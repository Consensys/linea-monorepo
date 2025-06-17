//go:build !fuzzlight

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

func TestCompiler(t *testing.T) {
	var (
		polSize    = 1 << 4
		nPols      = 16
		nPolsNoSIS = 3
		rows       = make([]ifaces.Column, nPols)
		// variables for multi-round
		nPolsMultiRound = []int{14, 8, 9, 16}
		numRounds       = 4
		rowsMultiRound  = make([][]ifaces.Column, numRounds)
		// variables for precomputed columns
		numPrecomputedsNoSIS = 4
		numPrecomputedsSIS   = 10
	)
	testCases := []struct {
		Explainer string
		Define    func(b *wizard.Builder)
		Prove     func(pr *wizard.ProverRuntime)
	}{
		{
			Explainer: "Vortex with a single round and using SIS",
			Define: func(b *wizard.Builder) {
				for i := range rows[:nPolsNoSIS] {
					rows[i] = b.RegisterCommit(ifaces.ColIDf("P_%v", i), polSize)
				}
				b.UnivariateEval("EVAL", rows[:nPolsNoSIS]...)
			},
			Prove: func(pr *wizard.ProverRuntime) {
				ys := make([]field.Element, nPolsNoSIS)
				x := field.NewElement(57) // the evaluation point

				// assign the rows with random polynomials and collect the ys
				for i, row := range rows[:nPolsNoSIS] {
					p := smartvectors.Rand(polSize)
					ys[i] = smartvectors.Interpolate(p, x)
					pr.AssignColumn(row.GetColID(), p)
				}

				pr.AssignUnivariate("EVAL", x, ys...)
			},
		},
		{
			Explainer: "Vortex with a single round but not SIS",
			Define: func(b *wizard.Builder) {
				for i := range rows {
					rows[i] = b.RegisterCommit(ifaces.ColIDf("P_%v", i), polSize)
				}
				b.UnivariateEval("EVAL", rows...)
			},
			Prove: func(pr *wizard.ProverRuntime) {
				ys := make([]field.Element, len(rows))
				x := field.NewElement(57) // the evaluation point

				// assign the rows with random polynomials and collect the ys
				for i, row := range rows {
					p := smartvectors.Rand(polSize)
					ys[i] = smartvectors.Interpolate(p, x)
					pr.AssignColumn(row.GetColID(), p)
				}

				pr.AssignUnivariate("EVAL", x, ys...)
			},
		},
		{
			Explainer: "Vortex with multiple round and only SIS, without precomputed columns",
			Define: func(b *wizard.Builder) {
				for round := 0; round < numRounds; round++ {
					var offsetIndex = 0
					// trigger the creation of a new round by declaring a dummy coin
					if round != 0 {
						_ = b.RegisterRandomCoin(coin.Namef("COIN_%v", round), coin.Field)
						// Compute the offsetIndex
						for i := 0; i < round; i++ {
							offsetIndex += nPols
						}
					}

					rowsMultiRound[round] = make([]ifaces.Column, nPols)
					for i := 0; i < nPols; i++ {
						rowsMultiRound[round][i] = b.RegisterCommit(ifaces.ColIDf("P_%v", offsetIndex+i), polSize)
					}
				}

				b.UnivariateEval("EVAL", utils.Join(rowsMultiRound...)...)
			},
			Prove: func(pr *wizard.ProverRuntime) {
				// Count the total number of polynomials
				numPolys := numRounds * nPols
				ys := make([]field.Element, numPolys)
				x := field.NewElement(57) // the evaluation point

				// assign the rows with random polynomials and collect the ys
				for round := 0; round < numRounds; round++ {
					var offsetIndex = 0
					if round != 0 {
						// let the prover know that it is free to go to the next
						// round by sampling the coin.
						_ = pr.GetRandomCoinField(coin.Namef("COIN_%v", round))
						// Compute the offsetIndex
						offsetIndex = nPols * round
					}

					for i, row := range rowsMultiRound[round] {
						p := smartvectors.Rand(polSize)
						ys[offsetIndex+i] = smartvectors.Interpolate(p, x)
						pr.AssignColumn(row.GetColID(), p)
					}
				}

				pr.AssignUnivariate("EVAL", x, ys...)
			},
		},
		{
			Explainer: "Vortex with multiple round and never SIS, without precomputed columns",
			Define: func(b *wizard.Builder) {
				for round := 0; round < numRounds; round++ {
					var offsetIndex = 0
					// trigger the creation of a new round by declaring a dummy coin
					if round != 0 {
						_ = b.RegisterRandomCoin(coin.Namef("COIN_%v", round), coin.Field)
						// Compute the offsetIndex
						for i := 0; i < round; i++ {
							offsetIndex += nPolsNoSIS
						}
					}

					rowsMultiRound[round] = make([]ifaces.Column, nPolsNoSIS)
					for i := 0; i < nPolsNoSIS; i++ {
						rowsMultiRound[round][i] = b.RegisterCommit(ifaces.ColIDf("P_%v", offsetIndex+i), polSize)
					}
				}

				b.UnivariateEval("EVAL", utils.Join(rowsMultiRound...)...)
			},
			Prove: func(pr *wizard.ProverRuntime) {
				// Count the total number of polynomials
				numPolys := numRounds * nPolsNoSIS
				ys := make([]field.Element, numPolys)
				x := field.NewElement(57) // the evaluation point

				// assign the rows with random polynomials and collect the ys
				for round := 0; round < numRounds; round++ {
					var offsetIndex = 0
					if round != 0 {
						// let the prover know that it is free to go to the next
						// round by sampling the coin.
						_ = pr.GetRandomCoinField(coin.Namef("COIN_%v", round))
						// Compute the offsetIndex
						offsetIndex = nPolsNoSIS * round
					}

					for i, row := range rowsMultiRound[round] {
						p := smartvectors.Rand(polSize)
						ys[offsetIndex+i] = smartvectors.Interpolate(p, x)
						pr.AssignColumn(row.GetColID(), p)
					}
				}

				pr.AssignUnivariate("EVAL", x, ys...)
			},
		},
		{
			Explainer: "Vortex with multiple rounds with both SIS and non-SIS rounds",
			Define: func(b *wizard.Builder) {
				for round := 0; round < numRounds; round++ {
					var offsetIndex = 0
					// trigger the creation of a new round by declaring a dummy coin
					if round != 0 {
						_ = b.RegisterRandomCoin(coin.Namef("COIN_%v", round), coin.Field)
						// Compute the offsetIndex
						for i := 0; i < round; i++ {
							offsetIndex += nPolsMultiRound[i]
						}
					}

					rowsMultiRound[round] = make([]ifaces.Column, nPolsMultiRound[round])
					for i := range nPolsMultiRound[round] {
						rowsMultiRound[round][i] = b.RegisterCommit(ifaces.ColIDf("P_%v", offsetIndex+i), polSize)
					}
				}

				b.UnivariateEval("EVAL", utils.Join(rowsMultiRound...)...)
			},
			Prove: func(pr *wizard.ProverRuntime) {
				// Count the total number of polynomials
				numPolys := 0
				for i := range nPolsMultiRound {
					numPolys += nPolsMultiRound[i]
				}
				ys := make([]field.Element, numPolys)
				x := field.NewElement(57) // the evaluation point

				// assign the rows with random polynomials and collect the ys
				for round := range rowsMultiRound {
					var offsetIndex = 0
					if round != 0 {
						// let the prover know that it is free to go to the next
						// round by sampling the coin.
						_ = pr.GetRandomCoinField(coin.Namef("COIN_%v", round))
						// Compute the offsetIndex
						for i := 0; i < round; i++ {
							offsetIndex += nPolsMultiRound[i]
						}
					}

					for i, row := range rowsMultiRound[round] {
						p := smartvectors.Rand(polSize)
						ys[offsetIndex+i] = smartvectors.Interpolate(p, x)
						pr.AssignColumn(row.GetColID(), p)
					}
				}

				pr.AssignUnivariate("EVAL", x, ys...)
			},
		},
		{
			Explainer: "Vortex with multiple rounds with both SIS and non-SIS, with precomputeds committed with no SIS",
			Define: func(b *wizard.Builder) {
				for round := 0; round < numRounds; round++ {
					var offsetIndex = 0
					// trigger the creation of a new round by declaring a dummy coin
					if round != 0 {
						_ = b.RegisterRandomCoin(coin.Namef("COIN_%v", round), coin.Field)
						// Compute the offsetIndex
						for i := 0; i < round; i++ {
							offsetIndex += nPolsMultiRound[i]
						}
					}

					rowsMultiRound[round] = make([]ifaces.Column, nPolsMultiRound[round])
					if round == 0 {
						for i := 0; i < numPrecomputedsNoSIS; i++ {
							p := smartvectors.Rand(polSize)
							rowsMultiRound[round][i] = b.RegisterPrecomputed(ifaces.ColIDf("PRE_COMP_%v", i), p)
						}
						for i := numPrecomputedsNoSIS; i < nPolsMultiRound[round]; i++ {
							rowsMultiRound[round][i] = b.RegisterCommit(ifaces.ColIDf("P_%v", i), polSize)
						}
						continue
					}
					for i := range nPolsMultiRound[round] {
						rowsMultiRound[round][i] = b.RegisterCommit(ifaces.ColIDf("P_%v", offsetIndex+i), polSize)
					}
				}

				b.UnivariateEval("EVAL", utils.Join(rowsMultiRound...)...)
			},
			Prove: func(pr *wizard.ProverRuntime) {
				// Count the total number of polynomials
				numPolys := 0
				for i := range nPolsMultiRound {
					numPolys += nPolsMultiRound[i]
				}
				ys := make([]field.Element, numPolys)
				x := field.NewElement(57) // the evaluation point

				// assign the rows with random polynomials and collect the ys
				for round := range rowsMultiRound {
					var offsetIndex = 0
					if round != 0 {
						// let the prover know that it is free to go to the next
						// round by sampling the coin.
						_ = pr.GetRandomCoinField(coin.Namef("COIN_%v", round))
						// Compute the offsetIndex
						for i := 0; i < round; i++ {
							offsetIndex += nPolsMultiRound[i]
						}
					}

					for i, row := range rowsMultiRound[round] {
						// For round 0 we need (numPolys - numPrecomputeds) polys, as the precomputed are
						// assigned in the define phase
						if i < numPrecomputedsNoSIS && round == 0 {
							p := pr.Spec.Precomputed.MustGet(row.GetColID())
							ys[i] = smartvectors.Interpolate(p, x)
							continue
						}
						p := smartvectors.Rand(polSize)
						ys[offsetIndex+i] = smartvectors.Interpolate(p, x)
						pr.AssignColumn(row.GetColID(), p)
					}
				}

				pr.AssignUnivariate("EVAL", x, ys...)
			},
		},
		{
			Explainer: "Vortex with multiple rounds with both SIS and non-SIS, with precomputeds committed with SIS",
			Define: func(b *wizard.Builder) {
				for round := 0; round < numRounds; round++ {
					var offsetIndex = 0
					// trigger the creation of a new round by declaring a dummy coin
					if round != 0 {
						_ = b.RegisterRandomCoin(coin.Namef("COIN_%v", round), coin.Field)
						// Compute the offsetIndex
						for i := 0; i < round; i++ {
							offsetIndex += nPolsMultiRound[i]
						}
					}

					rowsMultiRound[round] = make([]ifaces.Column, nPolsMultiRound[round])
					if round == 0 {
						for i := 0; i < numPrecomputedsSIS; i++ {
							p := smartvectors.Rand(polSize)
							rowsMultiRound[round][i] = b.RegisterPrecomputed(ifaces.ColIDf("PRE_COMP_%v", i), p)
						}
						for i := numPrecomputedsSIS; i < nPolsMultiRound[round]; i++ {
							rowsMultiRound[round][i] = b.RegisterCommit(ifaces.ColIDf("P_%v", i), polSize)
						}
						continue
					}
					for i := range nPolsMultiRound[round] {
						rowsMultiRound[round][i] = b.RegisterCommit(ifaces.ColIDf("P_%v", offsetIndex+i), polSize)
					}
				}

				b.UnivariateEval("EVAL", utils.Join(rowsMultiRound...)...)
			},
			Prove: func(pr *wizard.ProverRuntime) {
				// Count the total number of polynomials
				numPolys := 0
				for i := range nPolsMultiRound {
					numPolys += nPolsMultiRound[i]
				}
				ys := make([]field.Element, numPolys)
				x := field.NewElement(57) // the evaluation point

				// assign the rows with random polynomials and collect the ys
				for round := range rowsMultiRound {
					var offsetIndex = 0
					if round != 0 {
						// let the prover know that it is free to go to the next
						// round by sampling the coin.
						_ = pr.GetRandomCoinField(coin.Namef("COIN_%v", round))
						// Compute the offsetIndex
						for i := 0; i < round; i++ {
							offsetIndex += nPolsMultiRound[i]
						}
					}

					for i, row := range rowsMultiRound[round] {
						// For round 0 we need (numPolys - numPrecomputeds) polys, as the precomputed are
						// assigned in the define phase
						if i < numPrecomputedsSIS && round == 0 {
							p := pr.Spec.Precomputed.MustGet(row.GetColID())
							ys[i] = smartvectors.Interpolate(p, x)
							continue
						}
						p := smartvectors.Rand(polSize)
						ys[offsetIndex+i] = smartvectors.Interpolate(p, x)
						pr.AssignColumn(row.GetColID(), p)
					}
				}

				pr.AssignUnivariate("EVAL", x, ys...)
			},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.Explainer, func(t *testing.T) {
			logrus.Infof("Testing %s", tc.Explainer)
			compiled := wizard.Compile(tc.Define, vortex.Compile(4, vortex.WithOptionalSISHashingThreshold(9)))
			proof := wizard.Prove(compiled, tc.Prove)
			valid := wizard.Verify(compiled, proof)

			require.NoErrorf(t, valid, "the proof did not pass")
		})
	}
}
