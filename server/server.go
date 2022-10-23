package main

import (
	"bufio"
	"encoding/gob"
	"fmt"
	"net"
	"os"
	"strings"
)

// TODO: A MAP that can store all the connected clients. Was thinking something like {USERNAME: netConn}
// This way when the server recieves a message Struct, we check to see if its in the map, and handle the error if its not.
// If it is in the MAP, we could now construct a new Struct of the form Message{From, Content string}, ex--> {"Abdu: ", "Hello"}\

// Message struct to receive structs from the clients.
type Message struct {
	To      string
	From    string
	Content []string
}

// Global variable used to turn the server off
var serverStatus = "ON"

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

func receiveMessages(c net.Conn, incoming chan Message) {

	dec := gob.NewDecoder(c)

	// Move to incoming routine
	for {
		//Read data from connection
		message := &Message{}
		dec.Decode(message)
		temp := message.To

		//
		if temp == "" {
			fmt.Print("Client has exited...\n")
			c.Close()
		}

		incoming <- *message
	}

}

func sendMessages(c net.Conn, outgoing chan Message) {

	enc := gob.NewEncoder(c)

	for {
		var m = <-outgoing

		// TODO: NOTE FROM JOHN
		// NEED TO USE MAP TO CHECK IF MESSAGE SHOULD BE SENT TO THIS CLIENT
		// KEY INTO MAP USING THE USERNAME OF THE MESSAGE AND VERIFY THAT THE CORRESPONDING
		// PORT VALUE MATCHES THE PORT VALUE OF `c.RemoteAddr()`

		err := enc.Encode(m)
		if err != nil {
			fmt.Println(err)
			return
		}

	}
}

// TODO: When the ^^ stopChatroom goroutine receives the EXIT signal:
//   - Need to somehow communicate to all handleConnections go routines via a channel that the chatroom is exitting,
//   - Send the EXIT signal to the respective client via TCP.
func handleConnection(c net.Conn, signal chan string, incoming, outgoing chan Message) {
	// decoder used to recieve messages from the respective client.
	go receiveMessages(c, incoming)

	go sendMessages(c, outgoing)
}

func router(incoming, outgoing chan Message) {

	for {
		var m = <-incoming
		outgoing <- m
	}
}

func main() {
	// Handle errors for cmd line arguments
	arguments := os.Args
	if len(arguments) == 1 {
		fmt.Println("Please provide a port number!")
		return
	}

	// Server starts here
	fmt.Println("Server has started...")
	PORT := ":" + arguments[1]
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
	outgoing := make(chan Message, 5)

	go stopChatroom(signal)
	go router(incoming, outgoing)

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
		go handleConnection(c, signal, incoming, outgoing)

	}
}
