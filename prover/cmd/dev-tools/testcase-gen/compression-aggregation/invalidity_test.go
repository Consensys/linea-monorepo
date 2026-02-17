package main

import (
	"context"
	"math/big"
	"math/rand"
	"path/filepath"
	"testing"

	"github.com/consensys/linea-monorepo/prover/circuits"
	circuitInvalidity "github.com/consensys/linea-monorepo/prover/circuits/invalidity"
	keccak "github.com/consensys/linea-monorepo/prover/circuits/pi-interconnection/keccak/prover/circuits/pi-interconnection/keccak"
	smt_koalabear "github.com/consensys/linea-monorepo/prover/crypto/state-management/smt_koalabear"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/require"
)

func TestInvalidity(t *testing.T) {

	// Create a reproducible RNG
	// #nosec G404 --we don't need a cryptographic RNG for testing purpose
	rng := rand.New(rand.NewSource(seed))

	configFile = "../../../../config/config-integration-development.toml"
	viper.Set("assets_dir", "../../../../prover-assets")

	spec := &InvalidityProofSpec{
		ChainID:             big.NewInt(51),
		ExpectedBlockHeight: 1_000_000_000,
		FtxNumber:           1678,
	}

	ProcessInvaliditySpec(rng, spec, nil, "spec-invalidity.json")
}

func TestInvalidityFull(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping full-mode invalidity test in short mode")
	}

	const maxRlpByteSize = 4096

	// Step 1: Compile the invalidity-nonce-balance circuit (same as setup.go)
	t.Log("Compiling invalidity-nonce-balance circuit...")
	keccakComp := circuitInvalidity.MakeKeccakCompiledIOP(maxRlpByteSize, keccak.WizardCompilationParameters()...)
	builder := circuitInvalidity.NewBuilder(circuitInvalidity.Config{
		Depth:             smt_koalabear.DefaultDepth,
		KeccakCompiledIOP: keccakComp,
		MaxRlpByteSize:    maxRlpByteSize,
	})
	ccs, err := builder.Compile()
	require.NoError(t, err)
	t.Logf("Circuit compiled: %d constraints", ccs.GetNbConstraints())

	// Step 2: Generate the setup using unsafe (test-only) SRS
	t.Log("Generating setup with unsafe SRS...")
	srsProvider := circuits.NewUnsafeSRSProvider()
	setup, err := circuits.MakeSetup(
		context.Background(),
		circuits.InvalidityNonceBalanceCircuitID,
		ccs,
		srsProvider,
		map[string]any{},
	)
	require.NoError(t, err)

	// Step 3: Write setup assets to temp dir with the expected path structure
	// The config will resolve: <assets_dir>/<version>/<environment>/<circuit_id>/
	// integration-development.toml has version="4.0.0", environment="integration-development"
	assetsDir := t.TempDir()
	setupPath := filepath.Join(assetsDir, "4.0.0", "integration-development", string(circuits.InvalidityNonceBalanceCircuitID))
	t.Logf("Writing setup to %s", setupPath)
	require.NoError(t, setup.WriteTo(setupPath))

	// Step 4: Override viper settings for full mode
	configFile = "../../../../config/config-integration-development.toml"
	viper.Set("assets_dir", assetsDir)
	viper.Set("invalidity.prover_mode", "full")

	// Step 5: Run the test
	rng := rand.New(rand.NewSource(seed))
	spec := &InvalidityProofSpec{
		ChainID:             big.NewInt(51),
		ExpectedBlockHeight: 1_000_000_000,
		FtxNumber:           1678,
	}

	t.Log("Running ProcessInvaliditySpec in full mode...")
	ProcessInvaliditySpec(rng, spec, nil, "spec-invalidity.json")
	t.Log("Full mode invalidity test passed!")
}
