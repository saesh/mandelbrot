package main

import (
	"log"

	"github.com/saesh/mandelbrot/pkg/farm/discovery"
)

const (
	broadcastAddress = "239.0.0.0:5000"
)

func main() {
	broadcastService()
}

func broadcastService() {
	broadcaster, err := discovery.NewBroadcaster(broadcastAddress)
	if err != nil {
		log.Fatalf("could not create multicast broadcaster: %v", err)
	}

	// blocking
	broadcaster.Start()
}
