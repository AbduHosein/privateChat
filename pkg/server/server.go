package server

import (
	"bufio"
	"encoding/gob"
	"fmt"
	"log"
	"net"
	"os"
	"strings"
)

// Logger variable to be used by the server to log all messages on
// the server's activity.
var InfoLogger *log.Logger

// Struct to format messages sent in the chat room.
type Message struct {
	// Username of the destination client
	To string
	// Username of the source client
	From string
	// Body of the message
	Content string
}

// Struct to contain data for every connection with a client.
// `enc` & `dec` fields can be used to encode/decode messages
// over the TCP connection with a client. These fields are stored
// so that only one encoder and decoder object can be created for
// each connection with a client, and can then be shared by all
// GoRoutines running on the server.
type ClientConnection struct {
	// Network connection with the Client
	c net.Conn
	// Encoder which writes to the newtork connection
	enc *gob.Encoder
	// Decoder which reads from the network connection
	dec *gob.Decoder
}

// Wrapper struct to hold a map of all the Client's that are currently
// connected to the chat room. This struct is needed so that it can be
// used as a receiver:
//
//	-dipatch(m Message)
//	-dispatchMulti(content string)
//
// Using the struct as a receiver in prevents prop-drilling from having
// to pass the map down through subsequent function calls.
type Router struct {
	// Maps a Client's username to their ClientConnection struct
	table map[string]ClientConnection
}

// Simple error checking function.
func check(err error) {
	if err != nil {
		fmt.Printf("ERROR: %s\n", err)
		return
	}
}

// Function to be ran as a GoRoutine for each Client that connects to
// the server.
func receiveMessages(router *Router, conn ClientConnection) {

	for {

		// Read a new message from the connecton
		var m Message
		conn.dec.Decode(&m)

		// When the call to `conn.dec.Decode(&m)` throws an error, no message
		// is decoded into the struct, this occurs when a client throws a Signal
		// Interrupt, and the connection is aborted. In order to catch this case,
		// this condition checks for a NIL Message and skips to the next iteration
		// to decode the 'EXIT' message that is sent from a client when it catches
		// a Signal Interupt.
		if m.To == "" && m.From == "" && m.Content == "" {
			continue
		}

		// If `m` is an 'EXIT' message from a client addressed to a server...
		if m.To == "SERVER" && m.Content == "EXIT" {

			// ...delete the client connection from the router table...
			delete(router.table, m.From)

			// ...log the event to os.Stdout...
			InfoLogger.Printf("%s has left the chat.", string(m.From))
			fmt.Print(">> ")

			// ...and finally close the connection with the client.
			conn.c.Close()
			return

			// If message is not NULL and is not an 'EXIT' message...
		} else {
			// ...dispatch the message to the proper client.
			router.dispatch(m)
		}

	}

}

// Function to be run as a GoRoutine when a Client connects to the
// server. `handleConnection` initializes a `ClientConnection` struct
// and adds the new client to the `router.table` map.
func handleConnection(c net.Conn, router *Router) {

	// Create Encoders and Decoders to be used across the process.
	// Creating them at initialization and storing them in the
	// ClientConnection struct prevents multiple Encoders and Decoders
	// being created on the same `net.Conn` object.
	enc := gob.NewEncoder(c)
	dec := gob.NewDecoder(c)
	newConn := ClientConnection{c, enc, dec}

	// The protocol indicates that a client sends an initialziation
	// message to the server when it first connects.
	//
	// This initialization message is decoded here.
	var m Message
	err := dec.Decode(&m)
	check(err)

	// Initialize a new field in the `router.table` map.
	username := m.From
	router.table[username] = newConn
	InfoLogger.Printf("%s has joined the chat.", username)
	fmt.Print(">> ")

	// Start loop to receive messages from this client. See
	// `receiveMessages` for details.
	receiveMessages(router, newConn)

}

// Function to send deliver an incoming message to the proper client.
func (r *Router) dispatch(m Message) {

	username := m.To

	// If the username value is present in the Router table...
	if conn, ok := r.table[username]; ok {

		// ...use the `ClientConnection` to encode the message to the
		// destintation client.
		conn.enc.Encode(&m)

		// If the username value is not present in the Router table...
	} else {

		// ......use the `ClientConnection` to send the source client
		// an error message.
		conn := r.table[m.From]
		errMsg := Message{To: m.From, From: "SERVER", Content: "The user \"" + username + "\" is not online."}
		conn.enc.Encode(errMsg)
	}
}

// Function to deliver messages to all clients in the router table.
func (r *Router) dispatchMulti(content string) {

	// Loop over every client stored in `router.table`...
	for user, conn := range r.table {

		// ...for each connection, use the ClientConnections `enc`
		// field to encode a Message over the network connection.
		m := Message{user, "SERVER", content}
		err := conn.enc.Encode(m)
		check(err)

	}
}

// Function to be run as a GoRoutine to read user inputs from the command line.
func readCommandLine(r *Router) {

	for {

		fmt.Print(">> ")

		// Read input from the command line...
		text, err := bufio.NewReader(os.Stdin).ReadString('\n')
		check(err)

		// ...if user enters "EXIT"...
		if strings.TrimSpace(string(text)) == "EXIT" {

			// Log the event...
			InfoLogger.Println("The chat room is shutting down...")

			// ...dispatch an "EXIT" message to all clients on the network...
			r.dispatchMulti("EXIT")

			// ...and shutdown the server.
			os.Exit(0)
		}
	}
}

// Run the Chat Room server.
func Server(port string) {

	// Create a logger to log events on the server with a timestamp.
	InfoLogger = log.New(os.Stdout, "\r \rINFO: ", log.Ltime)

	// Start the server
	PORT := ":" + port
	l, err := net.Listen("tcp4", PORT)
	check(err)
	defer l.Close()

	// Log the event.
	InfoLogger.Println("Server has started...")

	// Create an instance of the router table to be used across all
	// GoRoutines created by the process.
	routerTable := make(map[string]ClientConnection)
	router := Router{routerTable}

	// Start a GoRoutine to read user inputs from the command line.
	go readCommandLine(&router)

	// Listen for new client connections...
	for {
		// ...when a new client connects...
		c, err := l.Accept()
		check(err)

		// ...start a GoRoutine to handle the connection with the new client.
		go handleConnection(c, &router)
	}

}
