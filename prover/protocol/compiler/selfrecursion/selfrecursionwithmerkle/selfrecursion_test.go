package selfrecursionwithmerkle_test

import (
	"fmt"
	"testing"

	"github.com/consensys/accelerated-crypto-monorepo/crypto/ringsis"
	"github.com/consensys/accelerated-crypto-monorepo/maths/common/smartvectors"
	"github.com/consensys/accelerated-crypto-monorepo/maths/field"
	"github.com/consensys/accelerated-crypto-monorepo/protocol/coin"
	"github.com/consensys/accelerated-crypto-monorepo/protocol/compiler"
	"github.com/consensys/accelerated-crypto-monorepo/protocol/compiler/dummy"
	"github.com/consensys/accelerated-crypto-monorepo/protocol/compiler/mimc"
	"github.com/consensys/accelerated-crypto-monorepo/protocol/compiler/selfrecursion/selfrecursionwithmerkle"
	"github.com/consensys/accelerated-crypto-monorepo/protocol/compiler/vortex"
	"github.com/consensys/accelerated-crypto-monorepo/protocol/ifaces"
	"github.com/consensys/accelerated-crypto-monorepo/protocol/wizard"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
)

// returns a dummy column name
func dummyColName(i int) ifaces.ColID {
	return ifaces.ColIDf("A%v", i)
}

// returns a dummy column name
func dummyCoinName(i int) coin.Name {
	return coin.Namef("A%v", i)
}

// name of the evaluation query
const QNAME ifaces.QueryID = "EVAL"

// sis instances
var sisInstances = []ringsis.Params{
	{LogTwoBound_: 8, LogTwoDegree: 1},
	{LogTwoBound_: 4, LogTwoDegree: 2},
	{LogTwoBound_: 8, LogTwoDegree: 3},
	{LogTwoBound_: 8, LogTwoDegree: 6},
}

// testcase type
type TestCase struct {
	Numpoly, NumRound, PolSize, NumOpenCol int
	SisInstance                            ringsis.Params
}

// tests-cases for all tests
var testcases []TestCase = []TestCase{
	{Numpoly: 32, NumRound: 1, PolSize: 32, NumOpenCol: 16, SisInstance: sisInstances[0]},
	{Numpoly: 32, NumRound: 2, PolSize: 32, NumOpenCol: 16, SisInstance: sisInstances[0]},
	{Numpoly: 2, NumRound: 2, PolSize: 32, NumOpenCol: 2, SisInstance: sisInstances[0]},
	{Numpoly: 32, NumRound: 3, PolSize: 32, NumOpenCol: 16, SisInstance: sisInstances[0]},
	{Numpoly: 32, NumRound: 1, PolSize: 16, NumOpenCol: 16, SisInstance: sisInstances[1]},
	{Numpoly: 32, NumRound: 3, PolSize: 32, NumOpenCol: 16, SisInstance: sisInstances[2]},
	{Numpoly: 27, NumRound: 1, PolSize: 32, NumOpenCol: 8, SisInstance: sisInstances[0]},
	{Numpoly: 32, NumRound: 1, PolSize: 32, NumOpenCol: 16, SisInstance: sisInstances[3]},
	{Numpoly: 27, NumRound: 3, PolSize: 32, NumOpenCol: 16, SisInstance: sisInstances[3]},
	{Numpoly: 29, NumRound: 1, PolSize: 32, NumOpenCol: 16, SisInstance: sisInstances[3]},
}

// generate a testcase protocol with given parameters
func generateProtocol(tc TestCase) (define func(*wizard.Builder), prove func(*wizard.ProverRuntime)) {

	// the define function creates a dummy protocol
	// with only univariate evaluations
	define = func(b *wizard.Builder) {
		cols := make([]ifaces.Column, tc.Numpoly)
		numColPerRound := tc.Numpoly / tc.NumRound

		for round := 0; round < tc.NumRound; round++ {
			// determine which columns should be declared for each round
			start, stop := round*numColPerRound, (round+1)*numColPerRound
			if round == tc.NumRound-1 {
				stop = tc.Numpoly
			}

			for i := start; i < stop; i++ {
				cols[i] = b.RegisterCommit(dummyColName(i), tc.PolSize)
			}

			if round < tc.NumRound-1 {
				b.RegisterRandomCoin(dummyCoinName(round), coin.Field)
			}
		}

		b.UnivariateEval(QNAME, cols...)
	}

	// the prove function assignes the univariate evaluation
	// and the columns with random values
	prove = func(run *wizard.ProverRuntime) {
		// the evaluation point
		x := field.NewElement(42)
		ys := make([]field.Element, tc.Numpoly)
		numColPerRound := tc.Numpoly / tc.NumRound

		for round := 0; round < tc.NumRound; round++ {
			// determine which columns should be declared for each round
			start, stop := round*numColPerRound, (round+1)*numColPerRound
			if round == tc.NumRound-1 {
				stop = tc.Numpoly
			}

			// assigns each column to a random value and evalutes it at x
			for i := start; i < stop; i++ {
				v := smartvectors.Rand(tc.PolSize)
				run.AssignColumn(dummyColName(i), v)
				ys[i] = smartvectors.Interpolate(v, x)
			}

			if round < tc.NumRound-1 {
				_ = run.GetRandomCoinField(dummyCoinName(round))
			}
		}

		run.AssignUnivariate(QNAME, x, ys...)
	}
	return define, prove
}

func TestSelfRecursionRandom(t *testing.T) {

	logrus.SetLevel(logrus.FatalLevel)

	for _, tc := range testcases {
		t.Run(fmt.Sprintf("testcase-%++v", tc), func(subT *testing.T) {
			define, prove := generateProtocol(tc)

			comp := wizard.Compile(
				define,
				vortex.Compile(
					2,
					vortex.MerkleMode,
					vortex.ForceNumOpenedColumns(16),
					vortex.WithSISParams(&tc.SisInstance),
				),
				selfrecursionwithmerkle.SelfRecurse,
				dummy.Compile,
			)

			proof := wizard.Prove(
				comp,
				prove,
			)

			err := wizard.Verify(comp, proof)
			require.NoError(subT, err)
		})
	}
}

func TestSelfRecursionMultiLayered(t *testing.T) {

	logrus.SetLevel(logrus.FatalLevel)
	tc := TestCase{Numpoly: 32, NumRound: 3, PolSize: 32, NumOpenCol: 16, SisInstance: sisInstances[0]}
	t.Run(fmt.Sprintf("testcase-%++v", tc), func(subT *testing.T) {
		define, prove := generateProtocol(tc)

		comp := wizard.Compile(
			define,
			vortex.Compile(
				2,
				vortex.MerkleMode,
				vortex.ForceNumOpenedColumns(16),
				vortex.WithSISParams(&tc.SisInstance),
			),
			selfrecursionwithmerkle.SelfRecurse,
			mimc.CompileMiMC,
			compiler.Arcane(1<<8, 1<<10, false),
			vortex.Compile(
				2,
				vortex.MerkleMode,
				vortex.ForceNumOpenedColumns(16),
				vortex.WithSISParams(&tc.SisInstance),
			),
			selfrecursionwithmerkle.SelfRecurse,
			mimc.CompileMiMC,
			compiler.Arcane(1<<11, 1<<13, false),
			vortex.Compile(
				2,
				vortex.MerkleMode,
				vortex.ForceNumOpenedColumns(16),
				vortex.WithSISParams(&tc.SisInstance),
			),
			dummy.Compile,
		)

		proof := wizard.Prove(
			comp,
			prove,
		)

		err := wizard.Verify(comp, proof)
		require.NoError(subT, err)
	})
}
