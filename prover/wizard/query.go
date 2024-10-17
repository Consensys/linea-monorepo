package wizard

import (
	"github.com/consensys/gnark/frontend"
	"github.com/consensys/zkevm-monorepo/prover/crypto/fiatshamir"
	"github.com/consensys/zkevm-monorepo/prover/maths/field"
	"github.com/consensys/zkevm-monorepo/prover/protocol/ifaces"
	"github.com/consensys/zkevm-monorepo/prover/symbolic"
)

// Query represents either a constraint or a wizard oracle query over columns
// of the protocol.
type Query interface {
	symbolic.Metadata
	computeResult(run Runtime) QueryResult
	computeResultGnark(api frontend.API, run RuntimeGnark) QueryResultGnark
	Check(run Runtime) error
	CheckGnark(api frontend.API, run RuntimeGnark)
	MarkAsCompiled() bool
	IsCompiled() bool
	Round() int
	id() id
	Explain() string
	Tags() []string
	DeferToVerifier()
	IsDeferredToVerifier() bool
}

// AllActiveQueries returns the list of all the queries that are neither
// ignored or defered to the verifier.
func (api *API) AllActiveQueries() []Query {

	var (
		all = api.queries.all()
		res = make([]Query, 0, len(all))
	)

	for i := range all {
		if all[i].IsCompiled() || all[i].IsDeferredToVerifier() {
			continue
		}

		res = append(res, all[i])
	}

	return res
}

type subQuery struct {
	round             int
	compiled          bool
	deferedToVerifier bool
}

func (q *subQuery) MarkAsCompiled() bool {
	res := q.compiled
	q.compiled = true
	return res
}

func (q *subQuery) IsCompiled() bool {
	return q.compiled
}

func (q *subQuery) Round() int {
	return q.round
}

func (q *subQuery) DeferToVerifier() {
	q.deferedToVerifier = true
	q.compiled = true
}

func (q *subQuery) IsDeferredToVerifier() bool {
	return q.deferedToVerifier
}

// rowLinComb utility function used to manually check permutation and inclusion
// constraints. Will return a linear combination of i-th element of
// each list.
func rowLinComb(alpha field.Element, i int, list []ifaces.ColAssignment) field.Element {
	var res field.Element
	for j := range list {
		res.Mul(&res, &alpha)
		x := list[j].Get(i)
		res.Add(&res, &x)
	}
	return res
}

// QueryResult represents the runtime parameters of a query. As explained in
// [Query], certain type of queries can require the prover to provide runtime
// parameters in order to make the predicate verifiable. This is the case for
// [github.com/consensys/zkevm-monorepo/protocol/query.UnivariateEval] which requires the user to provide an evaluation
// point (X) and one or more alleged evaluation points (Ys) (depending on whether
// the query is applied over one or more columns for the same X).
//
// The interface requires only a method to explain how the query parameter
// should update the Fiat-Shamir state.
type QueryResult interface {
	// Update fiat-shamir with the query parameters
	UpdateFS(*fiatshamir.State)
}

// QueryResultGnark mirrors exactly [QueryResult], but in a gnark circuit.
type QueryResultGnark interface {
	// Update fiat-shamir with the query parameters in a circuit
	UpdateFS(*fiatshamir.GnarkFiatShamir)
}

// QueryResNone represents the result of a query that does not returns any
// result: for instance a global constraint. In implements the [QueryRes]
// interface.
type QueryResNone struct{}

func (q *QueryResNone) UpdateFS(_ *fiatshamir.State) {}

// QueryResNoneGnark represents the result of a query that does not returns any
// result: for instance a global constraint. In implements the [QueryResGnark]
// interface.
type QueryResNoneGnark struct{}

func (q *QueryResNoneGnark) UpdateFS(_ *fiatshamir.GnarkFiatShamir) {}

// QueryResFE represents the result of a query that returns a single field
// element.
type QueryResFE struct{ R field.Element }

func (q *QueryResFE) UpdateFS(fs *fiatshamir.State) {
	fs.Update(q.R)
}

// QueryResFEGnark is as [QueryResFE] but in a gnark circuit
type QueryResFEGnark struct{ R frontend.Variable }

func (q *QueryResFEGnark) UpdateFS(fs *fiatshamir.GnarkFiatShamir) {
	fs.Update(q.R)
}
