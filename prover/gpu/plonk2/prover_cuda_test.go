//go:build cuda

package plonk2

import (
	"testing"

	"github.com/consensys/gnark-crypto/ecc"
	gnarkplonk "github.com/consensys/gnark/backend/plonk"
	"github.com/stretchr/testify/require"

	"github.com/consensys/linea-monorepo/prover/gpu"
)

func TestFullProverCPUFallbackTargetCurves_CUDA(t *testing.T) {
	dev, err := gpu.New()
	require.NoError(t, err, "creating CUDA device should succeed")
	defer dev.Close()

	for _, curveID := range []ecc.ID{ecc.BN254, ecc.BLS12_377, ecc.BW6_761} {
		t.Run(curveID.String(), func(t *testing.T) {
			fixture := newProverAPIFixture(t, curveID)
			fullWitness := newMulWitness(t, curveID, 3, 11, 33)
			publicWitness, err := fullWitness.Public()
			require.NoError(t, err, "extracting public witness should succeed")

			prover, err := NewProver(dev, fixture.ccs, fixture.pk)
			require.NoError(t, err, "creating plonk2 prover should succeed")
			defer prover.Close()

			proof, err := prover.Prove(fullWitness)
			require.NoError(t, err, "CPU fallback proof should succeed")
			require.NoError(t, gnarkplonk.Verify(proof, fixture.vk, publicWitness), "fallback proof should verify")
		})
	}
}

func TestFullProverDisabledFallbackTargetCurves_CUDA(t *testing.T) {
	dev, err := gpu.New()
	require.NoError(t, err, "creating CUDA device should succeed")
	defer dev.Close()

	for _, curveID := range []ecc.ID{ecc.BN254, ecc.BLS12_377, ecc.BW6_761} {
		t.Run(curveID.String(), func(t *testing.T) {
			fixture := newProverAPIFixture(t, curveID)
			fullWitness := newMulWitness(t, curveID, 3, 11, 33)

			prover, err := NewProver(dev, fixture.ccs, fixture.pk, WithCPUFallback(false))
			require.NoError(t, err, "creating plonk2 prover should succeed")
			defer prover.Close()

			_, err = prover.Prove(fullWitness)
			require.ErrorIs(t, err, errGPUProverNotWired, "GPU prover is not wired yet")
		})
	}
}
