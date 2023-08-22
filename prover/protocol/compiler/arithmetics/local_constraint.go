package arithmetics

import (
	"github.com/consensys/accelerated-crypto-monorepo/protocol/ifaces"
	"github.com/consensys/accelerated-crypto-monorepo/protocol/query"
	"github.com/consensys/accelerated-crypto-monorepo/protocol/variables"
	"github.com/consensys/accelerated-crypto-monorepo/protocol/wizard"
	"github.com/consensys/accelerated-crypto-monorepo/symbolic"
)

/*
Reduce a local query -> convert it into a global constraint for the same round

Basically,

P(g^5) + Q(g^-3) = 0 => L_1(x) [ P(g^-5 . x) + Q(g^3 . x) ] for x \in D
*/
func ReduceLocalConstraint(comp *wizard.CompiledIOP, q query.LocalConstraint, round int) {
	/*
		Compile down the query
	*/
	comp.QueriesNoParams.MarkAsIgnored(q.ID)

	domainSize := 0

	board := q.Board()

	for _, metadataInterface := range board.ListVariableMetadata() {
		if metadata, ok := metadataInterface.(ifaces.Column); ok {
			domainSize = metadata.Size()
			break
		}
	}

	lagrange := LagrangeOne(domainSize)
	newExpr := q.Expression.Mul(lagrange)
	newName := deriveName[ifaces.QueryID]("LOCAL", q.ID, "QUERY")
	// Make sure to never cancel the bounds here. It is a waste of
	// arithmetic degree + it wrongfully cancels on the last entry
	// if the local constraints involves a `-1` position
	comp.InsertGlobal(round, newName, newExpr, true)
}

/*
Returns a symnolic expression that evaluates (X^N - 1) / (X - 1)
(requires that N is a power of two)
*/
func LagrangeOne(n int) *symbolic.Expression {
	return variables.NewPeriodicSample(n, 0)
}
