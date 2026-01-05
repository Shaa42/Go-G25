package main

import (
	"elp-project/internal/audio"
	"log"
)

func main() {
	err := audio.ParseWav("assets/sample-3s.wav")
	if err != nil {
		log.Fatal("Can't open file")
	}
}
