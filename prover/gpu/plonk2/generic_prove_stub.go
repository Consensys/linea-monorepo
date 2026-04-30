//go:build !cuda

package plonk2

import (
	"github.com/consensys/gnark/backend"
	gnarkplonk "github.com/consensys/gnark/backend/plonk"
	"github.com/consensys/gnark/backend/witness"
)

func proveGenericGPUBackend(
	b *genericGPUBackend,
	fullWitness witness.Witness,
	_ ...backend.ProverOption,
) (gnarkplonk.Proof, error) {
	if err := genericPlainSolvePreflight(b.ccs, fullWitness); err != nil {
		return nil, err
	}
	return nil, errGPUProverNotWired
}
