package main

import "fmt"

func main() {
	fmt.Println("This is the main code for the client")

}

import (
	"fmt"
	"net"
)

func main() {
	listener, err := net.Listen("tcp", "localhost:6666")
	conn, err := listener.Accept()
	data := make([]byte, 1024)
	n, err := conn.Read(data)
	fmt.Println("Received:", string(buffer[:n]))

	conn.Write([]byte("Hello from server"))
}
