package vortex

import (
	"github.com/consensys/gnark/frontend"
	"github.com/consensys/linea-monorepo/prover/crypto/reedsolomon"
	"github.com/consensys/linea-monorepo/prover/maths/field/fext"
	"github.com/consensys/linea-monorepo/prover/maths/field/gnarkfext"
	"github.com/consensys/linea-monorepo/prover/maths/polynomials"
	"github.com/consensys/linea-monorepo/prover/maths/zk"
)

type GnarkProof struct {
	Columns           [][][]zk.WrappedVariable
	LinearCombination []gnarkfext.E4Gen
}

type GnarkVerifierInput struct {

	// alpha random coin used for the linear combination
	Alpha gnarkfext.E4Gen

	// X is the univariate evaluation point
	X gnarkfext.E4Gen

	// Ys are the alleged evaluation at point X
	Ys [][]gnarkfext.E4Gen

	// EntryList is the random coin representing the columns to open.
	EntryList []frontend.Variable
}

func GnarkVerify(api frontend.API, params Params, proof GnarkProof, vi GnarkVerifierInput, roots []frontend.Variable) error {

	err := GnarkCheckLinComb(api, proof.LinearCombination, vi.EntryList, vi.Alpha, proof.Columns)
	if err != nil {
		return err
	}

	// Batch the two Lagrange evaluations (GnarkCheckStatement and GnarkCheckIsCodeWord)
	// since they both evaluate the same polynomial on the same domain
	err = GnarkCheckStatementAndCodeWord(api, params, proof.LinearCombination, vi.Ys, vi.X, vi.Alpha)
	return err
}

// GnarkCheckStatementAndCodeWord combines GnarkCheckStatement and GnarkCheckIsCodeWord
// to share the Lagrange basis computation, saving approximately n multiplications.
func GnarkCheckStatementAndCodeWord(api frontend.API, params Params, linComb []gnarkfext.E4Gen,
	ys [][]gnarkfext.E4Gen, x, alpha gnarkfext.E4Gen) error {

	apiGen, err := zk.NewGenericApi(api)
	if err != nil {
		return err
	}

	ext4, err := gnarkfext.NewExt4(api)
	if err != nil {
		return err
	}

	// === Part 1: Prepare for codeword check (compute FFT inverse via hint) ===
	fftinv := fftHint(apiGen.Type())
	sizeFextUnpacked := len(linComb) * 4
	inputs := make([]zk.WrappedVariable, sizeFextUnpacked)
	for i := 0; i < len(linComb); i++ {
		inputs[4*i] = linComb[i].B0.A0
		inputs[4*i+1] = linComb[i].B0.A1
		inputs[4*i+2] = linComb[i].B1.A0
		inputs[4*i+3] = linComb[i].B1.A1
	}
	_res, err := apiGen.NewHint(fftinv, sizeFextUnpacked, inputs...)
	if err != nil {
		return err
	}

	res := make([]gnarkfext.E4Gen, len(linComb))
	for i := 0; i < len(linComb); i++ {
		res[i].B0.A0 = _res[4*i]
		res[i].B0.A1 = _res[4*i+1]
		res[i].B1.A0 = _res[4*i+2]
		res[i].B1.A1 = _res[4*i+3]
	}

	// === Part 2: Batch Lagrange evaluation at two points ===
	// Both evaluations use the same domain (Domains[1])
	var c fext.Element
	c.SetRandom()
	challenge := gnarkfext.NewE4Gen(c)

	// Batch evaluate linComb at both x (for statement check) and challenge (for codeword check)
	zs := []gnarkfext.E4Gen{x, challenge}
	evals := polynomials.GnarkEvaluateLagrangeExtBatch(
		api,
		linComb,
		zs,
		params.RsParams.Domains[1].Generator,
		params.RsParams.Domains[1].Cardinality)

	alphaY := evals[0]      // P(x) for statement check
	evalLag := evals[1]     // P(challenge) for codeword check

	// === Part 3: Statement check ===
	var yjoined []gnarkfext.E4Gen
	for i := 0; i < len(ys); i++ {
		yjoined = append(yjoined, ys[i]...)
	}
	alphaYPrime := polynomials.GnarkEvalCanonicalExt(api, yjoined, alpha)
	ext4.AssertIsEqual(&alphaY, &alphaYPrime)

	// === Part 4: Codeword check (Schwartz-Zippel) ===
	evalCan := polynomials.GnarkEvalCanonicalExt(api, res, challenge)
	apiGen.AssertIsEqual(evalLag.B0.A0, evalCan.B0.A0)
	apiGen.AssertIsEqual(evalLag.B0.A1, evalCan.B0.A1)
	apiGen.AssertIsEqual(evalLag.B1.A0, evalCan.B1.A0)
	apiGen.AssertIsEqual(evalLag.B1.A1, evalCan.B1.A1)

	// === Part 5: Assert last entries are zeroes (RS codeword property) ===
	zero := zk.ValueOf(0)
	for i := params.RsParams.NbColumns(); i < params.RsParams.NbEncodedColumns(); i++ {
		apiGen.AssertIsEqual(res[i].B0.A0, zero)
		apiGen.AssertIsEqual(res[i].B0.A1, zero)
		apiGen.AssertIsEqual(res[i].B1.A0, zero)
		apiGen.AssertIsEqual(res[i].B1.A1, zero)
	}

	return nil
}

func GnarkCheckIsCodeWord(api frontend.API, params reedsolomon.RsParams, linComb []gnarkfext.E4Gen) error {
	// This function is no longer used and can be removed or refactored if needed.
	return nil
}

func GnarkCheckLinComb(
	api frontend.API, linComb []gnarkfext.E4Gen,
	entryList []frontend.Variable, alpha gnarkfext.E4Gen,
	columns [][][]zk.WrappedVariable) error {

	apiGen, err := zk.NewGenericApi(api)
	if err != nil {
		return err
	}

	numCommitments := len(columns)

	for j, selectedColID := range entryList {

		// Will carry the concatenation of the columns for the same entry j
		fullCol := []zk.WrappedVariable{}

		for i := range numCommitments {
			// Entries of the selected columns #j contained in the commitment #i.
			fullCol = append(fullCol, columns[i][j]...)
		}

		// Check the linear combination is consistent with the opened column
		y := polynomials.GnarkEvalCanonical(api, fullCol, alpha)

		// check that y := linComb[selectedColID] coords by coords
		table := make([]zk.WrappedVariable, len(linComb))
		for k := 0; k < len(linComb); k++ {
			table[k] = linComb[k].B0.A0
		}
		v := apiGen.Mux(selectedColID, table...)
		apiGen.AssertIsEqual(y.B0.A0, v)

		for k := 0; k < len(linComb); k++ {
			table[k] = linComb[k].B0.A1
		}
		v = apiGen.Mux(selectedColID, table...)
		apiGen.AssertIsEqual(y.B0.A1, v)

		for k := 0; k < len(linComb); k++ {
			table[k] = linComb[k].B1.A0
		}
		v = apiGen.Mux(selectedColID, table...)
		apiGen.AssertIsEqual(y.B1.A0, v)

		for k := 0; k < len(linComb); k++ {
			table[k] = linComb[k].B1.A1
		}
		v = apiGen.Mux(selectedColID, table...)
		apiGen.AssertIsEqual(y.B1.A1, v)
	}

	return nil
}
