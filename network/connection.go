package network

import (
	"fmt"
	"log"
	"net"
	"sync"
	"time"

	"coinkit/util"
)

// How frequently in seconds to send keepalive pings
const keepalive = 10

// A Connection represents a two-way message channel.
// You can close it at any point, and it will close itself if it detects
// network problems.
type Connection struct {
	conn     net.Conn
	handler  func(*util.SignedMessage)
	outbox   chan *util.SignedMessage
	quit     chan bool
	closed   bool
	quitOnce sync.Once
}

// NewConnection creates a new logical connection given a network connection.
// handler will get called on all non-nil incoming messages.
func NewConnection(conn net.Conn, handler func(*util.SignedMessage)) *Connection {
	c := &Connection{
		conn:    conn,
		handler: handler,
		outbox:  make(chan *util.SignedMessage, 100),
		quit:    make(chan bool),
		closed:  false,
	}
	go c.runIncoming()
	go c.runOutgoing()
	return c
}

func (c *Connection) Close() {
	c.quitOnce.Do(func() {
		c.closed = true
		close(c.quit)
	})
}

func (c *Connection) IsClosed() bool {
	return c.closed
}

func (c *Connection) runIncoming() {
	for {
		// Wait for 2x the keepalive period
		c.conn.SetReadDeadline(time.Now().Add(2 * keepalive * time.Second))
		response, err := util.ReadSignedMessage(c.conn)
		if c.closed {
			break
		}
		if err != nil {
			log.Printf("connection error: %+c", err)
			c.Close()
			break
		}
		if response != nil {
			c.handler(response)
		}
	}
	c.handler = nil
}

func (c *Connection) runOutgoing() {
	for {
		var message *util.SignedMessage
		timer := time.NewTimer(time.Duration(keepalive * time.Second))
		select {
		case <-c.quit:
			return
		case <-timer.C:
			// Send a keepalive ping
			message = nil
		case message = <-c.outbox:
		}

		fmt.Fprintf(c.conn, util.SignedMessageToLine(message))
	}
}

// Send sends a message, but only if the queue is not full.
// It returns whether the message entered the outbox.
func (c *Connection) Send(message *util.SignedMessage) bool {
	select {
	case c.outbox <- message:
		return true
	default:
		log.Printf("Connection outbox overloaded, dropping message")
		return false
	}
}

// QuitChannel returns a channel that gets closed once, when the channel shuts down.
func (c *Connection) QuitChannel() chan bool {
	return c.quit
}
