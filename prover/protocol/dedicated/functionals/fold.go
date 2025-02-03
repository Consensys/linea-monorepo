package functionals

import (
	"fmt"

	"github.com/consensys/gnark/frontend"
	"github.com/consensys/linea-monorepo/prover/maths/common/smartvectors"
	"github.com/consensys/linea-monorepo/prover/maths/field"
	"github.com/consensys/linea-monorepo/prover/protocol/accessors"
	"github.com/consensys/linea-monorepo/prover/protocol/coin"
	"github.com/consensys/linea-monorepo/prover/protocol/ifaces"
	"github.com/consensys/linea-monorepo/prover/protocol/wizard"
	"github.com/consensys/linea-monorepo/prover/utils"
	"github.com/sirupsen/logrus"
)

func Fold(comp *wizard.CompiledIOP, h ifaces.Column, x ifaces.Accessor, innerDegree int) ifaces.Column {

	round := x.Round()

	foldedSize := h.Size() / innerDegree
	foldedName := ifaces.ColIDf("FOLDED_%v_%v_%v", h.GetColID(), x.Name(), innerDegree)
	folded := comp.InsertCommit(round, foldedName, foldedSize)

	if x.Round() <= h.Round() {
		logrus.Debugf("Unsafe, the coin is before the commitment : %v", foldedName)
	}

	comp.SubProvers.AppendToInner(round, func(assi *wizard.ProverRuntime) {
		// We need to compute an assignment for "folded"
		h := h.GetColAssignment(assi) // overshadows the handle
		x := x.GetVal(assi)           // overshadows the accessor

		foldedVal := make([]field.Element, foldedSize)
		for i := range foldedVal {
			subH := h.SubVector(i*innerDegree, (i+1)*innerDegree)
			foldedVal[i] = smartvectors.EvalCoeff(subH, x)
		}

		assi.AssignColumn(foldedName, smartvectors.NewRegular(foldedVal))
	})

	outerCoinName := coin.Namef("OUTER_COIN_%v", folded.GetColID())
	outerCoin := comp.InsertCoin(round+1, outerCoinName, coin.Field)
	outerCoinAcc := accessors.NewFromCoin(outerCoin)

	foldedEvalAcc := CoeffEval(comp, folded.String(), outerCoin, folded)
	hEvalAcc := EvalCoeffBivariate(comp, folded.String(), h, x, outerCoinAcc, innerDegree, folded.Size())

	verRound := utils.Max(outerCoinAcc.Round(), foldedEvalAcc.Round())

	// Check that the two evaluations yield the same result
	comp.InsertVerifier(verRound, func(a wizard.Runtime) error {
		if foldedEvalAcc.GetVal(a) != hEvalAcc.GetVal(a) {
			return fmt.Errorf("verifier of folding failed %v", foldedName)
		}
		return nil
	}, func(api frontend.API, wvc wizard.GnarkRuntime) {
		c := foldedEvalAcc.GetFrontendVariable(api, wvc)
		c_ := hEvalAcc.GetFrontendVariable(api, wvc)
		api.AssertIsEqual(c, c_)
	})

	return folded
}
