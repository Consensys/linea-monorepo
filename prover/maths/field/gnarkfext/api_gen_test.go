package gnarkfext

import (
	"testing"

	"github.com/consensys/gnark-crypto/ecc"
	"github.com/consensys/gnark-crypto/field/koalabear"
	"github.com/consensys/gnark/frontend"
	"github.com/consensys/gnark/frontend/cs/scs"
	"github.com/consensys/gnark/std/math/emulated"
	"github.com/consensys/linea-monorepo/prover/maths/field/fext"
	"github.com/consensys/linea-monorepo/prover/protocol/zk"
	"github.com/stretchr/testify/assert"
)

type ApiCircuitGen[T zk.FType] struct {
	A, B    E4Gen[T]
	ATimesB E4Gen[T]
}

func (c *ApiCircuitGen[T]) Define(api frontend.API) error {

	apiGen, err := zk.NewApi[T](api)
	if err != nil {
		return err
	}

	var atimesb E4Gen[T]
	atimesb.Mul(apiGen, &c.A, &c.B)

	atimesb.AssertIsEqual(apiGen, &c.ATimesB)

	return nil
}

func TestAPIGen(t *testing.T) {

	{
		var witness ApiCircuitGen[frontend.Variable]
		var a, b, tmp fext.Element
		a.SetRandom()
		b.SetRandom()

		witness.A = NewE4Gen[frontend.Variable](a)
		witness.B = NewE4Gen[frontend.Variable](b)
		tmp.Mul(&a, &b)
		witness.ATimesB = NewE4Gen[frontend.Variable](tmp)

		var circuit ApiCircuitGen[frontend.Variable]

		ccs, err := frontend.CompileU32(koalabear.Modulus(), scs.NewBuilder, &circuit)
		assert.NoError(t, err)

		fullWitness, err := frontend.NewWitness(&witness, koalabear.Modulus())
		assert.NoError(t, err)
		err = ccs.IsSolved(fullWitness)
		assert.NoError(t, err)
	}

	{
		var witness ApiCircuitGen[emulated.Element[emulated.KoalaBear]]
		var a, b, tmp fext.Element
		a.SetRandom()
		b.SetRandom()

		witness.A = NewE4Gen[emulated.Element[emulated.KoalaBear]](a)
		witness.B = NewE4Gen[emulated.Element[emulated.KoalaBear]](b)
		tmp.Mul(&a, &b)
		witness.ATimesB = NewE4Gen[emulated.Element[emulated.KoalaBear]](tmp)

		var circuit ApiCircuitGen[emulated.Element[emulated.KoalaBear]]

		ccs, err := frontend.Compile(ecc.BLS12_377.ScalarField(), scs.NewBuilder, &circuit)
		assert.NoError(t, err)

		fullWitness, err := frontend.NewWitness(&witness, ecc.BLS12_377.ScalarField())
		assert.NoError(t, err)
		err = ccs.IsSolved(fullWitness)
		assert.NoError(t, err)
	}

}
