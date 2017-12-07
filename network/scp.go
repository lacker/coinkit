package network

import (
)

// Stuff for implementing the Stellar Consensus Protocol. See:
// https://www.stellar.org/papers/stellar-consensus-protocol.pdf

// For now the network will just be tracking its own uptime.
type SlotValue struct {
	Uptime float64
}

type QuorumSlice struct {
	// Members is a list of public keys for nodes that occur in the quorum slice.
	// Typically includes ourselves.
	Members []string

	// The number of members we require for consensus, including ourselves.
	// The protocol can support other sorts of slices, like weighted or any wacky
	// thing, but for now we only do this simple "any k out of these n" voting.
	Threshold int
}

type NominateMessage interface {
	// What slot we are nominating values for
	Slot int

	Nominate []SlotValue
	Accept []SlotValue
	Slice QuorumSlice
}
