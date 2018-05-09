package data

import (
	"encoding/json"

	"github.com/lacker/coinkit/consensus"
)

// data.Block represents how the value for a single block gets stored to the database.
type Block struct {
	// Which block this is
	Slot int

	// The LedgerChunk for this block
	Chunk *LedgerChunk

	// The ballot numbers this node confirmed.
	C int
	H int

	// The quorum slice used to confirm this block
	D consensus.QuorumSlice
}

// ExternalizeMessage() constructs a message with the metadata for how we came to
// consensus on this block
func (b *Block) ExternalizeMessage() *consensus.ExternalizeMessage {
	return &consensus.ExternalizeMessage{
		I:  b.Slot,
		X:  b.Chunk.Hash(),
		Cn: b.C,
		Hn: b.H,
		D:  b.D,
	}
}

// OperationMessage() constructs a message with the chunk contents
func (b *Block) OperationMessage() *OperationMessage {
	return NewOperationMessageWithChunk(b.Chunk)
}

func (b *Block) String() string {
	bytes, err := json.MarshalIndent(b, "", "  ")
	if err != nil {
		panic(err)
	}
	return string(append(bytes, '\n'))
}
