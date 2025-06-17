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
		logrus.Infof("segment(total)=%v module=%v segment.index=%v", i, witnessLPP.ModuleName, witnessLPP.ModuleIndex)
		for k := range distWizard.LPPs {

			if !reflect.DeepEqual(distWizard.LPPs[k].ModuleNames(), witnessLPPs[i].ModuleName) {
				continue
			}

			moduleLPP = compiledLPPs[k]
		}

		if moduleLPP == nil {
			utils.Panic("module does not exists")
		}

		modSegID := &ModSegID{
			SegmentID:  i,
			ModuleName: string(witnessLPP.ModuleName),
			ModuleIdx:  witnessLPP.ModuleIndex,
		}

		run, err := modSegID.RunLPPProver(cfg, moduleLPP, witnessLPP)
		if err != nil {
			return nil, err
		}
		runs[i] = run
	}
	return runs, nil
}

func (ms *ModSegID) RunLPPProver(cfg *config.Config,
	moduleLPP *distributed.RecursedSegmentCompilation,
	witnessLPP *distributed.ModuleWitnessLPP) (*wizard.ProverRuntime, error) {

	logrus.Infof("RUNNING the LPP prover for segment(total)=%v module=%v segment.index=%v", ms.SegmentID, ms.ModuleName, ms.ModuleIdx)
	runtimeLPP := moduleLPP.ProveSegment(witnessLPP)
	logrus.Infof("Finished running the LPP prover")

	logrus.Infof("Extracting LPP-recursion witness and writing it to disk")
	recursionLPPWitness := recursion.ExtractWitness(runtimeLPP)
	moduleName := fmt.Sprintf("%v-%v-%v", ms.SegmentID, ms.ModuleName, ms.ModuleIdx)

	err := SerializeAndWriteRecursionWitness(cfg, moduleName, &recursionLPPWitness, true)
	if err != nil {
		return nil, fmt.Errorf("failed to serialize and write LPP-recursion witness: %w", err)
	}
	return runtimeLPP, nil
}
