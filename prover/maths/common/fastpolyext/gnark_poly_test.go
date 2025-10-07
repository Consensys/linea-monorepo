package fastpolyext

import (
	"testing"

	"github.com/consensys/gnark-crypto/field/koalabear"
	"github.com/consensys/gnark-crypto/field/koalabear/vortex"
	"github.com/consensys/gnark/frontend"
	"github.com/consensys/gnark/frontend/cs/scs"
	"github.com/consensys/linea-monorepo/prover/maths/field/fext"
	"github.com/consensys/linea-monorepo/prover/maths/field/gnarkfext"
	"github.com/stretchr/testify/assert"
)

type EvaluateLagrangeCircuit struct {
	X    gnarkfext.E4Gen   // point of evaluation
	Poly []gnarkfext.E4Gen // poly in Lagrange form
	R    gnarkfext.E4Gen   // expected result
}

func (c *EvaluateLagrangeCircuit) Define(api frontend.API) error {

	r := EvaluateLagrangeGnark(api, c.Poly, c.X)
	c.R.AssertIsEqual(api, r)

	return nil
}

func TestEvaluateLagrangeGnark(t *testing.T) {

	// sample random poly and random point
	size := 64
	poly := make([]fext.Element, size)
	for i := 0; i < size; i++ {
		poly[i].SetRandom()
	}
	var x fext.Element
	x.SetRandom()

	// eval lagrange
	r, err := vortex.EvalFextPolyLagrange(poly, x)
	assert.NoError(t, err)

	// test circuit
	var witness, circuit EvaluateLagrangeCircuit
	circuit.Poly = make([]gnarkfext.E4Gen, size)
	witness.Poly = make([]gnarkfext.E4Gen, size)
	for i := 0; i < size; i++ {
		witness.Poly[i] = gnarkfext.FromValue(poly[i])
	}
	witness.R = gnarkfext.FromValue(r)
	witness.X = gnarkfext.FromValue(x)

	ccs, err := frontend.CompileU32(koalabear.Modulus(), scs.NewBuilder, &circuit)
	assert.NoError(t, err)

	fullWitness, err := frontend.NewWitness(&witness, koalabear.Modulus())
	assert.NoError(t, err)
	err = ccs.IsSolved(fullWitness)
	assert.NoError(t, err)
}
