package network

import (
	"fmt"

	"coinkit/consensus"
	"coinkit/currency"
	"coinkit/util"
)

// A HistoryMessage is sent when the other node is far behind and needs to catch up
// to the current state.

type HistoryMessage struct {
	I int
	T *currency.TransactionMessage
	E *consensus.ExternalizeMessage
}

func (m *HistoryMessage) Slot() int {
	return m.I
}

func (m *HistoryMessage) MessageType() string {
	return "H"
}

func (m *HistoryMessage) String() string {
	return fmt.Sprintf("history i=%d: %s %s", m.I, m.T, m.E)
}

func init() {
	util.RegisterMessageType(&HistoryMessage{})
}
