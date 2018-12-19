package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"net"
	"sync"
	"time"

	"github.com/saesh/mandelbrot/pkg/farm/discovery"
	"github.com/saesh/mandelbrot/pkg/farm/node"
	gen "github.com/saesh/mandelbrot/pkg/generator"
	grpc "google.golang.org/grpc"
)

const (
	broadcastAddress = "239.0.0.0:5000"
)

type RenderNodeConfig struct {
	Hostname string
	IP       string
	Port     int
}

type HeadNode struct {
	Nodes []RenderNodeConfig
}

func main() {
	go broadcastService()
	startHeadNodeService()
}

func broadcastService() {
	broadcaster, err := discovery.NewBroadcaster(broadcastAddress)
	if err != nil {
		log.Fatalf("could not create multicast broadcaster: %v", err)
	}

	// blocking
	broadcaster.Start()
}

func startHeadNodeService() {
	headNode := &HeadNode{}

	srv := grpc.NewServer()

	node.RegisterHeadNodeServer(srv, headNode)

	l, err := net.Listen("tcp", ":8080")
	if err != nil {
		log.Fatalf("could not listen to :8080: %v", err)
	}
	log.Fatal(srv.Serve(l))
}

func (h *HeadNode) Register(ctx context.Context, registerRequest *node.RegisterRequest) (*node.Void, error) {
	log.Printf("render node registered: %v (%v:%v)\n", registerRequest.Hostname, registerRequest.Ip, registerRequest.Port)

	h.Nodes = append(h.Nodes, RenderNodeConfig{registerRequest.Hostname, registerRequest.Ip, int(registerRequest.Port)})

	log.Printf("number of nodes: %v\n", len(h.Nodes))

	if len(h.Nodes) == 2 {
		defer h.startRendering()
	}

	return &node.Void{}, nil
}

func (h *HeadNode) startRendering() error {

	// start rendering, TODO: move to own logicial component
	width := 3000
	height := 3000
	mb := gen.NewMandelbrot(width, height)

	mb.X = 0
	mb.Y = 0
	mb.R = 4

	nodesCount := len(h.Nodes)
	pixelPerNode := (width * height) / nodesCount

	for index, nodeConfig := range h.Nodes {
		configureRenderNode(nodeConfig, &node.RenderConfiguration{
			ColorPreset:   int32(mb.Colors),
			MaxIterations: int32(mb.MaxIterations),
			X:             float32(mb.X),
			Y:             float32(mb.Y),
			R:             float32(mb.R),
			StartIndex:    int32(index * pixelPerNode),
			EndIndex:      int32(index*pixelPerNode + pixelPerNode),
			Width:         int32(mb.Width),
			Height:        int32(mb.Height),
		})
	}

	var wg sync.WaitGroup
	wg.Add(len(h.Nodes))

	startTime := time.Now()
	for _, nodeConfig := range h.Nodes {
		go func(nodeConfig RenderNodeConfig) {
			timerStart := time.Now()
			log.Printf("[START] rendering on node: %v\n", nodeConfig.Hostname)
			conn, err := grpc.Dial(fmt.Sprintf("%v:%v", nodeConfig.IP, nodeConfig.Port), grpc.WithInsecure())
			if err != nil {
				wg.Done()
			}

			client := node.NewRenderNodeClient(conn)
			stream, err := client.IsMandelbrot(context.Background(), &node.Void{})
			if err != nil {
				log.Printf("Error writing to render node: %v", err)
			}
			waitc := make(chan struct{})

			// read results
			go func() {
				for {
					result, err := stream.Recv()
					if err == io.EOF {
						// read done.
						close(waitc)
						return
					}
					if err != nil {
						log.Fatalf("Failed to receive a compute result: %v", err)
					}
					// done read result here
					mb.ColorizeFunc(result.IsMandelbrot, int(result.Iteration), float64(result.Re), float64(result.Im), mb.MaxIterations, int(result.Index))
				}
			}()

			err = stream.CloseSend()
			if err != nil {
				log.Printf("error closing stream to client: %v\n", err)
			}
			<-waitc
			log.Printf("[DONE] rendering on node: %v, elapsed time: %v\n", nodeConfig.Hostname, time.Since(timerStart))
			wg.Done()
		}(nodeConfig)
	}

	wg.Wait()
	log.Printf("[DONE] total rendering time: %v", time.Since(startTime))
	mb.WriteJpeg("test.jpeg", 90)
	return nil
}

func configureRenderNode(nodeConfig RenderNodeConfig, config *node.RenderConfiguration) error {
	log.Printf("configuring render node: %v\n", nodeConfig.Hostname)
	conn, err := grpc.Dial(fmt.Sprintf("%v:%v", nodeConfig.IP, nodeConfig.Port), grpc.WithInsecure())
	if err != nil {
		return err
	}

	client := node.NewRenderNodeClient(conn)
	_, err = client.Configure(context.Background(), config)
	if err != nil {
		return err
	}

	return nil
}
