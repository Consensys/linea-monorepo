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
// queries relating to the same table. This allows saving in commitment because
// when grouping is possible, then we only need to commit to a single
// extract the table and the checked from the lookup query and ensures that the
// table are in canonical order. That is because we want to group lookups into
// the same columns but not in the same order.
func CompileLogDerivative(comp *wizard.CompiledIOP) {

	var (
		mainLookupCtx = captureModuleLookupTables(comp)
		lastRound     = comp.NumRounds() - 1

		// zCatalog stores a mapping (round, size) into ZCtx and helps finding
		// which Z context should be used to handle a part of a given permutation
		// query.
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
			tableCtx        = compileLookupTable(comp, round, lookupTable, checkedTables, includedFilters)
		)

		// push to zCatalog
		PushToZCatalog(tableCtx, zCatalog)
	}

	// Handle cases where only T part is present
	for _, lookupTable := range mainLookupCtx.LookupTables {
		tableName := lookUp.NameTable(lookupTable)
		if _, ok := mainLookupCtx.CheckedTables[tableName]; ok {
			continue
		}

		var (
			round    = mainLookupCtx.Rounds[tableName]
			tableCtx = compileLookupTable(comp, round, lookupTable, nil, nil)
		)

		// push to zCatalog
		PushToZCatalog(tableCtx, zCatalog)
	}

	// insert a  LogDerivativeSum for all the Sigma Columns .
	comp.InsertLogDerivativeSum(lastRound, "LogDerivativeSum", zCatalog)
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

// compileLookupTable applies the log-derivative lookup compilation context to
// the supplied table. round denotes the interaction round in which to start the
// compilation.
//
// It registers the following queries
//   - (1) The verifier queries that $\sum_{k=0\ldots n-1} (\Sigma_{S,k})[|S_k| - 1] == (\Sigma_T)[|T| - 1]$. Namely, the sum of the last entry of all $\Sigma_{S,k}$ equals the last entry of $\Sigma_T$
//   - (2) **(For all k)** the verifier makes a `Local` query : $(\Sigma_{S,k})[0] = \frac{1}{S_{k,0} + \gamma}$
//   - (3) The verifier makes a `Local` query : $(\Sigma_T)[0] = \frac{M_0}{T_0 + \gamma}$
//   - (4) **(For all k)** The verifier makes a `Global` query : $\left((\Sigma_{S,k})[i] - (\Sigma_{S,k})[i-1]\right)(S_{k,i} + \gamma) = 1$
//   - (5) The verifier makes a `Global` query : $\left((\Sigma_T)[i] - (\Sigma_T)[i-1]\right)(T_i + \gamma) = M_i$

// here we are looking up set of columns S in a single column T
func compileLookupTable(
	comp *wizard.CompiledIOP,
	round int,
	lookupTable []table,
	checkedTables []table,
	includedFilters []ifaces.Column,
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
			ctx.M[frag] = comp.InsertCommit(
				round,
				lookUp.DeriveTableNameWithIndex[ifaces.ColID](lookUp.LogDerivativePrefix, lookupTable, frag, "M"),
				lookupTable[frag][0].Size(),
			)

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
			ctx.M[frag] = comp.InsertCommit(
				round,
				lookUp.DeriveTableNameWithIndex[ifaces.ColID](lookUp.LogDerivativePrefix, lookupTable, frag, "M"),
				lookupTable[frag][0].Size(),
			)
		}

		for i := range ctx.S {
			ctx.S[i] = wizardutils.RandLinCombColSymbolic(alpha, checkedTables[i])
		}
	}

	ctx.Gamma = comp.InsertCoin(
		round+1,
		lookUp.DeriveTableName[coin.Name](lookUp.LogDerivativePrefix, lookupTable, "GAMMA"),
		coin.Field,
	)

	return ctx
}

// PushToZCatalog constructs the numerators and denominators for S and T of the
// stc into zCatalog for their corresponding rounds and size.
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
