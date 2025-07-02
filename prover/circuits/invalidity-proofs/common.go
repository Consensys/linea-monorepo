package badnonce

import (
	"github.com/consensys/gnark/frontend"
	gmimc "github.com/consensys/gnark/std/hash/mimc"
	"github.com/consensys/linea-monorepo/prover/crypto/state-management/smt"
)

// MerkleProofCircuit defines the circuit for validating the Merkle proofs
type MerkleProofCircuit struct {
	Proofs smt.GnarkProof
	Leaf   frontend.Variable
	Root   frontend.Variable
}

func (circuit *MerkleProofCircuit) Define(api frontend.API) error {
	h, err := gmimc.NewMiMC(api)
	if err != nil {
		return err
	}

	smt.GnarkVerifyMerkleProof(api, circuit.Proofs, circuit.Leaf, circuit.Root, &h)

	return nil
}

// Circuit defines a pre-image knowledge proof
// mimc( preImage) = public hash
type MimcCircuit struct {
	PreImage []frontend.Variable
	Hash     frontend.Variable
}

// Define declares the circuit's constraints
// Hash = mimc(PreImage)
func (circuit *MimcCircuit) Define(api frontend.API) error {
	// hash function
	mimc, _ := gmimc.NewMiMC(api)

	// mimc(preImage) == hash
	for _, toHash := range circuit.PreImage {
		mimc.Write(toHash)
	}

	api.AssertIsEqual(circuit.Hash, mimc.Sum())

	return nil
}
