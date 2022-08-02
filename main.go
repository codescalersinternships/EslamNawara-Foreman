package main

import "fmt"

func main() {

	foreman, err := parseProcfile("procfile.yml")
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
