package data

import (
	"coinkit/currency"
)

// data.Block represents how the value for a single block gets stored to the database.
type Block struct {
	// Which block this is
	Slot int

	// The LedgerChunk for this block
	// TODO: do not let this be nil
	Chunk *currency.LedgerChunk

	// The ballot numbers this node confirmed.
	C int
	H int
}
