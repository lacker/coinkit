package data

import (
	"fmt"

	"github.com/lacker/coinkit/util"
)

// CreateOperation is used to create a new document on the blockchain.
type CreateOperation struct {
	// Who is creating this document
	Signer string `json:"signer"`

	// The sequence number for this operation
	Sequence uint32 `json:"sequence"`

	// How much the creator is willing to pay to get this document registered
	Fee uint64 `json:"fee"`

	// The data to be created in the new document
	Data *JSONObject `json:"data"`
}

func (op *CreateOperation) String() string {
	return fmt.Sprintf("create owner=%s, data=%s", util.Shorten(op.Signer), op.Data)
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

// TODO: should this do something?
func (op *CreateOperation) Verify() bool {
	return true
}

func MakeTestCreateOperation(n int) *SignedOperation {
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

func (op *CreateOperation) Document(id uint64) *Document {
	data := op.Data.Copy()
	data.Set("id", id)
	return &Document{
		Data: data,
		Id:   id,
	}
}

func init() {
	RegisterOperationType(&CreateOperation{})
}
