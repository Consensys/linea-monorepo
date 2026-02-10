package invalidity

import (
	"math/big"

	"github.com/consensys/gnark/frontend"
	"github.com/consensys/gnark/std/compress"
	"github.com/consensys/linea-monorepo/prover/circuits/internal"
	wizardk "github.com/consensys/linea-monorepo/prover/circuits/pi-interconnection/keccak/prover/protocol/wizard"
	"github.com/consensys/linea-monorepo/prover/crypto/poseidon2_koalabear"
	"github.com/consensys/linea-monorepo/prover/maths/field/koalagnark"
	"github.com/consensys/linea-monorepo/prover/utils"
	"github.com/ethereum/go-ethereum/crypto"
)

// BadNonceBalanceCircuit defines the circuit for the transaction with a bad nonce or insufficient balance.
// The circuit does not do any range check over the transaction field,
// as it is supposed to be done by the contract.
//
// AccountTrie uses KoalaBear Poseidon2 for state membership verification.
// Other fields use BLS12-377 native arithmetic.
type BadNonceBalanceCircuit struct {
	// Transaction nonce (single BLS variable)
	TxNonce frontend.Variable
	// Transaction cost (single BLS variable)
	TxCost frontend.Variable
	// RLP-encoded payload prefixed with the type byte: txType || rlp(tx.inner)
	RLPEncodedTx []frontend.Variable
	// Sender address
	TxFromAddress frontend.Variable
	// AccountTrie of the sender (uses KoalaBear Poseidon2)
	AccountTrie AccountTrie
	// Hash of the transaction (split into two 16-byte chunks)
	TxHash [2]frontend.Variable
	// Keccak verifier circuit
	KeccakH wizardk.VerifierCircuit
	// Invalidity type: 0 = BadNonce, 1 = BadBalance
	InvalidityType frontend.Variable
	api            frontend.API
}

// Define represents the constraints relevant to [BadNonceBalanceCircuit]
func (circuit *BadNonceBalanceCircuit) Define(api frontend.API) error {
	// Store API for use in FunctionalPublicInputs
	circuit.api = api

	var (
		account = circuit.AccountTrie.Account
		hKey    = circuit.AccountTrie.LeafOpening.HKey
	)

	// Check that invalidity type is valid: 0 = BadNonce, 1 = BadBalance
	api.AssertIsBoolean(circuit.InvalidityType)

	// ========== NONCE CHECK ==========
	// Reconstruct account nonce from limbs (4 x 16-bit, big-endian)
	accountNonce := combine16BitLimbs(api, toNativeSlice(account.Nonce[:]))

	// Bad nonce: if invalidityType == 0 ----> tx nonce != account nonce + 1
	nonceDiff := api.Add(
		api.Mul(
			api.Sub(1, circuit.InvalidityType),
			api.Sub(circuit.TxNonce, api.Add(accountNonce, 1)),
		),
		circuit.InvalidityType)
	api.AssertIsDifferent(nonceDiff, 0)

	// ========== BALANCE CHECK ==========
	// Reconstruct account balance from limbs (16 x 16-bit, big-endian)
	accountBalance := combine16BitLimbs(api, toNativeSlice(account.Balance[:]))

	// Bad balance: if invalidityType == 1 ----> account balance < tx cost+1
	api.AssertIsLessOrEqual(
		api.Mul(accountBalance, circuit.InvalidityType),
		api.Add(circuit.TxCost, 1))

	// ========== ACCOUNT TRIE MEMBERSHIP ==========
	// Verify account is in the state trie using Poseidon2 Merkle proof
	circuit.AccountTrie.Define(api)

	// ========== ADDRESS VERIFICATION ==========
	// Check that sender address matches the account's HKey
	// HKey = Poseidon2(FromAddress)
	koalaHasher := poseidon2_koalabear.NewKoalagnarkMDHasher(api)
	addressElements := addressToKoalaElements(api, circuit.TxFromAddress)
	koalaHasher.Write(addressElements...)
	addressHash := koalaHasher.Sum()
	koalaAPI := koalagnark.NewAPI(api)
	for i := 0; i < 8; i++ {
		koalaAPI.AssertIsEqual(addressHash[i], hKey[i])
	}

	// ========== RLP CONSISTENCY ==========
	// Check that TxNonce matches the RLP encoding
	expectedNonce := ExtractNonceFromRLPZk(api, circuit.RLPEncodedTx)
	api.AssertIsEqual(expectedNonce, circuit.TxNonce)

	// Check that TxCost matches the RLP encoding
	expectedCost := ExtractTxCostFromRLPZk(api, circuit.RLPEncodedTx)
	api.AssertIsEqual(expectedCost, circuit.TxCost)

	// ========== KECCAK VERIFICATION (TODO) ==========
	CheckKeccakConsistency(api, circuit.RLPEncodedTx, circuit.TxHash, &circuit.KeccakH)
	circuit.KeccakH.Verify(api)

	return nil
}

func toNativeSlice(values []koalagnark.Element) []frontend.Variable {
	res := make([]frontend.Variable, len(values))
	for i := range values {
		res[i] = values[i].Native()
	}
	return res
}

func addressToKoalaElements(api frontend.API, addr frontend.Variable) []koalagnark.Element {
	// Address is 20 bytes, split into 10 x 16-bit limbs (big-endian order).
	// This avoids overflowing KoalaBear's 31-bit field element.
	const (
		addrBits  = 160
		chunkBits = 16
		chunks    = addrBits / chunkBits
	)

	bits := api.ToBinary(addr, addrBits)
	res := make([]koalagnark.Element, chunks)
	for i := 0; i < chunks; i++ {
		start := (chunks - 1 - i) * chunkBits
		end := start + chunkBits
		limb := api.FromBinary(bits[start:end]...)
		res[i] = koalagnark.WrapFrontendVariable(limb)
	}
	return res
}

// Allocate the circuit
func (cir *BadNonceBalanceCircuit) Allocate(config Config) {
	// Allocate the account trie
	cir.AccountTrie.Allocate(config)
	// Allocate the keccak verifier
	cir.KeccakH = *wizardk.AllocateWizardCircuit(config.KeccakCompiledIOP, 0)
	// Allocate the RLPEncodedTx to have a fixed size
	cir.RLPEncodedTx = make([]frontend.Variable, config.MaxRlpByteSize)
}

// Assign the circuit from [AssigningInputs], circuit is reinitialized
func (cir *BadNonceBalanceCircuit) Assign(assi AssigningInputs) {
	var (
		txCost  = assi.Transaction.Cost()
		balance = assi.AccountTrieInputs.Account.Balance
		txNonce = assi.Transaction.Nonce()
		acNonce = assi.AccountTrieInputs.Account.Nonce
		txHash  = crypto.Keccak256(assi.RlpEncodedTx)
		keccak  = wizardk.AssignVerifierCircuit(assi.KeccakCompiledIOP, assi.KeccakProof, 0)
	)

	cir.TxNonce = txNonce
	cir.TxCost = txCost
	cir.InvalidityType = int(assi.InvalidityType)
	cir.TxFromAddress = assi.FromAddress[:]
	cir.RLPEncodedTx = make([]frontend.Variable, assi.MaxRlpByteSize)

	// Assign the tx hash
	cir.TxHash[0] = txHash[0:LIMB_SIZE]
	cir.TxHash[1] = txHash[LIMB_SIZE:]

	if assi.RlpEncodedTx[0] != 0x02 {
		utils.Panic("only support typed 2 transactions, maybe the rlp is not prefixed with the type byte")
	}

	// Assign the RLP encoding
	elements := internal.FromBytesToElements(assi.RlpEncodedTx)
	rlpLen := len(elements)
	if rlpLen > assi.MaxRlpByteSize {
		utils.Panic("rlp encoding is too large: got %d, max %d", rlpLen, assi.MaxRlpByteSize)
	}

	copy(cir.RLPEncodedTx, elements)
	for i := len(elements); i < assi.MaxRlpByteSize; i++ {
		cir.RLPEncodedTx[i] = 0
	}

	// Sanity checks
	if assi.InvalidityType != BadNonce && assi.InvalidityType != BadBalance {
		utils.Panic("expected invalidity type BadNonce or BadBalance but received %v", assi.InvalidityType)
	}
	if txNonce == uint64(acNonce+1) && assi.InvalidityType == BadNonce {
		utils.Panic("tried to generate a bad-nonce proof for a possibly valid transaction")
	}
	if txCost.Cmp(balance) != 1 && assi.InvalidityType == BadBalance {
		utils.Panic("tried to generate a bad-balance proof for a possibly valid transaction")
	}

	// Assign the account trie
	cir.AccountTrie.Assign(assi.AccountTrieInputs)
	cir.KeccakH = *keccak
}

// FunctionalPublicInputs returns the functional public inputs of the circuit
func (c *BadNonceBalanceCircuit) FunctionalPublicInputs() FunctionalPublicInputsGnark {

	return FunctionalPublicInputsGnark{
		TxHash:        c.TxHash,
		FromAddress:   c.TxFromAddress,
		StateRootHash: reconstructRootHash(c.api, c.AccountTrie.MerkleProof.Root),
	}
}

// reconstructRootHash converts a Root octuplet to 2 BLS12-377 field elements
// Combining 4 elements (16 bytes) into one BLS field element using base 2^32
func reconstructRootHash(api frontend.API, root koalagnark.Octuplet) [2]frontend.Variable {

	// Convert koalagnark.Element to frontend.Variable
	rootVars := make([]frontend.Variable, 8)
	for i := range root {
		rootVars[i] = root[i].Native()
	}

	return [2]frontend.Variable{
		combine32BitLimbs(api, rootVars[:4]), // First 4 KoalaBear elements
		combine32BitLimbs(api, rootVars[4:]), // Last 4 KoalaBear elements
	}
}

// Helper function to combine KoalaBear elements using compress.ReadNum
func combine32BitLimbs(api frontend.API, vs []frontend.Variable) frontend.Variable {
	p32 := new(big.Int).Lsh(big.NewInt(1), 32) // base = 2^32
	return compress.ReadNum(api, vs, p32)
}

func combine16BitLimbs(api frontend.API, vs []frontend.Variable) frontend.Variable {
	p16 := new(big.Int).Lsh(big.NewInt(1), 16) // base = 2^16
	return compress.ReadNum(api, vs, p16)
}
