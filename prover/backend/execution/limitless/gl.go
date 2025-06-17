package limitless

import (
	"bytes"
	"path"

	"github.com/consensys/linea-monorepo/prover/config"
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

		logrus.Infof("RUNNING the GL Prover")
		runs[i] = RunGLProver(moduleGL, witnessGL)
		logrus.Infof("FINISHED Running the GL Prover")
	}
	return runs
}

func RunGLProver(moduleGL *distributed.RecursedSegmentCompilation, witnessGL *distributed.ModuleWitnessGL) *wizard.ProverRuntime {
	return moduleGL.ProveSegment(witnessGL)
}

func SerializeAndWriteWitnessGL(cfg *config.Config, witnessGLName string, witnessGL *distributed.ModuleWitnessGL) error {
	reader := bytes.NewReader(nil)
	filePath := cfg.PathforLimitlessProverAssets()
	filePath = path.Join(filePath, "witnesses")
	return serializeAndWrite(filePath, witnessGLName, witnessGL, reader)
}
