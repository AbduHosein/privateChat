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

func check(err error) {
	if err != nil {
		fmt.Println(err)
		return
	}
}

func receiveMessages(c net.Conn, exit chan string) {

	dec := gob.NewDecoder(c)

	for {
		var m Message
		err := dec.Decode(&m)
		check(err)
		//content := strings.Join(m.Content, " ")
		if strings.TrimSpace(string(m.Content)) == "EXIT" {
			exit <- "EXIT"
			return
		}

		fmt.Print(m.From + ": " + m.Content + ">> ")
	}
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

	exit := make(chan string, 1)
	next := make(chan string, 1)
	go receiveMessages(c, exit)

	// Following block of code reads client Stdin, formats it, then sends to the server using Gob.
	encoder := gob.NewEncoder(c)

	// INIT-MESSAGE
	//content := strings.Split(USERNAME, " ")
	content := USERNAME
	m := Message{"SERVER", c.LocalAddr().String(), content}
	encoder.Encode(&m)
	check(err)

	// NOTE FROM JOHN:
	// THIS FOR-LOOP WORKS TO READ FROM CLIENT'S COMMAND LINE
	// NEED TO IMPLEMENT GO-ROUTINE TO DECODE INCOMING MESSAGES FROM THE SERVER

	for {

		go func() {
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
				exit <- "EXIT"
				return
			}

			// Here the raw text string is split into a slice, so that I can store the slice values in the message struct.
			send := strings.Split(text, " ")

			// ensure there is more than 2 inputs, ie, the client is sending a message.
			if len(send) > 2 && send[1] == USERNAME {
				// Store the message content in a slice, so users can now send longer messages.
				content := send[2:]

				// Message struct created, and sent to server using encoder.
				message := &Message{send[0], send[1], strings.Join(content, " ")}
				encoder.Encode(message)
			} else {
				fmt.Println("Invalid arguments, please input: To {Your USERNAME} Message")
			}

			next <- "Continue"

		}()

		select {
		case <-next:
			continue
		case <-exit:
			fmt.Print("Client is exitting...")
			return

		}
	}
}
