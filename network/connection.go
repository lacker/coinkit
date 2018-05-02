package network

import (
	"time"

	"github.com/lacker/coinkit/data"
	"github.com/lacker/coinkit/util"
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
	sm := util.NewSignedMessage(message, kp)
	c.Send(sm)
}

// WaitToClear waits for the operation with this account + sequence number to clear.
func WaitToClear(c Connection, user string, sequence uint32) *data.Account {
	for {
		start := time.Now()
		util.Logger.Printf("XXX checking account info at %s", start)
		SendAnonymousMessage(c, &util.InfoMessage{Account: user})
		m := (<-c.Receive()).Message()
		elapsed := time.Now().Sub(start).Seconds()
		if elapsed > 1.0 {
			util.Logger.Printf("warning: server took %.2fs to get account info", elapsed)
		}
		dataMessage, ok := m.(*data.DataMessage)
		if ok {
			account := dataMessage.Accounts[user]
			if account != nil && account.Sequence >= sequence {
				return account
			}
		}

		util.Logger.Printf("XXX sleeping for a bit at %s", time.Now())
		time.Sleep(time.Millisecond * 100)
	}
}

func GetAccount(c Connection, user string) *data.Account {
	for {
		SendAnonymousMessage(c, &util.InfoMessage{Account: user})
		m := (<-c.Receive()).Message()
		dataMessage, ok := m.(*data.DataMessage)
		if !ok {
			util.Logger.Fatalf("expected a data message but got: %+v", m)
		}
		return dataMessage.Accounts[user]
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
