package query

import (
	"errors"
	"fmt"
	"sync"

	"github.com/consensys/gnark-crypto/ecc"
	"github.com/consensys/gnark/backend/witness"
	"github.com/consensys/gnark/frontend"
	"github.com/consensys/gnark/frontend/cs/scs"
	"github.com/consensys/linea-monorepo/prover/maths/field"
	"github.com/consensys/linea-monorepo/prover/protocol/ifaces"
	"github.com/consensys/linea-monorepo/prover/utils"
)

var _ ifaces.Query = &PlonkInWizard{}

// PlonkInWizard is a non-parametric query enforcing that a Plonk circuit
// whose public inputs are provided in the input columns is satisfiable.
// To be more precise the input column can specify several instances of
// of the same circuits, the number of instances is dynamic can be read
// for the Selector column. The input data should be formatted as follows:
//
//	|	DATA					|	SELECTOR 		|
//	| <inputs (instance 0)>		| 1 1 1 1 ...		| // Laying out the actual public inputs of the first instance
//	| <zero padding>			| 1 1 1 1 ...		| // Padding to the next power of two positions
//	| <inputs (instance 1)>		| 1 1 1 1 ...		|
//	| <zero padding>			| 1 1 1 1 ...		|
//	| <inputs (instance 2)>		| 1 1 1 1 ...		|
//	| <zero padding>			| 1 1 1 1 ...		|
//	| 0 0 0 0 ... 				| 0 0 0 0 ...		| // Both the selector and the data columns are right-padded with zeroes to highlight that no more instances are expected.
//
// Of course, both the Data and the Selector columns must have the same length
// and the maximal number of instances is given by:
//
// MaxNbInstance := len(data) / NextPowTwo()
//
// A function 'InputFiller' must also be provided to tell the query how to assign
// the circuit.
type PlonkInWizard struct {
	// ID is the unique identifier of the query
	ID ifaces.QueryID
	// Data is the column storing the values to provide as the public inputs of
	// the circuit instance.
	Data ifaces.Column
	// Selector is the binary-decreasing column indicating which portion of the
	// rows of [PlonkInWizard.Data] corresponds to actual public inputs to the
	// circuit to satisfy.
	Selector ifaces.Column
	// Circuit is the circuit to satisfy. The circuit must have zero "secret" values
	// meaning that it should be fully assignable just from the public inputs.
	Circuit frontend.Circuit
	// PlonkOptions are optional options to pass to the circuit when building it
	PlonkOptions []any
	// InputFiller returns an element to pad in the public input for the
	// circuit in case DataToCircuitMask is not full length of the circuit
	// input. If it is nil, then we use zero value.
	InputFiller func(circuitInstance, inputIndex int) field.Element
}

// Name implements the [ifaces.Query] interface
func (piw *PlonkInWizard) Name() ifaces.QueryID {
	return piw.ID
}

// Check implements the [ifaces.Query] interface
func (piw *PlonkInWizard) Check(run ifaces.Runtime) error {

	var (
		data         = piw.Data.GetColAssignment(run).IntoRegVecSaveAlloc()
		sel          = piw.Selector.GetColAssignment(run).IntoRegVecSaveAlloc()
		ccs, compErr = frontend.Compile(ecc.BLS12_377.ScalarField(), scs.NewBuilder, piw.Circuit)
	)

	if compErr != nil {
		return fmt.Errorf("while compiling the circuit: %w", compErr)
	}

	var (
		nbPublic       = ccs.GetNbPublicVariables()
		nbPublicPadded = utils.NextPowerOfTwo(nbPublic)
		wg             = &sync.WaitGroup{}
		errLock        = &sync.Mutex{}
		errSolver      error
	)

	for i := 0; !sel[i].IsZero(); i += nbPublicPadded {

		wg.Add(1)

		go func(i int) {

			defer wg.Done()

			var (
				locPubInputs  = data[i : i+nbPublic]
				locSelector   = sel[i : i+nbPublic]
				witness, _    = witness.New(ecc.BLS12_377.ScalarField())
				witnessFiller = make(chan any, nbPublic)
				currPos       = 0
			)

			for ; currPos < nbPublic; currPos++ {

				if locSelector[currPos].IsZero() {
					witnessFiller <- piw.InputFiller(i, currPos)
					continue
				}

				witnessFiller <- locPubInputs[currPos]
			}

			// closing the channel is necessary to prevent leaking and
			// also to let the witness "know" it is complete.
			close(witnessFiller)
			witness.Fill(nbPublic, 0, witnessFiller)

			if err := ccs.IsSolved(witness); err != nil {
				errLock.Lock()
				errSolver = errors.Join(errSolver, fmt.Errorf("error in solver instance=%v err=%w", i, err))
				errLock.Unlock()
			}
		}(i)
	}

	wg.Wait()

	if errSolver != nil {
		return fmt.Errorf("plonk-in-wizard verifier error: %w", errSolver)
	}

	return nil
}

// CheckGnark implements the [ifaces.Query] interface and will panic in this
// construction because we do not have a good way to check the query within a
// circuit
func (piw *PlonkInWizard) CheckGnark(api frontend.API, run ifaces.GnarkRuntime) {
	utils.Panic("UNSUPPORTED : can't check a PlonkInWizard query directly into the circuit, query-name=%v", piw.Name())
}
