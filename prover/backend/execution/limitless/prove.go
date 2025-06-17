package limitless

import (
	"bytes"
	"fmt"
	"os"
	"path"

	"github.com/consensys/linea-monorepo/prover/backend/execution"
	"github.com/consensys/linea-monorepo/prover/circuits"
	ckt_exec "github.com/consensys/linea-monorepo/prover/circuits/execution"
	"github.com/consensys/linea-monorepo/prover/config"
	"github.com/consensys/linea-monorepo/prover/maths/field"
	"github.com/consensys/linea-monorepo/prover/protocol/compiler/recursion"
	"github.com/consensys/linea-monorepo/prover/protocol/distributed"
	"github.com/consensys/linea-monorepo/prover/protocol/serialization"
	"github.com/consensys/linea-monorepo/prover/protocol/wizard"
	"github.com/consensys/linea-monorepo/prover/utils"
	"github.com/consensys/linea-monorepo/prover/utils/profiling"
	"github.com/consensys/linea-monorepo/prover/zkevm"
	"github.com/sirupsen/logrus"
)

// Asset struct to hold deserialized assets
type Asset struct {
	Zkevm      *zkevm.ZkEvm
	Disc       *distributed.StandardModuleDiscoverer
	DistWizard *distributed.DistributedWizard
}

// Prove function for the Assest struct
func (asset *Asset) Prove(cfg *config.Config, req *execution.Request) (*execution.Response, error) {
	// Set MonitorParams before any proving happens
	profiling.SetMonitorParams(cfg)

	defer func() {
		filepath := cfg.PathforLimitlessProverAssets()
		filepath = path.Join(filepath, "witness")
		os.RemoveAll(filepath)
	}()

	// Setup execution circuit
	var (
		setup       circuits.Setup
		errSetup    error
		chSetupDone = make(chan struct{})
	)
	go func() {
		setup, errSetup = circuits.LoadSetup(cfg, circuits.ExecutionCircuitID)
		close(chSetupDone)
	}()

	// Setup execution witness and output response
	var (
		out     = execution.CraftProverOutput(cfg, req)
		witness = execution.NewWitness(cfg, req, &out)
	)

	logrus.Info("Starting to run the bootstrapper")
	var (
		runtimeBoot             = wizard.RunProver(asset.DistWizard.Bootstrapper, asset.Zkevm.GetMainProverStep(witness.ZkEVM))
		witnessGLs, witnessLPPs = distributed.SegmentRuntime(runtimeBoot, asset.DistWizard)
	)
	logrus.Info("Finished running the bootstrapper")

	logrus.Info("Starting to run GL Prover")
	runGLs, err := RunProverGLs(cfg, asset.DistWizard, witnessGLs)
	if err != nil {
		return nil, err
	}
	SanityCheckProvers(asset.DistWizard, runGLs)
	logrus.Info("Finished running GL Prover")

	logrus.Info("Starting to generate shared Randomness")
	sharedRandomness := GetSharedRandomness(runGLs)
	logrus.Info("Finished generating shared Randomness")

	logrus.Info("Starting to run LPP Prover")
	runLPPs, err := RunProverLPPs(cfg, asset.DistWizard, sharedRandomness, witnessLPPs)
	if err != nil {
		return nil, err
	}
	SanityCheckProvers(asset.DistWizard, runLPPs)
	logrus.Info("Finished running LPP Prover")

	logrus.Info("Starting to run Conglomerator")
	proof, err := RunConglomerationProver(cfg, asset.DistWizard.CompiledConglomeration, witnessGLs, witnessLPPs)
	if err != nil {
		return nil, err
	}
	logrus.Info("Finished running Conglomerator")

	// wait for setup to be loaded
	<-chSetupDone
	if errSetup != nil {
		utils.Panic("could not load setup: %v", errSetup)
	}

	execution.ValidateSetupChecksum(setup, &cfg.TracesLimits)
	out.Proof = ckt_exec.MakeProof(&config.TracesLimits{}, setup, asset.Zkevm.WizardIOP, *proof, *witness.FuncInp)
	out.VerifyingKeyShaSum = setup.VerifyingKeyDigest()
	return &out, nil
}

func SanityCheckProvers(distWizard *distributed.DistributedWizard, runs []*wizard.ProverRuntime) {
	for i := range runs {
		logrus.Infof("sanity-checking prover run[%d]", i)
		SanityCheckConglomeration(distWizard.CompiledConglomeration, runs[i])
	}
}

func GetSharedRandomness(runs []*wizard.ProverRuntime) field.Element {
	witnesses := make([]recursion.Witness, len(runs))
	for i := range runs {
		witnesses[i] = recursion.ExtractWitness(runs[i])
	}

	comps := make([]*wizard.CompiledIOP, len(runs))
	for i := range runs {
		comps[i] = runs[i].Spec
	}
	return distributed.GetSharedRandomnessFromWitnesses(comps, witnesses)
}

// Unified function to read and deserialize all assets and compiled files
func ReadAndDeserAssets(config *config.Config) (*Asset, error) {
	if config == nil {
		return nil, fmt.Errorf("config is nil")
	}

	filePath := config.PathforLimitlessProverAssets()
	var readBuf bytes.Buffer

	// Initialize result struct
	assets := &Asset{
		Zkevm:      &zkevm.ZkEvm{},
		Disc:       &distributed.StandardModuleDiscoverer{},
		DistWizard: &distributed.DistributedWizard{},
	}

	// Define all files to read and deserialize
	files := []struct {
		name   string
		target any
	}{
		{name: "zkevm.bin", target: &assets.Zkevm},
		{name: "disc.bin", target: &assets.Disc},
		{name: "dw-raw.bin", target: &assets.DistWizard},
		{name: "dw-compiled-default.bin", target: &assets.DistWizard.CompiledDefault},
	}

	// Read and deserialize each file
	for _, file := range files {
		if err := readAndDeserialize(filePath, file.name, file.target, &readBuf); err != nil {
			return nil, err
		}
	}

	// Read and deserialize GL modules
	for i := 0; i < len(assets.DistWizard.GLs); i++ {
		var compiledGL *distributed.RecursedSegmentCompilation
		fileName := fmt.Sprintf("dw-compiled-gl-%d.bin", i)
		if err := readAndDeserialize(filePath, fileName, &compiledGL, &readBuf); err != nil {
			return nil, err
		}
		assets.DistWizard.CompiledGLs = append(assets.DistWizard.CompiledGLs, compiledGL)
	}

	// Read and deserialize LPP modules
	for i := 0; i < len(assets.DistWizard.LPPs); i++ {
		var compiledLPP *distributed.RecursedSegmentCompilation
		fileName := fmt.Sprintf("dw-compiled-lpp-%d.bin", i)
		if err := readAndDeserialize(filePath, fileName, &compiledLPP, &readBuf); err != nil {
			return nil, err
		}
		assets.DistWizard.CompiledLPPs = append(assets.DistWizard.CompiledLPPs, compiledLPP)
	}

	// Read and deserialize conglomeration compilation
	if err := readAndDeserialize(filePath, "dw-compiled-conglomeration.bin", &assets.DistWizard.CompiledConglomeration, &readBuf); err != nil {
		return nil, err
	}

	return assets, nil
}

func serializeAndWriteRecursionWitness(cfg *config.Config, witnessName string, witness *recursion.Witness, isLPP bool) error {
	reader := bytes.NewReader(nil)
	filePath := cfg.PathforLimitlessProverAssets()
	filePath = path.Join(filePath, "witness")
	if isLPP {
		filePath = path.Join(filePath, "lpp")
	} else {
		filePath = path.Join(filePath, "gl")
	}
	return serializeAndWrite(filePath, witnessName, witness, reader)
}

// Helper function to read and deserialize an object from a file
func readAndDeserialize(filePath string, fileName string, target any, readBuf *bytes.Buffer) error {
	readBuf.Reset()
	fullPath := path.Join(filePath, fileName)
	if err := utils.ReadFromFile(fullPath, readBuf); err != nil {
		return fmt.Errorf("failed to read %s: %w", fullPath, err)
	}
	if err := serialization.Deserialize(readBuf.Bytes(), target); err != nil {
		return fmt.Errorf("failed to deserialize %s: %w", fileName, err)
	}
	logrus.Infof("Read and deserialized %s from %s", fileName, fullPath)
	return nil
}
