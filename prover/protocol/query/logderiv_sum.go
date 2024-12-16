package query

import (
	"fmt"

	"github.com/consensys/gnark/frontend"
	"github.com/consensys/linea-monorepo/prover/crypto/fiatshamir"
	"github.com/consensys/linea-monorepo/prover/maths/common/smartvectors"
	"github.com/consensys/linea-monorepo/prover/maths/common/vector"
	"github.com/consensys/linea-monorepo/prover/maths/field"
	"github.com/consensys/linea-monorepo/prover/protocol/coin"
	"github.com/consensys/linea-monorepo/prover/protocol/ifaces"
	"github.com/consensys/linea-monorepo/prover/protocol/variables"
	"github.com/consensys/linea-monorepo/prover/symbolic"
	sym "github.com/consensys/linea-monorepo/prover/symbolic"
	"github.com/consensys/linea-monorepo/prover/utils"
)

// LogDerivativeSumInput stores the input to the query
type LogDerivativeSumInput struct {
	Round, Size int
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
	Inputs map[[2]int]*LogDerivativeSumInput

	ZNumeratorBoarded, ZDenominatorBoarded map[[2]int][]sym.ExpressionBoard

	Zs map[[2]int][]ifaces.Column
	// ZOpenings are the opening queries to the end of each Z.
	ZOpenings map[[2]int][]LocalOpening

	ID ifaces.QueryID
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
func NewLogDerivativeSum(inp map[[2]int]*LogDerivativeSumInput, id ifaces.QueryID) LogDerivativeSum {

	// check the length consistency
	for key := range inp {
		if len(inp[key].Numerator) != len(inp[key].Denominator) || len(inp[key].Numerator) == 0 {
			panic("Numerator and Denominator should have the same (no-zero) length")
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
				denominator       = EvalExprColumn(run, denBoard).IntoRegVecSaveAlloc()
				numerator         []field.Element
				packedZ           = field.BatchInvert(denominator)
			)

			if len(numeratorMetadata) == 0 {
				numerator = vector.Repeat(field.One(), r.Inputs[key].Size)
			}

			if len(numeratorMetadata) > 0 {
				numerator = EvalExprColumn(run, numBoard).IntoRegVecSaveAlloc()
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

func EvalExprColumn(run ifaces.Runtime, board symbolic.ExpressionBoard) smartvectors.SmartVector {

	var (
		metadata = board.ListVariableMetadata()
		inputs   = make([]smartvectors.SmartVector, len(metadata))
		length   = ExprIsOnSameLengthHandles(&board)
	)

	// Attempt to recover the size of the
	for i := range inputs {
		switch m := metadata[i].(type) {
		case ifaces.Column:
			inputs[i] = m.GetColAssignment(run)
		case coin.Info:
			v := run.GetRandomCoinField(m.Name)
			inputs[i] = smartvectors.NewConstant(v, length)
		case ifaces.Accessor:
			v := m.GetVal(run)
			inputs[i] = smartvectors.NewConstant(v, length)
		case variables.PeriodicSample:
			v := m.EvalCoset(length, 0, 1, false)
			inputs[i] = v
		case variables.X:
			v := m.EvalCoset(length, 0, 1, false)
			inputs[i] = v
		}
	}

	return board.Evaluate(inputs)
}

// ExprIsOnSameLengthHandles checks that all the variables of the expression
// that are [ifaces.Column] have the same size (and panics if it does not), then
// returns the match.
func ExprIsOnSameLengthHandles(board *symbolic.ExpressionBoard) int {

	var (
		metadatas = board.ListVariableMetadata()
		length    = 0
	)

	for _, m := range metadatas {
		switch metadata := m.(type) {
		case ifaces.Column:
			// Initialize the length with the first commitment
			if length == 0 {
				length = metadata.Size()
			}

			// Sanity-check the vector should all have the same length
			if length != metadata.Size() {
				utils.Panic("Inconsistent length for %v (has size %v, but expected %v)", metadata.GetColID(), metadata.Size(), length)
			}
		// The expression can involve random coins
		case coin.Info, variables.X, variables.PeriodicSample, ifaces.Accessor:
			// Do nothing
		default:
			utils.Panic("unknown type %T", metadata)
		}
	}

	// No commitment were found in the metadata, thus this call is broken
	if length == 0 {
		utils.Panic("declared a handle from an expression which does not contains any handle")
	}

	return length
}
