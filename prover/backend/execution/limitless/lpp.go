package limitless

import (
	"reflect"

	"github.com/consensys/linea-monorepo/prover/maths/field"
	"github.com/consensys/linea-monorepo/prover/protocol/distributed"
	"github.com/consensys/linea-monorepo/prover/protocol/wizard"
	"github.com/consensys/linea-monorepo/prover/utils"
	"github.com/sirupsen/logrus"
)

func RunProverLPPs(
	distWizard *distributed.DistributedWizard,
	sharedRandomness field.Element,
	witnessLPPs []*distributed.ModuleWitnessLPP,
) []*wizard.ProverRuntime {

	var (
		runs         = make([]*wizard.ProverRuntime, len(witnessLPPs))
		compiledLPPs = distWizard.CompiledLPPs
	)

	for i, witnessLPP := range witnessLPPs {
		var moduleLPP *distributed.RecursedSegmentCompilation
		witnessLPP.InitialFiatShamirState = sharedRandomness
		logrus.Infof("segment(total)=%v module=%v segment.index=%v", i, witnessLPP.ModuleNames, witnessLPP.ModuleIndex)
		for k := range distWizard.LPPs {

			if !reflect.DeepEqual(distWizard.LPPs[k].ModuleNames(), witnessLPPs[i].ModuleNames) {
				continue
			}

			moduleLPP = compiledLPPs[k]
		}

		if moduleLPP == nil {
			utils.Panic("module does not exists")
		}

		logrus.Infof("RUNNING THE LPP PROVER:")
		runs[i] = RunLPPProver(moduleLPP, witnessLPP)
		logrus.Infof("RUNNING THE LPP PROVER - DONE:")
	}
	return runs
}

func RunLPPProver(moduleLPP *distributed.RecursedSegmentCompilation, witnessLPP *distributed.ModuleWitnessLPP) *wizard.ProverRuntime {
	return moduleLPP.ProveSegment(witnessLPP)
}
