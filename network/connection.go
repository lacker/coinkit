package network

import (
	"coinkit/util"
)

type Connection interface {
	Close()
	IsClosed() bool
	Send(message *util.SignedMessage) bool
	QuitChannel() chan bool
}
