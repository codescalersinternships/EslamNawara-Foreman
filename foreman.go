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

func (foreman Foreman) startForeman() error {
	signals := make(chan os.Signal)

	graph := buildDepGraph(foreman)
	if isCyclic(graph) {
		return fmt.Errorf("Found conflect in dependencies")
	}
	services := topSort(graph)

	for _, service := range services {
		foreman.runProcess(service)
	}

	signal.Notify(signals, syscall.SIGCHLD)
	for {
		<-signals
		foreman.sigchildHandler()
	}
}

func (foreman *Foreman) runProcess(serviceName string) error {
    fmt.Println(serviceName + " started")
	service := foreman.services[serviceName]
	cmd := exec.Command(service.cmd, service.cmdArgs...)
	err := cmd.Start()
	if err != nil {
		return err
	}
	service.process = cmd.Process
	service.id = cmd.Process.Pid
	foreman.services[serviceName] = service

	go service.checker()
	return nil
}

func (service Service) checker() {
	ticker := time.NewTicker(interval)
	for {
		<-ticker.C
		checkCmd := exec.Command(service.check.cmd, service.check.cmdArgs...)
		err := checkCmd.Run()
		if err != nil {
			syscall.Kill(service.id, syscall.SIGINT)
			return
		}
	}

}

func (foreman Foreman) sigchildHandler() {
	for _, service := range foreman.services {
		childProcess, _ := process.NewProcess(int32(service.process.Pid))
		childStatus, _ := childProcess.Status()
		if childStatus == "Z" {
			service.process.Wait()
			if !service.runOnce {
				foreman.runProcess(service.serviceName)
			}
		}
	}
}
