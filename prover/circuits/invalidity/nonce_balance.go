package invalidity

import (
	"github.com/consensys/gnark/frontend"
	gmimc "github.com/consensys/gnark/std/hash/mimc"
	"github.com/consensys/linea-monorepo/prover/circuits/internal"
	"github.com/consensys/linea-monorepo/prover/crypto/fiatshamir"
	"github.com/consensys/linea-monorepo/prover/crypto/mimc/gkrmimc"
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
	TxCost frontend.Variable
	// RLP-encoded payload  prefixed with the type byte. txType || rlp(tx.inner)
	RLPEncodedTx []frontend.Variable // the RLP encoded of the unsigned transaction
	// sender address
	TxFromAddress frontend.Variable
	// AccountTrie of the sender
	AccountTrie AccountTrie
	// hash of the transaction
	TxHash [2]frontend.Variable
	// Keccak verifier circuit
	KeccakH wizard.VerifierCircuit
	// Invalidity type
	InvalidityType frontend.Variable
}

// Define represents the constraints relevant to [BadNonceBalanceCircuit]
func (circuit *BadNonceBalanceCircuit) Define(api frontend.API) error {

	var (
		account = circuit.AccountTrie.Account
		hKey    = circuit.AccountTrie.LeafOpening.HKey
	)

	// setup the hashing approach for the keccak verifier
	circuit.KeccakH.HasherFactory = gkrmimc.NewHasherFactory(api)
	circuit.KeccakH.FS = fiatshamir.NewGnarkFiatShamir(api, circuit.KeccakH.HasherFactory)

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

	// check that TxNonce matches the rlp encoding
	expectedNonce := ExtractNonceFromRLPZk(api, circuit.RLPEncodedTx)
	api.AssertIsEqual(expectedNonce, circuit.TxNonce)

	// check that TxCost matches the rlp encoding
	expectedCost := ExtractTxCostFromRLPZk(api, circuit.RLPEncodedTx)
	api.AssertIsEqual(expectedCost, circuit.TxCost)

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
	// TODO: re-enable this check once consistency check between nonce and rlp is implemented
	// if assi.InvalidityType == 0 || assi.InvalidityType == 1 {
	// 	utils.Panic("yet unsupported, since consistency check between nonce and rlp is missing")
	// }
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
	cir.TxHash[0] = a[0:LIMB_SIZE]
	cir.TxHash[1] = a[LIMB_SIZE:]

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
	if assi.InvalidityType != BadNonce && assi.InvalidityType != BadBalance {
		utils.Panic("expected invalidity type BadNonce or BadBalance but received %v", assi.InvalidityType)
	}
	if txNonce == uint64(acNonce+1) && assi.InvalidityType == BadNonce {
		utils.Panic("tried to generate a bad-nonce proof for a possibly valid transaction")
	}

	if txCost.Cmp(balance) != 1 && assi.InvalidityType == BadBalance {
		utils.Panic("tried to generate a bad-balance proof for a possibly valid transaction")
	}

	// assign the account trie
	cir.AccountTrie.Assign(assi.AccountTrieInputs)
	cir.KeccakH = keccak
}

// FunctionalPublicInputs returns the functional public inputs of the circuit
func (c *BadNonceBalanceCircuit) FunctionalPublicInputs() FunctionalPublicInputsGnark {
	return FunctionalPublicInputsGnark{
		TxHash:        c.TxHash,
		FromAddress:   c.TxFromAddress,
		StateRootHash: c.AccountTrie.MerkleProof.Root,
	}
}
