package main

import (
	"context"
	"flag"
	"log"
	"net"
	"net/http"

	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"github.com/saesh/mandelbrot/pkg/farm/discovery"
	"github.com/saesh/mandelbrot/pkg/farm/node"
	"github.com/tmc/grpc-websocket-proxy/wsproxy"
	grpc "google.golang.org/grpc"
)

var (
	broadcastAddress = flag.String("broadcastaddress", "239.0.0.0:5000", "broadcast address")
	grpcAddress      = flag.String("grpcaddress", ":8080", "listen grpc address")
	httpAddr         = flag.String("addr", ":8000", "listen http addr")
)

func run() error {
	go broadcastHeadNode(*broadcastAddress)

	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	if err := listenGRPC(*grpcAddress); err != nil {
		return err
	}

	mux := runtime.NewServeMux()
	opts := []grpc.DialOption{grpc.WithInsecure()}
	err := node.RegisterHeadNodeHandlerFromEndpoint(ctx, mux, *grpcAddress, opts)
	if err != nil {
		return err
	}

	log.Printf("WebSocket listening on %v\n", *httpAddr)
	http.ListenAndServe(*httpAddr, wsproxy.WebsocketProxy(mux))

	return nil
}

func broadcastHeadNode(address string) {
	broadcaster, err := discovery.NewBroadcaster(address)
	if err != nil {
		log.Printf("could not create multicast broadcaster: %v\n", err)
	}

	log.Printf("broadcasting on %v\n", address)

	// blocking
	broadcaster.Broadcast("mandelbrot-head-node")
}

func listenGRPC(address string) error {
	listener, err := net.Listen("tcp", address)
	if err != nil {
		return err
	}

	grpcServer := grpc.NewServer()
	node.RegisterHeadNodeServer(grpcServer, &node.HeadNode{})

	log.Printf("starting GRPC server on %v", address)
	go func() {
		if err := grpcServer.Serve(listener); err != nil {
			log.Println("GRPC server error:", err)
		}
	}()

	return nil
}

func main() {
	flag.Parse()

	if err := run(); err != nil {
		log.Fatal(err)
	}
}
