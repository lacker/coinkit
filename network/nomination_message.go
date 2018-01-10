package network

import (
)

// The nomination message format of the Stellar Consensus Protocol.
// Implements Message.
// See:
// https://www.stellar.org/papers/stellar-consensus-protocol.pdf
type NominationMessage struct {
	// What slot we are nominating values for
	I int

	// The values we have voted to nominate
	Nom []SlotValue

	// The values we have accepted as nominated
	Acc []SlotValue

	D QuorumSlice
}

func (m *NominationMessage) MessageType() string {
	return "Nomination"
}

func (m *NominationMessage) Slot() int {
	return m.I
}
