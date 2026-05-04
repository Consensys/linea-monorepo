package circuits

import (
	"bytes"
	"fmt"
	"os"

	"github.com/consensys/gnark/backend"
	"github.com/consensys/gnark/backend/plonk"
	"github.com/consensys/gnark/constraint/solver"
	"github.com/consensys/gnark/frontend"
	"github.com/consensys/gnark/test"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/sirupsen/logrus"

	plonk_bn254 "github.com/consensys/gnark/backend/plonk/bn254"
	"github.com/consensys/linea-monorepo/prover/gpu"
	gpuplonk2 "github.com/consensys/linea-monorepo/prover/gpu/plonk2"
)

const envGPUPlonk2 = "LINEA_PROVER_GPU_PLONK2"

// Generates a PlonkProof and sanity-checks it against the verifying key. Can
// take a list of options which can of either backend.ProverOption of backend.
// VerifierOption.
func ProveCheck(setup *Setup, assignment frontend.Circuit, opts ...any) (plonk.Proof, error) {

	proverOpts := []backend.ProverOption{}
	verifierOpts := []backend.VerifierOption{}
	solverOpts := []solver.Option{}

	// @alex: we cannot incrementally pass the solver options to the prover
	// options (they are overridden at every call). That's why we need to collect
	// them first before passing them to the prover options.

	for _, opt := range opts {
		switch o := opt.(type) {
		case solver.Option:
			solverOpts = append(solverOpts, o)
		case backend.ProverOption:
			proverOpts = append(proverOpts, o)
		case backend.VerifierOption:
			verifierOpts = append(verifierOpts, o)
		default:
			return nil, fmt.Errorf("unknown option type to prove-check: %++v", o)
		}
	}

	proverOpts = append(proverOpts, backend.WithSolverOptions(solverOpts...))

	logrus.Infof("Creating the witness")
	witness, err := frontend.NewWitness(assignment, setup.Circuit.Field())
	if err != nil {
		return nil, fmt.Errorf("while generating the gnark witness: %w", err)
	}

	logrus.Infof("Generating the proof")
	var proof plonk.Proof

	if envFlag(envGPUPlonk2) {
		if !gpu.Enabled {
			return nil, fmt.Errorf("%s=1 requires a binary built with the cuda tag", envGPUPlonk2)
		}
		dev := gpu.GetDevice()
		if dev == nil {
			return nil, fmt.Errorf("%s=1 requested gpu/plonk2, but no GPU device is available", envGPUPlonk2)
		}
		logrus.Infof(
			"Generating the proof with gpu/plonk2: circuit=%T provingKey=%T verifyingKey=%T",
			setup.Circuit,
			setup.ProvingKey,
			setup.VerifyingKey,
		)
		var gpuProver *gpuplonk2.Prover
		gpuProver, err = gpuplonk2.NewProver(
			dev,
			setup.Circuit,
			setup.ProvingKey,
			setup.VerifyingKey,
			gpuplonk2.WithEnabled(true),
			gpuplonk2.WithStrictMode(true),
		)
		if err != nil {
			return nil, fmt.Errorf("while creating the gpu/plonk2 prover: %w", err)
		}
		defer gpuProver.Close()
		proof, err = gpuProver.Prove(witness, proverOpts...)
	} else {
		proof, err = plonk.Prove(setup.Circuit, setup.ProvingKey, witness, proverOpts...)
	}
	logrus.Infof("Generated proof type %T", proof)
	if err != nil {
		// The error returned by the Plonk prover is usually not helpful at
		// all. So, in order to get more details, we run the "test" Solver.
		logrus.Errorf("plonk.Prove returned an error, using the test.IsSolved to get more details: %s", err.Error())
		errDetail := test.IsSolved(
			assignment,
			assignment,
			setup.Circuit.Field(),
			// this test engine prover option was no-op before and it was removed
			// test.WithBackendProverOptions(proverOpts...),
		)
		return nil, fmt.Errorf("while running the plonk prover: %w", errDetail)
	}

	logrus.Infof("Sanity-checking the proof")
	// Sanity-check : the proof must pass
	{
		pubwitness, err := witness.Public()
		if err != nil {
			panic(err)
		}

		err = plonk.Verify(proof, setup.VerifyingKey, pubwitness, verifierOpts...)
		if err != nil {
			panic(err)
		}
		// logrus.Infof("the proof passed with\nproof=%++v\nwit=%++v\nvkey=%++v\n", proof, pubwitness, pp.VK)
	}

	return proof, nil
}

func envFlag(name string) bool {
	switch os.Getenv(name) {
	case "1", "true", "TRUE", "True", "yes", "YES", "on", "ON":
		return true
	default:
		return false
	}
}

// Serializes the proof in an 0x prefixed hexstring
func SerializeProofRaw(proof plonk.Proof) string {
	var buf bytes.Buffer
	proof.WriteRawTo(&buf)
	return hexutil.Encode(buf.Bytes())
}

// Serializes the proof in an 0x prefixed hexstring
func SerializeProofSolidityBn254(proof plonk.Proof) string {
	buf := proof.(*plonk_bn254.Proof).MarshalSolidity()
	return hexutil.Encode(buf)
}
