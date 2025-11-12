package gnarkutil

import (
	"github.com/consensys/gnark-crypto/ecc"
	"github.com/consensys/gnark-crypto/field/koalabear"
	"github.com/consensys/gnark/backend/witness"
	"github.com/consensys/linea-monorepo/prover/maths/field/gnarkfext"
	"github.com/consensys/linea-monorepo/prover/maths/zk"
)

/*
Allocate a slice of field element
*/
// func AllocateSlice(n int) []zk.WrappedVariable {
// 	return make([]zk.WrappedVariable, n)
// }

/*
AllocateSliceExt allocates a slice of field extension elements
*/
func AllocateSliceExt(n int) []gnarkfext.E4Gen {
	return make([]gnarkfext.E4Gen, n)
}

// AsWitness converts a slice of field elements to a slice of witness variables
// of the same length with only public inputs.
func AsWitnessPublic(v []zk.WrappedVariable) witness.Witness {

	var (
		wit, _  = witness.New(ecc.BLS12_377.ScalarField())
		witChan = make(chan any, len(v))
	)

	for _, w := range v {
		witChan <- w
	}

	close(witChan)

	if err := wit.Fill(len(v), 0, witChan); err != nil {
		panic(err)
	}

	return wit
}

// AsWitnessPublicSmallField converts a slice of base field elements to a slice of witness variables
// of the same length with only public inputs.
func AsWitnessPublicSmallField(v []zk.WrappedVariable) witness.Witness {

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
