package bls

type Limits struct {
	// Number of G1 addition input instances. Single instance is approx 11000
	// PLONK constraints
	NbG1AddInputInstances int
	// Number of G2 addition input instances. Single instance is approx 25000
	// PLONK constraints.
	NbG2AddInputInstances int

	// Number of G1 scalar multiplication input instances. Single instance is
	// approx 596000 PLONK constraints
	NbG1MulInputInstances int
	// Number of G2 scalar multiplication input instances. Single instance is
	// approx 1.2 million PLONK constraints
	NbG2MulInputInstances int

	// Number of inputs per Miller loop circuits. Counted without the last
	// Miller loop which is done in the final exponentiation part. Single
	// instance is approx 1.68M PLONK constraints
	NbMillerLoopInputInstances int
	// Number of inputs per final exponentiation circuits. Single instance is
	// approx 1.74M PLONK constraints.
	NbFinalExpInputInstances int

	// Number of inputs per G1 subgroup membership circuits. Single instance is
	// approx 317000 PLONK constraints.
	NbG1MembershipInputInstances int
	// Number of inputs per G2 subgroup membership circuits. Single instance is
	// approx 378000 PLONK constraints.
	NbG2MembershipInputInstances int

	// Number of G1 map to curve input instances. Single instance is approx
	// 217000 PLONK constraints.
	NbG1MapToInputInstances int
	// Number of G2 map to curve input instances. Single instance is approx
	// 797000 PLONK constraints.
	NbG2MapToInputInstances int

	// Number of C1 membership input instances. Single instance is approx 4000
	// PLONK constraints
	NbC1MembershipInputInstances int
	// Number of C2 membership input instances. Single instance is approx 8000
	// PLONK constraints.
	NbC2MembershipInputInstances int

	// Number of point evaluation input instances. Single instance is approx
	// 3.19M PLONK constraints.
	NbPointEvalInputInstances int
	// Number of point evaluation failure input instances. Single instance is approx
	// 5.34M PLONK constraints.
	NbPointEvalFailureInputInstances int
}

func (l *Limits) nbAddInputInstances(g Group) int {
	switch g {
	case G1:
		return l.NbG1AddInputInstances
	case G2:
		return l.NbG2AddInputInstances
	default:
		panic("unknown group for bls add input instances")
	}
}

func (l *Limits) nbMapInputInstances(g Group) int {
	switch g {
	case G1:
		return l.NbG1MapToInputInstances
	case G2:
		return l.NbG2MapToInputInstances
	default:
		panic("unknown group for bls map input instances")
	}
}

func (l *Limits) nbMulInputInstances(g Group) int {
	switch g {
	case G1:
		return l.NbG1MulInputInstances
	case G2:
		return l.NbG2MulInputInstances
	default:
		panic("unknown group for bls mul input instances")
	}
}

func (l *Limits) nbCurveMembershipInputInstances(g Group) int {
	switch g {
	case G1:
		return l.NbC1MembershipInputInstances
	case G2:
		return l.NbC2MembershipInputInstances
	default:
		panic("unknown group for bls curve membership input instances")
	}
}

func (l *Limits) nbGroupMembershipInputInstances(g Group) int {
	switch g {
	case G1:
		return l.NbG1MembershipInputInstances
	case G2:
		return l.NbG2MembershipInputInstances
	default:
		panic("unknown group for bls group membership input instances")
	}
}
