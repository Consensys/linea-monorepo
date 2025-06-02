package vortex

import (
	"hash"
	"runtime"

	"github.com/consensys/gnark-crypto/field/koalabear/vortex"

	"github.com/consensys/linea-monorepo/prover/crypto/state-management/hashtypes"
	"github.com/consensys/linea-monorepo/prover/crypto/state-management/smt"
	"github.com/consensys/linea-monorepo/prover/maths/common/smartvectors"
	"github.com/consensys/linea-monorepo/prover/maths/field"
	"github.com/consensys/linea-monorepo/prover/utils"
	"github.com/consensys/linea-monorepo/prover/utils/parallel"
	"github.com/consensys/linea-monorepo/prover/utils/types"
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
func (p *Params) CommitMerkle(ps []smartvectors.SmartVector) (encodedMatrix EncodedMatrix, tree *smt.Tree, colHashes []field.Element) {

	if len(ps) > p.MaxNbRows {
		utils.Panic("too many rows: %v, capacity is %v\n", len(ps), p.MaxNbRows)
	}

	numRows := len(ps)
	numCols := utils.NextPowerOfTwo(p.NbColumns)
	sizeCodeWord := p.NumEncodedCols()

	logrus.Infof("Vortex compiler: RS encoding nrows=%v of ncol=%v to codeword-size=%v", numRows, numCols, numCols*p.BlowUpFactor)

	input := make([][]field.Element, numRows)
	parallel.Execute(numRows, func(start, end int) {
		for i := start; i < end; i++ {
			input[i] = make([]field.Element, numCols)
			for j := 0; j < numCols; j++ {
				input[i][j] = ps[i].Get(j)
			}
		}
	})

	// In Commit phase, it's not used, so set to 0 as a placeholder. numSelectedColumns is only used in the Open phase

	options := make([]vortex.Option, 0, 2)
	if p.HashFunc != nil {
		options = append(options, vortex.WithMerkleHash(p.HashFunc()))
	}
	if p.NoSisHashFunc != nil {
		options = append(options, vortex.WithNoSis(p.NoSisHashFunc()))
	}

	logrus.Infof("Vortex compiler: SIS hashing DONE")

	logrus.Infof("Vortex compiler: SIS merkle hashing START")
	// Hash the digest by chunk and build the tree using the chunk hashes as leaves.
	var leaves []types.Bytes32

	if !p.HasSisReplacement() {
		// The total length of colHashes is sizeCodeWord * p.Key.GnarkInternal.Degree
		// The total number of leaves is sizeCodeWord,
		// p.Key.GnarkInternal.Degree field.Element values from colHashes are packed to form one Merkle tree leaf.
		leaves = p.hashSisHash(colHashes)
	} else {
		leaves = make([]types.Bytes32, sizeCodeWord)
		for i := range leaves {
			currentChunkSlice := colHashes[8*i : 8*(i+1)]
			var currentChunkArray [8]field.Element
			copy(currentChunkArray[:], currentChunkSlice)
			leaves[i] = types.HashToBytes32(currentChunkArray)
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

// hashSisHash is used to hash the individual SIS hashes stored in colHashes.
// The function is reserved for the case where no NoSisHasher is provided to
// parameters of Vortex.
func (p *Params) hashSisHash(colHashes []field.Element) (bytes32Leaves []types.Bytes32) {

	// Case with SIS, the columns hashes all fit on several field.element
	// in that case, we need to hash them further. before merkleizing them.
	sizeCodeWord := p.NumEncodedCols()

	hashLeaves := make([]vortex.Hash, sizeCodeWord)

	bytes32Leaves = make([]types.Bytes32, sizeCodeWord)
	sisKeySize := p.Key.GnarkInternal.Degree

	// Poseidon2 blocksize 16
	const blockSize = 16
	if sizeCodeWord%blockSize == 0 {
		// we hash by blocks of 16 to leverage optimized SIMD implementation
		// of Poseidon2 which require 16 hashes to be computed independently.
		parallel.Execute(sizeCodeWord/blockSize, func(start, end int) {
			for block := start; block < end; block++ {
				b := block * blockSize
				sStart := b * sisKeySize
				sEnd := sStart + sisKeySize*blockSize
				vortex.HashPoseidon2x16(colHashes[sStart:sEnd], hashLeaves[b:b+blockSize], sisKeySize)
			}
		})
	} else {
		// unusual path; it means we have < 16 columns (tiny code words)
		// so we do the hashes one by one.
		for i := 0; i < sizeCodeWord; i++ {
			sStart := i * sisKeySize
			sEnd := sStart + sisKeySize
			hashLeaves[i] = vortex.HashPoseidon2(colHashes[sStart:sEnd])
		}
	}
	for i := 0; i < sizeCodeWord; i++ {
		bytes32Leaves[i] = types.HashToBytes32(hashLeaves[i])
	}
	return bytes32Leaves
}

// Uses the no-sis hash function to hash the columns
// TODO: we may want to output sizeCodeWord*8 field.Element,
// If we output sizeCodeWord field.Element, then each one mapped to one leaf will be embedding 32 bits into 32 bytes.
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
