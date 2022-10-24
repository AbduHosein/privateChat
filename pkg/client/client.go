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

// Logger  to be used by the client to log messages.
var MessageLogger *log.Logger

// Struct to format messages sent in the chat room.
type Message struct {
	// Username of the destination client
	To string
	// Username of the source client
	From string
	// Body of the message
	Content string
}

// Simple error checking function.
func check(err error) {
	if err != nil {
		fmt.Printf("ERROR: %s\n", err)
		return
	}
}

// Function to be run as a GoRoutine
func receiveMessages(c net.Conn, dec *gob.Decoder) {

	for {
		var m Message
		err := dec.Decode(&m)

		if err == nil {

			if m.From == "SERVER" && m.Content == "EXIT" {

				c.Close()
				os.Exit(0)

			} else {

				fmt.Fprint(os.Stdout, "\r \r")
				MessageLogger.Printf("\n----------------------\nFrom: \t %s\nContent: %s", m.From, m.Content)
				fmt.Print(">> ")

			}
		} else {
			os.Exit(0)
		}
	}
}

// Function to run as a GoRoutine to catch a SIGINT flag.
// This allows the client and server to gracefully close their
// connection if the user hits `Crtl + C` to exit the program.
func catchSignalInterrupt(c net.Conn, enc *gob.Encoder, username string) {

	// Make a channel of `os.Signal` objects and fill the channel
	// with instances of SIGINT flags. If the user hits `Crtl + C`,
	// the `sigs` channel will be filled with an `os.Signal` object.
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, os.Interrupt)

	for {

		// If a signal is caught...
		<-sigs

		// ...send an 'EXIT' message to the Server...
		m := Message{"SERVER", username, "EXIT"}
		err := enc.Encode(m)
		check(err)

		// ...then close the connection and EXIT the program.
		c.Close()
		os.Exit(0)
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
			fmt.Println("Invalid arguments, please input: To {Your USERNAME} Message")
		}
	}

}

// Function to run at startup to output basic chatroom details.
func printChatRoomDetails(address, username string) {

	fmt.Println("----------------------")
	fmt.Printf("Chatroom Server: %s\nUsername: \t %s\n", address, username)
	fmt.Println("----------------------")
}

func Client(address, username string) {

	// Create a logger to log messages on the client with a timestamp.
	MessageLogger = log.New(os.Stdout, "MESSAGE: ", log.Ltime)

	printChatRoomDetails(address, username)

	// Connect to the server. DIAL connection once and then use this
	// connection instance throughout the program.
	c, err := net.Dial("tcp", address)
	check(err)

	// Create encoder and decoder to be used throughout the process.
	// Use of these encoder and decoder prevents new encoders/decoders
	// from being created in other areas of the program which read/write
	// on the same connection.
	enc := gob.NewEncoder(c)
	dec := gob.NewDecoder(c)

	// Start a GoRoutine to receive incoming messages from the server.
	go receiveMessages(c, dec)

	// Start a GoRoutine to catch a Signal Interupt from the user to gracefully
	// shutdown the client.
	go catchSignalInterrupt(c, enc, username)

	// The protocol indicates that a client sends an initialziation
	// message to the server when it first connects.
	//
	// Send the server an initialziation message.
	m := Message{"SERVER", username, "INIT"}
	enc.Encode(&m)
	check(err)

	// Start a loop to read users command line inputs. See `readCommandLine`
	// for details.
	readCommandLine(enc, username)

}
