package gnarkfext

import (
	"math/big"
	"testing"

	"github.com/consensys/gnark-crypto/ecc"
	"github.com/consensys/gnark-crypto/field/koalabear"
	"github.com/consensys/gnark/frontend"
	"github.com/consensys/gnark/frontend/cs/scs"
	"github.com/consensys/linea-monorepo/prover/maths/field/fext"
	"github.com/stretchr/testify/assert"
)

var sizeExpo = 12345

type ApiCircuitGen struct {
	A, B    E4Gen
	AddAB   E4Gen
	SubAB   E4Gen
	MulAB   E4Gen
	SquareA E4Gen
	DivAB   E4Gen
	InvA    E4Gen
	ExpA    E4Gen
	n       big.Int
}

func (c *ApiCircuitGen) Define(api frontend.API) error {

	ext4, err := NewExt4(api)
	if err != nil {
		return err
	}
	addAB := ext4.Add(&c.A, &c.B)
	ext4.AssertIsEqual(addAB, &c.AddAB)

	subAB := ext4.Sub(&c.A, &c.B)
	ext4.AssertIsEqual(subAB, &c.SubAB)

	mulAB := ext4.Mul(&c.A, &c.B)
	ext4.AssertIsEqual(mulAB, &c.MulAB)

	squareA := ext4.Square(&c.A) // TODO Square mysteriously fails
	ext4.AssertIsEqual(squareA, &c.SquareA)

	divAB := ext4.Div(&c.A, &c.B)
	ext4.AssertIsEqual(divAB, &c.DivAB)

	invA := ext4.Inverse(&c.A)
	ext4.AssertIsEqual(invA, &c.InvA)

	expA := ext4.Exp(&c.A, &c.n)
	ext4.AssertIsEqual(expA, &c.ExpA)

	return nil
}

func testApiGenWitness() *ApiCircuitGen {
	var a, b, addab, subab, mulab, squarea, inva, divab, expa fext.Element
	var n big.Int
	n.SetUint64(uint64(sizeExpo))
	a.SetRandom()
	b.SetRandom()
	addab.Add(&a, &b)
	subab.Sub(&a, &b)
	mulab.Mul(&a, &b)
	squarea.Square(&a)
	divab.Div(&a, &b)
	inva.Inverse(&a)
	expa.Exp(a, &n)
	return &ApiCircuitGen{
		A:       NewE4Gen(a),
		B:       NewE4Gen(b),
		AddAB:   NewE4Gen(addab),
		SubAB:   NewE4Gen(subab),
		MulAB:   NewE4Gen(mulab),
		SquareA: NewE4Gen(squarea),
		DivAB:   NewE4Gen(divab),
		InvA:    NewE4Gen(inva),
		ExpA:    NewE4Gen(expa),
		n:       n,
	}
}

func TestAPIGen(t *testing.T) {

	{
		witness := testApiGenWitness()

		var circuit ApiCircuitGen
		circuit.n.SetUint64(uint64(sizeExpo))

		ccs, err := frontend.CompileU32(koalabear.Modulus(), scs.NewBuilder, &circuit)
		assert.NoError(t, err)

		fullWitness, err := frontend.NewWitness(witness, koalabear.Modulus())
		assert.NoError(t, err)
		err = ccs.IsSolved(fullWitness)
		assert.NoError(t, err)
	}

	{
		witness := testApiGenWitness()

		var circuit ApiCircuitGen
		circuit.n.SetUint64(uint64(sizeExpo))

		ccs, err := frontend.Compile(ecc.BLS12_377.ScalarField(), scs.NewBuilder, &circuit)
		assert.NoError(t, err)

		fullWitness, err := frontend.NewWitness(witness, ecc.BLS12_377.ScalarField())
		assert.NoError(t, err)
		err = ccs.IsSolved(fullWitness)
		assert.NoError(t, err)
	}

}
