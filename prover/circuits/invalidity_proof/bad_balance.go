package invalidity_proof

import (
	"github.com/consensys/gnark/frontend"
)

// BadBalanceCircuit defines the circuit for the transaction with insufficient balance.
// This means account.Balance is less than Tx.Value + Tx.GasFee
type BadBalanceCircuit struct {
	// Transaction value
	TxValue frontend.Variable
	// Transaction GasFee
	TxGasFee frontend.Variable
	// sender address
	TxFromAddress frontend.Variable
	AccountTrie   AccountTrie
}

// Define represent the constraints relevant to [BadBalanceCircuit]
func (circuit *BadBalanceCircuit) Define(api frontend.API) error {

	var (
		balance = circuit.AccountTrie.Account.Balance
	)
	// check that the Account.Balance < TxValue + TxGasFee + 1
	api.AssertIsLessOrEqual(balance, api.Add(circuit.TxValue, circuit.TxGasFee, 1))

	circuit.AccountTrie.Define(api)

	//@azam check that tx fields are related to  Tx.Hash  and then in the interconnection we show that
	//FTx.Hash is included in the RollingHash

	//@Azam check that FromAddress is consistent with the account
	return nil
}

// Allocate the circuit
func (c *BadBalanceCircuit) Allocate(config Config) {
	c.AccountTrie.Allocate(config)
}

// Assign the circuit from [AssigningInputs]
func (c *BadBalanceCircuit) Assign(assi AssigningInputs) {

	*c = BadBalanceCircuit{
		TxValue:       assi.Transaction.Value(),
		TxGasFee:      assi.Transaction.Cost(),
		TxFromAddress: assi.FromAddress,
	}

	c.AccountTrie.Assign(assi.AccountTrieInputs)

}
