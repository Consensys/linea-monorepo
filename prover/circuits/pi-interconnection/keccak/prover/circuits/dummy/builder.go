package dummy

import (
	"fmt"
	"math/big"

	"github.com/consensys/gnark/constraint"
	"github.com/consensys/gnark/frontend"
	"github.com/consensys/gnark/frontend/cs/scs"
	"github.com/consensys/linea-monorepo/prover/circuits/pi-interconnection/keccak/prover/circuits"
)

type builder struct {
	circID      circuits.MockCircuitID
	scalarField *big.Int
}

func NewBuilder(circID circuits.MockCircuitID, scalarField *big.Int) *builder {
	return &builder{circID: circID, scalarField: scalarField}
}

func (b *builder) Compile() (constraint.ConstraintSystem, error) {
	ccs, err := MakeCS(b.circID, b.scalarField)
	if err != nil {
		return nil, err
	}

	// do some historic sanity check
	var (
		nbCommit = len(ccs.GetCommitments().(constraint.PlonkCommitments))
		nbPublic = ccs.GetNbPublicVariables()
	)
	if nbCommit != 1 {
		return nil, fmt.Errorf("must have exactly 1 commitment, got %v", nbCommit)
	}
	if nbPublic != 1 {
		return nil, fmt.Errorf("must have exactly 1 public input, got %v", nbPublic)
	}

	return ccs, nil
}

// Compiles the circuit for a given circuit ID
func MakeCS(circID circuits.MockCircuitID, scalarField *big.Int) (constraint.ConstraintSystem, error) {
	circuit := &CircuitDummy{
		ID: int(circID),
	}
	scs, err := frontend.Compile(scalarField, scs.NewBuilder, circuit)
	if err != nil {
		return nil, fmt.Errorf("while calling frontend.Compile for the dummy circuit: %w", err)
	}
	return scs, nil
}
