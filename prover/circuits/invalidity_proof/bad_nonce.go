package invalidity_proof

import (
	"github.com/consensys/gnark/frontend"
	"github.com/consensys/linea-monorepo/prover/utils"
)

// BadNonceCircuit defines the circuit for the transaction with a bad nonce.
type BadNonceCircuit struct {
	// Transaction Nonce
	TxNonce frontend.Variable
	// sender address
	// TxFromAddress frontend.Variable
	// AccountTrie of the sender
	AccountTrie AccountTrie
}

// Define represent the constraints relevant to [BadNonceCircuit]
func (circuit *BadNonceCircuit) Define(api frontend.API) error {

	var (
		account = circuit.AccountTrie.Account
		diff    = api.Sub(circuit.TxNonce, api.Add(account.Nonce, 1))
	)

	// check that the FTx.Nonce = Account.Nonce + 1
	api.AssertIsDifferent(diff, 0)

	//@azam check that tx fields are related to  Tx.Hash  and then in the interconnection we show that
	//FTx.Hash and FromAddress  = HKey is included in the RollingHash
	return nil
}

// Allocate the circuit
func (c *BadNonceCircuit) Allocate(config Config) {
	c.AccountTrie.Allocate(config)
}

// Assign the circuit from [AssigningInputs]
func (c *BadNonceCircuit) Assign(assi AssigningInputs) {

	var (
		txNonce = assi.Transaction.Nonce()
		acNonce = assi.AccountTrieInputs.Account.Nonce
	)
	*c = BadNonceCircuit{
		TxNonce: txNonce,
	}

	// sanity-check
	if txNonce == uint64(acNonce+1) {
		utils.Panic("tried to generate a bad-nonce proof for a valid transaction")
	}

	c.AccountTrie.Assign(assi.AccountTrieInputs)
}

func (c *BadNonceCircuit) ExecutionCtx() frontend.Variable {
	return c.AccountTrie.MerkleProof.Root
}
