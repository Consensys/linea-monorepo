package vortex2

import (
	"hash"
	"runtime"

	"github.com/consensys/accelerated-crypto-monorepo/crypto/state-management/hashtypes"
	"github.com/consensys/accelerated-crypto-monorepo/crypto/state-management/smt"
	"github.com/consensys/accelerated-crypto-monorepo/maths/common/smartvectors"
	"github.com/consensys/accelerated-crypto-monorepo/maths/field"
	"github.com/consensys/accelerated-crypto-monorepo/utils"
	"github.com/consensys/accelerated-crypto-monorepo/utils/parallel"
	"github.com/consensys/accelerated-crypto-monorepo/utils/profiling"
	"github.com/sirupsen/logrus"
)

// Commitment represents a (vanilla-mode) Vortex commitment
type Commitment []field.Element

// MerkleCommitment represents a (merkle-mode) Vortex commitment
type MerkleCommitment field.Element

// Committed matrix
type CommittedMatrix []smartvectors.SmartVector

// Commits to a sequence of columns.
func (params *Params) Commit(ps []smartvectors.SmartVector) (commitment Commitment, committedMatrix CommittedMatrix) {

	// Sanity-check, all the vectors must have the right length
	for i := range ps {
		if ps[i].Len() != params.NbColumns {
			utils.Panic("bad length : expected %v columns but col %v has size %v", params.NbColumns, i, ps[i].Len())
		}
	}

	logrus.Infof("vortex encoding : %v rows of length %v", len(ps), ps[0].Len())

	// The committed matrix is obtained by encoding the input vectors
	// and laying them in rows.
	committedMatrix = make(CommittedMatrix, len(ps))
	parallel.Execute(len(ps), func(start, stop int) {
		for i := start; i < stop; i++ {
			committedMatrix[i] = params.rsEncode(ps[i])
		}
	})

	logrus.Infof("vortex committing : %v rows of length %v", len(committedMatrix), committedMatrix[0].Len())

	// And obtain the hash of the columns
	var colHashes []field.Element
	if !params.HasSisReplacement() {
		colHashes = params.Key.TransversalHash(committedMatrix)
	} else {
		colHashes = params.noSisTransversalHash(committedMatrix)
	}

	logrus.Infof("finished committing to %v rows", len(ps))

	return Commitment(colHashes), committedMatrix
}

// Commit to a sequence of columns and Merkle hash on top of that. Returns the tree
func (p *Params) CommitMerkle(ps []smartvectors.SmartVector) (commitMatrix CommittedMatrix, tree *smt.Tree, colHashes Commitment) {

	if !p.IsMerkleMode() {
		utils.Panic(
			"called CommitMerkle, but the params are not in MerkleMode." +
				"You may have forgotten to call WithMerkleMode over the Params",
		)
	}

	colHashes, commitMatrix = p.Commit(ps)

	// Hash the digest by chunk and build the tree using the chunk hashes as leaves.
	var leaves []hashtypes.Digest

	if !p.HasSisReplacement() {

		stopTimer := profiling.LogTimer("Vortex commit - SIS hashes")

		// Case with SIS, the columns hashes all fit on several field.element
		// in that case, we need to hash them further. before merkleizing them.
		chunkSize := p.Key.Degree
		numChunks := p.NumEncodedCols()
		leaves = make([]smt.Digest, numChunks)

		parallel.Execute(numChunks, func(start, stop int) {
			// Create the hasher in the parallel setting to avoid race conditions.
			hasher := p.HashFunc()
			for chunkID := start; chunkID < stop; chunkID++ {
				startChunk := chunkID * chunkSize
				hasher.Reset()

				for i := 0; i < chunkSize; i++ {
					fbytes := colHashes[startChunk+i].Bytes()
					hasher.Write(fbytes[:])
				}

				// Manually copies the hasher's digest into the leaves to
				// skip a verbose type conversion.
				copy(leaves[chunkID][:], hasher.Sum(nil))
			}
		})

		stopTimer()
	} else {
		leaves = make([]hashtypes.Digest, len(colHashes))
		for i := range leaves {
			leaves[i] = colHashes[i].Bytes()
		}
	}

	stopTimer := profiling.LogTimer("Vortex commit - Merkleize the column hashes")
	// We purposefully ignore the "MaxValue" of the
	// hashtypes.Hasher, because we do not need it.
	tree = smt.BuildComplete(
		leaves,
		func() hashtypes.Hasher {
			return hashtypes.Hasher{Hash: p.HashFunc()}
		},
	)
	stopTimer()
	return commitMatrix, tree, colHashes
}

// Uses the no-sis hash function to hash the columns
func (p *Params) noSisTransversalHash(v []smartvectors.SmartVector) []field.Element {

	// Assert, we are in no-sis mode
	if !p.HasSisReplacement() {
		panic("expected no-sis mode")
	}

	// Assert that all smart-vectors have the same numCols
	numCols := v[0].Len()
	for i := range v {
		if v[i].Len() != numCols {
			utils.Panic("Unexpected : all inputs smart-vectors should have the same length the first one has length %v, but #%v has length %v",
				numCols, i, v[i].Len())
		}
	}

	numRows := len(v)

	res := make([]field.Element, numCols)
	hashers := make([]hash.Hash, runtime.GOMAXPROCS(0))

	parallel.ExecuteThreadAware(
		numCols,
		func(threadID int) {
			hashers[threadID] = p.NoSisHashFunc()
		},
		func(col, threadID int) {
			hasher := hashers[threadID]
			hasher.Reset()
			for row := 0; row < numRows; row++ {
				x := v[row].Get(col)
				xBytes := x.Bytes()
				hasher.Write(xBytes[:])
			}

			digest := hasher.Sum(nil)
			res[col].SetBytes(digest)
		},
	)

	return res

}
