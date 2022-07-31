package main

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

type Foreman struct {
	services map[string]Service
}

type Service struct {
	serviceName string
	cmd         string
	runOnce     bool
	checks      Check
	deps        []string
}

type Check struct {
	cmd      string
	tcpPorts []string
	udpPorts []string
}

func parseProcfile(filePath string) (Foreman, error) {
	foreman := Foreman{services: make(map[string]Service)}
	yamlMap := make(map[string]map[string]any)

	data, err := os.ReadFile(filePath)
	if err != nil {
		return Foreman{}, fmt.Errorf("Failed to read the procfile")
	}

	err = yaml.Unmarshal([]byte(data), yamlMap)
	if err != nil {
		return Foreman{}, fmt.Errorf("Bad format, Can't parse the procfile")
	}

	for serviceName, info := range yamlMap {
		foreman.services[serviceName] = parseService(serviceName, info)
	}
	return foreman, nil
}

func parseService(serviceName string, info map[string]any) Service {
	service := Service{
		serviceName: serviceName,
		deps:        []string{},
	}
	for key, value := range info {
		switch key {
		case "cmd":
			service.cmd = value.(string)
		case "run_once":
			service.runOnce = value.(bool)
		case "deps":
			for _, dep := range value.([]any) {
				service.deps = append(service.deps, dep.(string))
			}
		case "checks":
			service.checks = parseCheck(value)
		}
	}
	return service
}

func parseCheck(value any) Check {
	checks := Check{}
	for checkKey, checkValue := range value.(map[string]any) {
		switch checkKey {
		case "cmd":
			checks.cmd = checkValue.(string)
		case "tcp_ports":
			for _, port := range checkValue.([]any) {
				checks.tcpPorts = append(checks.tcpPorts, fmt.Sprint(port.(int)))
			}
		case "udp_ports":
			for _, port := range checkValue.([]any) {
				checks.udpPorts = append(checks.udpPorts, fmt.Sprint(port.(int)))
			}
		}
	}
	return checks
}
