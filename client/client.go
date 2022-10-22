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
	Content []string
}

func main() {
	// Validate that client is supplying correct args.
	arguments := os.Args
	if len(arguments) != 3 {
		fmt.Println("Please provide host:port USERNAME")
		return
	}
	// Store args in local variables
	CONNECT := arguments[1]
	USERNAME := arguments[2]

	// TODO: Potentialy send the username over to server to be a key in our serveside Connections map that will store the various connections.

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
		text, err := reader.ReadString('\n')
		if err != nil {
			fmt.Println(err)
			return
		}

		// Exit client when the user types STOP.
		// TODO (STOP GO ROUTINE FROM SERVER): Add another condition here when you recieve STOP from the server.
		if strings.TrimSpace(string(text)) == "EXIT" {
			fmt.Println("TCP client exiting...")
			return
		}

		// Here the raw text string is split into a slice, so that I can store the slice values in the message struct.
		send := strings.Split(text, " ")

		// ensure there is more than 2 inputs, ie, the client is sending a message.
		if len(send) > 2 {
			// Store the message content in a slice, so users can now send longer messages.
			content := send[2:]

			// Message struct created, and sent to server using encoder.
			message := &Message{send[0], send[1], content}
			encoder.Encode(message)
		} else {
			fmt.Println("Invalid arguments, please input: To From Message")
		}
	}
}