package conglomeration

import (
	"fmt"

	"github.com/consensys/gnark/frontend"
	"github.com/consensys/linea-monorepo/prover/crypto/fiatshamir"
	"github.com/consensys/linea-monorepo/prover/crypto/mimc/gkrmimc"
	"github.com/consensys/linea-monorepo/prover/maths/field"
	"github.com/consensys/linea-monorepo/prover/protocol/coin"
	"github.com/consensys/linea-monorepo/prover/protocol/column"
	"github.com/consensys/linea-monorepo/prover/protocol/column/verifiercol"
	"github.com/consensys/linea-monorepo/prover/protocol/ifaces"
	"github.com/consensys/linea-monorepo/prover/protocol/query"
	"github.com/consensys/linea-monorepo/prover/protocol/wizard"
	"github.com/consensys/linea-monorepo/prover/utils/collection"
)

var (
	_ wizard.Runtime      = &runtimeTranslator{}
	_ wizard.GnarkRuntime = &gnarkRuntimeTranslator{}
)

// compTranslator is a builder struct for building a target [wizard.CompiledIOP]
// instances from another source [wizard.CompiledIOP]. All items in the built
// compiled IOP are prefixed with an identifier.
type compTranslator struct {
	Prefix string
	Target *wizard.CompiledIOP
}

// runtimeTranslator is an adapter structure prefixing every ColID and QueryID and
// coin.Name with a prefix string.
type runtimeTranslator struct {
	Prefix string
	Rt     wizard.Runtime
}

// gnarkRuntimeTranslator is as [runtimeTranslator] but for [wizard.GnarkRuntime]
type gnarkRuntimeTranslator struct {
	Prefix string
	Rt     wizard.GnarkRuntime
}

// InsertColumn inserts a new column in the target compiled IOP. The column name
// is prefixed with comp.Prefix. The function checks that the passed column does
// not have a precomputed status (e.g. either precomputed or verifying key).
func (comp *compTranslator) InsertColumn(col column.Natural) ifaces.Column {

	switch col.Status() {
	case column.Precomputed, column.VerifyingKey:
		panic("cannot insert a precomputed or verifying key column as normal column. Use [InsertPrecomputed] instead")
	}
	name := ifaces.ColID(comp.Prefix) + "." + col.ID
	return comp.Target.InsertColumn(col.Round(), name, col.Size(), col.Status())
}

// InsertPrecomputed inserts a new column as a precomputed column to the target
// compiled IOP. To differ with [InsertColumn], this method does also add the
// column to the list of precomputed columns.
func (comp *compTranslator) InsertPrecomputed(col column.Natural, ass ifaces.ColAssignment) ifaces.Column {
	name := ifaces.ColID(comp.Prefix) + "." + col.ID

	switch col.Status() {
	case column.VerifyingKey:
		// assertedly, the round of a precomputed column is always 0
		col := comp.Target.InsertColumn(0, name, col.Size(), col.Status())
		comp.Target.Precomputed.InsertNew(name, ass)
		return col
	case column.Precomputed, column.Ignored:
		return comp.Target.InsertPrecomputed(name, ass)
	default:
		panic(fmt.Sprintf("not a precomputed column: status=%v name=%v", col.Status().String(), col.ID))
	}
}

// InsertColumns inserts a list of columns in the target compiled IOP by adding
// a prefix to their names. The inputs columns are expected to be of type
// Natural or this will lead to a panic.
func (comp *compTranslator) InsertColumns(cols []ifaces.Column) []ifaces.Column {
	res := make([]ifaces.Column, 0, len(cols))
	for i := range cols {
		r := comp.InsertColumn(cols[i].(column.Natural))
		res = append(res, r)
	}
	return res
}

// GetColumn returns a column from the target compiled IOP.
func (comp *compTranslator) GetColumn(name ifaces.ColID) ifaces.Column {
	name = ifaces.ColID(comp.Prefix) + "." + name
	return comp.Target.Columns.GetHandle(name)
}

// InsertCoin inserts a new coin in the target compiled IOP. The coin name
// is prefixed with the comp.Prefix.
func (comp *compTranslator) InsertCoin(info coin.Info) coin.Info {
	name := coin.Name(comp.Prefix) + "." + info.Name
	switch info.Type {
	case coin.IntegerVec:
		return comp.Target.InsertCoin(info.Round, name, info.Type, info.Size, info.UpperBound)
	case coin.Field:
		return comp.Target.InsertCoin(info.Round, name, info.Type)
	default:
		panic("unknown coin type")
	}
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

	var q2 ifaces.Query
	switch q := q.(type) {
	case query.UnivariateEval:
		q2 = query.NewUnivariateEval(name, q.Pols...)
	case query.LocalOpening:
		q2 = query.NewLocalOpening(name, q.Pol)
	case query.InnerProduct:
		q2 = query.NewInnerProduct(name, q.A, q.Bs...)
	case query.GrandProduct:
		q2 = query.NewGrandProduct(q.Round, q.Inputs, name)
	case query.LogDerivativeSum:
		q2 = query.NewLogDerivativeSum(q.Round, q.Inputs, name)
	default:
		panic("unknown query type")
	}

	comp.Target.QueriesParams.AddToRound(round, name, q2)
	comp.Target.QueriesParams.MarkAsIgnored(q2.Name())
	return q2
}

// TranslateColumnList translates a collection of pre-inserted columns.
// If one of the columns provided in the list is nil, it will be ignored
// and the function will return nil at the same position in the returned
// list of column.
func (comp *compTranslator) TranslateColumnList(cols []ifaces.Column) []ifaces.Column {
	res := make([]ifaces.Column, 0, len(cols))
	for _, col := range cols {

		if col == nil {
			res = append(res, nil)
			continue
		}

		res = append(res, comp.GetColumn(col.GetColID()))
	}
	return res
}

// TranslateColumnVecVec translates a collection of pre-inserted columns
func (comp *compTranslator) TranslateColumnVecVec(cols collection.VecVec[ifaces.ColID]) collection.VecVec[ifaces.ColID] {
	var res = collection.NewVecVec[ifaces.ColID]()
	for r, vec := range cols.Inner() {
		for _, c := range vec {

			// If it does not exists, then it is a verifier column
			if !comp.Target.Columns.Exists(c) {
				res.AppendToInner(r, c)
				continue
			}

			res.AppendToInner(r, comp.GetColumn(c).GetColID())
		}
	}
	return res
}

// TranslateColumnSet translates a set of pre-inserted columns
func (comp *compTranslator) TranslateColumnSet(cols map[ifaces.ColID]struct{}) map[ifaces.ColID]struct{} {
	var res = make(map[ifaces.ColID]struct{})
	for col := range cols {

		// If it does not exists, then it is a verifier column
		if !comp.Target.Columns.Exists(col) {
			res[col] = struct{}{}
			continue
		}

		res[comp.GetColumn(col).GetColID()] = struct{}{}
	}
	return res
}

// TranslateUniEval returns a copied UnivariateEval query with the columns translated
// and the names translated. The returned query is registered in the translator comp.
func (comp *compTranslator) TranslateUniEval(round int, q query.UnivariateEval) query.UnivariateEval {
	newPols := make([]ifaces.Column, len(q.Pols))
	for i := range newPols {
		if _, ok := q.Pols[i].(verifiercol.VerifierCol); ok {
			newPols[i] = q.Pols[i]
			continue
		}

		newPols[i] = comp.GetColumn(q.Pols[i].GetColID())
	}
	var res = query.NewUnivariateEval(q.QueryID, newPols...)
	return comp.InsertQueryParams(round, res).(query.UnivariateEval)
}

func (run *runtimeTranslator) GetColumn(name ifaces.ColID) ifaces.ColAssignment {
	name = ifaces.ColID(run.Prefix) + "." + name
	return run.Rt.GetColumn(name)
}

func (run *runtimeTranslator) GetColumnAt(name ifaces.ColID, pos int) field.Element {
	name = ifaces.ColID(run.Prefix) + "." + name
	return run.Rt.GetColumnAt(name, pos)
}

func (run *runtimeTranslator) GetRandomCoinField(name coin.Name) field.Element {
	name = coin.Name(run.Prefix) + "." + name
	return run.Rt.GetRandomCoinField(name)
}

func (run *runtimeTranslator) GetRandomCoinIntegerVec(name coin.Name) []int {
	name = coin.Name(run.Prefix) + "." + name
	return run.Rt.GetRandomCoinIntegerVec(name)
}

func (run *runtimeTranslator) GetParams(id ifaces.QueryID) ifaces.QueryParams {
	id = ifaces.QueryID(run.Prefix) + "." + id
	return run.Rt.GetParams(id)
}

func (run *runtimeTranslator) GetSpec() *wizard.CompiledIOP {
	return run.Rt.GetSpec()
}

func (run *runtimeTranslator) GetPublicInput(name string) field.Element {
	name = run.Prefix + "." + name
	return run.Rt.GetPublicInput(name)
}

func (run *runtimeTranslator) GetGrandProductParams(name ifaces.QueryID) query.GrandProductParams {
	name = ifaces.QueryID(run.Prefix) + "." + name
	return run.Rt.GetGrandProductParams(name)
}

func (run *runtimeTranslator) GetLogDerivSumParams(name ifaces.QueryID) query.LogDerivSumParams {
	name = ifaces.QueryID(run.Prefix) + "." + name
	return run.Rt.GetLogDerivSumParams(name)
}

func (run *runtimeTranslator) GetLocalPointEvalParams(name ifaces.QueryID) query.LocalOpeningParams {
	name = ifaces.QueryID(run.Prefix) + "." + name
	return run.Rt.GetLocalPointEvalParams(name)
}

func (run *runtimeTranslator) GetInnerProductParams(name ifaces.QueryID) query.InnerProductParams {
	name = ifaces.QueryID(run.Prefix) + "." + name
	return run.Rt.GetInnerProductParams(name)
}

func (run *runtimeTranslator) GetUnivariateEval(name ifaces.QueryID) query.UnivariateEval {
	name = ifaces.QueryID(run.Prefix) + "." + name
	return run.Rt.GetUnivariateEval(name)
}

func (run *runtimeTranslator) GetUnivariateParams(name ifaces.QueryID) query.UnivariateEvalParams {
	name = ifaces.QueryID(run.Prefix) + "." + name
	return run.Rt.GetUnivariateParams(name)
}

func (run *runtimeTranslator) Fs() *fiatshamir.State {
	return run.Rt.Fs()
}

func (run *runtimeTranslator) FsHistory() [][2][]field.Element {
	return run.Rt.FsHistory()
}

func (run *runtimeTranslator) InsertCoin(name coin.Name, value any) {
	name = coin.Name(run.Prefix) + "." + name
	run.Rt.InsertCoin(name, value)
}

func (run *runtimeTranslator) GetState(name string) (any, bool) {
	name = run.Prefix + "." + name
	return run.Rt.GetState(name)
}

func (run *runtimeTranslator) SetState(name string, value any) {
	name = run.Prefix + "." + name
	run.Rt.SetState(name, value)
}

func (run *runtimeTranslator) GetQuery(name ifaces.QueryID) ifaces.Query {
	name = ifaces.QueryID(run.Prefix) + "." + name
	return run.Rt.GetQuery(name)
}

func (run *gnarkRuntimeTranslator) GetColumn(name ifaces.ColID) []frontend.Variable {
	name = ifaces.ColID(run.Prefix) + "." + name
	return run.Rt.GetColumn(name)
}

func (run *gnarkRuntimeTranslator) GetColumnAt(name ifaces.ColID, at int) frontend.Variable {
	name = ifaces.ColID(run.Prefix) + "." + name
	return run.Rt.GetColumnAt(name, at)
}

func (run *gnarkRuntimeTranslator) GetRandomCoinField(name coin.Name) frontend.Variable {
	name = coin.Name(run.Prefix) + "." + name
	return run.Rt.GetRandomCoinField(name)
}

func (run *gnarkRuntimeTranslator) GetRandomCoinIntegerVec(name coin.Name) []frontend.Variable {
	name = coin.Name(run.Prefix) + "." + name
	return run.Rt.GetRandomCoinIntegerVec(name)
}

func (run *gnarkRuntimeTranslator) GetParams(id ifaces.QueryID) ifaces.GnarkQueryParams {
	id = ifaces.QueryID(run.Prefix) + "." + id
	return run.Rt.GetParams(id)
}

func (run *gnarkRuntimeTranslator) GetSpec() *wizard.CompiledIOP {
	return run.Rt.GetSpec()
}

func (run *gnarkRuntimeTranslator) GetPublicInput(api frontend.API, name string) frontend.Variable {
	name = run.Prefix + "." + name
	return run.Rt.GetPublicInput(api, name)
}

func (run *gnarkRuntimeTranslator) GetGrandProductParams(name ifaces.QueryID) query.GnarkGrandProductParams {
	name = ifaces.QueryID(run.Prefix) + "." + name
	return run.Rt.GetGrandProductParams(name)
}

func (run *gnarkRuntimeTranslator) GetLogDerivSumParams(name ifaces.QueryID) query.GnarkLogDerivSumParams {
	name = ifaces.QueryID(run.Prefix) + "." + name
	return run.Rt.GetLogDerivSumParams(name)
}

func (run *gnarkRuntimeTranslator) GetLocalPointEvalParams(name ifaces.QueryID) query.GnarkLocalOpeningParams {
	name = ifaces.QueryID(run.Prefix) + "." + name
	return run.Rt.GetLocalPointEvalParams(name)
}

func (run *gnarkRuntimeTranslator) GetInnerProductParams(name ifaces.QueryID) query.GnarkInnerProductParams {
	name = ifaces.QueryID(run.Prefix) + "." + name
	return run.Rt.GetInnerProductParams(name)
}

func (run *gnarkRuntimeTranslator) GetUnivariateEval(name ifaces.QueryID) query.UnivariateEval {
	name = ifaces.QueryID(run.Prefix) + "." + name
	return run.Rt.GetUnivariateEval(name)
}

func (run *gnarkRuntimeTranslator) GetUnivariateParams(name ifaces.QueryID) query.GnarkUnivariateEvalParams {
	name = ifaces.QueryID(run.Prefix) + "." + name
	return run.Rt.GetUnivariateParams(name)
}

func (run *gnarkRuntimeTranslator) Fs() *fiatshamir.GnarkFiatShamir {
	return run.Rt.Fs()
}

func (run *gnarkRuntimeTranslator) FsHistory() [][2][]frontend.Variable {
	return run.Rt.FsHistory()
}

func (run *gnarkRuntimeTranslator) GetHasherFactory() *gkrmimc.HasherFactory {
	return run.Rt.GetHasherFactory()
}

func (run *gnarkRuntimeTranslator) InsertCoin(name coin.Name, value any) {
	name = coin.Name(run.Prefix) + "." + name
	run.Rt.InsertCoin(name, value)
}

func (run *gnarkRuntimeTranslator) GetState(name string) (any, bool) {
	name = run.Prefix + "." + name
	return run.Rt.GetState(name)
}

func (run *gnarkRuntimeTranslator) SetState(name string, value any) {
	name = run.Prefix + "." + name
	run.Rt.SetState(name, value)
}

func (run *gnarkRuntimeTranslator) GetQuery(name ifaces.QueryID) ifaces.Query {
	name = ifaces.QueryID(run.Prefix) + "." + name
	return run.Rt.GetQuery(name)
}
