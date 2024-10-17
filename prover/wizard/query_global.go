package wizard

import (
	"fmt"

	"github.com/consensys/gnark/frontend"
	"github.com/consensys/zkevm-monorepo/prover/symbolic"
)

// QueryGlobal is an arbitrary low-degree arithmetic constraint between columns
// and possibly Accessors. The constraint enforces that the arithmetic
// expression vanishes completely. Implictly, the expression is required to
// have a least one column as its variable. In case the expression includes
// shifted columns, the constraint becomes void where the positions of the
// shifted column wrap around. This behavior can be cancelled using the
// NoBoundCancel flag.
type QueryGlobal struct {
	Expr          *ColExpression
	NoBoundCancel bool
	domainSize    int
	metadata      *metadata
	*subQuery
}

// NewQueryGlobal declares and returns a new [QueryGlobal]
func (api *API) NewQueryGlobal(expr *symbolic.Expression, noBoundCancel ...bool) *QueryGlobal {

	var (
		e                = NewColExpression(expr) // This reduces the column expressions
		round            = e.Round()
		domainSize       = e.Size()
		effNoBoundCancel = false
	)

	if len(noBoundCancel) > 0 {
		effNoBoundCancel = noBoundCancel[0]
	}

	q := &QueryGlobal{
		domainSize:    domainSize,
		Expr:          e,
		NoBoundCancel: effNoBoundCancel,
		metadata:      api.newMetadata(),
		subQuery: &subQuery{
			round: round,
		},
	}

	api.queries.addToRound(q.round, q)
	return q
}

func (q QueryGlobal) Check(run Runtime) error {
	res := q.Expr.GetAssignment(run).IntoRegVecSaveAlloc()
	for i := range res {
		if !res[i].IsZero() {
			// @alex: implements the error-friendly message detailling the
			// involved expression.
			return fmt.Errorf("the expression did not cancel at row: %v", i)
		}
	}
	return nil
}

func (q QueryGlobal) CheckGnark(api frontend.API, run RuntimeGnark) {
	res := q.Expr.GetAssignmentGnark(api, run)
	for i := range res {
		api.AssertIsEqual(res[i], 0)
	}
}
