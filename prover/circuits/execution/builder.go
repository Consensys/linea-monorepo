package execution

import (
	"github.com/consensys/gnark-crypto/ecc/bls12-377/fr"
	"github.com/consensys/gnark/constraint"
	"github.com/consensys/gnark/frontend"
	"github.com/consensys/gnark/frontend/cs/scs"
	"github.com/consensys/zkevm-monorepo/prover/protocol/wizard"
	"github.com/consensys/zkevm-monorepo/prover/zkevm"
	"github.com/consensys/zkevm-monorepo/prover/zkevm/prover/publicInput"
)

type builder struct {
	comp      *wizard.CompiledIOP
	extractor *publicInput.FunctionalInputExtractor
}

func NewBuilder(z *zkevm.ZkEvm) *builder {
	return &builder{comp: z.WizardIOP, extractor: z.PublicInput.GetExtractor()}
}

func (b *builder) Compile() (constraint.ConstraintSystem, error) {
	return makeCS(b.comp, b.extractor), nil
}

// builds the circuit
func makeCS(comp *wizard.CompiledIOP, ext *publicInput.FunctionalInputExtractor) constraint.ConstraintSystem {
	circuit := Allocate(comp, ext)

	scs, err := frontend.Compile(fr.Modulus(), scs.NewBuilder, &circuit, frontend.WithCapacity(1<<24))
	if err != nil {
		panic(err)
	}
	return scs
}
