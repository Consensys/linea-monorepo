package gnarkutil

import "github.com/consensys/gnark/frontend"

/*
Allocate a slice of field element
*/
func AllocateSlice(n int) []frontend.Variable {
	return make([]frontend.Variable, n)
}

/*
Allocate a matrix of field element
*/
func AllocateDoubleSlice(outerSize, innerSize int) [][]frontend.Variable {
	res := make([][]frontend.Variable, outerSize)
	for i := range res {
		res[i] = make([]frontend.Variable, innerSize)
	}
	return res
}
