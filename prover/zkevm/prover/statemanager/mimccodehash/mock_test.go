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
		romLexInput = &RomLexInput{}
		mod         Module
	)

	cmp := wizard.Compile(func(build *wizard.Builder) {

		// Define romInput
		romInput = &RomInput{
			NBytes:  ctRom.GetCommit(build, "NBYTES"),
			Counter: ctRom.GetCommit(build, "COUNTER"),
		}

		for i := range common.NbLimbU128 {
			romInput.Acc[i] = ctRom.GetCommit(build, fmt.Sprintf("ACC_%d", i))
		}

		for i := range common.NbLimbU32 {
			romInput.CFI[i] = ctRom.GetCommit(build, fmt.Sprintf("CFI_%d", i))
			romInput.CodeSize[i] = ctRom.GetCommit(build, fmt.Sprintf("CODESIZE_%d", i))
		}

		// Define romLexInput
		for i := range common.NbLimbU256 {
			romLexInput.CodeHash[i] = ctRomLex.GetCommit(build, fmt.Sprintf("CODEHASH_%d", i))
		}

		for i := range common.NbLimbU32 {
			romLexInput.CFIRomLex[i] = ctRomLex.GetCommit(build, fmt.Sprintf("CFI_ROMLEX_%d", i))
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

		accNames := make([]string, len(romInput.Acc))
		for i := range romInput.Acc {
			accNames[i] = string(romInput.Acc[i].GetColID())
		}

		var ctRomCols = []string{"CFI_0", "CFI_1"}
		ctRomCols = append(ctRomCols, accNames[:]...)
		ctRomCols = append(ctRomCols, "NBYTES", "COUNTER")
		ctRomCols = append(ctRomCols, codeSizeNames[:]...)

		ctRom.Assign(run, ctRomCols[:]...)
		romInput.completeAssign(run)
		ctRomLex.Assign(run,
			append([]string{"CFI_ROMLEX_0", "CFI_ROMLEX_1"},
				codeHashNames...)...)
		mod.Assign(run)

		var ctRomColIds = []string{string(romInput.CFI[0].GetColID()), string(romInput.CFI[1].GetColID())}
		ctRomColIds = append(ctRomColIds, accNames[:]...)
		ctRomColIds = append(ctRomColIds, string(romInput.NBytes.GetColID()))
		ctRomColIds = append(ctRomColIds, string(romInput.Counter.GetColID()))
		ctRomColIds = append(ctRomColIds, codeSizeNames[:]...)

		ctRom.CheckAssignment(run,
			// TODO: add also auxiliary columns
			ctRomColIds[:]...,
		)

		ctRomLex.CheckAssignment(run,
			append(
				[]string{string(romLexInput.CFIRomLex[0].GetColID()), string(romLexInput.CFIRomLex[1].GetColID())},
				codeHashNames[:]...,
			)...,
		)
	})
	if err := wizard.Verify(cmp, proof); err != nil {
		t.Fatal("proof failed", err)
	}
	t.Log("proof succeeded")
}
