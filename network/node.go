package network

import (
	"coinkit/consensus"
	"coinkit/currency"
)

// A Node is a logical container for everything one node in the network handles.
type Node struct {
	publicKey string
	chain *consensus.Chain
	queue *currency.TransactionQueue
}

func NewNode(publicKey string, qs consensus.QuorumSlice) *Node {
	return &Node{
		publicKey: publicKey,
		chain: consensus.NewEmptyChain(publicKey, qs),
		queue: currency.NewTransactionQueue,
	}
}

// Handle handles an incoming message.
// It may return a message to be sent back to the original sender, or it may
// just return nil if it has no particular response.
func (node *Node) Handle(sender string, message util.Message) util.Message {
	if sender == node.publicKey {
		return
	}
	switch m := message.(type) {
	case *TransactionMessage:
		node.queue.Handle(sender, m)
	case *NominationMessage:
		node.chain.Handle(sender, m)
	case *PrepareMessage:
		node.chain.Handle(sender, m)
	case *ConfirmMessage:
		node.chain.Handle(sender, m)
	case *ExternalizeMessage:
		node.chain.Handle(sender, m)
	default:
		log.Printf("unrecognized message: %+v", m)
	}
}

func (node *Node) OutgoingMessages() []util.Message {
	answer := node.chain.OutgoingMessages()
	sharing := node.queue.SharingMessage()
	if sharing != nil {
		answer = append(answer, sharing)
	}
	return sharing
}
