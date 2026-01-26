package vortex

import (
	"github.com/consensys/gnark/frontend"
	"github.com/consensys/linea-monorepo/prover/maths/field/fext"
	"github.com/consensys/linea-monorepo/prover/maths/field/koalagnark"
	"github.com/consensys/linea-monorepo/prover/maths/polynomials"
)

type GnarkProof struct {
	Columns           [][][]koalagnark.Element
	LinearCombination []koalagnark.Ext
}

type GnarkVerifierInput struct {

	// alpha random coin used for the linear combination
	Alpha koalagnark.Ext

	// X is the univariate evaluation point
	X koalagnark.Ext

	// Ys are the alleged evaluation at point X
	Ys [][]koalagnark.Ext

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
func GnarkCheckStatementAndCodeWord(api frontend.API, params Params, linComb []koalagnark.Ext,
	ys [][]koalagnark.Ext, x, alpha koalagnark.Ext) error {

	koalaAPI := koalagnark.NewAPI(api)

	// === Part 1: Prepare for codeword check (compute FFT inverse via hint) ===
	fftinv := fftHint(koalaAPI.Type())
	sizeFextUnpacked := len(linComb) * 4
	inputs := make([]koalagnark.Element, sizeFextUnpacked)
	for i := 0; i < len(linComb); i++ {
		inputs[4*i] = linComb[i].B0.A0
		inputs[4*i+1] = linComb[i].B0.A1
		inputs[4*i+2] = linComb[i].B1.A0
		inputs[4*i+3] = linComb[i].B1.A1
	}
	_res, err := koalaAPI.NewHint(fftinv, sizeFextUnpacked, inputs...)
	if err != nil {
		return err
	}

	res := make([]koalagnark.Ext, len(linComb))
	for i := 0; i < len(linComb); i++ {
		res[i].B0.A0 = _res[4*i]
		res[i].B0.A1 = _res[4*i+1]
		res[i].B1.A0 = _res[4*i+2]
		res[i].B1.A1 = _res[4*i+3]
	}

	var c fext.Element
	c.SetRandom()
	challenge := koalagnark.NewExt(c)

	// === Part 2: Codeword check (Schwartz-Zippel) ===
	// Evaluate linComb at challenge (for codeword check)
	evalLag := polynomials.GnarkEvaluateLagrangeExt(
		api,
		linComb,
		challenge,
		params.RsParams.Domains[1].Generator,
		params.RsParams.Domains[1].Cardinality)

	evalCan := polynomials.GnarkEvalCanonicalExt(api, res, challenge)
	koalaAPI.AssertIsEqualExt(evalLag, evalCan)

	// === Part 3: Assert last entries are zeroes (RS codeword property) ===
	zero := koalaAPI.ZeroExt()
	for i := params.RsParams.NbColumns(); i < params.RsParams.NbEncodedColumns(); i++ {
		koalaAPI.AssertIsEqualExt(res[i], zero)
	}

	// === Part 4: Statement check ===
	alphaY := polynomials.GnarkEvalCanonicalExt(api, res, x)

	var yjoined []koalagnark.Ext
	for i := 0; i < len(ys); i++ {
		yjoined = append(yjoined, ys[i]...)
	}
	alphaYPrime := polynomials.GnarkEvalCanonicalExt(api, yjoined, alpha)
	koalaAPI.AssertIsEqualExt(alphaY, alphaYPrime)

	return nil
}

func GnarkCheckLinComb(
	api frontend.API, linComb []koalagnark.Ext,
	entryList []frontend.Variable, alpha koalagnark.Ext,
	columns [][][]koalagnark.Element) error {

	koalaAPI := koalagnark.NewAPI(api)

	numCommitments := len(columns)

	// Prepare all polynomials for batch evaluation at alpha ===
	// Collect all fullCol polynomials first
	fullCols := make([][]koalagnark.Element, len(entryList))

	for j := range entryList {

		// Will carry the concatenation of the columns for the same entry j
		fullCol := []koalagnark.Element{}

		for i := range numCommitments {
			// Entries of the selected columns #j contained in the commitment #i.
			fullCol = append(fullCol, columns[i][j]...)
		}

		fullCols[j] = fullCol
	}
	ys := polynomials.GnarkEvalCanonicalBatch(api, fullCols, alpha)

	for j, selectedColID := range entryList {
		// Check the linear combination is consistent with the opened column

		// check that y := linComb[selectedColID] coords by coords
		table := make([]koalagnark.Element, len(linComb))
		for k := 0; k < len(linComb); k++ {
			table[k] = linComb[k].B0.A0
		}
		v := koalaAPI.Mux(selectedColID, table...)
		koalaAPI.AssertIsEqual(ys[j].B0.A0, v)

		for k := 0; k < len(linComb); k++ {
			table[k] = linComb[k].B0.A1
		}
		v = koalaAPI.Mux(selectedColID, table...)
		koalaAPI.AssertIsEqual(ys[j].B0.A1, v)

		for k := 0; k < len(linComb); k++ {
			table[k] = linComb[k].B1.A0
		}
		v = koalaAPI.Mux(selectedColID, table...)
		koalaAPI.AssertIsEqual(ys[j].B1.A0, v)

		for k := 0; k < len(linComb); k++ {
			table[k] = linComb[k].B1.A1
		}
		v = koalaAPI.Mux(selectedColID, table...)
		koalaAPI.AssertIsEqual(ys[j].B1.A1, v)
	}

	return nil
}
