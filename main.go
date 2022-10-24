package main

import (
	"os"
	"privateChat/pkg/client"
	"privateChat/pkg/server"
)

func main() {

	run := os.Args[1]

	input := os.Args[2:]

	if run == "SERVER" {
		server.Server(input[0])
	} else if run == "CLIENT" {
		client.Client(input[0], input[1])
	}

}
