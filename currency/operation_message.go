package currency

import (
	"fmt"
	"sort"
	"strings"

	"github.com/lacker/coinkit/consensus"
	"github.com/lacker/coinkit/util"
)

// An OperationMessage has a list of operations. Each of the operations
// is separately signed by the sender, so that an OperationMessage can be
// used not just to inform the network you would like to issue an opertion,
// but also for nodes to share a set of known pending operations.

type OperationMessage struct {
	// Should be sorted and non-nil
	// Only contains operations that were not previously sent
	Operations []*util.SignedOperation

	// Contains any chunks that might be in the immediately following messages
	Chunks map[consensus.SlotValue]*LedgerChunk
}

func (m *OperationMessage) Slot() int {
	return 0
}

func (m *OperationMessage) MessageType() string {
	return "Operation"
}

func (m *OperationMessage) String() string {
	cnames := []string{}
	for name, _ := range m.Chunks {
		cnames = append(cnames, util.Shorten(string(name)))
	}
	return fmt.Sprintf("op %s chunks (%s)",
		util.StringifyOperations(m.Operations), strings.Join(cnames, ","))
}

// Orders the transactions
func NewOperationMessage(ops ...*util.SignedOperation) *OperationMessage {
	sort.Slice(ops, func(i, j int) bool {
		return util.HighestFeeFirst(ops[i], ops[j]) < 0
	})

	return &OperationMessage{
		Operations: ops,
		Chunks:     make(map[consensus.SlotValue]*LedgerChunk),
	}
}

func init() {
	util.RegisterMessageType(&OperationMessage{})
}
