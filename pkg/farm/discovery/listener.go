package discovery

import (
	"log"
	"net"
	"time"
)

const (
	maxDatagramSize = 8192
)

type Listener struct {
	connection *net.UDPConn
}

func NewListener(address string) (*Listener, error) {
	addr, err := net.ResolveUDPAddr("udp", address)
	if err != nil {
		return nil, err
	}

	conn, err := net.ListenMulticastUDP("udp", nil, addr)
	if err != nil {
		return nil, err
	}

	conn.SetReadBuffer(maxDatagramSize)

	return &Listener{connection: conn}, nil
}

func (l *Listener) Start(stopChan chan struct{}, handler func(*net.UDPAddr, int, []byte)) {
	t := time.NewTicker(500 * time.Microsecond)
	for {
		select {
		case <-stopChan:
		case <-t.C:
			l.read(handler)
		}
		break
	}
}

func (l *Listener) read(handler func(*net.UDPAddr, int, []byte)) {
	buffer := make([]byte, maxDatagramSize)
	numBytes, src, err := l.connection.ReadFromUDP(buffer)
	if err != nil {
		log.Fatal("UDP read failed:", err)
	}

	if handler != nil {
		handler(src, numBytes, buffer)
	}
}
