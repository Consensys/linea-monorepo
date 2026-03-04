package vortex_bn254

import (
	bn254fr "github.com/consensys/gnark-crypto/ecc/bn254/fr"
	"github.com/consensys/linea-monorepo/prover/crypto/poseidon2_bn254"
	"github.com/consensys/linea-monorepo/prover/crypto/state-management/smt_bn254"
	"github.com/consensys/linea-monorepo/prover/crypto/vortex"
	"github.com/consensys/linea-monorepo/prover/maths/common/smartvectors"
	"github.com/consensys/linea-monorepo/prover/maths/field"
	"github.com/consensys/linea-monorepo/prover/utils"
	"github.com/consensys/linea-monorepo/prover/utils/parallel"
)

// EncodedMatrix represents the witness of a Vortex matrix commitment
type EncodedMatrix = []smartvectors.SmartVector

// Commitment represents the root of a BN254 Merkle tree
type Commitment = bn254fr.Element

// Params embeds vortex.Params to define local methods
type Params struct {
	vortex.Params
}

func NewParams(rate, nbColumns, maxNbRows, logTwoDegree, logTwoBound int) Params {
	_params := vortex.NewParams(rate, nbColumns, maxNbRows, logTwoDegree, logTwoBound)
	return Params{*_params}
}

// CommitMerkleWithoutSIS commits to polysMatrix by hashing columns using BN254 Poseidon2
// and building a BN254 Merkle tree.
func (p *Params) CommitMerkleWithoutSIS(polysMatrix []smartvectors.SmartVector) (EncodedMatrix, Commitment, *smt_bn254.Tree, []bn254fr.Element) {
	if len(polysMatrix) > p.MaxNbRows {
		utils.Panic("too many rows: %v, capacity is %v\n", len(polysMatrix), p.MaxNbRows)
	}

	encodedMatrix := p.EncodeRows(polysMatrix)

	var commitment Commitment
	colHashes := p.noSisTransversalHash(encodedMatrix)

	tree := smt_bn254.BuildComplete(colHashes)

	commitment = tree.Root
	return encodedMatrix, commitment, tree, colHashes
}

// noSisTransversalHash hashes each column (transversally) using BN254 Poseidon2.
// Each column of KoalaBear elements is hashed to a single BN254 field element.
func (p *Params) noSisTransversalHash(v []smartvectors.SmartVector) []bn254fr.Element {
	nbCols := v[0].Len()
	for i := range v {
		if v[i].Len() != nbCols {
			utils.Panic("Unexpected: all inputs smart-vectors should have the same length the first one has length %v, but #%v has length %v",
				nbCols, i, v[i].Len())
		}
	}

	nbRows := len(v)
	res := make([]bn254fr.Element, nbCols)
	parallel.Execute(nbCols, func(start, end int) {
		curCol := make([]field.Element, nbRows)
		h := poseidon2_bn254.NewMDHasher()
		for i := start; i < end; i++ {
			for j := 0; j < nbRows; j++ {
				curCol[j] = v[j].Get(i)
			}
			h.WriteKoalabearElements(curCol...)
			res[i] = h.SumElement()
			h.Reset()
		}
	})

	return res
}

// EncodeRows encodes polysMatrix using Reed-Solomon.
func (params *Params) EncodeRows(ps []smartvectors.SmartVector) (encodedMatrix EncodedMatrix) {
	for i := range ps {
		if ps[i].Len() != params.NbColumns {
			utils.Panic("Bad length: expected %v columns but col %v has size %v", params.NbColumns, i, ps[i].Len())
		}
	}

	encodedMatrix = make(EncodedMatrix, len(ps))
	parallel.Execute(len(ps), func(start, stop int) {
		for i := start; i < stop; i++ {
			encodedMatrix[i] = params.RsParams.RsEncodeBase(ps[i])
		}
	})

	return encodedMatrix
}
