package bls

import (
	"github.com/consensys/gnark/std/algebra/emulated/sw_bls12381"
	"github.com/consensys/gnark/std/algebra/emulated/sw_emulated"
	"github.com/consensys/gnark/std/math/emulated/emparams"
)

// Type aliases for serialization registration
// These provide access to the concrete generic types needed by the serialization package

type (
	// G1 circuit types with sw_emulated.AffinePoint
	MultiAddCircuitG1        = multiAddCircuit[g1ElementWizard, sw_emulated.AffinePoint[emparams.BLS12381Fp]]
	MultiMulCircuitG1        = multiMulCircuit[g1ElementWizard, sw_emulated.AffinePoint[emparams.BLS12381Fp]]
	MultiMapCircuitG1        = multiMapCircuit[g1ElementWizard, sw_emulated.AffinePoint[emparams.BLS12381Fp]]
	MultiCheckableG1NonGroup = multiCheckableCircuit[nonGroupMembershipInstance[g1ElementWizard, sw_emulated.AffinePoint[emparams.BLS12381Fp]]]
	MultiCheckableG1NonCurve = multiCheckableCircuit[nonCurveMembershipInstance[g1ElementWizard, sw_emulated.AffinePoint[emparams.BLS12381Fp]]]

	// G2 circuit types
	MultiAddCircuitG2        = multiAddCircuit[g2ElementWizard, sw_bls12381.G2Affine]
	MultiMulCircuitG2        = multiMulCircuit[g2ElementWizard, sw_bls12381.G2Affine]
	MultiMapCircuitG2        = multiMapCircuit[g2ElementWizard, sw_bls12381.G2Affine]
	MultiCheckableG2NonGroup = multiCheckableCircuit[nonGroupMembershipInstance[g2ElementWizard, sw_bls12381.G2Affine]]
	MultiCheckableG2NonCurve = multiCheckableCircuit[nonCurveMembershipInstance[g2ElementWizard, sw_bls12381.G2Affine]]

	// Non-generic pairing and point eval circuits
	MultiMillerLoopMulCircuit      = multiMillerLoopMulCircuit
	MultiMillerLoopFinalExpCircuit = multiMillerLoopFinalExpCircuit
	MultiPointEvalCircuit          = multiPointEvalCircuit
	MultiPointEvalFailureCircuit   = multiPointEvalFailureCircuit
)

// Constructor functions for serialization registration
func NewMultiAddCircuitG1() *MultiAddCircuitG1 {
	return &MultiAddCircuitG1{}
}

func NewMultiMulCircuitG1() *MultiMulCircuitG1 {
	return &MultiMulCircuitG1{}
}

func NewMultiMulCircuitG2() *MultiMulCircuitG2 {
	return &MultiMulCircuitG2{}
}

func NewMultiCheckableCircuitG1NonGroup() *MultiCheckableG1NonGroup {
	return &MultiCheckableG1NonGroup{}
}

func NewMultiCheckableCircuitG2NonGroup() *MultiCheckableG2NonGroup {
	return &MultiCheckableG2NonGroup{}
}
