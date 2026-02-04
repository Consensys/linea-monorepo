package execution

import (
	public_input "github.com/consensys/linea-monorepo/prover/circuits/pi-interconnection/keccak/prover/public-input"
	"github.com/consensys/linea-monorepo/prover/circuits/pi-interconnection/keccak/prover/zkevm"
)

type Witness struct {
	FuncInp *public_input.Execution
	ZkEVM   *zkevm.Witness
}

// mustProveAndPass the prover (in the void). Does not takes a
// prover-step function performing the assignment but a function
// returning such a function. This is important to avoid side-effects
// when calling it twice.

// SanityCheckTraceChecksum ensures the checksum for the traces in the setup matches the one in the config
