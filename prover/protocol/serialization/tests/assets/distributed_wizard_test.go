package assets

import (
	"fmt"
	"testing"

	"github.com/consensys/linea-monorepo/prover/protocol/distributed"
	"github.com/consensys/linea-monorepo/prover/utils/test_utils"
)

// GetDistributed initializes a distributed wizard configuration using the
// ZkEVM's compiled IOP and a StandardModuleDiscoverer with preset parameters.
// The function compiles the necessary segments and produces a conglomerated
// distributed wizard, which is then returned.
func GetDistributed() *distributed.DistributedWizard {

	var (
		zkevm = test_utils.GetZkEVM()
		disc  = &distributed.StandardModuleDiscoverer{
			TargetWeight: 1 << 28,
			Affinities:   test_utils.GetAffinities(zkevm),
			Predivision:  1,
		}

		// This tests the compilation of the compiled-IOP
		distWizard = distributed.DistributeWizard(zkevm.WizardIOP, disc).
				CompileSegments().
				Conglomerate(20)
	)

	return distWizard
}

func TestDistributedWizard(t *testing.T) {
	dist := GetDistributed()

	runSerdeTest(t, dist.Bootstrapper, "DistributedWizard.Bootstrapper")

	runSerdeTest(t, dist.DefaultModule, "DistributedWizard.DefaultModule")

	runSerdeTest(t, dist.Disc, "DistributedWizard.Discoverer")

	for i := range dist.CompiledGLs {
		runSerdeTest(t, dist.CompiledGLs[i], fmt.Sprintf("DistributedWizard.CompiledGL-%v", i))
	}

	for i := range dist.CompiledLPPs {
		runSerdeTest(t, dist.CompiledLPPs[i], fmt.Sprintf("DistributedWizard.CompiledLPP-%v", i))
	}

	runSerdeTest(t, dist.CompiledConglomeration, "DistributedWizard.CompiledConglomeration")
}
