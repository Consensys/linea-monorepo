package wizard

import (
	"github.com/consensys/zkevm-monorepo/prover/crypto/fiatshamir"
	"github.com/consensys/zkevm-monorepo/prover/symbolic"
)

// Coin is an interface representing public random coins as part of a protocol.
type Coin interface {
	symbolic.Metadata
	Round() int
	sample(fs *fiatshamir.State) any
	id() id
	Explain() string
	Tags() []string
}
