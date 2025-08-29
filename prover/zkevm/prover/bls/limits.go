package bls

import "github.com/consensys/linea-monorepo/prover/utils"

type Limits struct {
	NbG1AddInputInstances   int
	NbG1AddCircuitInstances int

	NbG2AddInputInstances   int
	NbG2AddCircuitInstances int

	NbG1MulInputInstances   int
	NbG1MulCircuitInstances int
	NbG2MulInputInstances   int
	NbG2MulCircuitInstances int

	// Number of inputs per Miller loop circuits. Counted without the last
	// Miller loop which is done in the final exponentiation part.
	NbMillerLoopInputInstances int
	// Number of Miller loop circuits
	NbMillerLoopCircuitInstances int

	// Number of inputs per final exponentiation circuits
	NbFinalExpInputInstances int
	// Number of final exponentiation circuits
	NbFinalExpCircuitInstances int

	// Number of inputs per G1 subgroup membership circuits
	NbG1MembershipInputInstances int
	// Number of G1 subgroup membership circuits
	NbG1MembershipCircuitInstances int

	// Number of inputs per G2 subgroup membership circuits
	NbG2MembershipInputInstances int
	// Number of G2 subgroup membership circuits
	NbG2MembershipCircuitInstances int

	NbG1MapToInputInstances   int
	NbG1MapToCircuitInstances int

	NbG2MapToInputInstances   int
	NbG2MapToCircuitInstances int

	NbC1MembershipInputInstances   int
	NbC1MembershipCircuitInstances int
	NbC2MembershipInputInstances   int
	NbC2MembershipCircuitInstances int

	NbPointEvalInputInstances          int
	NbPointEvalCircuitInstances        int
	NbPointEvalFailureInputInstances   int
	NbPointEvalFailureCircuitInstances int
}

func (l *Limits) sizeAddIntegration(g group) int {
	switch g {
	case G1:
		return utils.NextPowerOfTwo(l.NbG1AddInputInstances*nbRowsPerG1Add) * utils.NextPowerOfTwo(l.NbG1AddCircuitInstances)
	case G2:
		return utils.NextPowerOfTwo(l.NbG2AddInputInstances*nbRowsPerG2Add) * utils.NextPowerOfTwo(l.NbG2AddCircuitInstances)
	default:
		panic("unknown group for bls add integration size")
	}
}

func (l *Limits) sizeMulIntegration(g group) int {
	switch g {
	case G1:
		return utils.NextPowerOfTwo(l.NbG1MulInputInstances*nbRowsPerG1Mul) * utils.NextPowerOfTwo(l.NbG1MulCircuitInstances)
	case G2:
		return utils.NextPowerOfTwo(l.NbG2MulInputInstances*nbRowsPerG2Mul) * utils.NextPowerOfTwo(l.NbG2MulCircuitInstances)
	default:
		panic("unknown group for bls mul integration size")
	}
}

func (l *Limits) sizeMulUnalignedIntegration(g group) int {
	switch g {
	case G1:
		return utils.NextPowerOfTwo(l.NbG1MulInputInstances) * utils.NextPowerOfTwo(l.NbG1MulCircuitInstances)
	case G2:
		return utils.NextPowerOfTwo(l.NbG2MulInputInstances) * utils.NextPowerOfTwo(l.NbG2MulCircuitInstances)
	default:
		panic("unknown group for bls mul unaligned integration size")
	}
}

func (l *Limits) sizePairUnalignedIntegration() int {
	return utils.NextPowerOfTwo(
		max(
			l.NbMillerLoopInputInstances*l.NbMillerLoopCircuitInstances,
			l.NbFinalExpInputInstances*l.NbFinalExpCircuitInstances,
		),
	)
}

func (l *Limits) sizePairMillerLoopIntegration() int {
	return utils.NextPowerOfTwo(l.NbMillerLoopInputInstances*nbRowsPerMillerLoop) * utils.NextPowerOfTwo(l.NbMillerLoopCircuitInstances)
}

func (l *Limits) sizePairFinalExpIntegration() int {
	return utils.NextPowerOfTwo(l.NbFinalExpInputInstances*nbRowsPerFinalExp) * utils.NextPowerOfTwo(l.NbFinalExpCircuitInstances)
}

func (l *Limits) nbAddCircuitInstances(g group) int {
	switch g {
	case G1:
		return l.NbG1AddCircuitInstances
	case G2:
		return l.NbG2AddCircuitInstances
	default:
		panic("unknown group for bls add circuit instances")
	}
}

func (l *Limits) nbMulCircuitInstances(g group) int {
	switch g {
	case G1:
		return l.NbG1MulCircuitInstances
	case G2:
		return l.NbG2MulCircuitInstances
	default:
		panic("unknown group for bls mul circuit instances")
	}
}

func (l *Limits) nbCurveMembershipCircuitInstances(g group) int {
	switch g {
	case G1:
		return l.NbC1MembershipCircuitInstances
	case G2:
		return l.NbC2MembershipCircuitInstances
	default:
		panic("unknown group for bls curve membership instances")
	}
}

func (l *Limits) nbGroupMembershipCircuitInstances(g group) int {
	switch g {
	case G1:
		return l.NbG1MembershipCircuitInstances
	case G2:
		return l.NbG2MembershipCircuitInstances
	default:
		panic("unknown group for bls group membership instances")
	}
}

func (l *Limits) nbMapCircuitInstances(g group) int {
	switch g {
	case G1:
		return l.NbG1MapToCircuitInstances
	case G2:
		return l.NbG2MapToCircuitInstances
	default:
		panic("unknown group for bls map circuit instances")
	}
}
