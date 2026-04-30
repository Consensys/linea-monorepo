//go:build !cuda

package plonk2

import gnarkplonk "github.com/consensys/gnark/backend/plonk"

func newLegacyProverGPUBackend(
	_ any,
	_ any,
	_ gnarkplonk.ProvingKey,
) (proverGPUBackend, error) {
	return nil, errGPUProverNotWired
}
