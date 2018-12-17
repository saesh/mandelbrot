package main

import (
	"fmt"
	"log"
	"net"

	"github.com/saesh/mandelbrot/pkg/farm/discovery"
)

var listener discovery.Listener

func main() {
	listener, err := discovery.NewListener("239.0.0.0:5000")
	if err != nil {
		log.Fatalf("could not create multicast listener: %v", err)
	}

	// read ip from service broadcaster multicast
	stopChan := make(chan struct{})
	ipChan := make(chan string)
	go listener.Start(stopChan, getBroadcasterIP(ipChan))
	var serviceIp string
	for ip := range ipChan {
		fmt.Printf("service detected at %v\n", ip)
		serviceIp = ip
		close(stopChan)
	}

	// register at service
	fmt.Printf("registering at service: %v\n", serviceIp)
}

func getBroadcasterIP(ipChan chan string) func(*net.UDPAddr, int, []byte) {
	return func(src *net.UDPAddr, numBytes int, bytes []byte) {
		ipChan <- src.IP.String()
		close(ipChan)
	}
}
