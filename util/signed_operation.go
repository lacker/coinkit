package util

import (
// "encoding/json"
// "fmt"
// "strings"
)

type SignedOperation struct {
	Operation Operation

	// The signature to prove that the sender has signed this
	// Nil if the transaction has not been signed
	Signature string
}

func NewSignedOperation(op Operation, kp *KeyPair) *SignedOperation {
	panic("TODO: implement")
}
