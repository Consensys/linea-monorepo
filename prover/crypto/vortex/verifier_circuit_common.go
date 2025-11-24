package vortex

import (
	"github.com/consensys/gnark/frontend"
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

	err = GnarkCheckStatement(api, params, proof.LinearCombination, vi.Ys, vi.X, vi.Alpha)

	return err
}

func GnarkCheckStatement(api frontend.API, params Params, linComb []gnarkfext.E4Gen,

	ys [][]gnarkfext.E4Gen, x, alpha gnarkfext.E4Gen) error {

	var yjoined []gnarkfext.E4Gen
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

	apiGen, err := zk.NewGenericApi(api)
	if err != nil {
		return err
	}
	apiGen.AssertIsEqual(alphaY.B0.A0, alphaYPrime.B0.A0)
	apiGen.AssertIsEqual(alphaY.B0.A1, alphaYPrime.B0.A1)
	apiGen.AssertIsEqual(alphaY.B1.A0, alphaYPrime.B1.A0)
	apiGen.AssertIsEqual(alphaY.B1.A1, alphaYPrime.B1.A1)

	return nil
}

// Put that in vortex common
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
