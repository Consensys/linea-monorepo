package main

import (
	"context"
	"fmt"

	"github.com/consensys/gnark-crypto/ecc"
	frBls "github.com/consensys/gnark-crypto/ecc/bls12-377/fr"
	frBn254 "github.com/consensys/gnark-crypto/ecc/bn254/fr"
	frBw6 "github.com/consensys/gnark-crypto/ecc/bw6-761/fr"
	"github.com/consensys/gnark/backend/plonk"
	emPlonk "github.com/consensys/gnark/std/recursion/plonk"
	"github.com/consensys/zkevm-monorepo/prover/circuits"
	"github.com/consensys/zkevm-monorepo/prover/circuits/aggregation"
	"github.com/consensys/zkevm-monorepo/prover/circuits/dummy"
	"github.com/consensys/zkevm-monorepo/prover/circuits/emulation"
	"github.com/consensys/zkevm-monorepo/prover/utils"
	"github.com/sirupsen/logrus"
)

func main() {

	// Mock circuits to aggregate
	var innerSetups []circuits.Setup
	nCircuits := 10
	logrus.Infof("Initializing many inner-circuits of %v\n", nCircuits)
	srsProvider := circuits.NewUnsafeSRSProvider() // This is a dummy SRS provider, not to use in prod.
	for i := 0; i < nCircuits; i++ {
		pp, _ := dummy.MakeUnsafeSetup(srsProvider, circuits.MockCircuitID(i), ecc.BLS12_377.ScalarField())
		innerSetups = append(innerSetups, pp)
	}

	// This collects the verifying keys from the public parameters
	vkeys := []plonk.VerifyingKey{}
	for _, setup := range innerSetups {
		vkeys = append(vkeys, setup.VerifyingKey)
	}

	// At this step, we will collect several proofs for the BW6 circuit and
	// several verification keys.
	bw6Vkeys := []plonk.VerifyingKey{}
	bw6Proofs := []plonk.Proof{}

	for _, nc := range []int{1, 5, 10, 20} {

		// Building aggregation circuit for max `nc` proofs
		logrus.Infof("Building aggregation circuit for size of %v\n", nc)
		ccs, err := aggregation.MakeCS(nc, []string{"blob-decompression-v1", "execution"}, nil, vkeys) // TODO @Tabaie add a PI key
		if err != nil {
			panic(err)
		}

		// Generates the setup
		logrus.Infof("Generating setup, will take a while to complete..")
		ppBw6, err := circuits.MakeSetup(context.Background(), circuits.CircuitID(fmt.Sprintf("aggregation-%d", nc)), ccs, circuits.NewUnsafeSRSProvider(), nil)
		if err != nil {
			panic(err)
		}

		// Generate proofs claims to aggregate
		nProofs := utils.Max(nc-3, 1)
		logrus.Infof("Generating a witness, %v dummy-proofs to aggregates", nProofs)
		innerProofClaims := make([]aggregation.ProofClaimAssignment, nProofs)
		for i := range innerProofClaims {

			// Assign the dummy circuit for a random value
			circID := i % nCircuits
			var x frBls.Element
			x.SetRandom()
			a := dummy.Assign(circuits.MockCircuitID(circID), x)

			// Stores the inner-proofs for later
			proof, err := circuits.ProveCheck(
				&innerSetups[circID], a,
				emPlonk.GetNativeProverOptions(ecc.BW6_761.ScalarField(), ecc.BLS12_377.ScalarField()),
				emPlonk.GetNativeVerifierOptions(ecc.BW6_761.ScalarField(), ecc.BLS12_377.ScalarField()),
			)
			if err != nil {
				panic(err)
			}

			innerProofClaims[i] = aggregation.ProofClaimAssignment{
				CircuitID:   circID,
				Proof:       proof,
				PublicInput: x,
			}
		}

		// Assigning the BW6 circuit
		logrus.Infof("Generating the aggregation proof for arity %v", nc)

		// TODO @Tabaie add a PI proof
		bw6Proof, err := aggregation.MakeProof(&ppBw6, nc, innerProofClaims, nil, frBw6.NewElement(10))
		if err != nil {
			panic(err)
		}

		bw6Proofs = append(bw6Proofs, bw6Proof)
		bw6Vkeys = append(bw6Vkeys, ppBw6.VerifyingKey)
	}

	// Then create the BN254 circuit
	logrus.Infof("Generating the emulation circuit")
	ccsEmulation, err := emulation.MakeCS(bw6Vkeys)
	if err != nil {
		panic(err)
	}

	logrus.Infof("Generating the setup for the emulation circuit")
	setupEmulation, err := circuits.MakeSetup(context.Background(), circuits.EmulationCircuitID, ccsEmulation, circuits.NewUnsafeSRSProvider(), nil)
	if err != nil {
		panic(err)
	}

	for k := range bw6Proofs {
		logrus.Infof("Generating the proof for the emulation circuit (BW6 Proof #%v)", k)
		_, err = emulation.MakeProof(&setupEmulation, k, bw6Proofs[k], frBn254.NewElement(10))
		if err != nil {
			panic(err)
		}
	}

}
