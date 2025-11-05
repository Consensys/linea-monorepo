package importpad

import (
	"encoding/hex"
	"testing"

	"fmt"
	"github.com/consensys/linea-monorepo/prover/crypto/keccak"
	"github.com/consensys/linea-monorepo/prover/crypto/sha2"
	"github.com/consensys/linea-monorepo/prover/protocol/compiler/dummy"
	"github.com/consensys/linea-monorepo/prover/protocol/ifaces"
	"github.com/consensys/linea-monorepo/prover/protocol/wizard"
	"github.com/consensys/linea-monorepo/prover/utils/csvtraces"
	"github.com/consensys/linea-monorepo/prover/zkevm/prover/common"
	"github.com/consensys/linea-monorepo/prover/zkevm/prover/hash/generic"
	"github.com/stretchr/testify/assert"
)

func TestImportAndPad(t *testing.T) {
	var testCases = []struct {
		Name        string
		ModFilePath string
		UseCase     generic.HashingUsecase
		PaddingFunc func(stream []byte) []byte
	}{
		{
			Name:        "Keccak",
			ModFilePath: "testdata/mod_keccak.csv",
			UseCase:     generic.KeccakUsecase,
			PaddingFunc: keccak.PadStream,
		},
		{
			Name:        "Sha2",
			ModFilePath: "testdata/mod_sha2.csv",
			UseCase:     generic.Sha2Usecase,
			PaddingFunc: sha2.PadStream,
		},
		{
			Name:        "MiMC",
			ModFilePath: "testdata/mod_mimc.csv",
			UseCase:     generic.MiMCUsecase,
		},
	}

	for _, uc := range testCases {
		t.Run(uc.Name, func(t *testing.T) {

			var (
				inp   ImportAndPadInputs
				mod   *Importation
				inpCt = csvtraces.MustOpenCsvFile("testdata/input.csv")
				modCt = csvtraces.MustOpenCsvFile(uc.ModFilePath)
			)

			comp := wizard.Compile(func(build *wizard.Builder) {

				inp = ImportAndPadInputs{
					Name: "TESTING",
					Src: generic.GenericByteModule{Data: generic.GenDataModule{
						HashNum: inpCt.GetCommit(build, "HASH_NUM"),
						Index:   inpCt.GetCommit(build, "INDEX"),
						ToHash:  inpCt.GetCommit(build, "TO_HASH"),
						NBytes:  inpCt.GetCommit(build, "NBYTES"),
						Limbs:   make([]ifaces.Column, common.NbLimbU128),
					}},
					PaddingStrategy: uc.UseCase,
				}

				for i := 0; i < common.NbLimbU128; i++ {
					inp.Src.Data.Limbs[i] = inpCt.GetCommit(build, fmt.Sprintf("LIMB_%d", i))
				}

				mod = ImportAndPad(build.CompiledIOP, inp, 64)

			}, dummy.Compile)

			proof := wizard.Prove(comp, func(run *wizard.ProverRuntime) {

				assign := []string{"HASH_NUM", "INDEX", "TO_HASH", "NBYTES"}
				for i := 0; i < len(inp.Src.Data.Limbs); i++ {
					assign = append(assign, fmt.Sprintf("LIMB_%d", i))
				}

				inpCt.Assign(run, assign...)

				mod.Run(run)

				toCheck := []string{
					string(mod.HashNum.GetColID()),
					string(mod.Index.GetColID()),
					string(mod.IsActive.GetColID()),
					string(mod.IsInserted.GetColID()),
					string(mod.IsPadded.GetColID()),
					string(mod.NBytes.GetColID()),
					string(mod.AccPaddedBytes.GetColID()),
				}
				for i := range mod.Limbs {
					toCheck = append(toCheck, string(mod.Limbs[i].GetColID()))
				}

				modCt.CheckAssignment(run, toCheck...)

				if uc.PaddingFunc != nil {
					checkPaddingAssignment(t, run, uc.PaddingFunc, mod)
				}
			})

			if err := wizard.Verify(comp, proof); err != nil {
				t.Fatal("proof failed", err)
			}

			t.Log("proof succeeded")

		})
	}
}

func checkPaddingAssignment(t *testing.T, run *wizard.ProverRuntime, paddingFunc func([]byte) []byte, mod *Importation) {

	var (
		paddedGdm = &generic.GenDataModule{
			HashNum: mod.HashNum,
			Limbs:   mod.Limbs,
			NBytes:  mod.NBytes,
			ToHash:  mod.IsActive,
			Index:   mod.Index,
		}

		inputStreams        = mod.Inputs.Src.Data.ScanStreams(run)
		actualPaddedStreams = paddedGdm.ScanStreams(run)
	)

	for i := range inputStreams {
		expectedPaddedStream := paddingFunc(inputStreams[i])
		assert.Equal(t,
			hex.EncodeToString(expectedPaddedStream),
			hex.EncodeToString(actualPaddedStreams[i]),
		)
	}
}
