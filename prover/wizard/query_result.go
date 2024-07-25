package wizard

import (
	"github.com/consensys/gnark/frontend"
	"github.com/consensys/zkevm-monorepo/prover/crypto/fiatshamir"
	"github.com/consensys/zkevm-monorepo/prover/maths/field"
)

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

// QueryResFE represents the result of a query that returns a single field
// element.
type QueryResFESlice struct{ R []field.Element }

func (q *QueryResFESlice) UpdateFS(fs *fiatshamir.State) {
	fs.UpdateVec(q.R)
}

// QueryResFEGnark is as [QueryResFE] but in a gnark circuit
type QueryResFESliceGnark struct{ R []frontend.Variable }

func (q *QueryResFESliceGnark) UpdateFS(fs *fiatshamir.GnarkFiatShamir) {
	fs.UpdateVec(q.R)
}
