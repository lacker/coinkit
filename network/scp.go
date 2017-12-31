package network

import (
	"fmt"
	"log"
	"time"
)

// Stuff for implementing the Stellar Consensus Protocol. See:
// https://www.stellar.org/papers/stellar-consensus-protocol.pdf
// When there are frustrating single-letter variable names, it's because we are
// making the names line up with the protocol paper.

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

// See page 21 of the protocol paper for more detail here.
type NominationState struct {
	// The values we have voted to nominate
	X []SlotValue

	// The values we have accepted as nominated
	Y []SlotValue

	// The values whose nomination we have confirmed
	Z []SlotValue

	// The last NominationMessage received from each node
	N map[string]*NominationMessage

	// Who we are
	publicKey string

	// Who we listen to for quorum
	D QuorumSlice

	// The number of non-duplicate messages this state has processed
	received int
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
		if HasSlotValue(m.Acc, v) {
			votedOrAccepted = append(votedOrAccepted, node)
			accepted = append(accepted, node)
			continue
		}
		if HasSlotValue(m.Nom, v) {
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
		log.Printf("%s accepts the nomination of %+v", s.publicKey, v)
		changed = true
		s.Y = append(s.Y, v)
	}

	// We confirm once a quorum has accepted
	if MeetsQuorum(s, accepted) {
		log.Printf("%s confirms the nomination of %+v", s.publicKey, v)		
		changed = true
		s.Z = append(s.Z, v)
		log.Printf("new s.Z: %+v", s.Z)
	}
	return changed
}

// Handles an incoming nomination message from a peer node
func (s *NominationState) Handle(node string, m *NominationMessage) {
	// What nodes we have seen new information about
	touched := []SlotValue{}

	// Check if there's anything new
	old, ok := s.N[node]
	var oldLenNom, oldLenAcc int
	if ok {
		oldLenNom = len(old.Nom)
		oldLenAcc = len(old.Acc)
	}
	if len(m.Nom) < oldLenNom {
		log.Printf("%s sent a stale message: %v", node, m)
		return
	}
	if len(m.Acc) < oldLenAcc {
		log.Printf("%s sent a stale message: %v", node, m)
		return
	}
	if len(m.Nom) == oldLenNom && len(m.Acc) == oldLenAcc {
		// It's just a dupe
		return
	}
	// Update our most-recent-message
	log.Printf("%s got nomination message from %s: %+v", s.publicKey, node, m)
	s.N[node] = m
	s.received++
	
	for i := oldLenNom; i < len(m.Nom); i++ {
		if !HasSlotValue(touched, m.Nom[i]) {
			touched = append(touched, m.Nom[i])
		}

		// If we don't have a candidate, we can support this new nomination
		if !HasSlotValue(s.X, m.Nom[i]) {
			log.Printf("%s supports the nomination of %+v", s.publicKey, m.Nom[i])
			s.X = append(s.X, m.Nom[i])
			log.Printf("new s.X: %+v", s.X)
		}
	}

	for i := oldLenAcc; i < len(m.Acc); i++ {
		if !HasSlotValue(touched, m.Acc[i]) {
			touched = append(touched, m.Acc[i])
		}
	}

	for _, v := range touched {
		s.MaybeAdvance(v)
	}
}

// See page 23 of the protocol paper for more detail here.
// The null ballot is represented by a nil.
type BallotState struct {
	// What phase of balloting we are in
	phase Phase
	
	// The current ballot we are trying to prepare and commit.
	b *Ballot

	// The highest two incompatible ballots that are accepted as prepared.
	// p is the highest, pPrime the next.
	// It's nil if there is no such ballot.
	p *Ballot
	pPrime *Ballot

	// [cn, hn] defines a range of ballot numbers that defines a range of
	// b-compatible ballots.
	// [0, 0] is the invalid range, rather than [0], since 0 is an invalid ballot
	// number.
	// In the Prepare phase, this is the range we have voted to commit (which
	// we do when we can confirm the ballot is prepared) but that we have not
	// aborted.
	// In the Confirm phase, this is the range we have accepted a commit.
	// In the Externalize phase, this is the range we have confirmed a commit.
	cn int
	hn int

	// The value to use in the next ballot, if this ballot fails.
	// We may have no idea what value we would use. In that case, z is nil.
	z *SlotValue
	
	// The latest PrepareMessage, ConfirmMessage, or ExternalizeMessage from
	// each peer
	M map[string]BallotMessage

	// Who we are
	publicKey string

	// Who we listen to for quorum
	D QuorumSlice

	// The number of non-duplicate messages this state has processed
	received int
}

func NewBallotState(publicKey string, qs QuorumSlice) *BallotState {
	return &BallotState{
		phase: Prepare,
		M: make(map[string]BallotMessage),
		publicKey: publicKey,
		D: qs,
	}
}

func (s *BallotState) PublicKey() string {
	return s.publicKey
}

func (s *BallotState) QuorumSlice(node string) (*QuorumSlice, bool) {
	if node == s.publicKey {
		return &s.D, true
	}
	m, ok := s.M[node]
	if !ok {
		return nil, false
	}
	qs := m.QuorumSlice()
	return &qs, true
}

// MaybeAcceptAsPrepared returns true if the ballot state changes.
func (s *BallotState) MaybeAcceptAsPrepared(n int, x SlotValue) bool {
	if s.phase != Prepare {
		return false
	}
	if n == 0 {
		return false
	}

	// Check if we already accept this as prepared
	if s.p != nil && s.p.n >= n && Equal(s.p.x, x) {
		return false
	}
	if s.pPrime != nil && s.pPrime.n >= n && Equal(s.pPrime.x, x) {
		return false
	}

	if s.pPrime != nil && s.pPrime.n >= n {
		// This is about an old ballot number, we don't care even if it is
		// accepted
		return false
	}
	
	// The rules for accepting are, if a quorum has voted or accepted,
	// we can accept.
	// Or, if a local blocking set has accepted, we can accept.
	votedOrAccepted := []string{}
	accepted := []string{}
	if s.b != nil && s.b.n >= n && Equal(s.b.x, x) {
		// We have voted for this
		votedOrAccepted = append(votedOrAccepted, s.publicKey)
	}

	for node, m := range s.M {
		if m.AcceptAsPrepared(n, x) {
			accepted = append(accepted, node)
			votedOrAccepted = append(votedOrAccepted, node)
			continue
		}
		if m.VoteToPrepare(n, x) {
			votedOrAccepted = append(votedOrAccepted, node)
		}
	}

	if !MeetsQuorum(s, votedOrAccepted) && !s.D.BlockedBy(accepted) {
		// We can't accept this as prepared yet
		return false
	}

	log.Printf("%s accepts as prepared: %d %+v", s.publicKey, n, x)
	
	if s.b != nil && s.hn <= n && !Equal(s.b.x, x) {
		// Accepting this as prepared means we have to abort b
		// TODO: should we set b to something?
		s.hn = 0
		s.cn = 0
		s.b = nil
	}
	
	// p and p prime should be the top two conflicting things we accept
	// as prepared. update them accordingly
	ballot := &Ballot{
		n: n,
		x: x,
	}
	
	if s.p == nil {
		s.p = ballot
	} else if Equal(s.p.x, x) {
		if n <= s.p.n {
			log.Fatal("should have short circuited already")
		}
		s.p.n = n
	} else if n >= s.p.n {
		s.pPrime = s.p
		s.p = ballot
	} else {
		// We already short circuited if it isn't worth bumping p prime
		s.pPrime = ballot
	}
	return true
}

// MaybeConfirmAsPrepared returns whether anything in the ballot state changed.
func (s *BallotState) MaybeConfirmAsPrepared(n int, x SlotValue) bool {
	if s.phase != Prepare {
		return false
	}
	if s.hn >= n {
		// We are already past this ballot
		return false
	}
	ballot := &Ballot{
		n: n,
		x: x,
	}
	
	// We confirm when a quorum accepts as prepared
	accepted := []string{}
	if gtecompat(s.p, ballot) || gtecompat(s.pPrime, ballot) {
		// We accept as prepared
		accepted = append(accepted, s.publicKey)
	}

	for node, m := range s.M {
		if m.AcceptAsPrepared(n, x) {
			accepted = append(accepted, node)
		}
	}

	if !MeetsQuorum(s, accepted) {
		return false
	}

	log.Printf("%s confirms as prepared: %d %+v", s.publicKey, n, x)
	
	// We can confirm this as prepared.
	// Time to vote to commit it
	if s.b != nil && !Equal(s.b.x, x) {
		// We have to abort b
		s.b = nil
		s.hn = 0
		s.cn = 0
	}

	if s.b == nil {
		// We weren't working on any ballot, but now we can work on this one
		s.b = ballot
		s.hn = n
		s.cn = n
		s.z = &x
	} else {
		// We were either working on a lower number, or had not confirmed
		// any numbers as prepared.
		// So bump the range
		s.hn = n
		if s.cn == 0 {
			s.cn = n
		}
	}
	return true
}

// MaybeAcceptAsCommitted returns whether anything in the ballot state changed.
func (s *BallotState) MaybeAcceptAsCommitted(n int, x SlotValue) bool {
	if s.phase == Externalize {
		return false
	}
	if s.phase == Confirm && s.cn <= n && n <= s.hn {
		// We already do accept this commit
		return false
	}

	votedOrAccepted := []string{}
	accepted := []string{}

	if s.phase == Prepare && s.b != nil &&
		Equal(s.b.x, x) && s.cn != 0 && s.cn <= n && n <= s.hn {
		// We vote to commit this
		votedOrAccepted = append(votedOrAccepted, s.publicKey)
	}

	for node, m := range s.M {
		if m.AcceptAsCommitted(n, x) {
			votedOrAccepted = append(votedOrAccepted, node)
			accepted = append(accepted, node)
		} else if m.VoteToCommit(n, x) {
			votedOrAccepted = append(votedOrAccepted, node)
		}
	}

	if !MeetsQuorum(s, votedOrAccepted) && !s.D.BlockedBy(accepted) {
		// We can't accept this commit yet
		return false
	}

	log.Printf("%s accepts as committed: %d %+v", s.publicKey, n, x)
	
	// We accept this commit
	s.phase = Confirm
	if s.b == nil || !Equal(s.b.x, x) {
		// Totally replace our old target value
		s.b = &Ballot{
			n: n,
			x: x,
		}
		s.cn = n
		s.hn = n
		s.z = &x
	} else {
		// Just update our range of acceptance
		if n < s.cn {
			s.cn = n
		}
		if n > s.hn {
			s.hn = n
		}
	}
	return true
}

// MaybeConfirmAsCommitted returns whether anything in the ballot state changed.
func (s *BallotState) MaybeConfirmAsCommitted(n int, x SlotValue) bool {
	if s.phase == Prepare {
		return false
	}
	if s.b == nil || !Equal(s.b.x, x) {
		return false
	}
	
	accepted := []string{}
	if s.phase == Confirm {
		if s.cn <= n && n <= s.hn {
			accepted = append(accepted, s.publicKey)
		}
	} else if s.cn <= n && n <= s.hn {
		// We already did confirm this as committed
		return false
	}

	for node, m := range s.M {
		if m.AcceptAsCommitted(n, x) {
			accepted = append(accepted, node)
		}
	}

	if !MeetsQuorum(s, accepted) {
		return false
	}
	
	log.Printf("%s confirms as committed: %d %+v", s.publicKey, n, x)
	
	if s.phase == Confirm {
		s.phase = Externalize
		s.cn = n
		s.hn = n
	} else {
		if n < s.cn {
			s.cn = n
		}
		if n > s.hn {
			s.hn = n
		}
	}

	return true
}

// Returns whether we needed to bump the ballot number.
// We bump the ballot number if the set of nodes with a higher
// ballot number is blocking.
// TODO: figure out what the distinction is between s.b.n and s.hn
// TODO: figure out if s.z and s.b.x are redundant
func (s *BallotState) MaybeNextBallot() bool {
	if s.z == nil || s.b == nil {
		return false
	}
	
	// Nodes with a higher ballot number
	higher := []string{}
	current := 0
	if s.b != nil {
		current = s.b.n
	}

	for node, m := range s.M {
		if m.BallotNumber() > current {
			higher = append(higher, node)
		}
	}

	if !s.D.BlockedBy(higher) {
		return false
	}

	// s.z and s.b.x should be equivalent here
	s.b.n++
	return true
}

// Update the stage of this ballot as needed
// See the handling algorithm on page 24 of the Mazieres paper.
// The investigate method does steps 1-8
func (s *BallotState) Investigate(n int, x SlotValue) {
	s.MaybeAcceptAsPrepared(n, x)
	s.MaybeConfirmAsPrepared(n, x)
	s.MaybeAcceptAsCommitted(n, x)
	s.MaybeConfirmAsCommitted(n, x)
}

func (s *BallotState) Handle(node string, message BallotMessage) {
	// If this message isn't new, skip it
	old, ok := s.M[node]
	if ok && Compare(old, message) >= 0 {
		return
	}
	log.Printf("%s got ballot message from %s: %+v", s.publicKey, node, message)
	s.received++
	s.M[node] = message

	for {
		// Investigate all ballots whose state might be updated
		// TODO: make sure we aren't missing ballot numbers internal to the
		// ranges
		// TODO: make sure a malformed message can't DDOS us here
		switch m := message.(type) {
		case *PrepareMessage:
			s.Investigate(m.Bn, m.Bx)
			s.Investigate(m.Pn, m.Px)
			s.Investigate(m.Ppn, m.Ppx)
		case *ConfirmMessage:
			s.Investigate(m.Hn, m.X)
		case *ExternalizeMessage:
			for i := m.Cn; i <= m.Hn; i++ {
				s.Investigate(i, m.X)
			}
		}

		// Step 9 of the processing algorithm
		if !s.MaybeNextBallot() {
			break
		}
	}
}

// MaybeInitializeValue initializes the value if it doesn't already have a value,
// and returns whether anything in the ballot state changed.
func (s *BallotState) MaybeInitializeValue(v SlotValue) bool {
	if s.z != nil {
		return false
	}
	s.z = &v
	if s.b == nil {
		s.b = &Ballot{
			n: 1,
			x: v,
		}
	}
	return true
}

func (s *BallotState) HasMessage() bool {
	return s.b != nil
}

func (s *BallotState) Message(slot int, qs QuorumSlice) Message {
	if !s.HasMessage() {
		panic("coding error")
	}

	switch s.phase {
	case Prepare:
		m := &PrepareMessage{
			T: Prepare,
			I: slot,
			Bn: s.b.n,
			Bx: s.b.x,
			Cn: s.cn,
			Hn: s.hn,
			D: qs,
		}
		if s.p != nil {
			m.Pn = s.p.n
			m.Px = s.p.x
		}
		if s.pPrime != nil {
			m.Ppn = s.pPrime.n
			m.Ppx = s.pPrime.x
		}
		return m

	case Confirm:
		m := &ConfirmMessage{
			T: Confirm,
			I: slot,
			X: s.b.x,
			Cn: s.cn,
			Hn: s.hn,
			D: qs,
		}
		if s.p != nil {
			m.Pn = s.p.n
		}
		return m

	case Externalize:
		return &ExternalizeMessage{
			T: Externalize,
			I: slot,
			X: s.b.x,
			Cn: s.cn,
			Hn: s.hn,
			D: qs,
		}
	}

	panic("code flow should not get here")
}

type ChainState struct {
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

func NewChainState(publicKey string, members []string, threshold int) *ChainState {
	log.Printf("I am %s", publicKey)
	qs := QuorumSlice{
		Members: members,
		Threshold: threshold,
	}
	return &ChainState{
		slot: 1,
		start: time.Now(),
		values: make(map[int]SlotValue),
		nState: NewNominationState(publicKey, qs),
		bState: NewBallotState(publicKey, qs),
		D: qs,
		publicKey: publicKey,
	}
}

// OutgoingMessages returns the outgoing messages.
// There can be zero or one nomination messages, and zero or one ballot messages.
func (cs *ChainState) OutgoingMessages() []Message {
	answer := []Message{}

	if !cs.nState.HasNomination() {
		// There's nothing to nominate. Let's nominate something.
		// TODO: if it's not our turn, wait instead of nominating
		comment := fmt.Sprintf(
			"this is %s at %s", cs.publicKey, time.Now().Format("15:04:05.00000"))
		v := MakeSlotValue(comment)
		log.Printf("%s nominates %+v", cs.publicKey, v)
		cs.nState.SetDefault(v)
	}

	answer = append(answer, &NominationMessage{
		I: cs.slot,
		Nom: cs.nState.X,
		Acc: cs.nState.Y,
		D: cs.D,
	})

	// If we aren't working on any ballot, but we do have a nomination, we can
	// optimistically start working on that ballot
	if cs.nState.HasNomination() && cs.bState.z == nil {
		cs.bState.MaybeInitializeValue(cs.nState.PredictValue())
	}
	
	if cs.bState.HasMessage() {
		answer = append(answer, cs.bState.Message(cs.slot, cs.D))
	}

	return answer
}

// Done returns whether this chain has externalized all the slots it is working on.
func (cs *ChainState) Done() bool {
	return cs.bState.phase == Externalize
}

// Handle handles an incoming message
func (cs *ChainState) Handle(sender string, message Message) {
	if sender == cs.publicKey {
		// It's one of our own returning to us, we can ignore it
		return
	}
	switch m := message.(type) {
	case *NominationMessage:
		cs.nState.Handle(sender, m)
	case *PrepareMessage:
		cs.bState.Handle(sender, m)
	case *ConfirmMessage:
		cs.bState.Handle(sender, m)
	case *ExternalizeMessage:
		cs.bState.Handle(sender, m)
	default:
		log.Printf("unrecognized message: %v", m)
	}
}

