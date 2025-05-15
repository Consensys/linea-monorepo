package bls

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

	// Number of inputs per G1 subgroup membership circuits
	NbG1MembershipInputInstances int
	// Number of G1 subgroup membership circuits
	NbG1MembershipCircuits int

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

func (l *Limits) nbG1MembershipChecks() int {
	return l.NbG1MembershipInputInstances * l.NbG1MembershipCircuits
}

func (l *Limits) nbG2MembershipChecks() int {
	return l.NbG2MembershipInputInstances * l.NbG2MembershipCircuits
}
