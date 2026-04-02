package vortex

import (
	"github.com/consensys/linea-monorepo/prover/crypto/reedsolomon"
	"github.com/consensys/linea-monorepo/prover/maths/common/smartvectors"
	"github.com/consensys/linea-monorepo/prover/maths/common/vectorext"
	"github.com/consensys/linea-monorepo/prover/maths/field"
	"github.com/consensys/linea-monorepo/prover/maths/field/fext"
	"github.com/consensys/linea-monorepo/prover/utils"
	"github.com/consensys/linea-monorepo/prover/utils/parallel"
)

// OpeningProof represents an opening proof for a Vortex commitment
type OpeningProof struct {

	// Columns [i][j][k] returns the k-th entry
	// of the j-th selected column of the i-th commitment
	Columns [][][]field.Element

	// Linear combination of the Reed-Solomon encoded polynomials to open.
	LinearCombination smartvectors.SmartVector
}

// Let x := randomCoin
// computes ∑ᵢxⁱ * v[i]
// n is the size of each vector v[i]
//
// TODO @thomaspiellard why not use directly smarvectorext.LinComb ??
func LinearCombination(proof *OpeningProof, v []smartvectors.SmartVector, randomCoin fext.Element) {

	if len(v) == 0 {
		utils.Panic("attempted to open an empty witness")
	}

	n := v[0].Len()
	linComb := make([]fext.Element, n)
	parallel.Execute(len(linComb), func(start, stop int) {

		x := fext.One()
		scratch := make(vectorext.Vector, stop-start)
		localLinComb := make(vectorext.Vector, stop-start)
		for i := range v {
			_sv := v[i]
			// we distinguish the case of a regular vector and constant to avoid
			// unnecessary allocations and copies
			switch _svt := _sv.(type) {
			case *smartvectors.Constant:
				cst := _svt.GetExt(0)
				cst.Mul(&cst, &x)
				for j := range localLinComb {
					localLinComb[j].Add(&localLinComb[j], &cst)
				}
				x.Mul(&x, &randomCoin)
				continue
			case *smartvectors.Regular:
				sv := field.Vector((*_svt)[start:stop])
				for i := range scratch {
					fext.SetFromBase(&scratch[i], &sv[i])
				}
			default:
				sv := _svt.SubVector(start, stop)
				sv.WriteInSliceExt(scratch)
			}
			scratch.ScalarMul(scratch, &x)
			localLinComb.Add(localLinComb, scratch)
			x.Mul(&x, &randomCoin)

		}
		copy(linComb[start:stop], localLinComb)
	})

	proof.LinearCombination = smartvectors.NewRegularExt(linComb)
}

// LinearCombinationStreaming computes ∑ᵢxⁱ * v[i] where v is formed by
// concatenating encodedRows (already RS-encoded) followed by originalRows
// (re-encoded on the fly using rsParams). This avoids materializing the full
// encoded matrix for the original rows.
//
// The ordering matters: encodedRows come first in the power series, then
// originalRows. This matches the NoSIS-before-SIS stacking convention.
func LinearCombinationStreaming(
	proof *OpeningProof,
	encodedRows []smartvectors.SmartVector,
	originalRows []smartvectors.SmartVector,
	rsParams *reedsolomon.RsParams,
	randomCoin fext.Element,
) {
	totalRows := len(encodedRows) + len(originalRows)
	if totalRows == 0 {
		utils.Panic("attempted to open an empty witness")
	}

	// Determine output length.
	var n int
	if len(encodedRows) > 0 {
		n = encodedRows[0].Len()
	} else {
		n = rsParams.NbEncodedColumns()
	}

	linComb := make([]fext.Element, n)

	// Process already-encoded rows using the efficient column-parallel approach.
	if len(encodedRows) > 0 {
		parallel.Execute(n, func(start, stop int) {
			x := fext.One()
			scratch := make(vectorext.Vector, stop-start)
			localLinComb := make(vectorext.Vector, stop-start)
			for i := range encodedRows {
				sv := encodedRows[i]
				switch svt := sv.(type) {
				case *smartvectors.Constant:
					cst := svt.GetExt(0)
					cst.Mul(&cst, &x)
					for j := range localLinComb {
						localLinComb[j].Add(&localLinComb[j], &cst)
					}
					x.Mul(&x, &randomCoin)
					continue
				case *smartvectors.Regular:
					svSlice := field.Vector((*svt)[start:stop])
					for j := range scratch {
						fext.SetFromBase(&scratch[j], &svSlice[j])
					}
				default:
					sub := svt.SubVector(start, stop)
					sub.WriteInSliceExt(scratch)
				}
				scratch.ScalarMul(scratch, &x)
				localLinComb.Add(localLinComb, scratch)
				x.Mul(&x, &randomCoin)
			}
			copy(linComb[start:stop], localLinComb)
		})
	}

	// Process original rows by re-encoding each one on the fly.
	// We use a row-streaming approach: for each row, re-encode it, then
	// accumulate x^i * encoded[col] into linComb across all columns.
	x := fext.One()
	// Advance x past the encoded rows.
	for range encodedRows {
		x.Mul(&x, &randomCoin)
	}

	for _, row := range originalRows {
		encoded := rsParams.RsEncodeBase(row)

		// Handle constant vectors efficiently.
		if cst, ok := encoded.(*smartvectors.Constant); ok {
			val := cst.GetExt(0)
			val.Mul(&val, &x)
			for col := range linComb {
				linComb[col].Add(&linComb[col], &val)
			}
			x.Mul(&x, &randomCoin)
			continue
		}

		// For regular/other vectors, parallelize over columns.
		xCopy := x // capture for closure
		parallel.Execute(n, func(start, stop int) {
			for col := start; col < stop; col++ {
				val := encoded.Get(col)
				var valExt fext.Element
				fext.SetFromBase(&valExt, &val)
				valExt.Mul(&valExt, &xCopy)
				linComb[col].Add(&linComb[col], &valExt)
			}
		})
		x.Mul(&x, &randomCoin)
	}

	proof.LinearCombination = smartvectors.NewRegularExt(linComb)
}
