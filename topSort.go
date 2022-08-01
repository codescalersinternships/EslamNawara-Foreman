package main

const (
	unvisited         = 0
	visited           = 1
	currentlyVisiting = 2
)

func topSort(graph map[string][]string) []string {
	var out []string
	state := make(map[string]int)
	var ts func(string)

	ts = func(node string) {
		if state[node] == visited {
			return
		}
		state[node] = visited
		for _, child := range graph[node] {
			ts(child)
		}
		out = append(out, node)
	}

	for node := range graph {
		ts(node)
	}

	return out
}

func isCyclic(graph map[string][]string) bool {
	state := make(map[string]int)
	cyclic := false
	var dfs func(string)

	dfs = func(node string) {
		if state[node] == visited {
			return
		}
		if state[node] == currentlyVisiting {
			cyclic = true
			return
		}

		state[node] = currentlyVisiting
		for _, child := range graph[node] {
			dfs(child)
		}
		state[node] = visited
	}

	for node := range graph {
		dfs(node)
	}
	return cyclic
}

func buildDepGraph(foreman Foreman) map[string][]string {
	graph := make(map[string][]string)
	for serviceName, info := range foreman.services {
		deps := make([]string, 0)
		for _, dep := range info.deps {
			deps = append(deps, dep)
		}
		graph[serviceName] = deps
	}
	return graph
}
