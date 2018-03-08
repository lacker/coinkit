package currency

import (
	"encoding/json"
	"testing"

	"github.com/lacker/coinkit/util"
)

func TestTestTransactionVerifies(t *testing.T) {
	st := makeTestTransaction(0)
	if !st.Verify() {
		t.Fatal("should verify")
	}
}

func TestTransactionVerification(t *testing.T) {
	kp1 := util.NewKeyPairFromSecretPhrase("bloop1")
	kp2 := util.NewKeyPairFromSecretPhrase("bloop2")
	tr := &Transaction{
		From:     kp1.PublicKey().String(),
		Sequence: 1,
		To:       kp2.PublicKey().String(),
		Amount:   uint64(10),
		Fee:      uint64(10),
	}
	if !tr.SignWith(kp1).Verify() {
		t.Fatal("normal verification should work")
	}
	bytes, _ := json.Marshal(tr)
	st := &SignedTransaction{
		Transaction: tr,
		Signature:   kp2.Sign(string(bytes)),
	}
	if st.Verify() {
		t.Fatal("the sender should have to sign")
	}
	tr.To = "invalidAddress"
	if tr.SignWith(kp1).Verify() {
		t.Fatal("address should be valid to get verified")
	}
}
