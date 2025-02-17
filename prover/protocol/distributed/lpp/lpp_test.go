package lpp_test

import (
	"testing"

	"github.com/consensys/linea-monorepo/prover/maths/common/smartvectors"
	"github.com/consensys/linea-monorepo/prover/maths/field"
	"github.com/consensys/linea-monorepo/prover/protocol/distributed"
	"github.com/consensys/linea-monorepo/prover/protocol/distributed/compiler/inclusion"
	"github.com/consensys/linea-monorepo/prover/protocol/distributed/lpp"
	md "github.com/consensys/linea-monorepo/prover/protocol/distributed/namebaseddiscoverer"
	"github.com/consensys/linea-monorepo/prover/protocol/ifaces"
	"github.com/consensys/linea-monorepo/prover/protocol/wizard"
	"github.com/stretchr/testify/require"
)

// It tests DistributedLogDerivSum.
func TestSeedGeneration(t *testing.T) {
	const (
		numSegModule0 = 2
		numSegModule1 = 2
		numSegModule2 = 1
	)

	var (
		coinLookup0Gamma field.Element
		coinLookup1Gamma field.Element
		coinLookup2Gamma field.Element
		coinLookup2Alpha field.Element
	)

	//initialComp
	define := func(b *wizard.Builder) {

		var (
			// columns from module0
			col01 = b.CompiledIOP.InsertCommit(0, "module0.col1", 4)
			col02 = b.CompiledIOP.InsertCommit(0, "module0.col2", 8)

			// columns from module1
			col10 = b.CompiledIOP.InsertCommit(0, "module1.col0", 8)
			col11 = b.CompiledIOP.InsertCommit(0, "module1.col1", 16)
			col12 = b.CompiledIOP.InsertCommit(0, "module1.col2", 4)
			col13 = b.CompiledIOP.InsertCommit(0, "module1.col3", 4)
			col14 = b.CompiledIOP.InsertCommit(0, "module1.col4", 16)
			col15 = b.CompiledIOP.InsertCommit(0, "module1.col5", 16)

			//  columns from module2
			col20 = b.CompiledIOP.InsertCommit(0, "module2.col0", 4)
			col21 = b.CompiledIOP.InsertCommit(0, "module2.col1", 4)
			col22 = b.CompiledIOP.InsertCommit(0, "module2.col2", 4)
		)

		// inclusion query: S \subset T , S in module0, T in module1.
		b.CompiledIOP.InsertInclusion(0, "lookup0",
			[]ifaces.Column{col10}, []ifaces.Column{col01})

		// conditional inclusion query : S\subset T, S in module1,T in module0.
		b.CompiledIOP.InsertInclusionConditionalOnIncluded(0, "lookup1",
			[]ifaces.Column{col02}, []ifaces.Column{col12}, col13)

		// double conditional inclusion query (multi-column): S in module2, T in module1.
		b.CompiledIOP.InsertInclusionDoubleConditional(0, "lookup2",
			[]ifaces.Column{col11, col14}, []ifaces.Column{col20, col21}, col15, col22)
	}

	// initialProver
	prover := func(run *wizard.ProverRuntime) {
		run.AssignColumn("module0.col1", smartvectors.ForTest(1, 2, 1, 2))
		run.AssignColumn("module0.col2", smartvectors.ForTest(3, 4, 9, 1, 0, 4, 8, 7))

		run.AssignColumn("module1.col0", smartvectors.ForTest(1, 1, 2, 1, 1, 1, 1, 2))
		run.AssignColumn("module1.col1", smartvectors.ForTest(3, 2, 1, 0, 0, 8, 11, 5, 5, 7, 2, 1, 0, 0, 9, 10))
		run.AssignColumn("module1.col2", smartvectors.ForTest(9, 1, 0, 9))
		run.AssignColumn("module1.col3", smartvectors.ForTest(0, 1, 1, 1))
		run.AssignColumn("module1.col4", smartvectors.ForTest(2, 3, 3, 3, 6, 8, 0, 0, 0, 8, 8, 0, 3, 4, 5, 11))
		run.AssignColumn("module1.col5", smartvectors.ForTest(1, 1, 1, 1, 0, 0, 0, 1, 1, 1, 1, 1, 1, 0, 1, 1))

		run.AssignColumn("module2.col0", smartvectors.ForTest(3, 10, 2, 7))
		run.AssignColumn("module2.col1", smartvectors.ForTest(2, 11, 8, 8))
		run.AssignColumn("module2.col2", smartvectors.ForTest(1, 1, 0, 1))
	}

	// initial compiledIOP is the parent to LPPComp and all the SegmentModuleComp objects.
	initialComp := wizard.Compile(define)
	// apply the LPP relevant compilers and generate the seed for initialComp
	lppComp := lpp.CompileLPPAndGetSeed(initialComp, distributed.IntoLogDerivativeSum)

	// Initialize the period separating module discoverer
	disc := &md.PeriodSeperatingModuleDiscoverer{}
	disc.Analyze(initialComp)

	// distribute the columns among modules and segments; this includes also multiplicity columns
	// for all the segments from the same module, compiledIOP object is the same.
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

	moduleComp2 := distributed.GetFreshSegmentModuleComp(distributed.SegmentModuleInputs{
		InitialComp:         initialComp,
		Disc:                disc,
		ModuleName:          "module2",
		NumSegmentsInModule: numSegModule2,
	})

	// distribute the query LogDerivativeSum among modules.
	// The seed is used to generate randomness for each moduleComp.
	inclusion.DistributeLogDerivativeSum(initialComp, moduleComp0, "module0", disc, numSegModule0)
	inclusion.DistributeLogDerivativeSum(initialComp, moduleComp1, "module1", disc, numSegModule1)
	inclusion.DistributeLogDerivativeSum(initialComp, moduleComp2, "module2", disc, numSegModule2)

	// run the initial runtime
	initialRuntime := wizard.ProverOnlyFirstRound(initialComp, prover)

	// compile and verify for lpp-Prover
	lppProof := wizard.Prove(lppComp, func(run *wizard.ProverRuntime) {
		run.ParentRuntime = initialRuntime
	})
	valid := wizard.Verify(lppComp, lppProof)
	require.NoError(t, valid)

	// check that all the modules see the same randomness for the same coins.
	for proverID := 0; proverID < numSegModule0; proverID++ {
		// get prover run time for module0
		runtime0 := wizard.RunProver(moduleComp0, func(run *wizard.ProverRuntime) {
			run.ParentRuntime = initialRuntime
			// inputs for vertical splitting of the witness
			run.ProverID = proverID
		})

		// get and compar the coins with the other segments/modules
		coin1 := runtime0.Coins.MustGet("TABLE_module0.col2_LOGDERIVATIVE_GAMMA_FieldFromSeed").(field.Element)
		coin0 := runtime0.Coins.MustGet("TABLE_module1.col0_LOGDERIVATIVE_GAMMA_FieldFromSeed").(field.Element)
		if coinLookup1Gamma.IsZero() {
			coinLookup0Gamma = coin0
			coinLookup1Gamma = coin1
		} else {
			require.Equal(t, coin0, coinLookup0Gamma)
			require.Equal(t, coin1, coinLookup1Gamma)
		}
	}

	// get prove  runtime for module1
	for proverID := 0; proverID < numSegModule1; proverID++ {
		runtime1 := wizard.RunProver(moduleComp1, func(run *wizard.ProverRuntime) {
			run.ParentRuntime = initialRuntime
			// inputs for vertical splitting of the witness
			run.ProverID = proverID
		})

		// get and compar the coins with the other segments/modules
		coin1 := runtime1.Coins.MustGet("TABLE_module0.col2_LOGDERIVATIVE_GAMMA_FieldFromSeed").(field.Element)
		coin0 := runtime1.Coins.MustGet("TABLE_module1.col0_LOGDERIVATIVE_GAMMA_FieldFromSeed").(field.Element)
		coin2Gamma := runtime1.Coins.MustGet("TABLE_module1.col5,module1.col1,module1.col4_LOGDERIVATIVE_GAMMA_FieldFromSeed").(field.Element)
		coin2Alpha := runtime1.Coins.MustGet("TABLE_module1.col5,module1.col1,module1.col4_LOGDERIVATIVE_ALPHA_FieldFromSeed").(field.Element)

		require.Equal(t, coin1, coinLookup1Gamma)
		require.Equal(t, coin0, coinLookup0Gamma)

		if coinLookup2Gamma.IsZero() {
			coinLookup2Gamma = coin2Gamma
			coinLookup2Alpha = coin2Alpha
		} else {
			require.Equal(t, coin2Gamma, coinLookup2Gamma)
			require.Equal(t, coin2Alpha, coinLookup2Alpha)
		}
	}

	// get prove  run time for module2
	for proverID := 0; proverID < numSegModule2; proverID++ {
		runtime2 := wizard.RunProver(moduleComp2, func(run *wizard.ProverRuntime) {
			run.ParentRuntime = initialRuntime
			// inputs for vertical splitting of the witness
			run.ProverID = proverID
		})

		// get and compare the coins with the other segments/modules
		coin2Gamma := runtime2.Coins.MustGet("TABLE_module1.col5,module1.col1,module1.col4_LOGDERIVATIVE_GAMMA_FieldFromSeed").(field.Element)
		coin2Alpha := runtime2.Coins.MustGet("TABLE_module1.col5,module1.col1,module1.col4_LOGDERIVATIVE_ALPHA_FieldFromSeed").(field.Element)

		require.Equal(t, coin2Gamma, coinLookup2Gamma)
		require.Equal(t, coin2Alpha, coinLookup2Alpha)
	}
}
