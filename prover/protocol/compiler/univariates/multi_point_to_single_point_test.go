package univariates_test

import (
	"testing"

	"github.com/consensys/zkevm-monorepo/prover/maths/common/smartvectors"
	"github.com/consensys/zkevm-monorepo/prover/maths/field"
	"github.com/consensys/zkevm-monorepo/prover/protocol/compiler/dummy"
	"github.com/consensys/zkevm-monorepo/prover/protocol/compiler/univariates"
	"github.com/consensys/zkevm-monorepo/prover/protocol/ifaces"
	"github.com/consensys/zkevm-monorepo/prover/protocol/wizard"
	"github.com/stretchr/testify/require"
)

func TestMPTS(t *testing.T) {
	var (
		P1      ifaces.ColID   = "P1"
		P2      ifaces.ColID   = "P2"
		EVAL_A  ifaces.QueryID = "EVAL_A"
		EVAL_B  ifaces.QueryID = "EVAL_B"
		PolSize int            = 4
	)

	definer := func(build *wizard.Builder) {
		P1 := build.RegisterCommit(P1, PolSize) // overshadowing
		P2 := build.RegisterCommit(P2, PolSize) // overshadowing
		build.UnivariateEval(EVAL_A, P1)
		build.UnivariateEval(EVAL_B, P2)
	}

	comp := wizard.Compile(
		definer,
		univariates.MultiPointToSinglePoint(PolSize),
		dummy.Compile,
	)

	hLProver := func(assi *wizard.ProverRuntime) {
		p1 := smartvectors.Rand(PolSize)
		p2 := smartvectors.Rand(PolSize)

		assi.AssignColumn(P1, p1)
		assi.AssignColumn(P2, p2)

		xa := field.NewElement(5)
		xb := field.NewElement(6)

		ya := smartvectors.Interpolate(p1, xa)
		yb := smartvectors.Interpolate(p2, xb)

		assi.AssignUnivariate(EVAL_A, xa, ya)
		assi.AssignUnivariate(EVAL_B, xb, yb)
	}

	proof := wizard.Prove(comp, hLProver)
	err := wizard.Verify(comp, proof)
	require.NoError(t, err)
}
