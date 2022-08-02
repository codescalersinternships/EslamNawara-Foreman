package main

import "fmt"

func main() {

	foreman, err := parseProcfile("procfile.yml")
    fmt.Println("up and running")
	if err != nil {
		fmt.Println(err)
		return
	}
	err = foreman.startForeman()
	if err != nil {
		fmt.Println(err)
		return
	}

}
