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
	// Alpha random coin used for the linear combination
	Alpha field.Ext

	// X is the univariate evaluation point
	X field.Ext

	// Ys are the alleged evaluation at point X
	Ys [][]field.Ext

	// EntryList is the random coin representing the columns to open.
	EntryList []int
}

// CheckStatement verifies that the linear combination U_alpha — given as T
// monomial coefficients — is consistent with the claimed evaluations ys at
// point x. Evaluates U_alpha(x) by Horner on the coefficients and checks it
// equals EvalCanonical(ys_joined, alpha).
func CheckStatement(coefficients []field.Ext, ys [][]field.Ext, x, alpha field.Ext) error {
	yJoined := slices.Concat(ys...)
	alphaY := polynomials.EvalCanonical(field.VecFromExt(coefficients), field.ElemFromExt(x)).AsExt()
	alphaYPrime := polynomials.EvalCanonical(field.VecFromExt(yJoined), field.ElemFromExt(alpha)).AsExt()

	if !alphaY.Equal(&alphaYPrime) {
		return fmt.Errorf("RowLincomb and Y are inconsistent")
	}

	return nil
}

// VerifierInputMultilinear holds the public inputs for the multilinear-evaluation
// variant of the Vortex verifier. It replaces the univariate point X with a
// multilinear evaluation point H.
type VerifierInputMultilinear struct {
	// Alpha is the random coin used for the linear combination.
	Alpha field.Ext

	// H is the multilinear evaluation point (len(H) = log2(NbColumns)).
	H []field.Ext

	// Ys[i][j] is the claimed multilinear evaluation of the j-th polynomial of
	// the i-th commitment at point H.
	Ys [][]field.Ext

	// EntryList is the random coin representing the columns to open.
	EntryList []int
}

// CheckStatementMultilinear verifies that the linear combination U_alpha —
// given as T monomial coefficients — is consistent with the claimed multilinear
// evaluations ys at point h.
//
// It recovers the T original row evaluations from the coefficients via the
// small-domain forward FFT (the inverse of LCEvalsToCoefficients), then checks
// EvalMultilin(rowEvals, h) == EvalCanonical(ys_joined, alpha).
func CheckStatementMultilinear(rsParams *reedsolomon.RsParams,
	coefficients []field.Ext, ys [][]field.Ext, h []field.Ext, alpha field.Ext,
) error {
	rowEvals := rsParams.EncodeFromMonomialsSmallDomain(coefficients)

	coords := make([]field.Gen, len(h))
	for i, hi := range h {
		coords[i] = field.ElemFromExt(hi)
	}

	lhs := polynomials.EvalMultilin(field.VecFromExt(rowEvals), coords).AsExt()

	yJoined := slices.Concat(ys...)
	rhs := polynomials.EvalCanonical(field.VecFromExt(yJoined), field.ElemFromExt(alpha)).AsExt()

	if !lhs.Equal(&rhs) {
		return fmt.Errorf("multilinear statement check failed: EvalMultilin=%v != EvalCanonical=%v", lhs, rhs)
	}
	return nil
}

// CheckLinComb verifies that U_alpha — given as T monomial coefficients —
// agrees with the alpha-Horner combination of the opened columns at every
// selected position. The verifier expands the coefficients to all N domain
// evaluations once via EncodeFromMonomials, then looks up U_alpha(ω^selectedColID).
func CheckLinComb(
	rsParams *reedsolomon.RsParams,
	coefficients []field.Ext,
	entryList []int,
	alpha field.Ext,
	columns [][][]field.Element,
) (err error) {

	evals := rsParams.EncodeFromMonomials(coefficients)
	numCommitments := len(columns)

	for j, selectedColID := range entryList {
		fullCol := []field.Element{}
		for i := range numCommitments {
			fullCol = append(fullCol, columns[i][j]...)
		}

		// U_alpha(ω^selectedColID) must equal α-Horner of the opened column
		y := polynomials.EvalCanonical(field.VecFromBase(fullCol), field.ElemFromExt(alpha)).AsExt()
		other := evals[selectedColID]

		if y != other {
			return fmt.Errorf("the linear combination is inconsistent %v : %v", y.String(), other.String())
		}
	}

	return nil
}
