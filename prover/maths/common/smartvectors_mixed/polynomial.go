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
		scratchBase := make(koalabear.Vector, n)
		scratchExt := make(extensions.Vector, n)

		for i := start; i < stop; i++ {
			// Check if it's base field or extension field
			if IsBase(vs[i]) {
				vsi := scratchBase
				switch v := vs[i].(type) {
				case *sv.Constant:
					results[i] = v.GetExt(0)
					continue
				case *sv.Regular:
					// no need to copy here.
					vsi = koalabear.Vector(*v)
				default:
					v.WriteInSlice(vsi)
				}
				results[i] = vLagrangeBasis.InnerProductByElement(vsi)
			} else {
				vsi := scratchExt
				switch v := vs[i].(type) {
				case *sv.ConstantExt:
					results[i] = v.Value
					continue
				case *sv.RegularExt:
					// no need to copy here.
					vsi = extensions.Vector(*v)
				default:
					v.WriteInSliceExt(vsi)
				}
				results[i] = vLagrangeBasis.InnerProduct(vsi)
			}
		}
	})

	return results
}
