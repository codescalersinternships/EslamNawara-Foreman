package main

import (
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"syscall"
	"time"

	"github.com/shirou/gopsutil/process"
)

const interval = 100 * time.Millisecond

func (foreman Foreman) startForeman() {
	signals := make(chan os.Signal)

	graph := buildDepGraph(foreman)
	if isCyclic(graph) {
		fmt.Println("Found conflect in dependencies")
		return
	}
	services := topSort(graph)

	for _, service := range services {
		foreman.runProcess(service)
	}

	signal.Notify(signals, syscall.SIGCHLD, syscall.SIGINT)
	for {
		signal := <-signals
		switch signal {
		case syscall.SIGCHLD:
			foreman.sigchildHandler()
		case syscall.SIGINT:
			foreman.sigintHandler()
		}
	}
}

func (foreman *Foreman) runProcess(serviceName string) error {
	service := foreman.services[serviceName]
	cmd := exec.Command(service.cmd, service.cmdArgs...)
	err := cmd.Start()
	if err != nil {
		return err
	}
	service.process = cmd.Process
	service.id = cmd.Process.Pid
	foreman.services[serviceName] = service
	if !service.runOnce {
		ticker := time.NewTicker(interval)
		go func() {
			for {
				<-ticker.C
				checkCmd := exec.Command(service.check.cmd, service.check.cmdArgs...)
				err = checkCmd.Run()
				if err != nil {
					syscall.Kill(service.id, syscall.SIGINT)
					return
				}
			}
		}()
	}
	return nil
}

func (foreman Foreman) sigchildHandler() {
	for _, service := range foreman.services {
		childProcess, _ := process.NewProcess(int32(service.process.Pid))
		childStatus, _ := childProcess.Status()
		if childStatus == "Z" {
			service.process.Wait()
			if !service.runOnce && foreman.active {
				foreman.runProcess(service.serviceName)
			}
		}
	}
}

func (foreman Foreman) sigintHandler() {
	foreman.active = false
	for _, service := range foreman.services {
		syscall.Kill(service.id, syscall.SIGINT)
	}
	os.Exit(0)
}
