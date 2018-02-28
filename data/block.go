package data

import ()

// data.Block represents how the value for a single block gets stored to the database.
type Block struct {
	// Which block this is
	Slot int

	// The block data
	Value string

	// The ballot numbers this node confirmed.
	C int
	H int
}
