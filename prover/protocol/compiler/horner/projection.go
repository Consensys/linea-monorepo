package horner

import (
	"github.com/consensys/linea-monorepo/prover/protocol/wizard"
	"github.com/consensys/linea-monorepo/prover/protocol/zk"
)

// CompileProjection compiles [query.Projection] queries
func CompileProjection[T zk.Element](comp *wizard.CompiledIOP[T]) {
	ProjectionToHorner(comp)
	CompileHorner(comp)
}
