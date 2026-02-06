package invalidity

import (
	"math/big"

	"github.com/consensys/gnark/frontend"
	"github.com/consensys/linea-monorepo/prover/circuits/internal"
	"github.com/consensys/linea-monorepo/prover/circuits/pi-interconnection/keccak/prover/protocol/wizard"
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
	KeccakH wizard.VerifierCircuit
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
	accountNonce := reconstructFromLimbs(api, toNativeSlice(account.Nonce[:]))

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
	accountBalance := reconstructFromLimbs(api, toNativeSlice(account.Balance[:]))

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

// reconstructFromLimbs reconstructs a value from 16-bit limbs (big-endian)
func reconstructFromLimbs(api frontend.API, limbs []frontend.Variable) frontend.Variable {
	result := frontend.Variable(0)
	n := len(limbs)
	for i := 0; i < n; i++ {
		shift := big.NewInt(1)
		shift.Lsh(shift, uint((n-1-i)*16))
		result = api.Add(result, api.Mul(limbs[i], shift))
	}
	return result
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
	cir.KeccakH = *wizard.AllocateWizardCircuit(config.KeccakCompiledIOP, 0)
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
		keccak  = wizard.AssignVerifierCircuit(assi.KeccakCompiledIOP, assi.KeccakProof, 0)
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
	root := c.AccountTrie.MerkleProof.Root
	api := c.api

	// Convert Root octuplet to 2 BLS12-377 field elements
	// This must match the Assign function in pi.go which uses ToBytes()
	// Each KoalaBear element is 31 bits, stored as 4 bytes big-endian
	// Combining 4 elements (16 bytes) into one BLS field element:
	// result = e[0] * 2^96 + e[1] * 2^64 + e[2] * 2^32 + e[3]

	shift96 := new(big.Int).Lsh(big.NewInt(1), 96)
	shift64 := new(big.Int).Lsh(big.NewInt(1), 64)
	shift32 := new(big.Int).Lsh(big.NewInt(1), 32)

	firstHalf := api.Add(
		api.Mul(root[0].Native(), shift96),
		api.Mul(root[1].Native(), shift64),
		api.Mul(root[2].Native(), shift32),
		root[3].Native(),
	)

	secondHalf := api.Add(
		api.Mul(root[4].Native(), shift96),
		api.Mul(root[5].Native(), shift64),
		api.Mul(root[6].Native(), shift32),
		root[7].Native(),
	)

	return FunctionalPublicInputsGnark{
		TxHash:        c.TxHash,
		FromAddress:   c.TxFromAddress,
		StateRootHash: [2]frontend.Variable{firstHalf, secondHalf},
	}
}
