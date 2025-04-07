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

// foldVerifierAction implements the VerifierAction interface for folding consistency.
type foldVerifierAction struct {
	foldedEvalAcc ifaces.Accessor
	hEvalAcc      ifaces.Accessor
	foldedName    ifaces.ColID
}

// Run executes the native verifier check for folding consistency.
func (a *foldVerifierAction) Run(run *wizard.VerifierRuntime) error {
	if a.foldedEvalAcc.GetVal(run) != a.hEvalAcc.GetVal(run) {
		return fmt.Errorf("verifier of folding failed %v", a.foldedName)
	}
	return nil
}

// RunGnark executes the gnark circuit verifier check for folding consistency.
func (a *foldVerifierAction) RunGnark(api frontend.API, wvc *wizard.WizardVerifierCircuit) {
	c := a.foldedEvalAcc.GetFrontendVariable(api, wvc)
	c_ := a.hEvalAcc.GetFrontendVariable(api, wvc)
	api.AssertIsEqual(c, c_)
}

func Fold(comp *wizard.CompiledIOP, h ifaces.Column, x ifaces.Accessor, innerDegree int) ifaces.Column {

	round := x.Round()

	foldedSize := h.Size() / innerDegree
	foldedName := ifaces.ColIDf("FOLDED_%v_%v_%v", h.GetColID(), x.Name(), innerDegree)
	folded := comp.InsertCommit(round, foldedName, foldedSize)

	if x.Round() <= h.Round() {
		logrus.Debugf("Unsafe, the coin is before the commitment : %v", foldedName)
	}

	comp.RegisterProverAction(round, &foldProverAction{
		h:           h,
		x:           x,
		innerDegree: innerDegree,
		foldedName:  foldedName,
		foldedSize:  foldedSize,
	})

	outerCoinName := coin.Namef("OUTER_COIN_%v", folded.GetColID())
	outerCoin := comp.InsertCoin(round+1, outerCoinName, coin.Field)
	outerCoinAcc := accessors.NewFromCoin(outerCoin)

	foldedEvalAcc := CoeffEval(comp, folded.String(), outerCoin, folded)
	hEvalAcc := EvalCoeffBivariate(comp, folded.String(), h, x, outerCoinAcc, innerDegree, folded.Size())

	verRound := utils.Max(outerCoinAcc.Round(), foldedEvalAcc.Round())

	// Check that the two evaluations yield the same result
	comp.RegisterVerifierAction(verRound, &foldVerifierAction{
		foldedEvalAcc: foldedEvalAcc,
		hEvalAcc:      hEvalAcc,
		foldedName:    foldedName,
	})

	return folded
}

// foldProverAction is the action to assign the folded column.
// It implements the [wizard.ProverAction] interface.
type foldProverAction struct {
	h           ifaces.Column
	x           ifaces.Accessor
	innerDegree int
	foldedName  ifaces.ColID
	foldedSize  int
}

// Run executes the foldProverAction over a [ProverRuntime]
func (a *foldProverAction) Run(assi *wizard.ProverRuntime) {
	// We need to compute an assignment for "folded"
	h := a.h.GetColAssignment(assi) // overshadows the handle
	x := a.x.GetVal(assi)           // overshadows the accessor

	foldedVal := make([]field.Element, a.foldedSize)
	for i := range foldedVal {
		subH := h.SubVector(i*a.innerDegree, (i+1)*a.innerDegree)
		foldedVal[i] = smartvectors.EvalCoeff(subH, x)
	}

	assi.AssignColumn(a.foldedName, smartvectors.NewRegular(foldedVal))
}
