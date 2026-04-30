//go:build cuda

package plonk2

import (
	"testing"

	"github.com/consensys/gnark-crypto/ecc"
	blsfr "github.com/consensys/gnark-crypto/ecc/bls12-377/fr"
	blskzg "github.com/consensys/gnark-crypto/ecc/bls12-377/kzg"
	bnfr "github.com/consensys/gnark-crypto/ecc/bn254/fr"
	bnkzg "github.com/consensys/gnark-crypto/ecc/bn254/kzg"
	bwfr "github.com/consensys/gnark-crypto/ecc/bw6-761/fr"
	bwkzg "github.com/consensys/gnark-crypto/ecc/bw6-761/kzg"
	gnarkplonk "github.com/consensys/gnark/backend/plonk"
	blsplonk "github.com/consensys/gnark/backend/plonk/bls12-377"
	bnplonk "github.com/consensys/gnark/backend/plonk/bn254"
	bwplonk "github.com/consensys/gnark/backend/plonk/bw6-761"
	"github.com/consensys/gnark/backend/witness"
	csbls12377 "github.com/consensys/gnark/constraint/bls12-377"
	csbn254 "github.com/consensys/gnark/constraint/bn254"
	csbw6761 "github.com/consensys/gnark/constraint/bw6-761"
	"github.com/consensys/gnark/frontend"
	"github.com/consensys/gnark/frontend/cs/scs"
	"github.com/stretchr/testify/require"

	"github.com/consensys/linea-monorepo/prover/gpu"
)

func TestGenericProverStatePreparesFixedCommitments_AllTargetCurves_CUDA(t *testing.T) {
	dev, err := gpu.New()
	require.NoError(t, err, "creating CUDA device should succeed")
	defer dev.Close()

	for _, curveID := range []ecc.ID{ecc.BN254, ecc.BLS12_377, ecc.BW6_761} {
		t.Run(curveID.String(), func(t *testing.T) {
			ccs, err := frontend.Compile(
				curveID.ScalarField(),
				scs.NewBuilder,
				newBenchChainCircuit(128),
			)
			require.NoError(t, err, "compiling benchmark circuit should succeed")

			srs, srsLagrange := testSRSAssets(t).loadForCCS(t, ccs)
			pk, vk, err := gnarkplonk.Setup(ccs, srs, srsLagrange)
			require.NoError(t, err, "PlonK setup should succeed")

			prover, err := NewProver(dev, ccs, pk, WithEnabled(true))
			require.NoError(t, err, "creating enabled prover should prepare generic GPU state")
			defer prover.Close()
			state := prover.genericState
			require.NotNil(t, state, "generic GPU state should be prepared")
			require.NotNil(t, state.kzg, "canonical SRS MSM should be resident")
			require.NotNil(t, state.lagKzg, "lagrange SRS MSM should be resident")
			require.NotNil(t, state.fft, "FFT domain should be resident")
			require.NotNil(t, state.perm, "permutation table should be resident")
			require.Equal(t, 8, state.fixedPolynomialCount(), "benchmark circuit should have no Qcp selectors")

			qlCanonicalRaw := make([]uint64, state.fixed.ql.RawWords())
			require.NoError(t, state.fixed.ql.CopyToHostRaw(qlCanonicalRaw), "copying Ql should succeed")
			got, err := state.kzg.CommitRaw(qlCanonicalRaw)
			require.NoError(t, err, "committing prepared Ql should succeed")

			switch typedVK := vk.(type) {
			case *bnplonk.VerifyingKey:
				requireBN254ProjectiveMatches(t, typedVK.Ql, got, "prepared BN254 Ql")
			case *blsplonk.VerifyingKey:
				requireBLS12377ProjectiveMatches(t, typedVK.Ql, got, "prepared BLS12-377 Ql")
			case *bwplonk.VerifyingKey:
				requireBW6761ProjectiveMatches(t, typedVK.Ql, got, "prepared BW6-761 Ql")
			default:
				t.Fatalf("unexpected verifying key type %T", vk)
			}
		})
	}
}

func TestGenericProverStateCommitsSolvedWireLagrangePolynomials_CUDA(t *testing.T) {
	dev, err := gpu.New()
	require.NoError(t, err, "creating CUDA device should succeed")
	defer dev.Close()

	for _, curveID := range []ecc.ID{ecc.BN254, ecc.BLS12_377, ecc.BW6_761} {
		t.Run(curveID.String(), func(t *testing.T) {
			ccs, err := frontend.Compile(
				curveID.ScalarField(),
				scs.NewBuilder,
				newBenchChainCircuit(128),
			)
			require.NoError(t, err, "compiling benchmark circuit should succeed")
			srs, srsLagrange := testSRSAssets(t).loadForCCS(t, ccs)
			pk, _, err := gnarkplonk.Setup(ccs, srs, srsLagrange)
			require.NoError(t, err, "PlonK setup should succeed")

			prover, err := NewProver(dev, ccs, pk, WithEnabled(true))
			require.NoError(t, err, "creating enabled prover should prepare generic GPU state")
			defer prover.Close()

			fullWitness, err := frontend.NewWitness(
				newBenchChainAssignment(curveID, 128),
				curveID.ScalarField(),
			)
			require.NoError(t, err, "creating witness should succeed")
			requireSolvedWireCommitmentMatches(t, prover.genericState, ccs, pk, fullWitness)
		})
	}
}

func requireSolvedWireCommitmentMatches(
	t *testing.T,
	state *genericProverState,
	ccs any,
	pk gnarkplonk.ProvingKey,
	fullWitness witness.Witness,
) {
	t.Helper()
	require.NotNil(t, state, "generic state should be prepared")

	switch spr := ccs.(type) {
	case *csbn254.SparseR1CS:
		solution, err := spr.Solve(fullWitness)
		require.NoError(t, err, "solving BN254 witness should succeed")
		l := []bnfr.Element(solution.(*csbn254.SparseR1CSSolution).L)
		typedPK := pk.(*bnplonk.ProvingKey)
		want, err := bnkzg.Commit(l, typedPK.KzgLagrange)
		require.NoError(t, err, "BN254 lagrange commitment should succeed")
		got, err := state.commitLagrangeRaw(genericRawBN254Fr(l))
		require.NoError(t, err, "BN254 generic GPU wire commitment should succeed")
		requireBN254ProjectiveMatches(t, want, got, "BN254 solved L wire")
	case *csbls12377.SparseR1CS:
		solution, err := spr.Solve(fullWitness)
		require.NoError(t, err, "solving BLS12-377 witness should succeed")
		l := []blsfr.Element(solution.(*csbls12377.SparseR1CSSolution).L)
		typedPK := pk.(*blsplonk.ProvingKey)
		want, err := blskzg.Commit(l, typedPK.KzgLagrange)
		require.NoError(t, err, "BLS12-377 lagrange commitment should succeed")
		got, err := state.commitLagrangeRaw(genericRawBLS12377Fr(l))
		require.NoError(t, err, "BLS12-377 generic GPU wire commitment should succeed")
		requireBLS12377ProjectiveMatches(t, want, got, "BLS12-377 solved L wire")
	case *csbw6761.SparseR1CS:
		solution, err := spr.Solve(fullWitness)
		require.NoError(t, err, "solving BW6-761 witness should succeed")
		l := []bwfr.Element(solution.(*csbw6761.SparseR1CSSolution).L)
		typedPK := pk.(*bwplonk.ProvingKey)
		want, err := bwkzg.Commit(l, typedPK.KzgLagrange)
		require.NoError(t, err, "BW6-761 lagrange commitment should succeed")
		got, err := state.commitLagrangeRaw(genericRawBW6761Fr(l))
		require.NoError(t, err, "BW6-761 generic GPU wire commitment should succeed")
		requireBW6761ProjectiveMatches(t, want, got, "BW6-761 solved L wire")
	default:
		t.Fatalf("unexpected constraint system type %T", ccs)
	}
}
