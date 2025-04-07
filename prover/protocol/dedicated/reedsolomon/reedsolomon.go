package reedsolomon

import (
	"fmt"

	"github.com/consensys/gnark/frontend"
	"github.com/consensys/linea-monorepo/prover/maths/common/smartvectors"
	"github.com/consensys/linea-monorepo/prover/maths/fft"
	"github.com/consensys/linea-monorepo/prover/protocol/accessors"
	"github.com/consensys/linea-monorepo/prover/protocol/coin"
	"github.com/consensys/linea-monorepo/prover/protocol/dedicated/functionals"
	"github.com/consensys/linea-monorepo/prover/protocol/ifaces"
	"github.com/consensys/linea-monorepo/prover/protocol/wizard"
)

const (
	REED_SOLOMON_COEFF       string = "REED_SOLOMON_COEFF"
	REED_SOLOMON_EVAL_CHECK  string = "REED_SOLOMON_EVAL_CHECK"
	REED_SOLOMON_COEFF_CHECK string = "REED_SOLOMON_COEFF_CHECK"
	REED_SOLOMON_BETA        string = "REED_SOLOMON_BETA"
)

// reedSolomonVerifierAction implements the VerifierAction interface for Reed-Solomon code checking.
type reedSolomonVerifierAction struct {
	coeffCheck ifaces.Accessor
	evalCheck  ifaces.Accessor
	h          ifaces.Column
}

// Run executes the native verifier check for Reed-Solomon consistency.
func (a *reedSolomonVerifierAction) Run(run *wizard.VerifierRuntime) error {
	y := a.coeffCheck.GetVal(run)
	y_ := a.evalCheck.GetVal(run)
	if y != y_ {
		return fmt.Errorf("reed-solomon check failed - %v is not a codeword", a.h.GetColID())
	}
	return nil
}

// RunGnark executes the gnark circuit verifier check for Reed-Solomon consistency.
func (a *reedSolomonVerifierAction) RunGnark(api frontend.API, wvc *wizard.WizardVerifierCircuit) {
	y := a.coeffCheck.GetFrontendVariable(api, wvc)
	y_ := a.evalCheck.GetFrontendVariable(api, wvc)
	api.AssertIsEqual(y, y_)
}

// Is code-member
func CheckReedSolomon(comp *wizard.CompiledIOP, rate int, h ifaces.Column) {

	round := h.Round()
	codeDim := h.Size() / rate
	coeff := comp.InsertCommit(
		round,
		ifaces.ColIDf("%v_%v", REED_SOLOMON_COEFF, h.GetColID()),
		codeDim,
	)

	beta := comp.InsertCoin(
		round+1,
		coin.Namef("%v_%v", REED_SOLOMON_BETA, h.GetColID()),
		coin.Field,
	)

	// Inserts the prover before calling the sub-wizard so that it is executed
	// before the sub-prover's wizards.
	comp.RegisterProverAction(round, &reedSolomonProverAction{
		h:       h,
		codeDim: codeDim,
		coeff:   coeff,
	})

	coeffCheck := functionals.CoeffEval(
		comp,
		fmt.Sprintf("%v_%v", REED_SOLOMON_COEFF_CHECK, h.GetColID()),
		beta,
		coeff,
	)

	evalCheck := functionals.Interpolation(
		comp,
		fmt.Sprintf("%v_%v", REED_SOLOMON_EVAL_CHECK, h.GetColID()),
		accessors.NewFromCoin(beta),
		h,
	)

	comp.RegisterVerifierAction(round+1, &reedSolomonVerifierAction{
		coeffCheck: coeffCheck,
		evalCheck:  evalCheck,
		h:          h,
	})
}

// reedSolomonProverAction is the action to assign the Reed-Solomon coefficients.
// It implements the [wizard.ProverAction] interface.
type reedSolomonProverAction struct {
	h       ifaces.Column
	codeDim int
	coeff   ifaces.Column
}

// Run executes the reedSolomonProverAction over a [ProverRuntime]
func (a *reedSolomonProverAction) Run(assi *wizard.ProverRuntime) {
	witness := a.h.GetColAssignment(assi)
	coeffs := smartvectors.FFTInverse(witness, fft.DIF, true, 0, 0, nil).SubVector(0, a.codeDim)
	assi.AssignColumn(a.coeff.GetColID(), coeffs)
}
