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
type Node struct {
	publicKey util.PublicKey
	chain     *consensus.Chain
	queue     *currency.TransactionQueue
	store     *data.DataStore
}

func NewNode(publicKey util.PublicKey, qs consensus.QuorumSlice) *Node {
	queue := currency.NewTransactionQueue(publicKey)

	return &Node{
		publicKey: publicKey,
		chain:     consensus.NewEmptyChain(publicKey, qs, queue),
		queue:     queue,
		store:     data.NewDataStore(),
	}
}

// Slot() returns the slot this node is currently working on
func (node *Node) Slot() int {
	return node.chain.Slot()
}

// Handle handles an incoming message.
// It may return a message to be sent back to the original sender, or it may
// just return nil if it has no particular response.
func (node *Node) Handle(sender string, message util.Message) util.Message {
	if sender == node.publicKey.String() {
		return nil
	}
	switch m := message.(type) {

	case *data.DataMessage:
		log.Printf("XXX there is a data message: %+v", m)
		return node.store.Handle(m)

	case *HistoryMessage:
		node.Handle(sender, m.T)
		node.Handle(sender, m.E)
		return nil

	case *currency.AccountMessage:
		return nil

	case *util.InfoMessage:
		if m.Account != "" {
			return node.queue.HandleInfoMessage(m)
		}
		if m.I != 0 {
			return node.chain.Handle(sender, m)
		}
		return nil

	case *currency.TransactionMessage:
		if node.queue.HandleTransactionMessage(m) {
			node.chain.ValueStoreUpdated()
		}
		return nil

	case *consensus.NominationMessage:
		return node.handleChainMessage(sender, m)
	case *consensus.PrepareMessage:
		return node.handleChainMessage(sender, m)
	case *consensus.ConfirmMessage:
		return node.handleChainMessage(sender, m)
	case *consensus.ExternalizeMessage:
		return node.handleChainMessage(sender, m)

	default:
		log.Printf("unrecognized message: %+v", m)
		return nil
	}
}

// A helper to handle the messages
func (node *Node) handleChainMessage(sender string, message util.Message) util.Message {
	response := node.chain.Handle(sender, message)

	externalize, ok := response.(*consensus.ExternalizeMessage)
	if !ok {
		return response
	}

	// Augment externalize messages into history messages
	t := node.queue.OldChunkMessage(externalize.I)
	return &HistoryMessage{
		T: t,
		E: externalize,
		I: externalize.I,
	}
}

func (node *Node) OutgoingMessages() []util.Message {
	answer := []util.Message{}
	sharing := node.queue.SharingMessage()
	if sharing != nil {
		answer = append(answer, sharing)
	}
	d := node.store.OutgoingMessage()
	if d != nil {
		answer = append(answer, d)
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
