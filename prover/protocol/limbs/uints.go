package limbs

import (
	"github.com/consensys/linea-monorepo/prover/protocol/distributed/pragmas"
	"github.com/consensys/linea-monorepo/prover/protocol/ifaces"
	"github.com/consensys/linea-monorepo/prover/protocol/wizard"
	"github.com/consensys/linea-monorepo/prover/utils"
)

// Uint[S, E] represents a register represented by a list of columns.
type Uint[S Size, E Endianness] struct {
	limbs[E]
}

// NewUint[S, E] creates a new [Uints] registering all its components.
func NewUint[S Size, E Endianness](comp *wizard.CompiledIOP, name ifaces.ColID, size int, prags ...pragmas.PragmaPair) Uint[S, E] {
	numLimbs := utils.DivExact(uintSize[S](), limbBitWidth)
	limbs := createLimb[E](comp, name, numLimbs, size, prags...)
	return Uint[S, E]{limbs: limbs}
}
