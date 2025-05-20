package fastpoly

import (
	"github.com/consensys/linea-monorepo/prover/maths/fft/fastpolyext"
	"github.com/consensys/linea-monorepo/prover/maths/field/fext"

	"github.com/consensys/linea-monorepo/prover/maths/field"
)

/*
Given the evaluations of a polynomial on a domain (whose
size must be a power of two, or panic), return an evaluation
at a chosen x.

As an input the user can specify that the inputs are given
on a coset.

Interpolate(poly []E1, x E4)
*/
func Interpolate(poly []field.Element, x fext.Element, oncoset ...bool) fext.Element {
	n := len(poly)
	polyext := make([]fext.Element, n)
	for i := 0; i < n; i++ {
		fext.FromBase(&polyext[i], &poly[i])
	}

	return fastpolyext.Interpolate(polyext, x, oncoset...)
}

// Batch version of Interpolate
func BatchInterpolate(polys [][]field.Element, x fext.Element, oncoset ...bool) []fext.Element {

	n := len(polys)
	m := len(polys[0])
	polysext := make([][]fext.Element, n)
	for i := 0; i < n; i++ {
		polysext[i] = make([]fext.Element, m)
	}

	for i := 0; i < n; i++ {
		for j := 0; j < m; j++ {
			fext.FromBase(&polysext[i][j], &polys[i][j])
		}

	}

	return fastpolyext.BatchInterpolate(polysext, x, oncoset...)
}
