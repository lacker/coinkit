package data

import (
	"fmt"
	"strings"

	"github.com/lacker/coinkit/util"
)

// An AccountMessage is used for sharing information about the state of
// accounts. Currently this is client-server rather than peer-peer.
// The client sends a blank AccountMessage missing the information it would
// like to know, and the server sends one back.
type AccountMessage struct {
	// The active slot when this message was created.
	// 0 means it is unknown.
	I int

	// The state of accounts as of the provided slot.
	// Nil values mean it is unknown.
	State map[string]*Account
}

func (m *AccountMessage) Slot() int {
	return m.I
}

func (m *AccountMessage) MessageType() string {
	return "Account"
}

func (m *AccountMessage) String() string {
	parts := []string{"account"}
	if m.I != 0 {
		parts = append(parts, fmt.Sprintf("i=%d", m.I))
	}
	for user, account := range m.State {
		parts = append(parts, fmt.Sprintf("%s=%s",
			util.Shorten(user), StringifyAccount(account)))
	}
	return strings.Join(parts, " ")
}

func init() {
	util.RegisterMessageType(&AccountMessage{})
}
