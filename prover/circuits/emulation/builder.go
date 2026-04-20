package emulation

import (
	"fmt"

	"github.com/consensys/gnark-crypto/ecc"
	"github.com/consensys/gnark/backend/plonk"
	"github.com/consensys/gnark/constraint"
	"github.com/consensys/gnark/frontend"
	"github.com/consensys/gnark/frontend/cs/scs"
)

type builder struct {
	innerVkeys []plonk.VerifyingKey
}

func NewBuilder(
	innerVkeys []plonk.VerifyingKey,
) *builder {
	return &builder{
		innerVkeys: innerVkeys,
	}
}

func (b *builder) Compile() (constraint.ConstraintSystem, error) {
	return MakeCS(b.innerVkeys)
}

// Generate the setup of the aggregation circuit. The provided innerCS can be
func MakeCS(
	innerVkeys []plonk.VerifyingKey,
) (constraint.ConstraintSystem, error) {

	outerCircuit, err := allocateOuterCircuit(innerVkeys)

	if err != nil {
		return nil, fmt.Errorf("while allocating the aggregation circuit: %w", err)
	}

	ccs, err := frontend.Compile(
		ecc.BN254.ScalarField(),
		scs.NewBuilder,
		outerCircuit,
		frontend.WithCapacity(1<<25),
	)

	if err != nil {
		return nil, fmt.Errorf("while compiling the aggregation circuit: %w", err)
	}

	return ccs, nil
}
