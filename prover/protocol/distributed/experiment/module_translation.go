package experiment

import (
	"github.com/consensys/linea-monorepo/prover/protocol/accessors"
	"github.com/consensys/linea-monorepo/prover/protocol/coin"
	"github.com/consensys/linea-monorepo/prover/protocol/column"
	"github.com/consensys/linea-monorepo/prover/protocol/column/verifiercol"
	"github.com/consensys/linea-monorepo/prover/protocol/ifaces"
	"github.com/consensys/linea-monorepo/prover/protocol/query"
	"github.com/consensys/linea-monorepo/prover/protocol/variables"
	"github.com/consensys/linea-monorepo/prover/protocol/wizard"
	"github.com/consensys/linea-monorepo/prover/symbolic"
	"github.com/consensys/linea-monorepo/prover/utils"
)

// moduleTranslator is a utily struct wrapping a [wizard.CompiledIOP] and
// implements the logic to build it and to translate it.
type moduleTranslator struct {
	Disc ModuleDiscoverer
	Wiop *wizard.CompiledIOP
}

// InsertColumn inserts a new column in the target compiled IOP. The column
// name the column names is kept identical to the original. The size is
// adjusted to match the one of the module following the [ModuleDiscoverer].
//
// The function returns the inserted column. If the column had alreadu been
// inserted the function is a no-op and returns the already inserted one.
//
// The function assumes and asserts that col.Round() == 0. Otherwise, it panics.
//
// IsLPP indicates if the column is part of the LPP part of the module and is
// inserted at round 1. Otherwise, it is inserted at round 0.
func (mt *moduleTranslator) InsertColumn(col column.Natural, atRound int) ifaces.Column {

	if col.Round() != 0 {
		utils.Panic("cannot translate a column with non-zero round %v", col.Round())
	}

	if mt.Wiop.Columns.Exists(col.ID) {
		return mt.Wiop.Columns.GetHandle(col.ID)
	}

	newSize := NewSizeOfColumn(mt.Disc, col)
	return mt.Wiop.InsertColumn(atRound, col.ID, newSize, col.Status())
}

// TranslateColumn returns an equivalent column from the new module. The
// function panics if the column cannot be resolved. It will happen if the
// column has an expected type or is defined from not resolvable items.
//
// The sizeHint argument is meant to deduce what the size of a translated
// [verifiercol.ConstCol]
func (mt *moduleTranslator) TranslateColumn(col ifaces.Column, sizeHint int) ifaces.Column {

	switch c := col.(type) {
	case column.Natural:
		return mt.Wiop.Columns.GetHandle(c.ID)
	case column.Shifted:
		return mt.TranslateColumn(c.Parent, sizeHint)
	case verifiercol.ConstCol:
		return verifiercol.NewConstantCol(c.F, sizeHint)
	default:
		utils.Panic("expected type of column: type: %T, name: %v", col, col.GetColID())
	}

	return nil
}

// TranslateExpression returns an expression corresponding to the provided
// expression but in term of the input module.
func (mt *moduleTranslator) TranslateExpression(expr *symbolic.Expression) *symbolic.Expression {

	sizeHint := NewSizeOfExpr(mt.Disc, expr)

	return expr.ReconstructBottomUp(func(e *symbolic.Expression, children []*symbolic.Expression) *symbolic.Expression {

		switch op := e.Operator.(type) {
		case symbolic.Variable:
			switch m := op.Metadata.(type) {
			case ifaces.Accessor:
				newAcc := mt.TranslateAccessor(m)
				return symbolic.NewVariable(newAcc)
			case ifaces.Column:
				newCol := mt.TranslateColumn(m, sizeHint)
				return symbolic.NewVariable(newCol)
			case coin.Info:
				newCoin := mt.TranslateCoin(m)
				return symbolic.NewVariable(newCoin)
			case variables.X, variables.PeriodicSample:
				return e
			}
		case symbolic.Constant:
			return e
		case symbolic.LinComb, symbolic.Product, symbolic.PolyEval:
			return e.SameWithNewChildren(children)
		}

		return e
	})
}

// TranslateCoin returns the equivalent coin from the new module.
// The function looks for a coin with the same name and panics if
// the coin was not found.
func (mt *moduleTranslator) TranslateCoin(coin coin.Info) coin.Info {
	return mt.Wiop.Coins.Data(coin.Name)
}

// TranslateAccessor returns an equivalent from the new module.
func (mt *moduleTranslator) TranslateAccessor(acc ifaces.Accessor) ifaces.Accessor {

	switch a := acc.(type) {

	case *accessors.FromConstAccessor:
		// The constant accessor has no information to translate
		// (it's just a field.Element) so we just return it.
		return a

	case *accessors.FromExprAccessor:
		newExpr := mt.TranslateExpression(a.Expr)
		return accessors.NewFromExpression(newExpr, a.ExprName)

	case *accessors.FromCoinAccessor:
		newCoin := mt.TranslateCoin(a.Info)
		return accessors.NewFromCoin(newCoin)

	case *accessors.FromPublicColumn:
		newCol := mt.TranslateColumn(a.Col, 1)
		return accessors.NewFromPublicColumn(newCol, a.Pos)

	case *accessors.FromLocalOpeningYAccessor:
		newLo := mt.TranslateQueryParam(a.Q).(query.LocalOpening)
		return accessors.NewLocalOpeningAccessor(newLo, a.QRound)

	default:
		utils.Panic("unexpected type of accessor: %T", acc)
	}

	return nil
}

// TranslateQueryParam returns an equivalent query from the new module.
// The function will only look for queries with the same name.
func (mt *moduleTranslator) TranslateQueryParam(query ifaces.Query) ifaces.Query {
	return mt.Wiop.QueriesParams.Data(query.Name())
}
