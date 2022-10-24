package main

import (
	"bufio"
	"encoding/gob"
	"fmt"
	"log"
	"net"
	"os"
	"strings"
)

var MessageLogger *log.Logger

// Message struct to relay to the server
type Message struct {
	To      string
	From    string
	Content []string
}

type ArgumentsError struct{}

func (m *ArgumentsError) Error() string {
	return "Please provide host:port USERNAME"
}

func check(err error) {
	if err != nil {
		fmt.Println(err)
		return
	}
}

func configure() (string, string, error) {

	arguments := os.Args
	if len(arguments) != 3 {
		fmt.Println("Please provide host:port USERNAME")
		return "", "", &ArgumentsError{}
	}
	// Store args in local variables
	port := arguments[1]
	username := arguments[2]

	return port, username, nil
}

func receiveMessages(c net.Conn) {

	dec := gob.NewDecoder(c)

	for {
		var m Message
		err := dec.Decode(&m)
		check(err)

		content := strings.Join(m.Content, " ")

		fmt.Fprint(os.Stdout, "\r \r")
		MessageLogger.Printf("\n-----------------\nFrom: \t %s\nContent: %s", m.From, content)
		fmt.Print(">> ")
	}
}

func readCommandLine(enc *gob.Encoder) {

	for {
		fmt.Print(">> ")

		// Reads input from command line, unformatted at this point
		text, err := bufio.NewReader(os.Stdin).ReadString('\n')
		check(err)

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

			// Message struct created, and sent to server using enc.
			message := &Message{send[0], send[1], content}
			enc.Encode(message)
		} else {
			fmt.Println("Invalid arguments, please input: To From Message")
		}
	}

}

func init() {

	MessageLogger = log.New(os.Stdout, "MESSAGE: ", log.Ltime)

}

func main() {

	ADDRESS, USERNAME, err := configure()
	check(err)

	// Printed this just to see the args
	fmt.Println("----------------------")
	fmt.Printf("Chatroom Server: %s\nUsername: \t %s\n", ADDRESS, USERNAME)
	fmt.Println("----------------------")

	// Connect to the server, DIAL is only used once.
	c, err := net.Dial("tcp", ADDRESS)
	if err != nil {
		fmt.Println(err)
		return
	}

	go receiveMessages(c)

	// Following block of code reads client Stdin, formats it, then sends to the server using Gob.
	enc := gob.NewEncoder(c)

	// Send an initializing message to the server so that the server can update its router table.
	content := strings.Split(USERNAME, " ")
	m := Message{"SERVER", c.LocalAddr().String(), content}
	enc.Encode(&m)
	check(err)

	readCommandLine(enc)

}
