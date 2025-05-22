package query

import (
	"fmt"

	"github.com/consensys/gnark-crypto/ecc"
	"github.com/consensys/gnark/backend/witness"
	"github.com/consensys/gnark/frontend"
	"github.com/consensys/gnark/frontend/cs/scs"
	"github.com/consensys/linea-monorepo/prover/protocol/ifaces"
	"golang.org/x/sync/errgroup"
)

// PlonkInWizardAligned is similar to [PlonkInWizard] in the statement made but
// uses a different and more versatile layout for the public inputs. The query
// allows using multi-ary projection queries from multiple sources. Unlike
// [PlonkInWizard], it allows specifying the inputs in a very flexible way and
// take cares of cherry-picking the right data using selectors.
type PlonkInWizardAligned struct {

	// ID is the name of the query
	ID ifaces.QueryID

	// Circuit is the circuit for which the data is aligned.
	Circuit frontend.Circuit

	// NbCircuitInstances is the number of gnark circuit instances we call. We
	// have to consider that for every circuit instance the PI column length has
	// to be padded to a power of two.
	NbCircuitInstances int

	// Selectors[i][row] indicates if Data[i][row] should be accounted for or
	// disregarded.
	Selectors []ifaces.Column

	// Datas is the set of columns holding the data to check with Plonk.
	Data []ifaces.Column

	// PlonkOptions are optional option to the plonk-in-wizard check
	PlonkOptions []PlonkOption
}

// Name implements the [ifaces.Query] interface.
func (q *PlonkInWizardAligned) Name() ifaces.QueryID {
	return q.ID
}

// Check implements the [ifaces.Query] interface.
func (q *PlonkInWizardAligned) Check(run ifaces.Runtime) error {

	projInputs := ProjectionSideMultiAry{}

	for i := range q.Data {
		projInputs.Columns = append(projInputs.Columns, []ifaces.Column{q.Data[i]})
		projInputs.Filters = append(projInputs.Filters, q.Selectors[i])
	}

	var (
		// dataIterator iterates on the input data as it would be understood by
		// the query.
		dataIterator = projInputs.NextIterator(run)
		ccs, compErr = frontend.Compile(ecc.BLS12_377.ScalarField(), scs.NewBuilder, q.Circuit)
		errGroup     = &errgroup.Group{}
	)

	if compErr != nil {
		return fmt.Errorf("while compiling the circuit: %w", compErr)
	}

	nbPublic := ccs.GetNbPublicVariables()

mainLoop:
	for i := 0; i < q.NbCircuitInstances; i++ {

		var witnessFiller chan any

		for j := 0; j < nbPublic; j++ {

			x, ok := dataIterator()

			if !ok && j == 0 {
				break mainLoop
			}

			if !ok {
				errGroup.Go(func() error { return fmt.Errorf("incomplete witness") })
				break mainLoop
			}

			if j == 0 {
				witnessFiller = make(chan any, nbPublic)
			}

			witnessFiller <- x
		}

		close(witnessFiller)
		witness, _ := witness.New(ecc.BLS12_377.ScalarField())

		// Note: having an error here is completely unexpected and that's why it
		// is a panic.
		if err := witness.Fill(nbPublic, 0, witnessFiller); err != nil {
			panic(err)
		}

		errGroup.Go(func() error {
			return ccs.IsSolved(witness)
		})
	}

	return errGroup.Wait()
}
