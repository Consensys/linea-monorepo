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
	columns [][][]fext.GenericFieldElem,
) (err error) {

	numCommitments := len(columns)

	var y fext.Element
	fullColBase := []field.Element{}
	fullColExt := []fext.Element{}
	ctr := 0

	for j, selectedColID := range entryList {
		// Will carry the concatenation of the columns for the same entry j

		for i := range numCommitments {
			// Entries of the selected columns #j contained in the commitment #i.

			for _, g := range columns[i][j] {
				ctr++
				if g.IsBase {
					fullColBase = append(fullColBase, g.Base)
				} else {
					fullColExt = append(fullColExt, g.Ext)
				}
			}

		}

		// snaity-check
		if len(fullColBase) != ctr && len(fullColExt) != ctr {
			return fmt.Errorf(" columns are a mixture of base and extension")
		}

		// Check the linear combination is consistent with the opened column
		if len(fullColExt) > 0 {
			y = EvalFextPolyHorner(fullColExt, alpha)
		} else {
			y = vortex.EvalBasePolyHorner(fullColBase, alpha)
		}

		if y != linComb.GetExt(selectedColID) {
			other := linComb.GetExt(selectedColID)
			return fmt.Errorf("the linear combination is inconsistent %v : %v", y.String(), other.String())
		}
	}

	return nil
}

func EvalFextPolyHorner(poly []fext.Element, x fext.Element) fext.Element {
	res := fext.Element{}
	for i := len(poly) - 1; i >= 0; i-- {
		res.Mul(&res, &x)
		res.Add(&res, &poly[i])
	}
	return res
}
