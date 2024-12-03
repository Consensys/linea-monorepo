package permutation

import (
	"github.com/consensys/linea-monorepo/prover/protocol/query"
	"github.com/consensys/linea-monorepo/prover/protocol/wizardutils"
)

const permutationStr = "PERMUTATION"

// DeriveName constructs a name for the permutation context
func DeriveName[R ~string](q query.Permutation, ss ...any) R {
	ss = append([]any{permutationStr, q}, ss...)
	return wizardutils.DeriveName[R](ss...)
}

// deriveName constructs a name for the permutation context
func deriveNameGen[R ~string](ss ...any) R {
	ss = append([]any{permutationStr}, ss...)
	return wizardutils.DeriveName[R](ss...)
}
