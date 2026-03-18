package reedsolomon

import (
	"fmt"

	"github.com/consensys/gnark-crypto/field/koalabear/fft"
	"github.com/consensys/gnark-crypto/utils"
	"github.com/consensys/gnark/frontend"
	"github.com/consensys/linea-monorepo/prover/maths/common/smartvectors"
	"github.com/consensys/linea-monorepo/prover/maths/field"
	"github.com/consensys/linea-monorepo/prover/maths/field/fext"
	"github.com/consensys/linea-monorepo/prover/maths/field/koalagnark"
	"github.com/consensys/linea-monorepo/prover/protocol/accessors"
	"github.com/consensys/linea-monorepo/prover/protocol/coin"
	"github.com/consensys/linea-monorepo/prover/protocol/dedicated/functionals"
	"github.com/consensys/linea-monorepo/prover/protocol/ifaces"
	"github.com/consensys/linea-monorepo/prover/protocol/wizard"
)

const (
	REED_SOLOMON_COEFF       string = "REED_SOLOMON_COEFF"
	REED_SOLOMON_EVALS       string = "REED_SOLOMON_EVALS"
	REED_SOLOMON_EVAL_CHECK  string = "REED_SOLOMON_EVAL_CHECK"
	REED_SOLOMON_COEFF_CHECK string = "REED_SOLOMON_COEFF_CHECK"
	REED_SOLOMON_BETA        string = "REED_SOLOMON_BETA"
)

type ReedSolomonProverAction struct {
	H       ifaces.Column
	Coeff   ifaces.Column
	CodeDim int
}

func (a *ReedSolomonProverAction) Run(assi *wizard.ProverRuntime) {
	witness := a.H.GetColAssignment(assi)
	domain := fft.NewDomain(uint64(witness.Len()), fft.WithCache())

	if a.H.IsBase() {
		coeffs := make([]field.Element, witness.Len())
		witness.WriteInSlice(coeffs)
		domain.FFTInverse(coeffs, fft.DIF, fft.WithNbTasks(2))
		utils.BitReverse(coeffs)
		// Take only the first `CodeDim` coefficients
		assi.AssignColumn(a.Coeff.GetColID(), smartvectors.NewRegular(coeffs[:a.CodeDim]))
		return
	}

	coeffs := make([]fext.Element, witness.Len())
	witness.WriteInSliceExt(coeffs)
	domain.FFTInverseExt(coeffs, fft.DIF, fft.WithNbTasks(2))
	utils.BitReverse(coeffs)
	// Take only the first `CodeDim` coefficients
	assi.AssignColumn(a.Coeff.GetColID(), smartvectors.NewRegularExt(coeffs[:a.CodeDim]))
}

type ReedSolomonVerifierAction struct {
	CoeffCheck ifaces.Accessor
	EvalCheck  ifaces.Accessor
	HColID     ifaces.ColID
}

func (a *ReedSolomonVerifierAction) Run(run wizard.Runtime) error {
	y := a.CoeffCheck.GetValExt(run)
	y_ := a.EvalCheck.GetValExt(run)

	if y != y_ {
		return fmt.Errorf("reed-solomon check failed - %v is not a codeword", a.HColID)
	}
	return nil
}

func (a *ReedSolomonVerifierAction) RunGnark(api frontend.API, wvc wizard.GnarkRuntime) {
	koalaAPI := koalagnark.NewAPI(api)
	y := a.CoeffCheck.GetFrontendVariableExt(api, wvc)
	y_ := a.EvalCheck.GetFrontendVariableExt(api, wvc)
	koalaAPI.AssertIsEqualExt(y, y_)
}

// CheckReedSolomonFromCoeff is the reverse of CheckReedSolomon: it takes a
// polynomial given in coefficient form (size T) and commits to its RS codeword
// (size T×rate). Consistency is verified via Schwartz-Zippel:
//
//	CanonicalEval(coeff, β) == LagrangeEval(evals, β)
//
// The returned column holds the N codeword evaluations and can be used
// directly with an inclusion lookup (Q, UalphaQ) ⊂ (I, evals).
func CheckReedSolomonFromCoeff(comp *wizard.CompiledIOP, rate int, coeff ifaces.Column) ifaces.Column {
	round := coeff.Round()
	evalSize := coeff.Size() * rate

	evals := comp.InsertCommit(
		round,
		ifaces.ColIDf("%v_%v", REED_SOLOMON_EVALS, coeff.GetColID()),
		evalSize,
		coeff.IsBase(),
	)

	beta := comp.InsertCoin(
		round+1,
		coin.Namef("%v_%v", REED_SOLOMON_BETA, coeff.GetColID()),
		coin.FieldExt,
	)

	comp.RegisterProverAction(round, &ReedSolomonFromCoeffProverAction{
		Coeff:    coeff,
		Evals:    evals,
		EvalSize: evalSize,
	})

	coeffCheck := functionals.CoeffEval(
		comp,
		fmt.Sprintf("%v_%v", REED_SOLOMON_COEFF_CHECK, coeff.GetColID()),
		beta,
		coeff,
	)

	evalCheck := functionals.Interpolation(
		comp,
		fmt.Sprintf("%v_%v", REED_SOLOMON_EVAL_CHECK, coeff.GetColID()),
		accessors.NewFromCoin(beta),
		evals,
	)

	comp.RegisterVerifierAction(round+1, &ReedSolomonVerifierAction{
		CoeffCheck: coeffCheck,
		EvalCheck:  evalCheck,
		HColID:     evals.GetColID(),
	})

	return evals
}

type ReedSolomonFromCoeffProverAction struct {
	Coeff    ifaces.Column
	Evals    ifaces.Column
	EvalSize int
}

func (a *ReedSolomonFromCoeffProverAction) Run(assi *wizard.ProverRuntime) {
	coeffSV := a.Coeff.GetColAssignment(assi)
	domain := fft.NewDomain(uint64(a.EvalSize), fft.WithCache())

	if a.Coeff.IsBase() {
		// Pad T coefficients to N with zeros, then FFT.
		coeffSlice := make([]field.Element, a.EvalSize)
		coeffSV.WriteInSlice(coeffSlice[:coeffSV.Len()])
		utils.BitReverse(coeffSlice)
		domain.FFT(coeffSlice, fft.DIT, fft.WithNbTasks(2))
		assi.AssignColumn(a.Evals.GetColID(), smartvectors.NewRegular(coeffSlice))
		return
	}

	// ext-field case
	coeffSlice := make([]fext.Element, a.EvalSize)
	coeffSV.WriteInSliceExt(coeffSlice[:coeffSV.Len()])
	utils.BitReverse(coeffSlice)
	domain.FFTExt(coeffSlice, fft.DIT, fft.WithNbTasks(2))
	assi.AssignColumn(a.Evals.GetColID(), smartvectors.NewRegularExt(coeffSlice))
}

// CheckReedSolomon verifies that evals (N elements) is a valid RS codeword of
// degree < N/rate. The prover hints the T = N/rate polynomial coefficients via
// an iFFT (hint = coefficients). Soundness is established by Schwartz-Zippel:
//
//	CanonicalEval(coeff, β) == LagrangeEval(evals, β)
//
// The returned column holds the T polynomial coefficients and can be used
// for cheaper Horner evaluations (degree T-1 vs N-1).
func CheckReedSolomon(comp *wizard.CompiledIOP, rate int, evals ifaces.Column) ifaces.Column {

	round := evals.Round()
	codeDim := evals.Size() / rate
	coeff := comp.InsertCommit(
		round,
		ifaces.ColIDf("%v_%v", REED_SOLOMON_COEFF, evals.GetColID()),
		codeDim,
		evals.IsBase(),
	)

	beta := comp.InsertCoin(
		round+1,
		coin.Namef("%v_%v", REED_SOLOMON_BETA, evals.GetColID()),
		coin.FieldExt,
	)

	// Prover computes iFFT(evals) and stores the first T=codeDim coefficients.
	comp.RegisterProverAction(round, &ReedSolomonProverAction{
		H:       evals,
		Coeff:   coeff,
		CodeDim: codeDim,
	})

	// Schwartz-Zippel: CanonicalEval(coeff, β) == LagrangeEval(evals, β).
	// In the gnark circuit both sides are computed from their respective
	// committed columns (coeff as hint, evals as the committed codeword).
	coeffCheck := functionals.CoeffEval(
		comp,
		fmt.Sprintf("%v_%v", REED_SOLOMON_COEFF_CHECK, evals.GetColID()),
		beta,
		coeff,
	)

	evalCheck := functionals.Interpolation(
		comp,
		fmt.Sprintf("%v_%v", REED_SOLOMON_EVAL_CHECK, evals.GetColID()),
		accessors.NewFromCoin(beta),
		evals,
	)

	comp.RegisterVerifierAction(round+1, &ReedSolomonVerifierAction{
		CoeffCheck: coeffCheck,
		EvalCheck:  evalCheck,
		HColID:     evals.GetColID(),
	})

	return coeff
}
