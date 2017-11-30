package network

import "fmt"
import "log"
import "net"
import "time"

type Peer struct {
	port int
	conn net.Conn
	outbox chan string
}

// Retries until it is connected
// TODO: reconnect after a disconnect
func (p *Peer) sendForever() {
	// Connect
	failCount := 0
	for {
		conn, err := net.Dial("tcp", fmt.Sprintf("localhost:%d", p.port))
		if err == nil {
			failCount = 0
			p.conn = conn
			break
		}

		failCount++
		log.Printf("dial failed. waiting %d seconds on port %d",
			failCount, p.port)
		time.Sleep(time.Duration(failCount) * time.Second)
	}

	// Send from the queue
	for {
		message := <-p.outbox
		fmt.Fprintf(p.conn, message + "\n")
	}
}

func (p *Peer) Send(message string) {
	go p.blockingSend(message)
}

func (p *Peer) blockingSend(message string) {
	p.outbox <- message
}

func NewPeer(port int) *Peer {
	p := &Peer{port: port, outbox: make(chan string)}
	go p.sendForever();
	return p
}
