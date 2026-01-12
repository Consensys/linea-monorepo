package polynomials

import (
	"testing"

	"github.com/consensys/gnark-crypto/ecc"
	"github.com/consensys/gnark-crypto/field/koalabear"
	"github.com/consensys/gnark/frontend"
	"github.com/consensys/gnark/frontend/cs/scs"
	"github.com/consensys/linea-monorepo/prover/maths/field"
	"github.com/consensys/linea-monorepo/prover/maths/field/fext"
	"github.com/consensys/linea-monorepo/prover/maths/field/gnarkfext"
	"github.com/stretchr/testify/assert"
)

type EvalCanonicalCircuit struct {
	Poly []frontend.Variable
	X    gnarkfext.Element
	Y    gnarkfext.Element
}

func (c *EvalCanonicalCircuit) Define(api frontend.API) error {

	y := GnarkEvalCanonical(api, c.Poly, c.X)

	gnarkfext.AssertIsEqual(api, c.Y, y)

	return nil
}

func TestGnarkEvalCanonical(t *testing.T) {

	size := 10
	poly := make([]fext.Element, size)
	for i := 0; i < size; i++ {
		poly[i].B0.A0.SetRandom()
	}
	var x fext.Element
	x.SetRandom()
	y := eval(poly, x)

	{
		var circuit, witness EvalCanonicalCircuit
		circuit.Poly = make([]frontend.Variable, size)
		witness.Poly = make([]frontend.Variable, size)
		for i := 0; i < size; i++ {
			witness.Poly[i] = field.NewFromKoala(poly[i].B0.A0)
		}
		witness.X = gnarkfext.AssignFromExt(x)
		witness.Y = gnarkfext.AssignFromExt(y)

		ccs, err := frontend.CompileU32(koalabear.Modulus(), scs.NewBuilder, &circuit)
		assert.NoError(t, err)

		fullWitness, err := frontend.NewWitness(&witness, koalabear.Modulus())
		assert.NoError(t, err)
		err = ccs.IsSolved(fullWitness)
		assert.NoError(t, err)
	}
	{
		var circuit, witness EvalCanonicalCircuit
		circuit.Poly = make([]frontend.Variable, size)
		witness.Poly = make([]frontend.Variable, size)
		for i := 0; i < size; i++ {
			witness.Poly[i] = field.NewFromKoala(poly[i].B0.A0)
		}
		witness.X = gnarkfext.AssignFromExt(x)
		witness.Y = gnarkfext.AssignFromExt(y)

		ccs, err := frontend.Compile(ecc.BLS12_377.ScalarField(), scs.NewBuilder, &circuit)
		assert.NoError(t, err)

		fullWitness, err := frontend.NewWitness(&witness, ecc.BLS12_377.ScalarField())
		assert.NoError(t, err)
		err = ccs.IsSolved(fullWitness)
		assert.NoError(t, err)
	}
}
