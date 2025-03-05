package query

import (
	"fmt"
	"math/big"

	"github.com/consensys/gnark/frontend"
	"github.com/consensys/linea-monorepo/prover/crypto/fiatshamir"
	"github.com/consensys/linea-monorepo/prover/maths/field"
	"github.com/consensys/linea-monorepo/prover/protocol/column"
	"github.com/consensys/linea-monorepo/prover/protocol/ifaces"
	"github.com/consensys/linea-monorepo/prover/symbolic"
	"github.com/consensys/linea-monorepo/prover/utils"
)

// HornerPart represents a part of a Horner evaluation query.
type HornerPart struct {
	// SignNegative indicates that the result should be negated.
	SignNegative bool
	// Coefficient is the coefficient of the term. It may be a
	// column or random linear combination of columns.
	Coefficient *symbolic.Expression
	// Selector is a boolean indicator column telling which terms
	// or [Coefficients] should be included in the Horner evaluation.
	Selector ifaces.Column
	// X is the "x" value in the Horner evaluation query. Most of the
	// time, the accessor will be a random coin. The typing to accessor
	// allows for more flexibility.
	X ifaces.Accessor
	// size indicates the size of which the horner part is running.
	// It is lazily computed thanks to the Size() column.
	size int
}

// Horner represents a Horner evaluation query. The query returns
// a sum (for every part) of values that are each evaluated as:
//
// ```
//
//	 def horner(selector, value, x, n0):
//			res := 0
//	 		count := 0
//			for sel, val in zip(selector, value) {
//	 	  	if sel {
//			  		res *= x
//			   		res += val
//			   		count += 1
//				}
//			}
//			return sign * res * (x ** n0), n0 + count
//
// ```
//
// where x and n0 are parameters of the query.
type Horner struct {
	// Round is the round of definition of the query
	Round int
	// ID is the identifier of the query in the [wizard.CompiledIOP]
	ID ifaces.QueryID
	// Parts are the parts of the query
	Parts []HornerPart
}

// HornerParamsParts represents the parameters for a part of a [Horner]
// evaluation query.
type HornerParamsPart struct {
	// N0 is an initial offset of the Horner query
	N0 int
	// N1 is the second offset of the Horner query
	N1 int
}

// HornerParams represents the parameters of the Horner evaluation query
type HornerParams struct {
	// Final result is the result of summing the Horner parts for every
	// queries.
	FinalResult field.Element
	// Parts are the parameters of the Horner parts
	Parts []HornerParamsPart
}

// NewHorner constructs a new Horner evaluation query. The function also
// performs all the well-formedness sanity-checks. In case of failure, it
// panics. In particular, it will panic if:
//
//   - the pairs [HornerPart.Selector] and [HornerPart.Coefficient] do not
//     share the same length.
func NewHorner(round int, id ifaces.QueryID, parts []HornerPart) Horner {

	for i := range parts {

		var (
			board = parts[i].Coefficient.Board()
			size  = column.ExprIsOnSameLengthHandles(&board)
		)

		mustBeNaturalOrVerifierCol(parts[i].Selector)

		if parts[i].Selector.Size() != size {
			utils.Panic("Horner part %v has a selector of size %v and a coefficient of size %v", i, parts[i].Selector.Size(), size)
		}

		parts[i].size = size
	}

	return Horner{
		Round: round,
		ID:    id,
		Parts: parts,
	}
}

// NewIncompleteHornerParams returns an incomplete [HornerParams] by using
// the [HornerParamsParts] as inputs. The function ignores the N1's values
// if they are provided and does not set
func NewHornerParamsFromIncomplete(run ifaces.Runtime, q Horner, parts []HornerParamsPart) HornerParams {
	res := HornerParams{
		Parts: parts,
	}
	res.SetResult(run, q)
	return res
}

// SetResult computes the result parameters of the Horner evaluation from
// an incomplete [HornerParams]. The object should however have a the X's
// and the N0's set. The function returns a pointer to its receiver so the
// call can be chained.
func (p *HornerParams) SetResult(run ifaces.Runtime, q Horner) *HornerParams {
	n1s, finalResult := p.GetResult(run, q)
	for i := range p.Parts {
		p.Parts[i].N1 = n1s[i]
	}
	p.FinalResult = finalResult
	return p
}

// GetResult computes and returns the result of the Horner evaluation
// from the (possibly incomplete) [HornerParams]. The function returns
// the list of the N1 values and the final result of the Horner evaluation
// query.
func (p *HornerParams) GetResult(run ifaces.Runtime, q Horner) (n1s []int, finalResult field.Element) {

	// note: this is ineffective as the final result is already allocated
	// and assigned with 0. The line is here for clarity.
	finalResult = field.Zero()
	n1s = make([]int, len(p.Parts))

	if len(q.Parts) != len(p.Parts) {
		utils.Panic("Horner query has %v parts but HornerParams has %v", len(q.Parts), len(p.Parts))
	}

	for i, part := range q.Parts {

		var (
			dataBoard = part.Coefficient.Board()
			data      = column.EvalExprColumn(run, dataBoard).IntoRegVecSaveAlloc()
			sel       = part.Selector.GetColAssignment(run).IntoRegVecSaveAlloc()
			n0        = p.Parts[i].N0
			x         = part.X.GetVal(run)
			count     = 0
			res       field.Element
			xN0       = new(field.Element).Exp(x, big.NewInt(int64(n0)))
		)

		for j := range data {
			if sel[j].IsOne() {
				res.Mul(&res, &x)
				res.Add(&res, &data[j])
				count++
			}
		}

		res.Mul(&res, xN0)

		if part.SignNegative {
			res.Neg(&res)
		}

		n1s[i] = n0 + count
		finalResult.Add(&finalResult, &res)
	}

	return n1s, finalResult
}

// Name implements the [ifaces.Query] interface
func (h *Horner) Name() ifaces.QueryID {
	return h.ID
}

// UpdateFS implements the [ifaces.QueryParams] interface. It updates
// FS with the parameters of the query.
func (h HornerParams) UpdateFS(fs *fiatshamir.State) {

	fs.Update(h.FinalResult)

	for _, part := range h.Parts {
		n1 := new(field.Element).SetInt64(int64(part.N1))
		fs.Update(*n1)
	}
}

// Check implements the [ifaces.Query] interface
func (h Horner) Check(run ifaces.Runtime) error {

	var (
		params           = run.GetParams(h.ID).(HornerParams)
		n1s, finalResult = params.GetResult(run, h)
	)

	if finalResult != params.FinalResult {
		return fmt.Errorf("expected final result %s but got %s", params.FinalResult.String(), finalResult.String())
	}

	for i, n1 := range n1s {
		if n1 != params.Parts[i].N1 {
			return fmt.Errorf("expected N1 %v but got %v", params.Parts[i].N1, n1)
		}
	}

	return nil
}

// CheckGnark implements the [ifaces.Query] interface. It will panic
// when called as this query is not intended to be explicitly verified
// by the verifier in a gnark circuit.
func (h *Horner) CheckGnark(api frontend.API, run ifaces.GnarkRuntime) {
	panic("Horner query is not intended to be explicitly verified by the verifier in a gnark circuit")
}

// Size returns the size of the columns taking part in a [HornerPart].
func (h *HornerPart) Size() int {
	return h.size
}
