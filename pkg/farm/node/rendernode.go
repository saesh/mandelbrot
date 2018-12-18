package node

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"sync"

	"github.com/saesh/mandelbrot/pkg/farm/discovery"
	g "github.com/saesh/mandelbrot/pkg/generator"
	grpc "google.golang.org/grpc"
)

const (
	broadcastAddress = "239.0.0.0:5000"
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
	fmt.Println("received render configuration")
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
				fmt.Println("stream EOF")
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

func Start() {
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

	startRenderNodeService()
}

func startRenderNodeService() {
	renderNode := &RenderNode{}

	srv := grpc.NewServer()

	RegisterRenderNodeServer(srv, renderNode)

	l, err := net.Listen("tcp", ":8081")
	if err != nil {
		log.Fatalf("could not listen to :8081: %v", err)
	}
	log.Fatal(srv.Serve(l))
}

func discoverHeadNodeIP() (string, error) {
	listener, err := discovery.NewListener(broadcastAddress)
	if err != nil {
		return "", err
	}

	stopChan := make(chan struct{})
	ipChan := make(chan string)

	go listener.Start(stopChan, getSourceIP(ipChan))

	var headNodeIP string
	for ip := range ipChan {
		headNodeIP = ip
		close(stopChan)
	}

	return headNodeIP, nil
}

func registerAtHeadNode(ip string) error {
	fmt.Printf("registering at head node: %v\n", ip)
	conn, err := grpc.Dial(fmt.Sprintf("%v:8080", ip), grpc.WithInsecure())
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

func getSourceIP(ipChan chan string) func(*net.UDPAddr, int, []byte) {
	return func(src *net.UDPAddr, numBytes int, bytes []byte) {
		ipChan <- src.IP.String()
		close(ipChan)
	}
}

func register(ctx context.Context, client HeadNodeClient) error {
	ip, err := externalIP()
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

func externalIP() (string, error) {
	ifaces, err := net.Interfaces()
	if err != nil {
		return "", err
	}
	for _, iface := range ifaces {
		if iface.Flags&net.FlagUp == 0 {
			continue // interface down
		}
		if iface.Flags&net.FlagLoopback != 0 {
			continue // loopback interface
		}
		addrs, err := iface.Addrs()
		if err != nil {
			return "", err
		}
		for _, addr := range addrs {
			var ip net.IP
			switch v := addr.(type) {
			case *net.IPNet:
				ip = v.IP
			case *net.IPAddr:
				ip = v.IP
			}
			if ip == nil || ip.IsLoopback() {
				continue
			}
			ip = ip.To4()
			if ip == nil {
				continue // not an ipv4 address
			}
			return ip.String(), nil
		}
	}
	return "", errors.New("are you connected to the network?")
}
