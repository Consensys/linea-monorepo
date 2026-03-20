package vortex

import (
	"errors"
	"fmt"

	"github.com/consensys/gnark-crypto/field/koalabear/vortex"
	"github.com/consensys/linea-monorepo/prover/crypto/reedsolomon"
	"github.com/consensys/linea-monorepo/prover/maths/common/smartvectors"
	"github.com/consensys/linea-monorepo/prover/maths/field"
	"github.com/consensys/linea-monorepo/prover/maths/field/fext"
	"github.com/consensys/linea-monorepo/prover/utils"
)

var (
	// ErrInvalidVerifierInputs flags a verifier input with invalid dimensions.
	ErrInvalidVerifierInputs = errors.New("invalid verification input")
)

// VerifierInput
type VerifierInput struct {
	// RsParams *reedsolomon.RsParams

	// alpha random coin used for the linear combination
	Alpha fext.Element

	// X is the univariate evaluation point
	X fext.Element

	// Ys are the alleged evaluation at point X
	Ys [][]fext.Element

	// EntryList is the random coin representing the columns to open.
	EntryList []int
}

// CheckStatement evaluates the polynomial in coefficient form at x using Horner
// and checks consistency with the alleged evaluations ys folded at alpha.
func CheckStatement(coefficients smartvectors.SmartVector, ys [][]fext.Element, x, alpha fext.Element) error {
	yJoined := utils.Join(ys...)

	res := make([]fext.Element, coefficients.Len())
	coefficients.WriteInSliceExt(res)

	alphaY := vortex.EvalFextPolyHorner(res, x)
	alphaYPrime := vortex.EvalFextPolyHorner(yJoined, alpha)

	if alphaY != alphaYPrime {
		return fmt.Errorf("RowLincomb (coeff mode) and Y are inconsistent")
	}
	return nil
}

// CheckLinComb checks the linear combination in coefficient mode: evaluates all
// N RS domain points via a single FFT and verifies each queried column via
// a direct index lookup.
func CheckLinComb(
	coefficients smartvectors.SmartVector,
	entryList []int,
	alpha fext.Element,
	columns [][][]field.Element,
	rsParams *reedsolomon.RsParams,
) error {
	// One FFT to get all N evaluations
	allEvals := rsParams.ExtCoefficientsToAllEvaluations(coefficients)
	numCommitments := len(columns)

	for j, selectedColID := range entryList {
		// Build full column (concatenate across commitments)
		fullCol := []field.Element{}
		for i := range numCommitments {
			fullCol = append(fullCol, columns[i][j]...)
		}

		// Check consistency: U_alpha(ω^selectedColID) == α-Horner of opened column
		uAlphaAtJ := allEvals[selectedColID]
		y := vortex.EvalBasePolyHorner(fullCol, alpha)
		if uAlphaAtJ != y {
			return fmt.Errorf("linear combination (coeff mode) inconsistent at query %d (col %d): got %v want %v",
				j, selectedColID, uAlphaAtJ.String(), y.String())
		}
	}
	return nil
}
