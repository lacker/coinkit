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

// Handle handles an incoming message
func (c *Chain) Handle(sender string, message Message) {
	if sender == c.publicKey {
		// It's one of our own returning to us, we can ignore it
		return
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
		return
	}

	block, ok := c.history[slot]
	if !ok {
		// We aren't working on this slot, ignore
		return
	}

	block.Handle(sender, message)
}
