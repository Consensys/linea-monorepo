package inclusion_test

import (
	"testing"

	"github.com/consensys/linea-monorepo/prover/maths/common/smartvectors"
	"github.com/consensys/linea-monorepo/prover/protocol/compiler/dummy"
	logderiv "github.com/consensys/linea-monorepo/prover/protocol/compiler/logderivativesum"
	"github.com/consensys/linea-monorepo/prover/protocol/distributed"
	"github.com/consensys/linea-monorepo/prover/protocol/distributed/compiler/inclusion"
	md "github.com/consensys/linea-monorepo/prover/protocol/distributed/namebaseddiscoverer"
	"github.com/consensys/linea-monorepo/prover/protocol/ifaces"
	"github.com/consensys/linea-monorepo/prover/protocol/wizard"
	"github.com/consensys/linea-monorepo/prover/symbolic"
	"github.com/stretchr/testify/require"
)

// It tests DistributedLogDerivSum.
func TestDistributedLogDerivSum(t *testing.T) {
	const (
		numSegModule0 = 2
		numSegModule1 = 4
	)

	//initialComp
	define := func(b *wizard.Builder) {
		col00 := b.CompiledIOP.InsertCommit(0, "module0.col0", 4)
		col01 := b.CompiledIOP.InsertCommit(0, "module0.col1", 4)
		b.CompiledIOP.InsertGlobal(0, "module0.global0",
			symbolic.Sub(col00, symbolic.Mul(2, col01)))

		col10 := b.CompiledIOP.InsertCommit(0, "module1.col0", 8)
		col11 := b.CompiledIOP.InsertCommit(0, "module1.col1", 8)
		// the inclusion query: S \subset T , S in module0, T in module1.
		b.CompiledIOP.InsertInclusion(0, "module1.lookup0", []ifaces.Column{col10}, []ifaces.Column{col01})
		b.CompiledIOP.InsertGlobal(0, "module1.global0",
			symbolic.Sub(col11, symbolic.Mul(2, col10)))
	}

	// initialProver
	prover := func(run *wizard.ProverRuntime) {
		run.AssignColumn("module0.col0", smartvectors.ForTest(2, 4, 1, 4))
		run.AssignColumn("module0.col1", smartvectors.ForTest(1, 2, 1, 2))

		run.AssignColumn("module1.col0", smartvectors.ForTest(1, 1, 2, 1, 1, 1, 1, 2))
		run.AssignColumn("module1.col1", smartvectors.ForTest(2, 2, 4, 2, 2, 2, 2, 4))
	}

	// in initialComp replace inclusion queries with a global LogDerivativeSum
	// it also create new columns relevant to the preparation such as multiplicity columns.
	initialComp := wizard.Compile(define, distributed.IntoLogDerivativeSum)

	// Initialize the period separating module discoverer
	disc := &md.PeriodSeperatingModuleDiscoverer{}
	disc.Analyze(initialComp)

	// distribute the columns among modules; this includes also multiplicity columns
	moduleComp0 := distributed.GetFreshSegmentModuleComp(
		distributed.SegmentModuleInputs{
			InitialComp:         initialComp,
			Disc:                disc,
			ModuleName:          "module0",
			NumSegmentsInModule: numSegModule0,
		},
	)
	moduleComp1 := distributed.GetFreshSegmentModuleComp(distributed.SegmentModuleInputs{
		InitialComp:         initialComp,
		Disc:                disc,
		ModuleName:          "module1",
		NumSegmentsInModule: numSegModule1,
	})

	// distribute the query LogDerivativeSum among modules.
	inclusion.DistributeLogDerivativeSum(initialComp, moduleComp0, "module0", disc)
	inclusion.DistributeLogDerivativeSum(initialComp, moduleComp1, "module1", disc)

	// This compiles the log-derivative queries into global/local queries.
	wizard.ContinueCompilation(moduleComp0, logderiv.CompileLogDerivSum, dummy.Compile)
	wizard.ContinueCompilation(moduleComp1, logderiv.CompileLogDerivSum, dummy.Compile)

	/*
		logderiv.CompileLogDerivSum(moduleComp0)
		logderiv.CompileLogDerivSum(moduleComp1)

		// This adds a dummy compilation step to control that all passes
		dummy.CompileAtProverLvl(moduleComp0)
		dummy.CompileAtProverLvl(moduleComp1)
	*/

	// run the initial runtime
	initialRuntime := wizard.RunProver(initialComp, prover)

	// Compile and prove for module0
	for proverID := 0; proverID < numSegModule0; proverID++ {
		proof0 := wizard.Prove(moduleComp0, func(run *wizard.ProverRuntime) {
			run.ParentRuntime = initialRuntime
			// inputs for vertical splitting
			run.ProverID = proverID
		})
		valid := wizard.Verify(moduleComp0, proof0)
		require.NoError(t, valid)
	}

	// Compile and prove for module1
	for proverID := 0; proverID < numSegModule1; proverID++ {
		proof1 := wizard.Prove(moduleComp1, func(run *wizard.ProverRuntime) {
			run.ParentRuntime = initialRuntime
			// inputs for vertical splitting
			run.ProverID = proverID
		})
		valid1 := wizard.Verify(moduleComp1, proof1)
		require.NoError(t, valid1)
	}
}
