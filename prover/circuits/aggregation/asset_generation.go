package aggregation

import (
	"fmt"

	"github.com/consensys/gnark-crypto/ecc"
	"github.com/consensys/gnark/backend/plonk"
	"github.com/consensys/gnark/constraint"
	"github.com/consensys/gnark/frontend"
	"github.com/consensys/gnark/frontend/cs/scs"
)

type builder struct {
	maxNbProofs   int
	vKeys         []plonk.VerifyingKey
	allowedInputs []string
	piVKey        plonk.VerifyingKey
}

func NewBuilder(
	maxNbProofs int,
	allowedInputs []string,
	piVKey plonk.VerifyingKey,
	vKeys []plonk.VerifyingKey,
) *builder {
	return &builder{
		piVKey:        piVKey,
		allowedInputs: allowedInputs,
		maxNbProofs:   maxNbProofs,
		vKeys:         vKeys,
	}
}

func (b *builder) Compile() (constraint.ConstraintSystem, error) {
	return MakeCS(b.maxNbProofs, b.allowedInputs, b.piVKey, b.vKeys)
}

// Initializes the bw6 aggregation circuit and returns a compiled constraint
// system.
func MakeCS(
	maxNbProofs int,
	allowedInputs []string,
	piVKey plonk.VerifyingKey,
	vKeys []plonk.VerifyingKey,
) (constraint.ConstraintSystem, error) {

	aggCircuit, err := AllocateCircuit(
		maxNbProofs,
		allowedInputs,
		piVKey,
		vKeys,
	)

	if err != nil {
		return nil, fmt.Errorf("while allocating the aggregation circuit: %w", err)
	}

	ccs, err := frontend.Compile(
		ecc.BW6_761.ScalarField(),
		scs.NewBuilder,
		aggCircuit,
		frontend.WithCapacity(1<<27),
	)

	if err != nil {
		return nil, fmt.Errorf("while compiling the aggregation circuit: %w", err)
	}

	return ccs, nil
}
