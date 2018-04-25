package util

import (
	"fmt"
	"strings"
)

// An InfoMessage is sent by a client that wishes to know information. It doesn't
// indicate any statement being made by the sender. The node-to-node protocol
// does not require InfoMessages so this is typically just sent by endpoint clients.
type InfoMessage struct {
	// When Block is nonzero, this message is requesting an ExternalizeMessage
	// containing the block for a particular slot.
	// If the block being requested is the next one, the server may optionally
	// wait a little while to send the block once it's finalized.
	Block int

	// When Account is nonempty, this message is requesting the account data for
	// this particular user.
	Account string
}

func (m *InfoMessage) Slot() int {
	return m.Block
}

func (m *InfoMessage) MessageType() string {
	return "Info"
}

func (m *InfoMessage) String() string {
	parts := []string{"info"}
	if m.Block != 0 {
		parts = append(parts, fmt.Sprintf("block=%d", m.Block))
	}
	if m.Account != "" {
		parts = append(parts, fmt.Sprintf("account=%s", Shorten(m.Account)))
	}
	return strings.Join(parts, " ")
}

func init() {
	RegisterMessageType(&InfoMessage{})
}
