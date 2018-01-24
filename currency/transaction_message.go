package currency

import (
	"sort"

	"coinkit/util"
)

// A TransactionMessage has a list of transactions. Each of the transactions
// is separately signed by the sender, so that a TransactionMessage can be
// used not just to inform the network you would like to make a transaction,
// but also for nodes to share a set of known transaction messages.

type TransactionMessage struct {
	// Should be sorted and non-nil
	Transactions []*SignedTransaction
}

func (m *TransactionMessage) Slot() int {
	return 0
}

func (m *TransactionMessage) MessageType() string {
	return "T"
}

// Orders the transactions
func NewTransactionMessage(ts ...*SignedTransaction) *TransactionMessage {
	sort.Slice(ts, func(i, j int) bool {
		return HighestPriorityFirst(ts[i], ts[j]) < 0
	})

	return &TransactionMessage{
		Transactions: ts,
	}
}

func init() {
	util.RegisterMessageType(&TransactionMessage{})
}
