package wizard

import (
	"github.com/consensys/gnark/frontend"
	"github.com/consensys/zkevm-monorepo/prover/maths/field"
	"github.com/consensys/zkevm-monorepo/prover/protocol/ifaces"
	"github.com/consensys/zkevm-monorepo/prover/symbolic"
)

type Query interface {
	symbolic.Metadata
	ComputeResult(run Runtime) QueryResult
	ComputeResultGnark(api frontend.API, run GnarkRuntime) QueryResultGnark
	Check(run Runtime) error
	CheckGnark(api frontend.API, run GnarkRuntime)
	MarkAsCompiled() bool
	IsCompiled() bool
	Round() int
	id() id
	Explain() string
	Tags() []string
	DeferToVerifier()
	IsDeferredToVerifier() bool
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
