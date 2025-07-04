package invalidity_proof

import (
	"github.com/consensys/gnark/frontend"
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

	//@azam check that FTx.Nonce is related to  FTx.Hash  and then in the interconnection we show that
	//FTx.Hash is included in the RollingHash
	return nil
}

// Allocate the circuit
func (c *BadNonceCircuit) Allocate(config Config) {
	c.AccountTrie.Allocate(config)
}

// Assign the circuit from [AssigningInputs]
func (c *BadNonceCircuit) Assign(assi AssigningInputs) {

	*c = BadNonceCircuit{
		TxNonce: assi.Transaction.Nonce(),
	}

	c.AccountTrie.Assign(assi.AccountTrieInputs)
}
