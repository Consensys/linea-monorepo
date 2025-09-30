package globalcs

import (
	"sync"

	"github.com/consensys/linea-monorepo/prover/protocol/accessors"
	"github.com/consensys/linea-monorepo/prover/protocol/wizard"
	"github.com/consensys/linea-monorepo/prover/protocol/zk"
	"github.com/consensys/linea-monorepo/prover/symbolic"
	"github.com/consensys/linea-monorepo/prover/symbolic/simplify"
)

// factorExpressionList applies [factorExpression] over a list of expressions
func factorExpressionList[T zk.Element](comp *wizard.CompiledIOP[T], exprList []*symbolic.Expression[T]) []*symbolic.Expression[T] {
	res := make([]*symbolic.Expression[T], len(exprList))
	var wg sync.WaitGroup

	for i, expr := range exprList {
		wg.Add(1)
		go func(i int, expr *symbolic.Expression[T]) {
			defer wg.Done()
			res[i] = factorExpression(comp, expr)
		}(i, expr)
	}

	wg.Wait()
	return res
}

// factorExpression factors expr and returns the factored expression. The
// resulting factored expression is cached in the file system as this is a
// compute intensive operation.
func factorExpression[T zk.Element](comp *wizard.CompiledIOP[T], expr *symbolic.Expression[T]) *symbolic.Expression[T] {
	flattenedExpr := flattenExpr(expr)
	return simplify.AutoSimplify(flattenedExpr)
}

// flattenExpr returns an expression equivalent to expr where the
// [accessors.FromExprAccessor] are inlined
func flattenExpr[T zk.Element](expr *symbolic.Expression[T]) *symbolic.Expression[T] {
	return expr.ReconstructBottomUp(func(e *symbolic.Expression[T], children []*symbolic.Expression[T]) (new *symbolic.Expression[T]) {

		v, isVar := e.Operator.(symbolic.Variable[T])
		if !isVar {
			return e.SameWithNewChildren(children)
		}

		fea, isFEA := v.Metadata.(*accessors.FromExprAccessor[T])

		if isFEA {
			return flattenExpr(fea.Expr)
		}

		return e
	})
}
