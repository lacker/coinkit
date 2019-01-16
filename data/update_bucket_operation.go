package data

import (
	"fmt"

	"github.com/lacker/coinkit/util"
)

type UpdateBucketOperation struct {
	// Who is creating this bucket
	Signer string `json:"signer"`

	// The sequence number for this operation
	Sequence uint32 `json:"sequence"`

	// The operation fee for entering an op into the blockchain
	Fee uint64 `json:"fee"`

	// The name of the bucket
	Name string `json:"name"`

	// The size of the bucket in megabytes
	Size uint32 `json:"size"`
}

func (op *UpdateBucketOperation) String() string {
	return fmt.Sprintf("UpdateBucket owner=%s, name=%s, size=%d",
		util.Shorten(op.Signer), op.Name, op.Size)
}

func (op *UpdateBucketOperation) OperationType() string {
	return "UpdateBucket"
}

func (op *UpdateBucketOperation) GetSigner() string {
	return op.Signer
}

func (op *UpdateBucketOperation) GetFee() uint64 {
	return op.Fee
}

func (op *UpdateBucketOperation) GetSequence() uint32 {
	return op.Sequence
}

// TODO: should this do something?
func (op *UpdateBucketOperation) Verify() bool {
	return true
}

func MakeTestUpdateBucketOperation(n int) *SignedOperation {
	mint := util.NewKeyPairFromSecretPhrase("mint")
	op := &UpdateBucketOperation{
		Signer:   mint.PublicKey().String(),
		Sequence: uint32(n),
		Name:     fmt.Sprintf("bucket%d", n),
		Size:     uint32(n * 2000),
	}
	return NewSignedOperation(op, mint)
}

func init() {
	RegisterOperationType(&UpdateBucketOperation{})
}
