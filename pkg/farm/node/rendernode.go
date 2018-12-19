package node

import (
	"context"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"sync"

	"github.com/saesh/mandelbrot/pkg/farm/discovery"
	g "github.com/saesh/mandelbrot/pkg/generator"
	"github.com/saesh/mandelbrot/pkg/util"
	grpc "google.golang.org/grpc"
)

const (
	broadcastAddress = "239.0.0.0:5000"
	headNodePort     = 8080
	grpcPort         = 8081
)

type RenderNode struct {
	RenderConfig *RenderConfiguration
	MB           *g.Mandelbrot
}

func (n *RenderNode) Configure(ctx context.Context, config *RenderConfiguration) (*Void, error) {
	n.RenderConfig = config
	n.MB = &g.Mandelbrot{
		Colors:        int(config.ColorPreset),
		MaxIterations: int(config.MaxIterations),
		R:             float64(config.R),
		X:             float64(config.X),
		Y:             float64(config.Y)}
	log.Println("received render configuration")
	return &Void{}, nil
}

func (n *RenderNode) IsMandelbrot(stream RenderNode_IsMandelbrotServer) error {
	coordinateChan := make(chan g.Coordinate, 100000)
	resultChan := make(chan g.MandelbrotResult, 100000)

	go g.IsMandelbrot(coordinateChan, resultChan, int(n.RenderConfig.MaxIterations))

	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		for {
			coordinate, err := stream.Recv()
			if err == io.EOF {
				log.Println("input stream EOF")
				defer close(coordinateChan)
				wg.Done()
				return
			}
			if err != nil {
				wg.Done()
				return
			}

			coordinateChan <- g.Coordinate{Re: float64(coordinate.Re), Im: float64(coordinate.Im), Index: int(coordinate.Index)}
		}
	}()

	go func() {
		var count = 0
		for r := range resultChan {
			count++
			computeResult := &ComputeResult{
				Re:           float32(r.Re),
				Im:           float32(r.Im),
				Iteration:    int32(r.Iterations),
				Index:        int32(r.Index),
				IsMandelbrot: r.IsMandelbrot}
			if err := stream.Send(computeResult); err != nil {
				log.Printf("Error sending compute result: %v", err)
			}
		}
		wg.Done()
	}()

	wg.Wait()

	return nil
}

// StartRenderNode is the main entry point for starting a render node.
// First, the head node is discovered and second a server is started
// for commands to be received
func StartRenderNode() {
	go func() {
		headNodeIP, err := discoverHeadNodeIP()
		if err != nil {
			log.Fatalf("could not discover head node: %v", err)
		}

		err = registerAtHeadNode(headNodeIP)
		if err != nil {
			log.Fatalf("could not register at head node: %v", err)
		}
	}()

	startRenderNodeServer()
}

func startRenderNodeServer() {
	srv := grpc.NewServer()

	RegisterRenderNodeServer(srv, &RenderNode{})

	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", grpcPort))
	if err != nil {
		log.Fatalf(fmt.Sprintf("could not listen to :%d: %v", grpcPort, err))
	}
	log.Fatal(srv.Serve(listener))
}

func discoverHeadNodeIP() (string, error) {
	listener, err := discovery.NewListener(broadcastAddress)
	if err != nil {
		return "", err
	}

	stopChan := make(chan struct{})
	ipChan := make(chan string)

	go listener.Start(stopChan, func(src *net.UDPAddr, numBytes int, bytes []byte) {
		ipChan <- src.IP.String()
		close(ipChan)
	})

	var headNodeIP string
	for ip := range ipChan {
		headNodeIP = ip
		close(stopChan)
	}

	return headNodeIP, nil
}

func registerAtHeadNode(ip string) error {
	log.Printf("registering at head node: %v\n", ip)
	conn, err := grpc.Dial(fmt.Sprintf("%v:%d", ip, headNodePort), grpc.WithInsecure())
	if err != nil {
		return err
	}

	client := NewHeadNodeClient(conn)
	err = register(context.Background(), client)
	if err != nil {
		return err
	}

	return nil
}

func register(ctx context.Context, client HeadNodeClient) error {
	ip, err := util.LocalNetworkIP()
	if err != nil {
		return err
	}

	hostname, err := os.Hostname()
	if err != nil {
		return err
	}

	request := &RegisterRequest{
		Hostname: hostname,
		Ip:       ip,
		Port:     grpcPort,
	}

	_, err = client.Register(ctx, request)

	return err
}

