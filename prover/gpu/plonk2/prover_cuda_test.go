//go:build cuda

package plonk2

import (
	"math/big"
	"os"
	"path/filepath"
	"testing"

	"github.com/consensys/gnark-crypto/ecc"
	blsfr "github.com/consensys/gnark-crypto/ecc/bls12-377/fr"
	bnfr "github.com/consensys/gnark-crypto/ecc/bn254/fr"
	bwfr "github.com/consensys/gnark-crypto/ecc/bw6-761/fr"
	gnarkplonk "github.com/consensys/gnark/backend/plonk"
	blsplonk "github.com/consensys/gnark/backend/plonk/bls12-377"
	bnplonk "github.com/consensys/gnark/backend/plonk/bn254"
	bwplonk "github.com/consensys/gnark/backend/plonk/bw6-761"
	"github.com/consensys/gnark/frontend"
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

func TestFullProverEnabledBLS12377GPUBackend_CUDA(t *testing.T) {
	dev, err := gpu.New()
	require.NoError(t, err, "creating CUDA device should succeed")
	defer dev.Close()

	setup := newFullProverBLS12377Setup(t, 128)
	defer setup.gpk.Close()
	tracePath := filepath.Join(t.TempDir(), "plonk2-gpu-trace.jsonl")

	prover, err := NewProver(
		dev,
		setup.spr,
		setup.pk,
		WithEnabled(true),
		WithStrictMode(true),
		WithLegacyBLS12377Backend(true),
		WithTrace(tracePath),
	)
	require.NoError(t, err, "creating enabled BLS12-377 GPU prover should succeed")
	defer prover.Close()

	proof, err := prover.Prove(setup.fullWitness)
	require.NoError(t, err, "BLS12-377 GPU proof should succeed")
	require.NoError(t, gnarkplonk.Verify(proof, setup.vk, setup.publicWitness), "GPU proof should verify")

	contents, err := os.ReadFile(tracePath)
	require.NoError(t, err, "trace file should be written")
	require.Contains(t, string(contents), `"phase":"legacy_bls12_377_gpu_prove"`, "trace should record the legacy GPU proof phase")
	require.Contains(t, string(contents), `"backend":"legacy_bls12_377_gpu"`, "trace should identify the backend")
}

func TestFullProverEnabledGenericBackendSelection_CUDA(t *testing.T) {
	dev, err := gpu.New()
	require.NoError(t, err, "creating CUDA device should succeed")
	defer dev.Close()

	for _, curveID := range []ecc.ID{ecc.BN254, ecc.BLS12_377, ecc.BW6_761} {
		t.Run(curveID.String(), func(t *testing.T) {
			fixture := newProverAPIFixture(t, curveID)
			prover, err := NewProver(dev, fixture.ccs, fixture.pk, WithEnabled(true))
			require.NoError(t, err, "creating enabled generic prover should succeed")
			defer prover.Close()

			require.NotNil(t, prover.gpuBackend, "enabled prover should select a GPU backend")
			require.Equal(t, "generic_gpu", prover.gpuBackend.Label(), "default enabled backend should be generic")
		})
	}
}

func TestFullProverStrictGenericBackendDoesNotFallback_CUDA(t *testing.T) {
	dev, err := gpu.New()
	require.NoError(t, err, "creating CUDA device should succeed")
	defer dev.Close()

	for _, curveID := range []ecc.ID{ecc.BN254, ecc.BLS12_377, ecc.BW6_761} {
		t.Run(curveID.String(), func(t *testing.T) {
			setup := newFullProverSetup(t, curveID, 128)
			tracePath := filepath.Join(t.TempDir(), "plonk2-generic-trace.jsonl")

			prover, err := NewProver(
				dev,
				setup.ccs,
				setup.pk,
				WithEnabled(true),
				WithStrictMode(true),
				WithTrace(tracePath),
			)
			require.NoError(t, err, "creating strict generic prover should succeed")
			defer prover.Close()

			proof, err := prover.Prove(setup.fullWitness)
			require.NoError(t, err, "strict generic GPU proof should succeed")
			require.NoError(t, gnarkplonk.Verify(proof, setup.vk, setup.publicWitness), "generic GPU proof should verify")

			contents, err := os.ReadFile(tracePath)
			require.NoError(t, err, "trace file should be written")
			require.Contains(t, string(contents), `"phase":"generic_gpu_prepare"`, "trace should record generic preparation")
			require.Contains(t, string(contents), `"phase":"generic_gpu_prove"`, "trace should record generic prove attempt")
			require.Contains(t, string(contents), `"phase":"generic_gpu_prove_done"`, "trace should record generic prove success")
			require.NotContains(t, string(contents), `"phase":"cpu_fallback"`, "strict generic path must not trace CPU fallback")
		})
	}
}

func TestFullProverStrictGenericInvalidWitnessFailsBeforeUnwired_CUDA(t *testing.T) {
	dev, err := gpu.New()
	require.NoError(t, err, "creating CUDA device should succeed")
	defer dev.Close()

	for _, curveID := range []ecc.ID{ecc.BN254, ecc.BLS12_377, ecc.BW6_761} {
		t.Run(curveID.String(), func(t *testing.T) {
			setup := newFullProverSetup(t, curveID, 128)
			assignment := newBenchChainAssignment(curveID, 128)
			assignment.Y = new(big.Int).Add(assignment.Y.(*big.Int), big.NewInt(1))
			invalidWitness, err := frontend.NewWitness(assignment, curveID.ScalarField())
			require.NoError(t, err, "creating invalid witness should succeed")

			prover, err := NewProver(
				dev,
				setup.ccs,
				setup.pk,
				WithEnabled(true),
				WithStrictMode(true),
			)
			require.NoError(t, err, "creating strict generic prover should succeed")
			defer prover.Close()

			_, err = prover.Prove(invalidWitness)
			require.Error(t, err, "invalid witness should fail in the generic strict path")
			require.NotErrorIs(t, err, errGPUProverNotWired, "invalid witness failure should not be hidden by readiness error")
		})
	}
}

func TestFullProverStrictGenericTamperedProofFailsVerification_CUDA(t *testing.T) {
	dev, err := gpu.New()
	require.NoError(t, err, "creating CUDA device should succeed")
	defer dev.Close()

	for _, curveID := range []ecc.ID{ecc.BN254, ecc.BLS12_377, ecc.BW6_761} {
		t.Run(curveID.String(), func(t *testing.T) {
			setup := newFullProverSetup(t, curveID, 128)
			prover, err := NewProver(dev, setup.ccs, setup.pk, WithEnabled(true), WithStrictMode(true))
			require.NoError(t, err, "creating strict generic prover should succeed")
			defer prover.Close()

			proof, err := prover.Prove(setup.fullWitness)
			require.NoError(t, err, "strict generic GPU proof should succeed")
			require.NoError(t, gnarkplonk.Verify(proof, setup.vk, setup.publicWitness), "untampered proof should verify")
			tamperProofClaim(t, proof)
			require.Error(t, gnarkplonk.Verify(proof, setup.vk, setup.publicWitness), "tampered GPU proof should fail verification")
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

func tamperProofClaim(t *testing.T, proof gnarkplonk.Proof) {
	t.Helper()
	switch typed := proof.(type) {
	case *bnplonk.Proof:
		one := bnfr.One()
		typed.BatchedProof.ClaimedValues[0].Add(&typed.BatchedProof.ClaimedValues[0], &one)
	case *blsplonk.Proof:
		one := blsfr.One()
		typed.BatchedProof.ClaimedValues[0].Add(&typed.BatchedProof.ClaimedValues[0], &one)
	case *bwplonk.Proof:
		one := bwfr.One()
		typed.BatchedProof.ClaimedValues[0].Add(&typed.BatchedProof.ClaimedValues[0], &one)
	default:
		t.Fatalf("unexpected proof type %T", proof)
	}
}
