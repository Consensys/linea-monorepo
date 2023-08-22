package query

import (
	"fmt"

	"github.com/consensys/accelerated-crypto-monorepo/maths/field"
	"github.com/consensys/accelerated-crypto-monorepo/protocol/ifaces"
	"github.com/consensys/accelerated-crypto-monorepo/utils"
	"github.com/consensys/gnark/frontend"
)

// Inclusion constraint
type Inclusion struct {
	Included  []ifaces.Column
	Including []ifaces.Column
	ID        ifaces.QueryID
}

/*
Construct a permutation. Will panic if it is mal-formed
*/
func NewInclusion(id ifaces.QueryID, included, including []ifaces.Column) Inclusion {
	/*
		Both side of the permutation must have the same number of columns
	*/
	if len(included) != len(including) {
		utils.Panic("a and b_ don't have the same number of commitments %v %v", len(including), len(included))
	}

	// This works because we already tested that both sides have the same number of columns
	nCol := len(included)

	// Check the existence of the handles
	for i := 0; i < nCol; i++ {
		included[i].MustExists()
		including[i].MustExists()
	}

	// All columns of including must have the same MaxSize
	_, err := utils.AllReturnEqual(
		ifaces.Column.Size,
		including,
	)

	// Same thing for included
	_, err2 := utils.AllReturnEqual(
		ifaces.Column.Size,
		included,
	)

	if err != nil || err2 != nil {
		utils.Panic("All columns of including and (resp.) included have the same size:\n\t%v\n\t%v", err, err2)
	}

	return Inclusion{Included: included, Including: including, ID: id}
}

/*
Test that the inclusion relation holds. It's a probabilistic
check - good enough for testing
*/
func (r Inclusion) Check(run ifaces.Runtime) error {

	including := make([]ifaces.ColAssignment, len(r.Included))
	included := make([]ifaces.ColAssignment, len(r.Including))

	// Populate the `including`
	for i, pol := range r.Including {
		including[i] = pol.GetColAssignment(run)
	}

	// Populate the included
	for i, pol := range r.Included {
		included[i] = pol.GetColAssignment(run)
	}

	/*
		Sample a random element alpha, usefull for multivalued inclusion checks
		It allows to reference multiple number through a linear combination
	*/
	var alpha field.Element
	_, err := alpha.SetRandom()
	if err != nil {
		// Cannot happen unless the entropy was exhausted
		panic(err)
	}

	/*
		Checks the dimensions of "included" and "including". They should
		have the same number of columns. And they both represent a proper
		matrix : all subslices in the same slice have the same length.
	*/
	if len(included) != len(including) {
		utils.Panic("Including (%v) and included (%v) should have the same length", len(including), len(included))
	}

	nrowIncluded, nrowIncluding := included[0].Len(), including[0].Len()
	for i := range included {
		/*
			We just tested that including and included had the same number
			of rows. So we can safely iterate on their columns in a single
			loop
		*/
		if included[i].Len() != nrowIncluded {
			utils.Panic("included is improper, col 0 has %v row but col %v has %v", nrowIncluded, i, included[i].Len())
		}
		if including[i].Len() != nrowIncluding {
			utils.Panic("including is improper, col 0 has %v row but col %v has %v", nrowIncluding, i, including[i].Len())
		}
	}

	/*
		Gather the elements of including in a set
	*/
	inclusionSet := make(map[field.Element]struct{})

	// Populate the inclusion set
	for i := 0; i < including[0].Len(); i++ {
		rand := rowLinComb(alpha, i, including)
		inclusionSet[rand] = struct{}{}
	}

	// Test that all elements are included in there
	for i := 0; i < included[0].Len(); i++ {
		rand := rowLinComb(alpha, i, included)
		if _, ok := inclusionSet[rand]; !ok {
			notFoundRow := []string{}
			for c := range included {
				x := included[c].Get(i)
				notFoundRow = append(notFoundRow, fmt.Sprintf("%v=%v", r.Included[c].GetColID(), x.String()))
			}
			return fmt.Errorf("row %v was not found in the `including` table : %v", i, notFoundRow)
		}
	}

	return nil
}

// GnarkCheck will panic in this construction because we do not have a good way
// to check the query within a circuit
func (i Inclusion) CheckGnark(api frontend.API, run ifaces.GnarkRuntime) {
	panic("UNSUPPORTED : can't check an inclusion query directly into the circuit")
}
