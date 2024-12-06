package inclusion

import (
	"github.com/consensys/linea-monorepo/prover/protocol/distributed/query"
	"github.com/consensys/linea-monorepo/prover/protocol/wizard"
)

// IntoLogDerivativeSum reduces all the inclusion queries into a LogDerivativeSum query.
func IntoLogDerivativeSum(comp *wizard.CompiledIOP) query.LogDerivativeSum {
	panic("unimplemented")
	// - scan the compiler for the inclusion queries, and
	// group them based on their lookup table (different S with the same T).
	// - declare two coins. The coins should not create any new round here.
	// So we declare them as variable rather than inserting them.
	// var alpha, beta coin.Info
	// - build the expressions for the LogDerivativeSum from the columns and the coins.
	// - return LogDerivatevSum.
}

// CompileDist compiles a LogDerivativeSum query distributedly.
func CompileDist(comp *wizard.CompiledIOP) {

}
