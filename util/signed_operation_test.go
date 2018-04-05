package util

import (
	"testing"
)

func TestSignedOperation(t *testing.T) {
	kp := NewKeyPairFromSecretPhrase("yo")
	op := &TestingOperation{
		Number: 8,
		Signer: kp.PublicKey().String(),
	}
	so := NewSignedOperation(op, kp)
	if !so.Verify() {
		t.Fatal("so should Verify")
	}
}
