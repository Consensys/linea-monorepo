// Copyright 2020-2024 Consensys Software Inc.
// Licensed under the Apache License, Version 2.0. See the LICENSE file for details.

package wizinplonkmimc

import (
	"fmt"
	"github.com/consensys/linea-monorepo/prover/crypto/mimc"
	"github.com/consensys/linea-monorepo/prover/maths/common/smartvectors"
	"github.com/consensys/linea-monorepo/prover/maths/field"
	"github.com/consensys/linea-monorepo/prover/protocol/compiler"
	mimcComp "github.com/consensys/linea-monorepo/prover/protocol/compiler/mimc"
	"github.com/consensys/linea-monorepo/prover/protocol/ifaces"
	"github.com/consensys/linea-monorepo/prover/protocol/wizard"
	"github.com/stretchr/testify/assert"
	"log"
	"testing"

	"github.com/consensys/gnark-crypto/ecc"
	"github.com/consensys/gnark/backend/plonk"
	"github.com/consensys/gnark/frontend/cs/scs"

	"github.com/consensys/gnark/frontend"
	"github.com/consensys/gnark/test/unsafekzg"
)

const (
	size = 16
)

// In this example we show how to use PLONK with KZG commitments. The circuit that is
// showed here is the same as in ../exponentiate.

// only the bitSize least significant bits of e are used
type Circuit struct {
	WizardVerifier wizard.WizardVerifierCircuit //`gnark:",secret"`
}

// Define declares the circuit's constraints
func (circuit *Circuit) Define(api frontend.API) error {
	circuit.WizardVerifier.Verify(api)
	return nil
}

func BenchmarkWizardInPlonkMiMC(bench *testing.B) {
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
	assert.NoErrorf(bench, wizard.Verify(comp, proof), "invalid proof")
	bench.StopTimer()

	// END OF WIZARD STUFF

	var circ Circuit
	// circ.wizardProof = proof
	wizardCirc, err := wizard.AllocateWizardCircuit(comp)
	circ.WizardVerifier = *wizardCirc

	// // building the circuit...
	ccs, err := frontend.Compile(ecc.BLS12_377.ScalarField(), scs.NewBuilder, &circ)
	if err != nil {
		fmt.Println("circuit compilation error")
	}

	// create the necessary data for KZG.
	// This is a toy example, normally the trusted setup to build ZKG
	// has been run before.
	// The size of the data in KZG should be the closest power of 2 bounding //
	// above max(nbConstraints, nbVariables).

	//scs := ccs.(*cs.SparseR1CS)
	scs := ccs
	srs, srsLagrange, err := unsafekzg.NewSRS(scs)
	if err != nil {
		panic(err)
	}

	// Correct data: the proof passes
	{
		// Witnesses instantiation. Witness is known only by the prover,
		// while public w is a public data known by the verifier.

		assigWizardVerifier := wizard.GetWizardVerifierCircuitAssignment(comp, proof)
		var assigCirc Circuit
		assigCirc.WizardVerifier = *assigWizardVerifier

		witnessFull, err := frontend.NewWitness(&assigCirc, ecc.BLS12_377.ScalarField())
		if err != nil {
			log.Fatal(err)
		}

		witnessPublic, err := frontend.NewWitness(&assigCirc, ecc.BLS12_377.ScalarField(), frontend.PublicOnly())
		if err != nil {
			log.Fatal(err)
		}

		// public data consists of the polynomials describing the constants involved
		// in the constraints, the polynomial describing the permutation ("grand
		// product argument"), and the FFT domains.
		pk, vk, err := plonk.Setup(ccs, srs, srsLagrange)
		//_, err := plonk.Setup(r1cs, kate, &publicWitness)
		if err != nil {
			log.Fatal(err)
		}
		bench.StartTimer()

		proof, err := plonk.Prove(ccs, pk, witnessFull)
		if err != nil {
			log.Fatal(err)
		}

		err = plonk.Verify(proof, vk, witnessPublic)
		if err != nil {
			log.Fatal(err)
		}
		bench.StopTimer()
	}

}
