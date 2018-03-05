package network

import (
	"log"

	"coinkit/consensus"
	"coinkit/currency"
	"coinkit/data"
	"coinkit/util"
)

// Node is the logical container for everything one node in the network handles.
// Node is not threadsafe.
// Everything within Node should be deterministic, for ease of testing. No channels
// or network connections. Database usage is okay though.
type Node struct {
	publicKey util.PublicKey
	chain     *consensus.Chain
	queue     *currency.TransactionQueue
	database  *data.Database
	slot      int
}

func NewNode(
	publicKey util.PublicKey, qs consensus.QuorumSlice, db *data.Database) *Node {
	queue := currency.NewTransactionQueue(publicKey)

	node := &Node{
		publicKey: publicKey,
		queue:     queue,
		database:  db,
		chain:     consensus.NewEmptyChain(publicKey, qs, queue),
		slot:      1,
	}

	if db != nil {
		node.slot = db.ForBlocks(func(b *data.Block) {
			m := b.ExternalizeMessage(qs)
			node.chain.AlreadyExternalized(m)
			node.queue.FinalizeChunk(b.Chunk)
		})
	}

	return node
}

// Slot() returns the slot this node is currently working on
func (node *Node) Slot() int {
	return node.slot
}

// Handle handles an incoming message.
// It may return a message to be sent back to the original sender
// The bool flag tells whether it has a response or not.
func (node *Node) Handle(sender string, message util.Message) (util.Message, bool) {
	if sender == node.publicKey.String() {
		return nil, false
	}
	switch m := message.(type) {

	case *HistoryMessage:
		node.Handle(sender, m.T)
		node.Handle(sender, m.E)
		return nil, false

	case *currency.AccountMessage:
		return nil, false

	case *util.InfoMessage:
		if m.Account != "" {
			answer := node.queue.HandleInfoMessage(m)
			return answer, answer != nil
		}
		if m.I != 0 {
			answer, ok := node.chain.Handle(sender, m)
			return answer, ok
		}
		return nil, false

	case *currency.TransactionMessage:
		if node.queue.HandleTransactionMessage(m) {
			node.chain.ValueStoreUpdated()
		}
		return nil, false

	case *consensus.NominationMessage:
		answer, ok := node.handleChainMessage(sender, m)
		return answer, ok
	case *consensus.PrepareMessage:
		answer, ok := node.handleChainMessage(sender, m)
		return answer, ok
	case *consensus.ConfirmMessage:
		answer, ok := node.handleChainMessage(sender, m)
		return answer, ok
	case *consensus.ExternalizeMessage:
		answer, ok := node.handleChainMessage(sender, m)
		return answer, ok

	default:
		log.Printf("unrecognized message: %+v", m)
		return nil, false
	}
}

// A helper to handle the messages
func (node *Node) handleChainMessage(sender string, message util.Message) (util.Message, bool) {
	response, hasResponse := node.chain.Handle(sender, message)

	if node.chain.Slot() > node.Slot() {
		// We have advanced.
		node.slot += 1

		if node.database != nil {
			// Let's save the old block.
			last := node.chain.GetLast()
			chunk := node.queue.OldChunk(last.I)
			block := &data.Block{
				Slot:  last.I,
				C:     last.Cn,
				H:     last.Hn,
				Chunk: chunk,
			}
			err := node.database.SaveBlock(block)
			if err != nil {
				panic(err)
			}
		}
	}

	if !hasResponse {
		return nil, false
	}

	externalize, ok := response.(*consensus.ExternalizeMessage)
	if !ok {
		return response, true
	}

	// Augment externalize messages into history messages
	t := node.queue.OldChunkMessage(externalize.I)
	return &HistoryMessage{
		T: t,
		E: externalize,
		I: externalize.I,
	}, true
}

func (node *Node) OutgoingMessages() []util.Message {
	answer := []util.Message{}
	sharing := node.queue.TransactionMessage()
	if sharing != nil {
		answer = append(answer, sharing)
	}
	for _, m := range node.chain.OutgoingMessages() {
		answer = append(answer, m)
	}
	return answer
}

func (node *Node) Stats() {
	node.chain.Stats()
	node.queue.Stats()
}

func (node *Node) Log() {
	node.chain.Log()
	node.queue.Log()
}
