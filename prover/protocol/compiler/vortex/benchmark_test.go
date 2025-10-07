package vortex_test

import (
	"testing"

	"github.com/consensys/linea-monorepo/prover/maths/common/smartvectors"
	"github.com/consensys/linea-monorepo/prover/maths/field"
	"github.com/consensys/linea-monorepo/prover/protocol/coin"
	"github.com/consensys/linea-monorepo/prover/protocol/compiler/vortex"
	"github.com/consensys/linea-monorepo/prover/protocol/ifaces"
	"github.com/consensys/linea-monorepo/prover/protocol/wizard"
	"github.com/consensys/linea-monorepo/prover/utils"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
)

func BenchmarkVortexForBenchmark(b *testing.B) {
	var (
		polySize  = []int{1 << 10, 1 << 14, 1 << 17}
		nPoly     = []int{1 << 10, 1 << 15, 1 << 20}
		numRounds = []int{1, 4, 8, 16}
	)

	// Run benchmarks for all combinations of numRounds, nPoly, and polSize
	for i := range numRounds {
		for n := range nPoly {
			for p := range polySize {
				benchmarkVortex(b, polySize[p], nPoly[n], numRounds[i])
			}
		}
	}

}

func benchmarkVortex(b *testing.B, polSize int, nPoly int, numRounds int) {

	logrus.Infof(" ------------ Benchmarking Vortex with numRounds=%v, nPoly=%v, PolySize=%v,-------------- ", numRounds, nPoly, polSize)
	var (
		rowsMultiRound = make([][]ifaces.Column, numRounds)
	)

	Define := func(b *wizard.Builder) {
		for round := 0; round < numRounds; round++ {
			var offsetIndex = 0
			// trigger the creation of a new round by declaring a dummy coin
			if round != 0 {
				_ = b.RegisterRandomCoin(coin.Namef("COIN_%v", round), coin.Field)
				// Compute the offsetIndex
				for i := 0; i < round; i++ {
					offsetIndex += nPoly
				}
			}

			rowsMultiRound[round] = make([]ifaces.Column, nPoly)
			for i := range nPoly {
				rowsMultiRound[round][i] = b.RegisterCommit(ifaces.ColIDf("P_%v", offsetIndex+i), polSize)
			}
		}

		b.UnivariateEval("EVAL", utils.Join(rowsMultiRound...)...)
	}

	Prove := func(pr *wizard.ProverRuntime) {
		ys := make([]field.Element, nPoly*numRounds)
		x := field.NewElement(57) // the evaluation point

		// assign the rows with random polynomials and collect the ys
		for round := range rowsMultiRound {
			var offsetIndex = 0
			if round != 0 {
				// let the prover know that it is free to go to the next
				// round by sampling the coin.
				_ = pr.GetRandomCoinField(coin.Namef("COIN_%v", round))
				// Compute the offsetIndex
				offsetIndex += nPoly * round
			}

			for i, row := range rowsMultiRound[round] {
				p := smartvectors.Rand(polSize)
				ys[offsetIndex+i] = smartvectors.Interpolate(p, x)
				pr.AssignColumn(row.GetColID(), p)
			}
		}

		pr.AssignUnivariate("EVAL", x, ys...)
	}

	compiled := wizard.Compile(Define, vortex.Compile(4,
		vortex.ForceNumOpenedColumns(10),
		vortex.WithOptionalSISHashingThreshold(512)))

	b.ResetTimer()
	var proof wizard.Proof
	for i := 0; i < b.N; i++ {
		proof = wizard.Prove(compiled, Prove)
	}
	valid := wizard.Verify(compiled, proof)

	require.NoErrorf(b, valid, "the proof did not pass")
}
