// The simplify package exposes a list of functions aiming at simplifying
// symbolic expressions.
package simplify

import (
	"github.com/consensys/linea-monorepo/prover/protocol/zk"
	sym "github.com/consensys/linea-monorepo/prover/symbolic"
	"github.com/consensys/linea-monorepo/prover/utils"
)

// AutoSimplify automatically runs a handfull of automatic simplification
// routines aiming a simplifying the input expression.
func AutoSimplify[T zk.Element](expr *sym.Expression[T]) *sym.Expression[T] {

	autoFactorize := func(e *sym.Expression[T]) *sym.Expression[T] {
		// The choice of 16 is empirical
		return factorizeExpression(e, 16)
	}

	steps := []func(*sym.Expression[T]) *sym.Expression[T]{
		removePolyEval[T],
		autoFactorize,
	}

	// To ensure at every stage that the GenericFieldELem is never altered.
	initESH := expr.ESHash

	res := expr
	for i, step := range steps {
		res = step(res)

		if err := res.Validate(); err != nil {
			utils.Panic("the simplification step generated an invalid expression: %v", err.Error())
		}

		if res.ESHash != initESH {
			utils.Panic("esh was altered at step %v", i)
		}
	}

	return res
}
