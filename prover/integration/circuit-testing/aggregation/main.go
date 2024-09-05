package main

import (
	"context"
	"fmt"
	"reflect"
	"sync"

	"github.com/consensys/gnark-crypto/ecc"
	frBls "github.com/consensys/gnark-crypto/ecc/bls12-377/fr"
	frBn254 "github.com/consensys/gnark-crypto/ecc/bn254/fr"
	frBw6 "github.com/consensys/gnark-crypto/ecc/bw6-761/fr"
	"github.com/consensys/gnark/backend/plonk"
	"github.com/consensys/gnark/constraint"
	"github.com/consensys/gnark/frontend"
	"github.com/consensys/gnark/frontend/cs/scs"
	emPlonk "github.com/consensys/gnark/std/recursion/plonk"
	"github.com/consensys/zkevm-monorepo/prover/circuits"
	"github.com/consensys/zkevm-monorepo/prover/circuits/aggregation"
	"github.com/consensys/zkevm-monorepo/prover/circuits/dummy"
	"github.com/consensys/zkevm-monorepo/prover/circuits/emulation"
	pi_interconnection "github.com/consensys/zkevm-monorepo/prover/circuits/pi-interconnection"
	"github.com/consensys/zkevm-monorepo/prover/config"
	dummyWizard "github.com/consensys/zkevm-monorepo/prover/protocol/compiler/dummy"
	public_input "github.com/consensys/zkevm-monorepo/prover/public-input"
	"github.com/consensys/zkevm-monorepo/prover/utils"
	"github.com/consensys/zkevm-monorepo/prover/utils/test_utils"
	"github.com/pkg/profile"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func main() {

	p := profile.Start()
	defer p.Stop()

	var t test_utils.FakeTestingT

	// Mock circuits to aggregate
	var innerSetups []circuits.Setup
	const nCircuits = 1 // TODO @Tabaie change to 10
	logrus.Infof("Initializing many inner-circuits of %v\n", nCircuits)
	srsProvider := circuits.NewUnsafeSRSProvider() // This is a dummy SRS provider, not to use in prod.
	for i := 0; i < nCircuits; i++ {
		logrus.Infof("\t%d/%d\n", i+1, nCircuits)
		pp, _ := dummy.MakeUnsafeSetup(srsProvider, circuits.MockCircuitID(i), ecc.BLS12_377.ScalarField())
		innerSetups = append(innerSetups, pp)
	}

	// This collects the verifying keys from the public parameters
	var vkeys []plonk.VerifyingKey
	for _, setup := range innerSetups {
		vkeys = append(vkeys, setup.VerifyingKey)
	}

	// At this step, we will collect several proofs for the BW6 circuit and
	// several verification keys.
	var (
		bw6Vkeys  []plonk.VerifyingKey
		bw6Proofs []plonk.Proof
	)

	//ncs := []int{1, 5, 10, 20} TODO @Tabaie change back to this
	ncs := []int{1}

	const hexZeroBlock = "0x0000000000000000000000000000000000000000000000000000000000000000"
	aggregationFPIHex := public_input.Aggregation{
		FinalShnarf:                             hexZeroBlock,
		ParentAggregationFinalShnarf:            hexZeroBlock,
		ParentStateRootHash:                     hexZeroBlock,
		ParentAggregationLastBlockTimestamp:     0,
		FinalTimestamp:                          0,
		LastFinalizedBlockNumber:                0,
		FinalBlockNumber:                        0,
		LastFinalizedL1RollingHash:              hexZeroBlock,
		L1RollingHash:                           hexZeroBlock,
		LastFinalizedL1RollingHashMessageNumber: 0,
		L1RollingHashMessageNumber:              0,
		L2MsgRootHashes:                         []string{},
		L2MsgMerkleTreeDepth:                    1,
	}
	aggregationPIBytes := aggregationFPIHex.Sum(nil)
	var aggregationPI frBw6.Element
	aggregationPI.SetBytes(aggregationPIBytes)

	logrus.Infof("Compiling interconnection circuit")
	piNbEachCircuit := utils.Max(ncs...)
	piSetup := getPiInterconnectionSetup(srsProvider, piNbEachCircuit)
	logrus.Infof("Assigning interconnection circuit")
	piAssignment, err := piSetup.c.Assign(pi_interconnection.Request{
		Decompressions: nil,
		Executions:     nil,
		Aggregation:    aggregationFPIHex,
	})
	assert.NoError(t, err)

	logrus.Infof("Generating PI proof")
	piW, err := frontend.NewWitness(&piAssignment, ecc.BLS12_377.ScalarField())
	assert.NoError(t, err)
	piProof, err := plonk.Prove(piSetup.Circuit, piSetup.ProvingKey, piW)
	assert.NoError(t, err)
	piPW, err := piW.Public()
	assert.NoError(t, err)
	piInfo := aggregation.PiInfo{
		Proof:         piProof,
		PublicWitness: piPW,
		ActualIndexes: []int{},
	}
	fmt.Println("aggregationPIBytes", utils.HexEncodeToString(aggregationPIBytes))

	for _, nc := range ncs {
		// Compile PI Interconnection sub-circuit

		// Building aggregation circuit for max `nc` proofs
		logrus.Infof("Building aggregation circuit for size of %v\n", nc)
		ccs, err := aggregation.MakeCS(
			nc,
			piSetup.Setup,
			vkeys)
		assert.NoError(t, err)

		// Generates the setup
		logrus.Infof("Generating setup, will take a while to complete..")
		ppBw6, err := circuits.MakeSetup(context.Background(), circuits.CircuitID(fmt.Sprintf("aggregation-%d", nc)), ccs, circuits.NewUnsafeSRSProvider(), nil)
		assert.NoError(t, err)

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
			assert.NoError(t, err)

			innerProofClaims[i] = aggregation.ProofClaimAssignment{
				CircuitID:   circID,
				Proof:       proof,
				PublicInput: x,
			}
		}

		// Assigning the BW6 circuit
		logrus.Infof("Generating the aggregation proof for arity %v", nc)

		bw6Proof, err := aggregation.MakeProof(&ppBw6, nc, innerProofClaims, piInfo, aggregationPI)
		assert.NoError(t, err)

		bw6Proofs = append(bw6Proofs, bw6Proof)
		bw6Vkeys = append(bw6Vkeys, ppBw6.VerifyingKey)
	}

	// Then create the BN254 circuit
	logrus.Infof("Generating the emulation circuit")
	ccsEmulation, err := emulation.MakeCS(bw6Vkeys)
	assert.NoError(t, err)

	logrus.Infof("Generating the setup for the emulation circuit")
	setupEmulation, err := circuits.MakeSetup(context.Background(), circuits.EmulationCircuitID, ccsEmulation, circuits.NewUnsafeSRSProvider(), nil)
	assert.NoError(t, err)

	for k := range bw6Proofs {
		logrus.Infof("Generating the proof for the emulation circuit (BW6 Proof #%v)", k)
		_, err = emulation.MakeProof(&setupEmulation, k, bw6Proofs[k], frBn254.NewElement(10))
		assert.NoError(t, err)
	}

}

func getPiInterconnectionSetup(srsProvider circuits.SRSProvider, numCircuits int) piInterconnectionSetup {
	var t test_utils.FakeTestingT

	circuit, err := pi_interconnection.Compile(config.PublicInput{
		MaxNbDecompression: numCircuits,
		MaxNbExecution:     numCircuits,
		MaxNbKeccakF:       numCircuits * 10,
		ExecutionMaxNbMsg:  1,
		L2MsgMerkleDepth:   1,
		L2MsgMaxNbMerkle:   1,
	}, dummyWizard.Compile)
	assert.NoError(t, err)

	var (
		cs            constraint.ConstraintSystem
		csCompilation sync.Mutex
	)
	csCompilation.Lock()
	go func() {
		var err error
		cs, err = frontend.Compile(ecc.BLS12_377.ScalarField(), scs.NewBuilder, circuit.Circuit) // always compiling the circuit in case it has changed
		csCompilation.Unlock()
		assert.NoError(t, err)
	}()

	cfg := config.Config{
		AssetsDir: "integration/circuit-testing/aggregation/.assets",
	}
	logrus.Infof("Loading PI setup")
	setup, err := circuits.LoadSetup(&cfg, circuits.PublicInputInterconnectionCircuitID)

	logrus.Infof("Compiling PI circuit")
	csCompilation.Lock()

	if err != nil || !reflect.DeepEqual(cs, setup.Circuit) { // TODO make sure this is a good way to compare cs's
		fmt.Println("commitment indexes", cs.GetCommitments().CommitmentIndexes())
		logrus.Infof("Loading failed. Creating PI setup")
		setup, err = circuits.MakeSetup(context.TODO(), circuits.PublicInputInterconnectionCircuitID, cs, srsProvider, nil)
		assert.NoError(t, err)
		logrus.Infof("Saving PI setup")
		/* assert.NoError(t, setup.WriteTo(cfg.PathForSetup(string(circuits.PublicInputInterconnectionCircuitID)))) No use storing the setup unless we can store SRS, TODO fix
		logrus.Infof("Copying SRS from gnark cache")	TODO Store SRS. Currently we're failing to load every time
		dst, err := os.Create(cfg.PathForSRS())
		assert.NoError(t, err)
		defer assert.NoError(t, dst.Close())
		homeDir, err := os.UserHomeDir()
		assert.NoError(t, err)
		src, err := os.Open(filepath.Join(homeDir, "."+"gnark", "kzg"))
		assert.NoError(t, err)
		defer assert.NoError(t, src.Close())
		_, err = io.Copy(dst, src)
		assert.NoError(t, err)*/
	}

	return piInterconnectionSetup{
		c:     circuit,
		Setup: setup,
	}
}

type piInterconnectionSetup struct {
	circuits.Setup
	c *pi_interconnection.Compiled
}
