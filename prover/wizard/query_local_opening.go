package wizard

import (
	"github.com/consensys/gnark/frontend"
	"github.com/consensys/zkevm-monorepo/prover/maths/field"
)

var (
	_ Accessor = &QueryLocalOpening{}
	_ Query    = &QueryLocalOpening{}
)

type QueryLocalOpening struct {
	Col      Column
	Position int
	metadata *metadata
	*subQuery
}

func (api *API) NewLocalOpening(col Column, pos int) *QueryLocalOpening {
	var (
		q = &QueryLocalOpening{
			metadata: api.newMetadata(),
			Col:      col,
			Position: pos,
			subQuery: &subQuery{
				round: col.Round(),
			},
		}
	)
	api.queries.addToRound(q.round, q)
	return q
}

func (q QueryLocalOpening) ComputeResult(run Runtime) QueryResult {
	// For efficiency, check in the runtime if the result is not already
	// available.
	return &QueryResFE{
		R: q.Col.GetAssignment(run).Get(q.Position),
	}
}

func (q QueryLocalOpening) ComputeResultGnark(api frontend.API, run GnarkRuntime) QueryResultGnark {
	return &QueryResFEGnark{
		R: q.Col.GetAssignmentGnark(api, run)[q.Position],
	}
}

func (q QueryLocalOpening) GetVal(run Runtime) field.Element {
	return run.getOrComputeQueryRes(&q).(*QueryResFE).R
}

func (q QueryLocalOpening) GetValGnark(api frontend.API, run GnarkRuntime) frontend.Variable {
	v, ok := run.tryGetQueryRes(&q)
	if !ok {
		panic("missing from the proof")
	}
	return v.(*QueryResFEGnark).R
}
