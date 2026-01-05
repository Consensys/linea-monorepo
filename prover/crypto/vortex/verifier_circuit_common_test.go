package vortex

import (
	"testing"

	"github.com/consensys/gnark-crypto/field/koalabear/fft"
	"github.com/consensys/gnark-crypto/utils"
	"github.com/consensys/gnark/frontend"
	"github.com/consensys/linea-monorepo/prover/crypto/reedsolomon"
	"github.com/consensys/linea-monorepo/prover/maths/field/fext"
	"github.com/consensys/linea-monorepo/prover/maths/field/gnarkfext"
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

}
