package logderivativesum

import (
	"github.com/consensys/linea-monorepo/prover/protocol/wizard"
	"github.com/consensys/linea-monorepo/prover/protocol/zk"
)

// CompileLookups compiles the [query.Inclusion] queries that are present
// in comp using the [LogDerivativeSum] approach. First, the queries are
// grouped into a big [LogDerivativeSum] query and then, the query is
// compiled using global-constraints and local-constraints.
func CompileLookups[T zk.Element](comp *wizard.CompiledIOP[T]) {
	LookupIntoLogDerivativeSum[T](comp)
	CompileLogDerivativeSum[T](comp)
}
