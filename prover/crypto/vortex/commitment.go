package vortex

import (
	"github.com/consensys/linea-monorepo/prover/crypto/poseidon2"
	"github.com/consensys/linea-monorepo/prover/crypto/state-management/smt"
	"github.com/consensys/linea-monorepo/prover/maths/common/smartvectors"
	"github.com/consensys/linea-monorepo/prover/maths/field"
	"github.com/consensys/linea-monorepo/prover/utils"
	"github.com/consensys/linea-monorepo/prover/utils/parallel"
	"github.com/consensys/linea-monorepo/prover/utils/profiling"
	"github.com/sirupsen/logrus"
)

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
// We apply SIS+Poseidon2 hashing on the columns to compute leaves
// Should be used when the number of rows to commit is more than the [ApplySISThreshold]
func (p *Params) CommitMerkleWithSIS(ps []smartvectors.SmartVector) (encodedMatrix EncodedMatrix, tree *smt.Tree, colHashes []field.Element) {

	if len(ps) > p.MaxNbRows {
		utils.
			Panic("too many rows: %v, capacity is %v\n", len(ps), p.MaxNbRows)
	}

	chAlloc := make(chan struct{}, 1)
	go func() {
		// this is a large alloc, we start before doing the encoding.
		// TODO @gbotrel not particularly efficient -- this still appears in the trace.
		// we should investigate a way to pre-allocate that and re-use it, seems useful to keep mostly
		// for self recursion, maybe in a context linked to that?
		colHashes = make([]field.Element, p.NumEncodedCols()*p.Key.OutputSize())
		close(chAlloc)
	}()

	timeEncoding := profiling.TimeIt(func() {
		encodedMatrix = p.encodeRows(ps)
	})
	timeSisHashing := profiling.TimeIt(func() {
		// colHashes stores concatenation of SIS hashes of the columns
		<-chAlloc
		colHashes = p.Key.TransversalHash(encodedMatrix, colHashes)
	})

	timeTree := profiling.TimeIt(func() {
		// Hash the SIS digests to obtain the leaves of the Merkle tree.
		leavesOcts := p.hashSisHash(colHashes)

		tree = smt.BuildComplete(
			leavesOcts,
		)
	})

	logrus.Infof(
		"[vortex-commitment-with-sis] numCol=%v numRow=%v numColEncoded=%v timeEncoding=%v timeSisHashing=%v timeMerkleizing=%v",
		p.NbColumns, len(ps), p.NumEncodedCols(), timeEncoding, timeSisHashing, timeTree,
	)

	return encodedMatrix, tree, colHashes
}

// Commit to a sequence of columns and Merkle hash on top of that. Returns the
// tree and an array containing the concatenated columns hashes. The final
// short commitment can be obtained from the returned tree as:
//
//	tree.Root()
//
// And can be safely converted to a field Element via
// [field.Element.SetBytesCanonical]
// We apply Poseidon2 hashing on the columns to compute leaves.
// Should be used when the number of rows to commit is less than the [ApplySISThreshold]
func (p *Params) CommitMerkleWithoutSIS(ps []smartvectors.SmartVector) (encodedMatrix EncodedMatrix, tree *smt.Tree, colHashes []field.Element) {

	if len(ps) > p.MaxNbRows {
		utils.Panic("too many rows: %v, capacity is %v\n", len(ps), p.MaxNbRows)
	}

	timeEncoding := profiling.TimeIt(func() {
		encodedMatrix = p.encodeRows(ps)
	})

	timeTree := profiling.TimeIt(func() {
		// colHashes stores the Poseidon2 hashes
		// of the columns.
		colHashesOcts := p.noSisTransversalHash(encodedMatrix)

		tree = smt.BuildComplete(
			colHashesOcts,
		)
	})

	logrus.Infof(
		"[vortex-commitment-without-sis] numCol=%v numRow=%v numColEncoded=%v timeEncoding=%v timeMerkleizing=%v",
		p.NbColumns, len(ps), p.NumEncodedCols(), timeEncoding, timeTree,
	)

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

	// The committed matrix is obtained by encoding the input vectors
	// and laying them in rows.
	encodedMatrix = make(EncodedMatrix, len(ps))
	parallel.Execute(len(ps), func(start, stop int) {
		for i := start; i < stop; i++ {
			encodedMatrix[i] = params._rsEncodeBase(ps[i])
		}
	})

	return encodedMatrix
}

// hashSisHash is used to hash the individual SIS hashes stored in colHashes.
// The function is reserved for the case where no NoSisHasher is provided to
// parameters of Vortex.
func (p *Params) hashSisHash(colHashes []field.Element) (leaves []field.Octuplet) {

	// Case with SIS, the columns hashes all fit on several field.element
	// in that case, we need to hash them further. before merkleizing them.
	chunkSize := p.Key.OutputSize()
	numChunks := p.NumEncodedCols()
	leaves = make([]field.Octuplet, numChunks)

	parallel.Execute(numChunks, func(start, stop int) {
		for chunkID := start; chunkID < stop; chunkID++ {
			startChunk := chunkID * chunkSize

			if p.LeafHashFunc != nil {
				panic("Default poseidon2 hash function is provided")
			} else {
				// Default LeafHashFunc: Using Poseidon2Sponge directly to avoid data conversion.
				leaves[chunkID] = poseidon2.Poseidon2Sponge(colHashes[startChunk : startChunk+chunkSize])
			}
		}

	})

	return leaves
}

// Uses the no-sis hash function to hash the columns. It uses the leafHasher
// function to hash the columns.
func (p *Params) noSisTransversalHash(v []smartvectors.SmartVector) []field.Octuplet {

	// Assert that all smart-vectors have the same numCols
	numCols := v[0].Len()
	for i := range v {
		if v[i].Len() != numCols {
			utils.Panic("Unexpected : all inputs smart-vectors should have the same length the first one has length %v, but #%v has length %v",
				numCols, i, v[i].Len())
		}
	}

	numRows := len(v)

	res := make([]field.Octuplet, numCols)

	if p.LeafHashFunc != nil {
		panic("Default poseidon2 hash function is provided")
	} else {
		// Default LeafHashFunc: Using Poseidon2Sponge directly to avoid data conversion.
		parallel.ExecuteThreadAware(
			numCols,
			func(threadID int) {
			},
			func(col, threadID int) {
				colElems := make([]field.Element, numRows)
				for row := 0; row < numRows; row++ {
					colElems[row] = v[row].Get(col)
				}
				res[col] = poseidon2.Poseidon2Sponge(colElems)
			},
		)
	}

	return res
}
