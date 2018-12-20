package discovery

import (
	"net"
	"time"
)

type Broadcaster struct {
	connection *net.UDPConn
}

var (
	interval = 1 * time.Second
)

func NewBroadcaster(broadcastAddress string) (*Broadcaster, error) {
	addr, err := net.ResolveUDPAddr("udp", broadcastAddress)
	if err != nil {
		return nil, err
	}

	conn, err := net.DialUDP("udp", nil, addr)
	if err != nil {
		return nil, err
	}

	return &Broadcaster{connection: conn}, nil
}

func (b *Broadcaster) Broadcast(message string) {
	for {
		b.connection.Write([]byte(message))
		time.Sleep(interval)
	}
}
