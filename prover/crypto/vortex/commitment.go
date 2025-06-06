package vortex

import (
	"github.com/consensys/gnark-crypto/field/koalabear"
	"github.com/consensys/gnark-crypto/field/koalabear/vortex"
	"github.com/consensys/linea-monorepo/prover/crypto/ringsis"
	"github.com/consensys/linea-monorepo/prover/crypto/state-management/hashtypes"
	"github.com/consensys/linea-monorepo/prover/crypto/state-management/smt"
	"github.com/consensys/linea-monorepo/prover/maths/common/mempool"
	"github.com/consensys/linea-monorepo/prover/maths/common/smartvectors"
	"github.com/consensys/linea-monorepo/prover/maths/field"
	"github.com/consensys/linea-monorepo/prover/utils"
	"github.com/consensys/linea-monorepo/prover/utils/parallel"
	"github.com/consensys/linea-monorepo/prover/utils/profiling"
	"github.com/consensys/linea-monorepo/prover/utils/types"
	"github.com/sirupsen/logrus"
)

// EncodedMatrix represents the witness of a Vortex matrix commitment, it is
// represented as an array of rows.
type EncodedMatrix []smartvectors.SmartVector

func (p *Params) Commit(ps []smartvectors.SmartVector) (encodedMatrix EncodedMatrix, tree *smt.Tree, colHashes []field.Element) {

	if len(ps) > p.MaxNbRows {
		utils.
			Panic("too many rows: %v, capacity is %v\n", len(ps), p.MaxNbRows)
	}
	encodedMatrix = p.encodeRows(ps)
	nbColumns := p.NumEncodedCols()

	if p.ColumnHasher == nil { // by default, we use sis
		colHashes = sisTransversalHash(p.Key, encodedMatrix)
	} else {
		colHashes = p.transversalHash(encodedMatrix)
	}
	// at this stage colHashes is a list of field.Element

	leaves := make([]types.Bytes32, nbColumns)
	sizeChunk := len(colHashes) / nbColumns

	if p.MerkleHasher == nil { // by default, we use poseidon2
		for i := 0; i < nbColumns; i++ {
			h := vortex.HashPoseidon2(colHashes[i*sizeChunk : (i+1)*sizeChunk])
			leaves[i] = types.HashToBytes32(h)
		}

	} else {
		merkleHasher := p.MerkleHasher()
		for i := 0; i < nbColumns; i++ {
			merkleHasher.Reset()
			for j := 0; j < sizeChunk; j++ {
				merkleHasher.Write(colHashes[i*sizeChunk+j].Marshal())
			}
			h := merkleHasher.Sum(nil)
			copy(leaves[i][:], h)
		}
	}

	tree = smt.BuildComplete(
		leaves,
		func() hashtypes.Hasher {
			return hashtypes.Hasher{Hash: p.MerkleHasher()}
		},
	)

	return encodedMatrix, tree, colHashes
}

func (p *Params) computeLeaves(colHashes []field.Element) []types.Bytes32 {

	nbElmts := p.NumEncodedCols()
	// TODO handle panic
	if len(colHashes)%nbElmts != 0 {
		panic("error computeLeavesPoseidon2")
	}
	sizeChunk := len(colHashes) / nbElmts
	res := make([]types.Bytes32, nbElmts)
	parallel.Execute(nbElmts, func(start, end int) {
		h := p.ColumnHasher()
		for i := start; i < end; i++ {
			h.Reset()
			s := i * sizeChunk
			for j := 0; j < sizeChunk; j++ {
				h.Write(colHashes[s].Marshal())
				s++
			}
			tmp := h.Sum(nil)
			copy(res[i][:], tmp)
		}
	})

	return res
}

func (p *Params) computeLeavesPoseidon2(colHashes []field.Element) []types.Bytes32 {

	nbElmts := p.NumEncodedCols()

	// TODO handle panic
	if len(colHashes)%nbElmts != 0 {
		panic("error computeLeavesPoseidon2")
	}
	res := make([]types.Bytes32, nbElmts)
	sizeChunk := len(colHashes) / nbElmts
	parallel.Execute(nbElmts, func(start, end int) {
		for i := start; i < end; i++ {
			h := vortex.HashPoseidon2(colHashes[i*sizeChunk : (i+1)*sizeChunk])
			res[i] = types.HashToBytes32(h)
		}
	})

	return res
}

// Uses the no-sis hash function to hash the columns
func (p *Params) transversalHash(v []smartvectors.SmartVector) []field.Element {

	nbRows := len(v)
	nbCols := p.NumEncodedCols()
	res := make([]field.Element, 8*nbCols) // each entry stores 8 koalabear elmts
	parallel.Execute(nbCols, func(start, stop int) {
		hasher := p.ColumnHasher()
		for col := start; col < stop; col++ {
			hasher.Reset()
			for row := 0; row < nbRows; row++ {
				cur := v[row].Get(col)
				curBytes := cur.Bytes()
				hasher.Write(curBytes[:])
			}
			h := hasher.Sum(nil)
			for j := 0; j < 8; j++ {
				res[col*8+j].SetBytes(h[j*8 : (j+1)*8])
			}
		}
	})

	return res
}

func sisTransversalHash(key ringsis.Key, m EncodedMatrix) []field.Element {

	nbRows := len(m)
	nbCols := m[0].Len()
	buffer := make([]koalabear.Element, nbRows)
	sisOutputSize := key.OutputSize()
	res := make([]field.Element, sisOutputSize*nbCols)
	parallel.Execute(nbCols, func(start, end int) {
		for i := start; i < end; i++ {
			for j := 0; j < nbRows; j++ {
				buffer[j] = m[j].Get(i)
			}
			hashCol := key.Hash(buffer)
			copy(res[i*sisOutputSize:(i+1)*sisOutputSize], hashCol)
		}
	})

	return res
}

// Commit to a sequence of columns and Merkle hash on top of that. Returns the
// tree and an array containing the concatenated columns hashes. The final
// short commitment can be obtained from the returned tree as:
//
//	tree.Root()
//
// And can be safely converted to a field Element via
// [field.Element.SetBytesCanonical]
// We apply SIS+MiMC hashing on the columns to compute leaves
// Should be used when the number of rows to commit is more than the [ApplySISThreshold]
func (p *Params) CommitMerkleWithSIS(ps []smartvectors.SmartVector) (encodedMatrix EncodedMatrix, tree *smt.Tree, colHashes []field.Element) {

	if len(ps) > p.MaxNbRows {
		utils.Panic("too many rows: %v, capacity is %v\n", len(ps), p.MaxNbRows)
	}

	timeEncoding := profiling.TimeIt(func() {
		encodedMatrix = p.encodeRows(ps)
	})
	timeSisHashing := profiling.TimeIt(func() {
		// colHashes stores concatenation of SIS+MiMC hashes of the columns
		// if isSISAppliedForRound is true, otherwise it stores the MiMC hashes
		// of the columns.
		colHashes = p.Key.TransversalHash(encodedMatrix)
	})

	timeTree := profiling.TimeIt(func() {
		// Hash the SIS digests to obtain the leaves of the Merkle tree.
		leaves := p.computeLeavesWithSis(colHashes)

		tree = smt.BuildComplete(
			leaves,
			func() hashtypes.Hasher {
				return hashtypes.Hasher{Hash: p.MerkleHasher()}
			},
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
// We apply MiMC hashing on the columns to compute leaves.
// Should be used when the number of rows to commit is less than the [ApplySISThreshold]
// TODO why are colHashes
// func (p *Params) CommitMerkleWithoutSIS(ps []smartvectors.SmartVector) (encodedMatrix EncodedMatrix, tree *smt.Tree, colHashes []field.Element) {

// 	if len(ps) > p.MaxNbRows {
// 		utils.Panic("too many rows: %v, capacity is %v\n", len(ps), p.MaxNbRows)
// 	}

// 	timeEncoding := profiling.TimeIt(func() {
// 		encodedMatrix = p.encodeRows(ps)
// 	})

// 	timeTree := profiling.TimeIt(func() {
// 		// colHashes stores the MiMC hashes
// 		// of the columns.
// 		leaves := p.transversalHash(encodedMatrix)

// 		tree = smt.BuildComplete(
// 			leaves,
// 			func() hashtypes.Hasher {
// 				return hashtypes.Hasher{Hash: p.MerkleHasher()}
// 			},
// 		)
// 	})

// 	logrus.Infof(
// 		"[vortex-commitment-without-sis] numCol=%v numRow=%v numColEncoded=%v timeEncoding=%v timeMerkleizing=%v",
// 		p.NbColumns, len(ps), p.NumEncodedCols(), timeEncoding, timeTree,
// 	)

// 	return encodedMatrix, tree, colHashes
// }

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

// computeLeavesWithSis is used to hash the individual SIS hashes stored in colHashes.
// The function is reserved for the case where no NoSisHasher is provided to
// parameters of Vortex.
func (p *Params) computeLeavesWithSis(colHashes []field.Element) (leaves []types.Bytes32) {

	// Case with SIS, the columns hashes all fit on several field.element
	// in that case, we need to hash them further. before merkleizing them.
	chunkSize := p.Key.OutputSize()
	numChunks := p.NumEncodedCols()
	leaves = make([]types.Bytes32, numChunks)

	parallel.Execute(numChunks, func(start, stop int) {
		hasher := p.ColumnHasher()
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
