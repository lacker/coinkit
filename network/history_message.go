package network

import (
	"fmt"

	"github.com/lacker/coinkit/consensus"
	"github.com/lacker/coinkit/data"
	"github.com/lacker/coinkit/util"
)

// A HistoryMessage is sent when the other node is far behind and needs to catch up
// to the current state.

type HistoryMessage struct {
	I int
	O *data.OperationMessage
	E *consensus.ExternalizeMessage
}

func (m *HistoryMessage) Slot() int {
	return m.I
}

func (m *HistoryMessage) MessageType() string {
	return "History"
}

func (m *HistoryMessage) String() string {
	return fmt.Sprintf("history i=%d: %s %s", m.I, m.O, m.E)
}

func init() {
	util.RegisterMessageType(&HistoryMessage{})
}
