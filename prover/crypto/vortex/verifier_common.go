package vortex

import (
	"errors"
	"fmt"

	"github.com/consensys/gnark-crypto/field/koalabear/vortex"
	"github.com/consensys/linea-monorepo/prover/crypto/reedsolomon"
	"github.com/consensys/linea-monorepo/prover/maths/common/smartvectors"
	"github.com/consensys/linea-monorepo/prover/maths/common/smartvectors_mixed"
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

func CheckStatement(linComb smartvectors.SmartVector, ys [][]fext.Element, x, alpha fext.Element) error {

	smartvectors_mixed.IsBase(linComb)

	// Check the consistency of Ys and proof.Linear combination
	yJoined := utils.Join(ys...)
	alphaY := smartvectors.EvaluateFextPolyLagrange(linComb, x)
	alphaYPrime := vortex.EvalFextPolyHorner(yJoined, alpha)

	if alphaY != alphaYPrime {
		return fmt.Errorf("RowLincomb and Y are inconsistent")
	}

	return nil
}

func CheckIsCodeWord(rsParams *reedsolomon.RsParams, v smartvectors.SmartVector) error {
	return rsParams.IsCodewordExt(v)
}

func CheckLinComb(
	linComb smartvectors.SmartVector,
	entryList []int,
	alpha fext.Element,
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

		if y != linComb.GetExt(selectedColID) {
			other := linComb.GetExt(selectedColID)
			return fmt.Errorf("the linear combination is inconsistent %v : %v", y.String(), other.String())
		}
	}

	return nil
}

// CheckStatementCoeff is like CheckStatement but for coefficient-mode U_alpha.
// coefficients holds T polynomial coefficients (E4); evaluation at x uses Horner.
func CheckStatementCoeff(coefficients smartvectors.SmartVector, ys [][]fext.Element, x, alpha fext.Element) error {
	yJoined := utils.Join(ys...)
	alphaY := reedsolomon.ExtCoefficientsEvalAt(coefficients, x)
	alphaYPrime := vortex.EvalFextPolyHorner(yJoined, alpha)

	if alphaY != alphaYPrime {
		return fmt.Errorf("RowLincomb (coeff mode) and Y are inconsistent")
	}
	return nil
}

// CheckLinCombCoeff is like CheckLinComb but for coefficient-mode U_alpha.
// coefficients holds T polynomial coefficients (E4).
// rsParams is used to evaluate U_alpha at all N RS domain points via a single
// FFT, then each queried column is checked by a direct index lookup.
func CheckLinCombCoeff(
	coefficients smartvectors.SmartVector,
	entryList []int,
	alpha fext.Element,
	columns [][][]field.Element,
	rsParams *reedsolomon.RsParams,
) error {
	// One FFT to get all N evaluations — O(N log N) vs O(K×T) Horner per query.
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
