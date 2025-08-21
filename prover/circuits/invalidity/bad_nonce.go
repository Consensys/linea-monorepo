package invalidity

import (
	"github.com/consensys/gnark/frontend"
	"github.com/consensys/linea-monorepo/prover/circuits/pi-interconnection/keccak"
	"github.com/consensys/linea-monorepo/prover/utils"
)

// BadNonceCircuit defines the circuit for the transaction with a bad nonce.
type BadNonceCircuit struct {
	// Transaction payload.
	TxPayload TxPayloadGnark
	// payload hash
	PayloadHash frontend.Variable
	// sender address
	TxFromAddress frontend.Variable
	// AccountTrie of the sender
	AccountTrie AccountTrie
}

// Define represents the constraints relevant to [BadNonceCircuit]
func (circuit *BadNonceCircuit) Define(api frontend.API) error {

	var (
		nonce   = circuit.TxPayload.Nonce
		account = circuit.AccountTrie.Account
		diff    = api.Sub(nonce, api.Add(account.Nonce, 1))
	)

	// check that the nonce != Account.Nonce + 1
	api.AssertIsDifferent(diff, 0)

	// check that the account matches the state root
	circuit.AccountTrie.Define(api)

	// check that sender address matches the account
	api.AssertIsEqual(circuit.AccountTrie.LeafOpening.HKey, circuit.TxFromAddress)

	// check that nonce matches payloadHash
	// @azam I stoped here.
	hsh := keccak.NewHasher(api, MaxNbKeccakF)
	circuit.TxPayload.Sum(api, hsh)

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
		TxPayload: txNonce,
	}

	// sanity-check
	if txNonce == uint64(acNonce+1) {
		utils.Panic("tried to generate a bad-nonce proof for a valid transaction")
	}

	c.AccountTrie.Assign(assi.AccountTrieInputs)
}

func (c *BadNonceCircuit) ExecutionCtx() []frontend.Variable {
	return []frontend.Variable{c.AccountTrie.MerkleProof.Root}
}
