package main

import (
	"context"
	"fmt"
	"log"
	"net"

	"github.com/saesh/mandelbrot/pkg/farm/discovery"
	"github.com/saesh/mandelbrot/pkg/farm/headnode"
	grpc "google.golang.org/grpc"
)

const (
	broadcastAddress = "239.0.0.0:5000"
)

type HeadNode struct{}

func main() {
	go broadcastService()

	var headNode HeadNode

	srv := grpc.NewServer()

	headnode.RegisterHeadNodeServer(srv, headNode)

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

func (HeadNode) Register(ctx context.Context, registerRequest *headnode.RegisterRequest) (*headnode.Void, error) {
	fmt.Printf("render node registered: %v (%v:%v)\n", registerRequest.Hostname, registerRequest.Ip, registerRequest.Port)
	return &headnode.Void{}, nil
}
