package vortex

import (
	"errors"
	"fmt"
	"slices"

	"github.com/consensys/gnark-crypto/field/koalabear/vortex"
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
	alphaYPrime := vortex.EvalFextPolyHorner(yJoined, alpha)

	if !alphaY.Equal(&alphaYPrime) {
		return fmt.Errorf("RowLincomb and Y are inconsistent")
	}

	return nil
}

// VerifierInputMultilinear holds the public inputs for multilinear-evaluation
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

// CheckStatementMultilinear verifies that the linear combination is consistent
// with the claimed multilinear evaluations ys at the point h.
//
// It decodes the N systematic Lagrange values from the encoded linComb, then
// checks EvalMultilin(decoded, h) == EvalFextPolyHorner(ys_joined, alpha).
func CheckStatementMultilinear(rsParams *reedsolomon.RsParams,
	linComb []field.Ext, ys [][]field.Ext, h []field.Ext, alpha field.Ext,
) error {
	decoded := rsParams.SystematicExt(linComb)

	coords := make([]field.Gen, len(h))
	for i, hi := range h {
		coords[i] = field.ElemFromExt(hi)
	}

	lhs := polynomials.EvalMultilin(field.VecFromExt(decoded), coords).AsExt()

	yJoined := slices.Concat(ys...)
	rhs := vortex.EvalFextPolyHorner(yJoined, alpha)

	if !lhs.Equal(&rhs) {
		return fmt.Errorf("multilinear statement check failed: EvalMultilin=%v != Horner=%v", lhs, rhs)
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
		y := vortex.EvalBasePolyHorner(fullCol, alpha)
		other := linComb[selectedColID]

		if y != other {
			return fmt.Errorf("the linear combination is inconsistent %v : %v", y.String(), other.String())
		}
	}

	return nil
}
