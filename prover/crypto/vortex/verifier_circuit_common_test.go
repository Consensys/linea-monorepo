package vortex

import (
	"testing"

	"github.com/consensys/gnark-crypto/ecc"
	"github.com/consensys/gnark-crypto/field/koalabear"
	"github.com/consensys/gnark-crypto/field/koalabear/fft"
	"github.com/consensys/gnark-crypto/utils"
	"github.com/consensys/gnark/frontend"
	"github.com/consensys/gnark/frontend/cs/scs"
	"github.com/consensys/linea-monorepo/prover/crypto/reedsolomon"
	sv "github.com/consensys/linea-monorepo/prover/maths/common/smartvectors"
	"github.com/consensys/linea-monorepo/prover/maths/common/smartvectors_mixed"
	"github.com/consensys/linea-monorepo/prover/maths/field/fext"
	"github.com/consensys/linea-monorepo/prover/maths/field/gnarkfext"
	"github.com/stretchr/testify/assert"
)

type StatementAndCodeWordCircuit struct {
	LinComb []gnarkfext.E4Gen
	Ys      [][]gnarkfext.E4Gen
	X       gnarkfext.E4Gen
	Alpha   gnarkfext.E4Gen
	params  Params
}

func (c *StatementAndCodeWordCircuit) Define(api frontend.API) error {
	return GnarkCheckStatementAndCodeWord(api, c.params, c.LinComb, c.Ys, c.X, c.Alpha)
}

func GenerateStatementAndCodeWordWitness(size, rate int) (*StatementAndCodeWordCircuit, *StatementAndCodeWordCircuit) {
	sizeCodeWord := size * rate
	rsParams := reedsolomon.NewRsParams(size, rate)

	// Generate a valid codeword (linComb)
	// Start with coefficients in canonical form (last entries are zero for RS codeword)
	coeffs := make([]fext.Element, sizeCodeWord)
	for i := 0; i < size; i++ {
		coeffs[i].SetRandom()
	}
	// FFT to get Lagrange basis representation
	d := fft.NewDomain(uint64(sizeCodeWord))
	d.FFTExt(coeffs, fft.DIF)
	utils.BitReverse(coeffs)

	// linComb is now a valid RS codeword in Lagrange basis
	linComb := make([]fext.Element, sizeCodeWord)
	copy(linComb, coeffs)

	// Generate random evaluation point x and alpha
	var x, alpha fext.Element
	x.SetRandom()
	alpha.SetRandom()

	// Compute P(x) where P is the polynomial represented by linComb in Lagrange basis
	// P(x) = Σᵢ Lᵢ(x) * linComb[i]
	linCombSv := sv.NewRegularExt(linComb)
	pXSlice := smartvectors_mixed.BatchEvaluateLagrange([]sv.SmartVector{linCombSv}, x)
	pX := pXSlice[0]

	// For the statement check to pass, we need:
	// P(x) == eval(yjoined, alpha) where yjoined is the concatenation of all ys
	//
	// We construct ys such that yjoined evaluated at alpha equals pX
	// Simplest approach: ys has one element [pX], so eval([pX], alpha) = pX
	ys := [][]fext.Element{{pX}}

	// Create circuit and witness
	var circuit, witness StatementAndCodeWordCircuit

	// Circuit (structure only)
	circuit.LinComb = make([]gnarkfext.E4Gen, sizeCodeWord)
	circuit.Ys = make([][]gnarkfext.E4Gen, len(ys))
	for i := range ys {
		circuit.Ys[i] = make([]gnarkfext.E4Gen, len(ys[i]))
	}
	circuit.params = Params{RsParams: rsParams}

	// Witness (actual values)
	witness.LinComb = make([]gnarkfext.E4Gen, sizeCodeWord)
	for i := 0; i < sizeCodeWord; i++ {
		witness.LinComb[i] = gnarkfext.NewE4Gen(linComb[i])
	}
	witness.Ys = make([][]gnarkfext.E4Gen, len(ys))
	for i := range ys {
		witness.Ys[i] = make([]gnarkfext.E4Gen, len(ys[i]))
		for j := range ys[i] {
			witness.Ys[i][j] = gnarkfext.NewE4Gen(ys[i][j])
		}
	}
	witness.X = gnarkfext.NewE4Gen(x)
	witness.Alpha = gnarkfext.NewE4Gen(alpha)
	witness.params = Params{RsParams: rsParams}

	return &circuit, &witness
}

func TestStatementAndCodeWord(t *testing.T) {
	size := 64
	rate := 2

	// native
	{
		circuit, witness := GenerateStatementAndCodeWordWitness(size, rate)

		ccs, err := frontend.CompileU32(koalabear.Modulus(), scs.NewBuilder, circuit)
		assert.NoError(t, err)

		fullWitness, err := frontend.NewWitness(witness, koalabear.Modulus())
		assert.NoError(t, err)
		err = ccs.IsSolved(fullWitness)
		assert.NoError(t, err)
	}
	// emulated
	{
		circuit, witness := GenerateStatementAndCodeWordWitness(size, rate)

		ccs, err := frontend.Compile(ecc.BLS12_377.ScalarField(), scs.NewBuilder, circuit)
		assert.NoError(t, err)

		fullWitness, err := frontend.NewWitness(witness, ecc.BLS12_377.ScalarField())
		assert.NoError(t, err)
		err = ccs.IsSolved(fullWitness)
		assert.NoError(t, err)
	}
}
