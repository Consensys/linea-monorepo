package query

import (
	"errors"
	"fmt"
	"sync"

	"github.com/consensys/gnark/frontend"
	"github.com/consensys/linea-monorepo/prover/crypto/fiatshamir"
	"github.com/consensys/linea-monorepo/prover/maths/common/vector"
	"github.com/consensys/linea-monorepo/prover/maths/field"
	"github.com/consensys/linea-monorepo/prover/protocol/column"
	"github.com/consensys/linea-monorepo/prover/protocol/ifaces"
	sym "github.com/consensys/linea-monorepo/prover/symbolic"
	"github.com/consensys/linea-monorepo/prover/utils"
	"github.com/consensys/linea-monorepo/prover/utils/parallel"
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

		if inp[key].Size != key {
			utils.Panic("the size must match the key: key=%v != size=%v", key, inp[key].Size)
		}

		for i, num := range inp[key].Numerator {
			if err := num.Validate(); err != nil {
				utils.Panic(" Numerator[%v] is not a valid expression", i)
			}

			if rs := column.ColumnsOfExpression(num); len(rs) == 0 {
				continue
			}

			b := num.Board()
			if key != column.ExprIsOnSameLengthHandles(&b) {
				utils.Panic("expression size mismatch")
			}
		}

		for i, den := range inp[key].Denominator {
			if err := inp[key].Denominator[i].Validate(); err != nil {
				utils.Panic(" Denominator[%v] is not a valid expression", i)
			}

			if rs := column.ColumnsOfExpression(den); len(rs) == 0 {
				continue
			}

			b := den.Board()
			if key != column.ExprIsOnSameLengthHandles(&b) {
				utils.Panic("expression size mismatch: qname=%v expression-size=%v expected-size=%v", id, column.ExprIsOnSameLengthHandles(&b), key)
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

// Compute returns the result value of the [LogDerivativeSum] query. It
// should be run by a runtime with access to the query columns. i.e
// either by a [wizard.ProverRuntime] or a [wizard.VerifierRuntime]
// but then the involved columns should all be public.
func (r LogDerivativeSum) Compute(run ifaces.Runtime) (field.Element, error) {

	// compute the actual sum from the Numerator and Denominator
	var (
		err       error
		actualSum = field.Zero()
		resLock   = &sync.Mutex{}
		inputs    = []struct {
			Num, Den *sym.Expression
			Size     int
		}{}
	)

	for key := range r.Inputs {
		for i := range r.Inputs[key].Numerator {
			inputs = append(inputs, struct {
				Num  *sym.Expression
				Den  *sym.Expression
				Size int
			}{
				Num:  r.Inputs[key].Numerator[i],
				Den:  r.Inputs[key].Denominator[i],
				Size: r.Inputs[key].Size,
			})
		}
	}

	parallel.Execute(len(inputs), func(start, stop int) {

		for k := 0; k < len(inputs); k++ {

			var (
				size                = inputs[k].Size
				num                 = inputs[k].Num
				den                 = inputs[k].Den
				numBoard            = num.Board()
				denBoard            = den.Board()
				numeratorMetadata   = numBoard.ListVariableMetadata()
				denominatorMetadata = denBoard.ListVariableMetadata()
				denominator         []field.Element
				numerator           []field.Element
			)

			if len(denominatorMetadata) == 0 {
				denominator = vector.Repeat(field.One(), size)
			} else {
				denominator = column.EvalExprColumn(run, denBoard).IntoRegVecSaveAlloc()
				for l := range denominator {
					if denominator[l].IsZero() {
						resLock.Lock()
						err = errors.Join(
							err,
							fmt.Errorf("denominator contains zeroes [position=%v] [size=%v] [term=%v]", l, size, k),
						)
						resLock.Unlock()
					}
				}
				denominator = field.BatchInvert(denominator)
			}

			if len(numeratorMetadata) == 0 {
				numerator = vector.Repeat(field.One(), size)
			} else {
				numerator = column.EvalExprColumn(run, numBoard).IntoRegVecSaveAlloc()
			}

			var (
				res = field.Zero()
				tmp field.Element
			)

			for k := range numerator {
				tmp.Mul(&numerator[k], &denominator[k])
				res.Add(&res, &tmp)
			}

			resLock.Lock()
			actualSum.Add(&actualSum, &res)
			resLock.Unlock()
		}
	})

	if err != nil {
		return field.Element{}, err
	}

	return actualSum, nil
}

// Test that global sum is correct
func (r LogDerivativeSum) Check(run ifaces.Runtime) error {

	var (
		params         = run.GetParams(r.ID).(LogDerivSumParams)
		actualSum, err = r.Compute(run)
	)

	if err != nil {
		return errors.New("expected a denominator without zeroes")
	}

	if actualSum != params.Sum {
		return fmt.Errorf("expected LogDerivativeSum = %s but got %s for the query %v", params.Sum.String(), actualSum.String(), r.ID)
	}

	return nil
}

// Test that global sum is correct
func (r LogDerivativeSum) CheckGnark(api frontend.API, run ifaces.GnarkRuntime) {
	panic("unexpected call")
}
