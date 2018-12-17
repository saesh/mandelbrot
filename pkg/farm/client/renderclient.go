package client

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net"
	"os"

	"github.com/saesh/mandelbrot/pkg/farm/discovery"
	"github.com/saesh/mandelbrot/pkg/farm/protocol"
	grpc "google.golang.org/grpc"
)

const (
	broadcastAddress = "239.0.0.0:5000"
)

func Start() {
	serviceIP, err := discoverServiceIP()
	if err != nil {
		log.Fatalf("could not discover service ip: %v", err)
	}

	err = registerAtService(serviceIP)
	if err != nil {
		log.Fatalf("could not register at server: %v", err)
	}
}

func discoverServiceIP() (string, error) {
	listener, err := discovery.NewListener(broadcastAddress)
	if err != nil {
		return "", err
	}

	// read ip from service broadcaster multicast
	stopChan := make(chan struct{})
	ipChan := make(chan string)

	go listener.Start(stopChan, getBroadcasterIP(ipChan))

	var serviceIP string
	for ip := range ipChan {
		serviceIP = ip
		close(stopChan)
	}

	return serviceIP, nil
}

func registerAtService(ip string) error {
	fmt.Printf("registering at service: %v\n", ip)
	conn, err := grpc.Dial(ip+":8080", grpc.WithInsecure())
	if err != nil {
		return err
	}

	client := protocol.NewMandelbrotClient(conn)
	err = register(context.Background(), client)
	if err != nil {
		return err
	}

	return nil
}

func getBroadcasterIP(ipChan chan string) func(*net.UDPAddr, int, []byte) {
	return func(src *net.UDPAddr, numBytes int, bytes []byte) {
		ipChan <- src.IP.String()
		close(ipChan)
	}
}

func register(ctx context.Context, client protocol.MandelbrotClient) error {
	ip, err := externalIP()
	if err != nil {
		return err
	}

	hostname, err := os.Hostname()
	if err != nil {
		return err
	}

	config := &protocol.ClientConfig{
		Hostname: hostname,
		Ip:       ip,
		Port:     8081,
	}

	_, err = client.Register(ctx, config)

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
