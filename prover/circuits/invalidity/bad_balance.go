package invalidity

import (
	"github.com/consensys/gnark/frontend"
	gmimc "github.com/consensys/gnark/std/hash/mimc"
	"github.com/consensys/linea-monorepo/prover/circuits/internal"
	"github.com/consensys/linea-monorepo/prover/protocol/wizard"
	"github.com/consensys/linea-monorepo/prover/utils"
	"github.com/ethereum/go-ethereum/core/types"
)

// BadBalanceCircuit defines the circuit for the transaction with insufficient balance.
// This means account.Balance is less than Tx.Value + Tx.GasFee
type BadBalanceCircuit struct {
	// Transaction cost = tx.Value + tx.Gas * tx.GasFeeCap
	TxCost frontend.Variable
	// sender address
	TxFromAddress frontend.Variable
	// RLP-encoded payload  prefixed with the type byte. txType || rlp(tx.inner)
	RLPEncodedTx []frontend.Variable
	// hash of the transaction
	TxHash [2]frontend.Variable
	// Keccak verifier circuit
	KeccakH wizard.VerifierCircuit
	// AccountTrie of the sender
	AccountTrie AccountTrie
}

// Define represent the constraints relevant to [BadBalanceCircuit]
func (circuit *BadBalanceCircuit) Define(api frontend.API) error {

	var (
		balance = circuit.AccountTrie.Account.Balance
		hKey    = circuit.AccountTrie.LeafOpening.HKey
	)
	// check that the Account.Balance < TxCost + 1
	api.AssertIsLessOrEqual(balance, api.Add(circuit.TxCost, 1))

	circuit.AccountTrie.Define(api)

	// check that sender address matches the account
	// Hash(FromAddress) == LeafOpening.HKey
	mimc, _ := gmimc.NewMiMC(api)
	mimc.Write(circuit.TxFromAddress)
	api.AssertIsEqual(hKey, mimc.Sum())

	// check that nonce matches the rlp encoding
	// expectedNonce := ExtractNonceFromRLPZk(api, circuit.RLPEncodedTx)
	// api.AssertIsEqual(expectedNonce, nonce)

	// check that rlpEncoding and TxHash are consistent with keccak input/output.
	checkKeccakConsistency(api, circuit.RLPEncodedTx, circuit.TxHash, &circuit.KeccakH)
	// verify keccak computation
	circuit.KeccakH.Verify(api)

	return nil
}

// Allocate the circuit
func (c *BadBalanceCircuit) Allocate(config Config) {
	c.AccountTrie.Allocate(config)
	c.RLPEncodedTx = make([]frontend.Variable, config.MaxRlpByteSize)
	c.KeccakH = *wizard.AllocateWizardCircuit(config.KeccakCompiledIOP, 0)
}

// Assign the circuit from [AssigningInputs]
func (c *BadBalanceCircuit) Assign(assi AssigningInputs) {
	var (
		txCost  = assi.Transaction.Cost()
		balance = assi.AccountTrieInputs.Account.Balance
		signer  = types.NewLondonSigner(assi.Transaction.ChainId())
		a       = signer.Hash(assi.Transaction)
		comp    = assi.KeccakCompiledIOP
		proof   = assi.KeccakProof
		keccak  = *wizard.AssignVerifierCircuit(comp, proof, 0)
	)

	c.TxCost = txCost
	c.TxFromAddress = assi.FromAddress[:]
	c.RLPEncodedTx = make([]frontend.Variable, assi.MaxRlpByteSize)

	// assign the tx hash
	c.TxHash[0] = a[0:16]
	c.TxHash[1] = a[16:32]

	if assi.RlpEncodedTx[0] != 0x02 {
		utils.Panic("only support typed 2 transactions, maybe the rlp is not prefixed with the type byte")
	}
	// assign the rlp encoding
	elements := internal.FromBytesToElements(assi.RlpEncodedTx)
	rlpLen := len(elements)
	if rlpLen > assi.MaxRlpByteSize {
		utils.Panic("rlp encoding is too large: got %d, max %d", rlpLen, assi.MaxRlpByteSize)
	}

	copy(c.RLPEncodedTx, elements)

	for i := len(elements); i < assi.MaxRlpByteSize; i++ {
		c.RLPEncodedTx[i] = 0
	}

	//sanity-check:
	if txCost.Cmp(balance) != 1 {
		utils.Panic("tried to generate a bad-balance proof for a valid transaction")
	}

	//assign the account trie
	c.AccountTrie.Assign(assi.AccountTrieInputs)
	c.KeccakH = keccak
}

func (c *BadBalanceCircuit) ExecutionCtx() []frontend.Variable {
	return []frontend.Variable{c.AccountTrie.MerkleProof.Root}
}
