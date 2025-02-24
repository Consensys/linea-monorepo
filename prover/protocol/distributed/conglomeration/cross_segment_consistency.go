package conglomeration

import (
	"errors"
	"fmt"

	"github.com/consensys/gnark/frontend"
	"github.com/consensys/linea-monorepo/prover/maths/field"
	"github.com/consensys/linea-monorepo/prover/protocol/distributed/constants"
	"github.com/consensys/linea-monorepo/prover/protocol/wizard"
)

// crossSegmentCheck is a verifier action that performs cross-segment checks:
// for instance, it checks that the log-derivative sums all sums to 0 and that
// the grand  product is 1. The goal is to ensure that the lookups, permutations
// in the original protocol are satisfied.
type CrossSegmentCheck struct {
	Ctxs []*recursionCtx
	skip bool
}

// Run implements the [wizard.VerifierAction], it handles the cross checks over
// the public inputs. for example the global sum over the LogDerivativeSum from
// different segments should be zero.
func (pir *CrossSegmentCheck) Run(run wizard.Runtime) error {

	var (
		logDerivSumAcc, grandSumAcc field.Element
		grandProductAcc             = field.One()
		err                         error
	)

	for i, ctx := range pir.Ctxs {

		var (
			wrappedRun  = &runtimeTranslator{Prefix: ctx.Translator.Prefix, Rt: run}
			tmpl        = ctx.Tmpl
			nextTmpl    = pir.Ctxs[(i+1)%len(pir.Ctxs)].Tmpl
			logDerivSum = tmpl.GetPublicInputAccessor(constants.LogDerivativeSumPublicInput).GetVal(wrappedRun)
			grandProd   = tmpl.GetPublicInputAccessor(constants.GrandProductPublicInput).GetVal(wrappedRun)
			grandSum    = tmpl.GetPublicInputAccessor(constants.GrandSumPublicInput).GetVal(wrappedRun)

			providerHash     = tmpl.GetPublicInputAccessor(constants.GlobalProviderPublicInput).GetVal(wrappedRun)
			nextReceiverHash = nextTmpl.GetPublicInputAccessor(constants.GlobalReceiverPublicInput).GetVal(wrappedRun)
		)

		logDerivSumAcc.Add(&logDerivSumAcc, &logDerivSum)
		grandSumAcc.Add(&grandSumAcc, &grandSum)
		grandProductAcc.Mul(&grandProductAcc, &grandProd)

		if providerHash != nextReceiverHash {
			err = errors.Join(err, fmt.Errorf("error in crosse checks for distributed global: "+
				"the provider of the current template is different from the receiver of the next template "))
		}
	}

	if logDerivSumAcc != field.Zero() {
		err = errors.Join(err, fmt.Errorf("the global sum over LogDerivSumParams is not zero,"+
			" maybe the same coin over different modules has different values"))
	}

	if grandProductAcc != field.One() {
		err = errors.Join(err, fmt.Errorf("the global product overGrandProductParams is not 1,"+
			" maybe the same coin over different modules has different values"))
	}

	if grandSumAcc != field.Zero() {
		err = errors.Join(err, fmt.Errorf("the global sum over GrandSumParams is not zero,"+
			" maybe the same coin over different modules has different values"))
	}

	if err != nil {
		return fmt.Errorf("[conglomeration.crossSegmentConsistency] %w", err)
	}

	return nil
}

// RunGnark implements the [wizard.VerifierAction]
func (pir *CrossSegmentCheck) RunGnark(api frontend.API, run wizard.GnarkRuntime) {

	var (
		logDerivSumAcc  = frontend.Variable(0)
		grandSumAcc     = frontend.Variable(0)
		grandProductAcc = frontend.Variable(1)
	)

	for i, ctx := range pir.Ctxs {

		var (
			wrappedRun  = &gnarkRuntimeTranslator{Prefix: ctx.Translator.Prefix, Rt: run}
			tmpl        = ctx.Tmpl
			nextTmpl    = pir.Ctxs[(i+1)%len(pir.Ctxs)].Tmpl
			logDerivSum = tmpl.GetPublicInputAccessor(constants.LogDerivativeSumPublicInput).GetFrontendVariable(api, wrappedRun)
			grandProd   = tmpl.GetPublicInputAccessor(constants.GrandProductPublicInput).GetFrontendVariable(api, wrappedRun)
			grandSum    = tmpl.GetPublicInputAccessor(constants.GrandSumPublicInput).GetFrontendVariable(api, wrappedRun)

			providerHash     = tmpl.GetPublicInputAccessor(constants.GlobalProviderPublicInput).GetFrontendVariable(api, wrappedRun)
			nextReceiverHash = nextTmpl.GetPublicInputAccessor(constants.GlobalReceiverPublicInput).GetFrontendVariable(api, wrappedRun)
		)

		logDerivSumAcc = api.Add(logDerivSumAcc, logDerivSum)
		grandSumAcc = api.Add(grandSumAcc, grandSum)
		grandProductAcc = api.Mul(grandProductAcc, grandProd)

		api.AssertIsEqual(providerHash, nextReceiverHash)

	}

	api.AssertIsEqual(logDerivSumAcc, field.Zero())
	api.AssertIsEqual(grandProductAcc, field.One())
	api.AssertIsEqual(grandSumAcc, field.Zero())
}

func (v *CrossSegmentCheck) Skip() {
	v.skip = true
}

func (v *CrossSegmentCheck) IsSkipped() bool {
	return v.skip
}
