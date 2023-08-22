package bigrange_test

import (
	"testing"

	"github.com/consensys/accelerated-crypto-monorepo/maths/common/smartvectors"
	"github.com/consensys/accelerated-crypto-monorepo/protocol/compiler/dummy"
	"github.com/consensys/accelerated-crypto-monorepo/protocol/dedicated/bigrange"
	"github.com/consensys/accelerated-crypto-monorepo/protocol/ifaces"
	"github.com/consensys/accelerated-crypto-monorepo/protocol/wizard"
	"github.com/stretchr/testify/require"
)

func TestBigRangeFullField(t *testing.T) {

	define := func(b *wizard.Builder) {
		P := b.RegisterCommit("P", 16)
		bigrange.BigRange(b.CompiledIOP, ifaces.ColumnAsVariable(P), 16, 16, "BIGRANGE")
	}

	prover := func(run *wizard.ProverRuntime) {
		run.AssignColumn("P", smartvectors.Rand(16))
	}

	comp := wizard.Compile(define, dummy.Compile)
	proof := wizard.Prove(comp, prover)
	err := wizard.Verify(comp, proof)
	require.NoError(t, err)

}
