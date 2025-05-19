// pragmas allows to define some pragma that are used by the distributed protocol compiler
package pragmas

import (
	"github.com/consensys/linea-monorepo/prover/maths/field"
	"github.com/consensys/linea-monorepo/prover/protocol/column"
	"github.com/consensys/linea-monorepo/prover/protocol/ifaces"
)

type Pragma = string

const (
	// FullColumnPragma is used to mark that a column is a full column.
	// This will help the module discoverer to determine that the column is
	// actually full-density and not a regular column with a suboptimal representation.
	FullColumnPragma Pragma = "fullcolumn-pragma"

	// RightPadded is used to indicate that a column is right-padded
	RightPadded Pragma = "rightpadded-pragma"

	// LeftPadded is used to indicate that a column is left-padded
	LeftPadded Pragma = "leftpadded-pragma"

	None Pragma = "none-pragma"

	// Paddable is used to indicate that a column is paddable
	Paddable Pragma = "paddable-pragma"
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

// IsRightPadded checks if a column is right-padded.
func IsRightPadded(col ifaces.Column) bool {
	nat := col.(column.Natural)
	_, ok := nat.GetPragma(RightPadded)
	return ok
}

// IsLeftPadded checks if a column is left-padded.
func IsLeftPadded(col ifaces.Column) bool {
	nat := col.(column.Natural)
	_, ok := nat.GetPragma(LeftPadded)
	return ok
}

// MarkRightPadded marks a column as right-padded.
func MarkRightPadded(col ifaces.Column) {
	nat := col.(column.Natural)
	nat.SetPragma(RightPadded, true)
}

// MarkLeftPadded marks a column as left-padded.
func MarkLeftPadded(col ifaces.Column) {
	nat := col.(column.Natural)
	nat.SetPragma(LeftPadded, true)
}

// IsPaddable checks if a column is paddable.
func IsPaddable(col ifaces.Column) (field.Element, bool) {
	nat := col.(column.Natural)
	v, ok := nat.GetPragma(Paddable)
	var v_ field.Element
	if ok {
		v_ = v.(field.Element)
	}
	return v_, ok
}

// MarkPaddable marks a column as paddable.
func MarkPaddable(col ifaces.Column, v field.Element) {
	nat := col.(column.Natural)
	nat.SetPragma(Paddable, v)
}
