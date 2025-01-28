// Copyright 2020-2024 Consensys Software Inc.
// Licensed under the Apache License, Version 2.0. See the LICENSE file for details.

package mymimc

import (
	"github.com/consensys/gnark/frontend"
	"github.com/consensys/gnark/std/hash/mimc"
)

const (
	size = 8
)

// Circuit defines a pre-image knowledge proof
// mimc(secret preImage) = public hash
type Circuit struct {
	// struct tag on a variable is optional
	// default uses variable name and secret visibility.
	PreImage []frontend.Variable `gnark:",public"`
	Hash     []frontend.Variable `gnark:",public"`
}

// Define declares the circuit's constraints
// Hash = mimc(PreImage)
func (circuit *Circuit) Define(api frontend.API) error {
	// hash function
	mimc, _ := mimc.NewMiMC(api)

	// specify constraints
	// mimc(preImage) == hash

	mimc.SetState([]frontend.Variable{0})
	for i := 0; i < size; i++ {
		mimc.Write(circuit.PreImage[i])
		hash := mimc.Sum()
		api.AssertIsEqual(circuit.Hash[i], hash)
	}

	return nil
}
