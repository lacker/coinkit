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

// Creates a node for a blockchain that starts with one mint account having a balance.
func NewNodeWithMint(publicKey util.PublicKey, qs consensus.QuorumSlice,
	db *data.Database, mint util.PublicKey, balance uint64) *Node {

	queue := data.NewOperationQueue(publicKey)
	if balance != 0 {
		queue.SetBalance(mint.String(), balance)
	}

	node := &Node{
		publicKey: publicKey,
		queue:     queue,
		database:  db,
		chain:     consensus.NewEmptyChain(publicKey, qs, queue),
		slot:      1,
	}

	if db != nil {
		loaded := db.ForBlocks(func(b *data.Block) {
			m := b.ExternalizeMessage(qs)
			node.chain.AlreadyExternalized(m)
			node.queue.FinalizeChunk(b.Chunk)
		})
		util.Logger.Printf("loaded %d old blocks from the database", loaded)
		node.slot = loaded + 1
	}

	return node
}

func NewNode(
	publicKey util.PublicKey, qs consensus.QuorumSlice, db *data.Database) *Node {
	var invalid util.PublicKey
	return NewNodeWithMint(publicKey, qs, db, invalid, 0)
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

	case *util.InfoMessage:
		if node.database == nil {
			util.Logger.Fatal("InfoMessages require a database to fulfill")
		}

		// TODO: fulfill all InfoMessages from the database, instead of
		// doing this stuff below. Then Node could just not handle InfoMessages
		if m.Account != "" {
			answer := node.queue.HandleInfoMessage(m)
			if answer == nil {
				util.Logger.Fatal("answer was nil")
			}
			return answer, answer != nil
		}

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
			err := node.database.InsertBlock(block)
			if err != nil {
				panic(err)
			}
			node.database.Commit()
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
