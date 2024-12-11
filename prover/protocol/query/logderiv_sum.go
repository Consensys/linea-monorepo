package query

import (
	"github.com/consensys/gnark/frontend"
	"github.com/consensys/linea-monorepo/prover/maths/field"
	"github.com/consensys/linea-monorepo/prover/protocol/ifaces"
	sym "github.com/consensys/linea-monorepo/prover/symbolic"
)

// LogDerivativeSumInput stores the input to the query
type LogDerivativeSumInput struct {
	Round, Size int
	Numerator   []*sym.Expression // T -> -M, S -> +Filter
	Denominator []*sym.Expression // S or T -> ({S,T} + X)
}

// LogDerivativeSum is the context of LogDerivativeSum.
// The fields are maps from [round, size].
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

// NewLogDerivativeSum creates the new context LogDerivativeSum.
func NewLogDerivativeSum(inp map[[2]int]*LogDerivativeSumInput) LogDerivativeSum {

	// add some sanity checks here

	return LogDerivativeSum{
		Inputs: inp,
	}

}

// Name implements the [ifaces.Query] interface
func (r LogDerivativeSum) Name() ifaces.QueryID {
	return r.ID
}

// Constructor for the query parameters/result
func NewLogDeriveSumParams(sum field.Element) LogDerivSumParams {
	return LogDerivSumParams{Sum: sum}
}

// Test that global sum is correct
func (r LogDerivativeSum) Check(run ifaces.Runtime) error {

	return nil
}

// Test that global sum is correct
func (r LogDerivativeSum) CheckGnark(api frontend.API, run ifaces.GnarkRuntime) {

}
