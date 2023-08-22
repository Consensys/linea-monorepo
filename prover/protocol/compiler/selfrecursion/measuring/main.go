package main

import (
	"github.com/consensys/accelerated-crypto-monorepo/backend/files"
	"github.com/consensys/accelerated-crypto-monorepo/protocol/coin"
	"github.com/consensys/accelerated-crypto-monorepo/protocol/compiler/logdata"
	"github.com/consensys/accelerated-crypto-monorepo/protocol/compiler/selfrecursion"
	"github.com/consensys/accelerated-crypto-monorepo/protocol/compiler/specialqueries"
	"github.com/consensys/accelerated-crypto-monorepo/protocol/compiler/vortex"
	"github.com/consensys/accelerated-crypto-monorepo/protocol/ifaces"
	"github.com/consensys/accelerated-crypto-monorepo/protocol/wizard"
)

func main() {

	const NUM_COL int = 1 << 15
	const NUM_ROW int = 1 << 16
	const NUM_ROUND int = 5
	const BLOWUP int = 2

	define := func(b *wizard.Builder) {

		rows := make([]ifaces.Column, NUM_ROW)
		chunkSize := NUM_ROW / NUM_ROUND

		for roundID := 0; roundID < NUM_ROUND; roundID++ {
			// Sample a dummy random coin at the beginning of the round
			// to create a new round
			if roundID > 0 {
				b.RegisterRandomCoin(coin.Namef("COIN%v", roundID), coin.Field)
			}

			// For each round, we assign only some parts of the columns
			start, stop := roundID*chunkSize, (roundID+1)*chunkSize
			if roundID == NUM_ROUND-1 {
				// Also add the leftover columns in the last round
				stop = NUM_ROW
			}

			for rowID := start; rowID < stop; rowID++ {
				rows[rowID] = b.RegisterCommit(ifaces.ColIDf("P%v", rowID), NUM_COL)
			}
		}

		// And create the query
		b.UnivariateEval("QUERY", rows...)
	}

	f := files.MustOverwrite("./stats.csv")
	defer f.Close()

	wizard.Compile(define,
		vortex.Compile(BLOWUP),
		selfrecursion.SelfRecurse,
		specialqueries.RangeProof,
		specialqueries.CompileFixedPermutations,
		specialqueries.LogDerivativeLookupCompiler,
		specialqueries.CompilePermutations,
		specialqueries.CompileInnerProduct,
		logdata.GenCSV(f),
	)

}
