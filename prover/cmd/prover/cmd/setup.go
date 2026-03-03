package cmd

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/consensys/linea-monorepo/prover/utils/signal"

	"github.com/sirupsen/logrus"

	"github.com/consensys/linea-monorepo/prover/circuits"
	"github.com/consensys/linea-monorepo/prover/circuits/finalwrap"
	"github.com/consensys/linea-monorepo/prover/config"
	"github.com/consensys/linea-monorepo/prover/tree"
	"github.com/consensys/linea-monorepo/prover/utils/gnarkutil"
	"github.com/consensys/linea-monorepo/prover/zkevm"
)

type SetupArgs struct {
	Force      bool
	Circuits   string
	AssetsDir  string
	ConfigFile string
}

// AllCircuits lists the circuits available for setup in the tree aggregation pipeline.
// The old BLS12-377 PLONK execution, BW6-761 aggregation, and BN254 emulation
// circuits have been removed — they are replaced by tree-aggregation + final-wrap.
var AllCircuits = []circuits.CircuitID{
	circuits.TreeAggregationCircuitID,
	circuits.FinalWrapCircuitID,
}

// Setup orchestrates the setup process for the tree aggregation pipeline.
//
// In the new pipeline, the execution prover uses the wizard IOP directly
// (compiled via FullZkEvm) and does not need PLONK keys. The only PLONK
// setup needed is for the BN254 final wrap circuit.
//
// The setup flow is:
//  1. Compile execution wizard IOP (via FullZkEvm)
//  2. Compile tree aggregation levels (in-memory, recompiled at proving time)
//  3. Compile and run PLONK setup for the BN254 final wrap circuit
func Setup(ctx context.Context, args SetupArgs) error {
	const cmdName = "setup"

	// Register GKR gates and hints before any circuit compilation.
	// This is needed for Poseidon2 KoalaBear gates used in the Vortex verifier.
	gnarkutil.RegisterHintsAndGkrGates()

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

	// srs provider (needed for BN254 final wrap PLONK setup)
	srsProvider, err := circuits.NewSRSStore(cfg.PathForSRS())
	if err != nil {
		return fmt.Errorf("%s failed to create SRS provider: %w", cmdName, err)
	}

	// Setup tree aggregation and final wrap circuits
	if inCircuits[circuits.TreeAggregationCircuitID] || inCircuits[circuits.FinalWrapCircuitID] {
		if err := setupTreeAggregationCircuits(ctx, cfg, args.Force, srsProvider, inCircuits); err != nil {
			return err
		}
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

// setupTreeAggregationCircuits handles setup for the tree aggregation and final
// wrap circuits. These circuits depend on the execution circuit's CompiledIOP,
// so they require special handling.
//
// The tree aggregation levels are compiled in-memory but not persisted to disk —
// they are recompiled on-the-fly at proving time (deterministic from config).
// Only the BN254 final wrap circuit needs a PLONK setup saved to disk.
func setupTreeAggregationCircuits(ctx context.Context, cfg *config.Config, force bool,
	srsProvider circuits.SRSProvider, inCircuits map[circuits.CircuitID]bool,
) error {

	// Both tree aggregation and final wrap need the execution zkEVM's CompiledIOP
	limits := cfg.TracesLimits
	logrus.Info("Compiling execution zkEVM for tree aggregation setup...")
	fullZkEvm := zkevm.FullZkEvm(&limits, cfg)
	leafComp := fullZkEvm.LeafCompiledIOP()

	if inCircuits[circuits.TreeAggregationCircuitID] {
		logrus.Infof("Setting up %s", circuits.TreeAggregationCircuitID)

		maxDepth := cfg.TreeAggregation.MaxDepth
		if maxDepth < 1 {
			maxDepth = 1
		}

		logrus.Infof("Compiling tree aggregation with depth=%d", maxDepth)
		treeAgg := tree.CompileTreeAggregation(leafComp, maxDepth)
		logrus.Infof("Tree aggregation compiled: %d levels", treeAgg.Depth())

		// If final wrap is also requested, compile it using the tree root
		if inCircuits[circuits.FinalWrapCircuitID] {
			logrus.Infof("Setting up %s", circuits.FinalWrapCircuitID)
			rootComp := treeAgg.RootCompiledIOP()
			builder := finalwrap.NewBuilder(rootComp)
			if err := updateSetup(ctx, cfg, force, srsProvider, circuits.FinalWrapCircuitID, builder, nil); err != nil {
				return err
			}
		}
	} else if inCircuits[circuits.FinalWrapCircuitID] {
		// Final wrap without tree aggregation: wrap the execution leaf directly
		logrus.Infof("Setting up %s (wrapping execution leaf directly)", circuits.FinalWrapCircuitID)
		builder := finalwrap.NewBuilder(leafComp)
		if err := updateSetup(ctx, cfg, force, srsProvider, circuits.FinalWrapCircuitID, builder, nil); err != nil {
			return err
		}
	}

	return nil
}
