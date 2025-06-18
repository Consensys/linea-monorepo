package limitless

import (
	"bytes"
	"fmt"
	"path"

	"github.com/consensys/linea-monorepo/prover/config"
	"github.com/consensys/linea-monorepo/prover/protocol/compiler/recursion"
	"github.com/consensys/linea-monorepo/prover/protocol/distributed"
	"github.com/consensys/linea-monorepo/prover/protocol/serialization"
	"github.com/consensys/linea-monorepo/prover/protocol/wizard"
	"github.com/consensys/linea-monorepo/prover/utils"
	"github.com/sirupsen/logrus"
)

func RunConglomerationProver(cfg *config.Config, cong *distributed.ConglomeratorCompilation,
	witnessGLs []*distributed.ModuleWitnessGL, witnessLPPs []*distributed.ModuleWitnessLPP,
) (*wizard.Proof, error) {

	var (
		filePath     = path.Join(cfg.PathforLimitlessProverAssets(), "witness")
		recurWitGLs  = make([]recursion.Witness, len(witnessGLs))
		recurWitLPPs = make([]recursion.Witness, len(witnessLPPs))
	)

	var readBuf *bytes.Buffer
	for i, witnessLPP := range witnessLPPs {
		var recurWitLPP *recursion.Witness
		filePath = path.Join(filePath, "lpp")
		fileName := fmt.Sprintf("%v-%v-%v", i, witnessLPP.ModuleName, witnessLPP.ModuleIndex)
		err := serialization.ReadAndDeserialize(filePath, fileName, &recurWitLPP, readBuf)
		if err != nil {
			return nil, fmt.Errorf("failed to read and deserialize LPP-recursion witness: %w", err)
		}
		recurWitLPPs[i] = *recurWitLPP
	}

	for i, witnessGL := range witnessGLs {
		var recurWitGL *recursion.Witness
		filePath = path.Join(filePath, "gl")
		fileName := fmt.Sprintf("%v-%v-%v", i, witnessGL.ModuleName, witnessGL.ModuleIndex)
		err := serialization.ReadAndDeserialize(filePath, fileName, &recurWitGL, readBuf)
		if err != nil {
			return nil, fmt.Errorf("failed to read and deserialize GL-recursion witness: %w", err)
		}
		recurWitGLs[i] = *recurWitGL
	}

	logrus.Info("Starting to prove conglomerator")
	proof := cong.Prove(recurWitGLs, recurWitLPPs)
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
