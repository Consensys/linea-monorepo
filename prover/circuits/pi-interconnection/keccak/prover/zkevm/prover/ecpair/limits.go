package ecpair

import (
	"github.com/consensys/linea-monorepo/prover/circuits/pi-interconnection/keccak/prover/utils"
)

// Limits for bounds for different calls to gnark circuits
type Limits struct {
	// Number of inputs per Miller loop circuits. Counted without the last
	// Miller loop which is done in the final exponentiation part.
	NbMillerLoopInputInstances int
	// Number of Miller loop circuits
	NbMillerLoopCircuits int

	// Number of inputs per final exponentiation circuits
	NbFinalExpInputInstances int
	// Number of final exponentiation circuits
	NbFinalExpCircuits int

	// Number of inputs per G2 subgroup membership circuits
	NbG2MembershipInputInstances int
	// Number of G2 subgroup membership circuits
	NbG2MembershipCircuits int
}

func (l *Limits) nbMillerLoops() int {
	return l.NbMillerLoopInputInstances * l.NbMillerLoopCircuits
}

func (l *Limits) nbFinalExps() int {
	return l.NbFinalExpInputInstances * l.NbFinalExpCircuits
}

func (l *Limits) nbG2MembershipChecks() int {
	return l.NbG2MembershipInputInstances * l.NbG2MembershipCircuits
}

func (l *Limits) sizeMillerLoopPart() int {
	return l.nbMillerLoops() * (nbG1Limbs + nbG2Limbs + 2*nbGtLimbs)
}

func (l *Limits) sizeFinalExpPart() int {
	return l.nbFinalExps() * (nbG1Limbs + nbG2Limbs + nbGtLimbs + 2)
}

func (l *Limits) sizeG2MembershipPart() int {
	return l.nbG2MembershipChecks() * (nbG2Limbs + 1)
}

func (l *Limits) sizeECPair() int {
	return utils.NextPowerOfTwo(l.sizeMillerLoopPart() + l.sizeFinalExpPart() + l.sizeG2MembershipPart())
}
