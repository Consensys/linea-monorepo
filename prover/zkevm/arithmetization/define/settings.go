package define

import "github.com/consensys/linea-monorepo/prover/config"

// Settings specifies the parameters for the arithmetization part of the
// zk-EVM.
type Settings struct {
	// Configuration object specifying the columns limits
	Traces        *config.TracesLimits
	ColDepthLimit int
	NumColLimit   int
}
