package inclusion

import "github.com/consensys/linea-monorepo/prover/protocol/ifaces"

// MainLookupCtx stores the compilation context of all the lookup queries
// altogether.
type mainLookupCtx struct {

	// LookupTables stores all the lookup table the compiler encounters. They are
	// sorted in canonical order. This used to derive a determistic ordering
	// of the lookup LookupTables. (We want to ensure the compiler yields always
	// exactly the same result for replicability).
	//
	// To illustrates its structure, the following sub-statement
	//
	// 		table[numTable][frag]
	//
	// refers to to the fragment #frag of the the table #numTable.
	LookupTables [][]table

	// CheckedTables stores all the checked column by lookup table. The key is
	// obtained as nameTable(lookupTable) where lookup is sorted in
	// canonical order.
	CheckedTables map[string][]table

	// IncludedFilters stores all the filters for the checked columns and `nil`
	// if no filter is applied. As for [checkedTables] they are stored by
	// lookup table name and in the same order for each key.
	IncludedFilters map[string][]ifaces.Column

	// Rounds stores the interaction round assigned to each lookupTable. The
	// round is obtained by taking the max of the declaration Rounds of the
	// Inclusion queries using the corresponding lookup table.
	Rounds map[string]int

	// it stores the multiplicity of T is S.
	//  For the multiColum case it collapse T and S and then counts the multiplicity.
	// note that here the collapsing does not need the same randomness as the compilation.
	// since the multiplicity is the same w.r.t any randomness.
	mTables map[string]table
}
