package execution

import (
	"github.com/consensys/gnark-crypto/ecc/bls12-377/fr"
	"github.com/consensys/gnark/constraint"
	"github.com/consensys/gnark/frontend"
	"github.com/consensys/gnark/frontend/cs/scs"
	"github.com/consensys/linea-monorepo/prover/config"
	"github.com/consensys/linea-monorepo/prover/protocol/wizard"
)

type limitlessBuilder struct {
	congWIOP *wizard.CompiledIOP
}

func NewLimitlessBuilder(congWIOP *wizard.CompiledIOP) *limitlessBuilder {
	return &limitlessBuilder{congWIOP: congWIOP}
}

func (b *limitlessBuilder) Compile() (constraint.ConstraintSystem, error) {
	return makeLimitlessCS(b.congWIOP), nil
}

// Makes the constraint system for the execution-limitless circuit
func makeLimitlessCS(congWIOP *wizard.CompiledIOP) constraint.ConstraintSystem {
	circuit := AllocateLimitless(congWIOP, &config.TracesLimits{})
	scs, err := frontend.Compile(fr.Modulus(), scs.NewBuilder, &circuit, frontend.WithCapacity(1<<24))
	if err != nil {
		panic(err)
	}
	return scs
}
