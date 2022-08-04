package main

import (
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/shirou/gopsutil/process"
)

const interval = 50 * time.Millisecond

// start all services from yaml file
func (foreman Foreman) StartForeman() error {
	signals := make(chan os.Signal)

	graph := buildDepGraph(foreman)
	if isCyclic(graph) {
		return fmt.Errorf("Found conflect in dependencies")
	}
	services := topSort(graph)

	for _, service := range services {
		err := foreman.runService(service)
		if err != nil {
			return fmt.Errorf("failed to start the process %s", service)
		}
	}

	signal.Notify(signals, syscall.SIGCHLD, syscall.SIGINT)
	for {
		signal := <-signals
		switch signal {
		case syscall.SIGCHLD:
			foreman.sigchildHandler()
		case syscall.SIGINT:
			foreman.sigIntHandler()
		}
	}
}

// start one service and wait for it
func (foreman *Foreman) runService(serviceName string) error {
	fmt.Println(serviceName + " started")
	service := foreman.services[serviceName]
	service.status = true
	cmd := exec.Command("bash", "-c", service.cmd)
	err := cmd.Start()
	if err != nil {
		return err
	}
	service.process = cmd.Process
	foreman.services[serviceName] = service

	go foreman.checker(serviceName)
	return nil
}

func (foreman Foreman) checker(serviceName string) {
	service := foreman.services[serviceName]
	ticker := time.NewTicker(interval)
	for {
		<-ticker.C
		checkCmd := exec.Command("bash", "-c", service.check.cmd)
		err := checkCmd.Run()
		if err != nil {
			syscall.Kill(service.process.Pid, syscall.SIGINT)
			return
		}
		for _, dep := range service.deps {
			if !foreman.services[dep].status {
				syscall.Kill(service.process.Pid, syscall.SIGINT)
				return
			}
		}
		checkPort := func(portType string) {
			var ports []string
			switch portType {
			case "tcp":
				ports = service.check.tcpPorts
			case "udp":
				ports = service.check.udpPorts
			}
			for _, port := range ports {
				cmd := fmt.Sprintf("netstat -lnptu | grep %s | grep %s -m 1 | awk '{print $7}'", portType, port)
				out, _ := exec.Command("bash", "-c", cmd).Output()
				pid, err := strconv.Atoi(strings.Split(string(out), "/")[0])
				if err != nil || pid != service.process.Pid {
					fmt.Println(service.serviceName + " checher failed")
					syscall.Kill(service.process.Pid, syscall.SIGINT)
					return
				}
			}
		}
		checkPort("udp")
		checkPort("tcp")
	}
}

// handler for signal child for child process
func (foreman *Foreman) sigchildHandler() {
	for _, service := range foreman.services {
		childProcess, _ := process.NewProcess(int32(service.process.Pid))
		childStatus, _ := childProcess.Status()
		if childStatus == "Z" {
			service.status = false
			foreman.services[service.serviceName] = service
			fmt.Println(service.serviceName + " stopped")
			service.process.Wait()
			if !service.runOnce && foreman.active {
				foreman.runService(service.serviceName)
			}
		}
	}
}

// handler for signal interrupt for the main process
func (foreman *Foreman) sigIntHandler() {
	foreman.active = false
	for _, service := range foreman.services {
		syscall.Kill(service.process.Pid, syscall.SIGINT)
	}
	os.Exit(0)
}
