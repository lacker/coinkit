package network

import (
)

// Stuff for implementing the Stellar Consensus Protocol. See:
// https://www.stellar.org/papers/stellar-consensus-protocol.pdf

// For now each block just has a list of comments.
// This isn't supposed to be useful, it's just for testing.
type SlotValue struct {
	Comments []string
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

type NominateMessage struct {
	// What slot we are nominating values for
	Slot int

	Nominate []SlotValue
	Accept []SlotValue
	Slice QuorumSlice
}

func (m *NominateMessage) MessageType() string {
	return "Nominate"
}

type NominationState struct {
	nominated []SlotValue
	accepted []SlotValue
	candidates []SlotValue

	// The last NominateMessage received from each node
	received map[string]*NominateMessage
}

// Ballot phases
type Phase int
const (
	Prepare Phase = iota
	Confirm
	Externalize
)

type Ballot struct {
	counter int
	value SlotValue
}

type BallotMessage struct {
	// TODO
}

type BallotState struct {
	// The current ballot we are trying to prepare and commit.
	// Lowest valid ballot is 1.
	current int

	// The highest two ballots that are accepted as prepared.
	// 0 means there is no such accepted ballot
	highestAccepted int
	nextHighestAccepted int

	// TODO: more stuff from pg 23
}

type StateBuilder struct {
	// Which slot is actively being built
	slot int

	// Values for past slots that have already achieved consensus
	map[int]SlotValue values

	NominationState nomination
	BallotState ballot
}

func NewStateBuilder() *StateBuilder {
	// TODO,
}

