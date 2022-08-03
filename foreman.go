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

const interval = 100 * time.Millisecond

func (foreman Foreman) startForeman() error {
	signals := make(chan os.Signal)

	graph := buildDepGraph(foreman)
	if isCyclic(graph) {
		return fmt.Errorf("Found conflect in dependencies")
	}
	services := topSort(graph)

	for _, service := range services {
		err := foreman.runProcess(service)
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

func (foreman *Foreman) runProcess(serviceName string) error {
	fmt.Println(serviceName + " started")
	service := foreman.services[serviceName]
	cmd := exec.Command("bash", "-c", service.cmd)
	err := cmd.Start()
	if err != nil {
		return err
	}
	service.process = cmd.Process
	foreman.services[serviceName] = service

	go service.checker()
	return nil
}

func (service Service) checker() {
	ticker := time.NewTicker(interval)
	for {
		<-ticker.C
		checkCmd := exec.Command("bash", "-c", service.check.cmd)
		err := checkCmd.Run()
		if err != nil {
			syscall.Kill(service.process.Pid, syscall.SIGINT)
			return
		}

		if len(service.check.tcpPorts) > 0 {
			ports := service.check.tcpPorts
			for _, port := range ports {
				cmd := fmt.Sprintf("sudo netstat -lnptu | grep tcp | grep %s -m 1 | awk '{print $7}'", port)
				out, _ := exec.Command("bash", "-c", cmd).Output()
				pid, err := strconv.Atoi(strings.Split(string(out), "/")[0])
				if err != nil || pid != service.process.Pid {
					fmt.Println(service.serviceName + " checher failed")
					syscall.Kill(service.process.Pid, syscall.SIGINT)
					return
				}

			}
		}
		if len(service.check.udpPorts) > 0 {
			ports := service.check.udpPorts
			for _, port := range ports {
				cmd := fmt.Sprintf("sudo netstat -lnptu | grep udp | grep %s -m 1 | awk '{print $7}'", port)
				out, _ := exec.Command("bash", "-c", cmd).Output()
				pid, err := strconv.Atoi(strings.Split(string(out), "/")[0])
				if err != nil || pid != service.process.Pid {
					fmt.Println(service.serviceName + " checher failed")
					syscall.Kill(service.process.Pid, syscall.SIGINT)
					return
				}

			}
		}

	}

}

func (foreman Foreman) sigchildHandler() {
	for _, service := range foreman.services {
		childProcess, _ := process.NewProcess(int32(service.process.Pid))
		childStatus, _ := childProcess.Status()
		if childStatus == "Z" {
			fmt.Println(service.serviceName + " stopped")
			service.process.Wait()
			if !service.runOnce && foreman.active {
				foreman.runProcess(service.serviceName)
			}
		}
	}
}

func (foreman *Foreman) sigIntHandler() {
	foreman.active = false
	for _, service := range foreman.services {
		syscall.Kill(service.process.Pid, syscall.SIGINT)
	}
	os.Exit(0)
}
