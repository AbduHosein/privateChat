package main

import (
	"encoding/gob"
	"fmt"
	"net"
	"os"
	"strings"
)

// TODO: A MAP that can store all the connected clients. Was thinking something like {USERNAME: netConn}
// This way when the server recieves a message Struct, we check to see if its in the map, and handle the error if its not.
// If it is in the MAP, we could now construct a new Struct of the form Message{From, Content string}, ex--> {"Abdu: ", "Hello"}
type Message struct {
	To      string
	From    string
	Content []string
}

// TODO: have a channel that can recieve values from the main() server block.
// Was thinking we could have a channel to recieve the "STOP" signal, and a channel to recieve the Message struct
func handleConnection(c net.Conn) {
	// decoder used to recieve messages from the respective client.
	dec := gob.NewDecoder(c)

	// encoder to send messages to the respective client.
	//enc := gob.NewEncoder(c)

	for {
		//Read data from connection
		fmt.Print(">> ")

		//temp := strings.TrimSpace(string(netData)) from earlier code I took from concurrentTCP linode article.
		message := &Message{}
		dec.Decode(message)
		temp := message.To

		// I think here the only condition that matters is "" cause if the client exits, thats what is stored in the temp variable.
		if temp == "STOP" || temp == "" {
			fmt.Print("Client has exited...\n>> ")
			break
		}

		// For testing purposes only, wont need this later on...
		// Printing the To and From values as is... and joining the Content value into one string, seperated by " ".
		// NOTE: message.Content is a slice containing the message values in each index.
		fmt.Print("{To field, From field, content}: " + message.To + " " + message.From + " " + strings.Join(message.Content, " "))
		//fmt.Print(message.Content[1])
	}
	c.Close()
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

	// This block handles incoming connections
	// TODO (STOP GO ROUTINE FROM SERVER): launch a stopChatroom go routine that will read from Stdin from server-side.
	// If it reads "STOP", send the "STOP" string to all client threads.
	for {
		c, err := l.Accept()
		if err != nil {
			fmt.Println(err)
			return
		}
		fmt.Println("Server has connected to a new client...")
		go handleConnection(c)
	}
}
