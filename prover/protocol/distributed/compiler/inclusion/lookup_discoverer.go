package inclusion

import (
	"github.com/consensys/linea-monorepo/prover/maths/field"
	"github.com/consensys/linea-monorepo/prover/protocol/column"
	"github.com/consensys/linea-monorepo/prover/protocol/column/verifiercol"
	lookUp "github.com/consensys/linea-monorepo/prover/protocol/compiler/lookup"
	"github.com/consensys/linea-monorepo/prover/protocol/ifaces"
	"github.com/consensys/linea-monorepo/prover/protocol/query"
	"github.com/consensys/linea-monorepo/prover/protocol/wizard"
)

// table is an alias for a list of column. We use it in the scope of the lookup
// compiler as a shorthand to make the code more eye-parseable.
type table = []ifaces.Column

// captureModuleLookupTables inspects comp and looks for Inclusion queries.
// It groups the matched queries by lookup table and marks them as ignored.
// It creates a list of lookupTables and checkedTables present in the module.
//
// The input is a compiledIOP object that stores the columns relevant to the module (in its Column field)
// Note that for a lookup query the module may contain only S or T table (and not necessarily both).
func captureModuleLookupTables(moduleComp *wizard.CompiledIOP) lookUp.MainLookupCtx {
	var (
		ctx = lookUp.MainLookupCtx{
			LookupTables:    [][]table{},
			CheckedTables:   map[string][]table{},
			IncludedFilters: map[string][]ifaces.Column{},
			Rounds:          map[string]int{},
		}
	)

	// Collect all the lookup queries into "lookups"
	for _, qName := range moduleComp.QueriesNoParams.AllUnignoredKeys() {

		// Filter out non-lookup queries
		lookup, ok := moduleComp.QueriesNoParams.Data(qName).(query.Inclusion)
		if !ok {
			continue
		}

		// Determine if the query is relevant to the module
		relevantPart := determineRelevantPart(lookup, moduleComp.Columns)
		if relevantPart == "" {
			continue
		}

		// This ensures that the lookup query is not used again in the
		// compilation process. We know that the query was already ignored at
		// the beginning because we are iterating over the unignored keys.
		moduleComp.QueriesNoParams.MarkAsIgnored(qName)

		var (
			// checkedTable corresponds to the "included" table and lookupTable
			// corresponds to the including table.
			checkedTable, lookupTable = lookUp.GetTableCanonicalOrder(lookup)
			tableName                 = lookUp.NameTable(lookupTable)
			// includedFilters stores the query.IncludedFilter parameter. If the
			// query has no includedFilters on the Included side. Then this is
			// left as nil.
			includedFilter ifaces.Column
		)

		if lookup.IsFilteredOnIncluding() {
			var (
				checkedLen = checkedTable[0].Size()
				ones       = verifiercol.NewConstantCol(field.One(), checkedLen)
			)

			checkedTable = append([]ifaces.Column{ones}, checkedTable...)
			for frag := range lookupTable {
				lookupTable[frag] = append([]ifaces.Column{lookup.IncludingFilter[frag]}, lookupTable[frag]...)
			}

			tableName = lookUp.NameTable(lookupTable)
		}

		if lookup.IsFilteredOnIncluded() {
			includedFilter = lookup.IncludedFilter
		}

		// In case this is the first iteration where we encounter the lookupTable
		// we need to add entries in the registering maps.
		if _, ok := ctx.CheckedTables[tableName]; !ok {
			ctx.IncludedFilters[tableName] = []ifaces.Column{}
			ctx.CheckedTables[tableName] = []table{}
			ctx.LookupTables = [][]table{}
			ctx.Rounds[tableName] = 0
		}

		// Add only the relevant part to the context
		if relevantPart == "Both" {
			ctx.IncludedFilters[tableName] = append(ctx.IncludedFilters[tableName], includedFilter)
			ctx.CheckedTables[tableName] = append(ctx.CheckedTables[tableName], checkedTable)
			ctx.LookupTables = append(ctx.LookupTables, lookupTable)
		}
		if relevantPart == "S" {
			ctx.IncludedFilters[tableName] = append(ctx.IncludedFilters[tableName], includedFilter)
			ctx.CheckedTables[tableName] = append(ctx.CheckedTables[tableName], checkedTable)
		} else if relevantPart == "T" {
			ctx.LookupTables = append(ctx.LookupTables, lookupTable)
		}

		ctx.Rounds[tableName] = max(ctx.Rounds[tableName], moduleComp.QueriesNoParams.Round(lookup.ID))

	}

	return ctx
}

// determineRelevantPart checks if the lookup query involves columns from the module and returns the relevant part (S or T).
func determineRelevantPart(lookup query.Inclusion, moduleColumns column.Store) string {
	var hasT, hasS bool
	// Check if any column in S part is in the module
	if moduleColumns.Exists(lookup.Included[0].GetColID()) {
		hasS = true
	}

	// Check if any column in T part is in the module
	if moduleColumns.Exists(lookup.Including[0][0].GetColID()) {
		hasT = true
	}
	if hasT && hasS {
		return "Both"
	}
	if hasT && !hasS {
		return "T"
	}
	if hasS && !hasT {
		return "S"
	} else {
		panic("the module contains a query that is not relevant to any column")
	}

}
