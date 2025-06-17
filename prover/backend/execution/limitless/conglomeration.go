package limitless

import (
	"fmt"

	"github.com/consensys/linea-monorepo/prover/protocol/compiler/recursion"
	"github.com/consensys/linea-monorepo/prover/protocol/distributed"
	"github.com/consensys/linea-monorepo/prover/protocol/wizard"
	"github.com/consensys/linea-monorepo/prover/utils"
	"github.com/sirupsen/logrus"
)

func RunConglomerationProver(cong *distributed.ConglomeratorCompilation, runGLs, runLPPs []*wizard.ProverRuntime) (*wizard.Proof, error) {
	var (
		witLPPs = make([]recursion.Witness, len(runLPPs))
		witGLs  = make([]recursion.Witness, len(runGLs))
	)

	for i := range runLPPs {
		witLPPs[i] = recursion.ExtractWitness(runLPPs[i])
	}

	for i := range runGLs {
		witGLs[i] = recursion.ExtractWitness(runGLs[i])
	}

	logrus.Info("Starting to prove conglomerator")
	proof := cong.Prove(witGLs, witLPPs)
	logrus.Info("Finished proving conglomerator")

	logrus.Info("Start sanity-checking conglomeration proof")
	err := wizard.Verify(cong.Wiop, proof)
	if err != nil {
		return nil, fmt.Errorf("could not verify conglomeration proof: %v", err)
	}
	logrus.Info("Finished sanity-checking conglomeration proof")
	return &proof, nil
}

func SanityCheckConglomeration(cong *distributed.ConglomeratorCompilation, run *wizard.ProverRuntime) {
	stopRound := recursion.VortexQueryRound(cong.ModuleGLIops[0])
	err := wizard.VerifyUntilRound(cong.ModuleGLIops[0], run.ExtractProof(), stopRound+1)

	if err != nil {
		utils.Panic("could not verify proof: %v", err)
	}
}
