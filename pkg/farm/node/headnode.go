package node

import (
	"context"
	"fmt"
	"io"
	"log"
	"sync"
	"time"

	gen "github.com/saesh/mandelbrot/pkg/generator"
	grpc "google.golang.org/grpc"
)

type RenderNodeConfig struct {
	Hostname string
	IP       string
	Port     int
}

type HeadNode struct {
	Nodes           []RenderNodeConfig
	pixelChannel    chan *Pixel
	RequiredClients int
}

var pixelChannel = make(chan *Pixel, 20000000)

func (h *HeadNode) Register(ctx context.Context, registerRequest *RegisterRequest) (*Void, error) {
	log.Printf("render node registered: %v (%v:%v)\n", registerRequest.Hostname, registerRequest.Ip, registerRequest.Port)

	h.Nodes = append(h.Nodes, RenderNodeConfig{registerRequest.Hostname, registerRequest.Ip, int(registerRequest.Port)})

	log.Printf("number of nodes: %v\n", len(h.Nodes))

	if len(h.Nodes) == h.RequiredClients {
		defer h.startRendering()
	}

	return &Void{}, nil
}

func (h *HeadNode) Results(void *Void, srv HeadNode_ResultsServer) error {
	log.Println("WebSocket connection, waiting for compute results")

	for pixel := range pixelChannel {
		if err := srv.Send(pixel); err != nil {
			log.Printf("Error sending pixel: %v", err)
			return err
		}
	}

	return nil
}

func (h *HeadNode) startRendering() error {

	// start rendering, TODO: move to own logicial component
	width := 100
	height := 100
	mb := gen.NewMandelbrot(width, height)

	mb.X = 0
	mb.Y = 0
	mb.R = 4

	nodesCount := len(h.Nodes)
	pixelPerNode := (width * height) / nodesCount

	for nodeIndex, nodeConfig := range h.Nodes {
		configureRenderNode(nodeConfig, &RenderConfiguration{
			ColorPreset:   int32(mb.Colors),
			MaxIterations: int32(mb.MaxIterations),
			X:             float32(mb.X),
			Y:             float32(mb.Y),
			R:             float32(mb.R),
			StartIndex:    int32(nodeIndex * pixelPerNode),
			EndIndex:      int32(nodeIndex*pixelPerNode + pixelPerNode),
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

			client := NewRenderNodeClient(conn)
			stream, err := client.IsMandelbrot(context.Background(), &Void{})
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

					r, g, b, _ := mb.ColorizeFunc(result.IsMandelbrot, int(result.Iteration), float64(result.Re), float64(result.Im), mb.MaxIterations, int(result.Index))

					pixelChannel <- &Pixel{
						Index: result.Index,
						R:     r,
						G:     g,
						B:     b}
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
	close(pixelChannel)
	log.Printf("[DONE] total rendering time: %v", time.Since(startTime))
	mb.WriteJpeg("output.jpeg", 90)
	return nil
}

func configureRenderNode(nodeConfig RenderNodeConfig, config *RenderConfiguration) error {
	log.Printf("configuring render node: %v\n", nodeConfig.Hostname)
	conn, err := grpc.Dial(fmt.Sprintf("%v:%v", nodeConfig.IP, nodeConfig.Port), grpc.WithInsecure())
	if err != nil {
		return err
	}

	client := NewRenderNodeClient(conn)
	_, err = client.Configure(context.Background(), config)
	if err != nil {
		return err
	}

	return nil
}
