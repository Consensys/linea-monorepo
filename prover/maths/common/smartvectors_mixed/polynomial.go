package smartvectors_mixed

import (
	sv "github.com/consensys/linea-monorepo/prover/maths/common/smartvectors"

	"github.com/consensys/gnark-crypto/field/koalabear"
	"github.com/consensys/gnark-crypto/field/koalabear/extensions"
	"github.com/consensys/gnark-crypto/field/koalabear/vortex"
	"github.com/consensys/linea-monorepo/prover/maths/field/fext"
	"github.com/consensys/linea-monorepo/prover/utils/parallel"
)

// BatchEvaluateLagrange - Extended version supporting both base and extension fields
func BatchEvaluateLagrange(vs []sv.SmartVector, x fext.Element, oncoset ...bool) []fext.Element {
	results := make([]fext.Element, len(vs))
	if len(vs) == 0 {
		return results
	}

	n := vs[0].Len()
	lagrangeBasis, err := vortex.ComputeLagrangeBasisAtX(n, x, oncoset...)
	if err != nil {
		panic(err)
	}
	vLagrangeBasis := extensions.Vector(lagrangeBasis)

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
