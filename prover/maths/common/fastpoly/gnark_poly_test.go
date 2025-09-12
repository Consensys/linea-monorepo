package fastpoly

import (
	"testing"

	"github.com/consensys/gnark-crypto/field/koalabear/vortex"

	"github.com/consensys/gnark-crypto/field/koalabear"
	"github.com/consensys/gnark/frontend"
	"github.com/consensys/gnark/frontend/cs/scs"
	"github.com/consensys/linea-monorepo/prover/maths/field"
	"github.com/consensys/linea-monorepo/prover/maths/field/fext"
	"github.com/stretchr/testify/assert"
)

type EvaluateLagrangeCircuit struct {
	X    frontend.Variable   // point of evaluation
	Poly []frontend.Variable // poly in Lagrange form
	R    frontend.Variable   // expected result
}

func (c *EvaluateLagrangeCircuit) Define(api frontend.API) error {

	r := EvaluateLagrangeGnark(api, c.Poly, c.X)
	api.AssertIsEqual(c.R, r)

	return nil
}

func TestEvaluateLagrangeGnark(t *testing.T) {

	// sample random poly and random point
	size := 64
	poly := make([]field.Element, size)
	for i := 0; i < size; i++ {
		poly[i].SetRandom()
	}
	var x field.Element
	x.SetRandom()
	var xExt fext.Element
	fext.SetFromBase(&xExt, &x)
	// eval lagrange
	r, _ := vortex.EvalBasePolyLagrange(poly, xExt)
	// test circuit
	var witness, circuit EvaluateLagrangeCircuit
	circuit.Poly = make([]frontend.Variable, size)
	witness.Poly = make([]frontend.Variable, size)
	for i := 0; i < size; i++ {
		witness.Poly[i] = poly[i]
	}
	witness.R = r
	witness.X = x

	ccs, err := frontend.CompileU32(koalabear.Modulus(), scs.NewBuilder, &circuit)
	assert.NoError(t, err)

	fullWitness, err := frontend.NewWitness(&witness, koalabear.Modulus())
	assert.NoError(t, err)
	err = ccs.IsSolved(fullWitness)
	assert.NoError(t, err)
}
