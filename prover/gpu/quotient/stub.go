//go:build !cuda

// Stub for non-CUDA builds. Guard calls with gpu.Enabled.
package quotient

import (
	"github.com/consensys/linea-monorepo/prover/gpu"
	"github.com/consensys/linea-monorepo/prover/protocol/ifaces"
	"github.com/consensys/linea-monorepo/prover/protocol/wizard"
	"github.com/consensys/linea-monorepo/prover/symbolic"
)

func RunGPU(
	_ *gpu.Device,
	_ *wizard.ProverRuntime,
	_ int,
	_ []int,
	_ []symbolic.ExpressionBoard,
	_ [][]ifaces.Column,
	_ [][]ifaces.Column,
	_ [][]ifaces.Column,
	_ map[int][]int,
) error {
	panic("gpu: cuda required")
}
