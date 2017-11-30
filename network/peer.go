package network

import "bufio"
import "fmt"
import "log"
import "net"
import "time"

type Peer struct {
	port int
	conn net.Conn
	outbox chan string
	connected bool
}

// connect is idempotent
func (p *Peer) connect() {
	if p.connected {
		return
	}
	failCount := 0
	for {
		conn, err := net.Dial("tcp", fmt.Sprintf("localhost:%d", p.port))
		if err == nil {
			p.conn = conn
			p.connected = true
			return
		}

		failCount++
		log.Printf("dial failed. waiting %d seconds on port %d",
			failCount, p.port)
		time.Sleep(time.Duration(failCount) * time.Second)
	}
}

func (p *Peer) disconnect() {
	if p.conn != nil {
		p.conn.Close()
	}
	p.connected = false
}

// sendForever should handle disconnects or unresponsive peers.
func (p *Peer) sendForever() {
	// Send from the queue
	for {
		message := <-p.outbox
		for {
			p.connect()
			log.Printf("sending message: %s", message)
			fmt.Fprintf(p.conn, message + "\n")

			// If we get an ok, great.
			// If we don't get an ok, disconnect and try again.
			p.conn.SetReadDeadline(time.Now().Add(5 * time.Second))
			_, err := bufio.NewReader(p.conn).ReadString('\n')
			if err == nil {
				break
			}
			log.Print("did not receive an ok: ", err)
			p.disconnect()
		}
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
