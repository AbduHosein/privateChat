package client

import (
	"bufio"
	"encoding/gob"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"strings"
)

var MessageLogger *log.Logger

// Message struct to relay to the server
type Message struct {
	To      string
	From    string
	Content string
}

type ArgumentsError struct{}

func (m *ArgumentsError) Error() string {
	return "Please provide host:port USERNAME"
}

func check(err error) {
	if err != nil {
		fmt.Println("ERROR:")
		fmt.Println(err)
		return
	}
}

func receiveMessages(c net.Conn) {

	dec := gob.NewDecoder(c)

	for {
		var m Message
		err := dec.Decode(&m)

		if err == nil {

			if m.From == "SERVER" {

				c.Close()
				os.Exit(0)

			} else {

				fmt.Fprint(os.Stdout, "\r \r")
				MessageLogger.Printf("\n----------------------\nFrom: \t %s\nContent: %s", m.From, m.Content)
				fmt.Print(">> ")

			}
		} else {
			print("Server has closed unexpectedly... ")
			c.Close()
			os.Exit(1)
		}
	}
}

func readCommandLine(enc *gob.Encoder, username string) {

	for {
		fmt.Print(">> ")

		// Reads input from command line, unformatted at this point
		text, err := bufio.NewReader(os.Stdin).ReadString('\n')
		check(err)

		// Exit client when the user types STOP.
		if strings.TrimSpace(string(text)) == "EXIT" {
			fmt.Println("TCP client exiting...")

			m := Message{"SERVER", username, "EXIT"}
			enc.Encode(m)
			return
		}

		// Here the raw text string is split into a slice, so that I can store the
		// slice values in the message struct.
		send := strings.Split(text, " ")

		// ensure there is more than 2 inputs, ie, the client is sending a message.
		if len(send) > 1 {
			// Store the message content in a slice, so users can now send longer messages.
			content := strings.Join(send[1:], " ")

			// Message struct created, and sent to server using enc.
			message := &Message{send[0], username, content}
			enc.Encode(message)
		} else {
			fmt.Println("Invalid arguments, please input: To Message")
		}
	}

}

func catchSignalInterrupt(c net.Conn, username string) {

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, os.Interrupt)

	for {
		<-sigs

		enc := gob.NewEncoder(c)

		m := Message{"SERVER", username, "EXIT"}
		err := enc.Encode(m)
		check(err)

		c.Close()
		os.Exit(0)
	}
}

func Client(address, username string) {

	MessageLogger = log.New(os.Stdout, "MESSAGE: ", log.Ltime)

	// Printed this just to see the args
	fmt.Println("----------------------")
	fmt.Printf("Chatroom Server: %s\nUsername: \t %s\n", address, username)
	fmt.Println("----------------------")

	// Connect to the server, DIAL is only used once.
	c, err := net.Dial("tcp", address)
	check(err)

	go receiveMessages(c)
	go catchSignalInterrupt(c, username)

	// Following block of code reads client Stdin, formats it, then sends to the server using Gob.
	enc := gob.NewEncoder(c)

	// Send an initializing message to the server so that the server can update its router table.
	m := Message{"SERVER", c.LocalAddr().String(), username}
	enc.Encode(&m)
	check(err)

	readCommandLine(enc, username)

}
