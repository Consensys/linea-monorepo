package limitless

import (
	"errors"
	"fmt"

	"github.com/consensys/linea-monorepo/prover/backend/execution"
	"github.com/consensys/linea-monorepo/prover/config"
	"github.com/consensys/linea-monorepo/prover/protocol/compiler/horner"
	"github.com/consensys/linea-monorepo/prover/protocol/compiler/logderivativesum"
	"github.com/consensys/linea-monorepo/prover/protocol/compiler/mimc"
	"github.com/consensys/linea-monorepo/prover/protocol/compiler/permutation"
	"github.com/consensys/linea-monorepo/prover/protocol/distributed"
	"github.com/consensys/linea-monorepo/prover/protocol/wizard"
	"github.com/consensys/linea-monorepo/prover/zkevm"
)

// One-per-Segment
type GLSubModReq struct {
	// Unique per every sub request
	ReqID string

	ModuleID  string
	SegmentID int

	// Addr. of the main instance from which the worker nodes need to
	// fetch the traces and then send the GL/LPP sub-response to
	MainAddr string
}

type Segment struct {
	SegmentID int // ModuleIndex
	GlobalID  int // i from witnessGLs
}

type ModuleSegmentMap map[distributed.ModuleName][]Segment
type DistMetadata struct {
	Registry ModuleSegmentMap
}

// GLSubModReq slice will be dispatched to the queue ideally
func InitBootstrapper(cfg *config.Config, req *execution.Request, targetWeight int) (
	[]GLSubModReq, *DistMetadata, error) {

	// Initialize module discoverer
	disc := &distributed.StandardModuleDiscoverer{
		TargetWeight: targetWeight,
	}

	// Get zkEVM instance
	zkevmInstance := zkevm.FullZkEvm(&cfg.TracesLimits, cfg)

	// Distribute the wizard protocol
	distWizard := distributed.DistributeWizard(zkevmInstance.WizardIOP, disc)
	distWizard.CompileModules(
		mimc.CompileMiMC,
		logderivativesum.LookupIntoLogDerivativeSumWithSegmenter(disc),
		permutation.CompileIntoGdProduct,
		horner.ProjectionToHorner,
	)

	// Generate zkEVM witness from request and config
	out := execution.CraftProverOutput(cfg, req)
	witness := execution.NewWitness(cfg, req, &out)
	zkWitness := witness.ZkEVM
	if zkWitness == nil {
		return nil, nil, errors.New("failed to generate zkEVM witness")
	}

	// Run the bootstrapper
	runtimeBoot := wizard.RunProver(distWizard.Bootstrapper, zkevmInstance.GetMainProverStep(zkWitness))
	if runtimeBoot == nil {
		return nil, nil, errors.New("bootstrapper prover failed")
	}

	// Extract and verify the proof
	proof := runtimeBoot.ExtractProof()
	if err := wizard.Verify(distWizard.Bootstrapper, proof); err != nil {
		return nil, nil, errors.New("bootstrapper verification failed: " + err.Error())
	}

	// Segment the runtime into GL and LPP witnesses
	witnessGLs, _ := distributed.SegmentRuntime(runtimeBoot, &distWizard)

	// Build ModuleSegmentMap
	moduleSegmentMap := buildModuleSegmentMap(witnessGLs)

	// Create GLSubModReq for each GL segment
	var glSubModReqs []GLSubModReq
	for i, witnessGL := range witnessGLs {
		// TODO: Simple unique ID to start with later will be replaced by guaranteed unique ID mechanisms
		reqID := fmt.Sprintf("sub-req-GL-%d", i)
		glSubModReqs = append(glSubModReqs, GLSubModReq{
			ReqID: reqID,

			ModuleID:  string(witnessGL.ModuleName),
			SegmentID: witnessGL.ModuleIndex,

			// Will be embedded in the config file later
			// Initially worker and main instances are the same
			MainAddr: "127.0.0.1",
		})
	}

	// Create DistMetadata
	distMetadata := &DistMetadata{
		Registry: moduleSegmentMap,
	}

	return glSubModReqs, distMetadata, nil
}

// buildModuleSegmentMap constructs the ModuleSegmentMap from the list of GL witnesses
func buildModuleSegmentMap(witnessGLs []*distributed.ModuleWitness) ModuleSegmentMap {
	m := make(ModuleSegmentMap)
	for i, witnessGL := range witnessGLs {
		module := witnessGL.ModuleName
		m[module] = append(m[module], Segment{
			SegmentID: witnessGL.ModuleIndex,
			GlobalID:  i,
		})
	}
	return m
}
