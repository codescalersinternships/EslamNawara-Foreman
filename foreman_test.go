package main

import (
	"testing"
)

const testProcfile = "tests/Procfile-test"
const testBadProcfile = "tests/Procfile-bad-test"
const testCyclicProcfile = "tests/Procfile-cyclic-test"

func TestNew(t *testing.T) {
	t.Run("Parse existing procfile with correct syntax", func(t *testing.T) {
		want := Foreman{
			services: map[string]Service{},
			active:   true,
		}
		sleeper := Service{
			serviceName: "sleeper",
			process:     nil,
			cmd:         "sleep infinity",
			runOnce:     true,
			deps:        []string{"hello"},
			check: Check{
				cmd:      "ls",
				tcpPorts: []string{"4759", "1865"},
				udpPorts: []string{"4500", "3957"},
			},
		}
		want.services["sleeper"] = sleeper

		hello := Service{
			serviceName: "hello",
			process:     nil,
			cmd:         `echo "hello"`,
			runOnce:     true,
			deps:        []string{},
		}
		want.services["hello"] = hello

		got, _ := parseProcfile(testProcfile)

		assertForeman(t, got, want)
	})

	t.Run("Run existing file with bad yml syntax", func(t *testing.T) {
		_, err := parseProcfile(testBadProcfile)
		if err == nil {
			t.Error("Expected error: yaml: unmarshal errors")
		}
	})

	t.Run("Run non-existing file", func(t *testing.T) {
		_, err := parseProcfile("unknown_file")
		want := "open unknown_file: no such file or directory"
		assertError(t, err, want)
	})
}

func TestBuildDependencyGraph(t *testing.T) {
	foreman, _ := parseProcfile("Procfile")

	got := buildDepGraph(foreman)
	want := make(map[string][]string)
	want["service_ping"] = []string{"service_redis"}
	want["service_sleep"] = []string{"service_ping"}

	assertGraph(t, got, want)
}

func TestIsCyclic(t *testing.T) {
	t.Run("run cyclic graph", func(t *testing.T) {
		foreman, _ := parseProcfile(testCyclicProcfile)
		graph := buildDepGraph(foreman)
		got := isCyclic(graph)
		if !got {
			t.Error("got:true, want:false")
		}
	})

	t.Run("run acyclic graph", func(t *testing.T) {
		foreman, _ := parseProcfile(testProcfile)
		graph := buildDepGraph(foreman)
		got := isCyclic(graph)
		if got {
			t.Error("got:false, want:true")
		}
	})
}

func TestTopSort(t *testing.T) {
	foreman, _ := parseProcfile("tests/Procfile")
	depGraph := buildDepGraph(foreman)
	got := topSort(depGraph)
	want := []string{"service_redis", "service_ping", "service_sleep"}
	assertList(t, got, want)
}

// Helpers
func assertForeman(t *testing.T, got, want Foreman) {
	t.Helper()

	for serviceName, service := range got.services {
		assertService(t, service, want.services[serviceName])
	}
}

func assertService(t *testing.T, got, want Service) {
	t.Helper()

	if got.serviceName != want.serviceName {
		t.Errorf("got:\n%q\nwant:\n%q", got.serviceName, want.serviceName)
	}

	if got.process != want.process {
		t.Errorf("got:\n%v\nwant:\n%v", got.process, want.process)
	}

	if got.cmd != want.cmd {
		t.Errorf("got:\n%q\nwant:\n%q", got.cmd, want.cmd)
	}

	if got.cmd != want.cmd {
		t.Errorf("got:\n%q\nwant:\n%q", got.cmd, want.cmd)
	}

	if got.runOnce != want.runOnce {
		t.Errorf("got:\n%t\nwant:\n%t", got.runOnce, want.runOnce)
	}

	assertList(t, got.deps, want.deps)
}

func assertChecks(t *testing.T, got, want Check) {
	t.Helper()

	if got.cmd != want.cmd {
		t.Errorf("got:\n%q\nwant:\n%q", got.cmd, want.cmd)
	}

	assertList(t, got.tcpPorts, want.tcpPorts)
	assertList(t, got.udpPorts, want.udpPorts)
}

func assertList(t *testing.T, got, want []string) {
	t.Helper()

	if len(want) != len(got) {
		t.Errorf("got:\n%v\nwant:\n%v", got, want)
	}

	n := len(want)
	for i := 0; i < n; i++ {
		if got[i] != want[i] {
			t.Errorf("got:\n%v\nwant:\n%v", got, want)
		}
	}
}

func assertError(t *testing.T, err error, want string) {
	t.Helper()

	if err == nil {
		t.Fatal("Expected Error: open unknown_file: no such file or directory")
	}

	if err.Error() != want {
		t.Errorf("got:\n%q\nwant:\n%q", err.Error(), want)
	}
}

func assertGraph(t *testing.T, got, want map[string][]string) {
	t.Helper()

	for key, value := range got {
		assertList(t, value, want[key])
	}
}
