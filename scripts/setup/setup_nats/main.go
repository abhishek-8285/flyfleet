package main

import (
	"fmt"
	"log"

	"github.com/nats-io/nats.go"
)

func main() {
	nc, err := nats.Connect("nats://localhost:4222")
	if err != nil {
		log.Fatal(err)
	}
	defer nc.Close()

	js, err := nc.JetStream()
	if err != nil {
		log.Fatal(err)
	}

	_, err = js.AddStream(&nats.StreamConfig{
		Name:     "fleet-telemetry",
		Subjects: []string{"fleet.telemetry.*"},
	})
	if err != nil {
		log.Printf("Stream error: %v", err)
	} else {
		fmt.Println("Stream fleet-telemetry created.")
	}
}
