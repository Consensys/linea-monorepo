package vortex

import (
	"hash"
	"runtime"

	"github.com/consensys/zkevm-monorepo/prover/crypto/state-management/hashtypes"
	"github.com/consensys/zkevm-monorepo/prover/crypto/state-management/smt"
	"github.com/consensys/zkevm-monorepo/prover/maths/common/mempool"
	"github.com/consensys/zkevm-monorepo/prover/maths/common/smartvectors"
	"github.com/consensys/zkevm-monorepo/prover/maths/field"
	"github.com/consensys/zkevm-monorepo/prover/utils"
	"github.com/consensys/zkevm-monorepo/prover/utils/parallel"
	"github.com/consensys/zkevm-monorepo/prover/utils/types"
	"github.com/sirupsen/logrus"
)

// MerkleCommitment represents a (merkle-mode) Vortex commitment
type MerkleCommitment field.Element

// EncodedMatrix represents the witness of a Vortex matrix commitment, it is
// represented as an array of rows.
type EncodedMatrix []smartvectors.SmartVector

// Commit to a sequence of columns and Merkle hash on top of that. Returns the
// tree and an array containing the concatenated columns hashes. The final
// short commitment can be obtained from the returned tree as:
//
//	tree.Root()
//
// And can be safely converted to a field Element via
// [field.Element.SetBytesCanonical]
func (p *Params) CommitMerkle(ps []smartvectors.SmartVector) (encodedMatrix EncodedMatrix, tree *smt.Tree, colHashes []field.Element) {

	if len(ps) > p.MaxNbRows {
		utils.Panic("too many rows: %v, capacity is %v\n", len(ps), p.MaxNbRows)
	}

	logrus.Infof("Vortex compiler: RS encoding nrows=%v of ncol=%v to codeword-size=%v", len(ps), p.NbColumns, p.NbColumns*p.BlowUpFactor)
	encodedMatrix = p.encodeRows(ps)
	logrus.Infof("Vortex compiler: RS encoding DONE")
	logrus.Infof("Vortex compiler: SIS hashing nrows=%v of ncol=%v to codeword-size=%v", len(ps), p.NbColumns, p.NbColumns*p.BlowUpFactor)
	colHashes = p.hashColumns(encodedMatrix)
	logrus.Infof("Vortex compiler: SIS hashing DONE")

	logrus.Infof("Vortex compiler: SIS merkle hashing START")
	// Hash the digest by chunk and build the tree using the chunk hashes as leaves.
	var leaves []types.Bytes32

	if !p.HasSisReplacement() {
		leaves = p.hashSisHash(colHashes)
	} else {
		leaves = make([]types.Bytes32, len(colHashes))
		for i := range leaves {
			leaves[i] = colHashes[i].Bytes()
		}
	}

	tree = smt.BuildComplete(
		leaves,
		func() hashtypes.Hasher {
			return hashtypes.Hasher{Hash: p.HashFunc()}
		},
	)
	logrus.Infof("Vortex compiler: SIS merkle hashing DONE")

	return encodedMatrix, tree, colHashes
}

// encodeRows returns the encodes `ps` using Reed-Solomon. ps is interpreted as
// a list of rows of the Vortex witness and encodedMatrix is obtained by
// encoding each of the [smartvectors.SmartVector] it contains separately.
func (params *Params) encodeRows(ps []smartvectors.SmartVector) (encodedMatrix EncodedMatrix) {

	// Sanity-check, all the vectors must have the right length
	for i := range ps {
		if ps[i].Len() != params.NbColumns {
			utils.Panic("Bad length : expected %v columns but col %v has size %v", params.NbColumns, i, ps[i].Len())
		}
	}

	// The pool will be responsible for holding the coefficients that are
	// intermediary steps in creating the rs encoded rows.
	pool := mempool.CreateFromSyncPool(params.NbColumns)

	// The committed matrix is obtained by encoding the input vectors
	// and laying them in rows.
	encodedMatrix = make(EncodedMatrix, len(ps))
	parallel.Execute(len(ps), func(start, stop int) {
		localPool := mempool.WrapsWithMemCache(pool)
		for i := start; i < stop; i++ {
			encodedMatrix[i] = params.rsEncode(ps[i], localPool)
		}
		localPool.TearDown()
	})

	return encodedMatrix
}

// hashColumns returns a slice storing the hashes of the column of
// `encodedMatrix` sequentially.
//
// When SIS is used, `colHashes` stores the concatenation of the SIS hashes.
func (params *Params) hashColumns(encodedMatrix EncodedMatrix) (colHashes []field.Element) {
	// And obtain the hash of the columns
	if !params.HasSisReplacement() {
		return params.Key.TransversalHash(encodedMatrix)
	}
	return params.noSisTransversalHash(encodedMatrix)
}

// hashSisHash is used to hash the individual SIS hashes stored in colHashes.
// The function is reserved for the case where no NoSisHasher is provided to
// parameters of Vortex.
func (p *Params) hashSisHash(colHashes []field.Element) (leaves []types.Bytes32) {

	// Case with SIS, the columns hashes all fit on several field.element
	// in that case, we need to hash them further. before merkleizing them.
	chunkSize := p.Key.OutputSize()
	numChunks := p.NumEncodedCols()
	leaves = make([]types.Bytes32, numChunks)

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

	return leaves
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
