package network

import "fmt"
import "log"
import "net"
import "time"

type Peer struct {
	port int
	conn net.Conn
	live bool
	failCount int
}

// Retries until it is connected
// TODO: make this threadsafe
// TODO: reconnect after a disconnect
func (p *Peer) ensureConnected() {
	if p.live {
		return
	}
	for {
		conn, err := net.Dial("tcp", fmt.Sprintf("localhost:%d", p.port))
		if err == nil {
			p.failCount = 0
			p.conn = conn
			p.live = true
			return
		}

		p.failCount++
		log.Printf("dial failed. waiting %d seconds on port %d",
			p.failCount, p.port)
		time.Sleep(time.Duration(p.failCount) * time.Second)
	}
}

func (p *Peer) Send(message string) {
	p.ensureConnected()
	fmt.Fprintf(p.conn, message + "\n")
}

func NewPeer(port int) *Peer {
	return &Peer{port: port}
}
