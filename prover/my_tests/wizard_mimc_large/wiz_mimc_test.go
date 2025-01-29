// Copyright 2020-2024 Consensys Software Inc.
// Licensed under the Apache License, Version 2.0. See the LICENSE file for details.

package wizmimc

import (
	"fmt"
	"github.com/consensys/linea-monorepo/prover/crypto/mimc"
	"github.com/consensys/linea-monorepo/prover/crypto/ringsis"
	"github.com/consensys/linea-monorepo/prover/maths/common/smartvectors"
	"github.com/consensys/linea-monorepo/prover/maths/field"
	"github.com/consensys/linea-monorepo/prover/protocol/compiler"
	mimcComp "github.com/consensys/linea-monorepo/prover/protocol/compiler/mimc"
	"github.com/consensys/linea-monorepo/prover/protocol/compiler/selfrecursion"
	"github.com/consensys/linea-monorepo/prover/protocol/compiler/vortex"
	"github.com/consensys/linea-monorepo/prover/protocol/ifaces"
	"github.com/consensys/linea-monorepo/prover/protocol/wizard"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

const (
	size = 8
)

func outputProverFunc() func(run *wizard.ProverRuntime) {
	prover := func(run *wizard.ProverRuntime) {
		data := make([]field.Element, size)
		for _, elem := range data {
			elem.SetRandom()
		}
		colPre1 := smartvectors.NewRegular(data)

		hash := make([]field.Element, size)
		inter := make([]field.Element, size)

		state := field.Zero() // the initial state is zero
		for i := 0; i < len(hash); i++ {
			// first, hash the HI part of the fetched log message
			mimcBlock := colPre1.Get(i)
			// debugString.WriteString(mimcBlock.)
			state = mimc.BlockCompression(state, mimcBlock)
			hash[i].Set(&state)

			// the data in hashSecond is used to initialize the next initial state, stored in the inter column
			if i+1 < len(hash) {
				inter[i+1] = hash[i]
			}
		}

		run.AssignColumn("PRE_COL_1", colPre1)
		run.AssignColumn("COL_MIMC_1", smartvectors.NewRegular(hash))
		run.AssignColumn("COL_INTER_1", smartvectors.NewRegular(inter))

		//run.GetRandomCoinField(coin.Namef("Coin"))

	}
	return prover
}

// In this example we show how to use PLONK with KZG commitments. The circuit that is
// showed here is the same as in ../exponentiate.

func BenchmarkWizardMiMC(bench *testing.B) {
	bench.StopTimer()

	define := func(b *wizard.Builder) {
		preimages := b.RegisterCommit("PRE_COL_1", size)
		hash := b.RegisterCommit("COL_MIMC_1", size)
		inter := b.RegisterCommit("COL_INTER_1", size)
		b.CompiledIOP.InsertMiMC(0, ifaces.QueryIDf("%s_%s", "TEST", "MIMC_CONSTRAINT"), preimages, inter, hash)

	}

	//vortex.Compile(2, vortex.WithDryThreshold(0))
	//comp := wizard.Compile(define, globalcs.Compile, vortex.Compile(2, vortex.WithDryThreshold(0)))
	//comp := wizard.Compile(define, compiler.Arcane(16, 64))
	// comp := wizard.Compile(define, vortex.Compile(2, vortex.WithDryThreshold(0)))
	sisInstance := ringsis.Params{LogTwoBound: 16, LogTwoDegree: 6}
	comp := wizard.Compile(define, mimcComp.CompileMiMC,
		compiler.Arcane(size, size),
		vortex.Compile(
			2,
			vortex.WithSISParams(&sisInstance),
		),
		selfrecursion.SelfRecurse,
	)

	// START BENCHMARK
	prover := outputProverFunc()
	proof := wizard.Prove(comp, prover)

	// copy wizard proof into a new one
	secondProof := wizard.Proof{
		Messages:      proof.Messages,
		QueriesParams: proof.QueriesParams,
	}

	bench.StartTimer()
	timeStart := time.Now()
	checkErr := wizard.Verify(comp, secondProof)
	assert.NoErrorf(bench, checkErr, "INVALID proof")
	timeEnd := time.Now()
	bench.StopTimer()

	proofSize := 0
	allValues := proof.Messages.InnerMap()
	for _, value := range allValues {
		proofSize += value.Len() * field.Bytes
	}
	bench.ReportMetric(float64(proofSize), "Wizard-proof-size")

	fmt.Println("Wizard proof size ", proofSize)
	customTime := timeEnd.Sub(timeStart).Nanoseconds()
	fmt.Println("Custom timings raw", customTime)
	fmt.Println("Benchmark iterations", bench.N)
	bench.ReportMetric(float64(customTime), "Custom-Timing")

	/*bench.StartTimer()
	proof2 := wizard.Prove(comp, prover)
	checkErr2 := wizard.Verify(comp, proof2)
	assert.NoErrorf(bench, checkErr2, "INVALID proof")
	bench.StopTimer()*/

}
