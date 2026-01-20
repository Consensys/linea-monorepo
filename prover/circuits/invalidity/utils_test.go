package invalidity_test

import (
	"math/big"
	"testing"

	"github.com/consensys/gnark-crypto/ecc"
	"github.com/consensys/gnark/frontend"
	"github.com/consensys/gnark/frontend/cs/scs"
	"github.com/consensys/linea-monorepo/prover/backend/ethereum"
	"github.com/consensys/linea-monorepo/prover/circuits/invalidity"
	"github.com/consensys/linea-monorepo/prover/crypto/mimc"
	"github.com/consensys/linea-monorepo/prover/crypto/state-management/accumulator"
	"github.com/consensys/linea-monorepo/prover/crypto/state-management/hashtypes"
	"github.com/consensys/linea-monorepo/prover/crypto/state-management/smt"
	"github.com/consensys/linea-monorepo/prover/maths/common/vector"
	"github.com/consensys/linea-monorepo/prover/maths/field"
	. "github.com/consensys/linea-monorepo/prover/utils/types"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/go-playground/assert/v2"
	"github.com/stretchr/testify/require"
)

// TestNonceFromRLP tests the extraction of nonce from RLP-encoded EIP-1559 transactions.
//
// This test verifies both the native Go implementation (ExtractNonceFromRLP) and the
// ZK circuit implementation (ExtractNonceFromRLPZk) produce the correct nonce value.
//
// Test approach:
// 1. Create an EIP-1559 transaction from test case data
// 2. RLP-encode the transaction for signing
// 3. Extract nonce using the native Go function and verify it matches tx.Nonce()
// 4. Compile a ZK circuit that extracts the nonce and asserts equality
// 5. Verify the circuit is satisfied with the witness values
func TestNonceFromRLP(t *testing.T) {
	var (
		tx        = types.NewTx(&tcases[1].Tx)
		encodedTx = ethereum.EncodeTxForSigning(tx)
		a         byte
	)
	extractedNonce, _ := invalidity.ExtractNonceFromRLP(encodedTx)
	require.Equal(t, tx.Nonce(), extractedNonce, "extractedNonce and actualNonce are different")

	// Create witness with actual values
	witness := circuitExtractNonceFromRLP{
		RlpEncodedtx: make([]frontend.Variable, len(encodedTx)),
		Nonce:        tx.Nonce(),
	}

	for i := range encodedTx {
		a = encodedTx[i]
		witness.RlpEncodedtx[i] = a
	}

	// Create circuit for compilation
	circuit := circuitExtractNonceFromRLP{
		RlpEncodedtx: make([]frontend.Variable, len(encodedTx)),
	}

	ccs, err := frontend.Compile(ecc.BLS12_377.ScalarField(), scs.NewBuilder, &circuit, frontend.IgnoreUnconstrainedInputs())
	if err != nil {
		t.Fatal(err)
	}

	// solve the circuit
	twitness, err := frontend.NewWitness(&witness, ecc.BLS12_377.ScalarField())
	if err != nil {
		t.Fatal(err)
	}
	err = ccs.IsSolved(twitness)
	if err != nil {
		t.Fatal(err)
	}

}

// circuitExtractNonceFromRLP is a ZK circuit for testing nonce extraction from RLP-encoded transactions.
// RlpEncodedtx contains the raw transaction bytes as circuit variables.
// Nonce is the expected nonce value to assert against the extracted value.
type circuitExtractNonceFromRLP struct {
	RlpEncodedtx []frontend.Variable
	Nonce        frontend.Variable
}

func (c circuitExtractNonceFromRLP) Define(api frontend.API) error {
	gnarkExtractedNonce := invalidity.ExtractNonceFromRLPZk(api, c.RlpEncodedtx)
	api.AssertIsEqual(gnarkExtractedNonce, c.Nonce)
	return nil
}

// TestTxCostFromRLP tests the extraction of transaction cost from RLP-encoded EIP-1559 transactions.
//
// Transaction cost is defined as: cost = value + gasLimit × maxFeePerGas
// This is used for the "invalid balance" check in the invalidity circuit:
// if cost > sender.Balance, the transaction is invalid.
//
// This test verifies both the native Go implementation (ExtractTxCostFromRLP) and the
// ZK circuit implementation (ExtractTxCostFromRLPZk) correctly extract and compute the cost.
//
// EIP-1559 transaction fields used:
// - maxFeePerGas (field index 3): Maximum fee per gas unit the sender is willing to pay
// - gasLimit (field index 4): Maximum gas units the transaction can consume
// - value (field index 6): Amount of ETH to transfer
//
// Test approach:
// 1. Create an EIP-1559 transaction from test case data
// 2. RLP-encode the transaction for signing
// 3. Calculate expected cost from transaction fields
// 4. Extract cost using the native Go function and verify it matches
// 5. Compile a ZK circuit that extracts the cost and asserts equality
// 6. Verify the circuit is satisfied with the witness values
func TestTxCostFromRLP(t *testing.T) {
	var (
		tx        = types.NewTx(&tcases[1].Tx)
		encodedTx = ethereum.EncodeTxForSigning(tx)
		a         byte
	)

	// Calculate expected cost: value + gasLimit * maxFeePerGas
	expectedCost := tx.Value().Uint64() + tx.Gas()*tx.GasFeeCap().Uint64()

	extractedCost, err := invalidity.ExtractTxCostFromRLP(encodedTx)
	require.NoError(t, err)
	require.Equal(t, expectedCost, extractedCost, "extractedCost and expectedCost are different")

	// Create witness with actual values
	witness := circuitExtractTxCostFromRLP{
		RlpEncodedtx: make([]frontend.Variable, len(encodedTx)),
		Cost:         expectedCost,
	}

	for i := range encodedTx {
		a = encodedTx[i]
		witness.RlpEncodedtx[i] = a
	}

	// Create circuit for compilation
	circuit := circuitExtractTxCostFromRLP{
		RlpEncodedtx: make([]frontend.Variable, len(encodedTx)),
	}

	ccs, err := frontend.Compile(ecc.BLS12_377.ScalarField(), scs.NewBuilder, &circuit, frontend.IgnoreUnconstrainedInputs())
	if err != nil {
		t.Fatal(err)
	}

	// solve the circuit
	twitness, err := frontend.NewWitness(&witness, ecc.BLS12_377.ScalarField())
	if err != nil {
		t.Fatal(err)
	}
	err = ccs.IsSolved(twitness)
	if err != nil {
		t.Fatal(err)
	}
}

// circuitExtractTxCostFromRLP is a ZK circuit for testing transaction cost extraction from RLP-encoded transactions.
// RlpEncodedtx contains the raw transaction bytes as circuit variables.
// Cost is the expected transaction cost (value + gasLimit × maxFeePerGas) to assert against the extracted value.
type circuitExtractTxCostFromRLP struct {
	RlpEncodedtx []frontend.Variable
	Cost         frontend.Variable
}

func (c circuitExtractTxCostFromRLP) Define(api frontend.API) error {
	gnarkExtractedCost := invalidity.ExtractTxCostFromRLPZk(api, c.RlpEncodedtx)
	api.AssertIsEqual(gnarkExtractedCost, c.Cost)
	return nil
}

// generate a tree for testing
func getMerkleProof(t *testing.T) (smt.Proof, Bytes32, Bytes32) {

	config := &smt.Config{
		HashFunc: hashtypes.MiMC,
		Depth:    10,
	}

	// Generate random field elements and cast them into Bytes32es
	leavesFr := vector.Rand(1 << config.Depth)
	leaves := make([]Bytes32, len(leavesFr))
	for i := range leaves {
		leaves[i] = Bytes32(leavesFr[i].Bytes())
	}

	// And generate the tree
	tree := smt.BuildComplete(leaves, config.HashFunc)

	// Make a valid Bytes32
	leafs, _ := tree.GetLeaf(0)
	proofs, _ := tree.Prove(0)

	// Directly verify the proof
	valid := proofs.Verify(config, leafs, tree.Root)
	require.Truef(t, valid, "pos #%v, proof #%v", 0, proofs)

	return proofs, leafs, tree.Root
}

// test [badnonce.MerkleProofCircuit]
func TestMerkleProofs(t *testing.T) {

	// generate witness
	proofs, leafs, root := getMerkleProof(t)

	var witness invalidity.MerkleProofCircuit

	witness.Proofs.Siblings = make([]frontend.Variable, len(proofs.Siblings))
	for j := 0; j < len(proofs.Siblings); j++ {
		witness.Proofs.Siblings[j] = proofs.Siblings[j][:]
	}
	witness.Proofs.Path = proofs.Path
	witness.Leaf = leafs[:]

	witness.Root = root[:]

	// compile circuit
	var circuit invalidity.MerkleProofCircuit
	circuit.Proofs.Siblings = make([]frontend.Variable, len(proofs.Siblings))

	ccs, err := frontend.Compile(ecc.BLS12_377.ScalarField(), scs.NewBuilder, &circuit, frontend.IgnoreUnconstrainedInputs())
	if err != nil {
		t.Fatal(err)
	}

	// solve the circuit
	twitness, err := frontend.NewWitness(&witness, ecc.BLS12_377.ScalarField())
	if err != nil {
		t.Fatal(err)
	}
	err = ccs.IsSolved(twitness)
	if err != nil {
		t.Fatal(err)
	}

}

// test [badnonce.MimcCircuit]
func TestMimcCircuit(t *testing.T) {

	scs, err := frontend.Compile(
		ecc.BLS12_377.ScalarField(),
		scs.NewBuilder,
		&invalidity.MimcCircuit{PreImage: make([]frontend.Variable, 4)},
	)
	if err != nil {
		t.Fatal(err)
	}
	require.NoError(t, err)

	assignment := invalidity.MimcCircuit{
		PreImage: []frontend.Variable{0, 1, 2, 3},
		Hash:     mimc.HashVec(vector.ForTest(0, 1, 2, 3)),
	}

	witness, err := frontend.NewWitness(&assignment, ecc.BLS12_377.ScalarField())
	require.NoError(t, err)

	err = scs.IsSolved(witness)
	require.NoError(t, err)

}

// it test the Mimc Hashing over [types.Account]
func TestMimcAccount(t *testing.T) {

	var (
		buf field.Element
		// generate Mimc witness for Hash(Account)
		a = tcases[1].Account

		witMimc invalidity.MimcCircuit

		account = GnarkAccount{
			Nonce:    a.Nonce,
			Balance:  a.Balance,
			CodeSize: a.CodeSize,
		}
		accountSlice = []frontend.Variable{}

		config = &smt.Config{
			HashFunc: hashtypes.MiMC,
			Depth:    10,
		}
	)

	account.StorageRoot = *buf.SetBytes(a.StorageRoot[:])
	account.MimcCodeHash = *buf.SetBytes(a.MimcCodeHash[:])
	account.KeccakCodeHashMSB = *buf.SetBytes(a.KeccakCodeHash[16:])
	account.KeccakCodeHashLSB = *buf.SetBytes(a.KeccakCodeHash[:16])

	witMimc.PreImage = append(accountSlice,
		account.Nonce,
		account.Balance,
		account.StorageRoot,
		account.MimcCodeHash,
		account.KeccakCodeHashMSB,
		account.KeccakCodeHashLSB,
		account.CodeSize,
	)
	hash := accumulator.Hash(config, a)
	witMimc.Hash = *buf.SetBytes(hash[:])

	//compile the circuit
	scs, err := frontend.Compile(
		ecc.BLS12_377.ScalarField(),
		scs.NewBuilder,
		&invalidity.MimcCircuit{PreImage: make([]frontend.Variable, len(witMimc.PreImage))},
	)
	if err != nil {
		t.Fatal(err)
	}
	require.NoError(t, err)

	witness, err := frontend.NewWitness(&witMimc, ecc.BLS12_377.ScalarField())
	require.NoError(t, err)

	err = scs.IsSolved(witness)
	require.NoError(t, err)

}

// it creates a merkle tree for the given [accumulator.LeafOpening] and config
func genShomei(t *testing.T, tcases []TestCases, config *smt.Config) (*smt.Tree, []smt.Proof, []Bytes32) {

	var leaves = []Bytes32{}
	for _, c := range tcases {

		leaf := accumulator.Hash(config, &accumulator.LeafOpening{
			Prev: c.Leaf.Prev,
			Next: c.Leaf.Next,
			HKey: c.Leaf.HKey,
			HVal: accumulator.Hash(config, c.Account),
		})

		leaves = append(leaves, leaf)
	}

	// Build the same tree by adding the leaves one by one
	tree := smt.NewEmptyTree(config)
	for i := range leaves {
		tree.Update(i, leaves[i])
	}

	var (
		leafs  = []Bytes32{}
		proofs = []smt.Proof{}
	)
	// Make a valid Bytes32
	for i := range leaves {

		leaf, _ := tree.GetLeaf(i)
		proof, _ := tree.Prove(i)

		// Directly verify the proof
		valid := proof.Verify(config, leaf, tree.Root)
		require.Truef(t, valid, "pos #%v, proof #%v", 0, proof)

		leafs = append(leafs, leaf)
		proofs = append(proofs, proof)
	}

	return tree, proofs, leaves
}

// it gets a leaf via its position and check it has the expected value.
func TestShomei(t *testing.T) {

	config := &smt.Config{
		HashFunc: hashtypes.MiMC,
		Depth:    10,
	}
	tree, _, leaves := genShomei(t, tcases, config)

	for i := range leaves {
		leaf, _ := tree.GetLeaf(i)
		c := tcases[i]

		expectedLeaf := accumulator.Hash(config, &accumulator.LeafOpening{
			Prev: c.Leaf.Prev,
			Next: c.Leaf.Next,
			HKey: c.Leaf.HKey,
			HVal: accumulator.Hash(config, c.Account),
		})

		assert.Equal(t, leaf, expectedLeaf)

	}

}

type TestCases struct {
	Account        Account
	Leaf           accumulator.LeafOpening
	Tx             types.DynamicFeeTx
	FromAddress    common.Address
	TxHash         common.Hash
	InvalidityType invalidity.InvalidityType
}

var tcases = []TestCases{

	{
		Account: Account{
			Balance: big.NewInt(0),
		},
		Leaf: accumulator.LeafOpening{
			Prev: 0,
			Next: 1,
			HKey: hKeyFromAddress(common.HexToAddress("0x00aed6")),
			HVal: hValFromAccount(Account{
				Balance: big.NewInt(0),
			}),
		},
		Tx: types.DynamicFeeTx{
			ChainID:   big.NewInt(59144), // Linea mainnet chain ID
			Nonce:     1,                 // valid nonce
			Value:     big.NewInt(1),     //invalid balance
			Gas:       1,
			GasFeeCap: big.NewInt(1), // gas price
		},
		FromAddress:    common.HexToAddress("0x00aed6"),
		TxHash:         common.HexToHash("0x3f1d2e2b4c3f4e5d6c7b8a9b0c1d2e3f4a5b6c7d8e9f0a1b2c3d4e5f6a7b8c9d"),
		InvalidityType: 1,
	},
	{
		// EOA
		Account: Account{
			Nonce:          65,
			Balance:        big.NewInt(5690),
			StorageRoot:    Bytes32FromHex("0x00aed60bedfcad80c2a5e6a7a3100e837f875f9aa71d768291f68f894b0a3d11"),
			MimcCodeHash:   Bytes32FromHex("0x007298fd87d3039ffea208538f6b297b60b373a63792b4cd0654fdc88fd0d6ee"),
			KeccakCodeHash: FullBytes32FromHex("0xc5d2460186f7233c927e7db2dcc703c0e500b653ca82273b7bfad8045d85a470"),
			CodeSize:       0,
		},

		Leaf: accumulator.LeafOpening{
			Prev: 0,
			Next: 2,
			HKey: hKeyFromAddress(common.HexToAddress("0x00aed7")),
			HVal: hValFromAccount(Account{
				Nonce:          65,
				Balance:        big.NewInt(5690),
				StorageRoot:    Bytes32FromHex("0x00aed60bedfcad80c2a5e6a7a3100e837f875f9aa71d768291f68f894b0a3d11"),
				MimcCodeHash:   Bytes32FromHex("0x007298fd87d3039ffea208538f6b297b60b373a63792b4cd0654fdc88fd0d6ee"),
				KeccakCodeHash: FullBytes32FromHex("0xc5d2460186f7233c927e7db2dcc703c0e500b653ca82273b7bfad8045d85a470"),
				CodeSize:       0,
			}),
		},
		Tx: types.DynamicFeeTx{
			ChainID:   big.NewInt(59144), // Linea mainnet chain ID
			Nonce:     65,                // invalid nonce
			Value:     big.NewInt(5700),  // invalid value
			Gas:       1,
			GasFeeCap: big.NewInt(1), // gas price
		},
		FromAddress:    common.HexToAddress("0x00aed7"),
		TxHash:         common.HexToHash("0x4f1d2e2b4c3f4e5d6c7b8a9b0c1d2e3f4a5b6c7d8e9f0a1b2c3d4e5f6a7b8c9e"),
		InvalidityType: 0, //  0 = BadNonce, 1 = BadBalance
	},
	{
		// Another EOA
		Account: Account{
			Nonce:          65,
			Balance:        big.NewInt(835),
			StorageRoot:    Bytes32FromHex("0x007942bb21022172cbad3ffc38d1c59e998f1ab6ab52feb15345d04bbf859f14"),
			MimcCodeHash:   Bytes32FromHex("0x007298fd87d3039ffea208538f6b297b60b373a63792b4cd0654fdc88fd0d6ee"),
			KeccakCodeHash: FullBytes32FromHex("0xc5d2460186f7233c927e7db2dcc703c0e500b653ca82273b7bfad8045d85a470"),
			CodeSize:       0,
		},
		Leaf: accumulator.LeafOpening{
			Prev: 1,
			Next: 3,
			HKey: hKeyFromAddress(common.HexToAddress("0x00aed8")),
			HVal: hValFromAccount(Account{
				Nonce:          65,
				Balance:        big.NewInt(835),
				StorageRoot:    Bytes32FromHex("0x007942bb21022172cbad3ffc38d1c59e998f1ab6ab52feb15345d04bbf859f14"),
				MimcCodeHash:   Bytes32FromHex("0x007298fd87d3039ffea208538f6b297b60b373a63792b4cd0654fdc88fd0d6ee"),
				KeccakCodeHash: FullBytes32FromHex("0xc5d2460186f7233c927e7db2dcc703c0e500b653ca82273b7bfad8045d85a470"),
				CodeSize:       0,
			}),
		},
		Tx: types.DynamicFeeTx{
			ChainID:   big.NewInt(59144), // Linea mainnet chain ID
			Nonce:     63,                // invalid nonce
			Value:     big.NewInt(800),   // valid value
			Gas:       1,
			GasFeeCap: big.NewInt(1), // gas price
		},
		FromAddress:    common.HexToAddress("0x00aed8"),
		TxHash:         common.HexToHash("0x5f1d2e2b4c3f4e5d6c7b8a9b0c1d2e3f4a5b6c7d8e9f0a1b2c3d4e5f6a7b8c9f"),
		InvalidityType: 0,
	},
}

func hKeyFromAddress(add common.Address) Bytes32 {
	mimc := mimc.NewMiMC()
	mimc.Write(add.Bytes())
	return Bytes32(mimc.Sum(nil))
}

func hValFromAccount(a Account) Bytes32 {
	mimc := mimc.NewMiMC()
	a.WriteTo(mimc)
	return Bytes32(mimc.Sum(nil))
}
