package execution

import (
	"fmt"
	"path/filepath"

	"github.com/consensys/gnark-crypto/ecc"
	"github.com/consensys/linea-monorepo/prover/circuits"
	"github.com/consensys/linea-monorepo/prover/circuits/dummy"
	"github.com/consensys/linea-monorepo/prover/circuits/execution"
	"github.com/consensys/linea-monorepo/prover/config"
	public_input "github.com/consensys/linea-monorepo/prover/public-input"
	"github.com/consensys/linea-monorepo/prover/utils"
	"github.com/consensys/linea-monorepo/prover/utils/exit"
	"github.com/consensys/linea-monorepo/prover/utils/profiling"
	"github.com/consensys/linea-monorepo/prover/zkevm"
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
		traces = &cfg.TracesLimitsLarge
	}

	switch cfg.Execution.ProverMode {
	case config.ProverModeDev, config.ProverModePartial:
		if cfg.Execution.ProverMode == config.ProverModePartial {

			logrus.Info("Running the PARTIAL prover")

			// And run the partial-prover with only the main steps. The generated
			// proof is sanity-checked to ensure that the prover never outputs
			// invalid proofs.
			partial := zkevm.FullZkEVMCheckOnly(traces, cfg)
			proof := partial.ProveInner(w.ZkEVM)
			if err := partial.VerifyInner(proof); err != nil {
				utils.Panic("The prover did not pass: %v", err)
			}
		}

		srsProvider, err := circuits.NewSRSStore(cfg.PathForSRS())
		if err != nil {
			panic(err.Error())
		}

		setup, err := dummy.MakeUnsafeSetup(srsProvider, circuits.MockCircuitIDExecution, ecc.BLS12_377.ScalarField())
		if err != nil {
			panic(err.Error())
		}

		return dummy.MakeProof(&setup, w.FuncInp.SumAsField(), circuits.MockCircuitIDExecution), setup.VerifyingKeyDigest()

	case config.ProverModeFull:
		logrus.Info("Running the FULL prover")

		// Run the full prover to obtain the intermediate proof
		logrus.Info("Get Full IOP")
		fullZkEvm := zkevm.FullZkEvm(traces, cfg)

		var (
			setup       circuits.Setup
			errSetup    error
			chSetupDone = make(chan struct{})
		)

		circuitID := circuits.ExecutionCircuitID
		if large {
			circuitID = circuits.ExecutionLargeCircuitID
		}

		// Sanity-check trace limits checksum between setup and config
		if err := SanityCheckTracesChecksum(circuitID, traces, cfg); err != nil {
			utils.Panic("traces checksum in the setup manifest does not match the one in the config: %v", err)
		}

		// Start loading the setup
		go func() {
			logrus.Infof("Loading setup - circuitID: %s", circuitID)
			setup, errSetup = circuits.LoadSetup(cfg, circuitID)
			close(chSetupDone)
		}()

		// Generates the inner-proof and sanity-check it so that we ensure that
		// the prover nevers outputs invalid proofs.
		proof := fullZkEvm.ProveInner(w.ZkEVM)

		logrus.Info("Sanity-checking the inner-proof")
		if err := fullZkEvm.VerifyInner(proof); err != nil {
			exit.OnUnsatisfiedConstraints(fmt.Errorf("the sanity-check of the inner-proof did not pass: %v", err))
		}

		// wait for setup to be loaded
		<-chSetupDone
		if errSetup != nil {
			utils.Panic("could not load setup: %v", errSetup)
		}

		// TODO: implements the collection of the functional inputs from the prover response
		return execution.MakeProof(traces, setup, fullZkEvm.WizardIOP, proof, *w.FuncInp), setup.VerifyingKeyDigest()

	case config.ProverModeBench:

		// Run the full prover to obtain the intermediate proof
		logrus.Info("Get Full IOP")
		fullZkEvm := zkevm.FullZkEvm(traces, cfg)

		// Generates the inner-proof and sanity-check it so that we ensure that
		// the prover nevers outputs invalid proofs.
		proof := fullZkEvm.ProveInner(w.ZkEVM)

		logrus.Info("Sanity-checking the inner-proof")
		if err := fullZkEvm.VerifyInner(proof); err != nil {
			utils.Panic("The prover did not pass: %v", err)
		}
		return "", ""

	case config.ProverModeCheckOnly:

		fullZkEvm := zkevm.FullZkEVMCheckOnly(traces, cfg)
		// this will panic to alert errors, so there is no need to handle or
		// sanity-check anything.
		logrus.Infof("Prover starting the prover")
		_ = fullZkEvm.ProveInner(w.ZkEVM)
		logrus.Infof("Prover checks passed")
		return "", ""

	default:
		panic("not implemented")
	}
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
