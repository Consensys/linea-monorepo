package execution

import (
	"github.com/consensys/gnark-crypto/ecc"
	"github.com/consensys/linea-monorepo/prover/circuits"
	"github.com/consensys/linea-monorepo/prover/circuits/dummy"
	"github.com/consensys/linea-monorepo/prover/circuits/execution"
	"github.com/consensys/linea-monorepo/prover/config"
	"github.com/consensys/linea-monorepo/prover/maths/field"
	"github.com/consensys/linea-monorepo/prover/utils"
	"github.com/consensys/linea-monorepo/prover/utils/profiling"
	"github.com/consensys/linea-monorepo/prover/zkevm"
	"github.com/sirupsen/logrus"
)

func Prove(cfg *config.Config, req *Request, large bool) (*Response, error) {
	traces := &cfg.TracesLimits
	if large {
		traces = &cfg.TracesLimitsLarge
	}

	var resp Response

	// TODO @gbotrel wrap profiling in the caller; so that we can properly return errors
	profiling.ProfileTrace("execution",
		cfg.Debug.Profiling,
		cfg.Debug.Tracing,
		func() {
			// Compute the prover's output.
			// WARN: CraftProverOutput calls functions that can panic.
			out, smTraces := CraftProverOutput(cfg, req)

			if cfg.Execution.ProverMode != config.ProverModeProofless {
				// Development, Partial, Full or Full-large Mode

				// inputs of the prover
				proverInps := &zkevm.Witness{
					ExecTracesFPath: req.ConflatedExecTraceFilepath(cfg.Execution.ConflatedTracesDir),
					SMTraces:        smTraces,
				}

				// WARN:
				// - MustProveAndPass can panic
				// - Execution prover calls function that can panic
				// - NewFromString can panic
				out.Proof, out.VerifyingKeyShaSum = mustProveAndPass(
					proverInps,
					cfg,
					traces,
					field.NewFromString(out.DebugData.FinalHash),
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
	witness *zkevm.Witness,
	cfg *config.Config,
	traces *config.TracesLimits,
	publicInput field.Element,
) (proofHexString string, vkeyShaSum string) {

	switch cfg.Execution.ProverMode {
	case config.ProverModeDev, config.ProverModePartial:
		if cfg.Execution.ProverMode == config.ProverModePartial {
			logrus.Info("Running the CHECKER before the PARTIAL prover")
			// If in partial-mode, we still want to make sure that the trace
			// was correctly generated even though we do not prove it.
			// Run the checker with all the options
			checker := zkevm.CheckerZkEvm(traces)
			proof := checker.ProveInner(witness)
			if err := checker.VerifyInner(proof); err != nil {
				utils.Panic("(PARTIAL-MODE ONLY) The checker did not pass: %v", err)
			}

			logrus.Info("Running the PARTIAL prover")

			// And run the partial-prover with only the main steps. The generated
			// proof is sanity-checked to ensure that the prover never outputs
			// invalid proofs.
			partial := zkevm.PartialZkEvm(traces)
			proof = partial.ProveInner(witness)
			if err := partial.VerifyInner(proof); err != nil {
				utils.Panic("The prover did not pass: %v", err)
			}
		}

		srsProvider, err := circuits.NewSRSStore(cfg.PathForSRS())
		if err != nil {
			utils.Panic(err.Error())
		}

		setup, err := dummy.MakeUnsafeSetup(srsProvider, circuits.MockCircuitIDExecution, ecc.BLS12_377.ScalarField())
		if err != nil {
			utils.Panic(err.Error())
		}

		return dummy.MakeProof(&setup, publicInput, circuits.MockCircuitIDExecution), setup.VerifiyingKeyDigest()

	case config.ProverModeFull:
		logrus.Info("Running the FULL prover")

		// Run the full prover to obtain the intermediate proof
		logrus.Info("Get Full IOP")
		fullZkEvm := zkevm.FullZkEvm(&cfg.Execution.Features, traces)

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
		// the prover nevers outputs invalid proofs.
		proof := fullZkEvm.ProveInner(witness)

		logrus.Info("Sanity-checking the inner-proof")
		if err := fullZkEvm.VerifyInner(proof); err != nil {
			utils.Panic("The prover did not pass: %v", err)
		}

		// wait for setup to be loaded
		<-chSetupDone
		if errSetup != nil {
			utils.Panic("could not load setup: %v", errSetup)
		}

		// ensure the checksum for the traces in the setup matches the one in the config
		setupCfgChecksum, err := setup.Manifest.GetString("cfg_checksum")
		if err != nil {
			utils.Panic("could not get the traces checksum from the setup manifest: %v", err)
		}
		if setupCfgChecksum != traces.Checksum()+cfg.Execution.Features.Checksum() {
			utils.Panic("traces + features checksum in the setup manifest does not match the one in the config")
		}

		// TODO: implements the collection of the functional inputs from the prover response
		return execution.MakeProof(setup, fullZkEvm.WizardIOP, proof, execution.FunctionalPublicInput{}, publicInput), setup.VerifiyingKeyDigest()
	default:
		panic("not implemented")
	}
}
