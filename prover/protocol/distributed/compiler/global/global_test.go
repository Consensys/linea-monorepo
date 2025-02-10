package global_test

import (
	"testing"

	"github.com/consensys/linea-monorepo/prover/maths/common/smartvectors"
	"github.com/consensys/linea-monorepo/prover/protocol/column"
	"github.com/consensys/linea-monorepo/prover/protocol/compiler/dummy"
	"github.com/consensys/linea-monorepo/prover/protocol/distributed"
	"github.com/consensys/linea-monorepo/prover/protocol/distributed/compiler/global"
	md "github.com/consensys/linea-monorepo/prover/protocol/distributed/namebaseddiscoverer"
	"github.com/consensys/linea-monorepo/prover/protocol/wizard"
	"github.com/consensys/linea-monorepo/prover/symbolic"
	"github.com/stretchr/testify/require"
)

// It tests DistributedLogDerivSum.
func TestDistributedGlobal(t *testing.T) {
	const (
		numSegModule = 1
	)

	//initialComp
	define := func(b *wizard.Builder) {

		var (
			col0 = b.CompiledIOP.InsertCommit(0, "module.col0", 8)
			col1 = b.CompiledIOP.InsertCommit(0, "module.col1", 8)
			col2 = b.CompiledIOP.InsertCommit(0, "module.col2", 8)
			col3 = b.CompiledIOP.InsertCommit(0, "module.col3", 8)
		)

		b.CompiledIOP.InsertGlobal(0, "global0",
			symbolic.Sub(
				col1, column.Shift(col0, 3),
				symbolic.Mul(2, column.Shift(col2, 2)),
				symbolic.Neg(column.Shift(col3, 3)),
			),
		)

	}

	// initialProver
	prover := func(run *wizard.ProverRuntime) {
		run.AssignColumn("module.col0", smartvectors.ForTest(3, 0, 2, 1, 4, 1, 13, 0))
		run.AssignColumn("module.col1", smartvectors.ForTest(1, 7, 1, 11, 2, 1, 0, 2))
		run.AssignColumn("module.col2", smartvectors.ForTest(7, 0, 1, 3, 0, 4, 1, 0))
		run.AssignColumn("module.col3", smartvectors.ForTest(2, 14, 0, 2, 3, 0, 10, 0))

	}

	// initial compiledIOP is the parent to all the SegmentModuleComp objects.
	initialComp := wizard.Compile(define)

	// Initialize the period separating module discoverer
	disc := &md.PeriodSeperatingModuleDiscoverer{}
	disc.Analyze(initialComp)

	// distribute the columns among modules and segments.
	moduleComp := distributed.GetFreshCompGL(
		distributed.SegmentModuleInputs{
			InitialComp:         initialComp,
			Disc:                disc,
			ModuleName:          "module",
			NumSegmentsInModule: numSegModule,
		},
	)

	// distribute the query among segments.
	global.DistributeGlobal(global.DistributionInputs{
		ModuleComp:  moduleComp,
		InitialComp: initialComp,
		Disc:        disc,
		ModuleName:  "module",
		NumSegments: numSegModule,
	})

	// This dummy compiles the global/local queries of the segment.
	wizard.ContinueCompilation(moduleComp, dummy.Compile)

	// run the initial runtime
	initialRuntime := wizard.ProverOnlyFirstRound(initialComp, prover)

	// Compile and prove for module
	for proverID := 0; proverID < numSegModule; proverID++ {
		proof := wizard.Prove(moduleComp, func(run *wizard.ProverRuntime) {
			run.ParentRuntime = initialRuntime
			// inputs for vertical splitting of the witness
			run.ProverID = proverID
		})
		valid := wizard.Verify(moduleComp, proof)
		require.NoError(t, valid)
	}

}
