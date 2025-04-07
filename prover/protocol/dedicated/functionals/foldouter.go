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

// foldOuterVerifierAction implements the VerifierAction interface for outer folding consistency.
type foldOuterVerifierAction struct {
	foldedEvalAcc ifaces.Accessor
	hEvalAcc      ifaces.Accessor
	foldedName    ifaces.ColID
}

// Run executes the native verifier check for outer folding consistency.
func (a *foldOuterVerifierAction) Run(run *wizard.VerifierRuntime) error {
	if a.foldedEvalAcc.GetVal(run) != a.hEvalAcc.GetVal(run) {
		return fmt.Errorf("verifier of folding failed %v", a.foldedName)
	}
	return nil
}

// RunGnark executes the gnark circuit verifier check for outer folding consistency.
func (a *foldOuterVerifierAction) RunGnark(api frontend.API, wvc *wizard.WizardVerifierCircuit) {
	c := a.foldedEvalAcc.GetFrontendVariable(api, wvc)
	c_ := a.hEvalAcc.GetFrontendVariable(api, wvc)
	api.AssertIsEqual(c, c_)
}

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

	comp.RegisterProverAction(round, &foldOuterProverAction{
		h:           h,
		x:           x,
		outerDegree: outerDegree,
		innerDegree: innerDegree,
		foldedName:  foldedName,
		folded:      folded,
	})

	innerCoinName := coin.Namef("INNER_COIN_%v", folded.GetColID())
	innerCoin := comp.InsertCoin(round+1, innerCoinName, coin.Field)
	innerCoinAcc := accessors.NewFromCoin(innerCoin)

	foldedEvalAcc := CoeffEval(comp, folded.String(), innerCoin, folded)
	hEvalAcc := EvalCoeffBivariate(comp, folded.String(), h, innerCoinAcc, x, innerDegree, outerDegree)

	verRound := utils.Max(innerCoinAcc.Round(), foldedEvalAcc.Round())

	// Check that the two evaluations yield the same result
	comp.RegisterVerifierAction(verRound, &foldOuterVerifierAction{
		foldedEvalAcc: foldedEvalAcc,
		hEvalAcc:      hEvalAcc,
		foldedName:    foldedName,
	})

	return folded
}

// foldOuterProverAction is the action to assign the folded outer column.
// It implements the [wizard.ProverAction] interface.
type foldOuterProverAction struct {
	h           ifaces.Column
	x           ifaces.Accessor
	outerDegree int
	innerDegree int
	foldedName  ifaces.ColID
	folded      ifaces.Column
}

// Run executes the foldOuterProverAction over a [ProverRuntime]
func (a *foldOuterProverAction) Run(assi *wizard.ProverRuntime) {
	// We need to compute an assignment for "folded"
	h := a.h.GetColAssignment(assi) // overshadows the handle
	x := a.x.GetVal(assi)           // overshadows the accessor

	// Split h in "outerDegree" segments of size "innerDegree"
	innerChunks := make([]smartvectors.SmartVector, a.outerDegree)
	for i := range innerChunks {
		innerChunks[i] = h.SubVector(i*a.innerDegree, (i+1)*a.innerDegree)
	}

	// Assign the folding as the RLC of the chunks using powers of x
	foldedVal := smartvectors.PolyEval(innerChunks, x)
	assi.AssignColumn(a.foldedName, foldedVal)
}
