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

// TODO: A MAP that can store all the connected clients. Was thinking something like {USERNAME: netConn}
// This way when the server recieves a message Struct, we check to see if its in the map, and handle the error if its not.
// If it is in the MAP, we could now construct a new Struct of the form Message{From, Content string}, ex--> {"Abdu: ", "Hello"}\

var InfoLogger *log.Logger

// Message struct to receive structs from the clients.
type Message struct {
	To      string
	From    string
	Content string
}

type ClientConnection struct {
	// Network address of the Client
	address string
	// TCP connection between server and client
	c net.Conn
	// Encoder which writes to the structs net.Conn
	enc gob.Encoder
	// Decoder which reads from the structs net.Conn
	dec gob.Decoder
}

type Router struct {
	incoming chan Message

	// Maps from a process's username to their ClientConnection address
	table map[string]ClientConnection
}

// Global variable used to turn the server off
var serverStatus = "ON"

func check(err error) {
	if err != nil {
		fmt.Println("ERROR:")
		fmt.Println(err)
		return
	}
}

// TODO: When EXIT is received, need to communicate via a channel to all the active handleconnection routines.
// stopChatroom readers user input from Stdin and updates the serverStatus global variable when the EXIT msg is inputted.
func stopChatroom(ch chan string) {

	for {
		fmt.Print(">> ")
		reader := bufio.NewReader(os.Stdin)
		text, err := reader.ReadString('\n')
		if err != nil {
			fmt.Println(err)
			return
		}

		if strings.TrimSpace(string(text)) == "EXIT" {
			fmt.Println("Server is shutting down... Press Ctrl + C to exit")
			serverStatus = "OFF"
			return
		}
	}
}

func receiveMessages(c net.Conn, router Router, enc *gob.Encoder, dec *gob.Decoder) {

	for {
		//Read data from connection
		message := &Message{}
		dec.Decode(message)
		dest := message.To

		// Check if it is a message to be delivered to the server
		if dest == "SERVER" {

			if message.Content == "EXIT" {

				// Add in function/lines here to delete the exixting client from the Router struct
				delete(router.table, message.From)

				InfoLogger.Printf("%s has left the chat.", string(message.From))
				c.Close()
				break

			}

			// Skip to next iteration, as this message does not need to be dispatched.
			continue
		} else if dest == "" {
			// Accept empty message that comes when connection is closed
			continue
		} else {
			// Dispatch the message to the proper client
			router.dispatch(*message, c)
		}

	}

}

// TODO: When the ^^ stopChatroom goroutine receives the EXIT signal:
//   - Need to somehow communicate to all handleConnections go routines via a channel that the chatroom is exitting,
//   - Send the EXIT signal to the respective client via TCP.
func handleConnection(c net.Conn, signal chan string, router Router) {

	enc := gob.NewEncoder(c)
	dec := gob.NewDecoder(c)

	// Add new connection to the router table.
	newConn := ClientConnection{c.RemoteAddr().String(), c, *enc, *dec}

	var m Message
	err := dec.Decode(&m)
	check(err)

	//username := strings.Join(message.Content, "")
	username := m.Content
	router.table[username] = newConn

	InfoLogger.Printf("New Client Added to `Router`: %s under the alias %s", c.RemoteAddr().String(), username)

	go receiveMessages(c, router, enc, dec)

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
		errMsg := Message{From: "ERROR", Content: "The user " + destinationUsername + " is not online.\n"}
		connection.enc.Encode(&errMsg)
	}
}

func Server(port string) {

	InfoLogger = log.New(os.Stdout, "INFO: ", log.Ltime|log.Lshortfile)

	// Server starts here
	fmt.Println("Server has started...")
	PORT := ":" + port
	l, err := net.Listen("tcp4", PORT)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer l.Close()

	// Launch the thread that will read Stdin from user on server side and update serverStatus global variable.
	// TODO: implement the signal channel and pass to stopChatroom, and somehow use this channel to communication with all handleConnection routines.
	signal := make(chan string)
	incoming := make(chan Message, 5)

	routerTable := make(map[string]ClientConnection)
	router := Router{incoming, routerTable}

	go stopChatroom(signal)

	// This block handles incoming connections while serverstatus is ON.
	// TODO: Ensure that the program terminates when serverStatus is OFF, ie, make sure all handleConnection routines exit.
	for serverStatus == "ON" {
		c, err := l.Accept()
		if err != nil {
			fmt.Println(err)
			return
		}
		fmt.Println("Server has connected to a new client...")

		// TODO: make sure that these routines exit when signal feeds the "EXIT" string

		// Create a client connection type to manage the new connection and
		// newSendChan := make(chan Message, 5)
		// newConn := ClientConnection{c.RemoteAddr().String(), newSendChan}
		// router.connections = append(router.connections, newConn)

		go handleConnection(c, signal, router)

	}
}
