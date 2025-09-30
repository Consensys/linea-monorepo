package recursion

import (
	"github.com/consensys/linea-monorepo/prover/protocol/coin"
	"github.com/consensys/linea-monorepo/prover/protocol/column"
	"github.com/consensys/linea-monorepo/prover/protocol/column/verifiercol"
	"github.com/consensys/linea-monorepo/prover/protocol/ifaces"
	"github.com/consensys/linea-monorepo/prover/protocol/query"
	"github.com/consensys/linea-monorepo/prover/protocol/wizard"
	"github.com/consensys/linea-monorepo/prover/protocol/zk"
	"github.com/consensys/linea-monorepo/prover/utils/collection"
)

// compTranslator is a builder struct for building a target [wizard.CompiledIOP[T]]
// instances from another source [wizard.CompiledIOP[T]]. All items in the built
// compiled IOP are prefixed with an identifier.
type compTranslator[T zk.Element] struct {
	Prefix string
	Target *wizard.CompiledIOP[T][T]
}

// addPrefixToID adds a prefix <prefix>.<message> to a string-like object message
func addPrefixToID[T ~string](prefix string, id T) T {
	return T(prefix) + T(".") + id
}

// AddPrecomputedColumn inserts a precomputed column. The inserted column is not
// a fake one and actually features a precomputed value. The inserted column retains
// the same status as the original one.
func (comp *compTranslator[T]) AddPrecomputed(srcComp *wizard.CompiledIOP[T], col ifaces.Column[T]) ifaces.Column[T] {

	var (
		nat          = col.(column.Natural[T])
		prefixedName = addPrefixToID(comp.Prefix, col.GetColID())
		ass          = srcComp.Precomputed.MustGet(nat.ID)
		newCol       = comp.Target.InsertColumn(0, prefixedName, col.Size(), nat.Status())
	)

	comp.Target.Precomputed.InsertNew(prefixedName, ass)
	return newCol
}

// AddColumnList translates and inserts a list of columns from the provided list by
// prefixing their names. If the columns are already inserted then the function
// returns the already inserted column. If fake is true, then the returned columns
// are FakeColumns.
func (comp *compTranslator[T]) AddColumnList(cols []ifaces.Column[T], fake bool, round int) []ifaces.Column[T] {

	res := make([]ifaces.Column[T], len(cols))

	for i, col := range cols {

		colRound := round
		if colRound < 0 && col != nil {
			colRound = col.Round()
		}

		res[i] = comp.AddColumnAtRound(col, fake, colRound)
	}

	return res
}

// AddColumn inserts in the translator. If the column is already inserted, then the
// function returns the already inserted column. If "fake" is true, the returned
// column is a [FakeColumn].
func (comp *compTranslator[T]) AddColumnAtRound(col ifaces.Column[T], fake bool, round int) ifaces.Column[T] {

	if col == nil {
		return nil
	}

	prefixedName := addPrefixToID(comp.Prefix, col.GetColID())

	if fake {
		return &column.FakeColumn[T]{ID: prefixedName}
	}

	if _, isVCol := col.(verifiercol.VerifierCol[T]); isVCol {
		return col
	}

	if comp.Target.Columns.Exists(prefixedName) {
		return comp.Target.Columns.GetHandle(prefixedName)
	}

	return comp.Target.InsertColumn(round, prefixedName, col.Size(), col.(column.Natural[T]).Status())
}

// AddColumnVecVec translates a collection of columns
func (comp *compTranslator[T]) AddColumnVecVec(cols collection.VecVec[ifaces.ColID]) collection.VecVec[ifaces.ColID] {

	res := collection.NewVecVec[ifaces.ColID]()

	for r, vec := range cols.GetInner() {
		for _, c := range vec {

			prefixedName := addPrefixToID(comp.Prefix, c)
			res.AppendToInner(r, prefixedName)
		}
	}

	return res
}

// TranslateColumnSet translates a set of pre-inserted columns
func (comp *compTranslator[T]) TranslateColumnSet(cols map[ifaces.ColID]struct{}) map[ifaces.ColID]struct{} {

	res := make(map[ifaces.ColID]struct{})

	for c := range cols {
		prefixedName := addPrefixToID(comp.Prefix, c)
		res[prefixedName] = struct{}{}
	}

	return res
}

// AddUniEval returns a copied UnivariateEval query with fake columns
// and names.
func (comp *compTranslator[T]) AddUniEval(round int, q query.UnivariateEval[T]) query.UnivariateEval[T] {
	queryID := addPrefixToID(comp.Prefix, q.QueryID)
	pols := comp.AddColumnList(q.Pols, true, round)
	return comp.Target.InsertUnivariate(round, queryID, pols)
}

// AddCoin adds a random coin with a prefixed name in the compiled IOP
func (comp *compTranslator[T]) AddCoinAtRound(info coin.Info, round int) coin.Info {
	name := addPrefixToID(comp.Prefix, info.Name)
	switch info.Type {
	case coin.IntegerVec:
		return comp.Target.InsertCoin(round, name, info.Type, info.Size, info.UpperBound)
	case coin.Field:
		return comp.Target.InsertCoin(round, name, info.Type)
	default:
		panic("unknown coin type")
	}
}
