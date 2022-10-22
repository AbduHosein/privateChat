package main

import (
	"encoding/gob"
	"fmt"
	"net"
	"os"
)

var count = 0

type Message struct {
	To      string
	From    string
	Content string
}

func handleConnection(c net.Conn) {

	for {
		//Read data from connection
		fmt.Print(">> ")
		dec := gob.NewDecoder(c)
		//netData, err := bufio.NewReader(c).ReadString('\n')
		//if err != nil {
		//	fmt.Println(err)
		//	return
		//}

		//temp := strings.TrimSpace(string(netData))
		message := &Message{}
		dec.Decode(message)
		temp := message.To
		if temp == "STOP" || temp == "" {
			fmt.Print("Client has exited...\n")
			break
		}
		fmt.Print("{To field, From field, content}: " + message.To + " " + message.From + " " + message.Content)
		//counter := strconv.Itoa(count) + "\n"
		//c.Write([]byte("Encoded struct received on serverW"))
	}
	c.Close()
}

func main() {
	arguments := os.Args
	if len(arguments) == 1 {
		fmt.Println("Please provide a port number!")
		return
	}
	fmt.Println("Server has started...")
	PORT := ":" + arguments[1]
	l, err := net.Listen("tcp4", PORT)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer l.Close()

	for {
		c, err := l.Accept()
		if err != nil {
			fmt.Println(err)
			return
		}
		fmt.Println("Server is awaiting a connection...")
		go handleConnection(c)
	}
}
