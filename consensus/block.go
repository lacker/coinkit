package consensus

import (
	"log"
	"strings"
	"time"

	"coinkit/util"
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

	// The hash of the previous block, or "" if this is the first one
	prevHash string
	
	// Who we care about
	D QuorumSlice

	// Who we are
	publicKey string
}

func NewBlock(publicKey string, qs QuorumSlice, slot int, prevHash string) *Block {
	nState := NewNominationState(publicKey, qs, prevHash)
	block := &Block{
		slot:      slot,
		start:     time.Now(),
		nState:    nState,
		bState:    NewBallotState(publicKey, qs, nState),
		prevHash:  prevHash,
		D:         qs,
		publicKey: publicKey,
	}
	block.MaybeNominateNewValue()
	return block
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
func (b *Block) OutgoingMessages() []util.Message {
	if b.external != nil {
		// This block is already externalized
		return []util.Message{b.external}
	}
	
	answer := []util.Message{b.nState.Message(b.slot, b.D)}

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

func (b *Block) MaybeNominateNewValue() {
	if b.nState.WantsToNominateNewValue() {
		comment := strings.Replace(b.publicKey, "node", "comment", 1)
		v := MakeSlotValue(comment)
		log.Printf("%s nominates %+v", b.publicKey, v)
		b.nState.NominateNewValue(v)
	}
}

// Handle handles an incoming message
func (b *Block) Handle(sender string, message util.Message) {
	if sender == b.publicKey {
		// It's one of our own returning to us, we can ignore it
		return
	}
	switch m := message.(type) {
	case *NominationMessage:
		b.nState.Handle(sender, m)
		b.MaybeNominateNewValue()
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
