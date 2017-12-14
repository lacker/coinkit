package network

import (
	"fmt"
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

type QuorumSlice struct {
	// Members is a list of public keys for nodes that occur in the quorum slice.
	// Members must be unique.
	// Typically includes ourselves.
	Members []string

	// The number of members we require for consensus, including ourselves.
	// The protocol can support other sorts of slices, like weighted or any wacky
	// thing, but for now we only do this simple "any k out of these n" voting.
	Threshold int
}

func (qs *QuorumSlice) atLeast(nodes []string, t int) bool {
	count := 0
	for _, member := range qs.Members {
		for _, node := range nodes {
			if node == member {
				count++
				if count >= t {
					return true
				}
				break
			}
		}
	}
	return false
}

func (qs *QuorumSlice) BlockedBy(nodes []string) bool {
	return qs.atLeast(nodes, len(qs.Members) - qs.Threshold + 1)
}

func (qs *QuorumSlice) SatisfiedWith(nodes []string) bool {
	return qs.atLeast(nodes, qs.Threshold)
}

type NominateMessage struct {
	// What slot we are nominating values for
	I int

	// The values we have voted to nominate
	X []SlotValue

	// The values we have accepted as nominated
	Y []SlotValue
	
	D QuorumSlice
}

func (m *NominateMessage) MessageType() string {
	return "Nominate"
}

// See page 21 of the protocol paper for more detail here.
type NominationState struct {
	// The values we have voted to nominate
	X []SlotValue

	// The values we have accepted as nominated
	Y []SlotValue

	// The values that we consider to be candidates 
	Z []SlotValue

	// The last NominateMessage received from each node
	N map[string]*NominateMessage

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
		N: make(map[string]*NominateMessage),
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

type QuorumFinder interface {
	QuorumSlice(node string) (*QuorumSlice, bool)
	PublicKey() string
}

// Returns whether this set of nodes meets the quorum for the network overall.
func MeetsQuorum(f QuorumFinder, nodes []string) bool {
	// Filter out the nodes in the potential quorum that do not have their
	// own quorum slices met
	hasUs := false
	filtered := []string{}
	for _, node := range nodes {
		qs, ok := f.QuorumSlice(node)
		if ok && qs.SatisfiedWith(nodes) {
			filtered = append(filtered, node)
			if node == f.PublicKey() {
				hasUs = true
			}
		}
	}
	if !hasUs {
		return false
	}
	if len(filtered) == len(nodes) {
		return true
	}
	return MeetsQuorum(f, filtered)
}

// MaybeAccept checks whether we should accept the nomination for this slot value,
// and adds it to our accepted list if appropriate.
// Returns whether we added v.
func (s *NominationState) MaybeAccept(v SlotValue) bool {
	if HasSlotValue(s.Y, v) {
		// We already did accept v's nomination
		return false
	}
	
	votedOrAccepted := []string{}
	if HasSlotValue(s.X, v) {
		votedOrAccepted = append(votedOrAccepted, s.publicKey)
	}
	accepted := []string{}
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

	// The rules are on page 13, section 5.3
	// Rule 1: if a quorum has either voted for the nomination or accepted the
	// nomination, we accept it.
	if MeetsQuorum(s, votedOrAccepted) {
		s.Y = append(s.Y, v)
		return true
	}
	// Rule 2: if a blocking set for us accepts the nomination, we accept it.
	if s.D.BlockedBy(accepted) {
		s.Y = append(s.Y, v)
		return true
	}
	return false
}

// Handles an incoming nomination message from a peer
func (s *NominationState) Handle(node string, m *NominateMessage) {
	// TODO
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
		sb.nState.SetDefault(MakeSlotValue(comment))
	}

	return &NominateMessage{
		I: sb.slot,
		X: sb.nState.X,
		Y: sb.nState.Y,
		D: sb.D,
	}
}

// Handle handles an incoming message
func (sb *StateBuilder) Handle(sender string, m Message) {
	// TODO: handle incoming nomination messages
}

