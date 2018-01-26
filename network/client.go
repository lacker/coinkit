package network

import (
	"fmt"
	"log"
	"net"
	"time"

	"coinkit/util"
)

// A Client is a network connection established to a Server.
// It will keep redialing even after disconnects.
type Client struct {
	port      int
	conn      net.Conn
	outbox    chan *util.SignedMessage
	connected bool
}

// connect is idempotent
func (p *Client) connect() {
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
		// log.Printf("dial failed. waiting %d seconds on port %d", failCount, p.port)
		time.Sleep(time.Duration(failCount) * time.Second)
	}
}

func (p *Client) disconnect() {
	if p.conn != nil {
		p.conn.Close()
	}
	p.connected = false
}

// sendForever should handle disconnects or unresponsive peers.
func (p *Client) sendForever() {
	// Send from the queue
	for {
		message := <-p.outbox
		for {
			p.connect()
			message.WriteTo(p.conn)

			// If we get an ok, great.
			// If we don't get an ok, disconnect and try again.
			p.conn.SetReadDeadline(time.Now().Add(5 * time.Second))
			_, err := util.ReadSignedMessage(p.conn)

			if err != nil {
				log.Printf("bad response from port %d: %+v", p.port, err)
				p.disconnect()
				continue
			}

			// TODO: handle the response
		}
	}
}

func (p *Client) Send(message *util.SignedMessage) {
	for {
		// Add to the outbox if we can
		select {
		case p.outbox <- message:
			return
		default:
			// The queue filled up
		}

		// Pop something off the outbox to be discarded if we can
		select {
		case _ = <-p.outbox:
		default:
			// There must be some racing. Just busy-add
		}
	}
}

func NewClient(port int) *Client {
	log.Printf("connecting to peer at port %d", port)
	// outbox has a buffer of buflen outgoing messages
	buflen := 1
	p := &Client{port: port, outbox: make(chan *util.SignedMessage, buflen)}
	go p.sendForever()
	return p
}
