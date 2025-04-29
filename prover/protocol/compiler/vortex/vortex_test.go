//go:build !fuzzlight

package vortex_test

import (
	"testing"

	"github.com/consensys/linea-monorepo/prover/crypto/ringsis"
	"github.com/consensys/linea-monorepo/prover/maths/common/smartvectors"
	"github.com/consensys/linea-monorepo/prover/maths/field"
	"github.com/consensys/linea-monorepo/prover/protocol/coin"
	"github.com/consensys/linea-monorepo/prover/protocol/compiler/vortex"
	"github.com/consensys/linea-monorepo/prover/protocol/ifaces"
	"github.com/consensys/linea-monorepo/prover/protocol/wizard"
	"github.com/consensys/linea-monorepo/prover/utils"
	"github.com/stretchr/testify/require"
)

func TestVortexSingleRoundMerkle(t *testing.T) {

	polSize := 1 << 4
	nPols := 16
	rows := make([]ifaces.Column, nPols)

	define := func(b *wizard.Builder) {
		for i := range rows {
			rows[i] = b.RegisterCommit(ifaces.ColIDf("P_%v", i), polSize)
		}
		b.UnivariateEval("EVAL", rows...)
	}

	prove := func(pr *wizard.ProverRuntime) {
		ys := make([]field.Element, len(rows))
		x := field.NewElement(57) // the evaluation point

		// assign the rows with random polynomials and collect the ys
		for i, row := range rows {
			p := smartvectors.Rand(polSize)
			ys[i] = smartvectors.Interpolate(p, x)
			pr.AssignColumn(row.GetColID(), p)
		}

		pr.AssignUnivariate("EVAL", x, ys...)
	}

	compiled := wizard.Compile(define, vortex.Compile(4))
	proof := wizard.Prove(compiled, prove)
	valid := wizard.Verify(compiled, proof)

	require.NoErrorf(t, valid, "the proof did not pass")
}

func TestVortexMultiRoundMerkle(t *testing.T) {

	polSize := 1 << 4
	nPols := 15
	numRounds := 4
	rows := make([][]ifaces.Column, numRounds)

	define := func(b *wizard.Builder) {
		for round := 0; round < numRounds; round++ {
			// trigger the creation of a new round by declaring a dummy coin
			if round != 0 {
				_ = b.RegisterRandomCoin(coin.Namef("COIN_%v", round), coin.Field)
			}

			rows[round] = make([]ifaces.Column, nPols)
			for i := range rows[round] {
				rows[round][i] = b.RegisterCommit(ifaces.ColIDf("P_%v", round*nPols+i), polSize)
			}
		}

		b.UnivariateEval("EVAL", utils.Join(rows...)...)
	}

	prove := func(pr *wizard.ProverRuntime) {
		ys := make([]field.Element, len(rows)*len(rows[0]))
		x := field.NewElement(57) // the evaluation point

		// assign the rows with random polynomials and collect the ys
		for round := range rows {
			// let the prover know that it is free to go to the next
			// round by sampling the coin.
			if round != 0 {
				_ = pr.GetRandomCoinField(coin.Namef("COIN_%v", round))
			}

			for i, row := range rows[round] {
				p := smartvectors.Rand(polSize)
				ys[round*nPols+i] = smartvectors.Interpolate(p, x)
				pr.AssignColumn(row.GetColID(), p)
			}
		}

		pr.AssignUnivariate("EVAL", x, ys...)
	}

	compiled := wizard.Compile(define, vortex.Compile(4))
	proof := wizard.Prove(compiled, prove)
	valid := wizard.Verify(compiled, proof)

	require.NoErrorf(t, valid, "the proof did not pass")
}

func TestVortexOneFullOneDryMerkle(t *testing.T) {

	polSize := 1 << 4
	nPols := 17
	rows := make([]ifaces.Column, nPols+1)

	define := func(b *wizard.Builder) {
		for i := 0; i < nPols; i++ {
			rows[i] = b.RegisterCommit(ifaces.ColIDf("P_%v", i), polSize)
		}

		// Sample a coin to specify that the following goes into a new round
		_ = b.RegisterRandomCoin("DUMMY_COIN", coin.Field)

		// registers a new commitment in what will be considered a dry round.
		// namely, a round where too few commitment are declared for us to
		// deem it worthwhile to send a commitment to the verifier instead of
		// directly sending the commitment.
		rows[nPols] = b.RegisterCommit(ifaces.ColIDf("P_%v", nPols), polSize)

		b.UnivariateEval("EVAL", rows...)
	}

	prove := func(pr *wizard.ProverRuntime) {
		ys := make([]field.Element, len(rows))
		x := field.NewElement(57) // the evaluation point

		for i := 0; i < nPols; i++ {
			p := smartvectors.Rand(polSize)
			ys[i] = smartvectors.Interpolate(p, x)
			pr.AssignColumn(rows[i].GetColID(), p)
		}

		// trigger the go to next round by querying the dummy random coin
		_ = pr.GetRandomCoinField("DUMMY_COIN")

		// assign the dry-round-column
		p := smartvectors.Rand(polSize)
		ys[nPols] = smartvectors.Interpolate(p, x)
		pr.AssignColumn(rows[nPols].GetColID(), p)

		// finally assigns the query
		pr.AssignUnivariate("EVAL", x, ys...)
	}

	compiled := wizard.Compile(
		define,
		vortex.Compile(
			4,
			vortex.WithSISParams(&ringsis.Params{LogTwoBound: 8, LogTwoDegree: 6}),
		),
	)
	proof := wizard.Prove(compiled, prove)
	valid := wizard.Verify(compiled, proof)

	require.NoErrorf(t, valid, "the proof did not pass")
}

func TestVortexSingleRoundMerkleNoSis(t *testing.T) {

	polSize := 1 << 4
	nPols := 16
	rows := make([]ifaces.Column, nPols)

	define := func(b *wizard.Builder) {
		for i := range rows {
			rows[i] = b.RegisterCommit(ifaces.ColIDf("P_%v", i), polSize)
		}
		b.UnivariateEval("EVAL", rows...)
	}

	prove := func(pr *wizard.ProverRuntime) {
		ys := make([]field.Element, len(rows))
		x := field.NewElement(57) // the evaluation point

		// assign the rows with random polynomials and collect the ys
		for i, row := range rows {
			p := smartvectors.Rand(polSize)
			ys[i] = smartvectors.Interpolate(p, x)
			pr.AssignColumn(row.GetColID(), p)
		}

		pr.AssignUnivariate("EVAL", x, ys...)
	}

	compiled := wizard.Compile(define, vortex.Compile(4))
	proof := wizard.Prove(compiled, prove)
	valid := wizard.Verify(compiled, proof)

	require.NoErrorf(t, valid, "the proof did not pass")
}

func TestVortexMultiRoundMerkleNoSis(t *testing.T) {

	polSize := 1 << 4
	nPols := 15
	numRounds := 4
	rows := make([][]ifaces.Column, numRounds)

	define := func(b *wizard.Builder) {
		for round := 0; round < numRounds; round++ {
			// trigger the creation of a new round by declaring a dummy coin
			if round != 0 {
				_ = b.RegisterRandomCoin(coin.Namef("COIN_%v", round), coin.Field)
			}

			rows[round] = make([]ifaces.Column, nPols)
			for i := range rows[round] {
				rows[round][i] = b.RegisterCommit(ifaces.ColIDf("P_%v", round*nPols+i), polSize)
			}
		}

		b.UnivariateEval("EVAL", utils.Join(rows...)...)
	}

	prove := func(pr *wizard.ProverRuntime) {
		ys := make([]field.Element, len(rows)*len(rows[0]))
		x := field.NewElement(57) // the evaluation point

		// assign the rows with random polynomials and collect the ys
		for round := range rows {
			// let the prover know that it is free to go to the next
			// round by sampling the coin.
			if round != 0 {
				_ = pr.GetRandomCoinField(coin.Namef("COIN_%v", round))
			}

			for i, row := range rows[round] {
				p := smartvectors.Rand(polSize)
				ys[round*nPols+i] = smartvectors.Interpolate(p, x)
				pr.AssignColumn(row.GetColID(), p)
			}
		}

		pr.AssignUnivariate("EVAL", x, ys...)
	}

	compiled := wizard.Compile(define, vortex.Compile(4))
	proof := wizard.Prove(compiled, prove)
	valid := wizard.Verify(compiled, proof)

	require.NoErrorf(t, valid, "the proof did not pass")
}

func TestVortexOneFullOneDryMerkleNoSis(t *testing.T) {

	polSize := 1 << 4
	nPols := 17
	rows := make([]ifaces.Column, nPols+1)

	define := func(b *wizard.Builder) {
		for i := 0; i < nPols; i++ {
			rows[i] = b.RegisterCommit(ifaces.ColIDf("P_%v", i), polSize)
		}

		// Sample a coin to specify that the following goes into a new round
		_ = b.RegisterRandomCoin("DUMMY_COIN", coin.Field)

		// registers a new commitment in what will be considered a dry round.
		// namely, a round where too few commitment are declared for us to
		// deem it worthwhile to send a commitment to the verifier instead of
		// directly sending the commitment.
		rows[nPols] = b.RegisterCommit(ifaces.ColIDf("P_%v", nPols), polSize)

		b.UnivariateEval("EVAL", rows...)
	}

	prove := func(pr *wizard.ProverRuntime) {
		ys := make([]field.Element, len(rows))
		x := field.NewElement(57) // the evaluation point

		for i := 0; i < nPols; i++ {
			p := smartvectors.Rand(polSize)
			ys[i] = smartvectors.Interpolate(p, x)
			pr.AssignColumn(rows[i].GetColID(), p)
		}

		// trigger the go to next round by querying the dummy random coin
		_ = pr.GetRandomCoinField("DUMMY_COIN")

		// assign the dry-round-column
		p := smartvectors.Rand(polSize)
		ys[nPols] = smartvectors.Interpolate(p, x)
		pr.AssignColumn(rows[nPols].GetColID(), p)

		// finally assigns the query
		pr.AssignUnivariate("EVAL", x, ys...)
	}

	compiled := wizard.Compile(
		define,
		vortex.Compile(4),
	)
	proof := wizard.Prove(compiled, prove)
	valid := wizard.Verify(compiled, proof)

	require.NoErrorf(t, valid, "the proof did not pass")
}
