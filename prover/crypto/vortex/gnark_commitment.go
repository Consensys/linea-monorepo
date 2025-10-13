package vortex

import (
	"github.com/consensys/gnark/frontend"
	"github.com/consensys/linea-monorepo/prover/crypto/state-management/smt"
	"github.com/consensys/linea-monorepo/prover/maths/field"
	"github.com/consensys/linea-monorepo/prover/maths/field/fext"
	"github.com/consensys/linea-monorepo/prover/maths/zk"
	"github.com/consensys/linea-monorepo/prover/utils"
	"github.com/consensys/linea-monorepo/prover/utils/types"
)

// Final circuit - commitment using Merkle trees
type VerifyOpeningCircuitMerkleTree struct {
	Proof      GProof                 `gnark:",public"`
	Roots      []zk.WrappedVariable   `gnark:",public"`
	X          zk.WrappedVariable     `gnark:",public"`
	RandomCoin zk.WrappedVariable     `gnark:",public"`
	Ys         [][]zk.WrappedVariable `gnark:",public"`
	EntryList  []zk.WrappedVariable   `gnark:",public"`
	Params     GParams
}

// allocate the variables for the verification circuit with Merkle trees
func AllocateCircuitVariablesWithMerkleTree(
	verifyCircuit *VerifyOpeningCircuitMerkleTree,
	proof OpeningProof,
	ys [][]fext.Element,
	entryList []int,
	roots []types.Bytes32) {

	verifyCircuit.Proof.LinearCombination = make([]zk.WrappedVariable, proof.LinearCombination.Len())

	verifyCircuit.Proof.Columns = make([][][]zk.WrappedVariable, len(proof.Columns))
	for i := 0; i < len(proof.Columns); i++ {
		verifyCircuit.Proof.Columns[i] = make([][]zk.WrappedVariable, len(proof.Columns[i]))
		for j := 0; j < len(proof.Columns[i]); j++ {
			verifyCircuit.Proof.Columns[i][j] = make([]zk.WrappedVariable, len(proof.Columns[i][j]))
		}
	}

	verifyCircuit.Proof.MerkleProofs = make([][]smt.GnarkProof, len(proof.MerkleProofs))
	for i := 0; i < len(proof.MerkleProofs); i++ {
		verifyCircuit.Proof.MerkleProofs[i] = make([]smt.GnarkProof, len(proof.MerkleProofs[i]))
		for j := 0; j < len(proof.MerkleProofs[i]); j++ {
			verifyCircuit.Proof.MerkleProofs[i][j].Siblings = make([]zk.WrappedVariable, len(proof.MerkleProofs[i][j].Siblings))
		}
	}

	verifyCircuit.EntryList = make([]zk.WrappedVariable, len(entryList))

	verifyCircuit.Ys = make([][]zk.WrappedVariable, len(ys))
	for i := 0; i < len(ys); i++ {
		verifyCircuit.Ys[i] = make([]zk.WrappedVariable, len(ys[i]))
	}

	verifyCircuit.Roots = make([]zk.WrappedVariable, len(roots))

}

// AssignCicuitVariablesWithMerkleTree assign the variables for the verification circuit with Merkle trees
func AssignCicuitVariablesWithMerkleTree(
	verifyCircuit *VerifyOpeningCircuitMerkleTree,
	proof OpeningProof,
	ys [][]fext.Element,
	entryList []int,
	roots []types.Bytes32) {

	frLinComb := make([]field.Element, proof.LinearCombination.Len())
	proof.LinearCombination.WriteInSlice(frLinComb)
	for i := 0; i < proof.LinearCombination.Len(); i++ {
		verifyCircuit.Proof.LinearCombination[i] = zk.ValueOf(frLinComb[i])
	}

	for i := 0; i < len(proof.Columns); i++ {
		for j := 0; j < len(proof.Columns[i]); j++ {
			for k := 0; k < len(proof.Columns[i][j]); k++ {
				verifyCircuit.Proof.Columns[i][j][k] = zk.ValueOf(proof.Columns[i][j][k])
			}
		}
	}

	var buf field.Element
	for i := 0; i < len(proof.MerkleProofs); i++ {
		for j := 0; j < len(proof.MerkleProofs[i]); j++ {
			verifyCircuit.Proof.MerkleProofs[i][j].Path = zk.ValueOf(proof.MerkleProofs[i][j].Path)
			for k := 0; k < len(proof.MerkleProofs[i][j].Siblings); k++ {
				buf.SetBytes(proof.MerkleProofs[i][j].Siblings[k][:])
				verifyCircuit.Proof.MerkleProofs[i][j].Siblings[k] = zk.ValueOf(buf)
			}
		}
	}

	for i := 0; i < len(entryList); i++ {
		verifyCircuit.EntryList[i] = zk.ValueOf(entryList[i])
	}

	for i := 0; i < len(ys); i++ {
		for j := 0; j < len(ys[i]); j++ {
			verifyCircuit.Ys[i][j] = zk.ValueOf(ys[i][j])
		}
	}

	for i := 0; i < len(roots); i++ {
		buf.SetBytes(roots[i][:])
		verifyCircuit.Roots[i] = zk.ValueOf(buf)
	}

}

func (circuit *VerifyOpeningCircuitMerkleTree) Define(api frontend.API) error {

	// Generic checks
	err := GnarkVerifyOpeningWithMerkleProof(
		api,
		circuit.Params,
		circuit.Roots,
		circuit.Proof,
		circuit.X,
		circuit.Ys,
		circuit.RandomCoin,
		circuit.EntryList,
	)
	return err
}

func GnarkVerifyOpeningWithMerkleProof(
	api frontend.API,
	params GParams,
	roots []zk.WrappedVariable,
	proof GProof,
	x zk.WrappedVariable,
	ys [][]zk.WrappedVariable,
	randomCoin zk.WrappedVariable,
	entryList []zk.WrappedVariable,
) error {

	if !params.HasNoSisHasher() {
		utils.Panic("the verifier circuit can only be instantiated using a NoSisHasher")
	}

	// Generic checks
	selectedColsHashes, err := GnarkVerifyCommon(
		api,
		params,
		proof.GProofWoMerkle,
		x,
		ys,
		randomCoin,
		entryList,
	)
	if err != nil {
		return err
	}

	hasher, _ := params.HasherFunc(api)
	hasher.Reset()

	for i, root := range roots {
		for j, entry := range entryList {

			// Hash the SIS hash
			var leaf = selectedColsHashes[i][j]

			// Check the Merkle-proof for the obtained leaf
			smt.GnarkVerifyMerkleProof(api, proof.MerkleProofs[i][j], leaf, root, hasher)

			// And check that the Merkle proof is related to the correct entry
			api.AssertIsEqual(proof.MerkleProofs[i][j].Path, entry)
		}
	}

	return nil
}
