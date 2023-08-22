package verifiercol

import (
	"github.com/consensys/accelerated-crypto-monorepo/protocol/column"
	"github.com/consensys/accelerated-crypto-monorepo/protocol/ifaces"
	"github.com/consensys/accelerated-crypto-monorepo/protocol/wizard"
	"github.com/consensys/accelerated-crypto-monorepo/utils"
)

type VerifierCol interface {
	ifaces.Column
	Split(comp *wizard.CompiledIOP, from, to int) ifaces.Column
}

// AssertIsPublicCol returns true if the column is public
// TODO @AlexandreBelling, this function seems at the wrong place in
// this package. We should consider rework the package organization
// to make it work in a cleaner manner.
func AssertIsPublicCol(comp *wizard.CompiledIOP, col ifaces.Column) {
	// Sanity-check, the columns should be accessible to the verifier
	grandParents := column.RootParents(col)
	for _, grandpa := range grandParents {
		// Careful, that the grandpa might be a verifier defined column
		// in this case, this is OK because this corresponds to a public
		// column but it will panic if we try to get its status so we need
		// to check first.
		if _, ok := grandpa.(VerifierCol); ok {
			continue
		}

		// Else, we can deduce the public-hood of the column from looking
		// at its
		status := comp.Columns.Status(grandpa.GetColID())
		if !status.IsPublic() {
			utils.Panic(
				"commitment %v has grandpa %v of size %v but is not public (status %v)",
				col, grandpa.GetColID(), grandpa.Size(), status.String(),
			)
		}
	}
}
