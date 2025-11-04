package smt

import (
	"testing"

	"github.com/consensys/gnark-crypto/ecc"
	"github.com/consensys/gnark/frontend"
	"github.com/consensys/gnark/frontend/cs/scs"
	"github.com/consensys/linea-monorepo/prover/crypto/poseidon2"

	"github.com/consensys/linea-monorepo/prover/maths/field"
	"github.com/consensys/linea-monorepo/prover/maths/zk"

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

		leafs[pos], _ = tree.GetLeaf(pos)
		proofs[pos], _ = tree.Prove(pos)

		// Directly verify the proof
		valid := proofs[pos].Verify(config, leafs[pos], tree.Root)
		require.Truef(t, valid, "pos #%v, proof #%v", pos, proofs[pos])
	}

	return proofs, leafs, tree.Root
}

type MerkleProofCircuit struct {
	Proofs []GnarkProof      `gnark:",public"`
	Leafs  []poseidon2.GHash `gnark:",public"`
	Root   poseidon2.GHash
}

func (circuit *MerkleProofCircuit) Define(api frontend.API) error {

	h, err := poseidon2.NewGnarkHasher(api)
	if err != nil {
		return err
	}
	for i := 0; i < len(circuit.Proofs); i++ {
		GnarkVerifyMerkleProof(api, circuit.Proofs[i], circuit.Leafs[i], circuit.Root, h)
	}
	return nil
}

func getCircuitAndWitness(t *testing.T) (MerkleProofCircuit, MerkleProofCircuit) {

	// generate witness
	proofs, leafs, root := getMerkleProof(t)
	nbProofs := len(proofs)
	var witness MerkleProofCircuit
	witness.Proofs = make([]GnarkProof, nbProofs)
	witness.Leafs = make([]poseidon2.GHash, nbProofs)
	var buf field.Octuplet
	for i := 0; i < nbProofs; i++ {
		witness.Proofs[i].Siblings = make([]poseidon2.GHash, len(proofs[i].Siblings))
		for j := 0; j < len(proofs[i].Siblings); j++ {
			buf = proofs[i].Siblings[j]
			for k := 0; k < 8; k++ {
				witness.Proofs[i].Siblings[j][k] = zk.ValueOf(buf[k])
			}
		}
		witness.Proofs[i].Path = zk.ValueOf(proofs[i].Path)
		buf = leafs[i]
		for k := 0; k < 8; k++ {
			witness.Leafs[i][k] = zk.ValueOf(buf[k])
		}
	}
	buf = root
	for k := 0; k < 8; k++ {
		witness.Root[k] = zk.ValueOf(buf[k])
	}

	// compile circuit
	var circuit MerkleProofCircuit
	circuit.Proofs = make([]GnarkProof, nbProofs)
	circuit.Leafs = make([]poseidon2.GHash, nbProofs)
	for i := 0; i < nbProofs; i++ {
		circuit.Proofs[i].Siblings = make([]poseidon2.GHash, len(proofs[i].Siblings))
	}

	return circuit, witness

}

func TestMerkleProofGnark(t *testing.T) {

	{
		circuit, witness := getCircuitAndWitness(t)

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

}
