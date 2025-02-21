package global_test

import (
	"errors"
	"testing"

	"github.com/consensys/linea-monorepo/prover/maths/common/smartvectors"
	"github.com/consensys/linea-monorepo/prover/maths/field"
	"github.com/consensys/linea-monorepo/prover/protocol/column"
	"github.com/consensys/linea-monorepo/prover/protocol/column/verifiercol"
	"github.com/consensys/linea-monorepo/prover/protocol/compiler/dummy"
	"github.com/consensys/linea-monorepo/prover/protocol/distributed/compiler/global"
	"github.com/consensys/linea-monorepo/prover/protocol/distributed/constants"
	md "github.com/consensys/linea-monorepo/prover/protocol/distributed/namebaseddiscoverer"
	segcomp "github.com/consensys/linea-monorepo/prover/protocol/distributed/segment_comp.go"
	"github.com/consensys/linea-monorepo/prover/protocol/wizard"
	"github.com/consensys/linea-monorepo/prover/symbolic"
	"github.com/stretchr/testify/require"
)

// It tests DistributedLogDerivSum.
func TestDistributedGlobal(t *testing.T) {
	const (
		numSegModule = 2
	)

	var (
		allVerfiers = []wizard.Runtime{}
	)

	//initialComp
	define := func(b *wizard.Builder) {

		var (
			col0 = b.CompiledIOP.InsertCommit(0, "module.col0", 8)
			col1 = b.CompiledIOP.InsertCommit(0, "module.col1", 8)
			col2 = b.CompiledIOP.InsertCommit(0, "module.col2", 8)
			col3 = b.CompiledIOP.InsertCommit(0, "module.col3", 8)

			fibonacci = b.CompiledIOP.InsertCommit(0, "module.fibo", 16)
			verifCol  = verifiercol.NewConstantCol(field.One(), 16)
			// verifCol1 = column.Shift(verifCol, 2)
		)

		b.CompiledIOP.InsertGlobal(0, "global0",
			symbolic.Sub(
				col1, column.Shift(col0, 3),
				symbolic.Mul(2, column.Shift(col2, 2)),
				symbolic.Neg(column.Shift(col3, 3)),
			),
			// check boundaries
			true,
		)

		b.CompiledIOP.InsertGlobal(0, "fibonacci",
			symbolic.Sub(
				symbolic.Mul(fibonacci, verifCol),
				column.Shift(fibonacci, -1),
				column.Shift(fibonacci, -2)),
		)

	}

	// initialProver
	prover := func(run *wizard.ProverRuntime) {
		run.AssignColumn("module.col0", smartvectors.ForTest(3, 0, 2, 1, 4, 1, 13, 0))
		run.AssignColumn("module.col1", smartvectors.ForTest(1, 7, 1, 11, 2, 1, 0, 2))
		run.AssignColumn("module.col2", smartvectors.ForTest(7, 0, 1, 3, 0, 4, 1, 0))
		run.AssignColumn("module.col3", smartvectors.ForTest(2, 14, 0, 2, 3, 0, 10, 0))

		run.AssignColumn("module.fibo", smartvectors.ForTest(1, 1, 2, 3, 5, 8, 13, 21, 34, 55, 89, 144, 233, 377, 610, 987))

	}

	// initial compiledIOP is the parent to all the SegmentModuleComp objects.
	initialComp := wizard.Compile(define)
	var segID int

	// Initialize the module discoverer
	disc := md.QueryBasedDiscoverer{
		SimpleDiscoverer: &md.PeriodSeperatingModuleDiscoverer{},
	}
	disc.Analyze(initialComp)

	// distribute the columns among modules and segments.
	moduleComp := segcomp.GetFreshGLComp(
		segcomp.SegmentInputs{
			InitialComp:         initialComp,
			Disc:                disc,
			ModuleName:          "module",
			NumSegmentsInModule: numSegModule,
			SegID:               segID,
		},
	)

	// distribute the query among segments.
	global.DistributeGlobal(global.DistributionInputs{
		ModuleComp:  moduleComp,
		InitialComp: initialComp,
		Disc:        disc.SimpleDiscoverer,
		ModuleName:  "module",
		NumSegments: numSegModule,
		SegID:       segID,
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
		vRunTime, valid := wizard.VerifyWithRuntime(moduleComp, proof)
		require.NoError(t, valid)

		allVerfiers = append(allVerfiers, vRunTime)
	}
	// apply the crosse checks over the public inputs.
	require.NoError(t, checkConsistency(allVerfiers))

}

func checkConsistency(runs []wizard.Runtime) error {

	for i := range runs {

		var (
			hashProvider     = runs[i].GetPublicInput(constants.GlobalProviderPublicInput)
			hashNextReceiver = runs[(i+1)%len(runs)].GetPublicInput(constants.GlobalReceiverPublicInput)
		)

		if hashProvider != hashNextReceiver {
			return errors.New("the provider and the next receiver have different values")
		}
	}

	return nil
}
