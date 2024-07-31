package execdatahashing

import (
	"testing"

	"github.com/consensys/zkevm-monorepo/prover/crypto/mimc"
	"github.com/consensys/zkevm-monorepo/prover/maths/field"
	"github.com/consensys/zkevm-monorepo/prover/protocol/compiler/dummy"
	"github.com/consensys/zkevm-monorepo/prover/protocol/wizard"
	"github.com/consensys/zkevm-monorepo/prover/utils/csvtraces"
	"github.com/stretchr/testify/assert"
)

func TestCorrectness(t *testing.T) {

	var (
		inp   *ExecDataHashingInput
		mod   *execData
		inpCt = csvtraces.MustOpenCsvFile("testdata/input.csv")
	)

	comp := wizard.Compile(func(build *wizard.Builder) {
		inp = &ExecDataHashingInput{
			Data:     inpCt.GetCommit(build, "DATA"),
			Selector: inpCt.GetCommit(build, "SELECTOR"),
		}

		mod = HashExecutionData(build.CompiledIOP, inp)
	}, dummy.Compile)

	proof := wizard.Prove(comp, func(run *wizard.ProverRuntime) {
		inpCt.Assign(run, "DATA", "SELECTOR")
		mod.Run(run)

		var (
			fHash  = manuallyComputeFinalHash(run, inp)
			fHash2 = run.GetLocalPointEvalParams(mod.FinalHashOpening.ID).Y
		)

		assert.Equalf(t, fHash.String(), fHash2.String(), "the hash computed by the module is not the correct one")
	})

	if err := wizard.Verify(comp, proof); err != nil {
		t.Fatal("proof failed", err)
	}

	t.Log("proof succeeded")

}

func manuallyComputeFinalHash(
	run *wizard.ProverRuntime,
	inp *ExecDataHashingInput,
) field.Element {

	var (
		data         = inp.Data.GetColAssignment(run).IntoRegVecSaveAlloc()
		selector     = inp.Selector.GetColAssignment(run).IntoRegVecSaveAlloc()
		actualInputs = make([]field.Element, 0)
	)

	for row := range selector {
		if selector[row].IsOne() {
			actualInputs = append(actualInputs, data[row])
		}
	}

	return mimc.HashVec(actualInputs)
}
