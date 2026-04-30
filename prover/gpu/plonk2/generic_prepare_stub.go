//go:build !cuda

package plonk2

import (
	gnarkplonk "github.com/consensys/gnark/backend/plonk"
	"github.com/consensys/gnark/constraint"

	"github.com/consensys/linea-monorepo/prover/gpu"
)

type genericProverState struct{}

func newGenericProverState(
	_ *gpu.Device,
	_ constraint.ConstraintSystem,
	_ gnarkplonk.ProvingKey,
) (*genericProverState, error) {
	return nil, errGPUProverNotWired
}

func (s *genericProverState) Close() {}

func (s *genericProverState) fixedPolynomialCount() int { return 0 }
