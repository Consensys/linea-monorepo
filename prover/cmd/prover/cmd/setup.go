package cmd

import (
	"context"
	"crypto/sha256"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	pi_interconnection "github.com/consensys/linea-monorepo/prover/circuits/pi-interconnection"
	"github.com/consensys/linea-monorepo/prover/protocol/serde"
	"github.com/consensys/linea-monorepo/prover/utils/signal"

	blob_v1 "github.com/consensys/linea-monorepo/prover/lib/compressor/blob/v1"
	"github.com/sirupsen/logrus"

	"github.com/consensys/gnark-crypto/ecc"
	"github.com/consensys/gnark/backend/plonk"
	"github.com/consensys/linea-monorepo/prover/circuits"
	"github.com/consensys/linea-monorepo/prover/circuits/aggregation"
	daconfig "github.com/consensys/linea-monorepo/prover/circuits/dataavailability/config"
	blobdecompression "github.com/consensys/linea-monorepo/prover/circuits/dataavailability/v2"
	"github.com/consensys/linea-monorepo/prover/circuits/dummy"
	"github.com/consensys/linea-monorepo/prover/circuits/emulation"
	"github.com/consensys/linea-monorepo/prover/circuits/execution"
	"github.com/consensys/linea-monorepo/prover/config"
	"github.com/consensys/linea-monorepo/prover/utils"
	"github.com/consensys/linea-monorepo/prover/zkevm"
)

type SetupArgs struct {
	Force      bool
	Circuits   string
	AssetsDir  string
	ConfigFile string
}

var AllCircuits = []circuits.CircuitID{
	circuits.ExecutionCircuitID,
	circuits.ExecutionLargeCircuitID,
	circuits.ExecutionLimitlessCircuitID,
	circuits.DataAvailabilityV2CircuitID,
	circuits.PublicInputInterconnectionCircuitID,
	circuits.AggregationCircuitID,
	circuits.EmulationCircuitID,
	circuits.EmulationDummyCircuitID, // we want to generate Verifier.sol for this one
}

// Setup orchestrates the setup process for specified circuits, ensuring assets are generated or updated as needed.
func Setup(ctx context.Context, args SetupArgs) error {
	const cmdName = "setup"

	// This allows the user to dump stacktraces by sending a SIGUSR1 to the
	// current process.
	signal.RegisterStackTraceDumpHandler()

	// Read config from file
	cfg, err := config.NewConfigFromFile(args.ConfigFile)
	if err != nil {
		return fmt.Errorf("%s failed to read config file: %w", cmdName, err)
	}

	// Parse inCircuits
	inCircuits, err := parseCircuitInputs(args.Circuits)
	if err != nil {
		return fmt.Errorf("%s unknown circuit: %w", cmdName, err)
	}

	// Create assets dir if needed (example; efs://prover-assets/v0.1.0/)
	if err := os.MkdirAll(filepath.Join(cfg.AssetsDir, cfg.Version), 0755); err != nil {
		return fmt.Errorf("%s failed to create assets directory: %w", cmdName, err)
	}

	// srs provider
	srsProvider, err := circuits.NewSRSStore(cfg.PathForSRS())
	if err != nil {
		return fmt.Errorf("%s failed to create SRS provider: %w", cmdName, err)
	}

	// Setup non-aggregation and non-emulation circuits first
	// For each circuit, we start by compiling the circuit, and
	// then we do a SHA-sum and compare against the one in the manifest.json
	for _, c := range AllCircuits {
		if !inCircuits[c] || c == circuits.AggregationCircuitID ||
			c == circuits.EmulationCircuitID {
			// we skip aggregation/emulation circuits in this first loop since the setup is more complex
			continue
		}
		logrus.Infof("Setting up circuit %s", c)

		// Build the circuit
		builder, extraFlags, err := createCircuitBuilder(c, cfg, args)
		if err != nil {
			return fmt.Errorf("%s failed to create builder for circuit %s: %w", cmdName, c, err)
		}

		if err := updateSetup(ctx, cfg, args.Force, srsProvider, c, builder, extraFlags); err != nil {
			return err
		}

		if c == circuits.ExecutionCircuitID || c == circuits.ExecutionLargeCircuitID {
			if err := serializeInnerCircuit(cfg, c); err != nil {
				return err
			}
		}
	}

	// Early exit if no aggregation or emulation circuits
	if !inCircuits[circuits.AggregationCircuitID] && !inCircuits[circuits.EmulationCircuitID] {
		// we are done
		return nil
	}

	// Get verifying key for public-input circuit
	piSetup, err := circuits.LoadSetup(cfg, circuits.PublicInputInterconnectionCircuitID)
	if err != nil {
		return fmt.Errorf("%s failed to load public input interconnection setup: %w", cmdName, err)
	}

	// Collect verifying keys for ALL circuits (using global circuit ID mapping)
	// The IsAllowedCircuitID bitmask in the config determines which ones are actually allowed at runtime
	allVks, err := collectAllVerifyingKeys(ctx, cfg, srsProvider)
	if err != nil {
		return err
	}

	// Setup aggregation circuits
	allowedVkForEmulation, err := setupAggregationCircuits(ctx, cfg, args.Force, srsProvider, inCircuits, &piSetup, allVks)
	if err != nil {
		return err
	}

	// Setup emulation circuit if needed
	if inCircuits[circuits.EmulationCircuitID] {
		logrus.Infof("setting up %s", circuits.EmulationCircuitID)
		builder := emulation.NewBuilder(allowedVkForEmulation)
		return updateSetup(ctx, cfg, args.Force, srsProvider, circuits.EmulationCircuitID, builder, nil)
	}

	logrus.Infof("Done setting up circuits and writing the assets to disk :)")

	return nil
}

// updateSetup: Runs the setup for the given circuit if needed.
// It first compiles the circuit, then checks if the files already exist, and if so, if the checksums match.
// If the files already exist and the checksums match, it skips the setup.
// else it does the setup and writes the assets to disk.
func updateSetup(ctx context.Context, cfg *config.Config, force bool,
	srsProvider circuits.SRSProvider, circuit circuits.CircuitID,
	builder circuits.Builder, extraFlags map[string]any,
) error {
	if extraFlags == nil {
		extraFlags = make(map[string]any)
	}

	// compile the circuit
	logrus.Infof("Compiling circuit %s", circuit)
	ccs, err := builder.Compile()
	if err != nil {
		return fmt.Errorf("failed to compile circuit %s: %w", circuit, err)
	}

	// Derive the asset paths
	setupPath := cfg.PathForSetup(string(circuit))
	manifestPath := filepath.Join(setupPath, config.ManifestFileName)

	if !force {
		// we may want to skip setup if the files already exist
		// and the checksums match
		// read manifest if already exists
		if manifest, err := circuits.ReadSetupManifest(manifestPath); err == nil {
			circuitDigest, err := circuits.CircuitDigest(ccs)
			if err != nil {
				return fmt.Errorf("failed to compute circuit digest for circuit %s: %w", circuit, err)
			}

			if manifest.Checksums.Circuit == circuitDigest {
				logrus.Infof("skipping %s (already setup)", circuit)
				return nil
			}
		}
	}

	// run the actual setup
	logrus.Infof("plonk setup for %s", circuit)
	setup, err := circuits.MakeSetup(ctx, circuit, ccs, srsProvider, extraFlags)
	if err != nil {
		return fmt.Errorf("failed to setup circuit %s: %w", circuit, err)
	}

	logrus.Infof("writing assets for %s", circuit)
	return setup.WriteTo(setupPath)
}

// serializeInnerCircuit serializes the compiled inner circuit (wizard IOP) to disk
// if the config has serialization enabled.
func serializeInnerCircuit(cfg *config.Config, c circuits.CircuitID) error {
	hasSer, innerCircuitPath, compressed := cfg.ExecutionCircuitBin(string(c))
	if !hasSer {
		return nil
	}

	limits := cfg.TracesLimits
	if c == circuits.ExecutionLargeCircuitID {
		limits.SetLargeMode()
	}
	// FullZkEvm is memoized via sync.Once — returns the cached compiled instance
	zkEvm := zkevm.FullZkEvm(&limits, cfg)

	logrus.Infof("Serializing inner circuit for %s to %s (compressed=%v)", c, innerCircuitPath, compressed)
	if err := serde.StoreToDisk(innerCircuitPath, zkEvm, compressed); err != nil {
		return fmt.Errorf("failed to serialize inner circuit for %s: %w", c, err)
	}
	return nil
}

// parseCircuitInputs: Converts the comma-separated circuit string into a map of enabled circuits.
func parseCircuitInputs(circuitsStr string) (map[circuits.CircuitID]bool, error) {
	inCircuits := make(map[circuits.CircuitID]bool)
	for _, c := range AllCircuits {
		inCircuits[c] = false
	}
	for _, c := range strings.Split(circuitsStr, ",") {
		circuitID := circuits.CircuitID(c)
		if _, ok := inCircuits[circuitID]; !ok {
			return nil, fmt.Errorf("invalid circuit: %s", c)
		}
		inCircuits[circuitID] = true
	}
	return inCircuits, nil
}

// createCircuitBuilder: Constructs the appropriate circuit builder and extra flags based on the circuit ID.
func createCircuitBuilder(c circuits.CircuitID, cfg *config.Config, args SetupArgs,
) (circuits.Builder, map[string]any, error) {
	extraFlags := make(map[string]any)
	switch c {
	case circuits.ExecutionCircuitID, circuits.ExecutionLargeCircuitID:
		limits := cfg.TracesLimits
		if c == circuits.ExecutionLargeCircuitID {
			limits.SetLargeMode()
		}
		extraFlags["cfg_checksum"] = limits.Checksum()
		zkEvm := zkevm.FullZkEvm(&limits, cfg)
		return execution.NewBuilder(zkEvm), extraFlags, nil

	case circuits.ExecutionLimitlessCircuitID:

		panic("uncomment when the limitless prover works")

		// executionLimitlessPath := cfg.PathForSetup("execution-limitless")
		// limits := cfg.TracesLimits
		// extraFlags["cfg_checksum"] = limits.Checksum()

		// logrus.Info("Setting up limitless prover assets")
		// asset := zkevm.NewLimitlessZkEVM(cfg)

		// // Unlike for the other circuits, the limitless prover assets are written
		// // to disk directly before returning the circuit builder. The reason is
		// // that the limitless prover assets are large and we want to avoid keeping
		// // them in memory. The second reason is that returning them alongside the
		// // build would change the structure of the function for just one case.
		// logrus.Infof("Writing limitless prover assets to path: %s", executionLimitlessPath)
		// if err := asset.Store(cfg); err != nil {
		// 	return nil, nil, fmt.Errorf("failed to write limitless prover assets: %w", err)
		// }
		// compCong := asset.DistWizard.CompiledConglomeration
		// asset = nil
		// runtime.GC()

		// return execution.NewBuilderLimitless(compCong.Wiop, &limits), extraFlags, nil

	case circuits.DataAvailabilityV2CircuitID:
		extraFlags["maxUsableBytes"] = blob_v1.MaxUsableBytes
		extraFlags["maxUncompressedBytes"] = cfg.DataAvailability.MaxUncompressedNbBytes
		extraFlags["dictNbBytes"] = cfg.DataAvailability.DictNbBytes
		extraFlags["maxNbBatches"] = cfg.DataAvailability.MaxNbBatches
		return blobdecompression.NewBuilder(daconfig.FromGlobalConfig(cfg.DataAvailability)), extraFlags, nil

	case circuits.PublicInputInterconnectionCircuitID:
		return pi_interconnection.NewBuilder(cfg.PublicInputInterconnection), extraFlags, nil

	case circuits.EmulationDummyCircuitID:
		// we can get the Verifier.sol from there.
		return dummy.NewBuilder(circuits.MockCircuitIDEmulation, ecc.BN254.ScalarField()), extraFlags, nil

	default:
		return nil, nil, fmt.Errorf("unsupported circuit: %s", c)
	}
}

// collectAllVerifyingKeys: Gathers verifying keys for ALL circuits in circuits.GlobalCircuitIDMapping.
// The aggregation circuit always has access to all VKs. The IsAllowedCircuitID bitmask
// in the config controls which circuits are actually allowed at runtime.
func collectAllVerifyingKeys(ctx context.Context, cfg *config.Config, srsProvider circuits.SRSProvider) ([]plonk.VerifyingKey, error) {
	allCircuitNames := circuits.GetAllCircuitNames()
	allVks := make([]plonk.VerifyingKey, len(allCircuitNames))

	for i, circuitName := range allCircuitNames {
		if isDummyCircuit(circuitName) {
			curveID, mockID, err := getDummyCircuitParams(circuitName)
			if err != nil {
				return nil, err
			}
			vk, err := getDummyCircuitVK(ctx, srsProvider, circuits.CircuitID(circuitName), dummy.NewBuilder(mockID, curveID.ScalarField()))
			if err != nil {
				return nil, err
			}
			allVks[i] = vk
			continue
		}

		// derive the asset paths
		setupPath := cfg.PathForSetup(circuitName)
		vkPath := filepath.Join(setupPath, config.VerifyingKeyFileName)
		vk := plonk.NewVerifyingKey(ecc.BLS12_377)
		if err := circuits.ReadVerifyingKey(vkPath, vk); err != nil {
			return nil, fmt.Errorf("failed to read verifying key for circuit %s: %w", circuitName, err)
		}
		allVks[i] = vk
	}
	return allVks, nil
}

// getDummyCircuitParams returns the curve and mock ID for a dummy circuit.
func getDummyCircuitParams(cID string) (ecc.ID, circuits.MockCircuitID, error) {
	switch circuits.CircuitID(cID) {
	case circuits.ExecutionDummyCircuitID:
		return ecc.BLS12_377, circuits.MockCircuitIDExecution, nil
	case circuits.DataAvailabilityDummyCircuitID:
		return ecc.BLS12_377, circuits.MockCircuitIDDecompression, nil
	case circuits.EmulationDummyCircuitID:
		return ecc.BN254, circuits.MockCircuitIDEmulation, nil
	default:
		return 0, 0, fmt.Errorf("unknown dummy circuit: %s", cID)
	}
}

// setupAggregationCircuits: Configures aggregation circuits and collects their verifying keys for emulation.
func setupAggregationCircuits(ctx context.Context, cfg *config.Config, force bool,
	srsProvider circuits.SRSProvider, inCircuits map[circuits.CircuitID]bool,
	piSetup *circuits.Setup, allowedVkForAggregation []plonk.VerifyingKey,
) ([]plonk.VerifyingKey, error) {
	if !inCircuits[circuits.AggregationCircuitID] {
		return nil, nil
	}

	// we need to compute the digest of the verifying keys & store them in the manifest
	// for the aggregation circuits to be able to check compatibility at run time with the proofs
	extraFlags := map[string]any{
		"allowedVkForAggregationDigests": listOfChecksums(allowedVkForAggregation),
	}

	allowedVkForEmulation := make([]plonk.VerifyingKey, 0, len(cfg.Aggregation.NumProofs))
	for _, numProofs := range cfg.Aggregation.NumProofs {
		c := circuits.CircuitID(fmt.Sprintf("%s-%d", string(circuits.AggregationCircuitID), numProofs))
		logrus.Infof("setting up %s (numProofs=%d)", c, numProofs)

		// Always pass ALL verifying keys - the IsAllowedCircuitID bitmask controls which are actually allowed
		builder := aggregation.NewBuilder(numProofs, *piSetup, allowedVkForAggregation)
		if err := updateSetup(ctx, cfg, force, srsProvider, c, builder, extraFlags); err != nil {
			return nil, err
		}

		// read the verifying key
		setupPath := cfg.PathForSetup(string(c))
		vkPath := filepath.Join(setupPath, config.VerifyingKeyFileName)
		vk := plonk.NewVerifyingKey(ecc.BW6_761)
		if err := circuits.ReadVerifyingKey(vkPath, vk); err != nil {
			return nil, fmt.Errorf("failed to read verifying key for circuit %s: %w", c, err)
		}
		allowedVkForEmulation = append(allowedVkForEmulation, vk)
	}
	return allowedVkForEmulation, nil
}

func isDummyCircuit(cID string) bool {
	switch circuits.CircuitID(cID) {
	case circuits.ExecutionDummyCircuitID, circuits.DataAvailabilityDummyCircuitID, circuits.EmulationDummyCircuitID:
		return true
	}
	return false

}

func getDummyCircuitVK(ctx context.Context, srsProvider circuits.SRSProvider, circuit circuits.CircuitID, builder circuits.Builder) (plonk.VerifyingKey, error) {
	// compile the circuit
	logrus.Infof("compiling %s", circuit)
	ccs, err := builder.Compile()
	if err != nil {
		return nil, fmt.Errorf("failed to compile circuit %s: %w", circuit, err)
	}
	setup, err := circuits.MakeSetup(ctx, circuit, ccs, srsProvider, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to setup circuit %s: %w", circuit, err)
	}

	return setup.VerifyingKey, nil
}

// listOfChecksums Computes a list of SHA256 checksums for a list of assets, the result is given
// in hexstring.
func listOfChecksums[T io.WriterTo](assets []T) []string {
	res := make([]string, len(assets))
	h := sha256.New()
	for i := range assets {
		h.Reset()
		_, err := assets[i].WriteTo(h)
		if err != nil {
			// It is unexpected that writing in a hasher could possibly fail.
			panic(err)
		}
		digest := h.Sum(nil)
		res[i] = utils.HexEncodeToString(digest)
	}
	return res
}
