package badnonce_test

import (
	"math/big"
	"testing"

	"github.com/consensys/linea-monorepo/prover/crypto/state-management/accumulator"
	"github.com/consensys/linea-monorepo/prover/crypto/state-management/smt"
	. "github.com/consensys/linea-monorepo/prover/utils/types"
	"github.com/stretchr/testify/require"
)

// it create a merkle tree for the given accounts and config
func genShomei(t *testing.T, tcases []TestCases, config *smt.Config) (*smt.Tree, []smt.Proof, []Bytes32) {

	/*	config := &smt.Config{
		HashFunc: hashtypes.MiMC,
		Depth:    depth,
	}*/

	var leaves []Bytes32
	for _, c := range tcases {
		account := AccountForHash{
			Acc: c.Account,
		}

		leaves = append(leaves, accumulator.Hash(config, account))
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

type TestCases struct {
	Account   Account
	HexString string
}

var tcases = []TestCases{

	{
		Account: Account{
			Balance: big.NewInt(0),
		},
		HexString: "0x000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000",
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
		HexString: "0x0000000000000000000000000000000000000000000000000000000000000041000000000000000000000000000000000000000000000000000000000000163a00aed60bedfcad80c2a5e6a7a3100e837f875f9aa71d768291f68f894b0a3d11007298fd87d3039ffea208538f6b297b60b373a63792b4cd0654fdc88fd0d6eec5d2460186f7233c927e7db2dcc703c0e500b653ca82273b7bfad8045d85a4700000000000000000000000000000000000000000000000000000000000000000",
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
		HexString: "0x00000000000000000000000000000000000000000000000000000000000000410000000000000000000000000000000000000000000000000000000000000343007942bb21022172cbad3ffc38d1c59e998f1ab6ab52feb15345d04bbf859f14007298fd87d3039ffea208538f6b297b60b373a63792b4cd0654fdc88fd0d6eec5d2460186f7233c927e7db2dcc703c0e500b653ca82273b7bfad8045d85a4700000000000000000000000000000000000000000000000000000000000000000",
	},
}
