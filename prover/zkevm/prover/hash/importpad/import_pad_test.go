package importpad

import (
	"encoding/hex"
	"testing"

	"github.com/consensys/linea-monorepo/prover/crypto/keccak"
	"github.com/consensys/linea-monorepo/prover/crypto/sha2"
	"github.com/consensys/linea-monorepo/prover/protocol/compiler/dummy"
	"github.com/consensys/linea-monorepo/prover/protocol/wizard"
	"github.com/consensys/linea-monorepo/prover/utils/csvtraces"
	"github.com/consensys/linea-monorepo/prover/zkevm/prover/hash/generic"
	"github.com/stretchr/testify/assert"
)

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

func TestImportAndPad(t *testing.T) {

	for _, uc := range testCases {
		t.Run(uc.Name, func(t *testing.T) {

			var (
				inp   ImportAndPadInputs
				mod   *importation
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
						Limb:    inpCt.GetCommit(build, "LIMBS"),
					}},
					PaddingStrategy: uc.UseCase,
				}

				mod = ImportAndPad(build.CompiledIOP, inp, 64)

			}, dummy.Compile)

			proof := wizard.Prove(comp, func(run *wizard.ProverRuntime) {

				inpCt.Assign(run,
					"HASH_NUM",
					"INDEX",
					"TO_HASH",
					"NBYTES",
					"LIMBS",
				)

				mod.Run(run)

				modCt.CheckAssignment(run,
					string(mod.HashNum.GetColID()),
					string(mod.Index.GetColID()),
					string(mod.IsActive.GetColID()),
					string(mod.IsInserted.GetColID()),
					string(mod.IsPadded.GetColID()),
					string(mod.Limbs.GetColID()),
					string(mod.NBytes.GetColID()),
					string(mod.AccPaddedBytes.GetColID()),
				)

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

func checkPaddingAssignment(t *testing.T, run *wizard.ProverRuntime, paddingFunc func([]byte) []byte, mod *importation) {

	var (
		paddedGdm = &generic.GenDataModule{
			HashNum: mod.HashNum,
			Limb:    mod.Limbs,
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
