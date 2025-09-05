package gnarkfext

import (
	"testing"

	"github.com/consensys/gnark-crypto/ecc"
	"github.com/consensys/gnark-crypto/field/koalabear"
	"github.com/consensys/gnark/frontend"
	"github.com/consensys/gnark/frontend/cs/scs"
	"github.com/consensys/linea-monorepo/prover/maths/field/fext"
	"github.com/consensys/linea-monorepo/prover/protocol/zk"
	"github.com/stretchr/testify/assert"
)

type ApiCircuitGen[T zk.Element] struct {
	A, B    E4Gen[T]
	ATimesB E4Gen[T]
}

func (c *ApiCircuitGen[T]) Define(api frontend.API) error {

	ext4, err := NewExt4[T](api)
	if err != nil {
		return err
	}
	atimesb := ext4.Mul(&c.A, &c.B)
	ext4.AssertIsEqual(atimesb, &c.ATimesB)

	return nil
}

func testApiGenWitness[T zk.Element]() *ApiCircuitGen[T] {
	var a, b, tmp fext.Element
	a.SetRandom()
	b.SetRandom()
	tmp.Mul(&a, &b)
	return &ApiCircuitGen[T]{
		A:       NewE4Gen[T](a),
		B:       NewE4Gen[T](b),
		ATimesB: NewE4Gen[T](tmp),
	}
}

func TestAPIGen(t *testing.T) {

	{
		witness := testApiGenWitness[zk.NativeElement]()

		var circuit ApiCircuitGen[zk.NativeElement]

		ccs, err := frontend.CompileU32(koalabear.Modulus(), scs.NewBuilder, &circuit)
		assert.NoError(t, err)

		fullWitness, err := frontend.NewWitness(witness, koalabear.Modulus())
		assert.NoError(t, err)
		err = ccs.IsSolved(fullWitness)
		assert.NoError(t, err)
	}

	{
		witness := testApiGenWitness[zk.EmulatedElement]()

		var circuit ApiCircuitGen[zk.EmulatedElement]

		ccs, err := frontend.Compile(ecc.BLS12_377.ScalarField(), scs.NewBuilder, &circuit)
		assert.NoError(t, err)

		fullWitness, err := frontend.NewWitness(witness, ecc.BLS12_377.ScalarField())
		assert.NoError(t, err)
		err = ccs.IsSolved(fullWitness)
		assert.NoError(t, err)
	}

}
