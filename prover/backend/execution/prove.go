package execution

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/consensys/gnark-crypto/ecc"
	"github.com/consensys/linea-monorepo/prover/circuits"
	"github.com/consensys/linea-monorepo/prover/circuits/dummy"
	"github.com/consensys/linea-monorepo/prover/config"
	"github.com/consensys/linea-monorepo/prover/protocol/compiler/recursion"
	"github.com/consensys/linea-monorepo/prover/protocol/serde"
	"github.com/consensys/linea-monorepo/prover/protocol/wizard"
	public_input "github.com/consensys/linea-monorepo/prover/public-input"
	"github.com/consensys/linea-monorepo/prover/utils"
	"github.com/consensys/linea-monorepo/prover/utils/exit"
	"github.com/consensys/linea-monorepo/prover/utils/profiling"
	"github.com/consensys/linea-monorepo/prover/zkevm"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/sirupsen/logrus"
)

type Witness struct {
	FuncInp *public_input.Execution
	ZkEVM   *zkevm.Witness
}

func Prove(cfg *config.Config, req *Request, large bool) (*Response, error) {

	// Set MonitorParams before any proving happens
	profiling.SetMonitorParams(cfg)

	// This instructs the [exit] package to actually exit when [OnLimitOverflow]
	// or [OnSatisfiedConstraints] are called.
	exit.SetIssueHandlingMode(exit.ExitAlways)

	var resp Response

	// TODO @gbotrel wrap profiling in the caller; so that we can properly return errors
	profiling.ProfileTrace("execution",
		cfg.Debug.Profiling,
		cfg.Debug.Tracing,
		func() {
			// Compute the prover's output.
			// WARN: CraftProverOutput calls functions that can panic.
			out := CraftProverOutput(cfg, req)

			if cfg.Execution.ProverMode != config.ProverModeProofless {
				// Development, Partial, Full or Full-large Mode

				// WARN:
				// - MustProveAndPass can panic
				// - Execution prover calls function that can panic
				// - NewFromString can panic
				out.Proof, out.VerifyingKeyShaSum = mustProveAndPass(
					cfg,
					NewWitness(cfg, req, &out),
					large,
				)

				out.Version = cfg.Version
				out.ProverMode = cfg.Execution.ProverMode
				out.VerifierIndex = uint(cfg.Aggregation.VerifierID) // TODO @gbotrel revisit
			} else {
				// Proofless Mode
				// task.ProverConfig.Mode() == config.ProverProofless
				logrus.Infof("Running the prover in proofless mode")
				out.Version = cfg.Version
				out.ProverMode = config.ProverModeProofless
			}

			resp = out
		})

	return &resp, nil
}

// mustProveAndPass the prover (in the void). Does not takes a
// prover-step function performing the assignment but a function
// returning such a function. This is important to avoid side-effects
// when calling it twice.
func mustProveAndPass(
	cfg *config.Config,
	w *Witness,
	large bool,
) (proofHexString string, vkeyShaSum string) {

	traces := &cfg.TracesLimits
	if large {
		traces.SetLargeMode()
	}

	switch cfg.Execution.ProverMode {
	case config.ProverModeDev, config.ProverModePartial:

		if cfg.Execution.ProverMode == config.ProverModePartial {

			logrus.Info("Running the PARTIAL prover")

			// And run the partial-prover with only the main steps. The generated
			// proof is sanity-checked to ensure that the prover never outputs
			// invalid proofs.
			checkOnlyZkEvm := zkevm.FullZkEVMCheckOnly(traces, cfg)
			proof := checkOnlyZkEvm.ProveInner(w.ZkEVM)
			if err := checkOnlyZkEvm.VerifyInner(proof); err != nil {
				utils.Panic("The prover did not pass: %v", err)
			}
		}

		srsProvider, err := circuits.NewSRSStore(cfg.PathForSRS())
		if err != nil {
			utils.Panic("could not create SRS provider: %v", err.Error())
		}

		setup, err := dummy.MakeUnsafeSetup(srsProvider, circuits.MockCircuitIDExecution, ecc.BLS12_377.ScalarField())
		if err != nil {
			utils.Panic("could not generate the mocked setup %v", err.Error())
		}

		return dummy.MakeProof(&setup, w.FuncInp.SumAsField(), circuits.MockCircuitIDExecution), setup.VerifyingKeyDigest()

	case config.ProverModeFull:

		logrus.Info("Running the FULL prover (tree aggregation mode)")

		circuitID := circuits.ExecutionCircuitID
		if large {
			circuitID = circuits.ExecutionLargeCircuitID
		}

		// Try loading serialized inner circuit, fall back to compilation
		var fullZkEvm *zkevm.ZkEvm
		enabled, innerPath := cfg.ExecutionCircuitBin(string(circuitID))
		if enabled {
			if _, err := os.Stat(innerPath); err == nil {
				logrus.Infof("Loading serialized inner circuit from %s", innerPath)
				var loaded zkevm.ZkEvm
				closer, loadErr := serde.LoadFromDisk(innerPath, &loaded, false)
				if loadErr == nil {
					defer closer.Close()
					fullZkEvm = &loaded
				} else {
					logrus.Warnf("Failed to load inner circuit: %v. Falling back to compilation.", loadErr)
				}
			} else {
				logrus.Warnf("Serialization enabled but %s not found. Falling back to compilation.", innerPath)
			}
		}
		if fullZkEvm == nil {
			logrus.Info("Get Full IOP")
			fullZkEvm = zkevm.FullZkEvm(traces, cfg)
		}

		// Run the full prover to obtain the recursion witness for tree
		// aggregation. This replaces the old BLS12-377 PLONK wrapping path.
		recursionWit, _ := fullZkEvm.ProveForTree(w.ZkEVM)

		// Serialize the recursion witness for later tree aggregation.
		// The aggregation job will deserialize these and feed them into
		// the binary tree aggregation pipeline.
		witBytes, err := serde.Serialize(recursionWit)
		if err != nil {
			utils.Panic("could not serialize the recursion witness: %v", err)
		}

		proofHexString = hexutil.Encode(witBytes)
		logrus.Infof("Serialized recursion witness: %d bytes", len(witBytes))
		return proofHexString, ""

	case config.ProverModeBench:

		panic("uncomment, when benchmark prover is ready for testing")

		// // Run the full prover to obtain the intermediate proof
		// logrus.Info("Get Full IOP")
		// fullZkEvm := zkevm.FullZkEvm(traces, cfg)

		// // Generates the inner-proof and sanity-check it so that we ensure that
		// // the prover nevers outputs invalid proofs.
		// proof := fullZkEvm.ProveInner(w.ZkEVM)

		// logrus.Info("Sanity-checking the inner-proof")
		// if err := fullZkEvm.VerifyInner(proof); err != nil {
		// 	utils.Panic("The prover did not pass: %v", err)
		// }
		// return "", ""

	case config.ProverModeCheckOnly:

		checkOnlyZkEvm := zkevm.FullZkEVMCheckOnly(traces, cfg)
		// this will panic to alert errors, so there is no need to handle or
		// sanity-check anything.
		logrus.Infof("Prover starting the prover")
		_ = checkOnlyZkEvm.ProveInner(w.ZkEVM)
		logrus.Infof("Prover checks passed")
		return "", ""

	case config.ProverModeEncodeOnly:

		panic("uncomment; when it is ready. Do we actually need it")

		// profiling.ProfileTrace("encode-decode-no-circuit", true, false, func() {
		// 	filepath := "/tmp/wizard-assignment/blob-" + strconv.Itoa(rand.Int()) + ".bin" //nolint:gosec // Ignoring weak randomness error

		// 	encodeOnlyZkEvm := zkevm.EncodeOnlyZkEvm(traces)
		// 	numChunks := runtime.GOMAXPROCS(0)

		// 	// Serialize the assignment
		// 	encodingDuration := time.Now()
		// 	encodeOnlyZkEvm.AssignAndEncodeInChunks(filepath, w.ZkEVM, numChunks)

		// 	// Deserialize the assignment
		// 	decodingDuration := time.Now()
		// 	_, errDec := serialization.DeserializeAssignment(filepath, numChunks)
		// 	if errDec != nil {
		// 		panic(fmt.Sprintf("Error during deserialization: %v", errDec))
		// 	}
		// 	fmt.Printf("[Encoding Summary] took %v sec to encode an assignmente and write it into the files \n", time.Since(encodingDuration).Seconds())
		// 	fmt.Printf("[Decoding Summary] took %v sec to read the files and decode it into an assignment\n", time.Since(decodingDuration).Seconds())
		// })

		// os.Exit(0)
		// return "", ""

	default:
		panic("not implemented")
	}
}

// TreeExecutionResult carries the wizard-level artifacts from the execution
// prover, suitable for feeding into the tree aggregation pipeline.
type TreeExecutionResult struct {
	// RecursionWitness is the recursion witness extracted from the execution
	// prover's IOP, stopped at the Vortex query round.
	RecursionWitness recursion.Witness
	// LeafCompiledIOP is the CompiledIOP from which the witness was extracted.
	// This must match the tree aggregation's leaf ChildComp.
	LeafCompiledIOP *wizard.CompiledIOP
	// FuncInp is the functional public input for this execution.
	FuncInp *public_input.Execution
}

// ProveForTree runs the execution prover in tree aggregation mode. Instead of
// wrapping the wizard proof in BLS12-377 PLONK, it returns the recursion
// witness for feeding into the tree aggregation pipeline.
func ProveForTree(cfg *config.Config, req *Request, large bool) (*TreeExecutionResult, error) {
	profiling.SetMonitorParams(cfg)
	exit.SetIssueHandlingMode(exit.ExitAlways)

	var result *TreeExecutionResult

	profiling.ProfileTrace("execution-tree",
		cfg.Debug.Profiling,
		cfg.Debug.Tracing,
		func() {
			out := CraftProverOutput(cfg, req)
			w := NewWitness(cfg, req, &out)

			traces := &cfg.TracesLimits
			if large {
				traces.SetLargeMode()
			}

			logrus.Info("Running the FULL prover for tree aggregation")
			fullZkEvm := zkevm.FullZkEvm(traces, cfg)

			recursionWitness, leafComp := fullZkEvm.ProveForTree(w.ZkEVM)

			result = &TreeExecutionResult{
				RecursionWitness: recursionWitness,
				LeafCompiledIOP:  leafComp,
				FuncInp:          w.FuncInp,
			}
		})

	return result, nil
}

// SanityCheckTraceChecksum ensures the checksum for the traces in the setup matches the one in the config
func SanityCheckTracesChecksum(circuitID circuits.CircuitID, traces *config.TracesLimits, cfg *config.Config) error {

	// read setup manifest
	manifestPath := filepath.Join(cfg.PathForSetup(string(circuitID)), config.ManifestFileName)
	manifest, err := circuits.ReadSetupManifest(manifestPath)
	if err != nil {
		utils.Panic("could not read the setup manifest: %v", err)
	}

	// read manifest traces checksum
	setupCfgChecksum, err := manifest.GetString("cfg_checksum")
	if err != nil {
		utils.Panic("could not get the traces checksum from the setup manifest: %v", err)
	}

	if setupCfgChecksum != traces.Checksum() {
		// This check is failing on prod but works locally.
		// @alex: since this is a setup-related constraint, it would likely be
		// more interesting to directly include that information in the setup
		// instead of the config. That way we are guaranteed to not pass the
		// wrong value at runtime.
		return fmt.Errorf("setup (%s): '%s' vs config: '%s'", circuitID, setupCfgChecksum, traces.Checksum())
	}

	return nil
}
