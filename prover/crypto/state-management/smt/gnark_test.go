package smt

import (
	"testing"

	"github.com/consensys/accelerated-crypto-monorepo/crypto/state-management/hashtypes"
	"github.com/consensys/gnark-crypto/ecc"
	"github.com/consensys/gnark-crypto/ecc/bn254/fr"
	"github.com/consensys/gnark/frontend"
	"github.com/consensys/gnark/frontend/cs/scs"
	gmimc "github.com/consensys/gnark/std/hash/mimc"
	"github.com/stretchr/testify/require"
)

func getMerkleProof(t *testing.T) ([]Proof, []hashtypes.Digest, hashtypes.Digest) {

	config := &Config{
		HashFunc: hashtypes.MiMC,
		Depth:    40,
	}

	tree := NewEmptyTree(config)

	// Only contains empty leaves
	nbProofs := 1
	proofs := make([]Proof, nbProofs)
	leafs := make([]hashtypes.Digest, nbProofs)
	for pos := 0; pos < nbProofs; pos++ {

		// Make a valid digest
		leafs[pos] = tree.GetLeaf(pos)
		proofs[pos] = tree.Prove(pos)

		// Directly verify the proof
		valid := proofs[pos].Verify(config, leafs[pos], tree.Root)
		require.Truef(t, valid, "pos #%v, proof #%v", pos, proofs[pos])
	}

	return proofs, leafs, tree.Root
}

type MerkleProofCircuit struct {
	Proofs []GnarkProof        `gnark:"public"`
	Leafs  []frontend.Variable `gnark:"public"`
	Root   frontend.Variable
}

func (circuit *MerkleProofCircuit) Define(api frontend.API) error {

	h, err := gmimc.NewMiMC(api)
	if err != nil {
		return err
	}
	for i := 0; i < len(circuit.Proofs); i++ {
		GnarkVerifyMerkleProof(api, circuit.Proofs[i], circuit.Leafs[i], circuit.Root, &h)
	}
	return nil
}

func TestMerkleProof(t *testing.T) {

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
			buf.SetBytes(proofs[i].Siblings[j][:])
			witness.Proofs[i].Siblings[j] = buf.String()
		}
		witness.Proofs[i].Path = proofs[i].Path
		buf.SetBytes(leafs[i][:])
		witness.Leafs[i] = buf.String()
	}
	buf.SetBytes(root[:])
	witness.Root = buf.String()

	// compile circuit
	var circuit MerkleProofCircuit
	circuit.Proofs = make([]GnarkProof, nbProofs)
	circuit.Leafs = make([]frontend.Variable, nbProofs)
	for i := 0; i < nbProofs; i++ {
		circuit.Proofs[i].Siblings = make([]frontend.Variable, len(proofs[i].Siblings))
	}
	ccs, err := frontend.Compile(ecc.BN254.ScalarField(), scs.NewBuilder, &circuit, frontend.IgnoreUnconstrainedInputs())
	if err != nil {
		t.Fatal(err)
	}

	// solve the circuit
	twitness, err := frontend.NewWitness(&witness, ecc.BN254.ScalarField())
	if err != nil {
		t.Fatal(err)
	}
	err = ccs.IsSolved(twitness)
	if err != nil {
		t.Fatal(err)
	}

}
