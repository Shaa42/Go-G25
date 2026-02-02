// Server: receive WavData via gob and reply with per-chunk ACK
package main

import (
	"encoding/gob"
	"io"
	"log"
	"net"
	"sync"

	"elp-project/internal/audio"
	"elp-project/internal/processor"
)

const WORKERS = 8

func worker(tasks <-chan audio.WavDataChunk, results chan<- audio.WavDataChunk, wg *sync.WaitGroup) {
	defer wg.Done()
	for chunk := range tasks {
		HandleSample(&chunk)
		results <- chunk
	}
}

// Send chunks to the client in the correct order
func sender(enc *gob.Encoder, results <-chan audio.WavDataChunk, done chan<- struct{}) {
	pending := make(map[int]audio.WavDataChunk)
	nextToSend := 1

	for chunk := range results {
		pending[chunk.ChunkID] = chunk

		// Send all consecutive chunks that are ready
		for {
			c, exists := pending[nextToSend]
			if !exists {
				break
			}

			if err := enc.Encode(c); err != nil {
				log.Println("Server: failed to send", err)
				close(done)
				return
			}
			log.Printf("Sender: chunk %d sent", nextToSend)

			delete(pending, nextToSend)
			nextToSend++
		}
	}
	close(done)
}

// Treat a single chunk of audio
func HandleSample(wdc *audio.WavDataChunk) {
	samplesFloat32 := wdc.ConvSampByteToFloat32()
	samplesFloat32 = processor.SubDB(samplesFloat32, 16.0)
	wdc.ConvSampFloat32ToByte(samplesFloat32)
}

// Handle the connection with the client
func handleConn(conn net.Conn) {
	defer conn.Close()
	log.Println("Server: new connection from", conn.RemoteAddr())

	dec := gob.NewDecoder(conn)
	enc := gob.NewEncoder(conn)

	tasks := make(chan audio.WavDataChunk, WORKERS*2)
	results := make(chan audio.WavDataChunk, WORKERS*2)
	done := make(chan struct{})

	// Start all workers
	var wg sync.WaitGroup
	for range WORKERS {
		wg.Add(1)
		go worker(tasks, results, &wg)
	}

	// Close results when all workers are done
	go func() {
		wg.Wait()
		close(results)
	}()

	// Start sender
	go sender(enc, results, done)

	for {
		// Wait for the client to send over data
		var wdc audio.WavDataChunk
		if err := dec.Decode(&wdc); err != nil {
			if err == io.EOF {
				log.Println("Server: client closed connection", conn.RemoteAddr())
			} else {
				log.Println("Server: gob decode error:", err)
			}
			break
		}
		tasks <- wdc

	}

	close(tasks)
	<-done
	log.Println("Server: finished processing for", conn.RemoteAddr())

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
