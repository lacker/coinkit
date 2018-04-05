package util

import (
	"encoding/json"
	"reflect"
)

type SignedOperation struct {
	Operation Operation

	// The signature to prove that the sender has signed this
	// Nil if the transaction has not been signed
	Signature string
}

func NewSignedOperation(op Operation, kp *KeyPair) *SignedOperation {
	if op == nil || reflect.ValueOf(op).IsNil() {
		Logger.Fatal("cannot sign nil operation")
	}

	if kp.PublicKey().String() != op.GetSigner() {
		Logger.Fatal("you can only sign your own operations")
	}

	bytes, err := json.Marshal(op)
	if err != nil {
		Logger.Fatal("failed to sign operation because json encoding failed")
	}
	sig := kp.Sign(string(bytes))
	return &SignedOperation{
		Operation: op,
		Signature: sig,
	}
}

func (s *SignedOperation) Verify() bool {
	if s.Operation == nil || reflect.ValueOf(s.Operation).IsNil() {
		return false
	}
	pk, err := ReadPublicKey(s.Operation.GetSigner())
	if err != nil {
		return false
	}
	bytes, err := json.Marshal(s.Operation)
	if err != nil {
		return false
	}
	if !VerifySignature(pk, string(bytes), s.Signature) {
		return false
	}
	if !s.Operation.Verify() {
		return false
	}

	return true
}
