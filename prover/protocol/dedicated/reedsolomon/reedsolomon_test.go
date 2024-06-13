package reedsolomon_test

import (
	"testing"

	"github.com/consensys/zkevm-monorepo/prover/maths/common/smartvectors"
	"github.com/consensys/zkevm-monorepo/prover/maths/fft"
	"github.com/consensys/zkevm-monorepo/prover/protocol/compiler"
	"github.com/consensys/zkevm-monorepo/prover/protocol/compiler/dummy"
	"github.com/consensys/zkevm-monorepo/prover/protocol/dedicated/reedsolomon"
	"github.com/consensys/zkevm-monorepo/prover/protocol/wizard"
	"github.com/stretchr/testify/require"
)

func TestReedSolomon(t *testing.T) {

	wp := smartvectors.ForTest(1, 2, 4, 8, 16, 32, 64, 128, 0, 0, 0, 0, 0, 0, 0, 0)
	wp = smartvectors.FFT(wp, fft.DIF, true, 0, 0)

	definer := func(b *wizard.Builder) {
		p := b.RegisterCommit("P", wp.Len())
		reedsolomon.CheckReedSolomon(b.CompiledIOP, 2, p)
	}

	prover := func(run *wizard.ProverRuntime) {
		run.AssignColumn("P", wp)
	}

	compiled := wizard.Compile(definer,
		compiler.Arcane(8, 8),
		dummy.Compile,
	)

	proof := wizard.Prove(compiled, prover)
	err := wizard.Verify(compiled, proof)
	require.NoError(t, err)

}
