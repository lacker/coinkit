package network

import (
	"github.com/lacker/coinkit/consensus"
	"github.com/lacker/coinkit/data"
	"github.com/lacker/coinkit/util"
)

// Node is the logical container for everything one node in the network handles.
// Node is not threadsafe.
// Everything within Node should be deterministic, for ease of testing. No channels
// or network connections. Database usage is okay though.
type Node struct {
	publicKey util.PublicKey
	chain     *consensus.Chain
	queue     *data.OperationQueue
	database  *data.Database
	slot      int
}

// NewNode uses the standard account initialization settings from data.Airdrop
func NewNode(
	publicKey util.PublicKey, qs consensus.QuorumSlice, db *data.Database) *Node {
	node := newNodeWithAccounts(publicKey, qs, db, data.Airdrop)

	// We check on startup that our block history matches our current account data
	if db != nil {
		err := db.CheckAccountsMatchBlocks()
		if err != nil {
			util.Printf("check failed: accounts do not match blocks")
			return nil
		}
	}

	return node
}

// Creates a node for a blockchain that starts out with the provided accounts airdropped.
func newNodeWithAccounts(publicKey util.PublicKey, qs consensus.QuorumSlice,
	db *data.Database, accounts []*data.Account) *Node {

	var slot int
	var queue *data.OperationQueue
	var chain *consensus.Chain

	// Figure out the current slot
	if db != nil {
		last := db.LastBlock()
		if last != nil {
			// We are resuming where we left off, based on the database
			slot = last.Slot + 1
			queue = data.NewOperationQueue(publicKey, db, slot)
			chain = consensus.NewChain(
				publicKey, qs, queue, last.ExternalizeMessage())
		}
	}

	if chain == nil {
		// This is initial startup, so do the airdrop
		slot = 1
		queue = data.NewOperationQueue(publicKey, db, slot)
		chain = consensus.NewEmptyChain(publicKey, qs, queue)
		for _, account := range accounts {
			queue.SetBalance(account.Owner, account.Balance)
		}
		if db != nil {
			db.Commit()
		}
	}

	return &Node{
		publicKey: publicKey,
		queue:     queue,
		database:  db,
		chain:     chain,
		slot:      slot,
	}
}

// Creates a new memory-only node where nobody has any money
func newTestingNode(publicKey util.PublicKey, qs consensus.QuorumSlice) *Node {
	return newNodeWithAccounts(publicKey, qs, nil, []*data.Account{})
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
		node.Handle(sender, m.O)
		node.Handle(sender, m.E)
		return nil, false

	case *data.OperationMessage:
		if node.queue.HandleOperationMessage(m) {
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
		util.Logger.Printf("Node received unexpected message: %+v", m)
		return nil, false
	}
}

// A helper to handle the messages
func (node *Node) handleChainMessage(sender string, message util.Message) (util.Message, bool) {
	response, hasResponse := node.chain.Handle(sender, message)

	if node.chain.Slot() > node.Slot() {
		// We have advanced.
		node.slot += 1
	}

	if !hasResponse {
		return nil, false
	}

	externalize, ok := response.(*consensus.ExternalizeMessage)
	if !ok {
		return response, true
	}

	// Augment externalize messages into history messages
	om := node.queue.OldChunkMessage(externalize.I)
	return &HistoryMessage{
		O: om,
		E: externalize,
		I: externalize.I,
	}, true
}

func (node *Node) OutgoingMessages() []util.Message {
	answer := []util.Message{}
	sharing := node.queue.OperationMessage()
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
