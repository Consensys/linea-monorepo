package vortex

import (
	"errors"
	"fmt"

	"github.com/consensys/gnark-crypto/field/koalabear/vortex"
	"github.com/consensys/linea-monorepo/prover/crypto/reedsolomon"
	"github.com/consensys/linea-monorepo/prover/maths/koalabear/field"
	"github.com/consensys/linea-monorepo/prover/maths/koalabear/polynomials"
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
	Alpha field.Ext

	// X is the univariate evaluation point
	X field.Ext

	// Ys are the alleged evaluation at point X
	Ys [][]field.Ext

	// EntryList is the random coin representing the columns to open.
	EntryList []int
}

func CheckStatement(linComb []field.Ext, ys [][]field.Ext, x, alpha field.Ext) error {

	// Check the consistency of Ys and proof.Linear combination
	yJoined := utils.Join(ys...)
	alphaY := polynomials.EvalLagrange(field.VecFromExt(linComb), field.ElemFromExt(x))
	alphaYPrime := vortex.EvalFextPolyHorner(yJoined, alpha)

	if !alphaY.Equal(&alphaYPrime) {
		return fmt.Errorf("RowLincomb and Y are inconsistent")
	}

	return nil
}

func CheckIsCodeWord(rsParams *reedsolomon.RsParams, v []field.Ext) error {
	return rsParams.IsCodewordExt(v)
}

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
