package dummy_test

import (
	"testing"

	"github.com/consensys/zkevm-monorepo/prover/maths/common/smartvectors"
	"github.com/consensys/zkevm-monorepo/prover/protocol/compiler/dummy"
	"github.com/consensys/zkevm-monorepo/prover/protocol/wizard"
	"github.com/stretchr/testify/require"
)

// Simply runs the lazy commit over a simple commitment
func TestLazyCommit(t *testing.T) {

	define := func(b *wizard.Builder) {
		_ = b.RegisterCommit("P", 1<<6)
	}

	prover := func(pr *wizard.ProverRuntime) {
		pr.AssignColumn("P", smartvectors.Rand(1<<6))
	}

	comp := wizard.Compile(define, dummy.LazyCommit)
	proof := wizard.Prove(comp, prover)
	err := wizard.Verify(comp, proof)

	require.NoError(t, err)
}
