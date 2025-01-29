// Copyright 2020-2024 Consensys Software Inc.
// Licensed under the Apache License, Version 2.0. See the LICENSE file for details.

package combined

import (
	"fmt"
	"github.com/consensys/linea-monorepo/prover/maths/common/smartvectors"
	"github.com/consensys/linea-monorepo/prover/maths/field"
	"github.com/consensys/linea-monorepo/prover/protocol/compiler/globalcs"
	"github.com/consensys/linea-monorepo/prover/protocol/compiler/univariates"
	"github.com/consensys/linea-monorepo/prover/protocol/compiler/vortex"
	"github.com/consensys/linea-monorepo/prover/protocol/ifaces"
	"github.com/consensys/linea-monorepo/prover/protocol/wizard"
	sym "github.com/consensys/linea-monorepo/prover/symbolic"
	"github.com/stretchr/testify/assert"
	"log"
	"testing"

	"github.com/consensys/gnark-crypto/ecc"
	"github.com/consensys/gnark/backend/plonk"
	"github.com/consensys/gnark/frontend/cs/scs"

	"github.com/consensys/gnark/frontend"
	"github.com/consensys/gnark/test/unsafekzg"
)

// In this example we show how to use PLONK with KZG commitments. The circuit that is
// showed here is the same as in ../exponentiate.

// Circuit y == x**e
// only the bitSize least significant bits of e are used
type Circuit struct {
	// tagging a variable is optional
	// default uses variable name and secret visibility.
	WizardVerifier wizard.WizardVerifierCircuit //`gnark:",secret"`
	//wizardProof    wizard.Proof                 `gnark:",secret"`
	/*
		p1             frontend.Variable            `gnark:",public"`
		p2             frontend.Variable            `gnark:",public"`
		coin frontend.Variable*/
}

// Define declares the circuit's constraints
// y == x**e
func (circuit *Circuit) Define(api frontend.API) error {
	circuit.WizardVerifier.Verify(api)
	return nil
}

func BenchmarkTestPlonk(t *testing.B) {

	/*
		define2 := func(b *wizard.Builder) {
			var (
				size = 16

				col1 = b.RegisterCommit("P1", size)
				col2 = b.RegisterCommit("P2", size)

				//coin = b.RegisterRandomCoin(coin.Namef("Coin"), coin.Field)
				univ = b.CompiledIOP.InsertUnivariate(0, "univ", []ifaces.Column{col1, col2})
				_    = univ
			)
			//b.CompiledIOP.InsertLocal(0, "local", sym.Sub(col1, field.Zero()))
			//b.CompiledIOP.InsertLocal(0, "local2", sym.Sub(col2, field.Zero()))
			//b.CompiledIOP.InsertGlobal(0, "global", sym.Sub(col1, sym.Mul(col2, 2)))

		}
		prover2 := func(run *wizard.ProverRuntime) {
			var (
				col1 = smartvectors.ForTest(0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0)
				col2 = smartvectors.ForTest(0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0)
			)
			run.AssignColumn("P1", col1)
			run.AssignColumn("P2", col2)

			run.AssignUnivariate("univ", field.Zero(), field.Zero(), field.Zero())

			//run.GetRandomCoinField(coin.Namef("Coin"))

		}*/

	define := func(b *wizard.Builder) {
		var (
			size = 16

			col1 = b.RegisterCommit("P1", size)
			col2 = b.RegisterCommit("P2", size)

			//coin = b.RegisterRandomCoin(coin.Namef("Coin"), coin.Field)
			univ = b.CompiledIOP.InsertUnivariate(0, "univ", []ifaces.Column{col1, col2})
			_    = univ
		)
		//b.CompiledIOP.InsertLocal(0, "local", sym.Sub(col1, field.Zero()))
		//b.CompiledIOP.InsertLocal(0, "local2", sym.Sub(col2, field.Zero()))
		b.CompiledIOP.InsertGlobal(0, "global", sym.Sub(col1, sym.Mul(col2, 2)))

	}
	prover := func(run *wizard.ProverRuntime) {
		var (
			col1 = smartvectors.ForTest(0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0)
			col2 = smartvectors.ForTest(0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0)
		)
		run.AssignColumn("P1", col1)
		run.AssignColumn("P2", col2)

		run.AssignUnivariate("univ", field.Zero(), field.Zero(), field.Zero())

		//run.GetRandomCoinField(coin.Namef("Coin"))

	}

	//vortex.Compile(2, vortex.WithDryThreshold(0))
	//comp := wizard.Compile(define, globalcs.Compile, vortex.Compile(2, vortex.WithDryThreshold(0)))
	comp := wizard.Compile(define, globalcs.Compile, univariates.MultiPointToSinglePoint(16), vortex.Compile(2, vortex.WithDryThreshold(0)))
	proof := wizard.Prove(comp, prover)
	assert.NoErrorf(t, wizard.Verify(comp, proof), "invalid proof")

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

		proof, err := plonk.Prove(ccs, pk, witnessFull)
		if err != nil {
			log.Fatal(err)
		}

		err = plonk.Verify(proof, vk, witnessPublic)
		if err != nil {
			log.Fatal(err)
		}
	}

}
