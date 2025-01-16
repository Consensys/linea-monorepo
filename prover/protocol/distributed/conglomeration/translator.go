package conglomeration

import (
	"github.com/consensys/linea-monorepo/prover/protocol/coin"
	"github.com/consensys/linea-monorepo/prover/protocol/column"
	"github.com/consensys/linea-monorepo/prover/protocol/ifaces"
	"github.com/consensys/linea-monorepo/prover/protocol/query"
	"github.com/consensys/linea-monorepo/prover/protocol/wizard"
	"github.com/consensys/linea-monorepo/prover/utils/collection"
)

// compTranslator is a builder struct for building a target [wizard.CompiledIOP]
// instances from another source [wizard.CompiledIOP]. All items in the built
// compiled IOP are prefixed with an identifier.
type compTranslator struct {
	Prefix string
	Target *wizard.CompiledIOP
}

// InsertColumn inserts a new column in the target compiled IOP. The column name
// is prefixed with comp.Prefix.
func (comp *compTranslator) InsertColumn(round int, name ifaces.ColID, size int, status column.Status) ifaces.Column {
	name = ifaces.ColID(comp.Prefix) + "." + name
	return comp.Target.InsertColumn(round, name, size, status)
}

// GetColumn returns a column from the target compiled IOP.
func (comp *compTranslator) GetColumn(name ifaces.ColID) ifaces.Column {
	name = ifaces.ColID(comp.Prefix) + "." + name
	return comp.Target.Columns.GetHandle(name)
}

// InsertCoin inserts a new coin in the target compiled IOP. The coin name
// is prefixed with the comp.Prefix.
func (comp *compTranslator) InsertCoin(round int, name coin.Name, type_ coin.Type, size ...int) coin.Info {
	name = coin.Name(comp.Prefix) + "." + name
	return comp.Target.InsertCoin(round, name, type_, size...)
}

// GetCoin returns a coin with the prefixed name in the target compiled IOP.
// It panics if the prefixed coin is not found.
func (comp *compTranslator) GetCoin(name coin.Name) coin.Info {
	name = coin.Name(comp.Prefix) + "." + name
	return comp.Target.Coins.Data(name)
}

// InsertQueryParams inserts a new query in the target compiled IOP prefixing the
// query name however the inner-fields of the query are not prefixed or translated.
// So it should be preferrably applied only over "Ignored" queries as the content of
// the inserted query will be invalid.
func (comp *compTranslator) InsertQueryParams(round int, q ifaces.Query) ifaces.Query {
	name := ifaces.QueryID(comp.Prefix) + "." + q.Name()
	q = copyQueryWithName(name, q)
	comp.Target.QueriesParams.AddToRound(round, name, q)
	return q
}

// copyQueryWithName returns a copy of the query with a new name.
func copyQueryWithName(name ifaces.QueryID, q ifaces.Query) ifaces.Query {
	switch q := q.(type) {
	case query.UnivariateEval:
		return query.NewUnivariateEval(name, q.Pols...)
	case query.LocalOpening:
		return query.NewLocalOpening(name, q.Pol)
	case query.InnerProduct:
		return query.NewInnerProduct(name, q.A, q.Bs...)
	case query.GrandProduct:
		return query.NewGrandProduct(q.Round, q.Inputs, name)
	case query.LogDerivativeSum:
		return query.NewLogDerivativeSum(q.Round, q.Inputs, name)
	default:
		panic("unknown query type")
	}
}

// TranslateColumnList translates a collection of pre-inserted columns
func (comp *compTranslator) TranslateColumnList(cols []ifaces.Column) []ifaces.Column {
	var res []ifaces.Column
	for _, col := range cols {
		res = append(res, comp.GetColumn(col.GetColID()))
	}
	return res
}

// TranslateColumnVecVec translates a collection of pre-inserted columns
func (comp *compTranslator) TranslateColumnVecVec(cols collection.VecVec[ifaces.ColID]) collection.VecVec[ifaces.ColID] {
	var res = collection.NewVecVec[ifaces.ColID]()
	for r, vec := range cols.Inner() {
		for _, c := range vec {
			res.AppendToInner(r, comp.GetColumn(c).GetColID())
		}
	}
	return res
}

// TranslateColumnSet translates a set of pre-inserted columns
func (comp *compTranslator) TranslateColumnSet(cols map[ifaces.ColID]struct{}) map[ifaces.ColID]struct{} {
	var res = make(map[ifaces.ColID]struct{})
	for col := range cols {
		res[comp.GetColumn(col).GetColID()] = struct{}{}
	}
	return res
}
