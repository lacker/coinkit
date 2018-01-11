package network

import (
	"log"
)

// Chain creates the blockchain, one Block at a time.
type Chain struct {
	// The block we are currently working on
	current *Block

	// Maps slot number to previous blocks
	// Every block in here should be externalized
	history map[int]*Block

	// The quorum logic we use for future blocks
	D QuorumSlice

	// Who we are
	publicKey string
}

// Handle handles an incoming message.
// It may return a message to be sent back to the original sender, or it may
// just return nil if it has no particular response.
func (c *Chain) Handle(sender string, message Message) Message {
	if sender == c.publicKey {
		// It's one of our own returning to us, we can ignore it
		return nil
	}

	slot := message.Slot()
	if slot == 0 {
		log.Fatalf("slot should not be zero in %+v", message)
	}

	if slot == c.current.slot {
		c.current.Handle(sender, message)
		if c.current.Done() {
			// This block is done, let's move on to the next one
			c.history[slot] = c.current
			c.current = NewBlock(c.publicKey, c.D, slot + 1)
		}
		return nil
	}

	block, ok := c.history[slot]
	if !ok {
		// We aren't working on this slot, ignore
		return nil
	}

	block.Handle(sender, message)
	return nil
}

func NewEmptyChain(publicKey string, qs QuorumSlice) *Chain {
	return &Chain{
		current: NewBlock(publicKey, qs, 1),
		history: make(map[int]*Block),
		D: qs,
		publicKey: publicKey,
	}
}

func (c *Chain) OutgoingMessages() []Message {
	answer := c.current.OutgoingMessages()

	prev := c.history[c.current.slot-1]
	if prev != nil {
		// We also send out the externalize data for the previous block
		answer = append(answer, prev.OutgoingMessages()...)
	}

	return answer
}
