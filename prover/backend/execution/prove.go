package execution

import (
	"github.com/consensys/gnark-crypto/ecc"
	"github.com/consensys/linea-monorepo/prover/circuits"
	"github.com/consensys/linea-monorepo/prover/circuits/dummy"
	"github.com/consensys/linea-monorepo/prover/circuits/execution"
	"github.com/consensys/linea-monorepo/prover/config"
	public_input "github.com/consensys/linea-monorepo/prover/public-input"
	"github.com/consensys/linea-monorepo/prover/utils"
	"github.com/consensys/linea-monorepo/prover/utils/profiling"
	"github.com/consensys/linea-monorepo/prover/zkevm"
	"github.com/sirupsen/logrus"
)

type Witness struct {
	FuncInp *public_input.Execution
	ZkEVM   *zkevm.Witness
}

// Prove function orchestrates the proving process
func Prove(cfg *config.Config, req *Request, large bool) (*Response, error) {
	// Set MonitorParams before any proving happens
	profiling.SetMonitorParams(cfg)

	// Select traces based on the mode (large or regular)
	traces := selectTraces(cfg, large)

	var resp Response

	// Profile the execution trace
	profiling.ProfileTrace("execution",
		cfg.Debug.Profiling,
		cfg.Debug.Tracing,
		func() {
			// Compute the prover's output
			// WARN: CraftProverOutput calls functions that can panic.
			out := CraftProverOutput(cfg, req)

			if cfg.Execution.ProverMode != config.ProverModeProofless {
				// Development, Partial, Full, or Full-large Mode

				// WARN:
				// - MustProveAndPass can panic
				// - Execution prover calls functions that can panic
				// - NewFromString can panic
				out.Proof, out.VerifyingKeyShaSum = mustProveAndPass(cfg, traces, NewWitness(cfg, req, &out))

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

// Select traces based on the mode (large or regular)
func selectTraces(cfg *config.Config, large bool) *config.TracesLimits {
	if large {
		return &cfg.TracesLimitsLarge
	}
	return &cfg.TracesLimits
}

// mustProveAndPass handles the proving process based on the prover mode
func mustProveAndPass(cfg *config.Config, traces *config.TracesLimits, w *Witness) (proofHexString string, vkeyShaSum string) {
	switch cfg.Execution.ProverMode {
	case config.ProverModeDev, config.ProverModePartial:
		return handlePartialProver(cfg, traces, w)

	case config.ProverModeFull:
		return handleFullProver(cfg, traces, w)

	case config.ProverModeBench:
		return handleBenchProver(cfg, traces, w)

	case config.ProverModeCheckOnly:
		return handleCheckOnlyProver(cfg, traces, w)

	default:
		panic("not implemented")
	}
}

// Handle partial prover mode
func handlePartialProver(cfg *config.Config, traces *config.TracesLimits, w *Witness) (string, string) {
	if cfg.Execution.ProverMode == config.ProverModePartial {
		logrus.Info("Running the PARTIAL prover")

		// Run the partial-prover with only the main steps. The generated
		// proof is sanity-checked to ensure that the prover never outputs
		// invalid proofs.
		partial := zkevm.FullZkEVMCheckOnly(traces, cfg)
		proof := partial.ProveInner(w.ZkEVM)
		if err := partial.VerifyInner(proof); err != nil {
			utils.Panic("The prover did not pass: %v", err)
		}
	}

	// Generate the setup for the partial prover
	srsProvider, err := circuits.NewSRSStore(cfg.PathForSRS())
	if err != nil {
		panic(err.Error())
	}

	setup, err := dummy.MakeUnsafeSetup(srsProvider, circuits.MockCircuitIDExecution, ecc.BLS12_377.ScalarField())
	if err != nil {
		panic(err.Error())
	}

	return dummy.MakeProof(&setup, w.FuncInp.SumAsField(), circuits.MockCircuitIDExecution), setup.VerifyingKeyDigest()
}

// Handle full prover mode
func handleFullProver(cfg *config.Config, traces *config.TracesLimits, w *Witness) (string, string) {
	logrus.Info("Running the FULL prover")

	// Run the full prover to obtain the intermediate proof
	logrus.Info("Get Full IOP")
	fullZkEvm := zkevm.FullZkEvm(traces, cfg)

	var (
		setup       circuits.Setup
		errSetup    error
		chSetupDone = make(chan struct{})
	)
	go func() {
		setup, errSetup = circuits.LoadSetup(cfg, circuits.ExecutionCircuitID)
		close(chSetupDone)
	}()

	// Generates the inner-proof and sanity-check it so that we ensure that
	// the prover never outputs invalid proofs.
	proof := fullZkEvm.ProveInner(w.ZkEVM)

	logrus.Info("Sanity-checking the inner-proof")
	if err := fullZkEvm.VerifyInner(proof); err != nil {
		utils.Panic("The prover did not pass: %v", err)
	}

	// Wait for setup to be loaded
	<-chSetupDone
	if errSetup != nil {
		utils.Panic("could not load setup: %v", errSetup)
	}

	// Ensure the checksum for the traces in the setup matches the one in the config
	ValidateSetupChecksum(setup, traces)

	// TODO: implements the collection of the functional inputs from the prover response
	return execution.MakeProof(traces, setup, fullZkEvm.WizardIOP, proof, *w.FuncInp), setup.VerifyingKeyDigest()
}

// Validate setup checksum
func ValidateSetupChecksum(setup circuits.Setup, traces *config.TracesLimits) {
	setupCfgChecksum, err := setup.Manifest.GetString("cfg_checksum")
	if err != nil {
		utils.Panic("could not get the traces checksum from the setup manifest: %v", err)
	}

	if setupCfgChecksum != traces.Checksum() {
		// This check is failing on prod but works locally.
		// @alex: since this is a setup-related constraint, it would likely be
		// more interesting to directly include that information in the setup
		// instead of the config. That way we are guaranteed to not pass the
		// wrong value at runtime.
		utils.Panic("traces checksum in the setup manifest does not match the one in the config")
	}
}

// Handle bench prover mode
func handleBenchProver(cfg *config.Config, traces *config.TracesLimits, w *Witness) (string, string) {
	logrus.Info("Get Full IOP")
	fullZkEvm := zkevm.FullZkEvm(traces, cfg)

	// Generates the inner-proof and sanity-check it so that we ensure that
	// the prover never outputs invalid proofs.
	proof := fullZkEvm.ProveInner(w.ZkEVM)

	logrus.Info("Sanity-checking the inner-proof")
	if err := fullZkEvm.VerifyInner(proof); err != nil {
		utils.Panic("The prover did not pass: %v", err)
	}
	return "", ""
}

// Handle check-only prover mode
func handleCheckOnlyProver(cfg *config.Config, traces *config.TracesLimits, w *Witness) (string, string) {
	fullZkEvm := zkevm.FullZkEVMCheckOnly(traces, cfg)

	// This will panic to alert errors, so there is no need to handle or
	// sanity-check anything.
	logrus.Infof("Prover starting the prover")
	_ = fullZkEvm.ProveInner(w.ZkEVM)
	logrus.Infof("Prover checks passed")
	return "", ""
}
