package backend

import (
	"github.com/consensys/accelerated-crypto-monorepo/backend/config"
	"github.com/consensys/accelerated-crypto-monorepo/backend/coordinator"
	"github.com/consensys/accelerated-crypto-monorepo/backend/files"
	"github.com/consensys/accelerated-crypto-monorepo/backend/prover"
	"github.com/consensys/accelerated-crypto-monorepo/maths/field"
	"github.com/consensys/accelerated-crypto-monorepo/protocol/wizard"
	"github.com/consensys/accelerated-crypto-monorepo/utils/profiling"
	"github.com/consensys/accelerated-crypto-monorepo/zkevm"
	"github.com/sirupsen/logrus"
)

func RunCorsetAndProver(inputPath, outputPath string) {
	profiling.ProfileTrace("testnet-prover",
		config.MustGetProver().ProfilingEnabled,
		config.MustGetProver().TracingEnabled,
		func() {
			inFile := files.MustRead(inputPath)
			// Only deserializing the block data here
			in := &coordinator.ProverInput{}
			in.Read(inFile)

			// Craft the prover's input prior to runnin the prover itself
			// this is necessary because we will need some of the intermediate
			// values to assigns (the transaction hashing, the transaction
			// signature, the state-manager's merkle-proofs etc...).

			// This computes the prover output
			out, smTraces := coordinator.CraftProverOutput(
				in,
				config.MustGetLayer2().MustGetMsgServiceContract(),
			)

			// config of the prover
			proconf := config.MustGetProver()

			// setup the prover options according to the config. This is used
			// to pass auxiliary information to the prover's runtime (tx signatures,
			// state-manager traces etc...)
			proverOptions := prover.NewProverOptions()

			if proconf.WithStateManager {
				proverOptions.WithStateManager(smTraces)
			}

			if proconf.WithKeccak {
				proverOptions.WithKeccak()
			}

			if proconf.WithEcdsa {
				proverOptions.WithEcdsa(in.GetRawSignaturesVerificationInputs)
			}

			// The prover can be run in 3 modes, notrace, light, full
			// notrace, skips all the prover's steps and generates a dummy proof
			// light, generates a proof checking a tiny fraction of the constraints
			// full, checks everything
			mode := proconf.Mode()
			if proconf.SkipTraces {
				logrus.Warnf("The PROVER_SKIP_TRACES environment variable is set, so the prover will not be run")
				mode = prover.NOTRACE
			}

			// In light-mode : run the prover over the vortex and make sure it passes
			// in dummy mode before generating the light-proof.
			out.Proof = prover.MustProveAndPass(
				mode,
				in.Exprover(config.MustGetProver().ConflatedTracesDir),
				proverOptions,
				field.NewFromString(out.DebugData.FinalHash),
			)

			// Finally write the prover version in the output file
			out.Version = config.MustGetProver().Version

			// If the dev-light functionality is enabled, we pass a flag in the output file
			out.ProverMode = config.MustGetProver().Mode()
			out.VerifierIndex = config.MustGetProver().VerifierIndex()

			// And write the output
			out.WriteInFile(outputPath)
		})
}

func RunCorsetAndChecker(inputPath string) {
	inFile := files.MustReadCompressed(inputPath)
	prover.MustPassCheck(
		func(run *wizard.ProverRuntime) {
			zkevm.AssignFromCorset(inFile, run)
		},
		prover.NewProverOptions(),
	)
}
