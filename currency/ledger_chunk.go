package currency

// A LedgerChunk is the information in one block of the blockchain.
type LedgerChunk struct {
	Transactions []*SignedTransaction

	// The state of accounts after these transactions have been processed.
	// This only includes account information for the accounts that are
	// mentioned in the transactions.
	State map[string]*Account
}
