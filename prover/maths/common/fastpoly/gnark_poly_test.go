package fastpoly

import (
	"testing"

	"github.com/consensys/gnark-crypto/field/koalabear"
	"github.com/consensys/gnark-crypto/field/koalabear/vortex"
	"github.com/consensys/gnark/frontend"
	"github.com/consensys/gnark/frontend/cs/scs"
	"github.com/consensys/linea-monorepo/prover/maths/field"
	"github.com/consensys/linea-monorepo/prover/maths/field/fext"
	"github.com/consensys/linea-monorepo/prover/maths/field/gnarkfext"
	"github.com/consensys/linea-monorepo/prover/maths/zk"
	"github.com/stretchr/testify/assert"
)

type EvaluateLagrangeCircuit struct {
	X    gnarkfext.Element   // point of evaluation
	Poly []frontend.Variable // poly in Lagrange form
	R    gnarkfext.Element   // expected result
}

func (c *EvaluateLagrangeCircuit) Define(api frontend.API) error {

	r := EvaluateLagrangeGnarkMixed(api, c.Poly, c.X)

	gnarkfext.AssertIsEqual(api, c.R, r)

	return nil
}

func getWitnessAndCircuit(t zk.VType) (EvaluateLagrangeCircuit, EvaluateLagrangeCircuit) {

	// sample random poly and random point
	size := 8
	poly := make([]field.Element, size)
	for i := 0; i < size; i++ {
		poly[i].SetRandom()
	}
	var x fext.Element
	x.SetRandom()

	// eval lagrange
	// r, err := vortex.EvalBasePolyLagrange(poly, x)
	r, err := vortex.EvalBasePolyLagrange(poly, x)
	if err != nil {
		panic(err)
	}
	// test circuit
	var witness, circuit EvaluateLagrangeCircuit
	circuit.Poly = make([]frontend.Variable, size)
	witness.Poly = make([]frontend.Variable, size)
	for i := 0; i < size; i++ {
		witness.Poly[i] = field.NewFromKoala(poly[i])
	}
	witness.R = gnarkfext.AssignFromExt(r)
	witness.X = gnarkfext.AssignFromExt(x)

	return circuit, witness

}

func TestEvaluateLagrangeGnark(t *testing.T) {

	{
		circuit, witness := getWitnessAndCircuit(zk.Native)
		ccs, err := frontend.CompileU32(koalabear.Modulus(), scs.NewBuilder, &circuit)
		assert.NoError(t, err)

		fullWitness, err := frontend.NewWitness(&witness, koalabear.Modulus())
		assert.NoError(t, err)
		err = ccs.IsSolved(fullWitness)
		assert.NoError(t, err)
	}

	{
		circuit, witness := getWitnessAndCircuit(zk.Emulated)
		ccs, err := frontend.CompileU32(koalabear.Modulus(), scs.NewBuilder, &circuit)
		assert.NoError(t, err)

		fullWitness, err := frontend.NewWitness(&witness, koalabear.Modulus())
		assert.NoError(t, err)
		err = ccs.IsSolved(fullWitness)
		assert.NoError(t, err)
	}
}
