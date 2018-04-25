package data

import (
	"fmt"
	"strings"

	"github.com/lacker/coinkit/util"
)

// A DataMessage is used to respond when a client requests data.
type DataMessage struct {
	// I is the last finalized slot occurring in the data snapshot used for this data.
	I int

	// The contents of an account, keyed by owner.
	// A nil value means there is no account for the owner key.
	Accounts map[string]*Account

	// The contents of some blocks, keyed by slot.
	// Nil values mean that the block is unknown because it has not been finalized yet.
	Blocks map[int]*Block
}

func (m *DataMessage) Slot() int {
	return m.I
}

func (m *DataMessage) MessageType() string {
	return "Data"
}

func (m *DataMessage) String() string {
	parts := []string{"data", fmt.Sprintf("slot=%d", m.Slot())}
	for owner, account := range m.Accounts {
		parts = append(parts, fmt.Sprintf("a:%s=%s",
			util.Shorten(owner), StringifyAccount(account)))
	}
	for i, _ := range m.Blocks {
		parts = append(parts, fmt.Sprintf("block%d", i))
	}
	return strings.Join(parts, " ")
}

func init() {
	util.RegisterMessageType(&DataMessage{})
}
