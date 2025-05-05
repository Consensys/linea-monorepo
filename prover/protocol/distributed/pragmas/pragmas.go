// pragmas allows to define some pragma that are used by the distributed protocol compiler
package pragmas

import (
	"github.com/consensys/linea-monorepo/prover/protocol/column"
	"github.com/consensys/linea-monorepo/prover/protocol/ifaces"
)

const (
	// FullColumnPragma is used to mark that a column is a full column.
	// This will help the module discoverer to determine that the column is
	// actually full-density and not a regular column with a suboptimal representation.
	FullColumnPragma = "fullcolumn-pragma"
)

// MarkFullColumn marks a column as full-column.
func MarkFullColumn(col ifaces.Column) {
	nat := col.(column.Natural)
	nat.SetPragma(FullColumnPragma, true)
}

// IsFullColumn checks if a column is a full-column.
func IsFullColumn(col ifaces.Column) bool {
	nat := col.(column.Natural)
	_, ok := nat.GetPragma(FullColumnPragma)
	return ok
}
