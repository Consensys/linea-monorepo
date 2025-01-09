package inclusion_test

import (
	"testing"

	"github.com/consensys/linea-monorepo/prover/maths/common/smartvectors"
	logderiv "github.com/consensys/linea-monorepo/prover/protocol/compiler/logderivativesum"
	"github.com/consensys/linea-monorepo/prover/protocol/distributed"
	"github.com/consensys/linea-monorepo/prover/protocol/distributed/compiler/inclusion"
	md "github.com/consensys/linea-monorepo/prover/protocol/distributed/module_discoverer"
	"github.com/consensys/linea-monorepo/prover/protocol/ifaces"
	"github.com/consensys/linea-monorepo/prover/protocol/wizard"
	"github.com/consensys/linea-monorepo/prover/symbolic"
	"github.com/stretchr/testify/require"
)

// It tests if the expression is properly distributed among modules.
// It does not test the prover steps.
func TestDistributedLogDerivSum(t *testing.T) {

	var (
		col01 ifaces.Column
	)
	// moduleComp0
	define0 := func(b0 *wizard.Builder) {
		col00 := b0.CompiledIOP.InsertCommit(0, "module0.col0", 4)
		col01 = b0.CompiledIOP.InsertCommit(0, "module0.col1", 4)
		b0.CompiledIOP.InsertGlobal(0, "module0.global0",
			symbolic.Sub(col00, symbolic.Mul(2, col01)))
	}
	// module1
	define1 := func(b *wizard.Builder) {
		col10 := b.CompiledIOP.InsertCommit(0, "module1.col0", 8)
		col11 := b.CompiledIOP.InsertCommit(0, "module1.col1", 8)
		// S \subset T , S in module0, T in module1.
		b.CompiledIOP.InsertInclusion(0, "module1.lookup0", []ifaces.Column{col10}, []ifaces.Column{col01})
		b.CompiledIOP.InsertGlobal(0, "module1.global0",
			symbolic.Sub(col11, symbolic.Mul(2, col10)))
	}
	define := func(b *wizard.Builder) {
		define0(b)
		define1(b)
	}

	// prover for module0
	prover0 := func(run *wizard.ProverRuntime) {
		run.AssignColumn("module0.col0", smartvectors.ForTest(2, 4, 1, 4))
		run.AssignColumn("module0.col1", smartvectors.ForTest(1, 2, 1, 2))
	}
	// prover for module1
	prover1 := func(run *wizard.ProverRuntime) {
		run.AssignColumn("module1.col0", smartvectors.ForTest(1, 1, 2, 1, 1, 1, 1, 2))
		run.AssignColumn("module1.col1", smartvectors.ForTest(2, 2, 4, 2, 2, 2, 2, 4))
	}

	prover := func(run *wizard.ProverRuntime) {
		prover0(run)
		prover1(run)
	}

	// in initialComp replace inclusion queries with a global LogDerivativeSum
	initialComp := wizard.Compile(define, distributed.IntoLogDerivativeSum)
	moduleComp0 := wizard.Compile(define0)
	moduleComp1 := wizard.Compile(define1)

	// Initialize the period separating module discoverer
	disc := &md.PeriodSeperatingModuleDiscoverer{}
	disc.Analyze(initialComp)

	proof := wizard.Prove(initialComp, prover)
	initialProver := proof.RunTime

	// distribute the shares to modules.
	inclusion.DistributeLogDerivativeSum(initialComp, moduleComp0, "module0", disc, initialProver)
	inclusion.DistributeLogDerivativeSum(initialComp, moduleComp1, "module1", disc, initialProver)

	logderiv.CompileLogDerivSum(moduleComp0)
	proof0 := wizard.Prove(moduleComp0, prover0)
	valid := wizard.Verify(moduleComp0, proof0)
	require.NoError(t, valid)

	logderiv.CompileLogDerivSum(moduleComp1)
	proof1 := wizard.Prove(moduleComp1, prover1)
	valid1 := wizard.Verify(moduleComp1, proof1)
	require.NoError(t, valid1)
}
