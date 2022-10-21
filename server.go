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
	fmt.Print(">> ")
	for {
		//Read data from connection
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
		if temp == "STOP" {
			break
		}
		fmt.Println("{To field, From field, content}: ", message.To, message.From, message.Content)
		//counter := strconv.Itoa(count) + "\n"
		c.Write([]byte("Encoded struct received on server"))
	}
	c.Close()
}

func main() {
	arguments := os.Args
	if len(arguments) == 1 {
		fmt.Println("Please provide a port number!")
		return
	}

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
		go handleConnection(c)
		count++
	}
}
