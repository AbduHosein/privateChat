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

var InfoLogger *log.Logger

// Message struct to receive structs from the clients.
type Message struct {
	To      string
	From    string
	Content string
}

// Client connection struct that will store info associated with individual clients.
type ClientConnection struct {
	// Network address of the Client
	address string
	// TCP connection between server and client
	c net.Conn
	// Encoder which writes to the structs net.Conn
	enc gob.Encoder
}

// Router struct that handles incoming messages, and stores individual clients' ClientConnection structs with their usernames as keys.
type Router struct {
	incoming chan Message

	// Maps from a process's username to their ClientConnection address
	table map[string]ClientConnection
}

// stopChatroom readers user input from Stdin and sends the EXIT signal to the main process when EXIT is inputted on the server.
func stopChatroom(signal chan string) {
	for {
		fmt.Print(">> ")
		reader := bufio.NewReader(os.Stdin)
		text, err := reader.ReadString('\n')
		if err != nil {
			//fmt.Println(err)
			return
		}

		if strings.TrimSpace(string(text)) == "EXIT" {
			fmt.Println("Server is shutting down... Press Ctrl + C to exit")
			signal <- "EXIT"
			return
		}
	}
}

func receiveMessages(c net.Conn, router Router) {

	dec := gob.NewDecoder(c)

	// Move to incoming routine
	for {
		//Read data from connection
		message := &Message{}
		dec.Decode(message)
		temp := message.To

		// Check for blank message,
		if temp == "" {
			fmt.Print("Client has exited...\n")
			c.Close()
			break
		}

		// Check if it is a message from the server
		if temp == "SERVER" {

			enc := gob.NewEncoder(c)
			newConn := ClientConnection{c.RemoteAddr().String(), c, *enc}

			//username := strings.Join(message.Content, "")
			username := message.Content
			router.table[username] = newConn

			InfoLogger.Printf("New Client Added to `Router`: %s under the alias %s", c.RemoteAddr().String(), username)

			// Skip to next iteration, as this message does not need to be dispatched.
			continue
		}

		// Dispatch the message to the proper client
		router.dispatch(*message, c)
	}

}

func acceptConnections(l net.Listener, ch chan net.Conn) {
	for {
		c, err := l.Accept()
		if err != nil {
			fmt.Println(err)
			ch <- nil
			return
		}
		fmt.Println("Server has connected to a new client...")
		ch <- c
	}

}

func handleConnection(c net.Conn, router Router) {
	go receiveMessages(c, router)
}

func (r Router) dispatch(m Message, c net.Conn) {

	// Username value in the Message.To field
	destinationUsername := m.To

	// Check that this username value is present in the Router table
	if connection, ok := r.table[destinationUsername]; ok {
		//connection := r.table[destinationUsername]
		connection.enc.Encode(&m)

	} else {
		// Send the client an error message that this user is not online.
		connection := r.table[m.From]
		errMsg := Message{From: "ERROR", Content: "The user \"" + destinationUsername + "\" is not online.\n"}
		connection.enc.Encode(&errMsg)
	}
}

func main() {
	// Handle errors for cmd line arguments
	arguments := os.Args
	if len(arguments) == 1 {
		fmt.Println("Please provide a port number!")
		return
	}

	InfoLogger = log.New(os.Stdout, "INFO: ", log.Ltime|log.Lshortfile)

	// Server starts here
	fmt.Println("Server has started...")
	PORT := ":" + arguments[1]
	l, err := net.Listen("tcp4", PORT)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer l.Close()

	// Channels used to communicate between go routines, and the main process.
	accept := make(chan net.Conn)
	signal := make(chan string, 1)
	incoming := make(chan Message, 5)

	// Initializing our router table to handle incoming messages, and store info regarding connected clients.
	routerTable := make(map[string]ClientConnection)
	router := Router{incoming, routerTable}

	// Launch thread to read the EXIT message from the user
	go stopChatroom(signal)

	// Launch thread to accept new connections and pass them over to the main process.
	go acceptConnections(l, accept)

	// Implemented a strategy from https://www.rudderstack.com/blog/implementing-graceful-shutdown-in-go/, "Breaking the loop" section, after facing issues
	// with updating the EXIT signal via a global variable--causing the program to hang (see commit history).
	// This strategy ensures that we break the loop, or continue the loop based on signal channels rather than static variables.
	for {

		select {
		default: // Default case, when a new connection is sent, store it and pass to the handleConnection routine to start the new client.
			c := <-accept
			go handleConnection(c, router)
		case <-signal: // When exit signal is received simply dispatch the exitSignal struct to all connected net.Conn variables from the router table.
			for key, value := range router.table {
				// Construct the exit struct.
				exitSignal := Message{From: "Server", Content: "EXIT"}
				exitSignal.To = key
				clientC := value.c
				router.dispatch(exitSignal, clientC)
			}
			return
		}
	}
}
