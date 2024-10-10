package ecdsa

import (
	"github.com/consensys/linea-monorepo/prover/protocol/ifaces"
	"github.com/consensys/linea-monorepo/prover/protocol/wizard"
	sym "github.com/consensys/linea-monorepo/prover/symbolic"
)

// -- utils. Copied from prover/zkevm/prover/statemanager/statesummary/state_summary.go

// isZeroWhenInactive constraints the column to cancel when inactive.
func isZeroWhenInactive(comp *wizard.CompiledIOP, c, isActive ifaces.Column) {
	comp.InsertGlobal(
		ROUND_NR,
		ifaces.QueryIDf("%v_IS_ZERO_WHEN_INACTIVE", c.GetColID()),
		sym.Sub(c, sym.Mul(c, isActive)),
	)
}
