package sishashing

import (
	"fmt"
	"math/big"

	"github.com/consensys/accelerated-crypto-monorepo/crypto/ringsis"
	"github.com/consensys/accelerated-crypto-monorepo/maths/common/poly"
	"github.com/consensys/accelerated-crypto-monorepo/maths/common/smartvectors"
	"github.com/consensys/accelerated-crypto-monorepo/maths/field"
	"github.com/consensys/accelerated-crypto-monorepo/protocol/accessors"
	"github.com/consensys/accelerated-crypto-monorepo/protocol/coin"
	"github.com/consensys/accelerated-crypto-monorepo/protocol/dedicated/functionals"
	"github.com/consensys/accelerated-crypto-monorepo/protocol/ifaces"
	"github.com/consensys/accelerated-crypto-monorepo/protocol/wizard"
	"github.com/consensys/accelerated-crypto-monorepo/utils"
	"github.com/consensys/accelerated-crypto-monorepo/utils/gnarkutil"
	"github.com/consensys/gnark/frontend"
)

/*
Checks the opening of a ring-SIS hash. Note that, this does not perform
the range
*/
func RingSISCheck(
	// compiled wizard iop
	comp *wizard.CompiledIOP,
	// sis parameters
	key *ringsis.Key,
	// preimages of the hashes
	limbSplitPreimage ifaces.Column,
	// alleged hash of the above preimage
	allegedSisHash ifaces.Column,
) {

	instanceName := fmt.Sprintf(
		"LATTICE_CHECK_%v_%v_%v",
		key.Repr(),
		limbSplitPreimage.GetColID(),
		allegedSisHash.GetColID(),
	)

	// Commit to the precomputed key. NB, this is an idemnpotent operation
	// if a previous to the current function was made with the same key returns
	// `InsertPrecomputed` will simply return the original handle.
	laidoutKey := key.LaidOutKey()
	sisKey := comp.InsertPrecomputed(SisKeyName(key), smartvectors.NewRegular(laidoutKey))

	// compute the latest round of declaration
	maxRound := utils.Max(allegedSisHash.Round(), limbSplitPreimage.Round())

	// Range checks the preimage
	comp.InsertRange(limbSplitPreimage.Round(), ifaces.QueryIDf("%v_SHORTNESS", instanceName), limbSplitPreimage, 1<<key.LogTwoBound)

	// Introduces a messages correspondind to the sis hash, but
	// modulo X^n - 1. It's not really a hash per se (because this
	// is not binding), but it is used as an advice value.
	dualSisHash := comp.InsertProof(
		maxRound,
		ifaces.ColID(fmt.Sprintf("%v_DUAL", instanceName)),
		key.Degree,
	)

	// Assign the dual sis hash
	comp.SubProvers.AppendToInner(maxRound, func(assi *wizard.ProverRuntime) {
		limbSplitPreimageSV := limbSplitPreimage.GetColAssignment(assi)
		limbSplitPreimage := smartvectors.IntoRegVec(limbSplitPreimageSV)
		res := key.HashModXnMinus1(limbSplitPreimage)
		assi.AssignColumn(dualSisHash.GetColID(), smartvectors.NewRegular(res))
	})

	// declare a folding coin
	folderCoin := comp.InsertCoin(maxRound+1, coin.Namef("%v_FOLDER", instanceName), coin.Field)
	folder := accessors.AccessorFromCoin(folderCoin)

	// fold the key and the preimage
	foldedKey := functionals.Fold(comp, sisKey, folder, key.Degree)
	foldedPreimage := functionals.Fold(comp, limbSplitPreimage, folder, key.Degree)

	// then scalar product
	ipRound := utils.Max(foldedKey.Round(), foldedPreimage.Round())
	ipName := ifaces.QueryIDf("%v_IP", instanceName)
	comp.InsertInnerProduct(ipRound, ipName, foldedKey, []ifaces.Column{foldedPreimage})

	comp.SubProvers.AppendToInner(ipRound, func(assi *wizard.ProverRuntime) {
		// compute the inner-product
		foldedKey := foldedKey.GetColAssignment(assi)           // overshadows the handle
		foldedPreimage := foldedPreimage.GetColAssignment(assi) // overshadows the handle

		y := smartvectors.InnerProduct(foldedKey, foldedPreimage)
		assi.AssignInnerProduct(ipName, y)
	})

	// check the folding of the polynomial is correct
	comp.InsertVerifier(ipRound, func(a *wizard.VerifierRuntime) error {
		allegedSisHash := a.GetColumn(allegedSisHash.GetColID())
		dualSisHash := a.GetColumn(dualSisHash.GetColID())

		x := folder.GetVal(a)
		yAlleged := a.GetInnerProductParams(ipName).Ys[0]
		yDual := smartvectors.EvalCoeff(dualSisHash, x)
		yActual := smartvectors.EvalCoeff(allegedSisHash, x)

		/*
			If P(X) is of degree 2d

			And
				- Q(X) = P(X) mod X^n - 1
				- R(X) = P(X) mod X^n + 1

			Then, with CRT we have: 2P(X) = (X^n+1)Q(X) - (X^n-1)R(X)
			Here, we can identify at the point x

			yDual * (x^n+1) - yActual * (x^n-1) == 2 * yAlleged
		*/
		var xN, xNminus1, xNplus1 field.Element
		one := field.One()
		xN.Exp(x, big.NewInt(int64(key.Degree)))
		xNminus1.Sub(&xN, &one)
		xNplus1.Add(&xN, &one)

		var left, left0, left1, right field.Element
		left0.Mul(&xNplus1, &yDual)
		left1.Mul(&xNminus1, &yActual)
		left.Sub(&left0, &left1)

		right.Double(&yAlleged)

		if left != right {
			return fmt.Errorf("failed the consistency check of the ring-SIS : %v != %v", left.String(), right.String())
		}

		return nil
	}, func(api frontend.API, wvc *wizard.WizardVerifierCircuit) {

		allegedSisHash := wvc.GetColumn(allegedSisHash.GetColID())
		dualSisHash := wvc.GetColumn(dualSisHash.GetColID())

		x := folder.GetFrontendVariable(api, wvc)
		yAlleged := wvc.GetInnerProductParams(ipName).Ys[0]
		yDual := poly.EvaluateUnivariateGnark(api, dualSisHash, x)
		yActual := poly.EvaluateUnivariateGnark(api, allegedSisHash, x)

		/*
			Same thing as the above function but gnark-side
		*/

		one := field.One()
		xN := gnarkutil.Exp(api, x, key.Degree)
		xNminus1 := api.Sub(xN, one)
		xNplus1 := api.Add(xN, one)
		left0 := api.Mul(xNplus1, yDual)
		left1 := api.Mul(xNminus1, yActual)
		left := api.Add(left0, left1)
		right := api.Add(yAlleged, yAlleged) // i.e doubling yAlleged
		api.AssertIsEqual(left, right)
	})
}

func SisKeyName(key *ringsis.Key) ifaces.ColID {
	return ifaces.ColIDf("SISKEY_%v_%v_%v", key.LogTwoBound, key.LogTwoDegree, key.MaxNbFieldToHash)
}
