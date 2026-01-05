package main

import (
	"elp-project/internal/audio"
	"log"
)

func main() {
	_, err := audio.ParseWav("assets/sine_8k.wav")
	if err != nil {
		log.Fatal("Can't open file")
	}
}
