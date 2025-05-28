package v2

import (
	"github.com/consensys/linea-monorepo/prover/protocol/distributed"
	"github.com/consensys/linea-monorepo/prover/utils/test_utils"
)

var (
	zkEVM      = test_utils.GetZkEVM()
	affinities = test_utils.GetAffinities(zkEVM)
	discoverer = &distributed.StandardModuleDiscoverer{
		TargetWeight: 1 << 28,
		Affinities:   affinities,
		Predivision:  1,
	}
	// Dist. wizard
	dw = distributed.DistributeWizard(zkEVM.WizardIOP, discoverer).CompileSegments()

	// Recur seg compilation
	recurSegComp = dw.CompiledDefault

	testRec       = recurSegComp.Recursion
	testCompInput = testRec.InputCompiledIOP
	testCompRecur = recurSegComp.RecursionComp
)
