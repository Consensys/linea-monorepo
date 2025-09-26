package bls

type Limits struct {
	NbG1AddInputInstances int
	NbG2AddInputInstances int

	NbG1MulInputInstances int
	NbG2MulInputInstances int

	// Number of inputs per Miller loop circuits. Counted without the last
	// Miller loop which is done in the final exponentiation part.
	NbMillerLoopInputInstances int

	// Number of inputs per final exponentiation circuits
	NbFinalExpInputInstances int

	// Number of inputs per G1 subgroup membership circuits
	NbG1MembershipInputInstances int

	// Number of inputs per G2 subgroup membership circuits
	NbG2MembershipInputInstances int

	NbG1MapToInputInstances int
	NbG2MapToInputInstances int

	NbC1MembershipInputInstances int
	NbC2MembershipInputInstances int

	NbPointEvalInputInstances        int
	NbPointEvalFailureInputInstances int
}

func (l *Limits) nbAddInputInstances(g group) int {
	switch g {
	case G1:
		return l.NbG1AddInputInstances
	case G2:
		return l.NbG2AddInputInstances
	default:
		panic("unknown group for bls add input instances")
	}
}

func (l *Limits) nbMapInputInstances(g group) int {
	switch g {
	case G1:
		return l.NbG1MapToInputInstances
	case G2:
		return l.NbG2MapToInputInstances
	default:
		panic("unknown group for bls map input instances")
	}
}

func (l *Limits) nbMulInputInstances(g group) int {
	switch g {
	case G1:
		return l.NbG1MulInputInstances
	case G2:
		return l.NbG2MulInputInstances
	default:
		panic("unknown group for bls mul input instances")
	}
}

func (l *Limits) nbCurveMembershipInputInstances(g group) int {
	switch g {
	case G1:
		return l.NbC1MembershipInputInstances
	case G2:
		return l.NbC2MembershipInputInstances
	default:
		panic("unknown group for bls curve membership input instances")
	}
}

func (l *Limits) nbGroupMembershipInputInstances(g group) int {
	switch g {
	case G1:
		return l.NbG1MembershipInputInstances
	case G2:
		return l.NbG2MembershipInputInstances
	default:
		panic("unknown group for bls group membership input instances")
	}
}
