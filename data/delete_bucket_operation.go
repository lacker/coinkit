package data

import (
	"fmt"

	"github.com/lacker/coinkit/util"
)

type DeleteBucketOperation struct {
	// Who is creating this bucket
	Signer string `json:"signer"`

	// The sequence number for this operation
	Sequence uint32 `json:"sequence"`

	// The operation fee for entering an op into the blockchain
	Fee uint64 `json:"fee"`

	// The name of the bucket
	Name string `json:"name"`
}

func (op *DeleteBucketOperation) String() string {
	return fmt.Sprintf("DeleteBucket owner=%s, name=%s",
		util.Shorten(op.Signer), op.Name, op.Size)
}

func (op *DeleteBucketOperation) OperationType() string {
	return "DeleteBucket"
}

func (op *DeleteBucketOperation) GetSigner() string {
	return op.Signer
}

func (op *DeleteBucketOperation) GetFee() uint64 {
	return op.Fee
}

func (op *DeleteBucketOperation) GetSequence() uint32 {
	return op.Sequence
}

// TODO: should this do something?
func (op *DeleteBucketOperation) Verify() bool {
	return true
}

func MakeTestDeleteBucketOperation(n int) *SignedOperation {
	mint := util.NewKeyPairFromSecretPhrase("mint")
	op := &DeleteBucketOperation{
		Signer:   mint.PublicKey().String(),
		Sequence: uint32(n),
		Name:     fmt.Sprintf("bucket%d", n),
	}
	return NewSignedOperation(op, mint)
}

func init() {
	RegisterOperationType(&DeleteBucketOperation{})
}
