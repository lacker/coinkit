package network

import (
	"log"
	"strings"
	"time"
)

// Block implements the convergence algorithm for a single block,
// according to the Stellar Consensus Protocol. See:
// https://www.stellar.org/papers/stellar-consensus-protocol.pdf
// Most logic is not in the Block itself, but is delegated to the
// NominationState for the nomination phase and the BallotState for the
// ballot phase.
type Block struct {
	// Which slot this block state is building
	slot int

	// The time we started working on this slot
	start time.Time

	nState *NominationState
	bState *BallotState

	// This is nil before the block is finalized.
	// When it is finalized, this is all we need to keep around in order
	// to catch up old nodes.
	external *ExternalizeMessage

	// Who we care about
	D QuorumSlice

	// Who we are
	publicKey string
}

func NewBlock(publicKey string, qs QuorumSlice, slot int) *Block {
	nState := NewNominationState(publicKey, qs)
	return &Block{
		slot:      slot,
		start:     time.Now(),
		nState:    nState,
		bState:    NewBallotState(publicKey, qs, nState),
		D:         qs,
		publicKey: publicKey,
	}
}

func (block *Block) AssertValid() {
	block.nState.AssertValid()
	block.bState.AssertValid()
	if block.bState.phase == Externalize && block.external == nil {
		block.bState.Show()
		log.Fatalf("this block has externalized but block.external is not set")
	}
}

// OutgoingMessages returns the outgoing messages.
// There can be zero or one nomination messages, and zero or one ballot messages.
func (b *Block) OutgoingMessages() []Message {
	if b.external != nil {
		// This block is already externalized
		return []Message{b.external}
	}
	
	answer := []Message{}

	if !b.nState.HasNomination() {
		// There's nothing to nominate. Let's nominate something.
		// TODO: if it's not our turn, wait instead of nominating
		comment := strings.Replace(b.publicKey, "node", "comment", 1)
		v := MakeSlotValue(comment)
		log.Printf("%s nominates %+v", b.publicKey, v)
		b.nState.SetDefault(v)
	}

	answer = append(answer, &NominationMessage{
		I:   b.slot,
		Nom: b.nState.X,
		Acc: b.nState.Y,
		D:   b.D,
	})

	// If we aren't working on any ballot, try to start working on a ballot
	if b.bState.b == nil {
		b.bState.GoToNextBallot()
	}

	if b.bState.HasMessage() {
		m := b.bState.Message(b.slot, b.D)
		answer = append(answer, m)
	}

	return answer
}

func (b *Block) Done() bool {
	return b.external != nil
}

// Handle handles an incoming message
func (b *Block) Handle(sender string, message Message) {
	if sender == b.publicKey {
		// It's one of our own returning to us, we can ignore it
		return
	}
	switch m := message.(type) {
	case *NominationMessage:
		b.nState.Handle(sender, m)
	case *PrepareMessage:
		b.bState.Handle(sender, m)
	case *ConfirmMessage:
		b.bState.Handle(sender, m)
	case *ExternalizeMessage:
		b.bState.Handle(sender, m)
	default:
		log.Printf("unrecognized message: %v", m)
	}

	if b.bState.phase == Externalize && b.external == nil {
		b.external = b.bState.Message(b.slot, b.D).(*ExternalizeMessage)
	}
	
	b.AssertValid()
}

func (b *Block) HandleTimerTick() {
	b.bState.HandleTimerTick()
}
