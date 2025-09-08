package invalidity

import (
	"github.com/consensys/gnark/frontend"
	gmimc "github.com/consensys/gnark/std/hash/mimc"
	"github.com/consensys/linea-monorepo/prover/circuits/internal"
	"github.com/consensys/linea-monorepo/prover/protocol/wizard"
	"github.com/consensys/linea-monorepo/prover/utils"
	"github.com/ethereum/go-ethereum/crypto"
)

// BadNonceBalanceCircuit defines the circuit for the transaction with a bad nonce or insufficient balance.
// The circuit does not do any range check over the transaction field,
// as it is supposed to be done by the contract.
type BadNonceBalanceCircuit struct {
	//  Transaction payload.
	TxNonce frontend.Variable
	// transaction cost
	TxCost frontend.Variable `gnark:",secret"`
	// RLP-encoded payload  prefixed with the type byte. txType || rlp(tx.inner)
	RLPEncodedTx []frontend.Variable `gnark:",secret"`
	// sender address
	TxFromAddress frontend.Variable `gnark:",secret"`
	// AccountTrie of the sender
	AccountTrie AccountTrie `gnark:",secret"`
	// hash of the transaction
	TxHash [2]frontend.Variable `gnark:",secret"`
	// Keccak verifier circuit
	KeccakH wizard.VerifierCircuit `gnark:",secret"`
	// Invalidity type
	InvalidityType frontend.Variable `gnark:",secret"`
}

// Define represents the constraints relevant to [BadNonceBalanceCircuit]
func (circuit *BadNonceBalanceCircuit) Define(api frontend.API) error {

	var (
		account = circuit.AccountTrie.Account
		hKey    = circuit.AccountTrie.LeafOpening.HKey
	)

	//check that invalidity type is valid, it should be 0 or 1
	api.AssertIsBoolean(circuit.InvalidityType)

	// bad nonce; if invalidityType == 0 ----> tx nonce != account nonce + 1
	nonceDiff := api.Add(
		api.Mul(
			api.Sub(1, circuit.InvalidityType),
			api.Sub(circuit.TxNonce, api.Add(account.Nonce, 1)),
		),
		circuit.InvalidityType)
	api.AssertIsDifferent(nonceDiff, 0)

	// bad balance; if invalidityType == 1 ----> account balance < tx cost + 1
	api.AssertIsLessOrEqual(
		api.Mul(account.Balance, circuit.InvalidityType),
		api.Add(circuit.TxCost, 1))

	// check that the account matches the state root
	circuit.AccountTrie.Define(api)

	// check that sender address matches the account
	// Hash(FromAddress) == LeafOpening.HKey
	mimc, _ := gmimc.NewMiMC(api)
	mimc.Write(circuit.TxFromAddress)
	api.AssertIsEqual(hKey, mimc.Sum())

	// check that TxInfo matches the rlp encoding
	// expectedNonce := ExtractNonceFromRLPZk(api, circuit.RLPEncodedTx)
	// api.AssertIsEqual(expectedNonce, nonce)

	// check that rlpEncoding and TxHash are consistent with keccak input/output.
	checkKeccakConsistency(api, circuit.RLPEncodedTx, circuit.TxHash, &circuit.KeccakH)
	// verify keccak computation
	circuit.KeccakH.Verify(api)

	return nil
}

// Allocate the circuit
func (cir *BadNonceBalanceCircuit) Allocate(config Config) {
	// allocate the account trie
	cir.AccountTrie.Allocate(config)
	// allocate the keccak verifier
	cir.KeccakH = *wizard.AllocateWizardCircuit(config.KeccakCompiledIOP, 0)
	// allocate the RLPEncodedTx to have a fixed size
	cir.RLPEncodedTx = make([]frontend.Variable, config.MaxRlpByteSize)

}

// Assign the circuit from [AssigningInputs], circuit is reinitialized
func (cir *BadNonceBalanceCircuit) Assign(assi AssigningInputs) {

	var (
		txCost  = assi.Transaction.Cost()
		balance = assi.AccountTrieInputs.Account.Balance
		txNonce = assi.Transaction.Nonce()
		acNonce = assi.AccountTrieInputs.Account.Nonce
		a       = crypto.Keccak256(assi.RlpEncodedTx)
		// assign the keccak verifier
		keccak = *wizard.AssignVerifierCircuit(assi.KeccakCompiledIOP, assi.KeccakProof, 0)
	)

	cir.TxNonce = txNonce
	cir.TxCost = txCost
	cir.InvalidityType = int(assi.InvalidityType)
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

	// sanity-checks
	if assi.InvalidityType != 0 && assi.InvalidityType != 1 {
		utils.Panic("expected invalidity type 0 or 1 but received %v", assi.InvalidityType)
	}
	if txNonce == uint64(acNonce+1) && cir.InvalidityType == 0 {
		utils.Panic("tried to generate a bad-nonce proof for a possibly valid transaction")
	}

	if txCost.Cmp(balance) != 1 && cir.InvalidityType == 1 {
		utils.Panic("tried to generate a bad-balance proof for a possibly valid transaction")
	}

	// assign the account trie
	cir.AccountTrie.Assign(assi.AccountTrieInputs)
	cir.KeccakH = keccak
}

func (c *BadNonceBalanceCircuit) ExecutionCtx() []frontend.Variable {
	return []frontend.Variable{c.AccountTrie.MerkleProof.Root}
}
