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
	queue    chan *Request
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
		request := <-p.queue
		if request.Message == nil {
			log.Fatal("do not send nil messages through clients")
		}
		for {
			p.connect()
			util.WriteSignedMessage(p.conn, request.Message)

			// If we get an ok, great.
			// If we don't get an ok, disconnect and try again.
			p.conn.SetReadDeadline(time.Now().Add(5 * time.Second))
			response, err := util.ReadSignedMessage(p.conn)

			if err != nil {
				log.Printf("bad response from port %d: %+v", p.port, err)
				p.disconnect()
				continue
			}

			if request.Response != nil {
				request.Response <- response
			}
		}
	}
}

func (c *Client) Send(r *Request) {
	for {
		// Add to the queue if we can
		select {
		case c.queue <- r:
			return
		default:
			// The queue filled up
		}

		// Pop something off the queue to be discarded if we can
		select {
		case _ = <-c.queue:
		default:
			// There must be some racing. Just busy-add
		}
	}
}

// NewClient constructs a new client by connecting to the given port.
func NewClient(port int) *Client {
	log.Printf("connecting to node at port %d", port)
	// queue has a buffer of buflen outgoing messages
	buflen := 1
	p := &Client{
		port: port,
		queue: make(chan *Request, buflen),
	}
	go p.sendForever()
	return p
}
