package projection

=======
import (
	"github.com/consensys/linea-monorepo/prover/protocol/coin"
	"github.com/consensys/linea-monorepo/prover/protocol/wizard"
)

// CompileDist compiles the projection queries distributedly.
// It receives a compiledIOP object relevant to a segment.
// The seed is a random coin from randomness beacon (FS of all LPP commitments).
// All the compilation steps are similar to the permutation compilation apart from:
//   - random coins \alpha and \gamma are generated from the seed (and the tableName).
//   - no verifierAction is needed over the ZOpening.
//   - ZOpenings are declared as public input.
func CompileDist(comp *wizard.CompiledIOP, seed coin.Info) {

}
