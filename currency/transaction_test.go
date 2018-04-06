package currency

import (
	"testing"
)

func TestTestTransactionVerifies(t *testing.T) {
	st := makeTestTransaction(0)
	if !st.Verify() {
		t.Fatal("should verify")
	}
}
