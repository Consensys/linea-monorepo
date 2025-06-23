package badnonce_test

import (
	"testing"

	"github.com/consensys/gnark-crypto/ecc"
	"github.com/consensys/gnark-crypto/ecc/bls12-377/fr"
	"github.com/consensys/gnark/frontend"
	"github.com/consensys/gnark/frontend/cs/scs"
	badnonce "github.com/consensys/linea-monorepo/prover/circuits/invalidity-proofs"
	"github.com/consensys/linea-monorepo/prover/crypto/mimc"
	"github.com/consensys/linea-monorepo/prover/crypto/state-management/hashtypes"
	"github.com/consensys/linea-monorepo/prover/crypto/state-management/smt"
	"github.com/consensys/linea-monorepo/prover/maths/common/vector"
	. "github.com/consensys/linea-monorepo/prover/utils/types"
	"github.com/stretchr/testify/require"
)

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

func TestMerkleProofs(t *testing.T) {

	// generate witness
	proofs, leafs, root := getMerkleProof(t)

	var witness badnonce.MerkleProofCircuit
	var buf fr.Element

	witness.Proofs.Siblings = make([]frontend.Variable, len(proofs.Siblings))
	for j := 0; j < len(proofs.Siblings); j++ {
		buf.SetBytes(proofs.Siblings[j][:])
		witness.Proofs.Siblings[j] = buf.String()
	}
	witness.Proofs.Path = proofs.Path
	buf.SetBytes(leafs[:])
	witness.Leaf = buf.String()

	buf.SetBytes(root[:])
	witness.Root = buf.String()

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
