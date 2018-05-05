package data

import (
	"fmt"
	"strings"

	"github.com/lacker/coinkit/util"
)

// A QueryMessage is sent by a client that wishes to know information. It doesn't
// indicate any statement being made by the sender.
// Only a single top-level field should be filled in for a single QueryMessage.
// The response should be a DataMessage.
type QueryMessage struct {
	// When Account is nonempty, this message is requesting the account data for
	// this particular user.
	Account string

	// When Block is nonzero, this message is requesting data for a mined block.
	Block int
}

func (m *QueryMessage) Slot() int {
	return 0
}

func (m *QueryMessage) MessageType() string {
	return "Query"
}

func (m *QueryMessage) String() string {
	parts := []string{"info"}
	if m.Account != "" {
		parts = append(parts, fmt.Sprintf("account=%s", util.Shorten(m.Account)))
	}
	if m.Block != 0 {
		parts = append(parts, fmt.Sprintf("block=%d", m.Block))
	}
	return strings.Join(parts, " ")
}

func init() {
	util.RegisterMessageType(&QueryMessage{})
}
