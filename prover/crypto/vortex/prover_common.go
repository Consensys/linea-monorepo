package vortex

import (
	"github.com/consensys/gnark-crypto/field/koalabear/fft"
	"github.com/consensys/gnark-crypto/utils"
	"github.com/consensys/linea-monorepo/prover/crypto/reedsolomon"
	"github.com/consensys/linea-monorepo/prover/maths/common/smartvectors"
	"github.com/consensys/linea-monorepo/prover/maths/common/vectorext"
	"github.com/consensys/linea-monorepo/prover/maths/field"
	"github.com/consensys/linea-monorepo/prover/maths/field/fext"
	"github.com/consensys/linea-monorepo/prover/utils/parallel"
)

// OpeningProof represents an opening proof for a Vortex commitment
type OpeningProof struct {

	// Columns [i][j][k] returns the k-th entry
	// of the j-th selected column of the i-th commitment
	Columns [][][]field.Element

	// LinearCombination holds the T monomial coefficients of ∑ᵢ alpha^i * p_i,
	// where p_i are the original (pre-encoding) polynomials.
	LinearCombination []fext.Element
}

// LinearCombination computes ∑ᵢ randomCoin^i * v[i] directly in the T-length
// polynomial domain and converts the result to T monomial coefficients via a
// T-length iFFT (using rsParams.PolyDomain). The input vectors v[i] must be
// the original polynomials in Lagrange basis (length T = rsParams.NbColumns()),
// NOT the RS-encoded codewords.
//
// This avoids all N-length work: the linear combination runs over T instead of
// N entries, and the iFFT is O(T log T) instead of O(N log N).
func LinearCombination(proof *OpeningProof, rsParams *reedsolomon.RsParams, v []smartvectors.SmartVector, randomCoin fext.Element) {

	t := rsParams.NbColumns()
	linComb := make([]fext.Element, t)
	parallel.Execute(t, func(start, stop int) {

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

	// T-length iFFT: Lagrange evaluations over the standard T-point NTT domain
	// → monomial coefficients.
	rsParams.PolyDomain.FFTInverseExt(linComb, fft.DIF, fft.WithNbTasks(1))
	utils.BitReverse(linComb)
	proof.LinearCombination = linComb
}
