package main

import (
	"context"
	"fmt"
	"log"
	"net"

	"github.com/saesh/mandelbrot/pkg/farm/discovery"
	"github.com/saesh/mandelbrot/pkg/farm/protocol"
	grpc "google.golang.org/grpc"
)

const (
	broadcastAddress = "239.0.0.0:5000"
)

type MandelbrotServer struct{}

func main() {
	go broadcastService()

	var mandelbrot MandelbrotServer

	srv := grpc.NewServer()

	protocol.RegisterMandelbrotServer(srv, mandelbrot)

	l, err := net.Listen("tcp", ":8080")
	if err != nil {
		log.Fatalf("could not listen to :8080: %v", err)
	}
	log.Fatal(srv.Serve(l))
}

func broadcastService() {
	broadcaster, err := discovery.NewBroadcaster(broadcastAddress)
	if err != nil {
		log.Fatalf("could not create multicast broadcaster: %v", err)
	}

	// blocking
	broadcaster.Start()
}

func (MandelbrotServer) Register(ctx context.Context, config *protocol.ClientConfig) (*protocol.Void, error) {
	fmt.Printf("Client connected: %v (%v:%v)\n", config.Hostname, config.Ip, config.Port)
	return &protocol.Void{}, nil
}
