package polynomials

import (
	"testing"

	"github.com/consensys/gnark-crypto/ecc"
	"github.com/consensys/gnark-crypto/field/koalabear"
	"github.com/consensys/gnark-crypto/field/koalabear/fft"
	"github.com/consensys/gnark-crypto/utils"
	"github.com/consensys/gnark/frontend"
	"github.com/consensys/gnark/frontend/cs/scs"
	"github.com/consensys/linea-monorepo/prover/maths/field"
	"github.com/consensys/linea-monorepo/prover/maths/field/fext"
	"github.com/consensys/linea-monorepo/prover/maths/field/gnarkfext"
	"github.com/stretchr/testify/assert"
)

type LagrangeEvaluationCircuit struct {
	Poly   []gnarkfext.Element
	X      gnarkfext.Element
	Y      gnarkfext.Element
	Domain *fft.Domain
}

func (c *LagrangeEvaluationCircuit) Define(api frontend.API) error {

	y := GnarkEvaluateLagrangeExt(api, c.Poly, c.X, c.Domain.Generator, c.Domain.Cardinality)
	gnarkfext.AssertIsEqual(api, c.Y, y)

	return nil
}

func fftextinplace(poly []fext.Element, d *fft.Domain) {

	size := len(poly)

	// fft on each coordinates ... @thomas where is the fft ext ???
	buf := make([]field.Element, size)

	for i := 0; i < size; i++ {
		buf[i].Set(&poly[i].B0.A0)
	}
	d.FFT(buf, fft.DIF)
	utils.BitReverse(buf)
	for i := 0; i < size; i++ {
		poly[i].B0.A0.Set(&buf[i])
	}

	for i := 0; i < size; i++ {
		buf[i].Set(&poly[i].B0.A1)
	}
	d.FFT(buf, fft.DIF)
	utils.BitReverse(buf)
	for i := 0; i < size; i++ {
		poly[i].B0.A1.Set(&buf[i])
	}

	for i := 0; i < size; i++ {
		buf[i].Set(&poly[i].B1.A0)
	}
	d.FFT(buf, fft.DIF)
	utils.BitReverse(buf)
	for i := 0; i < size; i++ {
		poly[i].B1.A0.Set(&buf[i])
	}

	for i := 0; i < size; i++ {
		buf[i].Set(&poly[i].B1.A1)
	}
	d.FFT(buf, fft.DIF)
	utils.BitReverse(buf)
	for i := 0; i < size; i++ {
		poly[i].B1.A1.Set(&buf[i])
	}
}

func fftextInverseinplace(poly []fext.Element, d *fft.Domain) {

	size := len(poly)

	// fft on each coordinates ... @thomas where is the fft ext ???
	buf := make([]field.Element, size)

	for i := 0; i < size; i++ {
		buf[i].Set(&poly[i].B0.A0)
	}
	d.FFTInverse(buf, fft.DIF)
	utils.BitReverse(buf)
	for i := 0; i < size; i++ {
		poly[i].B0.A0.Set(&buf[i])
	}

	for i := 0; i < size; i++ {
		buf[i].Set(&poly[i].B0.A1)
	}
	d.FFTInverse(buf, fft.DIF)
	utils.BitReverse(buf)
	for i := 0; i < size; i++ {
		poly[i].B0.A1.Set(&buf[i])
	}

	for i := 0; i < size; i++ {
		buf[i].Set(&poly[i].B1.A0)
	}
	d.FFTInverse(buf, fft.DIF)
	utils.BitReverse(buf)
	for i := 0; i < size; i++ {
		poly[i].B1.A0.Set(&buf[i])
	}

	for i := 0; i < size; i++ {
		buf[i].Set(&poly[i].B1.A1)
	}
	d.FFTInverse(buf, fft.DIF)
	utils.BitReverse(buf)
	for i := 0; i < size; i++ {
		poly[i].B1.A1.Set(&buf[i])
	}
}

func TestLagrangeEvaluation(t *testing.T) {

	size := 64
	d := fft.NewDomain(uint64(size))
	poly := make([]fext.Element, size)
	for i := 0; i < size; i++ {
		poly[i].B0.A0.SetRandom()
	}

	var x fext.Element
	x.SetRandom()
	y := eval(poly, x)

	fftextinplace(poly, d)

	{
		var circuit, witness LagrangeEvaluationCircuit
		circuit.Poly = make([]gnarkfext.Element, size)
		circuit.Domain = d
		witness.Poly = make([]gnarkfext.Element, size)
		for i := 0; i < size; i++ {
			witness.Poly[i] = gnarkfext.AssignFromExt(poly[i])
			witness.X = gnarkfext.AssignFromExt(x)
			witness.Y = gnarkfext.AssignFromExt(y)
		}

		ccs, err := frontend.CompileU32(koalabear.Modulus(), scs.NewBuilder, &circuit)
		assert.NoError(t, err)

		fullWitness, err := frontend.NewWitness(&witness, koalabear.Modulus())
		assert.NoError(t, err)
		err = ccs.IsSolved(fullWitness)
		assert.NoError(t, err)
	}
	{
		var circuit, witness LagrangeEvaluationCircuit
		circuit.Poly = make([]gnarkfext.Element, size)
		circuit.Domain = d
		witness.Poly = make([]gnarkfext.Element, size)
		for i := 0; i < size; i++ {
			witness.Poly[i] = gnarkfext.AssignFromExt(poly[i])
			witness.X = gnarkfext.AssignFromExt(x)
			witness.Y = gnarkfext.AssignFromExt(y)
		}

		ccs, err := frontend.Compile(ecc.BLS12_377.ScalarField(), scs.NewBuilder, &circuit)
		assert.NoError(t, err)

		fullWitness, err := frontend.NewWitness(&witness, ecc.BLS12_377.ScalarField())
		assert.NoError(t, err)
		err = ccs.IsSolved(fullWitness)
		assert.NoError(t, err)
	}
}

type LagrangeAtZCircuit struct {
	X  gnarkfext.Element
	Li []gnarkfext.Element
	d  *fft.Domain
}

func (c *LagrangeAtZCircuit) Define(api frontend.API) error {
	li := gnarkComputeLagrangeAtZ(api, c.X, c.d.Generator, c.d.Cardinality)

	for i := 0; i < len(li); i++ {
		gnarkfext.AssertIsEqual(api, li[i], c.Li[i])
	}
	return nil
}

func TestGnarkComputeLagrangeAtZ(t *testing.T) {

	size := 8
	d := fft.NewDomain(uint64(size))
	li := make([]fext.Element, size)
	var x fext.Element
	x.SetRandom()
	for i := 0; i < size; i++ {
		buf := make([]fext.Element, size)
		buf[i].SetOne()
		fftextInverseinplace(buf, d)
		li[i] = eval(buf, x)
	}

	{
		var circuit, witness LagrangeAtZCircuit
		circuit.Li = make([]gnarkfext.Element, size)
		circuit.d = d
		witness.Li = make([]gnarkfext.Element, size)
		for i := 0; i < size; i++ {
			witness.Li[i] = gnarkfext.AssignFromExt(li[i])
		}
		witness.X = gnarkfext.AssignFromExt(x)
		ccs, err := frontend.CompileU32(koalabear.Modulus(), scs.NewBuilder, &circuit)
		assert.NoError(t, err)

		fullWitness, err := frontend.NewWitness(&witness, koalabear.Modulus())
		assert.NoError(t, err)
		err = ccs.IsSolved(fullWitness)
		assert.NoError(t, err)
	}

	{
		var circuit, witness LagrangeAtZCircuit
		circuit.Li = make([]gnarkfext.Element, size)
		circuit.d = d
		witness.Li = make([]gnarkfext.Element, size)
		for i := 0; i < size; i++ {
			witness.Li[i] = gnarkfext.AssignFromExt(li[i])
		}
		witness.X = gnarkfext.AssignFromExt(x)
		ccs, err := frontend.Compile(ecc.BLS12_377.ScalarField(), scs.NewBuilder, &circuit)
		assert.NoError(t, err)

		fullWitness, err := frontend.NewWitness(&witness, ecc.BLS12_377.ScalarField())
		assert.NoError(t, err)
		err = ccs.IsSolved(fullWitness)
		assert.NoError(t, err)
	}
}
