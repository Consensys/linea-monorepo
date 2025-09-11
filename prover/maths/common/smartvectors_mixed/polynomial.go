package smartvectors_mixed

import (
	"sync/atomic"

	sv "github.com/consensys/linea-monorepo/prover/maths/common/smartvectors"

	"github.com/consensys/gnark-crypto/field/koalabear/vortex"
	"github.com/consensys/linea-monorepo/prover/maths/common/fastpoly"
	"github.com/consensys/linea-monorepo/prover/maths/field"
	"github.com/consensys/linea-monorepo/prover/maths/field/fext"
	"github.com/consensys/linea-monorepo/prover/utils/parallel"
)

// BatchEvaluateLagrange - memory optimized version
func BatchEvaluateLagrange(vs []sv.SmartVector, x fext.Element, oncoset ...bool) []fext.Element {
	results := make([]fext.Element, len(vs))

	if len(vs) == 0 {
		return results
	}

	var (
		basePolys     [][]field.Element
		baseIndices   []int
		extPolys      [][]fext.Element
		extIndices    []int
		totalConstant uint64
	)

	// Process in parallel - first pass for constants and type classification
	processed := make([]bool, len(vs))

	parallel.Execute(len(vs), func(start, stop int) {
		for i := start; i < stop; i++ {
			isBaseField := IsBase(vs[i])

			if isBaseField {
				if con, ok := vs[i].(*sv.Constant); ok {
					fext.SetFromBase(&results[i], &con.Value)
					processed[i] = true
					atomic.AddUint64(&totalConstant, 1)
				}
			} else {
				if con, ok := vs[i].(*sv.ConstantExt); ok {
					results[i] = con.Value
					processed[i] = true
					atomic.AddUint64(&totalConstant, 1)
				}
			}
		}
	})

	// Early return if all constants
	if int(totalConstant) == len(vs) {
		return results
	}

	// Second pass - extract polynomials (sequential is fine here)
	for i, v := range vs {
		if processed[i] {
			continue // Skip constants
		}

		if IsBase(v) {
			poly, _ := v.IntoRegVecSaveAllocBase()
			basePolys = append(basePolys, poly)
			baseIndices = append(baseIndices, i)
		} else {
			poly := v.IntoRegVecSaveAllocExt()
			extPolys = append(extPolys, poly)
			extIndices = append(extIndices, i)
		}
	}

	// Batch evaluate polynomials
	if len(basePolys) > 0 {
		baseResults := fastpoly.BatchEvaluateLagrangeMixed(basePolys, x, oncoset...)
		for i, result := range baseResults {
			results[baseIndices[i]] = result
		}
	}

	if len(extPolys) > 0 {
		extResults, _ := vortex.BatchEvalFextPolyLagrange(extPolys, x, oncoset...)
		for i, result := range extResults {
			results[extIndices[i]] = result
		}
	}

	return results
}
