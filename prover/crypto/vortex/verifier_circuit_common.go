package vortex

import (
	"github.com/consensys/gnark/frontend"
	"github.com/consensys/linea-monorepo/prover/maths/field/koalagnark"
	"github.com/consensys/linea-monorepo/prover/maths/polynomials"
)

type GnarkProof struct {
	Columns                  [][][]koalagnark.Element
	LinearCombination        []koalagnark.Ext
	EncodedLinearCombination []koalagnark.Ext
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

	err := GnarkCheckLinComb(api, proof.EncodedLinearCombination, vi.EntryList, vi.Alpha, proof.Columns)
	if err != nil {
		return err
	}

	// Batch the two Lagrange evaluations (GnarkCheckStatement and GnarkCheckIsCodeWord)
	// since they both evaluate the same polynomial on the same domain
	err = GnarkCheckStatementAndCodeWord(api, params, proof.LinearCombination, vi.Ys, vi.X, vi.Alpha)
	return err
}

// GnarkCheckStatementAndCodeWord checks that the linear combination is consistent with the claimed evaluations.
// linComb is in Lagrange basis and evaluated using Lagrange interpolation.
// ys is in canonical (coefficient) form and evaluated using Horner's method.
func GnarkCheckStatementAndCodeWord(api frontend.API, params Params, linComb []koalagnark.Ext,
	ys [][]koalagnark.Ext, x, alpha koalagnark.Ext) error {

	koalaAPI := koalagnark.NewAPI(api)

	// Evaluate linComb at x using Lagrange interpolation (linComb is in Lagrange basis)
	alphaY := polynomials.GnarkEvaluateLagrangeExt(
		api,
		linComb,
		x,
		params.RsParams.Domains[0].Generator,
		params.RsParams.Domains[0].Cardinality)

	// Join ys and evaluate at alpha using Horner's method (ys is in canonical form)
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

	for j, selectedColID := range entryList {

		// Will carry the concatenation of the columns for the same entry j
		fullCol := []koalagnark.Element{}

		for i := range numCommitments {
			// Entries of the selected columns #j contained in the commitment #i.
			fullCol = append(fullCol, columns[i][j]...)
		}

		// Check the linear combination is consistent with the opened column
		y := polynomials.GnarkEvalCanonical(api, fullCol, alpha)

		// check that y := linComb[selectedColID] coords by coords
		table := make([]koalagnark.Element, len(linComb))
		for k := 0; k < len(linComb); k++ {
			table[k] = linComb[k].B0.A0
		}
		v := koalaAPI.Mux(selectedColID, table...)
		koalaAPI.AssertIsEqual(y.B0.A0, v)

		for k := 0; k < len(linComb); k++ {
			table[k] = linComb[k].B0.A1
		}
		v = koalaAPI.Mux(selectedColID, table...)
		koalaAPI.AssertIsEqual(y.B0.A1, v)

		for k := 0; k < len(linComb); k++ {
			table[k] = linComb[k].B1.A0
		}
		v = koalaAPI.Mux(selectedColID, table...)
		koalaAPI.AssertIsEqual(y.B1.A0, v)

		for k := 0; k < len(linComb); k++ {
			table[k] = linComb[k].B1.A1
		}
		v = koalaAPI.Mux(selectedColID, table...)
		koalaAPI.AssertIsEqual(y.B1.A1, v)
	}

	return nil
}
