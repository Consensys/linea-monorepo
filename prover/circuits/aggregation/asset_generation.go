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
	maxNbProofs int
	vKeys       []plonk.VerifyingKey
}

func NewBuilder(
	maxNbProofs int,
	vKeys []plonk.VerifyingKey,
) *builder {
	return &builder{
		maxNbProofs: maxNbProofs,
		vKeys:       vKeys,
	}
}

func (b *builder) Compile() (constraint.ConstraintSystem, error) {
	return MakeCS(b.maxNbProofs, b.vKeys)
}

// Initializes the bw6 aggregation circuit and returns a compiled constraint
// system.
func MakeCS(
	maxNbProofs int,
	vKeys []plonk.VerifyingKey,
) (constraint.ConstraintSystem, error) {

	aggCircuit, err := AllocateAggregationCircuit(
		maxNbProofs,
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
