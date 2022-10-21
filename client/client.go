package main

import (
	"bufio"
	"encoding/gob"
	"fmt"
	"net"
	"os"
	"strings"
)

type Message struct {
	To      string
	From    string
	Content string
}

func main() {
	arguments := os.Args
	if len(arguments) != 3 {
		fmt.Println("Please provide host:port USERNAME")
		return
	}

	CONNECT := arguments[1]
	USERNAME := arguments[2]
	fmt.Println(CONNECT, USERNAME)
	c, err := net.Dial("tcp", CONNECT)
	encoder := gob.NewEncoder(c)
	if err != nil {
		fmt.Println(err)
		return
	}

	for {
		reader := bufio.NewReader(os.Stdin)
		fmt.Print(">> ")
		text, _ := reader.ReadString('\n')
		// Sending message to the server
		//fmt.Fprintf(c, text+"\n")
		send := strings.Split(text, " ")
		//content := send[2:len(send)]
		message := &Message{send[0], send[1], send[2]}
		encoder.Encode(message)

		// Read response from the server.
		//message, _ := bufio.NewReader(c).ReadString('\n')

		// Print the server's response
		//fmt.Print("->: " + message)
		if strings.TrimSpace(string(text)) == "STOP" {
			fmt.Println("TCP client exiting...")
			return
		}
	}
}
