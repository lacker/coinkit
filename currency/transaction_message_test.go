package currency

import (
	"testing"

	"github.com/lacker/coinkit/util"
)

func TestTransactionMessages(t *testing.T) {
	kp1 := util.NewKeyPairFromSecretPhrase("key pair 1")
	kp2 := util.NewKeyPairFromSecretPhrase("key pair 2")
	t1 := Transaction{
		Sequence: 1,
		Amount:   100,
		Fee:      2,
		Signer:   kp1.PublicKey().String(),
		To:       kp2.PublicKey().String(),
	}
	t2 := Transaction{
		Sequence: 1,
		Amount:   50,
		Fee:      2,
		Signer:   kp2.PublicKey().String(),
		To:       kp1.PublicKey().String(),
	}
	s1 := t1.SignWith(kp1)
	s2 := t2.SignWith(kp2)
	message := NewTransactionMessage(s1, s2)

	m := util.EncodeThenDecodeMessage(message).(*TransactionMessage)
	if len(m.Transactions) != 2 {
		t.Fatal("expected len m.Transactions to be 2")
	}
	if !m.Transactions[0].Verify() {
		t.Fatal("expected m.Transactions[0].Verify()")
	}
	if !m.Transactions[1].Verify() {
		t.Fatal("expected m.Transactions[1].Verify()")
	}

}
