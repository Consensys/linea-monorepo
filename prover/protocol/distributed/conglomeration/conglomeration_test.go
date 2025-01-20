package conglomeration

import (
	"testing"

	"github.com/consensys/linea-monorepo/prover/crypto/ringsis"
	"github.com/consensys/linea-monorepo/prover/maths/common/smartvectors"
	"github.com/consensys/linea-monorepo/prover/maths/field"
	"github.com/consensys/linea-monorepo/prover/protocol/coin"
	"github.com/consensys/linea-monorepo/prover/protocol/compiler/dummy"
	"github.com/consensys/linea-monorepo/prover/protocol/compiler/vortex"
	"github.com/consensys/linea-monorepo/prover/protocol/ifaces"
	"github.com/consensys/linea-monorepo/prover/protocol/query"
	"github.com/consensys/linea-monorepo/prover/protocol/wizard"
	"github.com/stretchr/testify/require"
)

type (
	defineFuncType func(*wizard.Builder)
	proverFuncType func(int) func(*wizard.ProverRuntime)
)

type conglomerationTestCase struct {
	define   defineFuncType
	prove    proverFuncType
	numProof int
}

func TestConglomerationPureVortexSingleRound(t *testing.T) {

	var (
		numCol   = 16
		numRow   = 16
		numProof = 16
		a        []ifaces.Column
		u        query.UnivariateEval
	)

	define := func(builder *wizard.Builder) {
		for i := 0; i < numCol; i++ {
			a = append(a, builder.RegisterCommit(ifaces.ColIDf("a-%v", i), numRow))
		}
		u = builder.CompiledIOP.InsertUnivariate(0, "u", a)
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

	}

	prover := func(k int) func(run *wizard.ProverRuntime) {
		return func(run *wizard.ProverRuntime) {
			ys := make([]field.Element, 0, len(a))
			for i := range a {
				y := field.NewElement(uint64(i + k))
				run.AssignColumn(a[i].GetColID(), smartvectors.NewConstant(y, numRow))
				ys = append(ys, y)
			}
		}
	}

	runConglomerationTestCase(t, conglomerationTestCase{
		define:   define,
		prove:    prover,
		numProof: numProof,
	})
}

func runConglomerationTestCase(t *testing.T, tc conglomerationTestCase) {

	var (
		sisParams            = &ringsis.Params{LogTwoBound: 16, LogTwoDegree: 1}
		vortexCompFunc       = vortex.Compile(2, vortex.WithSISParams(sisParams), vortex.ForceNumOpenedColumns(2), vortex.PremarkAsSelfRecursed())
		numProof             = tc.numProof
		tmpl                 = wizard.Compile(wizard.DefineFunc(tc.define), vortexCompFunc)
		congDef, ctxsPHolder = ConglomerateDefineFunc(tmpl, numProof)
		cong                 = wizard.Compile(congDef, dummy.CompileAtProverLvl)
		ctxs                 = *ctxsPHolder
		lastRound            = ctxs[0].LastRound
	)

	witnesses := make([]Witness, numProof)
	for i := range witnesses {
		runtime := wizard.RunProverUntilRound(tmpl, tc.prove(i), lastRound+1)
		witnesses[i] = ExtractWitness(runtime)
	}

	proof := wizard.Prove(cong, ProveConglomeration(ctxs, witnesses))
	err := wizard.Verify(cong, proof)

	require.NoError(t, err)
}
