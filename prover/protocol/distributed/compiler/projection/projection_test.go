package dist_projection_test

import (
	"errors"
	"testing"

	"github.com/consensys/linea-monorepo/prover/maths/common/smartvectors"
	"github.com/consensys/linea-monorepo/prover/maths/field"
	"github.com/consensys/linea-monorepo/prover/protocol/compiler/distributedprojection"
	"github.com/consensys/linea-monorepo/prover/protocol/compiler/dummy"
	"github.com/consensys/linea-monorepo/prover/protocol/distributed"
	dist_projection "github.com/consensys/linea-monorepo/prover/protocol/distributed/compiler/projection"
	"github.com/consensys/linea-monorepo/prover/protocol/distributed/constants"
	"github.com/consensys/linea-monorepo/prover/protocol/distributed/lpp"
	md "github.com/consensys/linea-monorepo/prover/protocol/distributed/namebaseddiscoverer"
	"github.com/consensys/linea-monorepo/prover/protocol/ifaces"
	"github.com/consensys/linea-monorepo/prover/protocol/query"
	"github.com/consensys/linea-monorepo/prover/protocol/wizard"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
)

func TestDistributeProjection(t *testing.T) {
	const (
		numSegModuleA = 2
		numSegModuleB = 2
		numSegModuleC = 2
	)
	var (
		allVerfiers                                                   = []wizard.Runtime{}
		moduleAName                                                   = "moduleA"
		moduleBName                                                   = "moduleB"
		moduleCName                                                   = "moduleC"
		flagSizeA                                                     = 512
		flagSizeB                                                     = 256
		flagA, flagB, columnA, columnB, flagC, columnC                ifaces.Column
		colA0, colA1, colA2, colB0, colB1, colB2, colC0, colC1, colC2 ifaces.Column
	)
	testcases := []struct {
		Name              string
		DefineFunc        func(builder *wizard.Builder)
		InitialProverFunc func(run *wizard.ProverRuntime)
	}{
		{
			DefineFunc: func(builder *wizard.Builder) {
				flagA = builder.RegisterCommit(ifaces.ColID("moduleA.FilterA"), flagSizeA)
				flagB = builder.RegisterCommit(ifaces.ColID("moduleB.FliterB"), flagSizeB)
				flagC = builder.RegisterCommit(ifaces.ColID("moduleC.FliterC"), flagSizeB)
				columnA = builder.RegisterCommit(ifaces.ColID("moduleA.ColumnA"), flagSizeA)
				columnB = builder.RegisterCommit(ifaces.ColID("moduleB.ColumnB"), flagSizeB)
				columnC = builder.RegisterCommit(ifaces.ColID("moduleC.ColumnC"), flagSizeB)
				_ = builder.InsertProjection("ProjectionTest-A-B",
					query.ProjectionInput{ColumnA: []ifaces.Column{columnA}, ColumnB: []ifaces.Column{columnB}, FilterA: flagA, FilterB: flagB})
				_ = builder.InsertProjection("ProjectionTest-C-A",
					query.ProjectionInput{ColumnA: []ifaces.Column{columnC}, ColumnB: []ifaces.Column{columnA}, FilterA: flagC, FilterB: flagA})

			},
			InitialProverFunc: func(run *wizard.ProverRuntime) {
				// assign filters and columns
				var (
					flagAWit   = make([]field.Element, flagSizeA)
					columnAWit = make([]field.Element, flagSizeA)
					flagBWit   = make([]field.Element, flagSizeB)
					columnBWit = make([]field.Element, flagSizeB)
					flagCWit   = make([]field.Element, flagSizeB)
					columnCWit = make([]field.Element, flagSizeB)
				)
				for i := 0; i < 10; i++ {
					flagAWit[i] = field.One()
					columnAWit[i] = field.NewElement(uint64(i))
				}
				for i := flagSizeB - 10; i < flagSizeB; i++ {
					flagBWit[i] = field.One()
					flagCWit[i] = field.One()
					columnBWit[i] = field.NewElement(uint64(i - (flagSizeB - 10)))
					columnCWit[i] = field.NewElement(uint64(i - (flagSizeB - 10)))
				}
				run.AssignColumn(flagA.GetColID(), smartvectors.RightZeroPadded(flagAWit, flagSizeA))
				run.AssignColumn(flagB.GetColID(), smartvectors.RightZeroPadded(flagBWit, flagSizeB))
				run.AssignColumn(flagC.GetColID(), smartvectors.RightZeroPadded(flagCWit, flagSizeB))
				run.AssignColumn(columnB.GetColID(), smartvectors.RightZeroPadded(columnBWit, flagSizeB))
				run.AssignColumn(columnA.GetColID(), smartvectors.RightZeroPadded(columnAWit, flagSizeA))
				run.AssignColumn(columnC.GetColID(), smartvectors.RightZeroPadded(columnCWit, flagSizeB))
			},
		},
		{
			Name: "distribute-projection-multiple_projections-multi-columns",
			DefineFunc: func(builder *wizard.Builder) {
				flagA = builder.RegisterCommit(ifaces.ColID("moduleA.FilterA"), flagSizeA)
				flagB = builder.RegisterCommit(ifaces.ColID("moduleB.FliterB"), flagSizeB)
				flagC = builder.RegisterCommit(ifaces.ColID("moduleC.FliterB"), flagSizeB)
				colA0 = builder.RegisterCommit(ifaces.ColID("moduleA.ColumnA0"), flagSizeA)
				colA1 = builder.RegisterCommit(ifaces.ColID("moduleA.ColumnA1"), flagSizeA)
				colA2 = builder.RegisterCommit(ifaces.ColID("moduleA.ColumnA2"), flagSizeA)
				colB0 = builder.RegisterCommit(ifaces.ColID("moduleB.ColumnB0"), flagSizeB)
				colB1 = builder.RegisterCommit(ifaces.ColID("moduleB.ColumnB1"), flagSizeB)
				colB2 = builder.RegisterCommit(ifaces.ColID("moduleB.ColumnB2"), flagSizeB)
				colC0 = builder.RegisterCommit(ifaces.ColID("moduleC.ColumnC0"), flagSizeB)
				colC1 = builder.RegisterCommit(ifaces.ColID("moduleC.ColumnC1"), flagSizeB)
				colC2 = builder.RegisterCommit(ifaces.ColID("moduleC.ColumnC2"), flagSizeB)
				_ = builder.InsertProjection("ProjectionTest-A-B-Multicolum",
					query.ProjectionInput{ColumnA: []ifaces.Column{colA0, colA1, colA2}, ColumnB: []ifaces.Column{colB0, colB1, colB2}, FilterA: flagA, FilterB: flagB})
				_ = builder.InsertProjection("ProjectionTest-C-A-Multicolumn",
					query.ProjectionInput{ColumnA: []ifaces.Column{colC0, colC1, colC2}, ColumnB: []ifaces.Column{colA0, colA1, colA2}, FilterA: flagC, FilterB: flagA})

			},
			InitialProverFunc: func(run *wizard.ProverRuntime) {
				// assign filters and columns
				var (
					flagAWit = make([]field.Element, flagSizeA)
					flagBWit = make([]field.Element, flagSizeB)
					flagCWit = make([]field.Element, flagSizeB)
					colA0Wit = make([]field.Element, flagSizeA)
					colA1Wit = make([]field.Element, flagSizeA)
					colA2Wit = make([]field.Element, flagSizeA)
					colB0Wit = make([]field.Element, flagSizeB)
					colB1Wit = make([]field.Element, flagSizeB)
					colB2Wit = make([]field.Element, flagSizeB)
					colC0Wit = make([]field.Element, flagSizeB)
					colC1Wit = make([]field.Element, flagSizeB)
					colC2Wit = make([]field.Element, flagSizeB)
				)
				for i := 0; i < 10; i++ {
					flagAWit[i] = field.One()
					colA0Wit[i] = field.NewElement(uint64(i))
					colA1Wit[i] = field.NewElement(uint64(i + 1))
					colA2Wit[i] = field.NewElement(uint64(i + 2))
				}
				for i := flagSizeB - 10; i < flagSizeB; i++ {
					flagBWit[i] = field.One()
					flagCWit[i] = field.One()
					colB0Wit[i] = field.NewElement(uint64(i - (flagSizeB - 10)))
					colC0Wit[i] = field.NewElement(uint64(i - (flagSizeB - 10)))
					colB1Wit[i] = field.NewElement(uint64(i + 1 - (flagSizeB - 10)))
					colC1Wit[i] = field.NewElement(uint64(i + 1 - (flagSizeB - 10)))
					colB2Wit[i] = field.NewElement(uint64(i + 2 - (flagSizeB - 10)))
					colC2Wit[i] = field.NewElement(uint64(i + 2 - (flagSizeB - 10)))
				}
				run.AssignColumn(flagA.GetColID(), smartvectors.RightZeroPadded(flagAWit, flagSizeA))
				run.AssignColumn(flagB.GetColID(), smartvectors.RightZeroPadded(flagBWit, flagSizeB))
				run.AssignColumn(flagC.GetColID(), smartvectors.RightZeroPadded(flagCWit, flagSizeB))
				run.AssignColumn(colA0.GetColID(), smartvectors.RightZeroPadded(colA0Wit, flagSizeA))
				run.AssignColumn(colA1.GetColID(), smartvectors.RightZeroPadded(colA1Wit, flagSizeA))
				run.AssignColumn(colA2.GetColID(), smartvectors.RightZeroPadded(colA2Wit, flagSizeA))
				run.AssignColumn(colB0.GetColID(), smartvectors.RightZeroPadded(colB0Wit, flagSizeB))
				run.AssignColumn(colB1.GetColID(), smartvectors.RightZeroPadded(colB1Wit, flagSizeB))
				run.AssignColumn(colB2.GetColID(), smartvectors.RightZeroPadded(colB2Wit, flagSizeB))
				run.AssignColumn(colC0.GetColID(), smartvectors.RightZeroPadded(colC0Wit, flagSizeB))
				run.AssignColumn(colC1.GetColID(), smartvectors.RightZeroPadded(colC1Wit, flagSizeB))
				run.AssignColumn(colC2.GetColID(), smartvectors.RightZeroPadded(colC2Wit, flagSizeB))
			},
		},
	}
	for _, tc := range testcases {

		t.Run(tc.Name, func(t *testing.T) {

			// initialComp is defined according to the define function provided by the
			// test-case.
			initialComp := wizard.Compile(tc.DefineFunc)

			// apply the LPP relevant compilers and generate the seed for initialComp
			lppComp := lpp.CompileLPPAndGetSeed(initialComp)

			// Initialize the period separating module discoverer
			disc := &md.PeriodSeperatingModuleDiscoverer{}
			disc.Analyze(initialComp)

			// distribute the columns among modules and segments; this includes also multiplicity columns
			// for all the segments from the same module, compiledIOP object is the same.
			moduleCompA := distributed.GetFreshSegmentModuleComp(
				distributed.SegmentModuleInputs{
					InitialComp:         initialComp,
					Disc:                disc,
					ModuleName:          moduleAName,
					NumSegmentsInModule: numSegModuleA,
				},
			)

			moduleCompB := distributed.GetFreshSegmentModuleComp(distributed.SegmentModuleInputs{
				InitialComp:         initialComp,
				Disc:                disc,
				ModuleName:          moduleBName,
				NumSegmentsInModule: numSegModuleB,
			})

			moduleCompC := distributed.GetFreshSegmentModuleComp(distributed.SegmentModuleInputs{
				InitialComp:         initialComp,
				Disc:                disc,
				ModuleName:          moduleCName,
				NumSegmentsInModule: numSegModuleC,
			})

			// distribute the query LogDerivativeSum among modules.
			// The seed is used to generate randomness for each moduleComp.
			dist_projection.NewDistributeProjectionCtx(moduleAName, initialComp, moduleCompA, disc, numSegModuleA)
			dist_projection.NewDistributeProjectionCtx(moduleBName, initialComp, moduleCompB, disc, numSegModuleB)
			dist_projection.NewDistributeProjectionCtx(moduleCName, initialComp, moduleCompC, disc, numSegModuleC)

			// This compiles the log-derivative queries into global/local queries.
			wizard.ContinueCompilation(moduleCompA, distributedprojection.CompileDistributedProjection, dummy.Compile)
			wizard.ContinueCompilation(moduleCompB, distributedprojection.CompileDistributedProjection, dummy.Compile)
			wizard.ContinueCompilation(moduleCompC, distributedprojection.CompileDistributedProjection, dummy.Compile)

			// run the initial runtime
			initialRuntime := wizard.ProverOnlyFirstRound(initialComp, tc.InitialProverFunc)

			// compile and verify for lpp-Prover
			lppProof := wizard.Prove(lppComp, func(run *wizard.ProverRuntime) {
				run.ParentRuntime = initialRuntime
			})
			lppVerifierRuntime, valid := wizard.VerifyWithRuntime(lppComp, lppProof)
			require.NoError(t, valid)

			// Compile and prove for moduleA
			for proverID := 0; proverID < numSegModuleA; proverID++ {
				proofA := wizard.Prove(moduleCompA, func(run *wizard.ProverRuntime) {
					run.ParentRuntime = initialRuntime
					// inputs for vertical splitting of the witness
					run.ProverID = proverID
				})
				runtimeA, validA := wizard.VerifyWithRuntime(moduleCompA, proofA, lppVerifierRuntime)
				require.NoError(t, validA)

				allVerfiers = append(allVerfiers, runtimeA)

			}

			// Compile and prove for moduleB
			for proverID := 0; proverID < numSegModuleB; proverID++ {
				proofB := wizard.Prove(moduleCompB, func(run *wizard.ProverRuntime) {
					run.ParentRuntime = initialRuntime
					// inputs for vertical splitting of the witness
					run.ProverID = proverID
				})
				runtimeB, validB := wizard.VerifyWithRuntime(moduleCompB, proofB, lppVerifierRuntime)
				require.NoError(t, validB)

				allVerfiers = append(allVerfiers, runtimeB)

			}

			// Compile and prove for moduleC
			for proverID := 0; proverID < numSegModuleC; proverID++ {
				proofC := wizard.Prove(moduleCompC, func(run *wizard.ProverRuntime) {
					run.ParentRuntime = initialRuntime
					// inputs for vertical splitting of the witness
					run.ProverID = proverID
				})
				runtimeC, validC := wizard.VerifyWithRuntime(moduleCompC, proofC, lppVerifierRuntime)
				require.NoError(t, validC)

				allVerfiers = append(allVerfiers, runtimeC)
			}

			// apply the crosse checks over the public inputs.
			require.NoError(t, checkConsistency(allVerfiers))

		})

	}

}

func checkConsistency(runs []wizard.Runtime) error {

	var res = field.Zero()
	for i, run := range runs {
		distProjectionParams := run.GetPublicInput(constants.DistributedProjectionPublicInput)
		logrus.Printf("successfully retrieved public input for %v, param = %v", i, distProjectionParams)
		res.Add(&res, &distProjectionParams)
	}

	if !res.IsZero() {
		return errors.New("the distributed projection sums do not cancel each others")
	}

	return nil
}
