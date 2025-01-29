package xcomp

import (
	"fmt"

	"github.com/consensys/gnark/frontend"
	"github.com/consensys/linea-monorepo/prover/maths/field"
	"github.com/consensys/linea-monorepo/prover/protocol/accessors"
	"github.com/consensys/linea-monorepo/prover/protocol/wizard"
	"github.com/consensys/linea-monorepo/prover/utils/collection"
)

// GetCrossComp generates an (empty) compiledIOP object that is handling the crosse checks
// for example, the global sum over logDerivativeSum is zero.
func GetCrossComp(vRuntimes []*wizard.VerifierRuntime) *wizard.CompiledIOP {

	xComp := wizard.NewCompiledIOP()

	// initialize the PICollector
	va := PublicInputCollector{
		PIFromLogDeriv:  collection.NewMapping[string, field.Element](),
		PIFromGrandProd: collection.NewMapping[string, field.Element](),
		PIFromGrandSum:  collection.NewMapping[string, field.Element](),
	}
	// get the publicInputs values from different verifiers.
	for i, runtime := range vRuntimes {
		va.Index = i
		CollectPIfromVerifer(runtime, &va)
	}

	// register a verifier action to check the consistency of public inputs
	xComp.RegisterVerifierAction(0, &PublicInputChecker{PublicInputCollector: va})

	return xComp
}

// PublicInputCollector collects the public input values from different modules/segments.
type PublicInputCollector struct {
	// maps for collecting publicInputs from different modules.
	PIFromLogDeriv, PIFromGrandProd, PIFromGrandSum collection.Mapping[string, field.Element]
	// index for the verifier from which we are receiving the publicInputs
	Index int
}

// CollectPIfromVerifer adds the public inputs of a given verifier to the Collector.
func CollectPIfromVerifer(run *wizard.VerifierRuntime, pic *PublicInputCollector) {
	var (
		allPI = run.Spec.PublicInputs
	)

	for _, pi := range allPI {

		name := fmt.Sprintf("%v_%v", pi.Name, pic.Index)

		switch v := pi.Acc.(type) {

		case *accessors.FromLogDerivSumAccessor:
			pic.PIFromLogDeriv.InsertNew(name, v.GetVal(run))

		case *accessors.FromGrandProductAccessor:
			pic.PIFromLogDeriv.InsertNew(name, v.GetVal(run))

		case *accessors.FromGrandSumAccessor:
			pic.PIFromLogDeriv.InsertNew(name, v.GetVal(run))
		}
	}

}

type PublicInputChecker struct {
	PublicInputCollector
	skip bool
}

// Run implements the [wizard.VerifierAction], it handles the cross checks over the public inputs.
// for example the global sum over the LogDerivativeSum from different segments should be zero.
func (pir *PublicInputChecker) Run(run *wizard.VerifierRuntime) error {
	var (
		logDerivSum, grandSum field.Element
		grandProduct          = field.One()
	)

	for _, key := range pir.PIFromLogDeriv.ListAllKeys() {
		curr := pir.PIFromLogDeriv.MustGet(key)
		logDerivSum.Add(&logDerivSum, &curr)
	}
	for _, key := range pir.PIFromGrandProd.ListAllKeys() {
		curr := pir.PIFromGrandProd.MustGet(key)
		grandProduct.Add(&grandProduct, &curr)
	}
	for _, key := range pir.PIFromGrandSum.ListAllKeys() {
		curr := pir.PIFromGrandSum.MustGet(key)
		grandSum.Add(&grandSum, &curr)
	}

	if logDerivSum != field.Zero() {
		panic("the global sum over LogDerivSumParams is not zero," +
			" maybe the same coin over different modules has different values")
	}

	if grandProduct != field.One() {
		panic("the global product overGrandProductParams is not 1," +
			" maybe the same coin over different modules has different values")
	}

	if grandSum != field.Zero() {
		panic("the global sum over GrandSumParams is not zero," +
			" maybe the same coin over different modules has different values")
	}
	return nil

}

// RunGnark implements the [wizard.VerifierAction]
func (pir *PublicInputChecker) RunGnark(api frontend.API, run *wizard.WizardVerifierCircuit) {
	var (
		logDerivSum, grandProduct, grandSum frontend.API
	)

	for _, key := range pir.PIFromLogDeriv.ListAllKeys() {
		curr := pir.PIFromLogDeriv.MustGet(key)
		api.Add(logDerivSum, curr)
	}
	for _, key := range pir.PIFromGrandProd.ListAllKeys() {
		curr := pir.PIFromGrandProd.MustGet(key)
		api.Add(grandProduct, curr)
	}
	for _, key := range pir.PIFromGrandSum.ListAllKeys() {
		curr := pir.PIFromGrandSum.MustGet(key)
		api.Add(grandSum, curr)
	}

	api.AssertIsEqual(logDerivSum, field.Zero())
	api.AssertIsEqual(grandProduct, field.One())
	api.AssertIsEqual(grandSum, field.Zero())
}

func (v *PublicInputChecker) Skip() {
	v.skip = true
}

func (v *PublicInputChecker) IsSkipped() bool {
	return v.skip
}
