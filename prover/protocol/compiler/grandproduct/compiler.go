package grandproduct

import (
	"fmt"

	"github.com/consensys/gnark/frontend"
	"github.com/consensys/linea-monorepo/prover/maths/field"
	"github.com/consensys/linea-monorepo/prover/protocol/compiler/permutation"
	"github.com/consensys/linea-monorepo/prover/protocol/ifaces"
	"github.com/consensys/linea-monorepo/prover/protocol/query"
	"github.com/consensys/linea-monorepo/prover/protocol/wizard"
)

// CompileGrandProductDist compiles [query.GrandProduct] queries and
func CompileGrandProductDist(comp *wizard.CompiledIOP) {
	var (
		allProverActions = make([]permutation.ProverTaskAtRound, comp.NumRounds()+1)
		// zCatalog stores a mapping (round, size) into ZCtx and helps finding
		// which Z context should be used to handle a part of a given permutation
		// query.
		zCatalog = map[[2]int]*permutation.ZCtx{}
	)

	for _, qName := range comp.QueriesParams.AllUnignoredKeys() {
		// Filter out non grand product queries
		grandproduct, ok := comp.QueriesParams.Data(qName).(query.GrandProduct)
		if !ok {
			continue
		}

		// This ensures that the grand product query is not used again in the
		// compilation process. We know that the query was not already ignored at the beginning
		// because we are iterating over the unignored keys.
		comp.QueriesParams.MarkAsIgnored(qName)
		round := comp.QueriesParams.Round(qName)

		dispatchGrandProduct(zCatalog, round, grandproduct)
		var (
			verAction = FinalProductCheck{
				GrandProductID: qName,
			}
		)

		for entry, zC := range zCatalog {

			// z-packing compile; it imposes the correct accumulation over Numerator and Denominator.
			zC.Compile(comp)
			zRound := entry[0]

			// append prover action for the given zCatalog specific to a grand product query
			allProverActions[zRound] = append(allProverActions[zRound], zC)

			// collect all the zOpening for all the z columns for the specific grand product query
			verAction.ZOpenings = append(verAction.ZOpenings, zC.ZOpenings...)
		}
		// verifer step
		comp.RegisterVerifierAction(round, &verAction)
	}
	// prover step; Z assignments
	for proverRound := range allProverActions {
		if len(allProverActions[proverRound]) > 0 {
			comp.RegisterProverAction(proverRound, allProverActions[proverRound])
		}
	}

}

// dispatchGrandProduct applies the grand product argument compilation over
// a specific [query.GrandProduct] using z-packing technique.
func dispatchGrandProduct(
	zCatalog map[[2]int]*permutation.ZCtx,
	round int,
	q query.GrandProduct,
) {
	for size, gpInputs := range q.Inputs {
		var (
			catalogEntry = [2]int{round, size}
		)
		if _, ok := zCatalog[catalogEntry]; !ok {
			zCatalog[catalogEntry] = &permutation.ZCtx{
				Size:  size,
				Round: round,
			}
		}
		ctx := zCatalog[catalogEntry]
		ctx.NumeratorFactors = append(ctx.NumeratorFactors, gpInputs.Numerators...)
		ctx.DenominatorFactors = append(ctx.DenominatorFactors, gpInputs.Denominators...)
	}

}

// FinalProductCheck mutiplies the last entries of the z columns
// and check that it is equal to the query param, implementing the [wizard.VerifierAction]
type FinalProductCheck struct {
	// ZOpenings lists all the openings of all the zCtx
	ZOpenings []query.LocalOpening
	// query ID
	GrandProductID ifaces.QueryID
}

// Run implements the [wizard.VerifierAction]
func (f *FinalProductCheck) Run(run *wizard.VerifierRuntime) error {

	// zProd stores the product of the ending values of the zs as queried
	// in the protocol via the local opening queries.
	zProd := field.One()
	for k := range f.ZOpenings {
		temp := run.GetLocalPointEvalParams(f.ZOpenings[k].ID).Y
		zProd.Mul(&zProd, &temp)
	}

	claimedProd := run.GetGrandProductParams(f.GrandProductID).Y
	if zProd != claimedProd {
		return fmt.Errorf("grand product: the final evaluation check failed for %v\n"+
			"given %v but calculated %v,",
			f.GrandProductID, claimedProd.String(), zProd.String())
	}

	return nil
}

// RunGnark implements the [wizard.VerifierAction]
func (f *FinalProductCheck) RunGnark(api frontend.API, run *wizard.WizardVerifierCircuit) {

	claimedProd := run.GetGrandProductParams(f.GrandProductID).Prod
	// zProd stores the product of the ending values of the z columns
	zProd := frontend.Variable(field.One())
	for k := range f.ZOpenings {
		temp := run.GetLocalPointEvalParams(f.ZOpenings[k].ID).Y
		zProd = api.Mul(zProd, temp)
	}

	api.AssertIsEqual(zProd, claimedProd)
}
