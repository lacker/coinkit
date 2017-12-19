package network

import (
)

// See page 23 of the protocol paper for a description of balloting.
type BallotMessage interface {
	QuorumSlice() QuorumSlice
	Phase() Phase
	MessageType() string
}

// Ballot phases
// Invalid is 0 so that if we inadvertently create a new message the wrong way and
// leave things zeroed it will be obviously an invalid phase
type Phase int
const (
	Invalid Phase = iota
	Prepare
	Confirm
	Externalize
)

type Ballot struct {
	// An increasing counter, n >= 1, to ensure we can always have more ballots
	n int

	// The value this ballot proposes
	x SlotValue
}

func Compatible(ballot1 Ballot, ballot2 Ballot) bool {
	return Equal(ballot1.x, ballot2.x)
}

// PrepareMessage is the first phase of the three-phase ballot protocol
type PrepareMessage struct {
	// T is Prepare for a PrepareMessage
	T Phase
	
	// What slot we are preparing ballots for
	I int

	// The current ballot we are trying to prepare
	Bn int
	Bx SlotValue

	// The contents of state.p
	Pn int
	Px SlotValue

	// The contents of state.pPrime
	Ppn int
	Ppx SlotValue

	// Ballot numbers for c and h
	Cn int
	Hn int

	D QuorumSlice
}

func (m *PrepareMessage) QuorumSlice() QuorumSlice {
	return m.D
}

func (m *PrepareMessage) Phase() Phase {
	if m.T != Prepare {
		panic("m.T != Prepare")
	}
	return Prepare
}

func (m *PrepareMessage) MessageType() string {
	return "Prepare"
}

// ConfirmMessage is the second phase of the three-phase ballot protocol
type ConfirmMessage struct {
	// T is Confirm for a ConfirmMessage
	T Phase
	
	// What slot we are confirming ballots for
	I int

	// The current ballot we are trying to confirm
	Bn int
	Bx SlotValue

	// state.p.n
	Pn int

	// state.c.n
	Cn int

	// state.h.n
	Hn int

	D QuorumSlice
}

func (m *ConfirmMessage) QuorumSlice() QuorumSlice {
	return m.D
}

func (m *ConfirmMessage) Phase() Phase {
	if m.T != Confirm {
		panic("m.T != Confirm")
	}
	return Confirm
}

func (m *ConfirmMessage) MessageType() string {
	return "Confirm"
}

// ExternalizeMessage is the third phase of the three-phase ballot protocol
// Sent after we have confirmed a commit
type ExternalizeMessage struct {
	// T is Externalize for an ExternalizeMessage
	T Phase
	
	// What slot we are externalizing
	I int

	// The value at this slot
	X SlotValue

	// state.c.n
	Cn int

	// state.h.n
	Hn int

	D QuorumSlice
}

func (m *ExternalizeMessage) QuorumSlice() QuorumSlice {
	return m.D
}

func (m *ExternalizeMessage) Phase() Phase {
	if m.T != Externalize {
		panic("m.T != Externalize")
	}
	return Externalize
}

func (m *ExternalizeMessage) MessageType() string {
	return "Externalize"
}

// Compare returns -1 if ballot1 < ballot2
// 0 if ballot1 == ballot2
// 1 if ballot1 > ballot2
// Ballots are ordered by:
// (phase, b, p, p prime, h)
// This is only intended to be used to compare messages coming from the same node.
func Compare(ballot1 BallotMessage, ballot2 BallotMessage) int {
	phase1 := ballot1.Phase()
	phase2 := ballot2.Phase()
	if phase1 < phase2 {
		return -1
	}
	if phase1 > phase2 {
		return 1
	}
	switch b1 := ballot1.(type) {
	case *PrepareMessage:
		b2 := ballot2.(*PrepareMessage)
		if b1.Bn < b2.Bn {
			return -1
		}
		if b1.Bn > b2.Bn {
			return 1
		}
		if b1.Pn < b2.Pn {
			return -1
		}
		if b1.Pn > b2.Pn {
			return 1
		}
		if b1.Ppn < b2.Ppn {
			return -1
		}
		if b1.Ppn > b2.Ppn {
			return 1
		}
		if b1.Hn < b2.Hn {
			return -1
		}
		if b1.Hn > b2.Hn {
			return 1
		}
		return 0
	case *ConfirmMessage:
		b2 := ballot2.(*ConfirmMessage)
		if b1.Bn < b2.Bn {
			return -1
		}
		if b1.Bn > b2.Bn {
			return 1
		}
		if b1.Pn < b2.Pn {
			return -1
		}
		if b1.Pn > b2.Pn {
			return 1
		}
		if b1.Hn < b2.Hn {
			return -1
		}
		if b1.Hn > b2.Hn {
			return 1
		}
		return 0		
	case *ExternalizeMessage:
		b2 := ballot2.(*ExternalizeMessage)
		if b1.Hn < b2.Hn {
			return -1
		}
		if b1.Hn > b2.Hn {
			return 1
		}
		return 0		
	default:
		panic("programming error")
	}
}
