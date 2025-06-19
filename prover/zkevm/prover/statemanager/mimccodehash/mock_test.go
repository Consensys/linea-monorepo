package mimccodehash

import (
	"fmt"
	"github.com/consensys/linea-monorepo/prover/zkevm/prover/common"
	"os"
	"testing"

	"github.com/consensys/linea-monorepo/prover/protocol/compiler/dummy"
	"github.com/consensys/linea-monorepo/prover/protocol/wizard"
	"github.com/consensys/linea-monorepo/prover/utils/csvtraces"
)

func TestMiMCCodeHash(t *testing.T) {
	romFile, errRom := os.Open("testdata/rom_input.csv")
	if errRom != nil {
		t.Fatal(errRom)
	}
	romLexFile, errRomLex := os.Open("testdata/romlex_input.csv")
	if errRomLex != nil {
		t.Fatal(errRomLex)
	}
	defer romFile.Close()
	defer romLexFile.Close()

	ctRom, errRom := csvtraces.NewCsvTrace(romFile)
	ctRomLex, errRomLex := csvtraces.NewCsvTrace(romLexFile)
	if errRom != nil {
		t.Fatal(errRom)
	}
	if errRomLex != nil {
		t.Fatal(errRomLex)
	}

	var (
		romInput    *RomInput
		romLexInput *RomLexInput
		mod         Module
	)

	cmp := wizard.Compile(func(build *wizard.Builder) {

		// Define romInput
		romInput = &RomInput{
			CFI:     ctRom.GetCommit(build, "CFI"),
			Acc:     ctRom.GetCommit(build, "ACC"),
			NBytes:  ctRom.GetCommit(build, "NBYTES"),
			Counter: ctRom.GetCommit(build, "COUNTER"),
		}

		for i := range common.NbLimbU32 {
			romInput.CodeSize[i] = ctRom.GetCommit(build, fmt.Sprintf("CODESIZE_%d", i))
		}

		// Define romLexInput
		romLexInput = &RomLexInput{
			CFIRomLex: ctRomLex.GetCommit(build, "CFI_ROMLEX"),
		}

		for i := range common.NbLimbU256 {
			romLexInput.CodeHash[i] = ctRomLex.GetCommit(build, fmt.Sprintf("CODEHASH_%d", i))
		}

		mod = NewModule(
			build.CompiledIOP,
			Inputs{
				Name: "MIMC_CODE_HASH_TEST",
				Size: 1 << 13,
			},
		)

		// Check the consistency of different input connection via projection and lookup queries.
		mod.ConnectToRom(build.CompiledIOP, romInput, romLexInput)
	}, dummy.Compile)

	proof := wizard.Prove(cmp, func(run *wizard.ProverRuntime) {
		codeHashNames := make([]string, len(romLexInput.CodeHash))
		for i := range romLexInput.CodeHash {
			codeHashNames[i] = string(romLexInput.CodeHash[i].GetColID())
		}

		codeSizeNames := make([]string, len(romInput.CodeSize))
		for i := range romInput.CodeSize {
			codeSizeNames[i] = string(romInput.CodeSize[i].GetColID())
		}

		ctRom.Assign(run,
			append(
				[]string{
					"CFI",
					"ACC",
					"NBYTES",
					"COUNTER",
				},
				codeSizeNames[:]...,
			)...,
		)
		romInput.completeAssign(run)
		ctRomLex.Assign(run,
			append([]string{"CFI_ROMLEX"},
				codeHashNames...)...)
		mod.Assign(run)
		ctRom.CheckAssignment(run,
			// TODO: add also auxiliary columns
			append(
				[]string{string(romInput.CFI.GetColID()),
					string(romInput.Acc.GetColID()),
					string(romInput.NBytes.GetColID()),
					string(romInput.Counter.GetColID())},
				codeSizeNames[:]...,
			)...,
		)

		ctRomLex.CheckAssignment(run,
			append(
				[]string{string(romLexInput.CFIRomLex.GetColID())},
				codeHashNames[:]...,
			)...,
		)
	})
	if err := wizard.Verify(cmp, proof); err != nil {
		t.Fatal("proof failed", err)
	}
	t.Log("proof succeeded")
}
