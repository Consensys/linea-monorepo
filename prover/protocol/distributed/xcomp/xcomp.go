package xcomp

import "github.com/consensys/linea-monorepo/prover/protocol/wizard"

func GetCrossComp(comps []*wizard.CompiledIOP) *wizard.CompiledIOP {
	round := 0
	for _, comp := range comps {
		round = max(round, comp.NumRounds())
	}

	xComp := &wizard.CompiledIOP{}

	// check that the global sum is zero.
	initialComp.RegisterVerifierAction(1, &va)

	return xComp
}
