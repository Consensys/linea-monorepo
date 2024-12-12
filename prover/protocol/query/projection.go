package query

import (
	"github.com/consensys/gnark/frontend"
	"github.com/consensys/linea-monorepo/prover/protocol/ifaces"
)

type Projection struct {
	ID                 ifaces.QueryID
	ColumnsA, ColumnsB []ifaces.Column
	FilterA, FilterB   ifaces.Column
}

// NewInclusion constructs an inclusion. Will panic if it is mal-formed
func NewProjection(
	id ifaces.QueryID,
	columnsA, columnsB []ifaces.Column,
	filterA, filterB ifaces.Column,
) Projection {

	return Projection{ColumnsA: columnsA, ColumnsB: columnsB, ID: id, FilterA: filterA, FilterB: filterB}
}

// Name implements the [ifaces.Query] interface
func (r Projection) Name() ifaces.QueryID {
	return r.ID
}

// Check implements the [ifaces.Query] interface
func (r Projection) Check(run ifaces.Runtime) error {

	panic("unimplemented")
}

// GnarkCheck implements the [ifaces.Query] interface. It will panic in this
// construction because we do not have a good way to check the query within a
// circuit
func (i Projection) CheckGnark(api frontend.API, run ifaces.GnarkRuntime) {
	panic("UNSUPPORTED : can't check an Projection query directly into the circuit")
}
