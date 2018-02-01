package network

import (
	"fmt"
	"log"
	"net"
	"time"

	"coinkit/currency"
	"coinkit/util"
)

// A Client is a network connection established to a Server.
// It will keep redialing even after disconnects.
type Client struct {
	address   *Address
	conn      net.Conn
	queue     chan *Request
	connected bool
}

// connect is idempotent
func (c *Client) connect() {
	if c.connected {
		return
	}
	failCount := 0
	for {
		conn, err := net.Dial("tcp", c.address.String())
		if err == nil {
			c.conn = conn
			c.connected = true
			return
		}

		failCount++
		time.Sleep(time.Duration(failCount) * time.Second)
	}
}

func (c *Client) disconnect() {
	if c.conn != nil {
		c.conn.Close()
	}
	c.connected = false
}

// sendForever should handle disconnects or unresponsive peers.
func (c *Client) sendForever() {
	// Send from the queue
	for {
		request := <-c.queue
		if request.Timeout == 0 {
			log.Fatalf("you should use a timeout with clients")
		}
		line := request.GetLine()
		if len(line) == 0 {
			log.Fatalf("cannot send line: [%s]", line)
		}
		for {
			c.connect()
			fmt.Fprintf(c.conn, line)

			// If we get an ok, great.
			// If we don't get an ok, disconnect and try again.
			c.conn.SetReadDeadline(time.Now().Add(request.Timeout))
			response, err := util.ReadSignedMessage(c.conn)

			if err != nil {
				log.Printf("bad response from %s: %+v", c.address.String(), err)
				c.disconnect()
				continue
			}

			if request.Response != nil {
				request.Response <- response
			}
			break
		}
	}
}

// Send issues a request and will send the response to the response channel.
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
		case <-c.queue:
			log.Printf("send queue overloaded, dropping message")
		default:
			// There must be some racing. Wait a bit and try again
			time.Sleep(time.Millisecond)
		}
	}
}

// Sends a signed message and waits for the response.
func (c *Client) SendMessage(message *util.SignedMessage) *util.SignedMessage {
	response := make(chan *util.SignedMessage)
	request := &Request{
		Message:  message,
		Response: response,
		Timeout:  5 * time.Second,
	}
	// Wait on a response.
	// This hangs on network failure
	c.Send(request)
	sm := <-response
	return sm
}

// GetAccount fetches account data.
// Hangs on network failure, can log fatal on a malicious server.
func (c *Client) GetAccount(user string) *currency.Account {
	// Since this is public data we'll use a throwaway key and stay anonymous
	kp := util.NewKeyPair()

	message := currency.NewInquiryMessage(user)
	sm := util.NewSignedMessage(kp, message)
	response := c.SendMessage(sm)
	m := response.Message()
	am, ok := m.(*currency.AccountMessage)
	if !ok {
		log.Fatal("received non-account message: %+v", message)
	}
	return am.State[user]
}

// NewClient connects to the Server at the given address.
func NewClient(address *Address) *Client {
	// queue has a buffer of buflen outgoing messages
	buflen := 10
	p := &Client{
		address: address,
		queue:   make(chan *Request, buflen),
	}
	go p.sendForever()
	return p
}
