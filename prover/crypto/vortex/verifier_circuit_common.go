package vortex

import (
	"github.com/consensys/gnark/frontend"
	"github.com/consensys/gnark/std/selector"
	"github.com/consensys/linea-monorepo/prover/maths/field/gnarkfext"
	"github.com/consensys/linea-monorepo/prover/maths/polynomials"
)

type GnarkProof struct {
	Columns           [][][]frontend.Variable
	LinearCombination []gnarkfext.Element
}

type GnarkVerifierInput struct {

	// alpha random coin used for the linear combination
	Alpha gnarkfext.Element

	// X is the univariate evaluation point
	X gnarkfext.Element

	// Ys are the alleged evaluation at point X
	Ys [][]gnarkfext.Element

	// EntryList is the random coin representing the columns to open.
	EntryList []frontend.Variable
}

func GnarkVerify(api frontend.API, params Params, proof GnarkProof, vi GnarkVerifierInput, roots []frontend.Variable) error {
	err := GnarkCheckLinComb(api, proof.LinearCombination, vi.EntryList, vi.Alpha, proof.Columns)
	if err != nil {
		return err
	}
	err = GnarkCheckStatement(api, params, proof.LinearCombination, vi.Ys, vi.X, vi.Alpha)
	return err
}

func GnarkCheckStatement(api frontend.API, params Params, linComb []gnarkfext.Element,

	ys [][]gnarkfext.Element, x, alpha gnarkfext.Element) error {

	var yjoined []gnarkfext.Element
	for i := 0; i < len(ys); i++ {
		yjoined = append(yjoined, ys[i]...)
	}

	alphaY := polynomials.GnarkEvaluateLagrangeExt(
		api,
		linComb,
		x,
		params.RsParams.Domains[1].Generator,
		params.RsParams.Domains[1].Cardinality)
	alphaYPrime := polynomials.GnarkEvalCanonicalExt(api, yjoined, alpha)

	gnarkfext.AssertIsEqual(api, alphaY, alphaYPrime)

	return nil
}

// Put that in vortex common
func GnarkCheckLinComb(
	api frontend.API, linComb []gnarkfext.Element,
	entryList []frontend.Variable, alpha gnarkfext.Element,
	columns [][][]frontend.Variable) error {

	numCommitments := len(columns)

	for j, selectedColID := range entryList {

		// Will carry the concatenation of the columns for the same entry j
		fullCol := []frontend.Variable{}

		for i := range numCommitments {
			// Entries of the selected columns #j contained in the commitment #i.
			fullCol = append(fullCol, columns[i][j]...)
		}

		// Check the linear combination is consistent with the opened column
		y := polynomials.GnarkEvalCanonical(api, fullCol, alpha)

		// check that y := linComb[selectedColID] coords by coords
		table := make([]frontend.Variable, len(linComb))
		for k := 0; k < len(linComb); k++ {
			table[k] = linComb[k].B0.A0
		}

		v := selector.Mux(api, selectedColID, table...)
		api.AssertIsEqual(y.B0.A0, v)

		for k := 0; k < len(linComb); k++ {
			table[k] = linComb[k].B0.A1
		}
		v = selector.Mux(api, selectedColID, table...)
		api.AssertIsEqual(y.B0.A1, v)

		for k := 0; k < len(linComb); k++ {
			table[k] = linComb[k].B1.A0
		}
		v = selector.Mux(api, selectedColID, table...)
		api.AssertIsEqual(y.B1.A0, v)

		for k := 0; k < len(linComb); k++ {
			table[k] = linComb[k].B1.A1
		}
		v = selector.Mux(api, selectedColID, table...)
		api.AssertIsEqual(y.B1.A1, v)
	}

	return nil
}
