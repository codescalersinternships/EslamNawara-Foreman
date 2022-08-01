package main

import (
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"syscall"
	"time"
)

const interval = 100 * time.Millisecond

func startForeman() {
	signals := make(chan os.Signal)
	foreman, err := parseProcfile("procfile.yml")
	if err != nil {
		fmt.Println(err)
		return
	}
	graph := buildDepGraph(foreman)
	if isCyclic(graph) {
		fmt.Println("Found conflect in dependencies")
		return
	}
	services := topSort(graph)

	for _, service := range services {
		foreman.runProcess(service)
	}

	signal.Notify(signals, syscall.SIGINT, syscall.SIGCHLD)
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
	fmt.Println(serviceName + " started")
	service := foreman.services[serviceName]
	cmd := exec.Command(service.cmd, service.cmdArgs...)
	err := cmd.Start()
	if err != nil {
		return err
	}
	service.id = cmd.Process.Pid
	foreman.services[serviceName] = service
	fmt.Println("here")
	if !service.runOnce {
		ticker := time.NewTicker(interval)
		go func() {
			for {
				<-ticker.C
				cmd = exec.Command(service.check.cmd, service.check.cmdArgs...)
				err = cmd.Run()
				if err != nil {
					fmt.Println(serviceName + ": check failed ")
					syscall.Kill(service.id, syscall.SIGINT)
					return
				}
			}
		}()
	}
	return nil
}

func (foreman Foreman) sigintHandler() {
	foreman.active = false
	for _, service := range foreman.services {
		syscall.Kill(service.id, syscall.SIGINT)
	}
	os.Exit(0)
}

func (foreman Foreman) sigchildHandler() {
	for _, service := range foreman.services {
		err := syscall.Kill(service.id, 0)
		fmt.Println("here")
		if err != nil {
			fmt.Println("here 2")
			childProcess, _ := os.FindProcess(service.id)
			childProcess.Kill()
			if !service.runOnce && foreman.active {
				foreman.runProcess(service.serviceName)
			}
		}
	}
}
