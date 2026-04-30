package plonk2

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/consensys/gnark-crypto/ecc"
	gnarkplonk "github.com/consensys/gnark/backend/plonk"
	"github.com/consensys/gnark/backend/witness"
	"github.com/consensys/gnark/constraint"
	"github.com/consensys/gnark/frontend"
	"github.com/consensys/gnark/frontend/cs/scs"
	"github.com/consensys/gnark/test/unsafekzg"
	"github.com/stretchr/testify/require"
)

type proverAPIFixture struct {
	curveID       ecc.ID
	ccs           constraint.ConstraintSystem
	pk            gnarkplonk.ProvingKey
	vk            gnarkplonk.VerifyingKey
	validWitness  frontend.Circuit
	publicWitness frontend.Circuit
}

func TestProver_CPUFallbackSupportedCurves(t *testing.T) {
	for _, curveID := range []ecc.ID{ecc.BN254, ecc.BLS12_377, ecc.BW6_761} {
		t.Run(curveID.String(), func(t *testing.T) {
			fixture := newProverAPIFixture(t, curveID)
			fullWitness := newMulWitness(t, curveID, 3, 11, 33)
			publicWitness, err := fullWitness.Public()
			require.NoError(t, err, "extracting public witness should succeed")

			prover, err := NewProver(nil, fixture.ccs, fixture.pk)
			require.NoError(t, err, "creating prover should accept target curve")
			defer prover.Close()

			proof, err := prover.Prove(fullWitness)
			require.NoError(t, err, "CPU fallback should produce a proof")
			require.NoError(
				t,
				gnarkplonk.Verify(proof, fixture.vk, publicWitness),
				"CPU fallback proof should verify",
			)
		})
	}
}

func TestProver_DisabledCPUFallback(t *testing.T) {
	fixture := newProverAPIFixture(t, ecc.BN254)
	fullWitness := newMulWitness(t, ecc.BN254, 3, 11, 33)

	prover, err := NewProver(
		nil,
		fixture.ccs,
		fixture.pk,
		WithCPUFallback(false),
	)
	require.NoError(t, err, "creating prover should succeed")

	_, err = prover.Prove(fullWitness)
	require.ErrorIs(t, err, errGPUProverNotWired, "disabled fallback should not prove on CPU")
}

func TestProver_StrictModeRejectsFallback(t *testing.T) {
	fixture := newProverAPIFixture(t, ecc.BN254)
	fullWitness := newMulWitness(t, ecc.BN254, 3, 11, 33)

	prover, err := NewProver(nil, fixture.ccs, fixture.pk, WithStrictMode(true))
	require.NoError(t, err, "creating strict prover should succeed")

	_, err = prover.Prove(fullWitness)
	require.ErrorIs(t, err, errGPUProverNotWired, "strict mode should reject CPU fallback")
}

func TestProver_EnabledGPUFallsBackWhenAllowed(t *testing.T) {
	fixture := newProverAPIFixture(t, ecc.BN254)
	fullWitness := newMulWitness(t, ecc.BN254, 3, 11, 33)

	prover, err := NewProver(nil, fixture.ccs, fixture.pk, WithEnabled(true))
	require.NoError(t, err, "creating enabled prover should succeed")

	_, err = prover.Prove(fullWitness)
	require.NoError(t, err, "enabled unwired GPU prover should use CPU fallback when allowed")
}

func TestProver_MismatchedKeyAndConstraintCurves(t *testing.T) {
	bn254Fixture := newProverAPIFixture(t, ecc.BN254)
	blsFixture := newProverAPIFixture(t, ecc.BLS12_377)

	_, err := NewProver(nil, bn254Fixture.ccs, blsFixture.pk)
	require.Error(t, err, "mismatched curves should fail at construction")
	require.Contains(t, err.Error(), "does not match", "error should explain curve mismatch")
}

func TestProver_UnsupportedCurve(t *testing.T) {
	fixture := newProverAPIFixture(t, ecc.BLS12_381)

	_, err := NewProver(nil, fixture.ccs, fixture.pk)
	require.Error(t, err, "unsupported curves should fail at construction")
	require.Contains(t, err.Error(), "unsupported", "error should explain unsupported input")
}

func TestProver_CloseIsIdempotent(t *testing.T) {
	fixture := newProverAPIFixture(t, ecc.BN254)

	prover, err := NewProver(nil, fixture.ccs, fixture.pk)
	require.NoError(t, err, "creating prover should succeed")
	require.NoError(t, prover.Close(), "first close should succeed")
	require.NoError(t, prover.Close(), "second close should succeed")
}

func TestProvePackageFunction(t *testing.T) {
	fixture := newProverAPIFixture(t, ecc.BN254)
	fullWitness := newMulWitness(t, ecc.BN254, 3, 11, 33)
	publicWitness, err := fullWitness.Public()
	require.NoError(t, err, "extracting public witness should succeed")

	proof, err := Prove(nil, fixture.ccs, fixture.pk, fullWitness)
	require.NoError(t, err, "package-level Prove should use CPU fallback")
	require.NoError(
		t,
		gnarkplonk.Verify(proof, fixture.vk, publicWitness),
		"package-level proof should verify",
	)
}

func TestProver_CPUFallbackInvalidWitnessTargetCurves(t *testing.T) {
	for _, curveID := range []ecc.ID{ecc.BN254, ecc.BLS12_377, ecc.BW6_761} {
		t.Run(curveID.String(), func(t *testing.T) {
			fixture := newProverAPIFixture(t, curveID)
			invalidWitness := newMulWitness(t, curveID, 3, 11, 34)

			prover, err := NewProver(nil, fixture.ccs, fixture.pk)
			require.NoError(t, err, "creating prover should succeed")
			defer prover.Close()

			_, err = prover.Prove(invalidWitness)
			require.Error(t, err, "invalid witness should fail through the plonk2 prover API")
		})
	}
}

func TestProverTrace(t *testing.T) {
	fixture := newProverAPIFixture(t, ecc.BN254)
	fullWitness := newMulWitness(t, ecc.BN254, 3, 11, 33)
	tracePath := filepath.Join(t.TempDir(), "plonk2-trace.jsonl")

	prover, err := NewProver(nil, fixture.ccs, fixture.pk, WithTrace(tracePath))
	require.NoError(t, err, "creating traced prover should succeed")
	_, err = prover.Prove(fullWitness)
	require.NoError(t, err, "traced fallback proof should succeed")

	contents, err := os.ReadFile(tracePath)
	require.NoError(t, err, "trace file should be written")
	require.Contains(t, string(contents), `"phase":"prepare"`, "trace should include prepare event")
	require.Contains(t, string(contents), `"phase":"cpu_fallback"`, "trace should include CPU fallback event")

	var event map[string]any
	require.NoError(t, json.Unmarshal(contents[:jsonLineEnd(contents)], &event), "first trace event should be JSON")
	require.Equal(t, "plonk2_prover", event["event"], "trace event should identify plonk2 prover")
	require.Contains(t, event, "peak_bytes", "trace should include memory metadata")
}

func jsonLineEnd(contents []byte) int {
	for i, b := range contents {
		if b == '\n' {
			return i
		}
	}
	return len(contents)
}

func newProverAPIFixture(t *testing.T, curveID ecc.ID) proverAPIFixture {
	t.Helper()

	ccs, err := frontend.Compile(
		curveID.ScalarField(),
		scs.NewBuilder,
		&e2eMulCircuit{},
	)
	require.NoError(t, err, "compiling circuit should succeed")

	srs, srsLagrange, err := unsafekzg.NewSRS(ccs)
	require.NoError(t, err, "creating unsafe test SRS should succeed")

	pk, vk, err := gnarkplonk.Setup(ccs, srs, srsLagrange)
	require.NoError(t, err, "PlonK setup should succeed")

	return proverAPIFixture{
		curveID: curveID,
		ccs:     ccs,
		pk:      pk,
		vk:      vk,
	}
}

func newMulWitness(t *testing.T, curveID ecc.ID, x, y, z int) witness.Witness {
	t.Helper()

	witness, err := frontend.NewWitness(
		&e2eMulCircuit{X: x, Y: y, Z: z},
		curveID.ScalarField(),
	)
	require.NoError(t, err, "creating witness should succeed")
	return witness
}
