package rendernode

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net"
	"os"

	"github.com/saesh/mandelbrot/pkg/farm/discovery"
	"github.com/saesh/mandelbrot/pkg/farm/headnode"
	grpc "google.golang.org/grpc"
)

const (
	broadcastAddress = "239.0.0.0:5000"
	grpcPort         = 8081
)

func Start() {
	headNodeIP, err := discoverHeadNodeIP()
	if err != nil {
		log.Fatalf("could not discover head node: %v", err)
	}

	err = registerAtHeadNode(headNodeIP)
	if err != nil {
		log.Fatalf("could not register at head node: %v", err)
	}
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

	client := headnode.NewHeadNodeClient(conn)
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

func register(ctx context.Context, client headnode.HeadNodeClient) error {
	ip, err := externalIP()
	if err != nil {
		return err
	}

	hostname, err := os.Hostname()
	if err != nil {
		return err
	}

	request := &headnode.RegisterRequest{
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
