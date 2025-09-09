package gnarkfext

import (
	"testing"

	"github.com/consensys/gnark-crypto/field/koalabear"
	"github.com/consensys/gnark/frontend"
	"github.com/consensys/gnark/frontend/cs/scs"
	"github.com/consensys/linea-monorepo/prover/maths/field/fext"
	"github.com/stretchr/testify/assert"
)

type ApiCircuit struct {
	A, B Element
	// APlusB   Element
	// AMinusB  Element
	// NegA     Element
	// ATimesB  Element
	// ASquare  Element
	// AExp     Element
	// AMulByFp Element
	// AInverse Element
	ADivB    Element
	fp       int
	exponent int
}

func (c *ApiCircuit) Define(api frontend.API) error {

	var adivb Element
	// var aplusb, aminusb, nega, atimesb, asquare, aexp, mulabyfp, ainverse, adivb Element
	// aplusb.Add(api, c.A, c.B)
	// aminusb.Sub(api, c.A, c.B)
	// nega.Neg(api, c.A)
	// atimesb.Mul(api, c.A, c.B)
	// asquare.Square(api, c.A)
	// aexp = Exp(api, c.A, c.exponent)
	// mulabyfp.MulByFp(api, c.A, c.fp)
	// ainverse.Inverse(api, c.A)
	adivb.Div(api, c.A, c.B)

	// aplusb.AssertIsEqual(api, c.APlusB)
	// aminusb.AssertIsEqual(api, c.AMinusB)
	// nega.AssertIsEqual(api, c.NegA)
	// atimesb.AssertIsEqual(api, c.ATimesB)
	// asquare.AssertIsEqual(api, c.ASquare)
	// aexp.AssertIsEqual(api, c.AExp)
	// mulabyfp.AssertIsEqual(api, c.AMulByFp)
	// ainverse.AssertIsEqual(api, c.AInverse)
	adivb.AssertIsEqual(api, c.ADivB)

	return nil
}

func TestAPINative(t *testing.T) {

	var witness ApiCircuit
	var a, b, tmp fext.Element
	a.SetRandom()
	b.SetRandom()

	witness.A = FromValue(a)
	// witness.B = FromValue(b)
	// tmp.Add(&a, &b)
	// witness.APlusB = FromValue(tmp)
	// tmp.Sub(&a, &b)
	// witness.AMinusB = FromValue(tmp)
	// tmp.Neg(&a)
	// witness.NegA = FromValue(tmp)
	// tmp.Mul(&a, &b)
	// witness.ATimesB = FromValue(tmp)
	// tmp.Square(&a)
	// witness.ASquare = FromValue(tmp)
	// exponent := 101
	// tmp.Exp(a, big.NewInt(int64(exponent)))
	// witness.AExp = FromValue(tmp)
	// witness.exponent = exponent
	// fp := 101
	// var fpKoalabear field.Element
	// fpKoalabear.SetInt64(int64(fp))
	// tmp.MulByElement(&a, &fpKoalabear)
	// witness.AMulByFp = FromValue(tmp)
	// tmp.Inverse(&a)
	// witness.AInverse = FromValue(tmp)
	tmp.Div(&a, &b)
	witness.ADivB = FromValue(tmp)

	var circuit ApiCircuit
	// circuit.exponent = exponent
	// circuit.fp = fp

	_, err := frontend.CompileU32(koalabear.Modulus(), scs.NewBuilder, &circuit)
	assert.NoError(t, err)

	// fullWitness, err := frontend.NewWitness(&witness, koalabear.Modulus())
	// assert.NoError(t, err)
	// err = ccs.IsSolved(fullWitness)
	// assert.NoError(t, err)
}
