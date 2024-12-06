package query

import (
	"github.com/consensys/gnark/frontend"
	"github.com/consensys/linea-monorepo/prover/protocol/coin"
	"github.com/consensys/linea-monorepo/prover/protocol/ifaces"
	"github.com/consensys/linea-monorepo/prover/symbolic"
)

// LogDerivativeSum describes the mathematical expression required in the lookup.
// \sum_i f_i/(s_i+\gamma)-\sum_i m_i/(t_i+\gamma) = 1
type LogDerivativeSum struct {
	// for S it keeps the filter value f_i = 0/1
	// for T it keeps  -(m_i)
	// we have one expression per S_collapsed or T_collapsed.
	Nominator []symbolic.Expression
	// for S it keeps s_i+\gamma (s_i  = \sum _j \alpha^j.s_{i,j} might be collapsing of several columns)
	// for T keeps t_i+\gamma (t_i might be collapsing of several columns)
	// we have one expression per S_collapsed or T_collapsed.
	Dominator []symbolic.Expression
	// ID stores the identifier string of the query
	ID ifaces.QueryID
	// coins used in the Nominator/Dominator.
	// the number of coins depends on the number of different T tables.
	alpha, beta []coin.Info
}

// NewLogDerivativeSum constructs an wLogDerivativeSum. Will panic if it is mal-formed
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
