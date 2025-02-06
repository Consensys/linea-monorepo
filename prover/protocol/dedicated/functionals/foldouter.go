package functionals

import (
	"fmt"

	"github.com/consensys/gnark/frontend"
	"github.com/consensys/linea-monorepo/prover/maths/common/smartvectors"
	"github.com/consensys/linea-monorepo/prover/protocol/accessors"
	"github.com/consensys/linea-monorepo/prover/protocol/coin"
	"github.com/consensys/linea-monorepo/prover/protocol/ifaces"
	"github.com/consensys/linea-monorepo/prover/protocol/wizard"
	"github.com/consensys/linea-monorepo/prover/utils"
	"github.com/sirupsen/logrus"
)

// Same as fold but folds on the outer-variable rather than the inner variable
func FoldOuter(comp *wizard.CompiledIOP, h ifaces.Column, x ifaces.Accessor, outerDegree int) ifaces.Column {

	round := x.Round()

	foldedSize := h.Size() / outerDegree
	innerDegree := foldedSize
	foldedName := ifaces.ColIDf("FOLDED_OUTER_%v_%v_%v", h.GetColID(), x.Name(), outerDegree)
	folded := comp.InsertCommit(round, foldedName, foldedSize)

	if x.Round() <= h.Round() {
		logrus.Debugf("Unsafe, the coin is before the commitment : %v", foldedName)
	}

	comp.SubProvers.AppendToInner(round, func(assi *wizard.ProverRuntime) {
		// We need to compute an assignment for "folded"
		h := h.GetColAssignment(assi) // overshadows the handle
		x := x.GetVal(assi)           // overshadows the accessor

		// Split h in "outerDegree" segments of size "innerDegree"
		innerChunks := make([]smartvectors.SmartVector, outerDegree)
		for i := range innerChunks {
			innerChunks[i] = h.SubVector(i*innerDegree, (i+1)*innerDegree)
		}

		// Assign the folding as the RLC of the chunks using powers of x
		foldedVal := smartvectors.PolyEval(innerChunks, x)
		assi.AssignColumn(foldedName, foldedVal)
	})

	innerCoinName := coin.Namef("INNER_COIN_%v", folded.GetColID())
	innerCoin := comp.InsertCoin(round+1, innerCoinName, coin.Field)
	innerCoinAcc := accessors.NewFromCoin(innerCoin)

	foldedEvalAcc := CoeffEval(comp, folded.String(), innerCoin, folded)
	hEvalAcc := EvalCoeffBivariate(comp, folded.String(), h, innerCoinAcc, x, innerDegree, outerDegree)

	verRound := utils.Max(innerCoinAcc.Round(), foldedEvalAcc.Round())

	// Check that the two evaluations yield the same result
	comp.InsertVerifier(verRound, func(run wizard.Runtime) error {
		if foldedEvalAcc.GetVal(run) != hEvalAcc.GetVal(run) {
			return fmt.Errorf("verifier of folding failed %v", foldedName)
		}
		return nil
	}, func(api frontend.API, run wizard.GnarkRuntime) {
		c := foldedEvalAcc.GetFrontendVariable(api, run)
		c_ := hEvalAcc.GetFrontendVariable(api, run)
		api.AssertIsEqual(c, c_)
	})

	return folded
}
