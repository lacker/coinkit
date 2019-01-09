package data

import (
	"fmt"

	"github.com/lacker/coinkit/util"
)

type CreateBucketOperation struct {
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

func (op *CreateBucketOperation) String() string {
	return fmt.Sprintf("CreateBucket owner=%s, name=%s, size=%d",
		util.Shorten(op.signer), op.Name, op.Size)
}

func (op *CreateBucketOperation) OperationType() string {
	return "CreateBucket"
}

func (op *CreateBucketOperation) GetSigner() string {
	return op.Signer
}

func (op *CreateBucketOperation) GetFee() uint64 {
	return op.Fee
}

func (op *CreateOperation) GetSequence() uint32 {
	return op.Sequence
}

// TODO: should this do something?
func (op *CreateOperation) Verify() bool {
	return true
}

func init() {
	RegisterOperationType(&CreateBucketOperation{})
}
