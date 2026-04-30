//go:build cuda

package plonk2

import (
	"fmt"
	"testing"

	"github.com/consensys/gnark-crypto/ecc"
	gnarkbackend "github.com/consensys/gnark/backend"
	gnarkplonk "github.com/consensys/gnark/backend/plonk"
	blsplonk "github.com/consensys/gnark/backend/plonk/bls12-377"
	bnplonk "github.com/consensys/gnark/backend/plonk/bn254"
	bwplonk "github.com/consensys/gnark/backend/plonk/bw6-761"
	"github.com/consensys/gnark/constraint"
	"github.com/consensys/gnark/frontend"
	"github.com/consensys/gnark/frontend/cs/scs"
	"github.com/consensys/gnark/test/unsafekzg"
	"github.com/stretchr/testify/require"

	"github.com/consensys/linea-monorepo/prover/gpu"
)

type e2eCommitCircuit struct {
	X frontend.Variable
	Y frontend.Variable
	Z frontend.Variable `gnark:",public"`
}

func (c *e2eCommitCircuit) Define(api frontend.API) error {
	committer, ok := api.(frontend.Committer)
	if !ok {
		return fmt.Errorf("frontend does not implement Committer")
	}
	commitment, err := committer.Commit(c.X, c.Y)
	if err != nil {
		return err
	}
	api.AssertIsEqual(api.Mul(c.X, c.Y), c.Z)
	api.AssertIsEqual(api.Sub(commitment, commitment), 0)
	return nil
}

func TestGenericGPUBuildArtifactsCommitmentCircuit_CUDA(t *testing.T) {
	dev, err := gpu.New()
	require.NoError(t, err, "creating CUDA device should succeed")
	defer dev.Close()

	for _, curveID := range []ecc.ID{ecc.BN254, ecc.BLS12_377, ecc.BW6_761} {
		t.Run(curveID.String(), func(t *testing.T) {
			ccs, err := frontend.Compile(
				curveID.ScalarField(),
				scs.NewBuilder,
				&e2eCommitCircuit{},
			)
			require.NoError(t, err, "compiling commitment circuit should succeed")
			require.True(t, hasPlonkCommitments(ccs), "compiled circuit should have commitment metadata")

			srs, srsLagrange, err := unsafekzg.NewSRS(ccs)
			require.NoError(t, err, "creating unsafe test SRS should succeed")
			pk, _, err := gnarkplonk.Setup(ccs, srs, srsLagrange)
			require.NoError(t, err, "PlonK setup should succeed")

			prover, err := NewProver(dev, ccs, pk, WithEnabled(true), WithStrictMode(true))
			require.NoError(t, err, "creating strict generic prover should succeed")
			defer prover.Close()
			backend, ok := prover.gpuBackend.(*genericGPUBackend)
			require.True(t, ok, "enabled prover should use generic backend")

			fullWitness, err := frontend.NewWitness(
				&e2eCommitCircuit{X: 3, Y: 11, Z: 33},
				curveID.ScalarField(),
			)
			require.NoError(t, err, "creating commitment witness should succeed")

			proverConfig, err := gnarkbackend.NewProverConfig()
			require.NoError(t, err, "default prover config should be valid")
			artifacts, err := backend.buildArtifacts(fullWitness, &proverConfig)
			require.NoError(t, err, "generic backend should solve and commit artifacts")
			require.Len(t, artifacts.bsb22Raw, 1, "one BSB22 polynomial should be stored")
			require.NotEmpty(t, artifacts.bsb22Raw[0], "BSB22 polynomial should be stored")
			require.NotEmpty(t, artifacts.commitmentVals, "commitment value should be stored")
			require.NotEmpty(t, artifacts.betaRaw, "beta challenge should be stored")
			require.NotEmpty(t, artifacts.gammaRaw, "gamma challenge should be stored")
			require.NotEmpty(t, artifacts.alphaRaw, "alpha challenge should be stored")
			require.NotEmpty(t, artifacts.zRaw, "Z Lagrange polynomial should be stored")
			require.NotEmpty(t, artifacts.zCanonical, "Z canonical polynomial should be stored")
			require.Greater(t, len(artifacts.zBlinded), len(artifacts.zCanonical), "Z blinded polynomial should include blinding terms")
			for i := range artifacts.lroRaw {
				require.NotEmpty(t, artifacts.lroRaw[i], "LRO Lagrange polynomial should be stored")
				require.NotEmpty(t, artifacts.lroCanonical[i], "LRO canonical polynomial should be stored")
				require.Greater(
					t,
					len(artifacts.lroBlinded[i]),
					len(artifacts.lroCanonical[i]),
					"LRO blinded polynomial should include blinding terms",
				)
			}
			requireProofSkeletonHasCommitments(t, artifacts.proof, ccs)
		})
	}
}

func TestGenericGPUProverVerifiesCommitmentCircuit_CUDA(t *testing.T) {
	dev, err := gpu.New()
	require.NoError(t, err, "creating CUDA device should succeed")
	defer dev.Close()

	for _, curveID := range []ecc.ID{ecc.BN254, ecc.BLS12_377, ecc.BW6_761} {
		t.Run(curveID.String(), func(t *testing.T) {
			ccs, err := frontend.Compile(
				curveID.ScalarField(),
				scs.NewBuilder,
				&e2eCommitCircuit{},
			)
			require.NoError(t, err, "compiling commitment circuit should succeed")
			require.True(t, hasPlonkCommitments(ccs), "compiled circuit should have commitment metadata")

			srs, srsLagrange, err := unsafekzg.NewSRS(ccs)
			require.NoError(t, err, "creating unsafe test SRS should succeed")
			pk, vk, err := gnarkplonk.Setup(ccs, srs, srsLagrange)
			require.NoError(t, err, "PlonK setup should succeed")

			fullWitness, err := frontend.NewWitness(
				&e2eCommitCircuit{X: 3, Y: 11, Z: 33},
				curveID.ScalarField(),
			)
			require.NoError(t, err, "creating commitment witness should succeed")
			publicWitness, err := fullWitness.Public()
			require.NoError(t, err, "extracting public witness should succeed")

			prover, err := NewProver(dev, ccs, pk, WithEnabled(true), WithStrictMode(true))
			require.NoError(t, err, "creating strict generic prover should succeed")
			defer prover.Close()

			proof, err := prover.Prove(fullWitness)
			require.NoError(t, err, "generic GPU commitment proof should succeed")
			require.NoError(t, gnarkplonk.Verify(proof, vk, publicWitness), "generic GPU commitment proof should verify")
			requireProofSkeletonHasCommitments(t, proof, ccs)
		})
	}
}

func requireProofSkeletonHasCommitments(t *testing.T, proof any, ccs constraint.ConstraintSystem) {
	t.Helper()
	wantCommitments := plonkCommitmentCountForCCS(ccs)
	switch typed := proof.(type) {
	case *bnplonk.Proof:
		require.Len(t, typed.Bsb22Commitments, wantCommitments, "BN254 BSB22 commitment count should match metadata")
		require.False(t, typed.Bsb22Commitments[0].IsInfinity(), "BN254 BSB22 commitment should be non-zero")
		require.False(t, typed.Z.IsInfinity(), "BN254 Z commitment should be non-zero")
		for i := range typed.LRO {
			require.False(t, typed.LRO[i].IsInfinity(), "BN254 LRO commitment should be non-zero")
		}
	case *blsplonk.Proof:
		require.Len(t, typed.Bsb22Commitments, wantCommitments, "BLS12-377 BSB22 commitment count should match metadata")
		require.False(t, typed.Bsb22Commitments[0].IsInfinity(), "BLS12-377 BSB22 commitment should be non-zero")
		require.False(t, typed.Z.IsInfinity(), "BLS12-377 Z commitment should be non-zero")
		for i := range typed.LRO {
			require.False(t, typed.LRO[i].IsInfinity(), "BLS12-377 LRO commitment should be non-zero")
		}
	case *bwplonk.Proof:
		require.Len(t, typed.Bsb22Commitments, wantCommitments, "BW6-761 BSB22 commitment count should match metadata")
		require.False(t, typed.Bsb22Commitments[0].IsInfinity(), "BW6-761 BSB22 commitment should be non-zero")
		require.False(t, typed.Z.IsInfinity(), "BW6-761 Z commitment should be non-zero")
		for i := range typed.LRO {
			require.False(t, typed.LRO[i].IsInfinity(), "BW6-761 LRO commitment should be non-zero")
		}
	default:
		t.Fatalf("unexpected proof type %T", proof)
	}
}
