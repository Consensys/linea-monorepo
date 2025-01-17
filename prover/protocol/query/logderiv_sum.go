package query

import (
	"fmt"

	"github.com/consensys/gnark/frontend"
	"github.com/consensys/linea-monorepo/prover/crypto/fiatshamir"
	"github.com/consensys/linea-monorepo/prover/maths/common/vector"
	"github.com/consensys/linea-monorepo/prover/maths/field"
	"github.com/consensys/linea-monorepo/prover/protocol/column"
	"github.com/consensys/linea-monorepo/prover/protocol/ifaces"
	sym "github.com/consensys/linea-monorepo/prover/symbolic"
	"github.com/consensys/linea-monorepo/prover/utils"
)

// LogDerivativeSumInput stores the input to the query
type LogDerivativeSumInput struct {
	Size        int
	Numerator   []*sym.Expression // T -> -M, S -> +Filter
	Denominator []*sym.Expression // S or T -> ({S,T} + X)
}

// LogDerivativeSum is the context of LogDerivativeSum query.
// The fields are maps from [round, size].
// the aim of the query is to compute:
// \sum_{i,j} N_{i,j}/D_{i,j} where
// N_{i,j} is  the i-th element of the underlying column of  j-th Numerator
// D_{i,j} is  the i-th element of the underlying column of  j-th Denominator
type LogDerivativeSum struct {
	Round  int
	Inputs map[int]*LogDerivativeSumInput
	ID     ifaces.QueryID
}

// the result of the global Sum
type LogDerivSumParams struct {
	Sum field.Element // the sum of all the ZOpenings from different [round,size].
}

// Updates a Fiat-Shamir state
func (l LogDerivSumParams) UpdateFS(fs *fiatshamir.State) {
	fs.Update(l.Sum)
}

// NewLogDerivativeSum creates the new context LogDerivativeSum.
func NewLogDerivativeSum(round int, inp map[int]*LogDerivativeSumInput, id ifaces.QueryID) LogDerivativeSum {

	// check the length consistency
	for key := range inp {
		if len(inp[key].Numerator) != len(inp[key].Denominator) || len(inp[key].Numerator) == 0 {
			utils.Panic("Numerator and Denominator should have the same (no-zero) length, %v , %v", len(inp[key].Numerator), len(inp[key].Denominator))
		}
		for i := range inp[key].Numerator {

			if err := inp[key].Numerator[i].Validate(); err != nil {
				utils.Panic(" Numerator[%v] is not a valid expression", i)
			}

			if err := inp[key].Denominator[i].Validate(); err != nil {
				utils.Panic(" Denominator[%v] is not a valid expression", i)
			}
		}
	}

	return LogDerivativeSum{
		Round:  round,
		Inputs: inp,
		ID:     id,
	}

}

// Name implements the [ifaces.Query] interface
func (r LogDerivativeSum) Name() ifaces.QueryID {
	return r.ID
}

// Constructor for the query parameters/result
func NewLogDerivSumParams(sum field.Element) LogDerivSumParams {
	return LogDerivSumParams{Sum: sum}
}

// Test that global sum is correct
func (r LogDerivativeSum) Check(run ifaces.Runtime) error {
	params := run.GetParams(r.ID).(LogDerivSumParams)
	// compute the actual sum from the Numerator and Denominator
	actualSum := field.Zero()
	for key := range r.Inputs {
		for i, num := range r.Inputs[key].Numerator {

			var (
				numBoard          = num.Board()
				denBoard          = r.Inputs[key].Denominator[i].Board()
				numeratorMetadata = numBoard.ListVariableMetadata()
				denominator       = column.EvalExprColumn(run, denBoard).IntoRegVecSaveAlloc()
				numerator         []field.Element
				packedZ           = field.BatchInvert(denominator)
			)

			if len(numeratorMetadata) == 0 {
				numerator = vector.Repeat(field.One(), r.Inputs[key].Size)
			}

			if len(numeratorMetadata) > 0 {
				numerator = column.EvalExprColumn(run, numBoard).IntoRegVecSaveAlloc()
			}

			for k := range packedZ {
				packedZ[k].Mul(&numerator[k], &packedZ[k])
				if k > 0 {
					packedZ[k].Add(&packedZ[k], &packedZ[k-1])
				}
			}
			actualSum.Add(&actualSum, &packedZ[len(packedZ)-1])
		}
	}

	if actualSum != params.Sum {
		return fmt.Errorf("expected LogDerivativeSum = %s but got %s for the query %v", params.Sum.String(), actualSum.String(), r.ID)
	}

	return nil
}

// Test that global sum is correct
func (r LogDerivativeSum) CheckGnark(api frontend.API, run ifaces.GnarkRuntime) {
	/*params := run.GetParams(r.ID).(GnarkLogDerivSumParams)
	actualY := TBD
	api.AssertIsEqual(params.Y, actualY)
	*/
}
