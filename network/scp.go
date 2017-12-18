package network

import (
	"fmt"
	"log"
	"sort"
	"strings"
	"time"
)

// Stuff for implementing the Stellar Consensus Protocol. See:
// https://www.stellar.org/papers/stellar-consensus-protocol.pdf
// When there are frustrating single-letter variable names, it's because we are
// making the names line up with the protocol paper.

// For now each block just has a list of comments.
// This isn't supposed to be useful, it's just for testing.
type SlotValue struct {
	Comments []string
}

func MakeSlotValue(comment string) SlotValue {
	return SlotValue{Comments: []string{comment}}
}

// Combine is specific to what the slot values are
func Combine(a SlotValue, b SlotValue) SlotValue {
	joined := append(a.Comments, b.Comments...)
	sort.Strings(joined)
	answer := []string{}
	for _, item := range joined {
		if len(answer) == 0 || answer[len(answer)-1] != item {
			answer = append(answer, item)
		}
	}
	return SlotValue{Comments: answer}
}

// CombineSlice just runs Combine on each value in a slice
func CombineSlice(list []SlotValue) SlotValue {
	if len(list) == 0 {
		panic("CombineSlice should not be called on empty slices")
	}
	if len(list) == 1 {
		return list[0]
	}
	mid := len(list) / 2
	return Combine(CombineSlice(list[:mid]), CombineSlice(list[mid:]))
}

func HasSlotValue(list []SlotValue, v SlotValue) bool {
	k := strings.Join(v.Comments, ",")
	for _, s := range list {
		if strings.Join(s.Comments, ",") == k {
			return true
		}
	}
	return false
}

type NominationMessage struct {
	// What slot we are nominating values for
	I int

	// The values we have voted to nominate
	X []SlotValue

	// The values we have accepted as nominated
	Y []SlotValue
	
	D QuorumSlice
}

func (m *NominationMessage) MessageType() string {
	return "Nomination"
}

// See page 21 of the protocol paper for more detail here.
type NominationState struct {
	// The values we have voted to nominate
	X []SlotValue

	// The values we have accepted as nominated
	Y []SlotValue

	// The values that we consider to be candidates 
	Z []SlotValue

	// The last NominationMessage received from each node
	N map[string]*NominationMessage

	// Who we are
	publicKey string

	// Who we listen to for quorum
	D QuorumSlice
}

func NewNominationState(publicKey string, qs QuorumSlice) *NominationState {
	return &NominationState{
		X: make([]SlotValue, 0),
		Y: make([]SlotValue, 0),
		Z: make([]SlotValue, 0),
		N: make(map[string]*NominationMessage),
		publicKey: publicKey,
		D: qs,
	}
}

// HasNomination tells you whether this nomination state can currently send out
// a nominate message.
// If we have never received a nomination from a peer, and haven't had SetDefault
// called ourselves, then we won't have a nomination.
func (s *NominationState) HasNomination() bool {
	return len(s.X) > 0
}

func (s *NominationState) SetDefault(v SlotValue) {
	if s.HasNomination() {
		// We already have something to nominate
		return
	}
	s.X = []SlotValue{v}
}

// PredictValue can predict the value iff HasNomination is true. If not, panic
func (s *NominationState) PredictValue() SlotValue {
	if len(s.Z) > 0 {
		return CombineSlice(s.Z)
	}
	if len(s.Y) > 0 {
		return CombineSlice(s.Y)
	}
	if len(s.X) > 0 {
		return CombineSlice(s.X)
	}
	panic("PredictValue was called when HasNomination was false")
}

func (s *NominationState) QuorumSlice(node string) (*QuorumSlice, bool) {
	if node == s.publicKey {
		return &s.D, true
	}
	m, ok := s.N[node]
	if !ok {
		return nil, false
	}
	return &m.D, true
}

func (s *NominationState) PublicKey() string {
	return s.publicKey
}

// MaybeAdvance checks whether we should accept the nomination for this slot value,
// and adds it to our accepted list if appropriate.
// It also checks whether we should confirm the nomination.
// Returns whether we made any changes.
func (s *NominationState) MaybeAdvance(v SlotValue) bool {
	if HasSlotValue(s.Z, v) {
		// We already confirmed this, so we can't do anything more
		return false
	}
	
	changed := false	
	votedOrAccepted := []string{}
	accepted := []string{}
	if HasSlotValue(s.X, v) {
		votedOrAccepted = append(votedOrAccepted, s.publicKey)
	}
	if HasSlotValue(s.Y, v) {
		accepted = append(accepted, s.publicKey)
	}
	for node, m := range s.N {
		if HasSlotValue(m.Y, v) {
			votedOrAccepted = append(votedOrAccepted, node)
			accepted = append(accepted, node)
			continue
		}
		if HasSlotValue(m.X, v) {
			votedOrAccepted = append(votedOrAccepted, node)
		}
	}

	// The rules for accepting are on page 13, section 5.3
	// Rule 1: if a quorum has either voted for the nomination or accepted the
	// nomination, we accept it.
	// Rule 2: if a blocking set for us accepts the nomination, we accept it.
	accept := MeetsQuorum(s, votedOrAccepted) || s.D.BlockedBy(accepted)

	if accept && !HasSlotValue(s.Y, v) {
		// Accept this value
		log.Printf("I accept the nomination of %+v", v)
		changed = true
		s.Y = append(s.Y, v)
	}

	// We confirm once a quorum has accepted
	if MeetsQuorum(s, accepted) {
		log.Printf("I confirm the nomination of %+v", v)		
		changed = true
		s.Z = append(s.Z, v)
	}
	return changed
}

// Handles an incoming nomination message from a peer node
func (s *NominationState) Handle(node string, m *NominationMessage) {
	// What nodes we have seen new information about
	touched := []SlotValue{}

	// Check if there's anything new
	old, ok := s.N[node]
	var oldLenX, oldLenY int
	if ok {
		oldLenX = len(old.X)
		oldLenY = len(old.Y)
	}
	if len(m.X) < oldLenX {
		log.Printf("node %s sent a stale message: %v", node, m)
		return
	}
	if len(m.Y) < oldLenY {
		log.Printf("node %s sent a stale message: %v", node, m)
		return
	}
	if len(m.X) == oldLenX && len(m.Y) == oldLenY {
		// It's just a dupe
		return
	}
	// Update our most-recent-message
	s.N[node] = m
	
	for i := oldLenX; i < len(m.X); i++ {
		if !HasSlotValue(touched, m.X[i]) {
			touched = append(touched, m.X[i])
		}

		// If we don't have a candidate, we can support this new nomination
		if !HasSlotValue(s.X, m.X[i]) {
			log.Printf("I support the nomination of %+v", m.X[i])
			s.X = append(s.X, m.X[i])
		}
	}

	for i := oldLenY; i < len(m.Y); i++ {
		if !HasSlotValue(touched, m.Y[i]) {
			touched = append(touched, m.Y[i])
		}
	}

	for _, v := range touched {
		s.MaybeAdvance(v)
	}
}

// Ballot phases
type Phase int
const (
	Prepare Phase = iota
	Confirm
	Externalize
)

type Ballot struct {
	// An increasing counter, n >= 1, to ensure we can always have more ballots
	n int

	// The value this ballot proposes
	x SlotValue
}

// See page 23 of the protocol paper for more detail here.
type BallotState struct {
	// The current ballot we are trying to prepare and commit.
	b *Ballot

	// The highest two ballots that are accepted as prepared.
	// p is the highest, pPrime the next.
	// It's nil if there is no such ballot.
	p *Ballot
	pPrime *Ballot

	// In the Prepare phase, c is the lowest and h is the highest ballot
	// for which we have voted to commit but not accepted abort.
	// In the Confirm phase, c is the lowest and h is the highest ballot
	// for which we accepted commit.
	// In the Externalize phase, c is the lowest and h is the highest ballot
	// for which we confirmed commit.
	// If c is not nil, then c <= h <= b.
	c *Ballot
	h *Ballot

	// The value to use in the next ballot
	z SlotValue
	
	// The latest PrepareMessage, ConfirmMessage, or ExternalizeMessage from each peer
	M map[string]Message

	D QuorumSlice
}

// PrepareMessage is the first phase of the three-phase ballot protocol
type PrepareMessage struct {
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

func (m *PrepareMessage) MessageType() string {
	return "Prepare"
}

// ConfirmMessage is the second phase of the three-phase ballot protocol
type ConfirmMessage struct {
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

func (m *ConfirmMessage) MessageType() string {
	return "Confirm"
}

// ExternalizeMessage is the third phase of the three-phase ballot protocol
// Sent after we have confirmed a commit
type ExternalizeMessage struct {
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

func (m *ExternalizeMessage) MessageType() string {
	return "Externalize"
}

type StateBuilder struct {
	// Which slot is actively being built
	slot int

	// The time we started working on this slot
	start time.Time
	
	// Values for past slots that have already achieved consensus
	values map[int]SlotValue

	nState *NominationState
	bState *BallotState

	// Who we care about
	D QuorumSlice

	// Who we are
	publicKey string
}

func NewStateBuilder(publicKey string, members []string, threshold int) *StateBuilder {
	qs := QuorumSlice{
		Members: members,
		Threshold: threshold,
	}
	return &StateBuilder{
		slot: 1,
		start: time.Now(),
		values: make(map[int]SlotValue),
		nState: NewNominationState(publicKey, qs),
		D: qs,
		publicKey: publicKey,
	}
}

// OutgoingMessage returns nil if there should be no outgoing message at this time
func (sb *StateBuilder) OutgoingMessage() Message {
	// TODO: check if nomination is done and we should send a ballot message

	if !sb.nState.HasNomination() {
		// There's nothing to nominate. Let's nominate something.
		// TODO: if it's not our turn, wait instead of nominating
		comment := fmt.Sprintf(
			"this is %s at %s", sb.publicKey, time.Now().Format("15:04:05.00000"))
		v := MakeSlotValue(comment)
		log.Printf("I nominate %+v", v)
		sb.nState.SetDefault(v)
	}

	return &NominationMessage{
		I: sb.slot,
		X: sb.nState.X,
		Y: sb.nState.Y,
		D: sb.D,
	}
}

// Handle handles an incoming message
func (sb *StateBuilder) Handle(sender string, message Message) {
	switch m := message.(type) {
	case *NominationMessage:
		// log.Printf("handling: %+v", m)
		sb.nState.Handle(sender, m)
	default:
		log.Printf("unrecognized message: %v", m)
	}
}

