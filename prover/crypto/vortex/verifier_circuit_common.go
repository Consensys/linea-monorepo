package vortex

import (
	"github.com/consensys/gnark/frontend"
	"github.com/consensys/linea-monorepo/prover/maths/field/gnarkfext"
	"github.com/consensys/linea-monorepo/prover/maths/polynomials"
	"github.com/consensys/linea-monorepo/prover/maths/zk"
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

	linearCombination := make([]gnarkfext.E4Gen, len(proof.LinearCombination))
	for i := 0; i < len(proof.LinearCombination); i++ {
		linearCombination[i] = gnarkfext.NewE4GenFromFrontendExt(proof.LinearCombination[i])
	}
	alpha := gnarkfext.NewE4GenFromFrontendExt(vi.Alpha)
	x := gnarkfext.NewE4GenFromFrontendExt(vi.X)

	ys := make([][]gnarkfext.E4Gen, len(vi.Ys))
	for i := 0; i < len(vi.Ys); i++ {
		ys[i] = make([]gnarkfext.E4Gen, len(vi.Ys[i]))
		for j := 0; j < len(vi.Ys[i]); j++ {
			ys[i][j] = gnarkfext.NewE4GenFromFrontendExt(vi.Ys[i][j])
		}
	}

	columns := make([][][]zk.WrappedVariable, len(proof.Columns))
	for i := 0; i < len(proof.Columns); i++ {
		columns[i] = make([][]zk.WrappedVariable, len(proof.Columns[i]))
		for j := 0; j < len(proof.Columns[i]); j++ {
			columns[i][j] = make([]zk.WrappedVariable, len(proof.Columns[i][j]))
			for k := 0; k < len(proof.Columns[i][j]); k++ {
				columns[i][j][k] = zk.WrapFrontendVariable(proof.Columns[i][j][k])
			}
		}
	}
	err := GnarkCheckLinComb(api, linearCombination, vi.EntryList, alpha, columns)
	if err != nil {
		return err
	}
	err = GnarkCheckStatement(api, params, linearCombination, ys, x, alpha)
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

	ext4, err := gnarkfext.NewExt4(api)
	if err != nil {
		return err
	}

	ext4.AssertIsEqual(&alphaY, &alphaYPrime)

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
