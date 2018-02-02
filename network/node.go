package network

import (
	"log"

	"coinkit/consensus"
	"coinkit/currency"
	"coinkit/util"
)

// Node is the logical container for everything one node in the network handles.
// Node is not threadsafe.
type Node struct {
	publicKey string
	chain     *consensus.Chain
	queue     *currency.TransactionQueue
}

func NewNode(publicKey string, qs consensus.QuorumSlice) *Node {
	queue := currency.NewTransactionQueue(publicKey)

	return &Node{
		publicKey: publicKey,
		chain:     consensus.NewEmptyChain(publicKey, qs, queue),
		queue:     queue,
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
	if sender == node.publicKey {
		return nil
	}
	switch m := message.(type) {

	case *currency.AccountMessage:
		return nil

	case *util.InfoMessage:
		if m.Account != "" {
			answer, _ := node.queue.Handle(m)
			return answer
		}
		if m.I != 0 {
			return node.chain.Handle(sender, m)
		}
		return nil

	case *currency.TransactionMessage:
		response, updated := node.queue.Handle(m)
		if updated {
			node.chain.ValueStoreUpdated()
		}
		return response

	case *consensus.NominationMessage:
		return node.chain.Handle(sender, m)
	case *consensus.PrepareMessage:
		return node.chain.Handle(sender, m)
	case *consensus.ConfirmMessage:
		return node.chain.Handle(sender, m)
	case *consensus.ExternalizeMessage:
		return node.chain.Handle(sender, m)

	default:
		log.Printf("unrecognized message: %+v", m)
		return nil
	}
}

func (node *Node) OutgoingMessages() []util.Message {
	answer := []util.Message{}
	sharing := node.queue.SharingMessage()
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
