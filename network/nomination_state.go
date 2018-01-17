package network

import (
)

// The nomination state for the Stellar Consensus Protocol.
// See page 21 of:
// https://www.stellar.org/papers/stellar-consensus-protocol.pdf
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

	// The hash of the previous block, used to pseudorandomly determine
	// which node should heuristically start nominations
	prevHash string

	// A list of the nodes in priority order for who should initiate the
	// nomination
	priority []string
}

func NewNominationState(
	publicKey string, qs QuorumSlice, prevHash string) *NominationState {

	return &NominationState{
		X:         make([]SlotValue, 0),
		Y:         make([]SlotValue, 0),
		Z:         make([]SlotValue, 0),
		N:         make(map[string]*NominationMessage),
		publicKey: publicKey,
		D:         qs,
		prevHash:  prevHash,
		priority:  SeedSort(prevHash, qs.Members),
	}	
}

func (s *NominationState) Logf(format string, a ...interface{}) {
	// log.Printf(format, a...)
}

func (s *NominationState) Show() {
	s.Logf("nState for %s:", s.publicKey)
	s.Logf("X: %+v", s.X)
	s.Logf("Y: %+v", s.Y)
	s.Logf("Z: %+v", s.Z)
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
		answer := CombineSlice(s.Y)
		return answer
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

func (s *NominationState) AssertValid() {
	AssertNoDupes(s.X)
	AssertNoDupes(s.Y)
	AssertNoDupes(s.Z)
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
		s.Logf("%s accepts the nomination of %+v", s.publicKey, v)
		changed = true
		s.Logf("old s.Y: %+v", s.Y)
		AssertNoDupes(s.Y)
		s.Y = append(s.Y, v)
		accepted = append(accepted, s.publicKey)
		s.Logf("new s.Y: %+v", s.Y)
		AssertNoDupes(s.Y)
	}

	// We confirm once a quorum has accepted
	if MeetsQuorum(s, accepted) {
		s.Logf("%s confirms the nomination of %+v", s.publicKey, v)
		changed = true
		s.Z = append(s.Z, v)
		s.Logf("new s.Z: %+v", s.Z)
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
		s.Logf("%s sent a stale message: %v", node, m)
		return
	}
	if len(m.Acc) < oldLenAcc {
		s.Logf("%s sent a stale message: %v", node, m)
		return
	}
	if len(m.Nom) == oldLenNom && len(m.Acc) == oldLenAcc {
		// It's just a dupe
		return
	}
	// Update our most-recent-message
	s.Logf("\n\n%s got nomination message from %s:\n%+v", s.publicKey, node, m)
	s.N[node] = m
	s.received++

	for i := oldLenNom; i < len(m.Nom); i++ {
		if !HasSlotValue(touched, m.Nom[i]) {
			touched = append(touched, m.Nom[i])
		}

		// If we don't have a candidate, we can support this new nomination
		if !HasSlotValue(s.X, m.Nom[i]) {
			s.Logf("%s supports the nomination of %+v", s.publicKey, m.Nom[i])
			s.X = append(s.X, m.Nom[i])
			s.Logf("new s.X: %+v", s.X)
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


