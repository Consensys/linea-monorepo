package invalidity

import (
	"github.com/consensys/gnark/frontend"
	"github.com/consensys/linea-monorepo/prover/circuits/pi-interconnection/keccak"
	"github.com/consensys/linea-monorepo/prover/utils"
)

// BadNonceCircuit defines the circuit for the transaction with a bad nonce.
type BadNonceCircuit struct {
	//  Transaction payload.
	TxNonce frontend.Variable
	// RLP-encoded transaction as a slice of bytes
	RLPEncodedTx []frontend.Variable
	// payload hash
	PayloadHash frontend.Variable
	// sender address
	TxFromAddress frontend.Variable
	// AccountTrie of the sender
	AccountTrie AccountTrie
	// hash of the transaction
	TxHash frontend.Variable
}

// Define represents the constraints relevant to [BadNonceCircuit]
func (circuit *BadNonceCircuit) Define(api frontend.API) error {

	var (
		nonce   = circuit.TxNonce
		account = circuit.AccountTrie.Account
		diff    = api.Sub(nonce, api.Add(account.Nonce, 1))
		keccakH keccak.StrictHasherCircuit
		hshK    = keccakH.NewHasher(api)
	)

	// check that the nonce != Account.Nonce + 1
	api.AssertIsDifferent(diff, 0)

	// check that the account matches the state root
	circuit.AccountTrie.Define(api)

	// check that sender address matches the account
	api.AssertIsEqual(circuit.AccountTrie.LeafOpening.HKey, circuit.TxFromAddress)

	// check that nonce matches the rlp encoding
	expectedNonce := ExtractNonceFromRLPZk(api, circuit.RLPEncodedTx)
	api.AssertIsEqual(expectedNonce, nonce)

	// check that rlp encoding matches the transaction hash
	api.AssertIsEqual(Sum(api, &hshK, circuit.RLPEncodedTx), circuit.TxHash)

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

func (c *BadNonceCircuit) ExecutionCtx() []frontend.Variable {
	return []frontend.Variable{c.AccountTrie.MerkleProof.Root}
}
