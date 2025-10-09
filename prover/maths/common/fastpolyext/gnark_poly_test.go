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

	e4Api, err := gnarkfext.NewExt4(api)
	if err != nil {
		panic(err)
	}

	r := EvaluateLagrangeGnark(api, c.Poly, c.X)
	e4Api.AssertIsEqual(&c.R, &r)

	return nil
}

func getWitnessAndCircuit(t *testing.T) (EvaluateLagrangeCircuit, EvaluateLagrangeCircuit) {

	// sample random poly and random point
	size := 8
	poly := make([]fext.Element, size)
	for i := 0; i < size; i++ {
		poly[i].SetRandom()
	}
	var x fext.Element
	x.SetRandom()

	// eval lagrange
	// r, err := vortex.EvalFextPolyLagrange(poly, x)
	r, err := vortex.EvalFextPolyLagrange(poly, x)
	assert.NoError(t, err)

	// test circuit
	var witness, circuit EvaluateLagrangeCircuit
	circuit.Poly = make([]gnarkfext.E4Gen, size)
	witness.Poly = make([]gnarkfext.E4Gen, size)
	for i := 0; i < size; i++ {
		witness.Poly[i] = gnarkfext.NewE4Gen(poly[i])
	}
	witness.R = gnarkfext.NewE4Gen(r)
	witness.X = gnarkfext.NewE4Gen(x)

	return circuit, witness

}

func TestEvaluateLagrangeGnark(t *testing.T) {

	{
		circuit, witness := getWitnessAndCircuit(t)

		ccs, err := frontend.CompileU32(koalabear.Modulus(), scs.NewBuilder, &circuit)
		assert.NoError(t, err)

		fullWitness, err := frontend.NewWitness(&witness, koalabear.Modulus())
		assert.NoError(t, err)
		err = ccs.IsSolved(fullWitness)
		assert.NoError(t, err)
	}
}
