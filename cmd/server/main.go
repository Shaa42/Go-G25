// Server boilerplate

package main

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"net"
	"strings"
)

func handleConn(conn net.Conn) {
	defer conn.Close()
	fmt.Println("New connection.")

	for {
		// New message buffer
		var readBuffer *bufio.Reader = bufio.NewReader(conn)
		message, err := readBuffer.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				fmt.Println("Client closed the connection.")
				break
			} else {
				log.Fatal("Couldn't read string: ", err)
			}
		}

		message = strings.TrimSpace(message)

		log.Println(">>", message)
	}
}

func main() {
	// Open a TCP listening port
	fmt.Println("This is the main code for the server")
	listener, err := net.Listen("tcp", ":42069")
	if err != nil {
		log.Fatal("Couldn't open listening port.", err)
	}

	// Make sure the listening port is closed after the code execution
	defer listener.Close()

	for {
		// Accept incoming connexions
		conn, err := listener.Accept()
		if err != nil {
			log.Println("Couldn't accept connexion :", err)
			continue
		}

		go handleConn(conn)
	}

}
