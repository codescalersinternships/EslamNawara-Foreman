package main

import "fmt"

func main() {

	foreman, err := parseProcfile("procfile.yml")
	if err != nil {
		fmt.Println(err)
		return
	}
	foreman.startForeman()
}
