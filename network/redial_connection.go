package network

import (
	"sync"

	"coinkit/util"
)

// A RedialConnection is a Connection that will automatically redial when there
// is any connection failure that would normally close the
// connetion. You can close it yourself, though, and it will stay
// closed.
// Some messages (perhaps just one?) might get dropped during a reconnect.
type RedialConnection struct {
	conn     *Connection
	address  *Address
	handler  func(*util.SignedMessage)
	outbox   chan *util.SignedMessage
	quit     chan bool
	closed   bool
	quitOnce sync.Once
}

func NewRedialConnection(address *Address, handler func(*util.SignedMessage)) *RedialConnection {
	c := &RedialConnection{
		handler: handler,
		outbox:  make(chan *util.SignedMessage, 100),
		quit:    make(chan bool),
		closed:  false,
	}
	return c
}

// TODO: implement more functions
