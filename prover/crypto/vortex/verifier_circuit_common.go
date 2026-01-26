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

	err := GnarkCheckLinComb(api, params, proof.LinearCombination, vi.EntryList, vi.Alpha, proof.Columns)
	if err != nil {
		return err
	}

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
	api frontend.API, params Params, linComb []koalagnark.Ext,
	entryList []frontend.Variable, alpha koalagnark.Ext,
	columns [][][]koalagnark.Element) error {

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

	// gen is a base field element (generator of the multiplicative subgroup)
	gen := koalagnark.NewElement(params.RsParams.Domains[1].Generator)

	numCommitments := len(columns)
	numEntries := len(entryList)
	api.Println("entryList:", entryList)

	// === Part 2: Batch compute all z = gen^selectedColID values ===
	// Using ExpBaseVar since gen is a base field element (much cheaper than ExpExtVar)
	zs := make([]koalagnark.Element, numEntries)
	for j, selectedColID := range entryList {
		zs[j] = koalaAPI.ExpBaseVar(gen, selectedColID, 32)
	}

	// === Part 3: Batch evaluate res at all z points ===
	us := polynomials.GnarkEvalCanonicalExtBatch(api, res, zs)

	// === Part 4: Prepare all polynomials for batch evaluation at alpha ===
	// Collect all fullCol polynomials first
	fullCols := make([][]koalagnark.Element, numEntries)
	for j := range entryList {
		// Will carry the concatenation of the columns for the same entry j
		fullCol := []koalagnark.Element{}

		for i := range numCommitments {
			// Entries of the selected columns #j contained in the commitment #i.
			fullCol = append(fullCol, columns[i][j]...)
		}

		fullCols[j] = fullCol
	}

	// Batch compute all y values at once
	ys := polynomials.GnarkEvalCanonicalBatch(api, fullCols, alpha)

	// Assert all equalities
	for j := range entryList {
		koalaAPI.AssertIsEqualExt(ys[j], us[j])
	}

	return nil
}
