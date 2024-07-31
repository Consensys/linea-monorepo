package acc_module

import (
	"github.com/consensys/linea-monorepo/prover/maths/common/smartvectors"
	"github.com/consensys/linea-monorepo/prover/maths/field"
	"github.com/consensys/linea-monorepo/prover/protocol/ifaces"
	"github.com/consensys/linea-monorepo/prover/protocol/wizard"
)

// Extract a shallow copy of the active zone of a column. Meaning the unpadded
// area where the column encodes actual data.
func extractColLeftPadded(
	run *wizard.ProverRuntime,
	col ifaces.Column,
) []field.Element {

	// Fetchs the smart-vector and delimitate the active zone. Here we assume
	// that all the columns are zero-prepended. And have the same length. That's
	// why stop - density gives us the starting position for scanning the
	// witness.
	var (
		col_    = run.Columns.MustGet(col.GetColID())
		density = smartvectors.Density(col_)
		stop    = col_.Len()
		start   = stop - smartvectors.Density(col_)
	)

	// Calling subvector would result in an exception. Thus, we treat it as a
	// special case and return an empty vector.
	if density == 0 {
		return []field.Element{}
	}

	// Extract the assignments through a shallow copy.
	return col_.SubVector(start, stop).IntoRegVecSaveAlloc()
}

func extractColRightPadded(
	run *wizard.ProverRuntime,
	col ifaces.Column,
) []field.Element {

	// Fetchs the smart-vector and delimitate the active zone. Here we assume
	// that all the columns are zero-prepended. And have the same length. That's
	// why stop - density gives us the starting position for scanning the
	// witness.
	var (
		col_    = run.Columns.MustGet(col.GetColID())
		density = smartvectors.Density(col_)
		start   = 0
		stop    = start + density
	)

	// Calling subvector would result in an exception. Thus, we treat it as a
	// special case and return an empty vector.
	if density == 0 {
		return []field.Element{}
	}

	// Extract the assignments through a shallow copy.
	return col_.SubVector(start, stop).IntoRegVecSaveAlloc()
}
