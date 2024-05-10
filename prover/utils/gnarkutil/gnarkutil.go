package gnarkutil

import "github.com/consensys/gnark/frontend"

/*
Allocate a slice of field element
*/
func AllocateSlice(n int) []frontend.Variable {
	return make([]frontend.Variable, n)
}
