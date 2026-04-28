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

// MultilinearEval declares that N polynomials P_0, ..., P_{N-1} are all
// evaluated at the same fext-valued multilinear point. Each P_i is a column
// of size 2^NumVars whose entries represent the polynomial's hypercube
// evaluations in the gnark-crypto convention (X_1 = MSB of index).
type MultilinearEval struct {
	Pols    []ifaces.Column
	QueryID ifaces.QueryID
	NumVars int
	uuid    uuid.UUID `serde:"omit"`
}

// MultilinearEvalParams is the runtime witness for a [MultilinearEval] query:
// the evaluation point (length NumVars) and the claimed per-polynomial values.
type MultilinearEvalParams struct {
	Point []fext.Element
	Ys    []fext.Element
}

// NewMultilinearEval constructs and validates a MultilinearEval query. All
// columns must be distinct, non-empty, and of length exactly 2^numVars.
func NewMultilinearEval(id ifaces.QueryID, numVars int, pols ...ifaces.Column) MultilinearEval {
	if len(pols) == 0 {
		utils.Panic("MultilinearEval %v declared with zero polynomials", id)
	}
	if numVars <= 0 {
		utils.Panic("MultilinearEval %v: numVars must be positive, got %d", id, numVars)
	}
	expectedLen := 1 << numVars
	polsSet := collection.NewSet[ifaces.ColID]()
	for _, pol := range pols {
		if pol.Size() != expectedLen {
			utils.Panic("MultilinearEval %v: polynomial %v has size %d, expected 2^%d=%d",
				id, pol.GetColID(), pol.Size(), numVars, expectedLen)
		}
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
// The evaluation point is not included because the verifier can compute it
// independently (same convention as UnivariateEval).
func (p MultilinearEvalParams) UpdateFS(state fiatshamir.FS) {
	state.UpdateExt(p.Ys...)
}

// Check verifies the multilinear evaluation claims natively.
func (q MultilinearEval) Check(run ifaces.Runtime) error {
	params := run.GetParams(q.QueryID).(MultilinearEvalParams)

	if len(params.Point) != q.NumVars {
		return fmt.Errorf("MultilinearEval %v: point length %d != NumVars %d",
			q.QueryID, len(params.Point), q.NumVars)
	}
	if len(params.Ys) != len(q.Pols) {
		return fmt.Errorf("MultilinearEval %v: %d claimed Ys for %d polynomials",
			q.QueryID, len(params.Ys), len(q.Pols))
	}

	var errMsg string
	for k, pol := range q.Pols {
		wit := pol.GetColAssignment(run)
		vals := smartvectors.IntoRegVecExt(wit)
		actualY := multilinEvaluate(vals, params.Point)
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
		actualY := evalMultilinGnark(koalaAPI, col, params.Point)
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
