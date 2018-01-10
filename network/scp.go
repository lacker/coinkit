package network

import (
	"log"
	"strings"
	"time"
)

// Stuff for implementing the Stellar Consensus Protocol. See:
// https://www.stellar.org/papers/stellar-consensus-protocol.pdf
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
		Members:   members,
		Threshold: threshold,
	}
	return &ChainState{
		slot:      1,
		start:     time.Now(),
		values:    make(map[int]SlotValue),
		nState:    NewNominationState(publicKey, qs),
		bState:    NewBallotState(publicKey, qs),
		D:         qs,
		publicKey: publicKey,
	}
}

func (cs *ChainState) AssertValid() {
	cs.nState.AssertValid()
	cs.bState.AssertValid()
}

// OutgoingMessages returns the outgoing messages.
// There can be zero or one nomination messages, and zero or one ballot messages.
func (cs *ChainState) OutgoingMessages() []Message {
	answer := []Message{}

	if !cs.nState.HasNomination() {
		// There's nothing to nominate. Let's nominate something.
		// TODO: if it's not our turn, wait instead of nominating
		comment := strings.Replace(cs.publicKey, "node", "comment", 1)
		v := MakeSlotValue(comment)
		log.Printf("%s nominates %+v", cs.publicKey, v)
		cs.nState.SetDefault(v)
	}

	answer = append(answer, &NominationMessage{
		I:   cs.slot,
		Nom: cs.nState.X,
		Acc: cs.nState.Y,
		D:   cs.D,
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
		cs.bState.MaybeUpdateValue(cs.nState)
	case *PrepareMessage:
		cs.bState.Handle(sender, m)
	case *ConfirmMessage:
		cs.bState.Handle(sender, m)
	case *ExternalizeMessage:
		cs.bState.Handle(sender, m)
	default:
		log.Printf("unrecognized message: %v", m)
	}

	cs.AssertValid()
}

func (cs *ChainState) HandleTimerTick() {
	cs.bState.HandleTimerTick()
}
