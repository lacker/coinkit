package data

import ()

type AllocateOperation struct {
	// Who is performing this allocation. Can be either the bucket or provider owner
	Signer string `json:"signer"`

	// The sequence number for this operation
	Sequence uint32 `json:"sequence"`

	// The operation fee for entering an op into the blockchain
	Fee uint64 `json:"fee"`

	// The name of the bucket
	Name string `json:"name"`

	// The size of the bucket in megabytes
	ID uint64 `json:"id"`
}

func init() {
	RegisterOperationType(&AllocateOperation{})
}
