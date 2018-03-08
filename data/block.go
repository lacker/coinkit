package data

import (
	"github.com/lacker/coinkit/consensus"
	"github.com/lacker/coinkit/currency"
)

// data.Block represents how the value for a single block gets stored to the database.
type Block struct {
	// Which block this is
	Slot int

	// The LedgerChunk for this block
	Chunk *currency.LedgerChunk

	// The ballot numbers this node confirmed.
	C int
	H int
}

func (b *Block) ExternalizeMessage(d consensus.QuorumSlice) *consensus.ExternalizeMessage {
	return &consensus.ExternalizeMessage{
		I:  b.Slot,
		X:  b.Chunk.Hash(),
		Cn: b.C,
		Hn: b.H,
		D:  d,
	}
}
