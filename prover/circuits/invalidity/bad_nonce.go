package invalidity

import (
	"math/big"

	"github.com/consensys/gnark/frontend"
	gmimc "github.com/consensys/gnark/std/hash/mimc"
	"github.com/consensys/linea-monorepo/prover/circuits/blobdecompression/v0/compress"
	"github.com/consensys/linea-monorepo/prover/circuits/internal"
	"github.com/consensys/linea-monorepo/prover/protocol/ifaces"
	"github.com/consensys/linea-monorepo/prover/protocol/wizard"
	"github.com/consensys/linea-monorepo/prover/utils"
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
func (c *BadNonceCircuit) Allocate(config Config, comp *wizard.CompiledIOP) {
	c.AccountTrie.Allocate(config)
	keccak := wizard.AllocateWizardCircuit(comp, 0)
	c.KeccakH = *keccak
}

// Assign the circuit from [AssigningInputs]
func (c *BadNonceCircuit) Assign(assi AssigningInputs, comp *wizard.CompiledIOP, proof wizard.Proof) {

	var (
		txNonce = assi.Transaction.Nonce()
		acNonce = assi.AccountTrieInputs.Account.Nonce
		a       = assi.Transaction.Hash()
	)
	*c = BadNonceCircuit{
		TxNonce:       txNonce,
		TxFromAddress: assi.FromAddress[:],
	}

	c.TxHash[0] = a[:16]
	c.TxHash[1] = a[16:32]

	if assi.RlpEncodedTx[0] != 0x02 {
		utils.Panic("only support typed 2 transactions, maybe the rlp is not prefixed with the type byte")
	}
	// assign the rlp encoding
	c.RLPEncodedTx = internal.FromBytesToElements(assi.RlpEncodedTx)
	// sanity-check
	if txNonce == uint64(acNonce+1) {
		utils.Panic("tried to generate a bad-nonce proof for a valid transaction")
	}

	// assign the account trie
	c.AccountTrie.Assign(assi.AccountTrieInputs)
	// assign the keccak verifier
	c.KeccakH = *wizard.AssignVerifierCircuit(comp, proof, 0)
}

func (c *BadNonceCircuit) ExecutionCtx() []frontend.Variable {
	return []frontend.Variable{c.AccountTrie.MerkleProof.Root}
}

func checkKeccakConsistency(api frontend.API, rlpEncode []frontend.Variable, txHash [2]frontend.Variable, keccak *wizard.VerifierCircuit) {

	var (
		radix       = big.NewInt(256)
		ctr         = 0
		limbCol     = keccak.GetColumn(ifaces.ColIDf("TxHash_INVALIDITY_LIMBS"))
		hashHiCol   = keccak.GetColumn(ifaces.ColIDf("TxHash_INVALIDITY_HASH_HI"))
		hashLoCol   = keccak.GetColumn(ifaces.ColIDf("TxHash_INVALIDITY_HASH_LO"))
		isHashHiCol = keccak.GetColumn(ifaces.ColIDf("TxHash_INVALIDITY_IS_HASH_HI"))
		isHashLoCol = keccak.GetColumn(ifaces.ColIDf("TxHash_INVALIDITY_IS_HASH_LO"))
	)

	// check that the rlpEncoding matches the limb column
	if len(limbCol) < len(rlpEncode)/16+1 {
		utils.Panic("keccak limb column is not large enough to hold the rlp encoding")
	}

	// split the rlpEncoding into chunks of 16 bytes
	for len(rlpEncode) > 16 {
		v := rlpEncode[:16]
		curLimb := compress.ReadNum(api, v, radix)
		api.AssertIsEqual(limbCol[ctr], curLimb)
		ctr++
		rlpEncode = rlpEncode[16:]
	}
	// handle the last chunk
	if len(rlpEncode) > 0 {
		// left align and pad with zeros
		v := make([]frontend.Variable, 16)
		copy(v, rlpEncode)
		curLimb := compress.ReadNum(api, v, radix)
		api.AssertIsEqual(limbCol[ctr], curLimb)
	}

	// check that the hash output matches the txHash
	api.AssertIsEqual(hashHiCol[0], txHash[0])
	api.AssertIsEqual(hashLoCol[0], txHash[1])

	// check that isHashHi and isHashLo are set to 1
	api.AssertIsEqual(isHashHiCol[0], 1)
	api.AssertIsEqual(isHashLoCol[0], 1)

	// check that the rest of the limb column is padded with zeros
	// note that due to the collision resistance of keccak,
	// this check along side the above checks are enough,
	// and we dont need to check (nByte, hashNum, toHash, index) columns.
	for i := ctr + 1; i < len(limbCol); i++ {
		api.AssertIsEqual(limbCol[i], 0)
	}

}
