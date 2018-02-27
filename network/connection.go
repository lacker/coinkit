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
	Receive() chan *util.SignedMessage
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
		m := (<-c.Receive()).Message
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

		SendAnonymousMessage(c, &util.InfoMessage{I: m.Slot()})
		<-c.Receive()
	}
}

func GetAccount(c Connection, user string) *currency.Account {
	for {
		SendAnonymousMessage(c, &util.InfoMessage{Account: user})
		m := (<-c.Receive()).Message
		accountMessage, ok := m.(*currency.AccountMessage)
		if !ok {
			log.Fatalf("expected an account message but got: %+v", m)
		}
		return accountMessage.State[user]
	}
}

func recHelper(inbox chan *util.SignedMessage, quit chan bool) chan *util.SignedMessage {
	answer := make(chan *util.SignedMessage)
	go func() {
		select {
		case m := <-inbox:
			answer <- m
		case <-quit:
			answer <- nil
		}
	}()
	return answer
}
