package smt_bls12377

import (
	"testing"

	"github.com/consensys/gnark-crypto/ecc"
	"github.com/consensys/gnark-crypto/ecc/bls12-377/fr"
	"github.com/consensys/gnark/frontend"

	"github.com/consensys/gnark/frontend/cs/scs"
	gposeidon2 "github.com/consensys/gnark/std/hash/poseidon2"
	"github.com/consensys/linea-monorepo/prover/crypto/state-management/hashtypes"
	"github.com/consensys/linea-monorepo/prover/utils/types"
	"github.com/stretchr/testify/require"
)

func getMerkleProof(t *testing.T) ([]Proof, []types.Bytes32, types.Bytes32) {

	config := &Config{
		HashFunc: hashtypes.Poseidon2,
		Depth:    40,
	}

	tree := NewEmptyTree(config)

	// populate the tree
	nbLeaves := 10
	var tmp fr.Element
	for i := 0; i < nbLeaves; i++ {
		tmp.SetRandom()
		s := tmp.Bytes()
		buf := types.AsBytes32(s[:])
		tree.Update(i, buf)
	}
	nbProofs := 10
	proofs := make([]Proof, nbProofs)
	leafs := make([]types.Bytes32, nbProofs)
	for pos := 0; pos < nbProofs; pos++ {

		// Make a valid Bytes32
		leafs[pos], _ = tree.GetLeaf(pos)
		proofs[pos], _ = tree.Prove(pos)

		// Directly verify the proof
		valid := proofs[pos].Verify(config, leafs[pos], tree.Root)
		require.Truef(t, valid, "pos #%v, proof #%v", pos, proofs[pos])
	}

	return proofs, leafs, tree.Root
}

type MerkleProofCircuit struct {
	Proofs []GnarkProof        `gnark:",public"`
	Leafs  []frontend.Variable `gnark:",public"`
	Root   frontend.Variable
}

func (circuit *MerkleProofCircuit) Define(api frontend.API) error {

	h, err := gposeidon2.NewMerkleDamgardHasher(api)
	if err != nil {
		return err
	}
	for i := 0; i < len(circuit.Proofs); i++ {
		GnarkVerifyMerkleProof(api, circuit.Proofs[i], circuit.Leafs[i], circuit.Root, h)
	}
	return nil
}

func TestMerkleProofGnark(t *testing.T) {

	// generate witness
	proofs, leafs, root := getMerkleProof(t)
	nbProofs := len(proofs)
	var witness MerkleProofCircuit
	witness.Proofs = make([]GnarkProof, nbProofs)
	witness.Leafs = make([]frontend.Variable, nbProofs)
	var buf fr.Element
	for i := 0; i < nbProofs; i++ {
		witness.Proofs[i].Siblings = make([]frontend.Variable, len(proofs[i].Siblings))
		for j := 0; j < len(proofs[i].Siblings); j++ {
			siblingBytes := proofs[i].Siblings[j]
			buf.SetBytes(siblingBytes[:])
			witness.Proofs[i].Siblings[j] = buf.String()
		}
		witness.Proofs[i].Path = proofs[i].Path
		leafsBytes := leafs[i]
		buf.SetBytes(leafsBytes[:])
		witness.Leafs[i] = buf.String()
	}
	rootBytes := root

	buf.SetBytes(rootBytes[:])
	witness.Root = buf.String()

	// compile circuit
	var circuit MerkleProofCircuit
	circuit.Proofs = make([]GnarkProof, nbProofs)
	circuit.Leafs = make([]frontend.Variable, nbProofs)
	for i := 0; i < nbProofs; i++ {
		circuit.Proofs[i].Siblings = make([]frontend.Variable, len(proofs[i].Siblings))
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
