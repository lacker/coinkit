package util

import (
	"fmt"
	"strings"
)

// An InfoMessage is sent by a client that wishes to know information. It doesn't
// indicate any statement being made by the sender. The node-to-node protocol
// does not require InfoMessages so this is typically just sent by endpoint clients.
type InfoMessage struct {
	// When Account is nonempty, this message is requesting the account data for
	// this particular user.
	Account string
}

func (m *InfoMessage) Slot() int {
	return 0
}

func (m *InfoMessage) MessageType() string {
	return "Info"
}

func (m *InfoMessage) String() string {
	parts := []string{"info"}
	if m.Account != "" {
		parts = append(parts, fmt.Sprintf("account=%s", Shorten(m.Account)))
	}
	return strings.Join(parts, " ")
}

func init() {
	RegisterMessageType(&InfoMessage{})
}
