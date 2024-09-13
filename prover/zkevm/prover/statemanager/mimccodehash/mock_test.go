package mimccodehash

import (
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
			CFI:      ctRom.GetCommit(build, "CFI"),
			Acc:      ctRom.GetCommit(build, "ACC"),
			NBytes:   ctRom.GetCommit(build, "NBYTES"),
			Counter:  ctRom.GetCommit(build, "COUNTER"),
			CodeSize: ctRom.GetCommit(build, "CODESIZE"),
		}

		// Define romLexInput
		romLexInput = &RomLexInput{
			CFIRomLex:  ctRomLex.GetCommit(build, "CFI_ROMLEX"),
			CodeHashHi: ctRomLex.GetCommit(build, "CODEHASH_HI"),
			CodeHashLo: ctRomLex.GetCommit(build, "CODEHASH_LO"),
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
		ctRom.Assign(run,
			"CFI",
			"ACC",
			"NBYTES",
			"COUNTER",
			"CODESIZE")
		romInput.completeAssign(run)
		ctRomLex.Assign(run,
			"CFI_ROMLEX",
			"CODEHASH_HI",
			"CODEHASH_LO")
		mod.Assign(run)
		ctRom.CheckAssignment(run,
			// TODO: add also auxiliary columns
			string(romInput.CFI.GetColID()),
			string(romInput.Acc.GetColID()),
			string(romInput.NBytes.GetColID()),
			string(romInput.Counter.GetColID()),
			string(romInput.CodeSize.GetColID()),
		)
		ctRomLex.CheckAssignment(run,
			string(romLexInput.CFIRomLex.GetColID()),
			string(romLexInput.CodeHashHi.GetColID()),
			string(romLexInput.CodeHashLo.GetColID()),
		)
	})
	if err := wizard.Verify(cmp, proof); err != nil {
		t.Fatal("proof failed", err)
	}
	t.Log("proof succeeded")
}
