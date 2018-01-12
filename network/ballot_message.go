package network

import (
	"fmt"
)

// See page 23 of the protocol paper for a description of balloting.
type BallotMessage interface {
	QuorumSlice() QuorumSlice
	Phase() Phase
	MessageType() string
	Slot() int	

	// AcceptAsPrepared tells whether this message implies that the sender
	// accepts this ballot as prepared
	AcceptAsPrepared(n int, x SlotValue) bool

	// VoteToPrepare indicates whether this message is actively voting to prepare,
	// not whether some past message can be determined to have voted to prepare.
	VoteToPrepare(n int, x SlotValue) bool

	// AcceptCommit tells whether this message implies that the sender
	// accepts this commit
	AcceptAsCommitted(n int, x SlotValue) bool

	// VoteToCommit indicates whether this message is actively voting
	// to commit, not whether some past message can be determined to
	// have voted to commit
	VoteToCommit(n int, x SlotValue) bool

	// The highest ballot number this node is voting for
	// Used to decide when we should start going to a higher number
	BallotNumber() int

	// CouldEverVoteFor tells whether the node that sent this message
	// could ever have this ballot as its active ballot
	// TODO: what does this mean exactly for confirm and externalize?
	CouldEverVoteFor(n int, x SlotValue) bool
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

func (p Phase) String() string {
	switch p {
	case Invalid:
		return "Invalid"
	case Prepare:
		return "Prepare"
	case Confirm:
		return "Confirm"
	case Externalize:
		return "Externalize"
	default:
		panic(fmt.Sprintf("unknown phase: %+v", p))
	}
}

type Ballot struct {
	// An increasing counter, n >= 1, to ensure we can always have more ballots
	n int

	// The value this ballot proposes
	x SlotValue
}

func Compatible(ballot1 Ballot, ballot2 Ballot) bool {
	return Equal(ballot1.x, ballot2.x)
}

// Whether accepting a as prepared implies b is accepted as prepared
func gtecompat(a *Ballot, b *Ballot) bool {
	if a == nil || b == nil {
		return false
	}
	if a.n < b.n {
		return false
	}
	return Equal(a.x, b.x)
}

// Whether accepting a as prepared implies accepting b is aborted
func gteincompat(a *Ballot, b *Ballot) bool {
	if a == nil || b == nil {
		return false
	}
	if a.n < b.n {
		return false
	}
	return !Equal(a.x, b.x)
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

func (m *PrepareMessage) AcceptAsCommitted(n int, x SlotValue) bool {
	return false
}

func (m *PrepareMessage) VoteToCommit(n int, x SlotValue) bool {
	if m.Cn == 0 || m.Hn == 0 || !Equal(m.Bx, x) {
		return false
	}
	return m.Cn <= n && n <= m.Hn
}

func (m *PrepareMessage) CouldEverVoteFor(n int, x SlotValue) bool {
	if m.Bn > n {
		// Ballots don't go backwards
		return false
	}
	if m.Bn == n && !Equal(m.Bx, x) {
		// This message is currently voting *against*
		return false
	}
	return true
}

func (m *PrepareMessage) BallotNumber() int {
	return m.Bn
}

func (m *PrepareMessage) Slot() int {
	return m.I
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

	// The value that we are accepting a commit for.
	X SlotValue

	// state.p.n
	Pn int

	// The range of ballot numbers we accept a commit for.
	Cn int
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
	return Equal(m.X, x)
}

func (m *ConfirmMessage) VoteToPrepare(n int, x SlotValue) bool {
	return false
}

func (m *ConfirmMessage) AcceptAsCommitted(n int, x SlotValue) bool {
	return Equal(m.X, x) && m.Cn <= n && n <= m.Hn
}

func (m *ConfirmMessage) VoteToCommit(n int, x SlotValue) bool {
	return Equal(m.X, x)
}

func (m *ConfirmMessage) CouldEverVoteFor(n int, x SlotValue) bool {
	return Equal(x, m.X)
}

func (m *ConfirmMessage) BallotNumber() int {
	return m.Hn
}

func (m *ConfirmMessage) Slot() int {
	return m.I
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

	// The range of ballot numbers we confirm a commit for
	Cn int
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

func (m *ExternalizeMessage) AcceptAsCommitted(n int, x SlotValue) bool {
	return Equal(x, m.X) && m.Cn <= n && n <= m.Hn
}

func (m *ExternalizeMessage) VoteToCommit(n int, x SlotValue) bool {
	return false
}

func (m *ExternalizeMessage) CouldEverVoteFor(n int, x SlotValue) bool {
	return Equal(x, m.X)
}

func (m *ExternalizeMessage) BallotNumber() int {
	return m.Hn
}

func (m *ExternalizeMessage) Slot() int {
	return m.I
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
