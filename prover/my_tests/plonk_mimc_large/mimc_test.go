// Copyright 2020-2024 Consensys Software Inc.
// Licensed under the Apache License, Version 2.0. See the LICENSE file for details.

package mymimc

import (
	"fmt"
	"github.com/consensys/gnark-crypto/ecc"
	"github.com/consensys/gnark-crypto/ecc/bls12-377/fr"
	"github.com/consensys/gnark/backend/plonk"
	"github.com/consensys/gnark/frontend"
	"github.com/consensys/gnark/frontend/cs/scs"
	"github.com/consensys/gnark/test/unsafekzg"
	"github.com/consensys/linea-monorepo/prover/crypto/mimc"
	"github.com/consensys/linea-monorepo/prover/maths/common/smartvectors"
	"github.com/consensys/linea-monorepo/prover/maths/field"
	"log"
	"testing"
)

func BenchmarkPreimage(b *testing.B) {
	b.StopTimer()
	data := make([]fr.Element, size)
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

	var setupCircuit = Circuit{}
	setupCircuit.Hash = make([]frontend.Variable, size)
	setupCircuit.PreImage = make([]frontend.Variable, size)

	// // building the circuit...
	ccs, err := frontend.Compile(ecc.BLS12_377.ScalarField(), scs.NewBuilder, &setupCircuit)
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

		pk, vk, err := plonk.Setup(ccs, srs, srsLagrange)
		//_, err := plonk.Setup(r1cs, kate, &publicWitness)
		if err != nil {
			log.Fatal(err)
		}

		var mimcCircuit = Circuit{}
		mimcCircuit.Hash = make([]frontend.Variable, size)
		mimcCircuit.PreImage = make([]frontend.Variable, size)

		for i := 0; i < size; i++ {
			mimcCircuit.PreImage[i] = colPre1.Get(i)
			mimcCircuit.Hash[i] = hash[i]
		}

		w, err := frontend.NewWitness(&mimcCircuit, ecc.BLS12_377.ScalarField())

		wPub, err := frontend.NewWitness(&mimcCircuit, ecc.BLS12_377.ScalarField(), frontend.PublicOnly())

		b.StartTimer()
		proof, err := plonk.Prove(ccs, pk, w)
		if err != nil {
			log.Fatal(err)
		}

		err = plonk.Verify(proof, vk, wPub)
		if err != nil {
			log.Fatal(err)
		}
	}

	//ecc.BLS12_377

}
