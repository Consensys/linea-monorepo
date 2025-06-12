package limitless

import (
	"github.com/consensys/linea-monorepo/prover/protocol/distributed"
	"github.com/consensys/linea-monorepo/prover/protocol/wizard"
	"github.com/consensys/linea-monorepo/prover/utils"
	"github.com/sirupsen/logrus"
)

// runProverGLs executes the prover for each GL module segment. It takes in a list of
// compiled GL segments and corresponding witnesses, then runs the prover for each
// segment. The function logs the start and end times of the prover execution for each
// segment. It returns a slice of ProverRuntime instances, each representing the
// result of the prover execution for a segment.
func RunProverGLs(
	distWizard *distributed.DistributedWizard,
	witnessGLs []*distributed.ModuleWitnessGL,
) []*wizard.ProverRuntime {

	var (
		compiledGLs = distWizard.CompiledGLs
		runs        = make([]*wizard.ProverRuntime, len(witnessGLs))
	)

	for i, witnessGL := range witnessGLs {
		var moduleGL *distributed.RecursedSegmentCompilation

		logrus.Infof("segment(total)=%v module=%v segment.index=%v", i, witnessGL.ModuleName, witnessGL.ModuleIndex)
		for k := range distWizard.ModuleNames {
			if distWizard.ModuleNames[k] != witnessGLs[i].ModuleName {
				continue
			}
			moduleGL = compiledGLs[k]
		}

		if moduleGL == nil {
			utils.Panic("module GL does not exists")
		}

		logrus.Infof("RUNNING THE GL PROVER:")
		runs[i] = RunGLProver(moduleGL, witnessGL)
		logrus.Infof("RUNNING THE GL PROVER - DONE:")
	}
	return runs
}

func RunGLProver(moduleGL *distributed.RecursedSegmentCompilation, witnessGL *distributed.ModuleWitnessGL) *wizard.ProverRuntime {
	return moduleGL.ProveSegment(witnessGL)
}
