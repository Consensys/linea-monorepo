package distributed

import (
	"github.com/consensys/linea-monorepo/prover/maths/field"
	"github.com/consensys/linea-monorepo/prover/protocol/compiler/innerproduct"
	"github.com/consensys/linea-monorepo/prover/protocol/compiler/lookup"
	"github.com/consensys/linea-monorepo/prover/protocol/compiler/mimc"
	"github.com/consensys/linea-monorepo/prover/protocol/compiler/specialqueries"
	"github.com/consensys/linea-monorepo/prover/protocol/query"
	"github.com/consensys/linea-monorepo/prover/protocol/wizard"
	"github.com/consensys/linea-monorepo/prover/protocol/wizardutils"
	"github.com/consensys/linea-monorepo/prover/symbolic"
)

// prepare reduces any query to LPP or GL.
// For Inclusion query, it push the compilation one step further:
// all the inclusion queries are compiled into a LogDarivativeSum query,
// This is required due to the challenge that table M depends on whole the witness.
func prepare(comp *wizard.CompiledIOP) {
	mimc.CompileMiMC(comp)
	specialqueries.RangeProof(comp)
	specialqueries.CompileFixedPermutations(comp)
	innerproduct.Compile(comp)

	IntoLogDerivativeSum(comp)
}

// IntoLogDerivativeSum compiles  all the inclusion queries to a single LogDerivativeSum query that is ready for the split.
// This step is necessary for inclusion,
// as the M table depends on the whole witness and so can not be handled modules-wise without changing the API of WizardIOP.
func IntoLogDerivativeSum(comp *wizard.CompiledIOP) {

	var (
		mainLookupCtx = lookup.CaptureLookupTables(comp)
		lastRound     = comp.NumRounds() - 1
		// zCatalog stores a mapping (round, size) into query.LogDerivativeSumInput and helps finding
		// which Z context should be used to handle a part of a given inclusion
		// query.
		zCatalog = map[int]*query.LogDerivativeSumInput{}
	)

	// Skip the compilation phase if no lookup constraint is being used. Otherwise
	// it will register a verifier action that is not required and will be bugged.
	if len(mainLookupCtx.LookupTables) == 0 {
		return
	}

	// Step 1. construct the "per table" contexts and pack the Sigma's into
	// zCatalog.
	for _, lookupTable := range mainLookupCtx.LookupTables {

		var (
			// get checkedTables, rounds, Filters by lookupTableName
			lookupTableName = lookup.NameTable(lookupTable)
			checkTable      = mainLookupCtx.CheckedTables[lookupTableName]
			round           = mainLookupCtx.Rounds[lookupTableName]
			includedFilters = mainLookupCtx.IncludedFilters[lookupTableName]
			// collapse multiColumns to single Columns and commit to M.
			tableCtx = lookup.CompileLookupTable(comp, round, lookupTable, checkTable, includedFilters)
		)

		// push single-columns into zCatalog
		pushToZCatalog(tableCtx, zCatalog)

		a := lookup.MAssignmentTask{
			M:       tableCtx.M,
			S:       checkTable,
			T:       lookupTable,
			SFilter: includedFilters,
		}

		// assign the multiplicity column
		comp.SubProvers.AppendToInner(round, a.Run)
	}

	// insert a single LogDerivativeSum query for the global zCatalog.
	comp.InsertLogDerivativeSum(lastRound+1, "GlobalLogDerivativeSum", zCatalog)

	// assign parameters of LogDerivativeSum, it is just to prevent the panic attack in the prover
	comp.SubProvers.AppendToInner(lastRound+1, func(run *wizard.ProverRuntime) {
		run.AssignLogDerivSum("GlobalLogDerivativeSum", field.Zero())
	})
}

// pushToZCatalog constructs the numerators and denominators for the collapsed S and T
// into zCatalog, for their corresponding rounds and size.
func pushToZCatalog(stc lookup.SingleTableCtx, zCatalog map[int]*query.LogDerivativeSumInput) {

	// tableCtx push to -> zCtx
	// Process the T columns
	for frag := range stc.T {
		size := stc.M[frag].Size()

		key := size
		if zCatalog[key] == nil {
			zCatalog[key] = &query.LogDerivativeSumInput{
				Size: size,
			}
		}

		zCtxEntry := zCatalog[key]
		zCtxEntry.Numerator = append(zCtxEntry.Numerator, symbolic.Neg(stc.M[frag])) // no functions for num, denom here
		zCtxEntry.Denominator = append(zCtxEntry.Denominator, symbolic.Add(stc.Gamma, stc.T[frag]))
	}

	// Process the S columns
	for table := range stc.S {
		var (
			_, _, size = wizardutils.AsExpr(stc.S[table])
			sFilter    = symbolic.NewConstant(1)
		)

		if stc.SFilters[table] != nil {
			sFilter = symbolic.NewVariable(stc.SFilters[table])
		}

		key := size
		if zCatalog[key] == nil {
			zCatalog[key] = &query.LogDerivativeSumInput{
				Size: size,
			}
		}

		zCtxEntry := zCatalog[key]
		zCtxEntry.Numerator = append(zCtxEntry.Numerator, sFilter)
		zCtxEntry.Denominator = append(zCtxEntry.Denominator, symbolic.Add(stc.Gamma, stc.S[table]))
	}
}
