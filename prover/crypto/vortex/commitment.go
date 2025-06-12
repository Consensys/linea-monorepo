package vortex

import (
	"github.com/consensys/gnark-crypto/field/koalabear/vortex"
	"github.com/consensys/linea-monorepo/prover/crypto/state-management/hashtypes"
	"github.com/consensys/linea-monorepo/prover/crypto/state-management/smt"
	"github.com/consensys/linea-monorepo/prover/maths/common/mempool"
	"github.com/consensys/linea-monorepo/prover/maths/common/smartvectors"
	"github.com/consensys/linea-monorepo/prover/maths/field"
	"github.com/consensys/linea-monorepo/prover/utils"
	"github.com/consensys/linea-monorepo/prover/utils/parallel"
	"github.com/consensys/linea-monorepo/prover/utils/types"
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

	colHashes = p.transversalHash(encodedMatrix)

	// at this stage colHashes is a list of field.Element

	leaves := make([]types.Bytes32, nbColumns)
	sizeChunk := len(colHashes) / nbColumns
	for i := 0; i < nbColumns; i++ {
		leaves[i] = p.computeLeaf(colHashes[i*sizeChunk : (i+1)*sizeChunk])
	}

	tree = smt.BuildComplete(
		leaves,
		func() hashtypes.Hasher {
			return hashtypes.Hasher{Hash: p.MerkleHasher()}
		},
	)

	return encodedMatrix, tree, colHashes
}

func (p *Params) computeLeaf(leaf []field.Element) types.Bytes32 {

	var res types.Bytes32
	if p.MerkleHasher == nil { // by default, we use poseidon2
		h := vortex.HashPoseidon2(leaf)
		res = types.HashToBytes32(h)

	} else {
		merkleHasher := p.MerkleHasher()

		merkleHasher.Reset()
		for j := 0; j < len(leaf); j++ {
			merkleHasher.Write(leaf[j].Marshal())
		}
		h := merkleHasher.Sum(nil)
		copy(res[:], h)
	}
	return res
}

func (p *Params) hashColumn(column []field.Element) []field.Element {
	var res []field.Element
	if p.ColumnHasher != nil {
		hasher := p.ColumnHasher()
		res = make([]field.Element, 8)
		hasher.Reset()
		for k := 0; k < len(column); k++ {
			hasher.Write(column[k].Marshal())
		}
		h := hasher.Sum(nil)
		for j := 0; j < 8; j++ {
			res[j].SetBytes(h[j*8 : (j+1)*8])
		}
	} else {
		res = p.Key.Hash(column)
	}
	return res
}

func (p *Params) transversalHash(v []smartvectors.SmartVector) []field.Element {

	var res []field.Element
	nbColumns := v[0].Len()
	nbRows := len(v)
	var sizeChunk int
	if p.ColumnHasher == nil {
		sizeChunk = p.Key.OutputSize()
	} else {
		sizeChunk = 8
	}
	res = make([]field.Element, sizeChunk*nbColumns)
	parallel.Execute(nbColumns, func(start, end int) {
		ithCol := make([]field.Element, nbRows)
		for i := start; i < end; i++ {
			for j := 0; j < nbRows; j++ {
				ithCol[j] = v[j].Get(i)
			}
			curHash := p.hashColumn(ithCol)
			copy(res[i*sizeChunk:(i+1)*sizeChunk], curHash)
		}
	})

	return res
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
