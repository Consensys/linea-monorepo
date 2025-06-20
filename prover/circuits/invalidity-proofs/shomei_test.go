package badnonce_test

import (
	"math/big"
	"testing"

	"github.com/consensys/linea-monorepo/prover/crypto/state-management/accumulator"
	"github.com/consensys/linea-monorepo/prover/crypto/state-management/hashtypes"
	"github.com/consensys/linea-monorepo/prover/crypto/state-management/smt"
	. "github.com/consensys/linea-monorepo/prover/utils/types"
	"github.com/go-playground/assert/v2"
	"github.com/stretchr/testify/require"
)

// it creates a merkle tree for the given [accumulator.LeafOpening] and config
func genShomei(t *testing.T, tcases []TestCases, config *smt.Config) (*smt.Tree, []smt.Proof, []Bytes32) {

	var leaves []Bytes32
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
		leafs  []Bytes32
		proofs []smt.Proof
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
	},
}
