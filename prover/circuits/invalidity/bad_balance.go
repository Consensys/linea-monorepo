package invalidity

import (
	"github.com/consensys/gnark/frontend"
	"github.com/consensys/linea-monorepo/prover/utils"
)

// BadBalanceCircuit defines the circuit for the transaction with insufficient balance.
// This means account.Balance is less than Tx.Value + Tx.GasFee
type BadBalanceCircuit struct {
	// Transaction cost = tx.Value + tx.Gas * tx.GasFeeCap
	TxCost frontend.Variable
	// sender address
	// TxFromAddress frontend.Variable
	AccountTrie AccountTrie
}

// Define represent the constraints relevant to [BadBalanceCircuit]
func (circuit *BadBalanceCircuit) Define(api frontend.API) error {

	var (
		balance = circuit.AccountTrie.Account.Balance
	)
	// check that the Account.Balance < TxCost + 1
	api.AssertIsLessOrEqual(balance, api.Add(circuit.TxCost, 1))

	circuit.AccountTrie.Define(api)

	//@azam check that tx fields are related to  Tx.Hash  and then in the interconnection we show that
	//FTx.Hash and FromAddress  = HKey is included in the RollingHash

	return nil
}

// Allocate the circuit
func (c *BadBalanceCircuit) Allocate(config Config) {
	c.AccountTrie.Allocate(config)
}

// Assign the circuit from [AssigningInputs]
func (c *BadBalanceCircuit) Assign(assi AssigningInputs) {
	var (
		txCost  = assi.Transaction.Cost()
		balance = assi.AccountTrieInputs.Account.Balance
	)
	*c = BadBalanceCircuit{
		TxCost: txCost,
	}

	//sanity-check:
	if txCost.Cmp(balance) != 1 {
		utils.Panic("tried to generate a bad-balance proof for a valid transaction")
	}

	c.AccountTrie.Assign(assi.AccountTrieInputs)

}

func (c *BadBalanceCircuit) ExecutionCtx() []frontend.Variable {
	return []frontend.Variable{c.AccountTrie.MerkleProof.Root}
}
