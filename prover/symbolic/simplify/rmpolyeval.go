package simplify

import (
	sym "github.com/consensys/linea-monorepo/prover/symbolic"
)

// removePolyEval rewrites all the polyeval as an equivalent sums of products
func removePolyEval(e *sym.Expression) *sym.Expression {

	constructor := func(oldExpr *sym.Expression, newChildren []*sym.Expression) *sym.Expression {

		_, ok := oldExpr.Operator.(sym.PolyEval)
		if !ok {
			return oldExpr.SameWithNewChildren(newChildren)
		}

		x := newChildren[0]
		cs := newChildren[1:]

		acc := cs[0]
		xPowi := x

		for i := 1; i < len(cs); i++ {
			// Here we want to use the default constructor to ensure that we
			// will have a merged sum at the end.
			acc = sym.Add(acc, sym.Mul(xPowi, cs[i]))
			if i+1 < len(cs) {
				// We don't use the default construct because it will collapse the
				// xPowi into a single term. The intermediate are useful because
				// it tells the evaluator to reuse the intermediate terms instead of
				// computing x^i for every term.
				xPowi = sym.NewProduct([]*sym.Expression{xPowi, x}, []int{1, 1})
			}
		}

		if oldExpr.ESHash != acc.ESHash {
			panic("ESH was altered")
		}

		return acc
	}

	return e.ReconstructBottomUp(constructor)
}
