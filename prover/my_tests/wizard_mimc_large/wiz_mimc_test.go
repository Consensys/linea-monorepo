// Copyright 2020-2024 Consensys Software Inc.
// Licensed under the Apache License, Version 2.0. See the LICENSE file for details.

package wizmimc

import (
	"github.com/consensys/linea-monorepo/prover/crypto/mimc"
	"github.com/consensys/linea-monorepo/prover/maths/common/smartvectors"
	"github.com/consensys/linea-monorepo/prover/maths/field"
	"github.com/consensys/linea-monorepo/prover/protocol/compiler"
	mimcComp "github.com/consensys/linea-monorepo/prover/protocol/compiler/mimc"
	"github.com/consensys/linea-monorepo/prover/protocol/ifaces"
	"github.com/consensys/linea-monorepo/prover/protocol/wizard"
	"github.com/stretchr/testify/assert"

	"testing"
)

const (
	size = 8
)

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

	//vortex.Compile(2, vortex.WithDryThreshold(0))
	//comp := wizard.Compile(define, globalcs.Compile, vortex.Compile(2, vortex.WithDryThreshold(0)))
	//comp := wizard.Compile(define, compiler.Arcane(16, 64))
	// comp := wizard.Compile(define, vortex.Compile(2, vortex.WithDryThreshold(0)))
	comp := wizard.Compile(define, mimcComp.CompileMiMC, compiler.Arcane(16, 16))
	bench.StartTimer()
	proof := wizard.Prove(comp, prover)
	checkErr := wizard.Verify(comp, proof)
	assert.NoErrorf(bench, checkErr, "INVALID proof")
	bench.StopTimer()

	/*
		proofSize := 0
		allValues := proof.Messages.InnerMap()
		for _, value := range allValues {
			proofSize += value.Len() * field.Bytes
		}

		fmt.Println("Wizard proof size ", proofSize)*/

}
