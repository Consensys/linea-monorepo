package limitless

import (
	"fmt"
	"reflect"

	"github.com/consensys/linea-monorepo/prover/config"
	"github.com/consensys/linea-monorepo/prover/maths/field"
	"github.com/consensys/linea-monorepo/prover/protocol/compiler/recursion"
	"github.com/consensys/linea-monorepo/prover/protocol/distributed"
	"github.com/consensys/linea-monorepo/prover/protocol/wizard"
	"github.com/consensys/linea-monorepo/prover/utils"
	"github.com/sirupsen/logrus"
)

func RunProverLPPs(cfg *config.Config,
	distWizard *distributed.DistributedWizard,
	sharedRandomness field.Element,
	witnessLPPs []*distributed.ModuleWitnessLPP,
) ([]*wizard.ProverRuntime, error) {

	var (
		runs         = make([]*wizard.ProverRuntime, len(witnessLPPs))
		compiledLPPs = distWizard.CompiledLPPs
	)

	for i, witnessLPP := range witnessLPPs {
		var moduleLPP *distributed.RecursedSegmentCompilation
		witnessLPP.InitialFiatShamirState = sharedRandomness
		for k := range distWizard.LPPs {

			if !reflect.DeepEqual(distWizard.LPPs[k].ModuleNames(), witnessLPPs[i].ModuleName) {
				continue
			}

			moduleLPP = compiledLPPs[k]
		}

		if moduleLPP == nil {
			utils.Panic("module does not exists")
		}

		logrus.Infof("RUNNING the LPP-prover for segment(total)=%v module=%v segment.index=%v", i, witnessLPP.ModuleName, witnessLPP.ModuleIndex)
		fileName := fmt.Sprintf("%v-%v-%v", i, witnessLPP.ModuleName, witnessLPP.ModuleIndex)
		run, err := RunLPPProver(cfg, moduleLPP, witnessLPP, fileName)
		if err != nil {
			return nil, err
		}
		runs[i] = run
		logrus.Infof("Finished running the LPP-prover")
	}
	return runs, nil
}

func RunLPPProver(cfg *config.Config,
	moduleLPP *distributed.RecursedSegmentCompilation,
	witnessLPP *distributed.ModuleWitnessLPP, fileName string,
) (*wizard.ProverRuntime, error) {

	runtimeLPP := moduleLPP.ProveSegment(witnessLPP)

	logrus.Info("Extracting LPP-recursion witness and writing it to disk")
	recursionLPPWitness := recursion.ExtractWitness(runtimeLPP)

	err := serializeAndWriteRecursionWitness(cfg, fileName, &recursionLPPWitness, true)
	if err != nil {
		return nil, fmt.Errorf("failed to serialize and write LPP-recursion witness: %w", err)
	}
	logrus.Info("Finished extracting LPP-recursion witness and writing it to disk")
	return runtimeLPP, nil
}
