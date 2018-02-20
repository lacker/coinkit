package network

import (
	"log"

	"coinkit/currency"
	"coinkit/util"
)

type Connection interface {
	Close()
	IsClosed() bool
	Send(message *util.SignedMessage) bool
	QuitChannel() chan bool
	Receive() *util.SignedMessage
}

// SendAnonymousMessage uses a new random key to send a single message.
func SendAnonymousMessage(c Connection, message *util.InfoMessage) {
	kp := util.NewKeyPair()
	sm := util.NewSignedMessage(kp, message)
	c.Send(sm)
}

// WaitToClear waits for the transaction with this sequence number to clear.
func WaitToClear(c Connection, user string, sequence uint32) *currency.Account {
	for {
		SendAnonymousMessage(c, &util.InfoMessage{Account: user})
		m := c.Receive().Message()
		accountMessage, ok := m.(*currency.AccountMessage)
		if !ok {
			continue
		}
		account := accountMessage.State[user]
		if account == nil {
			continue
		}
		if account.Sequence >= sequence {
			return account
		}
		log.Printf("waiting for slot %d", m.Slot())
		SendAnonymousMessage(c, &util.InfoMessage{I: m.Slot()})
	}
}
