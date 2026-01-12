package polynomials

import (
	"testing"

	"github.com/consensys/gnark-crypto/ecc"
	"github.com/consensys/gnark-crypto/field/koalabear"
	"github.com/consensys/gnark/frontend"
	"github.com/consensys/gnark/frontend/cs/scs"
	"github.com/consensys/linea-monorepo/prover/maths/field/fext"
	"github.com/consensys/linea-monorepo/prover/maths/field/gnarkfext"
	"github.com/consensys/linea-monorepo/prover/maths/zk"
	"github.com/stretchr/testify/assert"
)

type EvalCanonicalCircuit struct {
	Poly []frontend.Variable
	X    gnarkfext.E4Gen
	Y    gnarkfext.E4Gen
}

func (c *EvalCanonicalCircuit) Define(api frontend.API) error {

	y := GnarkEvalCanonical(api, c.Poly, c.X)
	ext4, err := gnarkfext.NewExt4(api)
	if err != nil {
		return err
	}
	ext4.AssertIsEqual(&c.Y, &y)

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
			witness.Poly[i] = zk.ValueFromKoala(poly[i].B0.A0)
		}
		witness.X = gnarkfext.NewE4Gen(x)
		witness.Y = gnarkfext.NewE4Gen(y)

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
			witness.Poly[i] = zk.ValueFromKoala(poly[i].B0.A0)
		}
		witness.X = gnarkfext.NewE4Gen(x)
		witness.Y = gnarkfext.NewE4Gen(y)

		ccs, err := frontend.Compile(ecc.BLS12_377.ScalarField(), scs.NewBuilder, &circuit)
		assert.NoError(t, err)

		fullWitness, err := frontend.NewWitness(&witness, ecc.BLS12_377.ScalarField())
		assert.NoError(t, err)
		err = ccs.IsSolved(fullWitness)
		assert.NoError(t, err)
	}
}
