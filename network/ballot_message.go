package network

import (
)

// See page 23 of the protocol paper for a description of balloting.
type BallotMessage interface {
	QuorumSlice() QuorumSlice
	Phase() Phase
	MessageType() string

	// AcceptAsPrepared tells whether this message implies that the sender
	// accepts this ballot as prepared
	AcceptAsPrepared(n int, x SlotValue) bool

	// VoteToPrepare indicates whether this message is actively voting to prepare,
	// not whether some past message can be determined to have voted to prepare.
	VoteToPrepare(n int, x SlotValue) bool
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

// PrepareMessage is the first phase of the three-phase ballot protocol.
// This message is preparing a ballot.
// To prepare is to abort any conflicting ballots.
// This message is voting to prepare b, and also informing the receiver that
// we have accepted that both p and p prime have already been prepared.
type PrepareMessage struct {
	// T is Prepare for a PrepareMessage
	T Phase
	
	// What slot we are preparing ballots for
	I int

	// The ballot we are voting to prepare
	Bn int
	Bx SlotValue

	// The contents of state.p, which we accept as prepared
	Pn int
	Px SlotValue

	// The contents of state.pPrime, which we accept as prepared
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

func (m *PrepareMessage) AcceptAsPrepared(n int, x SlotValue) bool {
	// A prepare message accepts that both p and p prime are prepared.
	if Equal(m.Px, x) {
		return m.Pn >= n
	}
	if Equal(m.Ppx, x) {
		return m.Ppn >= n
	}
	return false
}

func (m *PrepareMessage) VoteToPrepare(n int, x SlotValue) bool {
	return Equal(x, m.Bx) && m.Bn >= n
}

// ConfirmMessage is the second phase of the three-phase ballot protocol
// "Confirm" seems like a bad name for this phase, it seems like it should be
// named "Commit". Because you are also confirming as part of nominate and prepare.
// I stuck with "Confirm" because that's what the paper calls it.
// Intuitively (sic), a confirm message is accepting a commit.
// The consensus can still get borked at this phase if we don't get a
// quorum confirming.
type ConfirmMessage struct {
	// T is Confirm for a ConfirmMessage
	T Phase
	
	// What slot we are confirming ballots for
	I int

	// The current ballot that we are accepting a commit for.
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

func (m *ConfirmMessage) AcceptAsPrepared(n int, x SlotValue) bool {
	return Equal(m.Bx, x)
}

func (m *ConfirmMessage) VoteToPrepare(n int, x SlotValue) bool {
	return false
}

// ExternalizeMessage is the third phase of the three-phase ballot protocol
// Sent after we have confirmed a commit.
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

func (m *ExternalizeMessage) AcceptAsPrepared(n int, x SlotValue) bool {
	return Equal(m.X, x)
}

func (m *ExternalizeMessage) VoteToPrepare(n int, x SlotValue) bool {
	return false
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
