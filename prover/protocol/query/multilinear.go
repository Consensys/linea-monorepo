package query

import (
	"errors"
	"fmt"
	"math/bits"

	"github.com/consensys/gnark/frontend"
	"github.com/consensys/linea-monorepo/prover/crypto/fiatshamir"
	"github.com/consensys/linea-monorepo/prover/maths/common/smartvectors"
	"github.com/consensys/linea-monorepo/prover/maths/field/fext"
	"github.com/consensys/linea-monorepo/prover/maths/field/koalagnark"
	"github.com/consensys/linea-monorepo/prover/protocol/ifaces"
	"github.com/consensys/linea-monorepo/prover/utils"
	"github.com/consensys/linea-monorepo/prover/utils/collection"
	"github.com/google/uuid"
)

// MultilinearEval declares that polynomials P_0, ..., P_{N-1} are each
// evaluated at their own fext-valued multilinear point. P_i is a column of
// size 2^NumVars[i]; its hypercube evaluations follow the gnark-crypto
// convention (X_1 = MSB of index).
type MultilinearEval struct {
	Pols    []ifaces.Column
	QueryID ifaces.QueryID
	NumVars []int
	uuid    uuid.UUID `serde:"omit"`
}

// MultilinearEvalParams is the runtime witness for a [MultilinearEval] query:
// Points[i] is the evaluation point for P_i (length NumVars[i]) and Ys[i] is
// the claimed evaluation P_i(Points[i]).
type MultilinearEvalParams struct {
	Points [][]fext.Element
	Ys     []fext.Element
}

// NewMultilinearEval constructs and validates a MultilinearEval query. NumVars
// for each polynomial is derived from its column size (must be a power of two).
func NewMultilinearEval(id ifaces.QueryID, pols ...ifaces.Column) MultilinearEval {
	if len(pols) == 0 {
		utils.Panic("MultilinearEval %v declared with zero polynomials", id)
	}
	polsSet := collection.NewSet[ifaces.ColID]()
	numVars := make([]int, len(pols))
	for i, pol := range pols {
		size := pol.Size()
		if size <= 0 || size&(size-1) != 0 {
			utils.Panic("MultilinearEval %v: polynomial %v size %d is not a power of two",
				id, pol.GetColID(), size)
		}
		numVars[i] = bits.TrailingZeros(uint(size))
		if polsSet.Insert(pol.GetColID()) {
			utils.Panic("MultilinearEval %v: duplicate polynomial %v", id, pol.GetColID())
		}
	}
	return MultilinearEval{QueryID: id, Pols: pols, NumVars: numVars, uuid: uuid.New()}
}

// Name implements [ifaces.Query].
func (q MultilinearEval) Name() ifaces.QueryID { return q.QueryID }

// UUID implements [ifaces.Query].
func (q MultilinearEval) UUID() uuid.UUID { return q.uuid }

// UpdateFS includes the claimed evaluation values in the Fiat-Shamir state.
// The evaluation points are not included because the verifier can compute them
// independently (same convention as UnivariateEval).
func (p MultilinearEvalParams) UpdateFS(state fiatshamir.FS) {
	state.UpdateExt(p.Ys...)
}

// Check verifies the multilinear evaluation claims natively.
func (q MultilinearEval) Check(run ifaces.Runtime) error {
	params := run.GetParams(q.QueryID).(MultilinearEvalParams)

	if len(params.Points) != len(q.Pols) {
		return fmt.Errorf("MultilinearEval %v: %d points for %d polynomials",
			q.QueryID, len(params.Points), len(q.Pols))
	}
	if len(params.Ys) != len(q.Pols) {
		return fmt.Errorf("MultilinearEval %v: %d claimed Ys for %d polynomials",
			q.QueryID, len(params.Ys), len(q.Pols))
	}

	var errMsg string
	for k, pol := range q.Pols {
		if len(params.Points[k]) != q.NumVars[k] {
			errMsg += fmt.Sprintf("MultilinearEval %v: poly %v point length %d != NumVars %d\n",
				q.QueryID, pol.GetColID(), len(params.Points[k]), q.NumVars[k])
			continue
		}
		wit := pol.GetColAssignment(run)
		vals := smartvectors.IntoRegVecExt(wit)
		actualY := multilinEvaluate(vals, params.Points[k])
		if actualY != params.Ys[k] {
			errMsg += fmt.Sprintf("MultilinearEval %v: poly %v expected %s got %s\n",
				q.QueryID, pol.GetColID(), params.Ys[k].String(), actualY.String())
		}
	}
	if errMsg != "" {
		return errors.New(errMsg)
	}
	return nil
}

// CheckGnark verifies the multilinear evaluation claims inside a gnark circuit.
func (q MultilinearEval) CheckGnark(api frontend.API, run ifaces.GnarkRuntime) {
	params := run.GetParams(q.QueryID).(GnarkMultilinearEvalParams)
	koalaAPI := koalagnark.NewAPI(api)

	for k, pol := range q.Pols {
		col := pol.GetColAssignmentGnarkExt(run)
		actualY := evalMultilinGnark(koalaAPI, col, params.Points[k])
		koalaAPI.AssertIsEqualExt(actualY, params.Ys[k])
	}
}

// multilinEvaluate evaluates the multilinear polynomial over fext given by
// vals (hypercube table, MSB-first index convention) at the given fext point
// via repeated folding. Does not mutate vals.
func multilinEvaluate(vals []fext.Element, point []fext.Element) fext.Element {
	n := len(point)
	if bits.OnesCount(uint(len(vals))) != 1 || bits.TrailingZeros(uint(len(vals))) != n {
		panic("multilinEvaluate: len(vals) must be 2^len(point)")
	}
	work := make([]fext.Element, len(vals))
	copy(work, vals)
	var t fext.Element
	for _, r := range point {
		mid := len(work) / 2
		for i := 0; i < mid; i++ {
			t.Sub(&work[i+mid], &work[i])
			t.Mul(&t, &r)
			work[i].Add(&work[i], &t)
		}
		work = work[:mid]
	}
	return work[0]
}

// evalMultilinGnark evaluates the multilinear polynomial given by col (fext
// hypercube table, MSB-first convention) at point inside a gnark circuit.
// Uses repeated folding: O(2^n) Ext multiplications over n rounds.
func evalMultilinGnark(api *koalagnark.API, col []koalagnark.Ext, point []koalagnark.Ext) koalagnark.Ext {
	work := make([]koalagnark.Ext, len(col))
	copy(work, col)
	for _, r := range point {
		mid := len(work) / 2
		for i := 0; i < mid; i++ {
			diff := api.SubExt(work[i+mid], work[i])
			rDiff := api.MulExt(r, diff)
			work[i] = api.AddExt(work[i], rDiff)
		}
		work = work[:mid]
	}
	return work[0]
}
