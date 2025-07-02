package badnonce_test

import (
	"math/big"
	"testing"

	"github.com/consensys/gnark-crypto/ecc"
	"github.com/consensys/gnark/frontend"
	"github.com/consensys/gnark/frontend/cs/scs"
	badnonce "github.com/consensys/linea-monorepo/prover/circuits/invalidity-proofs"
	"github.com/consensys/linea-monorepo/prover/crypto/mimc"
	"github.com/consensys/linea-monorepo/prover/crypto/state-management/accumulator"
	"github.com/consensys/linea-monorepo/prover/crypto/state-management/hashtypes"
	"github.com/consensys/linea-monorepo/prover/crypto/state-management/smt"
	"github.com/consensys/linea-monorepo/prover/maths/common/vector"
	"github.com/consensys/linea-monorepo/prover/maths/field"
	"github.com/consensys/linea-monorepo/prover/utils/types"
	. "github.com/consensys/linea-monorepo/prover/utils/types"
	"github.com/go-playground/assert/v2"
	"github.com/stretchr/testify/require"
)

// generate  a tree for testing
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

	var witness badnonce.MerkleProofCircuit

	witness.Proofs.Siblings = make([]frontend.Variable, len(proofs.Siblings))
	for j := 0; j < len(proofs.Siblings); j++ {
		witness.Proofs.Siblings[j] = proofs.Siblings[j][:]
	}
	witness.Proofs.Path = proofs.Path
	witness.Leaf = leafs[:]

	witness.Root = root[:]

	// compile circuit
	var circuit badnonce.MerkleProofCircuit
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
		&badnonce.MimcCircuit{PreImage: make([]frontend.Variable, 4)},
	)
	if err != nil {
		t.Fatal(err)
	}
	require.NoError(t, err)

	assignment := badnonce.MimcCircuit{
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

		witMimc badnonce.MimcCircuit

		account = types.GnarkAccount{
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
		&badnonce.MimcCircuit{PreImage: make([]frontend.Variable, len(witMimc.PreImage))},
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
	Account Account
	Leaf    accumulator.LeafOpening
	TxNonce int
}

var tcases = []TestCases{

	{
		Account: Account{
			Balance: big.NewInt(0),
		},
		Leaf: accumulator.LeafOpening{
			Prev: 0,
			Next: 1,
			HKey: Bytes32FromHex("0x00aed6"),
		},
		TxNonce: 1, // valid nonce
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
			HKey: Bytes32FromHex("0x00aed7"),
		},
		TxNonce: 65,
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
			HKey: Bytes32FromHex("0x00aed8"),
		},
		TxNonce: 67,
	},
}
