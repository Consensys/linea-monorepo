package query

import (
	"errors"
	"fmt"
	"sync"

	"github.com/consensys/gnark/frontend"
	"github.com/consensys/linea-monorepo/prover/crypto/fiatshamir"
	"github.com/consensys/linea-monorepo/prover/maths/common/smartvectors"
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

		for k := start; k < stop; k++ {

			var (
				size   = inputs[k].Size
				num    = inputs[k].Num
				den    = inputs[k].Den
				res, e = computeLogDerivativeSumPair(run, num, den, size)
			)

			if e != nil {
				resLock.Lock()
				err = e
				resLock.Unlock()
				return
			}

			resLock.Lock()
			if err != nil {
				resLock.Unlock()
				return
			}
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

// computeLogDerivativeSumPair computes the log derivative sum for a couple
// of numerator and denominator.
func computeLogDerivativeSumPair(run ifaces.Runtime, num, den *sym.Expression, size int) (field.Element, error) {

	var (
		numBoard            = num.Board()
		denBoard            = den.Board()
		numeratorMetadata   = numBoard.ListVariableMetadata()
		denominatorMetadata = denBoard.ListVariableMetadata()
		numerator           smartvectors.SmartVector
		denominator         smartvectors.SmartVector
		noNumerator         = len(numeratorMetadata) == 0
		noDenominator       = len(denominatorMetadata) == 0
		res                 field.Element
	)

	if noNumerator && noDenominator {
		return field.NewElement(uint64(size)), nil
	}

	if !noNumerator {
		numerator = column.EvalExprColumn(run, numBoard)
		numerator, _ = smartvectors.TryReduceSize(numerator)

		// If the denominator resolves into a constant equal to 1, then
		// this is identical to not having any denominator. If it is
		// zero, then we can directly return 0.
		if numC, ok := numerator.(*smartvectors.Constant); ok {

			v := numC.Val()

			if v.IsZero() {
				return field.Zero(), nil
			}

			if v.IsOne() {
				noNumerator = true
			}
		}
	}

	if !noDenominator {
		denominator = column.EvalExprColumn(run, denBoard)
		denominator, _ = smartvectors.TryReduceSize(denominator)

		for d := range denominator.IterateCompact() {
			if d.IsZero() {
				return field.Zero(), errors.New("denominator is zero")
			}
		}
	}

	if noNumerator {

		var (
			denominatorWindow = smartvectors.Window(denominator)
			res               = field.Zero()
		)

		denominatorWindow = field.BatchInvert(denominatorWindow)
		for i := range denominatorWindow {
			if denominatorWindow[i].IsZero() {
				return field.Element{}, errors.New("denominator is zero")
			}
			res.Add(&res, &denominatorWindow[i])
		}

		denominatorPadding, denominatorHasPadding := smartvectors.PaddingVal(denominator)
		if denominatorHasPadding {

			if denominatorPadding.IsZero() {
				return field.Zero(), fmt.Errorf("denominator padding is zero")
			}

			denominatorPadding.Inverse(&denominatorPadding)

			var (
				nbPadding        = denominator.Len() - len(denominatorWindow)
				nbPaddingAsField field.Element
			)

			nbPaddingAsField.SetInt64(int64(nbPadding))
			denominatorPadding.Mul(&denominatorPadding, &nbPaddingAsField)
			res.Add(&res, &denominatorPadding)
		}

		return res, nil
	}

	if noDenominator {

		var (
			numeratorWindow = smartvectors.Window(numerator)
			res             = field.Zero()
		)

		for i := range numeratorWindow {
			res.Add(&res, &numeratorWindow[i])
		}

		numeratorPadding, numeratorHasPadding := smartvectors.PaddingVal(numerator)
		if numeratorHasPadding {

			var (
				nbPadding        = numerator.Len() - len(numeratorWindow)
				nbPaddingAsField field.Element
			)

			nbPaddingAsField.SetInt64(int64(nbPadding))
			numeratorPadding.Mul(&numeratorPadding, &nbPaddingAsField)
			res.Add(&res, &numeratorPadding)
		}

		return res, nil
	}

	// This implementation should catch 99% of the remaining cases. This follows
	// from the fact that most vectors have a padded structure and we can take
	// advantage of that in the implementation.
	numeratorPCW, ok := numerator.(*smartvectors.PaddedCircularWindow)
	if ok {
		var (
			pv     = numeratorPCW.PaddingVal()
			offset = numeratorPCW.Offset()
			window = numeratorPCW.Window()
			size   = numeratorPCW.Len()
		)

		if pv.IsZero() && offset+len(window) <= size {

			var (
				start              = offset
				stop               = offset + len(window)
				denominatorNonZero = make([]field.Element, 0, len(window))
				numeratorNonZero   = make([]field.Element, 0, len(window))
			)

			for i := start; i < stop; i++ {

				if denominator.GetPtr(i).IsZero() {
					return field.Element{}, errors.New("denominator is zero")
				}

				if !numerator.GetPtr(i).IsZero() {
					denominatorNonZero = append(denominatorNonZero, denominator.Get(i))
					numeratorNonZero = append(numeratorNonZero, numerator.Get(i))
				}
			}

			denominatorNonZero = field.BatchInvert(denominatorNonZero)

			for i := range denominatorNonZero {
				var tmp field.Element
				tmp.Mul(&denominatorNonZero[i], &numeratorNonZero[i])
				res.Add(&res, &tmp)
			}

			return res, nil
		}
	}

	// This is the "main" case which corresponds to when both numerator and the
	// numerator are defined. In this case, we iterate over the numerator and
	// record all the pairs num/den such that num != 0. The initial capacity is
	// empirically divided by 16 as we will never need more than that 99% of
	// the time.
	//
	// The implementation is based on the following observation: the numerator
	// is always full of zeroes.
	var (
		denominatorNonZero = make([]field.Element, 0, denominator.Len()/16)
		numeratorNonZero   = make([]field.Element, 0, numerator.Len()/16)
	)

	for i := 0; i < numerator.Len(); i++ {

		if denominator.GetPtr(i).IsZero() {
			return field.Element{}, errors.New("denominator is zero")
		}

		if !numerator.GetPtr(i).IsZero() {

			denominatorNonZero = append(denominatorNonZero, denominator.Get(i))
			numeratorNonZero = append(numeratorNonZero, numerator.Get(i))
		}
	}

	denominatorNonZero = field.BatchInvert(denominatorNonZero)

	for i := range denominatorNonZero {
		var tmp field.Element
		tmp.Mul(&denominatorNonZero[i], &numeratorNonZero[i])
		res.Add(&res, &tmp)
	}

	return res, nil
}
