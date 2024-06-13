package execution

import (
	"github.com/consensys/gnark-crypto/ecc/bls12-377/fr"
	"github.com/consensys/gnark/constraint"
	"github.com/consensys/gnark/frontend"
	"github.com/consensys/gnark/frontend/cs/scs"
	"github.com/consensys/zkevm-monorepo/prover/protocol/wizard"
)

type builder struct {
	comp *wizard.CompiledIOP
}

func NewBuilder(comp *wizard.CompiledIOP) *builder {
	return &builder{comp: comp}
}

func (b *builder) Compile() (constraint.ConstraintSystem, error) {
	return MakeCS(b.comp), nil
}

// builds the circuit
func MakeCS(comp *wizard.CompiledIOP) constraint.ConstraintSystem {
	circuit := Allocate(comp)

	scs, err := frontend.Compile(fr.Modulus(), scs.NewBuilder, &circuit, frontend.WithCapacity(1<<24))
	if err != nil {
		panic(err)
	}
	return scs
}
