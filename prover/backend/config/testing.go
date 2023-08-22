package config

import (
	"path"
	"testing"
)

const (
	setupDir string = "/tmp/prover-tests/setup-data"
)

// Set up the environment variables for the test. This panics if called
// outside of a testing environment. Thus behaviour avoids the
func SetenvForTest(t *testing.T) {
	// Initialize the config
	t.Setenv("LAYER2_CHAIN_ID", "59140")
	t.Setenv("LAYER2_MESSAGE_SERVICE_CONTRACT", "0x0000000000000000000000000000000000000000")
	t.Setenv("PROVER_CONFLATED_TRACES_DIR", "./not/a/real/path")
	t.Setenv("PROVER_PKEY_FILE", path.Join(setupDir, "proving-key"))
	t.Setenv("PROVER_VKEY_FILE", path.Join(setupDir, "verifying-key"))
	t.Setenv("PROVER_R1CS_FILE", path.Join(setupDir, "r1cs"))
	t.Setenv("PROVER_SOL_VERIFIER", path.Join(setupDir, "verifierContract.sol"))
	t.Setenv("PROVER_MAX_CALLDATA_SIZE", "35000")
	t.Setenv("PROVER_DEV_LIGHT_VERSION", "true")
	t.Setenv("PROVER_VERSION", "not-a-real-version")
}
