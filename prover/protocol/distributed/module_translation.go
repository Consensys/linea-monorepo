package distributed

import (
	"github.com/consensys/linea-monorepo/prover/maths/common/smartvectors"
	"github.com/consensys/linea-monorepo/prover/protocol/accessors"
	"github.com/consensys/linea-monorepo/prover/protocol/coin"
	"github.com/consensys/linea-monorepo/prover/protocol/column"
	"github.com/consensys/linea-monorepo/prover/protocol/column/verifiercol"
	"github.com/consensys/linea-monorepo/prover/protocol/ifaces"
	"github.com/consensys/linea-monorepo/prover/protocol/query"
	"github.com/consensys/linea-monorepo/prover/protocol/wizard"
	"github.com/consensys/linea-monorepo/prover/symbolic"
	"github.com/consensys/linea-monorepo/prover/utils"
)

// ModuleTranslator is a utily struct wrapping a [wizard.CompiledIOP] and
// implements the logic to build it and to translate it.
type ModuleTranslator struct {
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
func (mt *ModuleTranslator) InsertColumn(col column.Natural, atRound int) ifaces.Column {

	if col.Round() != 0 {
		utils.Panic("cannot translate a column with non-zero round %v", col.Round())
	}

	if col.Status() == column.Precomputed || col.Status() == column.VerifyingKey {
		panic("use [InsertPrecomputed] for precomputed columns")
	}

	if mt.Wiop.Columns.Exists(col.ID) {
		return mt.Wiop.Columns.GetHandle(col.ID)
	}

	newSize := NewSizeOfColumn(mt.Disc, col)
	return mt.Wiop.InsertColumn(atRound, col.ID, newSize, col.Status())
}

// InsertPrecomputed is as [InsertColumn] but specificially works for precomputed
// columns.
func (mt *ModuleTranslator) InsertPrecomputed(col column.Natural, data smartvectors.SmartVector) ifaces.Column {

	if col.Round() != 0 {
		utils.Panic("cannot translate a column with non-zero round %v", col.Round())
	}

	if col.Status() != column.Precomputed && col.Status() != column.VerifyingKey {
		panic("use [InsertPrecomputed] for precomputed columns")
	}

	if mt.Wiop.Columns.Exists(col.ID) {
		return mt.Wiop.Columns.GetHandle(col.ID)
	}

	mt.Wiop.Precomputed.InsertNew(col.ID, data)
	return mt.Wiop.InsertColumn(0, col.ID, data.Len(), col.Status())
}

// TranslateColumn returns an equivalent column from the new module. The
// function panics if the column cannot be resolved. It will happen if the
// column has an expected type or is defined from not resolvable items.
//
// The sizeHint argument is meant to deduce what the size of a translated
// [verifiercol.ConstCol]
func (mt *ModuleTranslator) TranslateColumn(col ifaces.Column) ifaces.Column {

	switch c := col.(type) {
	case column.Natural:
		return mt.Wiop.Columns.GetHandle(c.ID)
	case column.Shifted:
		return column.Shifted{
			Parent: mt.TranslateColumn(c.Parent),
			Offset: c.Offset,
		}
	case verifiercol.ConstCol:
		return verifiercol.NewConstantCol(c.F, c.Size())
	default:
		utils.Panic("unexpected type of column: type: %T, name: %v", col, col.GetColID())
	}

	return nil
}

// TranslateColumnList returns a list of equivalent columns from the new module.
// The function panics if the column cannot be resolved. It will happen if the
// column has an expected type or is defined from not resolvable items.
func (mt *ModuleTranslator) TranslateColumnList(cols []ifaces.Column) []ifaces.Column {
	res := make([]ifaces.Column, len(cols))
	for i := range res {
		res[i] = mt.TranslateColumn(cols[i])
	}
	return res
}

// TranslateExpression returns an expression corresponding to the provided
// expression but in term of the input module. When the function encounters
// a [verifiercol.Constcol] as part of the expression, it converts it into
// a [symbolic.Constant] directly as this simplifies the later steps in the
// process and is strictly equivalent.
func (mt *ModuleTranslator) TranslateExpression(expr *symbolic.Expression) *symbolic.Expression {

	return expr.ReconstructBottomUpSingleThreaded(
		func(e *symbolic.Expression, children []*symbolic.Expression) *symbolic.Expression {
			switch op := e.Operator.(type) {
			case symbolic.Variable:
				switch m := op.Metadata.(type) {
				case ifaces.Accessor:
					newAcc := mt.TranslateAccessor(m)
					return symbolic.NewVariable(newAcc)
				case ifaces.Column:
					// When finding a constcol, it is always simpler to
					// convert it into a constant sub-expression. Also,
					// it is important to account for the fact that we
					// can absolutely encounter shifted version of a
					// constant col.
					root := column.RootParents(m)
					if constcol, isconst := root.(verifiercol.ConstCol); isconst {
						return symbolic.NewConstant(constcol.F)
					}
					newCol := mt.TranslateColumn(m)
					return symbolic.NewVariable(newCol)
				case coin.Info:
					newCoin := mt.TranslateCoin(m)
					return symbolic.NewVariable(newCoin)
				default:
					return e.SameWithNewChildren(children)
				}
			default:
				return e.SameWithNewChildren(children)
			}
		},
	)
}

// TranslateExpressionList returns a list of equivalent expressions from the new
// module.
func (mt *ModuleTranslator) TranslateExpressionList(exprs []*symbolic.Expression) []*symbolic.Expression {
	res := make([]*symbolic.Expression, len(exprs))
	for i := range res {
		res[i] = mt.TranslateExpression(exprs[i])
	}
	return res
}

// TranslateCoin returns the equivalent coin from the new module.
// The function looks for a coin with the same name and inserts it
// as a [coin.FieldFromSeed] if it is not found.
func (mt *ModuleTranslator) TranslateCoin(info coin.Info) coin.Info {
	if !mt.Wiop.Coins.Exists(info.Name) {
		mt.Wiop.InsertCoin(1, info.Name, coin.FieldFromSeed)
	}
	return mt.Wiop.Coins.Data(info.Name)
}

// TranslateAccessor returns an equivalent from the new module.
func (mt *ModuleTranslator) TranslateAccessor(acc ifaces.Accessor) ifaces.Accessor {

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
		newCol := mt.TranslateColumn(a.Col)
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
func (mt *ModuleTranslator) TranslateQueryParam(query ifaces.Query) ifaces.Query {
	return mt.Wiop.QueriesParams.Data(query.Name())
}

// InsertPlonkInWizard inserts a new PlonkInWizard query in the target compiled IOP
// by translating all the related columns
func (mt *ModuleTranslator) InsertPlonkInWizard(oldQuery *query.PlonkInWizard) *query.PlonkInWizard {

	newQuery := query.NewPlonkInWizard(
		oldQuery.ID,
		mt.TranslateColumn(oldQuery.Data),
		mt.TranslateColumn(oldQuery.Selector),
		oldQuery.Circuit,
		oldQuery.PlonkOptions,
	)

	mt.Wiop.InsertPlonkInWizard(newQuery)
	return newQuery
}

// InsertLogDerivative inserts a new LogDerivative query in the target compiled IOP
// by translating all the related columns
func (mt *ModuleLPP) InsertLogDerivative(
	round int,
	id ifaces.QueryID,
	logDerivativeArgs [][2]*symbolic.Expression,
) query.LogDerivativeSum {

	resInputs := map[int]*query.LogDerivativeSumInput{}

	for _, numDenPair := range logDerivativeArgs {

		var (
			num  = numDenPair[0]
			den  = numDenPair[1]
			size = NewSizeOfList(mt.Disc, num, den)
		)

		mt.addCoinFromExpression(num)
		mt.addCoinFromExpression(den)

		if _, hasSize := resInputs[size]; !hasSize {
			resInputs[size] = &query.LogDerivativeSumInput{
				Size: size,
			}
		}

		newInp := resInputs[size]
		newInp.Numerator = append(newInp.Numerator, mt.TranslateExpression(num))
		newInp.Denominator = append(newInp.Denominator, mt.TranslateExpression(den))
	}

	return mt.Wiop.InsertLogDerivativeSum(round, id, resInputs)
}

// InsertGrandProduct inserts a new GrandProduct query in the target compiled IOP
// by translating all the related columns
func (mt *ModuleLPP) InsertGrandProduct(
	round int,
	id ifaces.QueryID,
	args [][2]*symbolic.Expression,
) query.GrandProduct {

	resInputs := make(map[int]*query.GrandProductInput)

	for _, numDenPair := range args {

		var (
			num  = numDenPair[0]
			den  = numDenPair[1]
			size = NewSizeOfList(mt.Disc, num, den)
		)

		mt.addCoinFromExpression(num)
		mt.addCoinFromExpression(den)

		if _, hasSize := resInputs[size]; !hasSize {
			resInputs[size] = &query.GrandProductInput{
				Size: size,
			}
		}

		newInp := resInputs[size]
		newInp.Numerators = append(newInp.Numerators, mt.TranslateExpression(num))
		newInp.Denominators = append(newInp.Denominators, mt.TranslateExpression(den))
	}

	return mt.Wiop.InsertGrandProduct(round, id, resInputs)
}

// InsertHorner inserts a new Horner query in the target compiled IOP
// by translating all the related columns
func (mt *ModuleLPP) InsertHorner(
	round int,
	id ifaces.QueryID,
	parts []query.HornerPart,
) query.Horner {

	newParts := []query.HornerPart{}

	for _, oldPart := range parts {

		newPart := query.HornerPart{
			Coefficients: mt.TranslateExpressionList(oldPart.Coefficients),
			SignNegative: oldPart.SignNegative,
			Selectors:    mt.TranslateColumnList(oldPart.Selectors),
			X:            mt.TranslateAccessor(oldPart.X),
		}

		mt.addCoinFromExpression(newPart.Coefficients...)
		mt.addCoinFromAccessor(newPart.X)
		newParts = append(newParts, newPart)
	}

	return mt.Wiop.InsertHornerQuery(round, id, parts)
}

func (mt *ModuleTranslator) InsertCoin(name coin.Name, round int) {

	if mt.Wiop.Coins.Exists(name) {
		return
	}

	newInfo := coin.NewInfo(name, coin.FieldFromSeed, round)
	mt.Wiop.Coins.AddToRound(round, name, newInfo)
}
