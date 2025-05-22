package fastpoly

import (
	"testing"

	"github.com/consensys/gnark-crypto/field/koalabear/fft"
	"github.com/consensys/linea-monorepo/prover/maths/common/poly"
	"github.com/consensys/linea-monorepo/prover/maths/field"
	"github.com/consensys/linea-monorepo/prover/maths/field/fext"
)

func randomPoly(size int) []field.Element {
	res := make([]field.Element, size)
	for i := 0; i < size; i++ {
		res[i].SetRandom()
	}
	return res
}

func TestEvaluateLagrange(t *testing.T) {

	size := 64
	domain := fft.NewDomain(uint64(size))
	p := randomPoly(size)
	pLagrange := make([]field.Element, size)
	copy(pLagrange, p)
	domain.FFT(pLagrange, fft.DIF)
	fft.BitReverse(pLagrange)

	var x field.Element
	x.SetRandom()

	u := poly.Eval(p, x)
	v := EvaluateLagrange(pLagrange, x)

	tt := u.Equal(&v)
	if !tt {
		t.Fatal("Evaluate Lagrange failed")
	}

}

func TestEvaluateLagrangeFext(t *testing.T) {

	size := 64
	domain := fft.NewDomain(uint64(size))
	p := randomPoly(size)
	pLagrange := make([]field.Element, size)
	copy(pLagrange, p)
	domain.FFT(pLagrange, fft.DIF)
	fft.BitReverse(pLagrange)

	var x fext.Element
	x.SetRandom()

	u := poly.EvalOnExtField(p, x)
	v := EvaluateLagrangeOnFext(pLagrange, x)

	tt := u.Equal(&v)
	if !tt {
		t.Fatal("Evaluate Lagrange failed")
	}

}

func TestBatchLagrangeEvaluation(t *testing.T) {

	sizePoly := 64
	nbPoly := 20

	// sample a bunch of polynomials
	polys := make([][]field.Element, nbPoly)
	for i := 0; i < nbPoly; i++ {
		polys[i] = randomPoly(sizePoly)
	}

	// sample a random point
	x := fext.RandomElement()

	// compute canonical eval
	canEval := make([]fext.Element, nbPoly)
	for i := 0; i < nbPoly; i++ {
		canEval[i] = poly.EvalOnExtField(polys[i], x)
	}

	// change basis
	d := fft.NewDomain(uint64(sizePoly))
	for i := 0; i < nbPoly; i++ {
		d.FFT(polys[i], fft.DIF)
		fft.BitReverse(polys[i])
	}

	// compute lagrange eval
	lagEvalExt := BatchEvaluateLagrangeOnFext(polys, x)

	// check the result
	for i := 0; i < nbPoly; i++ {
		if !lagEvalExt[i].Equal(&canEval[i]) {
			t.Fatal("Error batch evaluation")
		}
	}

}
