package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
)

var (
	entering = make(chan chan<- string)
	leaving  = make(chan chan<- string)
	messages = make(chan string)
)

func main() {
	listener, err := net.Listen("tcp", "localhost:8001")
	if err != nil {
		log.Fatal(err)
	}
	go broadcaster()
	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Print(err)
			continue
		}
		go handleConn(conn)
	}
}

func broadcaster() {
	clients := make(map[chan<- string]string)
	for {
		select {
		case msg := <-messages:
			for ch := range clients {
				ch <- msg
			}
		case ch := <-entering:
			clients[ch] = ""
		case ch := <-leaving:
			delete(clients, ch)
			close(ch)
		}
	}
}

func handleConn(conn net.Conn) {
	// Here, I'm going to create a channel of string,
	// this channel will have the messaging functionality.
	// meaning the client writer will loop over this channel and prints the data
	// accordingly.

	ch := make(chan string)
	go clientWriter(conn, ch)

	who := conn.RemoteAddr().String()
	ch <- "You are " + who

	// When the user initially connects to the server,
	// at first the message of client idenity will be displayed
	// Secondly this will broadcast the arriving message to all the users.

	messages <- who + "has arrived"
	entering <- ch

	input := bufio.NewScanner(conn)
	for input.Scan() {
		messages <- who + ":" + input.Text()
	}
	leaving <- ch
	messages <- who + "has left"
	conn.Close()
}

func clientWriter(conn net.Conn, ch <-chan string) {
	for msg := range ch {
		fmt.Fprintln(conn, msg)
	}
}
