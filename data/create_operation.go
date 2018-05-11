package data

import (
	"fmt"

	"github.com/lacker/coinkit/util"
)

// CreateOperation is used to create a new document on the blockchain.
type CreateOperation struct {
	// Who is creating this document
	Signer string

	// The sequence number for this operation
	Sequence uint32

	// The data to be created in the new document
	Data *JSONObject

	// How much the creator is willing to pay to get this document registered
	Fee uint64
}

func (op *CreateOperation) String() string {
	return fmt.Sprintf("create by %s: %s", util.Shorten(op.Signer), op.Data)
}

func (op *CreateOperation) OperationType() string {
	return "Create"
}

func (op *CreateOperation) GetSigner() string {
	return op.Signer
}

func (op *CreateOperation) GetFee() uint64 {
	return op.Fee
}

func (op *CreateOperation) GetSequence() uint32 {
	return op.Sequence
}

func (op *CreateOperation) Verify() bool {
	return true
}

func makeTestCreateOperation(n int) *SignedOperation {
	mint := util.NewKeyPairFromSecretPhrase("mint")
	data := NewEmptyJSONObject()
	data.Set("foo", n)
	op := &CreateOperation{
		Signer:   mint.PublicKey().String(),
		Sequence: uint32(n),
		Data:     data,
		Fee:      0,
	}
	return NewSignedOperation(op, mint)
}

func init() {
	RegisterOperationType(&CreateOperation{})
}
