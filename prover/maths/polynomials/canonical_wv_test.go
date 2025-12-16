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
	Poly []zk.WrappedVariable
	X    gnarkfext.E4Gen
	Y    gnarkfext.E4Gen
}

func (c *EvalCanonicalCircuit) Define(api frontend.API) error {

	y := GnarkEvalCanonical(api, c.Poly, c.X)
	apiGen, err := zk.NewGenericApi(api)
	if err != nil {
		return err
	}
	apiGen.AssertIsEqual(c.Y.B0.A0, y.B0.A0)
	apiGen.AssertIsEqual(c.Y.B0.A1, y.B0.A1)
	apiGen.AssertIsEqual(c.Y.B1.A0, y.B1.A0)
	apiGen.AssertIsEqual(c.Y.B1.A1, y.B1.A1)

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
		circuit.Poly = make([]zk.WrappedVariable, size)
		witness.Poly = make([]zk.WrappedVariable, size)
		for i := 0; i < size; i++ {
			witness.Poly[i] = zk.ValueOf(poly[i].B0.A0.String())
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
		circuit.Poly = make([]zk.WrappedVariable, size)
		witness.Poly = make([]zk.WrappedVariable, size)
		for i := 0; i < size; i++ {
			witness.Poly[i] = zk.ValueOf(poly[i].B0.A0.String())
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
