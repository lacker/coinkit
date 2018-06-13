package data

import (
	"fmt"

	"github.com/lacker/coinkit/util"
)

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

func (op *DeleteOperation) String() string {
	return fmt.Sprintf("delete owner=%s, id=%d", util.Shorten(op.Signer), op.Id)
}

func (op *DeleteOperation) OperationType() string {
	return "Delete"
}

func (op *DeleteOperation) GetSigner() string {
	return op.Signer
}

func (op *DeleteOperation) GetFee() uint64 {
	return op.Fee
}

func (op *DeleteOperation) GetSequence() uint32 {
	return op.Sequence
}

// TODO: should this do something?
func (op *DeleteOperation) Verify() bool {
	return true
}

func init() {
	RegisterOperationType(&DeleteOperation{})
}
