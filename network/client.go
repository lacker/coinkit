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

	// We set closing to true and close the quit channel when the
	// client is closing
	closing bool
	quit    chan bool
}

// connect is idempotent
func (c *Client) connect() {
	if c.connected || c.closing {
		return
	}
	failCount := 0
	for {
		conn, err := net.Dial("tcp", c.address.String())
		if err == nil {
			if c.conn != nil {
				c.conn.Close()
			}
			c.conn = conn
			c.connected = true
			return
		}

		failCount++
		timer := time.NewTimer(time.Duration(failCount) * time.Second)
		select {
		case <-c.quit:
			return
		case <-timer.C:
			// Looping again will try to reconnect
		}
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
		var request *Request
		select {
		case <-c.quit:
			return
		case request = <-c.queue:
		}

		if request.Timeout == 0 {
			log.Fatalf("you should use a timeout with clients")
		}
		line := request.GetLine()
		if len(line) == 0 {
			log.Fatalf("cannot send line: [%s]", line)
		}

		for {
			c.connect()
			if c.closing {
				return
			}
			fmt.Fprintf(c.conn, line)

			// If we get an ok, great.
			// If we don't get an ok, disconnect and try again.
			c.conn.SetReadDeadline(time.Now().Add(request.Timeout))
			response, err := util.ReadSignedMessage(c.conn)

			if c.closing {
				return
			}

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

// Close() should be called when the client is no longer being used. Requests in
// progress may or may not have callbacks called. This is important to do so that
// we don't have eternal redials from clients that are no longer in use.
func (c *Client) Close() {
	c.closing = true
	close(c.quit)
	if c.conn != nil {
		c.conn.Close()
	}
}

// Send issues a request and will send the response to the response channel.
func (c *Client) Send(r *Request) {
	for {
		// Add to the queue if we can
		select {
		case c.queue <- r:
			return
		case <-c.quit:
			return
		default:
			// The queue filled up
		}

		// Pop something off the queue to be discarded if we can
		select {
		case <-c.queue:
			log.Printf("send queue overloaded, dropping message")
		case <-c.quit:
			return
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

// NewClient connects to the Server at the given address.
func NewClient(address *Address) *Client {
	// queue has a buffer of buflen outgoing messages
	buflen := 10
	p := &Client{
		address: address,
		queue:   make(chan *Request, buflen),
		closing: false,
		quit:    make(chan bool),
	}
	go p.sendForever()
	return p
}

func (c *Client) SendInfoMessage(message *util.InfoMessage) util.Message {
	// We can use an anonymous key with info messages
	kp := util.NewKeyPair()
	sm := util.NewSignedMessage(kp, message)
	response := c.SendMessage(sm)
	if response == nil {
		log.Fatal("got nil account message")
	}
	return response.Message()
}

// WaitToClear waits for the transaction with this sequence number to clear.
func (c *Client) WaitToClear(user string, sequence uint32) *currency.Account {
	for {
		m := c.SendInfoMessage(&util.InfoMessage{Account: user})
		account := m.(*currency.AccountMessage).State[user]
		if account.Sequence >= sequence {
			return account
		}
		log.Printf("waiting for slot %d", m.Slot())
		c.SendInfoMessage(&util.InfoMessage{I: m.Slot()})
	}
}

func (c *Client) GetAccount(user string) *currency.Account {
	m := c.SendInfoMessage(&util.InfoMessage{Account: user}).(*currency.AccountMessage)
	return m.State[user]
}
