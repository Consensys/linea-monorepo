package vortex_koalabear

import (
	vgnark "github.com/consensys/gnark-crypto/field/koalabear/vortex"
	"github.com/consensys/linea-monorepo/prover/crypto/poseidon2_koalabear"
	"github.com/consensys/linea-monorepo/prover/crypto/state-management/smt_koalabear"
	"github.com/consensys/linea-monorepo/prover/crypto/vortex"
	"github.com/consensys/linea-monorepo/prover/maths/common/smartvectors"
	"github.com/consensys/linea-monorepo/prover/maths/field"
	"github.com/consensys/linea-monorepo/prover/utils"
	"github.com/consensys/linea-monorepo/prover/utils/parallel"
)

// When SIS: hash_columns = poseidon2_koalabear(SIS(col))
// When no SIS: hash_columns = poseidon2_koalabear(encode_bls12377(col))

// EncodedMatrix represents the witness of a Vortex matrix commitment, it is
// represented as an array of rows.
type EncodedMatrix = []smartvectors.SmartVector

// Commitment represents the root of a Merkle tree
type Commitment = field.Octuplet

// Params Embeds vortex.Params to define local methods
type Params struct {
	vortex.Params
}

func NewParams(rate, nbColumns, maxNbRows, logTwoDegree, logTwoBound int) Params {

	_params := vortex.NewParams(
		rate,
		nbColumns,
		maxNbRows,
		logTwoDegree,
		logTwoBound)

	return Params{*_params}
}

// CommitMerkleWithSIS
//
// let h: koala* -> field.Octuplet
// 				a,b,.. -> poseidon2_koalabear(SIS(a,b,...)

// commits to ps by hashing the columns like this:
// [v11 v12 .... v1n ]
// ..
// [vm1 vm2 .... vmn ]
//
//	|             |
//	v             v
//
// [h() .. ....   h()] := v
// Compute MT of v
func (p *Params) CommitMerkleWithSIS(polysMatrix []smartvectors.SmartVector) (EncodedMatrix, Commitment, *smt_koalabear.Tree, []field.Element) {

	if len(polysMatrix) > p.MaxNbRows {
		utils.
			Panic("too many rows: %v, capacity is %v\n", len(polysMatrix), p.MaxNbRows)
	}

	encodedMatrix := p.EncodeRows(polysMatrix)

	var commitment Commitment

	leaf, colHashes := p.sisTransversalHash(encodedMatrix)

	tree := smt_koalabear.NewTree(
		leaf,
	)

	commitment = tree.Root

	return encodedMatrix, commitment, tree, colHashes
}

// CommitMerkleWithSISStreaming computes the same commitment as CommitMerkleWithSIS
// but processes rows in batches to improve cache locality during Ring-SIS hashing.
// The encoded matrix is still fully materialized and returned for use by the
// linear combination and opening phases.
//
// batchSize controls how many rows are RS-encoded and SIS-hashed at a time.
// If batchSize <= 0, defaults to max(1, NbRows/8).
func (p *Params) CommitMerkleWithSISStreaming(
	polysMatrix []smartvectors.SmartVector,
	batchSize int,
) (EncodedMatrix, Commitment, *smt_koalabear.Tree, []field.Element) {

	if len(polysMatrix) > p.MaxNbRows {
		utils.Panic("too many rows: %v, capacity is %v\n", len(polysMatrix), p.MaxNbRows)
	}

	nbRows := len(polysMatrix)
	if batchSize <= 0 {
		batchSize = nbRows / 8
		if batchSize < 1 {
			batchSize = 1
		}
	}

	// Sanity-check row lengths.
	for i := range polysMatrix {
		if polysMatrix[i].Len() != p.NbColumns {
			utils.Panic("Bad length : expected %v columns but col %v has size %v", p.NbColumns, i, polysMatrix[i].Len())
		}
	}

	nbEncodedCols := p.RsParams.NbEncodedColumns()
	encodedMatrix := make(EncodedMatrix, nbRows)
	hasher := p.Key.NewIncrementalHasher(nbEncodedCols)

	for start := 0; start < nbRows; start += batchSize {
		end := start + batchSize
		if end > nbRows {
			end = nbRows
		}

		// RS-encode this batch of rows.
		batch := polysMatrix[start:end]
		parallel.Execute(len(batch), func(s, e int) {
			for i := s; i < e; i++ {
				encodedMatrix[start+i] = p.RsParams.RsEncodeBase(batch[i])
			}
		})

		// Feed encoded rows to the incremental hasher.
		hasher.AbsorbBatch(encodedMatrix[start:end])
	}

	sisHashes := hasher.Finalize()
	leaves := p.sisHashesToMerkleLeaves(sisHashes)

	tree := smt_koalabear.NewTree(leaves)

	return encodedMatrix, tree.Root, tree, sisHashes
}

func (p *Params) sisTransversalHash(v []smartvectors.SmartVector) ([]field.Octuplet, []field.Element) {

	// sisHashes = [ [a, b ...], ... ] where [a, b, ...] is the sis hash of a column, a, b etc are on koalabear
	// let h = poseidon2_koalabear
	// compute [ h([a, b ...]), .. ]
	sisHashes := make([]field.Element, p.RsParams.NbEncodedColumns()*p.Key.OutputSize())
	sisHashes = p.Key.TransversalHash(v, sisHashes)
	leaves := p.sisHashesToMerkleLeaves(sisHashes)
	return leaves, sisHashes
}

// sisHashesToMerkleLeaves compresses SIS column hashes via Poseidon2 into
// Merkle tree leaves. sisHashes has length NbEncodedColumns * OutputSize.
func (p *Params) sisHashesToMerkleLeaves(sisHashes []field.Element) []field.Octuplet {
	chunkSize := p.Key.OutputSize()
	numCols := p.RsParams.NbEncodedColumns()
	leaves := make([]field.Octuplet, numCols)

	r := numCols % 16
	n := numCols / 16

	if chunkSize%8 != 0 {
		// TODO @gbotrel make the fast path generic with different SIS params
		parallel.Execute(numCols, func(start, stop int) {
			hasher := poseidon2_koalabear.NewMDHasher()
			for chunkID := start; chunkID < stop; chunkID++ {
				startChunk := chunkID * chunkSize
				hasher.Reset()
				hasher.WriteElements(sisHashes[startChunk : startChunk+chunkSize]...)
				leaves[chunkID] = hasher.SumElement()
			}
		})
		return leaves
	}

	// process the n full chunks of 16 columns using optimized SIMD implementation
	// if available
	parallel.Execute(n, func(start, stop int) {
		for chunkID := start; chunkID < stop; chunkID++ {
			startChunk := chunkID * 16 * chunkSize
			vgnark.CompressPoseidon2x16(sisHashes[startChunk:startChunk+16*chunkSize], chunkSize, leaves[chunkID*16:(chunkID+1)*16])
		}
	})

	// process the remaining r columns
	hasher := poseidon2_koalabear.NewMDHasher()
	for i := n * 16; i < n*16+r; i++ {
		startChunk := i * chunkSize
		hasher.Reset()
		hasher.WriteElements(sisHashes[startChunk : startChunk+chunkSize]...)
		leaves[i] = hasher.SumElement()
	}

	return leaves
}

// CommitMerkleWithoutSIS
//
// let h: koala* -> field.Octuplet
// 				a,b,.. -> poseidon2_koalabear(a,b,...)

// commits to ps by hashing the columns like this:
// [v11 v12 .... v1n ]
// ..
// [vm1 vm2 .... vmn ]
//
//	|             |
//	v             v
//
// [h() .. ....   h()] := v
// Compute MT of v
func (p *Params) CommitMerkleWithoutSIS(polysMatrix []smartvectors.SmartVector) (EncodedMatrix, Commitment, *smt_koalabear.Tree, []field.Element) {

	if len(polysMatrix) > p.MaxNbRows {
		utils.Panic("too many rows: %v, capacity is %v\n", len(polysMatrix), p.MaxNbRows)
	}

	encodedMatrix := p.EncodeRows(polysMatrix)

	var commitment Commitment

	colHashesOcts := p.noSisTransversalHash(encodedMatrix)
	colHashes := make([]field.Element, 0, len(colHashesOcts)*len(field.Octuplet{}))
	for i := range colHashesOcts {
		colHashes = append(colHashes, colHashesOcts[i][:]...)
	}

	tree := smt_koalabear.NewTree(
		colHashesOcts,
	)

	commitment = tree.Root
	return encodedMatrix, commitment, tree, colHashes
}

// Uses the no-sis hash function to hash the columns. It uses the leafHasher
// function to hash the columns.
func (p *Params) noSisTransversalHash(v []smartvectors.SmartVector) []field.Octuplet {

	// Assert that all smart-vectors have the same numCols
	nbCols := v[0].Len()
	for i := range v {
		if v[i].Len() != nbCols {
			utils.Panic("Unexpected : all inputs smart-vectors should have the same length the first one has length %v, but #%v has length %v",
				nbCols, i, v[i].Len())
		}
	}

	nbRows := len(v)
	res := make([]field.Octuplet, nbCols)
	parallel.Execute(nbCols, func(start, end int) {
		curCol := make([]field.Element, nbRows)
		h := poseidon2_koalabear.NewMDHasher()
		for i := start; i < end; i++ {
			for j := 0; j < nbRows; j++ {
				curCol[j] = v[j].Get(i)
			}
			h.WriteElements(curCol...)
			res[i] = h.SumElement()
			h.Reset()
		}
	})

	return res
}

// EncodeRows returns the encodes `ps` using Reed-Solomon. ps is interpreted as
// a list of rows of the Vortex witness and encodedMatrix is obtained by
// encoding each of the [smartvectors.SmartVector] it contains separately.
func (params *Params) EncodeRows(ps []smartvectors.SmartVector) (encodedMatrix EncodedMatrix) {

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
			encodedMatrix[i] = params.RsParams.RsEncodeBase(ps[i])
		}
	})

	return encodedMatrix
}
