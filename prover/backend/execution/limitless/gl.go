package limitless

import (
	"fmt"

	"github.com/consensys/linea-monorepo/prover/config"
	"github.com/consensys/linea-monorepo/prover/protocol/compiler/recursion"
	"github.com/consensys/linea-monorepo/prover/protocol/distributed"
	"github.com/consensys/linea-monorepo/prover/protocol/wizard"
	"github.com/consensys/linea-monorepo/prover/utils"
	"github.com/sirupsen/logrus"
)

// RunProverGLs executes the prover for each GL module segment. It takes in a list of
// compiled GL segments and corresponding witnesses, then runs the prover for each
// segment. The function logs the start and end times of the prover execution for each
// segment. It returns a slice of ProverRuntime instances, each representing the
// result of the prover execution for a segment.
func RunProverGLs(cfg *config.Config,
	distWizard *distributed.DistributedWizard,
	witnessGLs []*distributed.ModuleWitnessGL,
) ([]*wizard.ProverRuntime, error) {

	var (
		compiledGLs = distWizard.CompiledGLs
		runs        = make([]*wizard.ProverRuntime, len(witnessGLs))
	)

	for i, witnessGL := range witnessGLs {
		var moduleGL *distributed.RecursedSegmentCompilation

		// The wording is not 100% consistent and well-defined. Here, module denotes an horizontal split.
		// And module.index indicates the ID of the segment as part of a particular module.
		// For ex: it looks something like this:

		// | SegmentID   | Module        | Module.index  |
		// | 0           | A             | 0             |
		// | 1           | A             | 1             |
		// | 2           | A             | 2             |
		// | 3           | A             | 3             |
		// | 4           | A             | 4             |
		// | 5           | B             | 0             |
		// | 6           | B             | 1             |
		// | 7           | B             | 2             |
		// | 8           | B             | 3             |
		// | 9           | B             | 4             |
		// | 10          | C             | 0             |
		// | 11          | C             | 1             |
		// | 12          | C             | 2             |
		// | 13          | C             | 3             |

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

		modSegID := &ModSegID{
			SegmentID:  i,
			ModuleName: string(witnessGL.ModuleName),
			ModuleIdx:  witnessGL.ModuleIndex,
		}

		run, err := modSegID.RunGLProver(cfg, moduleGL, witnessGL)
		if err != nil {
			return nil, err
		}
		runs[i] = run

	}
	return runs, nil
}

// RunGLProver executes the prover for a single GL module segment and writes the `recursion.Witness`
// needed for Conglomeration to a file.
func (ms *ModSegID) RunGLProver(cfg *config.Config,
	moduleGL *distributed.RecursedSegmentCompilation,
	witnessGL *distributed.ModuleWitnessGL) (*wizard.ProverRuntime, error) {

	logrus.Infof("RUNNING the GL prover for segment(total)=%v module=%v segment.index=%v", ms.SegmentID, ms.ModuleName, ms.ModuleIdx)
	runtimeGL := moduleGL.ProveSegment(witnessGL)
	logrus.Infof("Finished running the GL prover")

	logrus.Infof("Extracting GL-recursion witness and writing it to disk")
	recursionGLWitness := recursion.ExtractWitness(runtimeGL)
	moduleName := fmt.Sprintf("%v-%v-%v", ms.SegmentID, ms.ModuleName, ms.ModuleIdx)

	err := SerializeAndWriteRecursionWitness(cfg, moduleName, &recursionGLWitness, false)
	if err != nil {
		return nil, fmt.Errorf("failed to serialize and write GL-recursion witness: %w", err)
	}
	return runtimeGL, nil
}
