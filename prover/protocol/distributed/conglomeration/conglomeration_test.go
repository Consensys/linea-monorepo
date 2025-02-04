package conglomeration_test

import (
	"testing"

	"github.com/consensys/linea-monorepo/prover/crypto/ringsis"
	"github.com/consensys/linea-monorepo/prover/maths/common/smartvectors"
	"github.com/consensys/linea-monorepo/prover/maths/field"
	"github.com/consensys/linea-monorepo/prover/protocol/accessors"
	"github.com/consensys/linea-monorepo/prover/protocol/coin"
	"github.com/consensys/linea-monorepo/prover/protocol/compiler"
	"github.com/consensys/linea-monorepo/prover/protocol/compiler/dummy"
	"github.com/consensys/linea-monorepo/prover/protocol/compiler/mimc"
	"github.com/consensys/linea-monorepo/prover/protocol/compiler/selfrecursion"
	"github.com/consensys/linea-monorepo/prover/protocol/compiler/vortex"
	"github.com/consensys/linea-monorepo/prover/protocol/distributed/conglomeration"
	"github.com/consensys/linea-monorepo/prover/protocol/distributed/constants"
	"github.com/consensys/linea-monorepo/prover/protocol/ifaces"
	"github.com/consensys/linea-monorepo/prover/protocol/query"
	"github.com/consensys/linea-monorepo/prover/protocol/wizard"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
)

type (
	defineFuncType func(*wizard.Builder)
	proverFuncType func(int) func(*wizard.ProverRuntime)
	compilerSuite  []func(*wizard.CompiledIOP)
)

var (
	commonSisParams  = &ringsis.Params{LogTwoBound: 16, LogTwoDegree: 1}
	commonVortexStep = vortex.Compile(
		2, // the inverse-rate of the RS code
		vortex.WithSISParams(commonSisParams),
		vortex.ForceNumOpenedColumns(2),
		vortex.PremarkAsSelfRecursed(),
	)
	vortexOnlyCompilationSuite = []func(*wizard.CompiledIOP){
		commonVortexStep,
	}
	arcaneCompilationSuite = []func(*wizard.CompiledIOP){
		compiler.Arcane(1<<8, 1<<10, false),
		commonVortexStep,
	}
	arcaneAndSelfRecCompilationSuite = []func(*wizard.CompiledIOP){
		compiler.Arcane(1, 1<<10, false),
		commonVortexStep,
		selfrecursion.SelfRecurse,
		mimc.CompileMiMC,
		compiler.Arcane(1, 1<<10, false),
		commonVortexStep,
	}
	// arcaneFullRecSelfRecCompilationSuite = []func(*wizard.CompiledIOP){
	// 	compiler.Arcane(1<<8, 1<<10, false),
	// 	commonVortexStep,
	// 	selfrecursion.SelfRecurse,
	// 	mimc.CompileMiMC,
	// 	compiler.Arcane(1<<8, 1<<10, false),
	// 	commonVortexStep,
	// 	fullrecursion.FullRecursion(true),
	// 	mimc.CompileMiMC,
	// 	compiler.Arcane(1<<8, 1<<10, false),
	// 	commonVortexStep,
	// }
)

type conglomerationTestCase struct {
	define   defineFuncType
	prove    proverFuncType
	numProof int
	suite    compilerSuite
}

func TestConglomerationPureVortexSingleRound(t *testing.T) {

	var (
		numCol   = 16
		numRow   = 16
		numProof = 2
		a        []ifaces.Column
		u        query.UnivariateEval
	)

	define := func(builder *wizard.Builder) {
		for i := 0; i < numCol; i++ {
			a = append(a, builder.RegisterCommit(ifaces.ColIDf("a-%v", i), numRow))
		}
		u = builder.CompiledIOP.InsertUnivariate(0, "u", a)
		builder.InsertPublicInput(constants.GrandProductPublicInput, accessors.NewConstant(field.NewElement(1)))
		builder.InsertPublicInput(constants.GrandSumPublicInput, accessors.NewConstant(field.NewElement(0)))
		builder.InsertPublicInput(constants.LogDerivativeSumPublicInput, accessors.NewConstant(field.NewElement(0)))
	}

	prover := func(k int) func(run *wizard.ProverRuntime) {
		return func(run *wizard.ProverRuntime) {
			ys := make([]field.Element, 0, len(a))
			for i := range a {
				y := field.NewElement(uint64(i + k))
				run.AssignColumn(a[i].GetColID(), smartvectors.NewConstant(y, numRow))
				ys = append(ys, y)
			}
			run.AssignUnivariate(u.QueryID, field.NewElement(0), ys...)
		}
	}

	runConglomerationTestCase(t, conglomerationTestCase{
		define:   define,
		prove:    prover,
		numProof: numProof,
		suite:    vortexOnlyCompilationSuite,
	})
}

func TestConglomerationPureVortexMultiRound(t *testing.T) {

	var (
		numRound = 4
		numCol   = 4
		numRow   = 16
		numProof = 16
		a        []ifaces.Column
	)

	define := func(builder *wizard.Builder) {

		allCols := make([]ifaces.Column, 0, numCol*numRound)

		for round := 0; round < numRound; round++ {

			if round > 0 {
				_ = builder.InsertCoin(round, coin.Namef("c-%v", round), coin.Field)
			}

			roundCols := make([]ifaces.Column, 0, numCol)

			for i := 0; i < numCol; i++ {
				newCol := builder.InsertCommit(round, ifaces.ColIDf("a-%v-%v", round, i), numRow)
				roundCols = append(roundCols, newCol)
			}

			if round > 0 {
				builder.SubProvers.AppendToInner(round, func(run *wizard.ProverRuntime) {
					for i := range roundCols {
						x := field.NewElement(uint64(round*numCol + i))
						run.AssignColumn(roundCols[i].GetColID(), smartvectors.NewConstant(x, numRow))
					}

					if round == numRound-1 {

					}
				})
			}

			if round == 0 {
				a = roundCols
			}

			allCols = append(allCols, roundCols...)
		}

		u := builder.CompiledIOP.InsertUnivariate(numRound-1, "u", allCols)

		builder.CompiledIOP.SubProvers.AppendToInner(numRound-1, func(run *wizard.ProverRuntime) {
			ys := make([]field.Element, 0, len(allCols))
			for _, col := range allCols {
				ys = append(ys, col.GetColAssignmentAt(run, 0))
			}
			run.AssignUnivariate(u.QueryID, field.NewElement(0), ys...)
		})

		builder.InsertPublicInput(constants.GrandProductPublicInput, accessors.NewConstant(field.NewElement(1)))
		builder.InsertPublicInput(constants.GrandSumPublicInput, accessors.NewConstant(field.NewElement(0)))
		builder.InsertPublicInput(constants.LogDerivativeSumPublicInput, accessors.NewConstant(field.NewElement(0)))
	}

	prover := func(k int) func(run *wizard.ProverRuntime) {
		return func(run *wizard.ProverRuntime) {
			for i := range a {
				y := field.NewElement(uint64(i + k))
				run.AssignColumn(a[i].GetColID(), smartvectors.NewConstant(y, numRow))
			}
		}
	}

	runConglomerationTestCase(t, conglomerationTestCase{
		define:   define,
		prove:    prover,
		numProof: numProof,
		suite:    vortexOnlyCompilationSuite,
	})
}

func TestConglomerationLookup(t *testing.T) {

	logrus.SetLevel(logrus.FatalLevel)

	tcs := []struct {
		name  string
		suite compilerSuite
	}{
		{
			name:  "arcane",
			suite: arcaneCompilationSuite,
		},
		{
			name:  "arcane/self-recursion",
			suite: arcaneAndSelfRecCompilationSuite,
		},
		// {
		// 	name:  "arcane/full-recursion/self-recursion",
		// 	suite: arcaneFullRecSelfRecCompilationSuite,
		// },
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {

			var (
				numCol   = 16
				numRow   = 16
				numProof = 4
				a        []ifaces.Column
			)

			define := func(builder *wizard.Builder) {
				for i := 0; i < numCol; i++ {
					a = append(a, builder.RegisterCommit(ifaces.ColIDf("a-%v", i), numRow))
					builder.Range(ifaces.QueryIDf("range-%v", i), a[i], 1<<8)
				}

				builder.InsertPublicInput(constants.GrandProductPublicInput, accessors.NewConstant(field.NewElement(1)))
				builder.InsertPublicInput(constants.GrandSumPublicInput, accessors.NewConstant(field.NewElement(0)))
				builder.InsertPublicInput(constants.LogDerivativeSumPublicInput, accessors.NewConstant(field.NewElement(0)))
			}

			prover := func(k int) func(run *wizard.ProverRuntime) {
				return func(run *wizard.ProverRuntime) {
					for i := range a {
						y := field.NewElement(uint64(i + k))
						run.AssignColumn(a[i].GetColID(), smartvectors.NewConstant(y, numRow))
					}
				}
			}

			runConglomerationTestCase(t, conglomerationTestCase{
				define:   define,
				prove:    prover,
				numProof: numProof,
				suite:    tc.suite,
			})
		})
	}
}

func runConglomerationTestCase(t *testing.T, tc conglomerationTestCase) {

	var (
		numProof             = tc.numProof
		tmpl                 = wizard.Compile(wizard.DefineFunc(tc.define), tc.suite...)
		congDef, ctxsPHolder = conglomeration.ConglomerateDefineFunc(tmpl, numProof)
		cong                 = wizard.Compile(congDef, dummy.CompileAtProverLvl)
		ctxs                 = *ctxsPHolder
		lastRound            = ctxs[0].LastRound
	)

	witnesses := make([]conglomeration.Witness, numProof)
	for i := range witnesses {
		runtime := wizard.RunProverUntilRound(tmpl, tc.prove(i), lastRound+1)
		witnesses[i] = conglomeration.ExtractWitness(runtime)
	}

	proof := wizard.Prove(cong, conglomeration.ProveConglomeration(ctxs, witnesses))
	err := wizard.Verify(cong, proof)

	require.NoError(t, err)
}
