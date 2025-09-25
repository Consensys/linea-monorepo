package fastpoly

import (
	"testing"

	"github.com/consensys/gnark-crypto/ecc"
	"github.com/consensys/gnark-crypto/field/koalabear"
	"github.com/consensys/gnark-crypto/field/koalabear/fft"
	"github.com/consensys/gnark/frontend"
	"github.com/consensys/gnark/frontend/cs/scs"
	"github.com/consensys/linea-monorepo/prover/maths/common/poly"
	"github.com/consensys/linea-monorepo/prover/maths/field"
	"github.com/consensys/linea-monorepo/prover/protocol/zk"
	"github.com/stretchr/testify/assert"
)

const sizePoly = 8

type Circuit[T zk.Element] struct {
	Poly []T
	X    T
	Y    T
}

func (c *Circuit[T]) Define(api frontend.API) error {

	apiGen, err := zk.NewApi[T](api)
	if err != nil {
		return err
	}
	res := EvaluateLagrangeGnarkGen(api, c.Poly, c.X)

	apiGen.AssertIsEqual(res, &c.Y)

	return nil
}

func testCircuitWitness[T zk.Element]() *Circuit[T] {

	var circuit Circuit[T]

	d := fft.NewDomain(uint64(sizePoly))
	p := make([]field.Element, sizePoly)
	for i := 0; i < sizePoly; i++ {
		//p[i].SetRandom()
		p[i].SetUint64(uint64(i + 2))
	}
	var x field.Element
	x.SetUint64(10)
	y := poly.Eval(p, x)
	circuit.X = *zk.ValueOf[T](x)
	circuit.Y = *zk.ValueOf[T](y)
	// fmt.Printf("y = %s\n", y.String())

	d.FFT(p, fft.DIF)
	fft.BitReverse(p)

	circuit.Poly = make([]T, sizePoly)
	for i := 0; i < sizePoly; i++ {
		circuit.Poly[i] = *zk.ValueOf[T](p[i])
	}

	return &circuit
}

func TestEvaluateLagrangeGnarkGen(t *testing.T) {

	{
		witness := testCircuitWitness[zk.NativeElement]()

		var circuit Circuit[zk.NativeElement]
		circuit.Poly = make([]zk.NativeElement, sizePoly)

		ccs, err := frontend.CompileU32(koalabear.Modulus(), scs.NewBuilder, &circuit)
		assert.NoError(t, err)

		fullWitness, err := frontend.NewWitness(witness, koalabear.Modulus())
		assert.NoError(t, err)
		err = ccs.IsSolved(fullWitness)
		assert.NoError(t, err)
	}

	{
		witness := testCircuitWitness[zk.EmulatedElement]()

		var circuit Circuit[zk.EmulatedElement]
		circuit.Poly = make([]zk.EmulatedElement, sizePoly)

		ccs, err := frontend.Compile(ecc.BLS12_377.ScalarField(), scs.NewBuilder, &circuit)
		assert.NoError(t, err)

		fullWitness, err := frontend.NewWitness(witness, ecc.BLS12_377.ScalarField())
		assert.NoError(t, err)
		err = ccs.IsSolved(fullWitness)
		assert.NoError(t, err)
	}

}
