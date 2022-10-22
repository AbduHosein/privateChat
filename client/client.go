package main

import (
	"bufio"
	"encoding/gob"
	"fmt"
	"net"
	"os"
	"strings"
)

// Message struct to relay to the server
type Message struct {
	To      string
	From    string
	Content string
}

func main() {
	// Validate that client is supplying correct args.
	arguments := os.Args
	if len(arguments) != 3 {
		fmt.Println("Please provide host:port USERNAME")
		return
	}
	// Store args in local variables
	// TODO: Potentialy send the username over to server to be a key in our serveside Connections map that will store the various connections.
	CONNECT := arguments[1]
	USERNAME := arguments[2]

	// Printed this just to see the args
	fmt.Println(CONNECT, USERNAME)

	// Connect to the server, DIAL is only used once.
	c, err := net.Dial("tcp", CONNECT)
	if err != nil {
		fmt.Println(err)
		return
	}

	// Following block of code reads client Stdin, formats it, then sends to the server using Gob.
	encoder := gob.NewEncoder(c)
	for {
		reader := bufio.NewReader(os.Stdin)
		fmt.Print(">> ")

		// Raw input from server, unformatted at this point
		text, _ := reader.ReadString('\n')

		// Exit client when the user types STOP.
		// TODO (STOP GO ROUTINE FROM SERVER): Add another condition here when you recieve STOP from the server.
		if strings.TrimSpace(string(text)) == "STOP" {
			fmt.Println("TCP client exiting...")
			return
		}

		// Here the raw text string is split into a slice, so that I can store the slice values in the message struct.
		send := strings.Split(text, " ")

		// TODO: the message content in the struct should be a slice that contains all the message contents.
		//content := send[2:] Experimenting here to pass over the message contents as a string, for messages longer than 1 word.

		// For now ensure that the message is len 3: To From Content
		if len(send) == 3 {
			// Message struct created, and sent to server using encoder.
			message := &Message{send[0], send[1], send[2]}
			encoder.Encode(message)
		} else {
			fmt.Println("Invalid arguments")
		}

	}
}
