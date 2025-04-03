package limitless

import (
	"errors"
	"math/rand/v2"
	"strings"
	"sync"

	"github.com/consensys/linea-monorepo/prover/backend/execution"
	"github.com/consensys/linea-monorepo/prover/config"
	"github.com/consensys/linea-monorepo/prover/maths/field"
	"github.com/consensys/linea-monorepo/prover/protocol/distributed"
	"github.com/consensys/linea-monorepo/prover/protocol/wizard"
	"github.com/consensys/linea-monorepo/prover/utils"
)

func MainProve(cfg *config.Config, req *execution.Request) (resp *execution.Response) {

	var (
		// Retrieve the setup first
		setup = GetSetup()

		// Retireve the singleton metadata instance
		distMetadata = GetDistMetada()

		// Extract the block range
		// For ex: 16302323-16302336.conflated.beta-v2.1-rc1_10.40.41.87_25df.lt" => 16302323-16302336
		blockRange = strings.Split(req.ConflatedExecutionTracesFile, ".")[0]
	)

	// Run the bootstrapper to obtain the GL/LPP witness traces
	bootstrapRes, err := setup.runBootstrapper(cfg, distMetadata, req, blockRange)
	if err != nil {
		utils.Panic(err.Error())
	}

	var (
		glProofs  = make([]wizard.Proof, len(bootstrapRes.witnessGLs))
		lppProofs = make([]wizard.Proof, len(bootstrapRes.witnessLPPs))
	)

	// Spin up GL-workers in parallel
	var wg sync.WaitGroup
	for i := range bootstrapRes.witnessGLs {
		wg.Add(1)
		go func(i int) {
			defer func() {
				distMetadata.SetKeyTrue(blockRange, string(bootstrapRes.witnessGLs[i].ModuleName),
					bootstrapRes.witnessGLs[i].ModuleIndex, i)
				wg.Done()
			}()
			glProofs[i] = setup.runGLSubProver(bootstrapRes.witnessGLs[i])
		}(i)
	}

	wg.Wait()

	// Mocked Randomness
	sharedRandomness := runMockRdnBeacon()

	// Spin up LPP-workers in parallel
	for i := range bootstrapRes.witnessLPPs {
		wg.Add(1)
		bootstrapRes.witnessLPPs[i].InitialFiatShamirState = sharedRandomness
		go func(i int) {
			defer func() {
				distMetadata.SetKeyTrue(blockRange, string(bootstrapRes.witnessLPPs[i].ModuleName),
					bootstrapRes.witnessLPPs[i].ModuleIndex, i)
				wg.Done()
			}()
			lppProofs[i] = setup.runLPPSubProver(bootstrapRes.witnessLPPs[i])
		}(i)
	}

	wg.Wait()

	if distMetadata.AreAllTrueForReqID(blockRange) {
		distMetadata.DeleteByReqID(blockRange)
		return runMockedConglomeration(glProofs, lppProofs)
	}

	return nil
}

type BootstrapRes struct {
	// To be compressed
	witnessGLs  []*distributed.ModuleWitness
	witnessLPPs []*distributed.ModuleWitness
}

func (setup *Setup) runBootstrapper(cfg *config.Config, distMetadata *DistMetadata,
	req *execution.Request, blockRange string) (*BootstrapRes, error) {

	// Generate zkEVM witness from request and config
	out := execution.CraftProverOutput(cfg, req)
	witness := execution.NewWitness(cfg, req, &out)
	zkWitness := witness.ZkEVM
	if zkWitness == nil {
		return nil, errors.New("failed to generate zkEVM witness")
	}

	// Run the bootstrapper
	runtimeBoot := wizard.RunProver(setup.DistWizard.Bootstrapper, setup.ZkEvmInstance.GetMainProverStep(zkWitness))
	if runtimeBoot == nil {
		return nil, errors.New("bootstrapper prover failed")
	}

	// Extract and verify the proof
	proof := runtimeBoot.ExtractProof()
	if err := wizard.Verify(setup.DistWizard.Bootstrapper, proof); err != nil {
		return nil, errors.New("bootstrapper verification failed: " + err.Error())
	}

	witnessGLs, witnessLPPs := distributed.SegmentRuntime(runtimeBoot, setup.DistWizard)

	// Update sub-reqs for GL sub-modules in metadata
	for i := range witnessGLs {
		distMetadata.AddKey(blockRange, string(witnessGLs[i].ModuleName), witnessGLs[i].ModuleIndex, i)
	}

	// Update sub-reqs for LPP sub-modules in metadata
	for i := range witnessLPPs {
		distMetadata.AddKey(blockRange, string(witnessLPPs[i].ModuleName), witnessLPPs[i].ModuleIndex, i)
	}

	return &BootstrapRes{
		witnessGLs:  witnessGLs,
		witnessLPPs: witnessLPPs,
	}, nil
}

func (setup *Setup) runGLSubProver(witnessGL *distributed.ModuleWitness) wizard.Proof {

	for k := range setup.DistWizard.ModuleNames {
		if setup.DistWizard.ModuleNames[k] != witnessGL.ModuleName {
			utils.Panic("GL-submodule name in wizard does not match with the corresponding witness")
		}
	}

	// TODO: Assuming valid index here
	moduleGL := setup.CompiledGLs[witnessGL.ModuleIndex]
	glSubProof := moduleGL.ProveSegment(witnessGL)
	return glSubProof
}

func (setup *Setup) runLPPSubProver(witnessLPP *distributed.ModuleWitness) wizard.Proof {

	for k := range setup.DistWizard.ModuleNames {
		if setup.DistWizard.ModuleNames[k] != witnessLPP.ModuleName {
			utils.Panic("LPP-submodule name in wizard does not match with the corresponding witness")
		}
	}

	// TODO: Assuming valid index here
	moduleGL := setup.CompiledLPPs[witnessLPP.ModuleIndex]
	lppSubProof := moduleGL.ProveSegment(witnessLPP)
	return lppSubProof
}

func runMockRdnBeacon() field.Element {

	var (
		// #nosec G404 --we don't need a cryptographic RNG for testing purpose
		rng              = rand.New(utils.NewRandSource(0))
		sharedRandomness = field.PseudoRand(rng)
	)

	return sharedRandomness
}

func runMockedConglomeration(glProofs, lppProofs []wizard.Proof) *execution.Response {
	return &execution.Response{}
}
