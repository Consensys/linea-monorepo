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
	X    gnarkfext.Element   // point of evaluation
	Poly []gnarkfext.Element // poly in Lagrange form
	R    gnarkfext.Element   // expected result
}

func (c *EvaluateLagrangeCircuit) Define(api frontend.API) error {

	r := EvaluateLagrangeGnark(api, c.Poly, c.X)
	gnarkfext.AssertIsEqual(api, c.R, r)

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
	circuit.Poly = make([]gnarkfext.Element, size)
	witness.Poly = make([]gnarkfext.Element, size)
	for i := 0; i < size; i++ {
		witness.Poly[i] = gnarkfext.AssignFromExt(poly[i])
	}
	witness.R = gnarkfext.AssignFromExt(r)
	witness.X = gnarkfext.AssignFromExt(x)

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
