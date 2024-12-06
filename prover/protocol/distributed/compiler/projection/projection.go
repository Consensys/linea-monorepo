package projection

import (
	"github.com/consensys/linea-monorepo/prover/protocol/coin"
	"github.com/consensys/linea-monorepo/prover/protocol/wizard"
)

// CompileDist compiles a GrandSum query distributedly.
// It receives a compiledIop object relevant to a segment.
// The seed is a random coin from randomness beacon (FS of all LPP commitments).
// All the compilation steps are similar to the permutation compilation apart from:
//   - random coins \alpha and \gamma are generated from the seed.
//   - no verifierAction is needed over the ZOpening.
//   - ZOpenings are declared as public input.
func CompileDist(comp *wizard.CompiledIOP, seed coin.Info) {

}
