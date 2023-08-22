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
	"github.com/consensys/accelerated-crypto-monorepo/protocol/dedicated/expr_handle"
	"github.com/consensys/accelerated-crypto-monorepo/protocol/dedicated/functionals"
	"github.com/consensys/accelerated-crypto-monorepo/protocol/ifaces"
	"github.com/consensys/accelerated-crypto-monorepo/protocol/wizard"
	"github.com/consensys/accelerated-crypto-monorepo/protocol/wizardutils"
	"github.com/consensys/accelerated-crypto-monorepo/utils"
	"github.com/consensys/accelerated-crypto-monorepo/utils/gnarkutil"
	"github.com/consensys/gnark/frontend"
)

func MultiRingSISCheck(
	// string identifier for the declared instance of RingSIS check
	instanceName string,
	// compiled wizard iop
	comp *wizard.CompiledIOP,
	key *ringsis.Key,
	limbSplitPreimage []ifaces.Column,
	// concatenated hashes
	concatenatedSisHash ifaces.Column,
) {

	// Commit to the precomputed key. NB, this is an idemnpotent operation
	// if a previous to the current function was made with the same key returns
	// `InsertPrecomputed` will simply return the original handle.
	laidoutKey := key.LaidOutKey()
	sisKey := comp.InsertPrecomputed(SisKeyName(key), smartvectors.NewRegular(laidoutKey))

	maxRound := wizardutils.MaxRound(append(limbSplitPreimage, concatenatedSisHash)...)

	// Range checks the preimage
	for _, preimage := range limbSplitPreimage {
		comp.InsertRange(
			maxRound,
			ifaces.QueryIDf("%v_%v_SHORTNESS", instanceName, preimage.GetColID()),
			preimage,
			1<<key.LogTwoBound,
		)
	}

	// Introduces a messages correspondind to the sis hash, but
	// modulo X^n - 1. It's not really a hash per se (because this
	// is not binding), but it is used as an advice value.
	concatenatedDualSisHash := comp.InsertProof(
		maxRound,
		ifaces.ColID(fmt.Sprintf("%v_DUAL", instanceName)),
		concatenatedSisHash.Size(),
	)

	// Assign the dual sis hash, by concatenating the dual sis hash of all columns
	comp.SubProvers.AppendToInner(maxRound, func(assi *wizard.ProverRuntime) {
		res := make([]field.Element, concatenatedSisHash.Size())
		for i := range limbSplitPreimage {
			limbSplitPreimageSV := limbSplitPreimage[i].GetColAssignment(assi)
			limbSplitPreimageI := smartvectors.IntoRegVec(limbSplitPreimageSV)
			hashI := key.HashModXnMinus1(limbSplitPreimageI)
			copy(res[i*key.Degree:(i+1)*key.Degree], hashI)
		}
		assi.AssignColumn(concatenatedDualSisHash.GetColID(), smartvectors.NewRegular(res))
	})

	// Declare a folding coin
	folderCoin := comp.InsertCoin(maxRound+1, coin.Namef("%v_FOLDER", instanceName), coin.Field)
	folder := accessors.AccessorFromCoin(folderCoin)

	// Collapse the preimages into a linear combination
	collapseCoin := comp.InsertCoin(maxRound+1, coin.Namef("%v_COLLAPSER", instanceName), coin.Field)
	linCombP := expr_handle.RandLinCombCol(
		comp,
		accessors.AccessorFromCoin(collapseCoin),
		limbSplitPreimage,
	)

	// fold the key and the (collapsed preimage)
	foldedKey := functionals.Fold(comp, sisKey, folder, key.Degree)
	foldedPreimage := functionals.Fold(comp, linCombP, folder, key.Degree)

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
		allegedSisHash := a.GetColumn(concatenatedSisHash.GetColID())
		dualSisHash := a.GetColumn(concatenatedDualSisHash.GetColID())

		x := folder.GetVal(a)
		t := a.GetRandomCoinField(collapseCoin.Name)
		yAlleged := a.GetInnerProductParams(ipName).Ys[0]
		yDual := smartvectors.EvalCoeffBivariate(dualSisHash, x, key.Degree, t)
		yActual := smartvectors.EvalCoeffBivariate(allegedSisHash, x, key.Degree, t)

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
		allegedSisHash := wvc.GetColumn(concatenatedSisHash.GetColID())
		dualSisHash := wvc.GetColumn(concatenatedDualSisHash.GetColID())

		x := folder.GetFrontendVariable(api, wvc)
		t := wvc.GetRandomCoinField(collapseCoin.Name)
		yAlleged := wvc.GetInnerProductParams(ipName).Ys[0]

		yDual := poly.GnarkEvalCoeffBivariate(api, dualSisHash, x, key.Degree, t)
		yActual := poly.GnarkEvalCoeffBivariate(api, allegedSisHash, x, key.Degree, t)

		/*
			If P(X) is of degree 2d

			And
				- Q(X) = P(X) mod X^n - 1
				- R(X) = P(X) mod X^n + 1

			Then, with CRT we have: 2P(X) = (X^n+1)Q(X) - (X^n-1)R(X)
			Here, we can identify at the point x

			yDual * (x^n+1) - yActual * (x^n-1) == 2 * yAlleged
		*/
		one := field.One()
		xN := gnarkutil.Exp(api, x, key.Degree)
		xNminus1 := api.Sub(xN, one)
		xNplus1 := api.Add(xN, one)
		left0 := api.Mul(xNplus1, yDual)
		left1 := api.Mul(xNminus1, yActual)
		left := api.Sub(left0, left1)
		right := api.Mul(yAlleged, 2)
		api.AssertIsEqual(left, right)
	})

}
