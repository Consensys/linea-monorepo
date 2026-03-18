package vortex

import (
	"testing"

	"github.com/consensys/gnark-crypto/ecc"
	"github.com/consensys/gnark-crypto/field/koalabear"
	koalaVortex "github.com/consensys/gnark-crypto/field/koalabear/vortex"
	"github.com/consensys/gnark/frontend"
	"github.com/consensys/gnark/frontend/cs/scs"
	"github.com/consensys/linea-monorepo/prover/crypto/fiatshamir"
	"github.com/consensys/linea-monorepo/prover/crypto/reedsolomon"
	"github.com/consensys/linea-monorepo/prover/maths/field"
	"github.com/consensys/linea-monorepo/prover/maths/field/fext"
	"github.com/consensys/linea-monorepo/prover/maths/field/koalagnark"
	"github.com/stretchr/testify/assert"
)

type StatementAndCodeWordCircuit struct {
	LinComb []koalagnark.Ext
	Evals   []koalagnark.Ext
	Ys      [][]koalagnark.Ext
	X       koalagnark.Ext
	Alpha   koalagnark.Ext
	params  Params
}

func (c *StatementAndCodeWordCircuit) Define(api frontend.API) error {
	var fs fiatshamir.GnarkFS
	if api.Compiler().Field().Cmp(field.Modulus()) == 0 {
		fs = fiatshamir.NewGnarkFSKoalabear(api)
	} else {
		fs = fiatshamir.NewGnarkFSBLS12377(api)
	}
	return GnarkCheckStatementAndCodeWord(api, fs, c.params, c.LinComb, c.Evals, c.Ys, c.X, c.Alpha)
}

func GenerateStatementAndCodeWordWitness(size, rate int) (*StatementAndCodeWordCircuit, *StatementAndCodeWordCircuit) {
	rsParams := reedsolomon.NewRsParams(size, rate)
	sizeCodeWord := size * rate

	// Generate T random coefficients
	coeffs := make([]fext.Element, size)
	for i := range coeffs {
		coeffs[i].SetRandom()
	}

	// Forward FFT to get N evaluations on the RS domain
	evals := rsParams.ExtCoefficientsToAllEvaluations(coeffs)

	// Generate random evaluation point x and alpha
	var x, alpha fext.Element
	x.SetRandom()
	alpha.SetRandom()

	// P(x) = Horner(coeffs, x)
	pX := koalaVortex.EvalFextPolyHorner(coeffs, x)

	// ys: single element [pX], so Horner([pX], alpha) = pX
	ys := [][]fext.Element{{pX}}

	// Create circuit and witness
	var circuit, witness StatementAndCodeWordCircuit

	circuit.LinComb = make([]koalagnark.Ext, size)
	circuit.Evals = make([]koalagnark.Ext, sizeCodeWord)
	circuit.Ys = make([][]koalagnark.Ext, len(ys))
	for i := range ys {
		circuit.Ys[i] = make([]koalagnark.Ext, len(ys[i]))
	}
	circuit.params = Params{RsParams: rsParams}

	witness.LinComb = make([]koalagnark.Ext, size)
	for i := range coeffs {
		witness.LinComb[i] = koalagnark.NewExt(coeffs[i])
	}
	witness.Evals = make([]koalagnark.Ext, sizeCodeWord)
	for i := range evals {
		witness.Evals[i] = koalagnark.NewExt(evals[i])
	}
	witness.Ys = make([][]koalagnark.Ext, len(ys))
	for i := range ys {
		witness.Ys[i] = make([]koalagnark.Ext, len(ys[i]))
		for j := range ys[i] {
			witness.Ys[i][j] = koalagnark.NewExt(ys[i][j])
		}
	}
	witness.X = koalagnark.NewExt(x)
	witness.Alpha = koalagnark.NewExt(alpha)
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
