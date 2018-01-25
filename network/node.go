package network

import (
	"log"
	
	"coinkit/consensus"
	"coinkit/currency"
	"coinkit/util"
)

// A Node is a logical container for everything one node in the network handles.
type Node struct {
	publicKey string
	chain *consensus.Chain
	queue *currency.TransactionQueue
}

func NewNode(publicKey string, qs consensus.QuorumSlice) *Node {
	queue := currency.NewTransactionQueue()
	
	return &Node{
		publicKey: publicKey,
		chain: consensus.NewEmptyChain(publicKey, qs, queue),
		queue: queue,
	}
}

// Handle handles an incoming message.
// It may return a message to be sent back to the original sender, or it may
// just return nil if it has no particular response.
func (node *Node) Handle(sender string, message util.Message) util.Message {
	if sender == node.publicKey {
		return nil
	}
	switch m := message.(type) {
	case *currency.TransactionMessage:
		node.queue.Handle(m)
		return nil
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
