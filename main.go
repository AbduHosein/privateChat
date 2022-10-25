package main

import (
	"fmt"
	"os"
	"privateChat/pkg/client"
	"privateChat/pkg/server"
)

// Main program to launch client or server based on cmd line args.
func main() {
	if len(os.Args) <= 1 {
		fmt.Printf("Please provide input in the form: \n")
		fmt.Printf("go run main.go SERVER {port} // To run the server\n")
		fmt.Printf("go run main.go CLIENT {host}:{port} {username} // To run a client\n\n")
	}

	// Store arguments
	run := os.Args[1]
	input := os.Args[2:]

	// Run server / client processes and handle invalid arguments with explanatory messages.
	if run == "SERVER" {
		server.Server(input[0])
	} else if run == "CLIENT" {
		client.Client(input[0], input[1])
	} else {
		fmt.Printf("Please provide input in the form: \n")
		fmt.Printf("go run main.go SERVER {port} // To run the server\n")
		fmt.Printf("go run main.go CLIENT {host}:{port} {username} // To run a client\n")
	}

}
