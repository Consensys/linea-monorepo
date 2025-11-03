package smt

import (
	"testing"

	"github.com/consensys/gnark-crypto/ecc"
	"github.com/consensys/gnark/frontend"
	"github.com/consensys/gnark/frontend/cs/scs"
	gmimc "github.com/consensys/gnark/std/hash/mimc"
	"github.com/consensys/linea-monorepo/prover/crypto/poseidon2"
	"github.com/consensys/linea-monorepo/prover/maths/field"
	"github.com/stretchr/testify/require"
)

func getMerkleProof(t *testing.T) ([]Proof, []field.Octuplet, field.Octuplet) {

	config := &Config{
		HashFunc: poseidon2.Poseidon2,
		Depth:    40,
	}

	tree := NewEmptyTree(config)

	// Only contains empty leaves
	nbProofs := 1
	proofs := make([]Proof, nbProofs)
	leafs := make([]field.Octuplet, nbProofs)
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
	Proofs []GnarkProof           `gnark:",public"`
	Leafs  [][8]frontend.Variable `gnark:",public"`
	Root   [8]frontend.Variable
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

func TestMerkleProofGnark(t *testing.T) {

	// generate witness
	proofs, leafs, root := getMerkleProof(t)
	nbProofs := len(proofs)
	var witness MerkleProofCircuit
	witness.Proofs = make([]GnarkProof, nbProofs)
	witness.Leafs = make([][8]frontend.Variable, nbProofs)
	for i := 0; i < nbProofs; i++ {
		witness.Proofs[i].Siblings = make([][8]frontend.Variable, len(proofs[i].Siblings))
		for j := 0; j < len(proofs[i].Siblings); j++ {
			for k := 0; k < 8; k++ {
				witness.Proofs[i].Siblings[j][k] = proofs[i].Siblings[j][k]
			}
		}
		witness.Proofs[i].Path = proofs[i].Path
		for k := 0; k < 8; k++ {
			witness.Leafs[i][k] = leafs[i][k].String()
		}
	}
	for k := 0; k < 8; k++ {
		witness.Root[k] = root[k].String()
	}
	// compile circuit
	var circuit MerkleProofCircuit
	circuit.Proofs = make([]GnarkProof, nbProofs)
	circuit.Leafs = make([][8]frontend.Variable, nbProofs)
	for i := 0; i < nbProofs; i++ {
		circuit.Proofs[i].Siblings = make([][8]frontend.Variable, len(proofs[i].Siblings))
	}
	ccs, err := frontend.Compile(ecc.BLS12_377.ScalarField(), scs.NewBuilder, &circuit)
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
