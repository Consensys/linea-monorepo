package smartvectors_mixed

import (
	"sync"

	sv "github.com/consensys/linea-monorepo/prover/maths/common/smartvectors"

	"github.com/consensys/gnark-crypto/field/koalabear/vortex"
	"github.com/consensys/linea-monorepo/prover/maths/field"
	"github.com/consensys/linea-monorepo/prover/maths/field/fext"
	"github.com/consensys/linea-monorepo/prover/utils/parallel"
)

// BatchEvaluateLagrange - Extended version supporting both base and extension fields
func BatchEvaluateLagrange(vs []sv.SmartVector, x fext.Element, oncoset ...bool) []fext.Element {
	results := make([]fext.Element, len(vs))
	if len(vs) == 0 {
		return results
	}

	// Pre-allocate with estimated capacities
	var (
		basePolys     = make([][]field.Element, 0, len(vs))
		baseIndices   = make([]int, 0, len(vs))
		extPolys      = make([][]fext.Element, 0, len(vs))
		extIndices    = make([]int, 0, len(vs))
		totalConstant uint64
	)

	// Work item for parallel processing
	type workItem struct {
		index         int
		isConstant    bool
		isBase        bool
		basePoly      []field.Element
		extPoly       []fext.Element
		constantValue fext.Element
	}

	workItems := make([]workItem, len(vs))

	// Parallel processing - classification and polynomial extraction
	parallel.Execute(len(vs), func(start, stop int) {
		for i := start; i < stop; i++ {
			item := workItem{index: i}

			// Check if it's base field or extension field
			if IsBase(vs[i]) {
				item.isBase = true
				// Check if it's a base field constant
				if con, ok := vs[i].(*sv.Constant); ok {
					fext.SetFromBase(&item.constantValue, &con.Value)
					item.isConstant = true
				} else {
					// Extract base polynomial
					item.basePoly, _ = vs[i].IntoRegVecSaveAllocBase()
				}
			} else {
				item.isBase = false
				// Check if it's an extension field constant
				if con, ok := vs[i].(*sv.ConstantExt); ok {
					item.constantValue = con.Value
					item.isConstant = true
				} else {
					// Extract extension polynomial
					item.extPoly = vs[i].IntoRegVecSaveAllocExt()
				}
			}

			workItems[i] = item
		}

	})

	// Sequential collection and result assignment
	for _, item := range workItems {
		if item.isConstant {
			results[item.index] = item.constantValue
			totalConstant++
		} else if item.isBase {
			basePolys = append(basePolys, item.basePoly)
			baseIndices = append(baseIndices, item.index)
		} else {
			extPolys = append(extPolys, item.extPoly)
			extIndices = append(extIndices, item.index)
		}
	}

	// Early return if all constants
	if int(totalConstant) == len(vs) {
		return results
	}

	// Batch evaluate polynomials in parallel if we have both types
	var baseWg, extWg sync.WaitGroup
	var baseResults []fext.Element
	var extResults []fext.Element
	var baseErr, extErr error

	// Evaluate base field polynomials
	if len(basePolys) > 0 {
		baseWg.Add(1)
		go func() {
			defer baseWg.Done()
			baseResults, baseErr = vortex.BatchEvalBasePolyLagrange(basePolys, x, oncoset...)
			if baseErr != nil {
				panic(baseErr)
			}
		}()
	}

	// Evaluate extension field polynomials
	if len(extPolys) > 0 {
		extWg.Add(1)
		go func() {
			defer extWg.Done()
			extResults, extErr = vortex.BatchEvalFextPolyLagrange(extPolys, x, oncoset...)
			if extErr != nil {
				panic(extErr)
			}
		}()
	}

	// Wait for both evaluations to complete
	baseWg.Wait()
	extWg.Wait()

	// Assign results back to original positions
	for i, result := range baseResults {
		results[baseIndices[i]] = result
	}
	for i, result := range extResults {
		results[extIndices[i]] = result
	}

	return results
}
