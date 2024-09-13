package plonk

import (
	"fmt"
	"reflect"

	"github.com/consensys/gnark-crypto/ecc"
	"github.com/consensys/gnark/backend/witness"
	"github.com/consensys/gnark/frontend"
	"github.com/consensys/linea-monorepo/prover/protocol/wizard"
)

// WitnessAssigner allows obtaining witness assignment for a circuit.
type WitnessAssigner interface {
	NumEffWitnesses(run *wizard.ProverRuntime) int
	Assign(run *wizard.ProverRuntime, i int) (private, public witness.Witness, err error)
}

type witnessFuncAssigner struct {
	circuit   frontend.Circuit
	assigners []func() frontend.Circuit
}

func (w *witnessFuncAssigner) NumEffWitnesses(run *wizard.ProverRuntime) int {
	return len(w.assigners)
}

func (w *witnessFuncAssigner) Assign(run *wizard.ProverRuntime, i int) (private, public witness.Witness, err error) {
	// does not use runtime. Assignment comes externally from the assign
	// function.
	if i >= len(w.assigners) {
		return nil, nil, fmt.Errorf("index out of range")
	}
	assign := w.assigners[i]
	assignment := assign()
	if w.circuit != nil && reflect.TypeOf(w.circuit) != reflect.TypeOf(assignment) {
		return nil, nil, fmt.Errorf("circuit and assignment do not have the same type")
	}
	witness, err := frontend.NewWitness(assignment, ecc.BLS12_377.ScalarField())
	if err != nil {
		return nil, nil, fmt.Errorf("new witness: %W", err)
	}
	pubWitness, err := witness.Public()
	if err != nil {
		return nil, nil, fmt.Errorf("public witness: %w", err)
	}
	return witness, pubWitness, nil
}

// NewSafeCircuitAssigner returns a WitnessAssigner that returns the private and
// public witness of the circuit. The assign function is called to get the
// assignment of the circuit.
//
// The function returns an error if the circuit and the assignment do not have
// the same type. For the unsafe version use [NewUnsafeCircuitAssigner].
func NewSafeCircuitAssigner(circuit frontend.Circuit, assigners ...func() frontend.Circuit) WitnessAssigner {
	return &witnessFuncAssigner{
		circuit:   circuit,
		assigners: assigners,
	}
}

// NewUnsafeCircuitAssigner returns a WitnessAssigner that returns the private
// and public witness of the circuit. The assign function is called to get the
// assignment of the circuit.
//
// The function does not check if the circuit and the assignment have the same
// type. For the safe version use [NewSafeCircuitAssigner].
func NewUnsafeCircuitAssigner(assigners ...func() frontend.Circuit) WitnessAssigner {
	return &witnessFuncAssigner{
		assigners: assigners,
	}
}
