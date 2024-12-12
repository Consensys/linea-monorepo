package distributed

import (
	"github.com/consensys/linea-monorepo/prover/protocol/compiler/innerproduct"
	"github.com/consensys/linea-monorepo/prover/protocol/compiler/lookup"
	"github.com/consensys/linea-monorepo/prover/protocol/compiler/mimc"
	"github.com/consensys/linea-monorepo/prover/protocol/compiler/specialqueries"
	"github.com/consensys/linea-monorepo/prover/protocol/ifaces"
	"github.com/consensys/linea-monorepo/prover/protocol/wizard"
)

// prepare reduces any query to LPP or GL.
// it prepares the columns that depends on whole the witness,e.g., M column for lookups.
func prepare(comp *wizard.CompiledIOP) {
	mimc.CompileMiMC(comp)
	specialqueries.RangeProof(comp)
	specialqueries.CompileFixedPermutations(comp)
	innerproduct.Compile(comp)

	prepareLookup(comp)
}

// It scans the initial compiledIOP and group all the checkedTables related to the same lookupTable,
// It creates a multiplicity column M for all such pairs (checkedTables, lookupTable).
func prepareLookup(comp *wizard.CompiledIOP) {
	mainLookupCtx := lookup.CaptureLookupTables(comp)

	for _, lookupTable := range mainLookupCtx.LookupTables {
		var (
			tableName = lookup.NameTable(lookupTable)
			round     = mainLookupCtx.Rounds[tableName]
			size      = lookupTable[0][0].Size()
			fragNum   = len(lookupTable[0])
			mTable    = make([]ifaces.Column, fragNum)
		)
		for frag := range lookupTable[0] {

			mTableID := ifaces.ColIDf("%v_%v_%v", tableName, "M", frag)
			mTable[frag] = comp.InsertCommit(round, mTableID, size)
		}

	}
}
