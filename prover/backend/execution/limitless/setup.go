package limitless

import (
	"sync"

	"github.com/consensys/linea-monorepo/prover/config"
	"github.com/consensys/linea-monorepo/prover/protocol/compiler/logdata"
	"github.com/consensys/linea-monorepo/prover/protocol/distributed"
	"github.com/consensys/linea-monorepo/prover/zkevm"
)

// Singleton instance and synchronization
var (
	setupInstance *Setup
	onceSetup     sync.Once
)

// Setup holds the components initialized during setup
// This will be part of the prover-assets
type Setup struct {
	ZkEvmInstance *zkevm.ZkEvm
	DistWizard    *distributed.DistributedWizard
	CompiledGLs   []*distributed.RecursedSegmentCompilation
	CompiledLPPs  []*distributed.RecursedSegmentCompilation
}

// InitLimitlessSetup initializes the Setup struct (internal use)
func InitLimitlessSetup(cfg *config.Config, targetWeight, minCompilationSize int) *Setup {
	// Do the prover setup only once
	onceSetup.Do(func() {
		// Initialize module discoverer
		disc := &distributed.StandardModuleDiscoverer{
			TargetWeight: targetWeight,
		}

		// Get zkEVM instance
		zkevmInstance := zkevm.FullZkEvm(&cfg.TracesLimits, cfg)

		// Distribute the wizard protocol
		distWizard := distributed.DistributeWizard(zkevmInstance.WizardIOP, disc)

		compiledGLs := make([]*distributed.RecursedSegmentCompilation, len(distWizard.GLs))
		compiledLPPs := make([]*distributed.RecursedSegmentCompilation, len(distWizard.LPPs))

		for i := range distWizard.GLs {
			if cells := logdata.GetWizardStats(distWizard.GLs[i].Wiop); cells.TotalCells() > minCompilationSize {
				compiledGLs[i] = distributed.CompileSegment(distWizard.GLs[i])
			}
		}

		for i := range distWizard.LPPs {
			if cells := logdata.GetWizardStats(distWizard.LPPs[i].Wiop); cells.TotalCells() > minCompilationSize {
				compiledLPPs[i] = distributed.CompileSegment(distWizard.LPPs[i])
			}
		}

		// Initialize the singleton instance
		setupInstance = &Setup{
			ZkEvmInstance: zkevmInstance,
			DistWizard:    &distWizard,
			CompiledGLs:   compiledGLs,
			CompiledLPPs:  compiledLPPs,
		}
	})
	return setupInstance
}

// GetSetup returns the singleton Setup instance, initializing it if necessary
func GetSetup() *Setup {
	return setupInstance
}
