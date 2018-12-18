package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"net"
	"sync"

	"github.com/saesh/mandelbrot/pkg/colors"
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
	fmt.Printf("render node registered: %v (%v:%v)\n", registerRequest.Hostname, registerRequest.Ip, registerRequest.Port)

	h.Nodes = append(h.Nodes, RenderNodeConfig{registerRequest.Hostname, registerRequest.Ip, int(registerRequest.Port)})

	fmt.Printf("number of nodes: %v\n", len(h.Nodes))

	if len(h.Nodes) == 2 {
		defer h.startRendering()
	}

	return &node.Void{}, nil
}

func (h *HeadNode) startRendering() error {

	// start rendering, TODO: move to own logicial component
	mb := &gen.Mandelbrot{}

	mb.Width = 400
	mb.Height = 400
	mb.MaxIterations = 300
	mb.Colors = colors.GradientUltraFractal

	mb.X = 0
	mb.Y = 0
	mb.R = 4

	for _, nodeConfig := range h.Nodes {
		configureRenderNode(nodeConfig, &node.RenderConfiguration{
			ColorPreset:   int32(mb.Colors),
			MaxIterations: int32(mb.MaxIterations),
			X:             float32(mb.X),
			Y:             float32(mb.Y),
			R:             float32(mb.R),
		})
	}

	var wg sync.WaitGroup
	wg.Add(len(h.Nodes))

	for _, nodeConfig := range h.Nodes {
		go func(nodeConfig RenderNodeConfig) {
			fmt.Printf("rendering on node: %v\n", nodeConfig.Hostname)
			conn, err := grpc.Dial(fmt.Sprintf("%v:%v", nodeConfig.IP, nodeConfig.Port), grpc.WithInsecure())
			if err != nil {
				wg.Done()
			}

			client := node.NewRenderNodeClient(conn)
			stream, err := client.IsMandelbrot(context.Background())
			if err != nil {
				log.Printf("Error writing to render node: %v", err)
			}
			waitc := make(chan struct{})
			go func() {
				for {
					result, err := stream.Recv()
					if err == io.EOF {
						// read done.
						fmt.Println("receiving from stream DONE")
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

			coordinates := mb.Coordinates((mb.Width * mb.Height) / len(h.Nodes))
			for coordinate := range coordinates {
				if err := stream.Send(&node.Coordinate{Re: float32(coordinate.Re), Im: float32(coordinate.Im), Index: int32(coordinate.Index)}); err != nil {
					log.Fatalf("Failed to send a coordinate: %v", err)
				}
			}
			fmt.Println("sending coordinates DONE")
			err = stream.CloseSend()
			if err != nil {
				log.Printf("error closing stream to client: %v\n", err)
			}
			<-waitc
			fmt.Println("rendering on remote nodes DONE")
			wg.Done()
		}(nodeConfig)
	}

	wg.Wait()
	mb.WriteJpeg("test.jpeg", 90)
	return nil
}

func configureRenderNode(nodeConfig RenderNodeConfig, config *node.RenderConfiguration) error {
	fmt.Printf("configuring render node: %v\n", nodeConfig.Hostname)
	conn, err := grpc.Dial(fmt.Sprintf("%v:%v", nodeConfig.IP, nodeConfig.Port), grpc.WithInsecure())
	if err != nil {
		return err
	}

	client := node.NewRenderNodeClient(conn)
	err = configure(context.Background(), client, config)
	if err != nil {
		return err
	}

	return nil
}

func configure(ctx context.Context, client node.RenderNodeClient, renderConfig *node.RenderConfiguration) error {
	_, err := client.Configure(ctx, renderConfig)

	return err
}
