package smartvectors_mixed

import (
	"math/big"

	sv "github.com/consensys/linea-monorepo/prover/maths/common/smartvectors"

	"github.com/consensys/gnark-crypto/field/koalabear"
	"github.com/consensys/gnark-crypto/field/koalabear/extensions"
	"github.com/consensys/gnark-crypto/field/koalabear/fft"
	"github.com/consensys/linea-monorepo/prover/maths/field/fext"
	"github.com/consensys/linea-monorepo/prover/utils/parallel"
)

// BatchEvaluateLagrange - Extended version supporting both base and extension fields
func BatchEvaluateLagrange(vs []sv.SmartVector, x fext.Element) []fext.Element {
	results := make([]fext.Element, len(vs))
	if len(vs) == 0 {
		return results
	}

	n := vs[0].Len()
	vLagrangeBasis := computeLagrangeBasisAtX(n, x)

	// Parallel processing - classification and polynomial extraction
	parallel.Execute(len(vs), func(start, stop int) {
		var scratchBase koalabear.Vector
		var scratchExt extensions.Vector

		for i := start; i < stop; i++ {
			// Check if it's base field or extension field
			if IsBase(vs[i]) {
				switch v := vs[i].(type) {
				case *sv.Constant:
					results[i] = v.GetExt(0)
					continue
				case *sv.Regular:
					// no need to copy here.
					results[i] = vLagrangeBasis.InnerProductByElement(koalabear.Vector(*v))
				default:
					if scratchBase == nil {
						scratchBase = make(koalabear.Vector, n)
					}
					v.WriteInSlice(scratchBase)
					results[i] = vLagrangeBasis.InnerProductByElement(koalabear.Vector(scratchBase))
				}

			} else {
				switch v := vs[i].(type) {
				case *sv.ConstantExt:
					results[i] = v.Value
					continue
				case *sv.RegularExt:
					// no need to copy here.
					results[i] = vLagrangeBasis.InnerProduct(extensions.Vector(*v))
				default:
					if scratchExt == nil {
						scratchExt = make(extensions.Vector, n)
					}
					v.WriteInSliceExt(scratchExt)
					results[i] = vLagrangeBasis.InnerProduct(scratchExt)
				}

			}
		}
	})

	return results
}

// TODO taken from gnark-crypto, tuned and adapted for our use case
// computeLagrangeBasisAtX computes (Lᵢ(x))_{i<n} and numerator for Lagrange basis evaluation
func computeLagrangeBasisAtX(n int, x fext.Element) extensions.Vector {

	generator, _ := fft.Generator(uint64(n))
	generatorInv := new(koalabear.Element).Inverse(&generator)
	one := koalabear.One()

	// (xⁿ - 1) / n
	var numerator fext.Element
	numerator.Exp(x, big.NewInt(int64(n)))
	numerator.B0.A0.Sub(&numerator.B0.A0, &one)

	cardInv := koalabear.NewElement(uint64(n))
	cardInv.Inverse(&cardInv)
	numerator.MulByElement(&numerator, &cardInv)
	numerator.Inverse(&numerator)

	// compute x-1, x/ω-1, x/ω²-1, ...
	res := make(extensions.Vector, n)
	res[0] = x
	for i := 1; i < n; i++ {
		res[i].MulByElement(&res[i-1], generatorInv)
	}
	isRootOfUnity := -1
	for i := range res {
		res[i].B0.A0.Sub(&res[i].B0.A0, &one)
		if res[i].IsZero() { // it means that x is a root of unity
			isRootOfUnity = i
			break
		}
	}
	if isRootOfUnity != -1 {
		res = make(extensions.Vector, n)
		res[isRootOfUnity].SetOne()
		return res
	}
	res.ScalarMul(res, &numerator)

	// 1/(x-1), 1/(x/ω-1), 1/(x/ω²-1), ...
	res = fext.ParBatchInvert(res, 0)

	return res
}
