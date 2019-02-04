package data

import (
	"fmt"

	"github.com/lacker/coinkit/util"
)

type AllocateOperation struct {
	// Who is performing this allocation. Can be either the bucket or provider owner
	Signer string `json:"signer"`

	// The sequence number for this operation
	Sequence uint32 `json:"sequence"`

	// The operation fee for entering an op into the blockchain
	Fee uint64 `json:"fee"`

	// The name of the bucket
	BucketName string `json:"bucketName"`

	// The id of the provider
	ProviderID uint64 `json:"providerID"`
}

func (op *AllocateOperation) String() string {
	return fmt.Sprintf("Allocate signer=%s, bucketName=%s, providerID=%d",
		util.Shorten(op.Signer), op.BucketName, op.ProviderID)
}

func (op *AllocateOperation) OperationType() string {
	return "Allocate"
}

func (op *AllocateOperation) GetSigner() string {
	return op.Signer
}

func (op *AllocateOperation) GetFee() uint64 {
	return op.Fee
}

func (op *AllocateOperation) GetSequence() uint32 {
	return op.Sequence
}

func (op *AllocateOperation) Verify() bool {
	if !IsValidBucketName(op.BucketName) {
		return false
	}
	if op.ProviderID == 0 {
		return false
	}
	return true
}

func init() {
	RegisterOperationType(&AllocateOperation{})
}
