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
		numSegModule = 2
	)

	//initialComp
	define := func(b *wizard.Builder) {

		var (
			col0 = b.CompiledIOP.InsertCommit(0, "module.col0", 8)
			col1 = b.CompiledIOP.InsertCommit(0, "module.col1", 8)
			// verifCol = verifiercol.NewConstantCol(field.NewElement(7), 8)
		)

		b.CompiledIOP.InsertGlobal(0, "global0",
			symbolic.Sub(col1, column.Shift(col0, 3)))

	}

	// initialProver
	prover := func(run *wizard.ProverRuntime) {
		run.AssignColumn("module.col0", smartvectors.ForTest(1, 0, 2, 1, 7, 1, 11, 2))
		run.AssignColumn("module.col1", smartvectors.ForTest(1, 7, 1, 11, 2, 1, 0, 2))
	}

	// initial compiledIOP is the parent to all the SegmentModuleComp objects.
	initialComp := wizard.Compile(define)

	// Initialize the period separating module discoverer
	disc := &md.PeriodSeperatingModuleDiscoverer{}
	disc.Analyze(initialComp)

	// distribute the columns among modules and segments; this includes also multiplicity columns
	// for all the segments from the same module, compiledIOP object is the same.
	moduleComp := distributed.GetFreshCompGL(
		distributed.SegmentModuleInputs{
			InitialComp:         initialComp,
			Disc:                disc,
			ModuleName:          "module",
			NumSegmentsInModule: numSegModule,
		},
	)

	// distribute the query LogDerivativeSum among modules.
	// The seed is used to generate randomness for each moduleComp.
	global.DistributeGlobal(global.DistributionInputs{
		ModuleComp:  moduleComp,
		InitialComp: initialComp,
		Disc:        disc,
		ModuleName:  "module",
		NumSegments: numSegModule,
	})

	// This compiles the log-derivative queries into global/local queries.
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
