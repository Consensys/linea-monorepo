package zkevm

import (
	"github.com/consensys/linea-monorepo/prover/circuits/pi-interconnection/keccak/prover/protocol/distributed"
)

var LimitlessCompilationParams = distributed.CompilationParams{
	FixedNbRowPlonkCircuit:       1 << 19,
	FixedNbRowExternalHasher:     1 << 15,
	FixedNbPublicInput:           1 << 10,
	InitialCompilerSize:          1 << 18,
	InitialCompilerSizeConglo:    1 << 13,
	ColumnProfileMPTS:            []int{17, 335, 37, 3, 5, 15, 0, 1},
	ColumnProfileMPTSPrecomputed: 22,
}

// LimitlessZkEVM defines the wizard responsible for proving execution of the EVM
// and the associated wizard circuits for the limitless prover protocol.
type LimitlessZkEVM struct {
	Zkevm      *ZkEvm
	DistWizard *distributed.DistributedWizard
}
