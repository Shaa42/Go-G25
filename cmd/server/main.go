// Server: receive WavData via gob and reply with per-chunk ACK
package main

import (
	"encoding/gob"
	"io"
	"log"
	"net"

	"elp-project/internal/audio"
	"elp-project/internal/processor"
)

// Handle the connection with the client
func handleConn(conn net.Conn) {
	defer conn.Close()
	log.Println("Server: new connection from", conn.RemoteAddr())

	dec := gob.NewDecoder(conn)
	enc := gob.NewEncoder(conn)
	for {
		// Wait for the client to send over data
		var wdc audio.WavDataChunk
		if err := dec.Decode(&wdc); err != nil {
			if err == io.EOF {
				log.Println("Server: client closed connection", conn.RemoteAddr())
				return
			}
			log.Println("Server: gob decode error:", err)
			return
		}

		// process the data
		HandleSample(&wdc)

		// Send data back to the client
		if err := enc.Encode(wdc); err != nil {
			log.Println("Server: gob encode ACK error:", err)
			return
		}

	}
}

func HandleSample(wdc *audio.WavDataChunk) {
	samplesFloat32 := wdc.ConvSampByteToFloat32()
	samplesFloat32 = processor.SubDB(samplesFloat32, 16.0)
	wdc.ConvSampFloat32ToByte(samplesFloat32)
}

func main() {
	// Open TCP listening server
	addr := ":42069"
	log.Println("Server: listening on", addr)

	// Open tcp socket on localhost:42069
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatal("Server: failed to listen:", err)
	}
	defer ln.Close()

	// Wait for incoming connection request and create new process to handle the new connection
	for {
		conn, err := ln.Accept()
		if err != nil {
			log.Println("Server: accept error:", err)
			continue
		}
		go handleConn(conn)
	}
}
