package data

import ()

type DeleteOperation struct {
	// Who is deleting the document. Must be the owner
	Signer string

	// The sequence number for this operation
	Sequence uint32

	// How much the updater is willing to pay to send this operation through
	Fee uint64

	// The id of the document to update
	Id uint64
}
