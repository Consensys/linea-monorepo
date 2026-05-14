package vortex

import (
	"errors"
	"fmt"
	"slices"

	"github.com/consensys/linea-monorepo/prover-ray/crypto/koalabear/reedsolomon"
	"github.com/consensys/linea-monorepo/prover-ray/maths/koalabear/field"
	"github.com/consensys/linea-monorepo/prover-ray/maths/koalabear/polynomials"
)

var (
	// ErrInvalidVerifierInputs flags a verifier input with invalid dimensions.
	ErrInvalidVerifierInputs = errors.New("invalid verification input")
)

// VerifierInput holds the public inputs needed by the verifier.
type VerifierInput struct {
	// RsParams *reedsolomon.RsParams

	// alpha random coin used for the linear combination
	Alpha field.Ext

	// X is the univariate evaluation point
	X field.Ext

	// Ys are the alleged evaluation at point X
	Ys [][]field.Ext

	// EntryList is the random coin representing the columns to open.
	EntryList []int
}

// CheckStatement verifies that the linear combination is consistent with the claimed evaluations ys at point x.
func CheckStatement(linComb []field.Ext, ys [][]field.Ext, x, alpha field.Ext) error {

	// Check the consistency of Ys and proof.Linear combination
	yJoined := slices.Concat(ys...)
	alphaY := polynomials.EvalLagrange(field.VecFromExt(linComb), field.ElemFromExt(x))
	alphaYPrime := evalFextPolyHorner(yJoined, alpha)

	if !alphaY.Equal(&alphaYPrime) {
		return fmt.Errorf("RowLincomb and Y are inconsistent")
	}

	return nil
}

// CheckIsCodeWord returns nil iff v is a valid Reed-Solomon codeword.
func CheckIsCodeWord(rsParams *reedsolomon.RsParams, v []field.Ext) error {
	return rsParams.IsCodewordExt(v)
}

// CheckLinComb verifies the linear combination of opened columns against the proof.
func CheckLinComb(
	linComb []field.Ext,
	entryList []int,
	alpha field.Ext,
	columns [][][]field.Element,
) (err error) {

	numCommitments := len(columns)

	for j, selectedColID := range entryList {
		// Will carry the concatenation of the columns for the same entry j
		fullCol := []field.Element{}

		for i := range numCommitments {
			// Entries of the selected columns #j contained in the commitment #i.
			fullCol = append(fullCol, columns[i][j]...)
		}

		// Check the linear combination is consistent with the opened column
		y := evalBasePolyHorner(fullCol, alpha)
		other := linComb[selectedColID]

		if !y.Equal(&other) {
			return fmt.Errorf("the linear combination is inconsistent %v : %v", y.String(), other.String())
		}
	}

	return nil
}

// evalFextPolyHorner evaluates a polynomial whose coefficients live in the
// extension field, at an extension point, using Horner's method.
func evalFextPolyHorner(poly []field.Ext, x field.Ext) field.Ext {
	var res field.Ext
	for i := len(poly) - 1; i >= 0; i-- {
		res.Mul(&res, &x)
		res.Add(&res, &poly[i])
	}
	return res
}

// evalBasePolyHorner evaluates a base-coefficient polynomial at an extension
// point. Each addition only touches the lifted constant slot of res.
func evalBasePolyHorner(poly []field.Element, x field.Ext) field.Ext {
	var res field.Ext
	for i := len(poly) - 1; i >= 0; i-- {
		res.Mul(&res, &x)
		res.B0.A0.Add(&res.B0.A0, &poly[i])
	}
	return res
}
