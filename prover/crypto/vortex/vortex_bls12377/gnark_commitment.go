package vortex

import (
	"hash"
	"runtime"

	"github.com/consensys/gnark-crypto/ecc/bls12-377/fr"
	"github.com/consensys/gnark/frontend"
	"github.com/consensys/linea-monorepo/prover/crypto/state-management/smt_bls12377"
	"github.com/consensys/linea-monorepo/prover/crypto/vortex"
	"github.com/consensys/linea-monorepo/prover/maths/common/smartvectors"
	"github.com/consensys/linea-monorepo/prover/maths/field"
	"github.com/consensys/linea-monorepo/prover/maths/field/fext"
	"github.com/consensys/linea-monorepo/prover/maths/field/gnarkfext"
	"github.com/consensys/linea-monorepo/prover/maths/zk"
	"github.com/consensys/linea-monorepo/prover/utils"
	"github.com/consensys/linea-monorepo/prover/utils/parallel"
	"github.com/consensys/linea-monorepo/prover/utils/profiling"
	"github.com/consensys/linea-monorepo/prover/utils/types"
	"github.com/sirupsen/logrus"
)

// Final circuit - commitment using Merkle trees
type VerifyOpeningCircuitMerkleTree struct {
	Proof      GProof              `gnark:",public"`
	Roots      []frontend.Variable `gnark:",public"`
	X          gnarkfext.E4Gen     `gnark:",public"`
	RandomCoin gnarkfext.E4Gen     `gnark:",public"`
	Ys         [][]gnarkfext.E4Gen `gnark:",public"`
	EntryList  []frontend.Variable `gnark:",public"`
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

	verifyCircuit.Proof.MerkleProofs = make([][]smt_bls12377.GnarkProof, len(proof.MerkleProofs))
	for i := 0; i < len(proof.MerkleProofs); i++ {
		verifyCircuit.Proof.MerkleProofs[i] = make([]smt_bls12377.GnarkProof, len(proof.MerkleProofs[i]))
		for j := 0; j < len(proof.MerkleProofs[i]); j++ {
			verifyCircuit.Proof.MerkleProofs[i][j].Siblings = make([]frontend.Variable, len(proof.MerkleProofs[i][j].Siblings))
		}
	}

	verifyCircuit.EntryList = make([]frontend.Variable, len(entryList))

	verifyCircuit.Ys = make([][]gnarkfext.E4Gen, len(ys))
	for i := 0; i < len(ys); i++ {
		verifyCircuit.Ys[i] = make([]gnarkfext.E4Gen, len(ys[i]))
	}

	verifyCircuit.Roots = make([]frontend.Variable, len(roots))

}

// AssignCicuitVariablesWithMerkleTree assign the variables for the verification circuit with Merkle trees
func AssignCicuitVariablesWithMerkleTree(
	verifyCircuit *VerifyOpeningCircuitMerkleTree,
	proof OpeningProof,
	ys [][]fext.Element,
	entryList []int,
	roots []types.Bytes32) {

	frLinComb := make([]fext.Element, proof.LinearCombination.Len())
	proof.LinearCombination.WriteInSliceExt(frLinComb)
	for i := 0; i < proof.LinearCombination.Len(); i++ {
		verifyCircuit.Proof.LinearCombination[i] = zk.ValueOf(frLinComb[i]) //write ext to zk.value
	}

	for i := 0; i < len(proof.Columns); i++ {
		for j := 0; j < len(proof.Columns[i]); j++ {
			for k := 0; k < len(proof.Columns[i][j]); k++ {
				verifyCircuit.Proof.Columns[i][j][k] = zk.ValueOf(proof.Columns[i][j][k])
			}
		}
	}

	var buf fr.Element
	for i := 0; i < len(proof.MerkleProofs); i++ {
		for j := 0; j < len(proof.MerkleProofs[i]); j++ {
			verifyCircuit.Proof.MerkleProofs[i][j].Path = proof.MerkleProofs[i][j].Path
			for k := 0; k < len(proof.MerkleProofs[i][j].Siblings); k++ {
				buf.SetBytes(proof.MerkleProofs[i][j].Siblings[k][:])
				verifyCircuit.Proof.MerkleProofs[i][j].Siblings[k] = buf.String()
			}
		}
	}

	for i := 0; i < len(entryList); i++ {
		verifyCircuit.EntryList[i] = entryList[i]
	}

	for i := 0; i < len(ys); i++ {
		for j := 0; j < len(ys[i]); j++ {
			verifyCircuit.Ys[i][j] = gnarkfext.NewE4Gen(ys[i][j])
		}
	}

	for i := 0; i < len(roots); i++ {
		buf.SetBytes(roots[i][:])
		verifyCircuit.Roots[i] = buf.String()
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
	roots []frontend.Variable,
	proof GProof,
	x gnarkfext.E4Gen,
	ys [][]gnarkfext.E4Gen,
	randomCoin gnarkfext.E4Gen,
	entryList []frontend.Variable,
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

			// TODO@yao: check if GnarkVerifyCommon compute the leaf, that is written by 7 field.Element to one frontend.Variable?
			// maybe add a new variable type for bigfield leaf? check leaf = sum (7 field.Element ) relationship

			// Check the Merkle-proof for the obtained leaf
			smt_bls12377.GnarkVerifyMerkleProof(api, proof.MerkleProofs[i][j], leaf, root, hasher)

			// And check that the Merkle proof is related to the correct entry
			api.AssertIsEqual(proof.MerkleProofs[i][j].Path, entry)
		}
	}

	return nil
}

/// below copy from github.com/consensys/linea-monorepo/prover/crypto/vortex/commitment.go

// Commit to a sequence of columns and Merkle hash on top of that. Returns the
// tree and an array containing the concatenated columns hashes. The final
// short commitment can be obtained from the returned tree as:
//
//	tree.Root()
//
// We apply Poseidon2 hashing on the columns to compute leaves.
// Should be used when the number of rows to commit is less than the [ApplySISThreshold]
func (p *Params) CommitMerkleWithoutSIS(encodedMatrix vortex.EncodedMatrix) (tree *smt_bls12377.Tree, colHashes []fr.Element) {

	timeTree := profiling.TimeIt(func() {
		// colHashes stores the Poseidon2 hashes
		// of the columns.
		colHashes = p.noSisTransversalHash(encodedMatrix)
		leaves := make([]types.Bytes32, len(colHashes))
		for i := range leaves {
			leaves[i] = colHashes[i].Bytes()
		}

		tree = smt_bls12377.BuildComplete(
			leaves,
			p.MerkleHashFunc,
		)
	})

	logrus.Infof(
		"[vortex-commitment-without-sis] numCol=%v numRow=%v numColEncoded=%v timeMerkleizing=%v",
		p.NbColumns, len(encodedMatrix), p.NumEncodedCols(), timeTree,
	)

	return tree, colHashes
}

// Uses the no-sis hash function to hash the columns
func (p *Params) noSisTransversalHash(v []smartvectors.SmartVector) []fr.Element {

	// Assert that all smart-vectors have the same numCols
	numCols := v[0].Len()
	for i := range v {
		if v[i].Len() != numCols {
			utils.Panic("Unexpected : all inputs smart-vectors should have the same length the first one has length %v, but #%v has length %v",
				numCols, i, v[i].Len())
		}
	}

	numRows := len(v)

	res := make([]fr.Element, numCols)
	hashers := make([]hash.Hash, runtime.GOMAXPROCS(0))

	parallel.ExecuteThreadAware(
		numCols,
		func(threadID int) {
			hashers[threadID] = p.LeafHashFunc()
		},
		func(col, threadID int) {
			hasher := hashers[threadID]
			hasher.Reset()
			xElems := make([]field.Element, numRows)
			for row := 0; row < numRows; row++ {
				xElems[row] = v[row].Get(col)
			}
			colBytes := EncodeKoalabearsToBytes(xElems)
			hasher.Write(colBytes)

			digest := hasher.Sum(nil)
			res[col].SetBytes(digest)
		},
	)

	return res
}
