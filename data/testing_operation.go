package data

import ()

type TestingOperation struct {
	Number  int
	Signer  string
	Invalid bool
}

func (op *TestingOperation) OperationType() string {
	return "Testing"
}

func (op *TestingOperation) String() string {
	return "Testing"
}

func (op *TestingOperation) GetSigner() string {
	return op.Signer
}

func (op *TestingOperation) Verify() bool {
	return !op.Invalid
}

func (op *TestingOperation) GetFee() uint64 {
	return 0
}

func (op *TestingOperation) GetSequence() uint32 {
	return 1
}

func init() {
	RegisterOperationType(&TestingOperation{})
}
