package functionals_test

import (
	"testing"

	"github.com/consensys/accelerated-crypto-monorepo/maths/common/smartvectors"
	"github.com/consensys/accelerated-crypto-monorepo/protocol/coin"
	"github.com/consensys/accelerated-crypto-monorepo/protocol/compiler"
	"github.com/consensys/accelerated-crypto-monorepo/protocol/compiler/dummy"
	"github.com/consensys/accelerated-crypto-monorepo/protocol/dedicated/functionals"
	"github.com/consensys/accelerated-crypto-monorepo/protocol/ifaces"
	"github.com/consensys/accelerated-crypto-monorepo/protocol/wizard"
	"github.com/stretchr/testify/require"
)

func TestEvalCoeff(t *testing.T) {

	var (
		x            coin.Info
		p            ifaces.Column
		acc          *ifaces.Accessor
		savedRuntime *wizard.ProverRuntime
	)

	wp := smartvectors.Rand(32)

	definer := func(b *wizard.Builder) {
		p = b.RegisterCommit("P", wp.Len())
		x = b.RegisterRandomCoin("X", coin.Field)
		acc = functionals.CoeffEval(b.CompiledIOP, "EVAL", x, p)

	}

	prover := func(run *wizard.ProverRuntime) {
		// Save the pointer toward the prover
		savedRuntime = run
		run.AssignColumn("P", wp)
	}

	compiled := wizard.Compile(definer,
		compiler.Arcane(1<<4, 1<<4),
		dummy.Compile,
	)

	proof := wizard.Prove(compiled, prover)

	xVal := savedRuntime.GetRandomCoinField(x.Name)
	accY := acc.GetVal(savedRuntime)
	expectedY := smartvectors.EvalCoeff(wp, xVal)

	require.Equal(t, accY, expectedY)

	err := wizard.Verify(compiled, proof)
	require.NoError(t, err)
}
