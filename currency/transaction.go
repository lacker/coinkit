package currency

import (
)

type Transaction struct {
	// Who is sending this money
	From string
	
	// The sequence number for this transaction
	Sequence uint32

	// Who is receiving this money
	To string
	
	// The amount of currency to transfer
	Amount uint64

	// How much the sender is willing to pay to get this transfer registered
	// This is on top of the amount
	Fee uint64
}
