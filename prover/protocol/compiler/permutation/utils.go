package permutation

import (
	"github.com/consensys/linea-monorepo/prover/protocol/query"
	"github.com/consensys/linea-monorepo/prover/protocol/wizardutils"
	"github.com/consensys/linea-monorepo/prover/protocol/zk"
)

const permutationStr = "PERMUTATION"

// deriveName constructs a name for the permutation context
func deriveName[R ~string, T zk.Element](q query.Permutation[T], ss ...any) R {
	ss = append([]any{permutationStr, q}, ss...)
	return wizardutils.DeriveName[R](ss...)
}

// deriveName constructs a name for the permutation context
func deriveNameGen[R ~string](ss ...any) R {
	ss = append([]any{permutationStr}, ss...)
	return wizardutils.DeriveName[R](ss...)
}
