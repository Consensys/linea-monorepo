package gnarkutil

import (
	"github.com/consensys/gnark-crypto/field/koalabear"
	"github.com/consensys/gnark/backend/witness"
	"github.com/consensys/gnark/frontend"
	"github.com/consensys/linea-monorepo/prover/maths/field/gnarkfext"
)

/*
Allocate a slice of field element
*/
func AllocateSlice(n int) []frontend.Variable {
	return make([]frontend.Variable, n)
}

/*
AllocateSliceExt allocates a slice of field extension elements
*/
func AllocateSliceExt(n int) []gnarkfext.Element {
	return make([]gnarkfext.Element, n)
}

// AsWitnessPublic converts a slice of small field elements to a slice
// of witness variables of the same length with only public inputs.
func AsWitnessPublic(v []any) witness.Witness {
	var (
		wit, _  = witness.New(koalabear.Modulus())
		witChan = make(chan any)
	)

	go func() {
		for _, w := range v {
			witChan <- w
		}
		close(witChan)
	}()

	if err := wit.Fill(len(v), 0, witChan); err != nil {
		panic(err)
	}

	return wit
}
