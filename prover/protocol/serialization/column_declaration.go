package serialization

import (
	"github.com/consensys/zkevm-monorepo/prover/protocol/column"
	"github.com/consensys/zkevm-monorepo/prover/protocol/ifaces"
	"github.com/consensys/zkevm-monorepo/prover/protocol/wizard"
)

// serializableColumnDecl is used to represent a "natural" column, meaning a
// column that is explicitly registered as part of the scheme. This is in
// oppposition with [serializableColumnRef] where such columns are encoded
// by just citing their names.
//
// Concretely, we need this because [column.Natural] has a complex structure
// that is deeply nested within [column.Store]. And this prevents directly
// applying the generic reflection-based serialization logic to it.
type serializableColumnDecl struct {
	Name   ifaces.ColID
	Round  int
	Status column.Status
	Size   int
}

// The function takes a Natural column as parameter rather than an
// [ifaces.Column]
func intoSerializableColDecl(c *column.Natural) *serializableColumnDecl {
	return &serializableColumnDecl{
		Name:   c.ID,
		Round:  c.Round(),
		Status: c.Status(),
		Size:   c.Size(),
	}
}

func (c *serializableColumnDecl) intoNaturalAndRegister(comp *wizard.CompiledIOP) ifaces.Column {
	return comp.InsertColumn(c.Round, c.Name, c.Size, c.Status)
}
