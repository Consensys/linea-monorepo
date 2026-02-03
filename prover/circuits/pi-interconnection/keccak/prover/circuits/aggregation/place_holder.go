package aggregation

import (
	"fmt"
	"sync"

	"github.com/consensys/gnark/frontend/cs/scs"

	"github.com/consensys/gnark-crypto/ecc"
	"github.com/consensys/gnark/constraint"
	"github.com/consensys/gnark/frontend"
	"github.com/consensys/linea-monorepo/prover/circuits/pi-interconnection/keccak/prover/utils"
)

var (
	// @alex: None of these values should be directly accessed. Use the getter
	// function instead. That way, this ensures that the instances are initialized.
	// We use this mechanism because the initialization mechanism is time
	// consuming and we would prefer to not have to do this whenever this
	// module is imported.

	// It is a dummy circuit with the same topography as the execution circuit. It
	// has the same number of constraints (up to the next power of 2) and uses 1
	// BBS commitment.
	_placeHolderCS   constraint.ConstraintSystem
	_placeHolderOnce = &sync.Once{}
)

// Initializes (if necessary) and returns the constraint system of the place-
// holder execution circuit.
func getPlaceHolderCS() constraint.ConstraintSystem {
	_initPlaceHolder()
	return _placeHolderCS
}

// The generic placeholder circuit is a circuit that does not do anything
// meaningful and is meant to be generatable for any number of target
// constraints, public inputs and BBS commitment. Its public inputs should
// always be assigned to zero.
type genericPlaceHolderCircuit struct {
	PubInputs     []frontend.Variable `gnark:",public"`
	Seed          frontend.Variable   `gnark:",secret"`
	nbConstraints int
	nbBbsCommit   int
}

func allocatePlaceHolderCircuit(nbPublicInput, nbConstraints, nbBbsCommit int) *genericPlaceHolderCircuit {
	return &genericPlaceHolderCircuit{
		PubInputs:     make([]frontend.Variable, nbPublicInput),
		nbConstraints: nbConstraints,
		nbBbsCommit:   nbBbsCommit,
	}
}

func (g *genericPlaceHolderCircuit) Define(api frontend.API) error {

	var (
		// This formula is handcrafted so that the circuit has the expected
		// number of constraints.
		nbMul     = g.nbConstraints - 3*g.nbBbsCommit - len(g.PubInputs) - 1
		commitApi = api.(frontend.Committer)
		_x        = frontend.Variable(1)
	)

	for i := 0; i < nbMul; i++ {
		_x = api.Mul(_x, g.Seed)
	}

	for i := 0; i < g.nbBbsCommit; i++ {
		r, err := commitApi.Commit(_x)
		if err != nil {
			return fmt.Errorf("while attempting to BBS commit: %w", err)
		}
		_x = api.Mul(_x, r)
	}

	// With overhelming probability, _x will not be zero since it is the product
	// of many random numbers over a large field. This means that we can easily
	// assign PublicInput == 0.
	expectedPub := api.IsZero(_x)

	for i := range g.PubInputs {
		api.AssertIsEqual(expectedPub, g.PubInputs[i])
	}

	return nil
}

func _initPlaceHolder() {
	_placeHolderOnce.Do(func() {
		var err error
		_placeHolderCircuit := allocatePlaceHolderCircuit(1, 1<<5, 1)
		_placeHolderCS, err = frontend.Compile(
			ecc.BLS12_377.ScalarField(),
			scs.NewBuilder,
			_placeHolderCircuit,
		)
		if err != nil {
			utils.Panic("could not generate the execution place holder circuit: %v", err)
		}
	})
}
