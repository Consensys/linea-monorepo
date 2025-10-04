package zk

import (
	"testing"

	"github.com/consensys/gnark-crypto/ecc"
	"github.com/consensys/gnark-crypto/field/koalabear"
	"github.com/consensys/gnark/frontend"
	"github.com/consensys/gnark/frontend/cs/scs"
	"github.com/stretchr/testify/assert"
)

type AddCircuit struct {
	A, B  WrappedVariable
	MulAB WrappedVariable
	AddAB WrappedVariable
	SubAB WrappedVariable
	DivAB WrappedVariable
	NegA  WrappedVariable
}

func (c *AddCircuit) Define(api frontend.API) error {

	genApi, err := NewGenericApi(api)
	if err != nil {
		return err
	}

	tmp := genApi.Mul(&c.A, &c.B)
	genApi.AssertIsEqual(tmp, &c.MulAB)

	tmp = genApi.Add(&c.A, &c.B)
	genApi.AssertIsEqual(tmp, &c.AddAB)

	tmp = genApi.Sub(&c.A, &c.B)
	genApi.AssertIsEqual(tmp, &c.SubAB)

	tmp = genApi.Div(&c.A, &c.B)
	genApi.AssertIsEqual(tmp, &c.DivAB)

	tmp = genApi.Neg(&c.A)
	genApi.AssertIsEqual(tmp, &c.NegA)

	return nil
}

func TestAddCircuit(t *testing.T) {

	{
		var witness AddCircuit
		var a, b, mulab, addab, subab, divab, nega koalabear.Element
		a.SetRandom()
		b.SetRandom()
		mulab.Mul(&a, &b)
		addab.Add(&a, &b)
		subab.Sub(&a, &b)
		divab.Div(&a, &b)
		nega.Neg(&a)

		witness.A = ValueOf(a.String())
		witness.B = ValueOf(b.String())
		witness.MulAB = ValueOf(mulab.String())
		witness.AddAB = ValueOf(addab.String())
		witness.SubAB = ValueOf(subab.String())
		witness.DivAB = ValueOf(divab.String())
		witness.NegA = ValueOf(nega.String())

		var circuit AddCircuit
		ccs, err := frontend.CompileU32(koalabear.Modulus(), scs.NewBuilder, &circuit)
		assert.NoError(t, err)

		fullWitness, err := frontend.NewWitness(&witness, koalabear.Modulus())
		assert.NoError(t, err)
		err = ccs.IsSolved(fullWitness)
		assert.NoError(t, err)
	}

	{
		var witness AddCircuit
		var a, b, mulab, addab, subab, divab, nega koalabear.Element
		a.SetRandom()
		b.SetRandom()
		mulab.Mul(&a, &b)
		addab.Add(&a, &b)
		subab.Sub(&a, &b)
		divab.Div(&a, &b)
		nega.Neg(&a)

		witness.A = ValueOf(a.String())
		witness.B = ValueOf(b.String())
		witness.MulAB = ValueOf(mulab.String())
		witness.AddAB = ValueOf(addab.String())
		witness.SubAB = ValueOf(subab.String())
		witness.DivAB = ValueOf(divab.String())
		witness.NegA = ValueOf(nega.String())
		var circuit AddCircuit
		ccs, err := frontend.Compile(ecc.BLS12_377.ScalarField(), scs.NewBuilder, &circuit)
		assert.NoError(t, err)

		fullWitness, err := frontend.NewWitness(&witness, ecc.BLS12_377.ScalarField())
		assert.NoError(t, err)
		err = ccs.IsSolved(fullWitness)
		assert.NoError(t, err)
	}

}
