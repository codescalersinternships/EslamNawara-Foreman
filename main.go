package main

import "fmt"

func main() {

	foreman, err := parseProcfile("tests/Procfile")
	if err != nil {
		fmt.Println(err)
		return
	}
	err = foreman.StartForeman()
	if err != nil {
		fmt.Println(err)
		return
	}

}
