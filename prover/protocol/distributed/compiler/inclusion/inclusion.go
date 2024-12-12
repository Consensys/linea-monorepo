package inclusion

import (
	"github.com/consensys/linea-monorepo/prover/protocol/coin"
	lookUp "github.com/consensys/linea-monorepo/prover/protocol/compiler/lookup"
	"github.com/consensys/linea-monorepo/prover/protocol/ifaces"
	"github.com/consensys/linea-monorepo/prover/protocol/query"
	"github.com/consensys/linea-monorepo/prover/protocol/wizard"
	"github.com/consensys/linea-monorepo/prover/protocol/wizardutils"
	"github.com/consensys/linea-monorepo/prover/symbolic"
)

// CompileLogDerivative scans `comp`, looking for Inclusion queries and compiles
// them using the LogDerivativeLookup technique. The compiler attempts to group
// queries relating to the same table (as such groups needs the same randomness).
//
//	The input is a wizard.CompiledIOP object relevant to the module.
//	It contains a list of the columns relevant to the module (inside its Columns field).
//
// For each T column inside the module, it also contains the M column.
//
// Note that for a lookup query the module may contain only the S or T columns (and not both).
func CompileLogDerivative(moduleComp *wizard.CompiledIOP) {

	var (
		// capture the S and T columns relevant to the module.
		mainLookupCtx = captureModuleLookupTables(moduleComp)
		lastRound     = moduleComp.NumRounds() - 1

		// zCatalog stores a mapping from (round, size) into [query.LogDerivativeSumInput].
		// it packs the sigma columns from the same (round,size) together.
		zCatalog = map[[2]int]*query.LogDerivativeSumInput{}
	)

	// Skip the compilation phase if no lookup constraint is being used. Otherwise
	// it will register a verifier action that is not required and will be bugged.
	if len(mainLookupCtx.LookupTables) == 0 && len(mainLookupCtx.CheckedTables) == 0 {
		return
	}

	// Step 1. construct the "per table" contexts and pack the Sigma's into
	// zCatalog.
	for tableName, checkedTables := range mainLookupCtx.CheckedTables {
		lookupTable := findLookupTableByName(mainLookupCtx.LookupTables, tableName)
		var (
			round           = mainLookupCtx.Rounds[tableName]
			includedFilters = mainLookupCtx.IncludedFilters[tableName]
			mTable          = mainLookupCtx.mTables[tableName]
			// it collapses multiColumn tables to single columns.
			tableCtx = collapsMultiColToSingleCol(moduleComp, round, lookupTable, checkedTables, includedFilters, mTable)
		)

		// push to zCatalog
		PushToZCatalog(tableCtx, zCatalog)
	}

	// Handle cases where only T part is present in the module
	for _, lookupTable := range mainLookupCtx.LookupTables {
		tableName := lookUp.NameTable(lookupTable)
		if _, ok := mainLookupCtx.CheckedTables[tableName]; ok {
			continue
		}

		var (
			round    = mainLookupCtx.Rounds[tableName]
			mTable   = mainLookupCtx.mTables[tableName]
			tableCtx = collapsMultiColToSingleCol(moduleComp, round, lookupTable, nil, nil, mTable)
		)

		// push to zCatalog
		PushToZCatalog(tableCtx, zCatalog)
	}

	// insert a  LogDerivativeSum for all the Sigma Columns in the module.
	moduleComp.InsertLogDerivativeSum(lastRound, "LogDerivativeSum", zCatalog)
}

// findLookupTableByName searches for a lookup table by its name in the list of lookup tables.
func findLookupTableByName(lookupTables [][]table, name string) []table {
	for _, lookupTable := range lookupTables {
		if lookUp.NameTable(lookupTable) == name {
			return lookupTable
		}
	}
	return nil
}

// It collapses the tables of MultiColumns to single columns.
// It also sample the Gamma coin for the rest of the compilation.
func collapsMultiColToSingleCol(
	comp *wizard.CompiledIOP,
	round int,
	lookupTable []table,
	checkedTables []table,
	includedFilters []ifaces.Column,
	mTable table,
) (ctx lookUp.SingleTableCtx) {

	ctx = lookUp.SingleTableCtx{
		TableName: lookUp.NameTable(lookupTable),
		S:         make([]*symbolic.Expression, len(checkedTables)),
		SFilters:  includedFilters,
		T:         make([]*symbolic.Expression, len(lookupTable)),
		M:         make([]ifaces.Column, len(lookupTable)),
	}

	var (
		// isMultiColumn indicates whether the lookup table (and thus the
		// checked tables) have the same number of
		isMultiColumn = (len(lookupTable) > 0 && len(lookupTable[0]) > 1) || (len(checkedTables) > 0 && len(checkedTables[0]) > 1)
	)

	if !isMultiColumn {
		for frag := range ctx.T {
			ctx.T[frag] = symbolic.NewVariable(lookupTable[frag][0])

		}

		for i := range ctx.S {
			ctx.S[i] = symbolic.NewVariable(checkedTables[i][0])
		}
	}

	if isMultiColumn {

		// alpha is the coin used to compute the linear combination of the
		// columns of T and S when they are (both) multi-columns.
		alpha := comp.InsertCoin(
			round+1,
			lookUp.DeriveTableName[coin.Name](lookUp.LogDerivativePrefix, lookupTable, "ALPHA"),
			coin.Field,
		)

		for frag := range ctx.T {
			ctx.T[frag] = wizardutils.RandLinCombColSymbolic(alpha, lookupTable[frag])
		}

		for i := range ctx.S {
			ctx.S[i] = wizardutils.RandLinCombColSymbolic(alpha, checkedTables[i])
		}
	}

	ctx.M = mTable

	ctx.Gamma = comp.InsertCoin(
		round+1,
		lookUp.DeriveTableName[coin.Name](lookUp.LogDerivativePrefix, lookupTable, "GAMMA"),
		coin.Field,
	)

	return ctx
}

// PushToZCatalog constructs the numerators and denominators for the collapsed S and T
// into zCatalog, for their corresponding rounds and size.
func PushToZCatalog(stc lookUp.SingleTableCtx, zCatalog map[[2]int]*query.LogDerivativeSumInput) {

	var (
		round = stc.Gamma.Round
	)

	// tableCtx push to -> zCtx
	// Process the T columns
	for frag := range stc.T {
		size := stc.M[frag].Size()

		key := [2]int{round, size}
		if zCatalog[key] == nil {
			zCatalog[key] = &query.LogDerivativeSumInput{
				Size:  size,
				Round: round,
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

		key := [2]int{round, size}
		if zCatalog[key] == nil {
			zCatalog[key] = &query.LogDerivativeSumInput{
				Size:  size,
				Round: round,
			}
		}

		zCtxEntry := zCatalog[key]
		zCtxEntry.Numerator = append(zCtxEntry.Numerator, sFilter)
		zCtxEntry.Denominator = append(zCtxEntry.Denominator, symbolic.Add(stc.Gamma, stc.S[table]))
	}
}
