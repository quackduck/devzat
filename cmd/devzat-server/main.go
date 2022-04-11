package main

import "fmt"

func main() {
	var server DevzatServer

	if err := server.Init(); err != nil {
		fmt.Print(err.Error())
	}
}
