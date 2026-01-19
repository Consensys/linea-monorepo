package vortex

import (
	"testing"

	"github.com/consensys/gnark-crypto/field/koalabear"
	"github.com/consensys/gnark-crypto/field/koalabear/fft"
	"github.com/consensys/gnark-crypto/utils"
	"github.com/consensys/gnark/frontend"
	"github.com/consensys/gnark/frontend/cs/scs"
	"github.com/consensys/linea-monorepo/prover/crypto/reedsolomon"
	"github.com/consensys/linea-monorepo/prover/maths/field/fext"
	"github.com/consensys/linea-monorepo/prover/maths/field/gnarkfext"
	"github.com/stretchr/testify/assert"
)

type IsCodeWordCircuit struct {
	Poly   []gnarkfext.E4Gen
	params reedsolomon.RsParams
}

func (c *IsCodeWordCircuit) Define(api frontend.API) error {

	GnarkCheckIsCodeWord(api, c.params, c.Poly)

	return nil
}

func GenerateCircuitWitness(size, rate int) (*IsCodeWordCircuit, *IsCodeWordCircuit) {

	sizeCodeWord := size * rate
	p := make([]fext.Element, sizeCodeWord)
	for i := 0; i < size; i++ {
		p[i].SetRandom()
	}
	d := fft.NewDomain(uint64(sizeCodeWord))
	d.FFTExt(p, fft.DIF)
	utils.BitReverse(p)
	rsParams := reedsolomon.NewRsParams(size, rate)

	var circuit, witness IsCodeWordCircuit
	circuit.Poly = make([]gnarkfext.E4Gen, sizeCodeWord)
	circuit.params = *rsParams
	witness.Poly = make([]gnarkfext.E4Gen, sizeCodeWord)
	for i := 0; i < sizeCodeWord; i++ {
		witness.Poly[i] = gnarkfext.NewE4Gen(p[i])
	}

	return &circuit, &witness

}

func TestIsCodeWord(t *testing.T) {

	size := 64
	rate := 4

	// native
	{
		circuit, witness := GenerateCircuitWitness(size, rate)

		ccs, err := frontend.CompileU32(koalabear.Modulus(), scs.NewBuilder, circuit)
		assert.NoError(t, err)

		fullWitness, err := frontend.NewWitness(witness, koalabear.Modulus())
		assert.NoError(t, err)
		err = ccs.IsSolved(fullWitness)
		assert.NoError(t, err)
	}

}
