package vortex

import (
	"github.com/consensys/gnark/frontend"
	"github.com/consensys/gnark/std/lookup/logderivlookup"
	"github.com/consensys/linea-monorepo/prover/crypto/fiatshamir"
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

// GnarkVerify verifies the vortex opening for coefficient-mode U_alpha.
// proof.LinearCombination holds T polynomial coefficients (E4).
// A forward FFT hint reconstructs N evaluations for the column-consistency lookup check.
func GnarkVerify(api frontend.API, fs fiatshamir.GnarkFS, params Params, proof GnarkProof, vi GnarkVerifierInput) error {
	koalaAPI := koalagnark.NewAPI(api)

	// Expand T coefficients → N evaluations via forward FFT hint.
	t := len(proof.LinearCombination)
	n := params.RsParams.NbEncodedColumns()
	fftfwd := fftFwdHint(koalaAPI.Type())
	inputsUnpacked := make([]koalagnark.Element, t*4)
	for i := 0; i < t; i++ {
		inputsUnpacked[4*i] = proof.LinearCombination[i].B0.A0
		inputsUnpacked[4*i+1] = proof.LinearCombination[i].B0.A1
		inputsUnpacked[4*i+2] = proof.LinearCombination[i].B1.A0
		inputsUnpacked[4*i+3] = proof.LinearCombination[i].B1.A1
	}
	evalsRaw, err := koalaAPI.NewHint(fftfwd, n*4, inputsUnpacked...)
	if err != nil {
		return err
	}
	evals := make([]koalagnark.Ext, n)
	for i := range evals {
		evals[i].B0.A0 = evalsRaw[4*i]
		evals[i].B0.A1 = evalsRaw[4*i+1]
		evals[i].B1.A0 = evalsRaw[4*i+2]
		evals[i].B1.A1 = evalsRaw[4*i+3]
	}

	// Column-consistency check: lookup evals[entryList[j]] == Horner(columns[j], alpha)
	err = GnarkCheckLinComb(api, evals, vi.EntryList, vi.Alpha, proof.Columns)
	if err != nil {
		return err
	}

	// Codeword + statement check using coefficients directly.
	return GnarkCheckStatementAndCodeWord(api, fs, params, proof.LinearCombination, evals, vi.Ys, vi.X, vi.Alpha)
}

// GnarkCheckStatementAndCodeWord combines the Schwartz-Zippel codeword check
// and statement check for coefficient-mode U_alpha.
// linComb holds T coefficients; evals holds N evaluations (from forward FFT hint).
func GnarkCheckStatementAndCodeWord(
	api frontend.API,
	fs fiatshamir.GnarkFS,
	params Params,
	linComb []koalagnark.Ext,
	evals []koalagnark.Ext,
	ys [][]koalagnark.Ext,
	x, alpha koalagnark.Ext) error {

	koalaAPI := koalagnark.NewAPI(api)

	// Bind T coefficients into the Fiat-Shamir transcript (same role as res in eval mode).
	fs.UpdateExt(linComb...)
	challenge := fs.RandomFieldExt()

	// Schwartz-Zippel: evalCan(linComb, challenge) == evalLag(evals, challenge)
	evalCan := polynomials.GnarkEvalCanonicalExt(api, linComb, challenge)
	evalLag := polynomials.GnarkEvaluateLagrangeExt(
		api,
		evals,
		challenge,
		params.RsParams.Domains[1].Generator,
		params.RsParams.Domains[1].Cardinality)
	koalaAPI.AssertIsEqualExt(evalCan, evalLag)

	// Statement check: evalCan(linComb, x) == Horner(ys_joined, alpha)
	alphaY := polynomials.GnarkEvalCanonicalExt(api, linComb, x)
	var yjoined []koalagnark.Ext
	for i := range ys {
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
	table1 := logderivlookup.New(api)
	table2 := logderivlookup.New(api)
	table3 := logderivlookup.New(api)
	table4 := logderivlookup.New(api)

	for k := range linComb {
		table1.Insert(linComb[k].B0.A0.Native())
		table2.Insert(linComb[k].B0.A1.Native())
		table3.Insert(linComb[k].B1.A0.Native())
		table4.Insert(linComb[k].B1.A1.Native())
	}

	v1 := table1.Lookup(entryList...)
	v2 := table2.Lookup(entryList...)
	v3 := table3.Lookup(entryList...)
	v4 := table4.Lookup(entryList...)

	// Construct the lookup results
	lookedUpValues := make([]koalagnark.Ext, len(entryList))
	for j := range entryList {
		lookedUpValues[j].B0.A0 = koalagnark.WrapFrontendVariable(v1[j])
		lookedUpValues[j].B0.A1 = koalagnark.WrapFrontendVariable(v2[j])
		lookedUpValues[j].B1.A0 = koalagnark.WrapFrontendVariable(v3[j])
		lookedUpValues[j].B1.A1 = koalagnark.WrapFrontendVariable(v4[j])
	}

	// Compare with ys
	for j := range entryList {
		koalaAPI.AssertIsEqualExt(ys[j], lookedUpValues[j])
	}

	return nil
}
