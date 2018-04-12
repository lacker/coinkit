package data

import (
	"testing"

	"github.com/lacker/coinkit/util"
)

func TestOperationMessages(t *testing.T) {
	kp1 := util.NewKeyPairFromSecretPhrase("key pair 1")
	kp2 := util.NewKeyPairFromSecretPhrase("key pair 2")
	t1 := &SendOperation{
		Sequence: 1,
		Amount:   100,
		Fee:      2,
		Signer:   kp1.PublicKey().String(),
		To:       kp2.PublicKey().String(),
	}
	t2 := &SendOperation{
		Sequence: 1,
		Amount:   50,
		Fee:      2,
		Signer:   kp2.PublicKey().String(),
		To:       kp1.PublicKey().String(),
	}
	s1 := util.NewSignedOperation(t1, kp1)
	s2 := util.NewSignedOperation(t2, kp2)
	message := NewOperationMessage(s1, s2)

	m := util.EncodeThenDecodeMessage(message).(*OperationMessage)
	if len(m.Operations) != 2 {
		t.Fatal("expected len m.Operations to be 2")
	}
	if !m.Operations[0].Verify() {
		t.Fatal("expected m.Operations[0].Verify()")
	}
	if !m.Operations[1].Verify() {
		t.Fatal("expected m.Operations[1].Verify()")
	}

}
