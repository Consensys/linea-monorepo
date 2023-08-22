package testing

import (
	"fmt"
	"os"
	"path"
	"strings"
	"testing"

	"github.com/consensys/accelerated-crypto-monorepo/backend"
	"github.com/consensys/accelerated-crypto-monorepo/backend/config"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
)

const SAMPLES_DIR string = "./"
const SHOULD_FAIL string = "shouldfail"

func TestSamples(t *testing.T) {

	files, err := os.ReadDir(SAMPLES_DIR)
	if err != nil {
		panic(err)
	}

	for _, file := range files {
		// We only walk down the dirs
		if !file.IsDir() {
			continue
		}

		// Ignore the profiling folder is any

		sampleName := file.Name()
		sampleDir := path.Join(SAMPLES_DIR, sampleName)

		// Recursively read the directory
		subDir, err := os.ReadDir(sampleDir)
		if err != nil {
			panic(err)
		}

		var proverInput string
		var shouldFail bool
		var hasTraces bool
		var noOutput bool

		for _, f := range subDir {
			// Detect if the test should fail
			if strings.Contains(f.Name(), "shouldfail") {
				shouldFail = true
			}

			// Detects the input file
			if strings.Contains(f.Name(), "getZkProof.json") {
				proverInput = f.Name()
			}

			// Detect if the test contains traces
			if strings.HasSuffix(f.Name(), ".gz") {
				hasTraces = true
			}

			// Detect if the test should ouput
			if strings.Contains(f.Name(), "no") && strings.Contains(f.Name(), "output") {
				noOutput = true
			}
		}

		// Skip the folder because it does not contain interesting files
		if len(proverInput) == 0 {
			t.Logf("Skipping `%v`  because it does not contain prover inputs", sampleName)
			continue
		}

		t.Run(sampleName, func(t *testing.T) {
			runTestSample(t, sampleDir, proverInput, hasTraces, shouldFail, noOutput)
		})
	}
}

func runTestSample(t *testing.T, sampleDir, fname string, hasTraces, shouldFail, noOutput bool) {

	// Set the level to trace
	logrus.SetLevel(logrus.TraceLevel)

	output := path.Join(sampleDir, fmt.Sprintf("%v-%v-zkProof.json", sampleDir, sampleDir))
	input := path.Join(sampleDir, fname)

	if noOutput {
		output = "/dev/null"
	}

	// Load the default config
	config.SetenvForTest(t)

	// Skip the traces by default
	t.Setenv("PROVER_SKIP_TRACES", "true")
	t.Setenv("LAYER2_MESSAGE_SERVICE_CONTRACT", "0xc499a572640b64ea1c8c194c43bc3e19940719dc")

	// Even in notrace mode, we need a proving key
	t.Setenv("PROVER_PKEY_FILE", "../../docker/setup/light/proving_key.bin")
	t.Setenv("PROVER_VKEY_FILE", "../../docker/setup/light/verifying_key.bin")
	t.Setenv("PROVER_R1CS_FILE", "../../docker/setup/light/circuit.bin")

	// And if we think, we can we activate the trace verification
	if WITH_CORSET && hasTraces {
		t.Setenv("PROVER_SKIP_TRACES", "false")
		t.Setenv("PROVER_CONFLATED_TRACES_DIR", sampleDir)
	}

	// And run the test
	if shouldFail {
		require.Panics(t, func() {
			backend.RunCorsetAndProver(input, output)
		})
	} else {
		backend.RunCorsetAndProver(input, output)
	}
}
