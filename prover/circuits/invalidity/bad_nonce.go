package invalidity

import (
	"github.com/consensys/gnark/frontend"
	gmimc "github.com/consensys/gnark/std/hash/mimc"
	"github.com/consensys/linea-monorepo/prover/circuits/internal"
	"github.com/consensys/linea-monorepo/prover/protocol/wizard"
	"github.com/consensys/linea-monorepo/prover/utils"
	"github.com/ethereum/go-ethereum/core/types"
)

// BadNonceCircuit defines the circuit for the transaction with a bad nonce.
type BadNonceCircuit struct {
	//  Transaction payload.
	TxNonce frontend.Variable
	// RLP-encoded payload  prefixed with the type byte. txType || rlp(tx.inner)
	RLPEncodedTx []frontend.Variable
	// sender address
	TxFromAddress frontend.Variable
	// AccountTrie of the sender
	AccountTrie AccountTrie
	// hash of the transaction
	TxHash [2]frontend.Variable
	// Keccak verifier circuit
	KeccakH wizard.VerifierCircuit
}

// Define represents the constraints relevant to [BadNonceCircuit]
func (circuit *BadNonceCircuit) Define(api frontend.API) error {

	var (
		nonce   = circuit.TxNonce
		account = circuit.AccountTrie.Account
		diff    = api.Sub(nonce, api.Add(account.Nonce, 1))
		hKey    = circuit.AccountTrie.LeafOpening.HKey
	)

	// check that the nonce != Account.Nonce + 1
	api.AssertIsDifferent(diff, 0)

	// check that the account matches the state root
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
func (cir *BadNonceCircuit) Allocate(config Config) {
	// allocate the account trie
	cir.AccountTrie.Allocate(config)
	cir.KeccakH = *wizard.AllocateWizardCircuit(config.KeccakCompiledIOP, 0)
	// allocate the RLPEncodedTx to have a fixed size
	cir.RLPEncodedTx = make([]frontend.Variable, config.MaxRlpByteSize)

}

// Assign the circuit from [AssigningInputs], circuit is reinitialized
func (cir *BadNonceCircuit) Assign(assi AssigningInputs) {

	var (
		txNonce = assi.Transaction.Nonce()
		acNonce = assi.AccountTrieInputs.Account.Nonce
		signer  = types.NewLondonSigner(assi.Transaction.ChainId())
		a       = signer.Hash(assi.Transaction).Bytes()
		comp    = assi.KeccakCompiledIOP
		proof   = assi.KeccakProof
		keccak  = *wizard.AssignVerifierCircuit(comp, proof, 0)
	)

	cir.TxNonce = txNonce
	cir.TxFromAddress = assi.FromAddress[:]
	cir.RLPEncodedTx = make([]frontend.Variable, assi.MaxRlpByteSize)

	// assign the tx hash
	cir.TxHash[0] = a[0:16]
	cir.TxHash[1] = a[16:32]

	if assi.RlpEncodedTx[0] != 0x02 {
		utils.Panic("only support typed 2 transactions, maybe the rlp is not prefixed with the type byte")
	}
	// assign the rlp encoding
	elements := internal.FromBytesToElements(assi.RlpEncodedTx)
	rlpLen := len(elements)
	if rlpLen > assi.MaxRlpByteSize {
		utils.Panic("rlp encoding is too large: got %d, max %d", rlpLen, assi.MaxRlpByteSize)
	}

	copy(cir.RLPEncodedTx, elements)

	for i := len(elements); i < assi.MaxRlpByteSize; i++ {
		cir.RLPEncodedTx[i] = 0
	}

	// sanity-check
	if txNonce == uint64(acNonce+1) {
		utils.Panic("tried to generate a bad-nonce proof for a valid transaction")
	}

	// assign the account trie
	cir.AccountTrie.Assign(assi.AccountTrieInputs)
	cir.KeccakH = keccak
}

func (c *BadNonceCircuit) ExecutionCtx() []frontend.Variable {
	return []frontend.Variable{c.AccountTrie.MerkleProof.Root}
}
