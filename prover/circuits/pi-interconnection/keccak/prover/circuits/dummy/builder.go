package dummy

import (
	"fmt"
	"math/big"

	"github.com/consensys/gnark/constraint"
	"github.com/consensys/gnark/frontend"
	"github.com/consensys/gnark/frontend/cs/scs"
	"github.com/consensys/linea-monorepo/prover/circuits/pi-interconnection/keccak/prover/circuits"
)

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
