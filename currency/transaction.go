package currency

import (
	"fmt"

	"github.com/lacker/coinkit/util"
)

type Transaction struct {
	// Who is sending this money
	Signer string

	// The sequence number for this transaction
	Sequence uint32

	// Who is receiving this money
	To string

	// The amount of currency to transfer
	Amount uint64

	// How much the sender is willing to pay to get this transfer registered
	// This is on top of the amount
	Fee uint64
}

func (t *Transaction) String() string {
	return fmt.Sprintf("send %d from %s -> %s, seq %d fee %d",
		t.Amount, util.Shorten(t.Signer), util.Shorten(t.To), t.Sequence, t.Fee)
}

func (t *Transaction) OperationType() string {
	return "Transaction"
}

func (t *Transaction) GetSigner() string {
	return t.Signer
}

func (t *Transaction) GetFee() uint64 {
	return t.Fee
}

func (t *Transaction) GetSequence() uint32 {
	return t.Sequence
}

func (t *Transaction) Verify() bool {
	if _, err := util.ReadPublicKey(t.To); err != nil {
		return false
	}
	return true
}

func makeTestTransaction(n int) *util.SignedOperation {
	kp := util.NewKeyPairFromSecretPhrase(fmt.Sprintf("blorp %d", n))
	dest := util.NewKeyPairFromSecretPhrase("destination")
	t := &Transaction{
		Signer:   kp.PublicKey().String(),
		Sequence: 1,
		To:       dest.PublicKey().String(),
		Amount:   uint64(n),
		Fee:      uint64(n),
	}
	return util.NewSignedOperation(t, kp)
}

func init() {
	util.RegisterOperationType(&Transaction{})
}
