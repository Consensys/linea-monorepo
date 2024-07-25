package wizard

import (
	"github.com/consensys/gnark/frontend"
	"github.com/consensys/zkevm-monorepo/prover/maths/common/smartvectors"
	"github.com/consensys/zkevm-monorepo/prover/symbolic"
	"github.com/consensys/zkevm-monorepo/prover/utils"
)

// ColExpression wraps [symbolic.Expression] and specializes it for the situation
// where it is built from wizard items (e.g. column, coins, accessors, query
// parameters)
//
// It implements the [Column] interface
type ColExpression struct {
	symbolic.Expression
	// board caches the value of [Expression.Board] so that it is not recomputed
	// every time.
	board *symbolic.ExpressionBoard
}

// computeBoard sets the [expression.board] field. It is not threadsafe.
func (e *ColExpression) getBoardSafe() symbolic.ExpressionBoard {
	if e.board == nil {
		board := (&e.Expression).Board()
		e.board = &board
	}
	return *e.board
}

// Round returns the latest round of the variables consistuting the expression.
// If the expression has no variables (e.g. it is a constant expression), then
// it returns 0.
func (e *ColExpression) Round() int {

	var (
		board  = e.getBoardSafe()
		leaves = board.ListVariableMetadata()
		round  = 0
	)

	for _, m := range leaves {
		v, ok := m.(interface{ Round() int })
		if !ok {
			utils.Panic("the expression has an appriopriate variable (%v) of type %T which does not have a round", m, m.String())
		}

		round = max(round, v.Round())
	}

	return 0
}

// Size checks that all the columns found in the expression have the same size
// and panics if no columns are found in the expression.
func (e *ColExpression) Size() int {

	var (
		board  = e.getBoardSafe()
		leaves = board.ListVariableMetadata()
		size   = 0
	)

	for _, m := range leaves {
		if c, ok := m.(Column); ok {
			if size == 0 {
				size = c.Size()
				continue
			}

			if size != c.Size() {
				panic("unequal size")
			}
		}
	}

	if size == 0 {
		panic("is a constant expression")
	}

	return size
}

func (e *ColExpression) String() string {
	return "/col-expression/" + e.ESHash.Text(16)
}

func (e *ColExpression) Shift(n int) Column {

	shifted := e.Expression.ReconstructBottomUp(
		func(e *symbolic.Expression, children []*symbolic.Expression) (new *symbolic.Expression) {

			v, isVar := e.Operator.(*symbolic.Variable)
			if !isVar {
				return e.SameWithNewChildren(children)
			}

			c, isCol := v.Metadata.(Column)
			if !isCol {
				return e
			}

			return symbolic.NewVariable(c.Shift(n))
		},
	)

	return &ColExpression{
		Expression: *shifted,
	}
}

// NewColExpression constructs an [expression]. It ensures that none of the children
// columns are themselves [expression] by absorbing them into the [expression].
func NewColExpression(e *symbolic.Expression) *ColExpression {

	newExpr := e.ReconstructBottomUp(
		func(e *symbolic.Expression, children []*symbolic.Expression) (new *symbolic.Expression) {

			v, isVar := e.Operator.(*symbolic.Variable)
			if !isVar {
				return e.SameWithNewChildren(children)
			}

			c, isColExpr := v.Metadata.(*ColExpression)
			if !isColExpr {
				return e
			}

			return &c.Expression
		},
	)

	return &ColExpression{
		Expression: *newExpr,
	}
}

func (e *ColExpression) GetAssignment(run Runtime) smartvectors.SmartVector {

	var (
		board  = e.getBoardSafe()
		leaves = board.ListVariableMetadata()
		size   = e.Size()
		inputs = make([]smartvectors.SmartVector, len(leaves))
	)

	for i := range leaves {
		switch v := leaves[i].(type) {
		case Accessor:
			inputs[i] = smartvectors.NewConstant(v.GetVal(run), size)
		case Column:
			inputs[i] = v.GetAssignment(run)
		}
	}

	return board.Evaluate(inputs)
}

func (e *ColExpression) GetAssignmentGnark(api frontend.API, run GnarkRuntime) []frontend.Variable {

	var (
		board  = e.getBoardSafe()
		leaves = board.ListVariableMetadata()
		size   = e.Size()
		inputs = make([][]frontend.Variable, len(leaves))
		res    = make([]frontend.Variable, size)
	)

	for i := range leaves {
		switch v := leaves[i].(type) {
		case Accessor:
			inputs[i] = []frontend.Variable{v.GetValGnark(api, run)}
		case Column:
			inputs[i] = v.GetAssignmentGnark(api, run)
		}
	}

	for i := range res {
		inputRow := make([]frontend.Variable, len(leaves))
		for j := range inputRow {

			if len(inputs[i]) == 1 {
				inputRow[i] = inputs[i][0]
				continue
			}

			inputRow[i] = inputs[i][j]
		}

		res[i] = board.GnarkEval(api, inputRow)
	}

	return res

}
