package currency

import (
	"fmt"
	"sort"
	"strings"

	"github.com/lacker/coinkit/consensus"
	"github.com/lacker/coinkit/util"
)

// A TransactionMessage has a list of transactions. Each of the transactions
// is separately signed by the sender, so that a TransactionMessage can be
// used not just to inform the network you would like to make a transaction,
// but also for nodes to share a set of known transaction messages.

type TransactionMessage struct {
	// Should be sorted and non-nil
	// Only contains transactions that were not previously sent
	Transactions []*SignedTransaction

	// Contains any chunks that might be in the immediately following messages
	Chunks map[consensus.SlotValue]*LedgerChunk
}

func (m *TransactionMessage) Slot() int {
	return 0
}

func (m *TransactionMessage) MessageType() string {
	return "T"
}

func (m *TransactionMessage) String() string {
	cnames := []string{}
	for name, _ := range m.Chunks {
		cnames = append(cnames, util.Shorten(string(name)))
	}
	return fmt.Sprintf("trans %s chunks (%s)",
		StringifyTransactions(m.Transactions), strings.Join(cnames, ","))
}

// Orders the transactions
func NewTransactionMessage(ts ...*SignedTransaction) *TransactionMessage {
	sort.Slice(ts, func(i, j int) bool {
		return HighestPriorityFirst(ts[i], ts[j]) < 0
	})

	return &TransactionMessage{
		Transactions: ts,
		Chunks: make(map[consensus.SlotValue]*LedgerChunk),
	}
}

func init() {
	util.RegisterMessageType(&TransactionMessage{})
}
