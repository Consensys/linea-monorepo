package lookup

import (
	"github.com/consensys/linea-monorepo/prover/protocol/coin"
	"github.com/consensys/linea-monorepo/prover/protocol/ifaces"
	"github.com/consensys/linea-monorepo/prover/symbolic"
)

// mainLookupCtx stores the compilation context of all the lookup queries
// altogether.
type mainLookupCtx struct {

	// lookupTables stores all the lookup table the compiler encounters. They are
	// sorted in canonical order. This used to derive a determistic ordering
	// of the lookup lookupTables. (We want to ensure the compiler yields always
	// exactly the same result for replicability).
	//
	// To illustrates its structure, the following sub-statement
	//
	// 		table[numTable][frag]
	//
	// refers to to the fragment #frag of the the table #numTable.
	lookupTables [][]table

	// checkedTables stores all the checked column by lookup table. The key is
	// obtained as nameTable(lookupTable) where lookup is sorted in
	// canonical order.
	checkedTables map[string][]table

	// includedFilters stores all the filters for the checked columns and `nil`
	// if no filter is applied. As for [checkedTables] they are stored by
	// lookup table name and in the same order for each key.
	includedFilters map[string][]ifaces.Column

	// rounds stores the interaction round assigned to each lookupTable. The
	// round is obtained by taking the max of the declaration rounds of the
	// Inclusion queries using the corresponding lookup table.
	rounds map[string]int
}

// singleTableCtx stores the compilation context for a single lookup query
// when it is compiled using the log-derivative lookup technique.
//
// A singleTableCtx relates to a lookup table rather than a lookup query. This
// means that multiple lookup queries that are related to the same table will be
// grouped into the same context. This allows optimizing the
type singleTableCtx struct {

	// TableName reflects the name of the lookup table being compiled.
	TableName string

	// M is the column storing the multiplicities of the entries of T within S.
	M []ifaces.Column

	// Gamma is the coin used to evaluate the sum of the inverse of the columns.
	Gamma coin.Info

	// S represents the list of the looked-up tables. Each entry of S
	// corresponds to a lookup query into T. The expressions that are stored
	// are either Variables pointing the related columns (in the single-column
	// lookup case) or random linear combinations (in Alpha) when the lookup
	// query is a multi-column lookup.
	S []*symbolic.Expression

	// SFilter stores the filters that are applied on S as column and `nil` if
	// no filter is applied over the column.
	SFilters []ifaces.Column

	// T represents the look-up table being currently compiled. The expression
	// is a variable if the lookup table has only a single column or a random
	// linear combination of column if the table has only a single column.
	T []*symbolic.Expression

	Name string
}
