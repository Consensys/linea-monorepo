package gnarkutil

import (
	"github.com/consensys/gnark/frontend"
	"github.com/consensys/linea-monorepo/prover/maths/field/fext/gnarkfext"
)

/*
Allocate a slice of field element
*/
func AllocateSlice(n int) []frontend.Variable {
	return make([]frontend.Variable, n)
}

func AllocateSliceExt(n int) []gnarkfext.Variable {
	return make([]gnarkfext.Variable, n)
}
