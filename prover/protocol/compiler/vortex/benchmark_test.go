package vortex_test

import (
	"fmt"
	"testing"

	"github.com/consensys/linea-monorepo/prover/maths/common/smartvectors"
	"github.com/consensys/linea-monorepo/prover/maths/field/fext"
	"github.com/consensys/linea-monorepo/prover/protocol/coin"
	"github.com/consensys/linea-monorepo/prover/protocol/compiler/vortex"
	"github.com/consensys/linea-monorepo/prover/protocol/ifaces"
	"github.com/consensys/linea-monorepo/prover/protocol/wizard"
	"github.com/consensys/linea-monorepo/prover/utils"
	"github.com/stretchr/testify/require"
)

func BenchmarkVortex(b *testing.B) {
	var (
		polySize  = []int{1 << 18, 1 << 19, 1 << 20}
		nPoly     = []int{1 << 11, 1 << 12, 1 << 13}
		numRounds = []int{1, 4, 8}
	)

	// Run benchmarks for all combinations of numRounds, nPoly, and polSize
	for i := range numRounds {
		for n := range nPoly {
			for p := range polySize {
				explainer := fmt.Sprintf("Running benchmark with numRounds=%v, nPoly=%v, PolySize=%v\n", numRounds[i], nPoly[n], polySize[p])
				b.Run(explainer, func(b *testing.B) {
					benchmarkVortex(b, polySize[p], nPoly[n], numRounds[i])
				})

			}
		}
	}

}

func benchmarkVortex(b *testing.B, polSize int, nPoly int, numRounds int) {

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
		ys := make([]fext.Element, nPoly*numRounds)
		x := fext.NewFromUintBase(57) // the evaluation point

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
				ys[offsetIndex+i] = smartvectors.EvaluateBasePolyLagrange(p, x)
				pr.AssignColumn(row.GetColID(), p)
			}
		}

		// 	pr.AssignUnivariate("EVAL", x, ys...)
		pr.AssignUnivariateExt("EVAL", x, ys...)
	}

	compiled := wizard.Compile(Define, vortex.Compile(2,
		vortex.ForceNumOpenedColumns(256),
		vortex.WithOptionalSISHashingThreshold(512)))

	b.ResetTimer()
	var proof wizard.Proof
	for i := 0; i < b.N; i++ {
		proof = wizard.Prove(compiled, Prove)
	}
	valid := wizard.Verify(compiled, proof)

	require.NoErrorf(b, valid, "the proof did not pass")
}
