package query

import (
	"github.com/consensys/gnark/frontend"
	"github.com/consensys/linea-monorepo/prover/protocol/coin"
	"github.com/consensys/linea-monorepo/prover/protocol/ifaces"
	"github.com/consensys/linea-monorepo/prover/symbolic"
)

// LogDerivativeSum describes the mathematical expression required in the lookup
// \sum_i f_i/(s_i+\gamma)-\sum_i m_i/(t_i+\gamma)
type LogDerivativeSum struct {
	// for S it keeps the filter value 0/1
	// for T it keeps  -(m_i)
	Nominator []symbolic.Expression
	// for S it keeps s_i+\gamma (s_i might be collapsing of several columns)
	// for T keeps t_i+\gamma (t_i might be collapsing of several columns)
	Dominator []symbolic.Expression
	// ID stores the identifier string of the query
	ID ifaces.QueryID
	// coins used in the Nominator/Dominator
	alpha, beta coin.Info
	// size is later used for z-Packing
	 size int
}

// NewLogDerivativeSum constructs an inclusion. Will panic if it is mal-formed
func NewLogDerivativeSum(
	id ifaces.QueryID,
	nominator []symbolic.Expression,
	dominator []symbolic.Expression,
) LogDerivativeSum {

	return LogDerivativeSum{Nominator: nominator, Dominator: dominator, ID: id}
}

// Name implements the [ifaces.Query] interface
func (r LogDerivativeSum) Name() ifaces.QueryID {
	return r.ID
}

// Check implements the [ifaces.Query] interface
func (r LogDerivativeSum) Check(run ifaces.Runtime) error {
	var errLU error
	return errLU
}

// GnarkCheck implements the [ifaces.Query] interface. It will panic in this
// construction because we do not have a good way to check the query within a
// circuit
func (i LogDerivativeSum) CheckGnark(api frontend.API, run ifaces.GnarkRuntime) {
	panic("UNSUPPORTED : can't check an inclusion query directly into the circuit")
}

// it is very similar to LogDerivaitiveSum, but with the difference that it can have prover step.
// this is the query that is imposed over the segment.
type LogDerivativeSegment struct {
	logDerivSum  LogDerivativeSum
	// the coin used for generating alpha and beta
	seed coin.Info
}

// all the methods relevant to the query interface.
