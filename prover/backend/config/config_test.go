package config_test

import (
	"testing"

	"github.com/consensys/accelerated-crypto-monorepo/backend/config"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func TestEthConfig(t *testing.T) {

	// Normally the test does not have CHAIN_ID env variable
	// so it will return an error.
	ethConf, err := config.GetLayer2()
	assert.Error(t, err)
	assert.Nil(t, ethConf)

	// Set the mandatory CHAIN_ID so that the
	t.Setenv("LAYER2_CHAIN_ID", "1")
	t.Setenv("LAYER2_MESSAGE_SERVICE_CONTRACT", "0xaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa")

	// Now, this should return the right result
	ethConf, err = config.GetLayer2()
	assert.NoError(t, err)
	assert.NotNil(t, ethConf)

	// Test that we have the right initial values
	assert.Equal(t, false, ethConf.ChainIdOverride)

	// Now set all the other values
	t.Setenv("LAYER2_CHAIN_ID_OVERRIDE", "true")

	// Reload the conf to get the new configuration
	ethConf = config.MustGetLayer2()

	// Test that we have the right initial values
	assert.Equal(t, 1, ethConf.ChainId)
	assert.Equal(t, true, ethConf.ChainIdOverride)
}

func TestProverConfig(t *testing.T) {

	// Normally the test does not have CHAIN_ID env variable
	// so it will return an error.
	proConf, err := config.GetProver()
	assert.Error(t, err)
	assert.Nil(t, proConf)

	var (
		testPkeyFile           = "./pkey-file"
		testVkeyFile           = "./vkey-file"
		testR1csFile           = "./r1cs-file"
		testVersion            = "coordinator-before-move-PoA-mvp-58-ge76c5996"
		testConflatedTracesDir = "/shared/traces/conflated-traces"
	)

	// Set the required fields
	t.Setenv("PROVER_PKEY_FILE", testPkeyFile)
	t.Setenv("PROVER_VKEY_FILE", testVkeyFile)
	t.Setenv("PROVER_R1CS_FILE", testR1csFile)
	t.Setenv("PROVER_VERSION", testVersion)
	t.Setenv("PROVER_CONFLATED_TRACES_DIR", testConflatedTracesDir)

	// Now, this should return the right result
	proConf, err = config.GetProver()
	assert.NoError(t, err)
	assert.NotNil(t, proConf)

	// Check the non-required fields
	assert.Equal(t, false, proConf.ProfilingEnabled)
	assert.Equal(t, false, proConf.TracingEnabled)
	assert.Equal(t, false, proConf.SkipTraces)
	assert.Equal(t, true, proConf.DevLightVersion)
	assert.Equal(t, testVersion, proConf.Version)
	assert.Equal(t, testConflatedTracesDir, proConf.ConflatedTracesDir)

	// Set the non required field
	t.Setenv("PROVER_PROFILING_ENABLED", "true")
	t.Setenv("PROVER_TRACING_ENABLED", "true")
	t.Setenv("PROVER_DEV_LIGHT_VERSION", "false")
	t.Setenv("PROVER_VERSION", "pewpewpew")
	t.Setenv("PROVER_SKIP_TRACES", "true")
	t.Setenv("PROVER_WITH_ECDSA", "true")
	t.Setenv("PROVER_WITH_KECCAK", "true")
	t.Setenv("PROVER_WITH_STATE_MANAGER", "true")

	proConf = config.MustGetProver()

	// Check the non-required fields
	assert.Equal(t, testPkeyFile, proConf.PKeyFile)
	assert.Equal(t, testR1csFile, proConf.R1CSFile)
	assert.Equal(t, true, proConf.ProfilingEnabled)
	assert.Equal(t, true, proConf.TracingEnabled)
	assert.Equal(t, false, proConf.DevLightVersion)
	assert.Equal(t, "pewpewpew", proConf.Version)
	assert.Equal(t, true, proConf.SkipTraces)
	assert.Equal(t, true, proConf.WithEcdsa)
	assert.Equal(t, true, proConf.WithKeccak)
	assert.Equal(t, true, proConf.WithStateManager)
}

func TestLogging(t *testing.T) {

	config.SetenvForTest(t)

	config.InitApp()

	// Default log level shoud be info
	assert.Equal(t, logrus.StandardLogger().Level, logrus.InfoLevel)

	// Try setting it to trace
	t.Setenv("LOG_LEVEL", "trace")
	config.InitApp()
	assert.Equal(t, logrus.StandardLogger().Level, logrus.TraceLevel)

	t.Setenv("LOG_LEVEL", "debug")
	config.InitApp()
	assert.Equal(t, logrus.StandardLogger().Level, logrus.DebugLevel)

	t.Setenv("LOG_LEVEL", "info")
	config.InitApp()
	assert.Equal(t, logrus.StandardLogger().Level, logrus.InfoLevel)

	t.Setenv("LOG_LEVEL", "warn")
	config.InitApp()
	assert.Equal(t, logrus.StandardLogger().Level, logrus.WarnLevel)

	t.Setenv("LOG_LEVEL", "error")
	config.InitApp()
	assert.Equal(t, logrus.StandardLogger().Level, logrus.ErrorLevel)
}
