package modexp

import (
	"fmt"
	"github.com/consensys/linea-monorepo/prover/protocol/ifaces"
	"testing"

	"github.com/consensys/linea-monorepo/prover/protocol/compiler/dummy"
	"github.com/consensys/linea-monorepo/prover/protocol/wizard"
	"github.com/consensys/linea-monorepo/prover/utils/csvtraces"
	"github.com/consensys/linea-monorepo/prover/zkevm/prover/common"
)

func TestModExpAntichamber(t *testing.T) {

	testCases := []struct {
		InputFName, ModuleFName string
	}{
		{
			InputFName:  "testdata/single_256_bits_input.csv",
			ModuleFName: "testdata/single_256_bits_module.csv",
		},
		{
			InputFName:  "testdata/single_4096_bits_input.csv",
			ModuleFName: "testdata/single_4096_bits_module.csv",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.InputFName, func(t *testing.T) {

			var (
				inp   Input
				mod   *Module
				inpCt = csvtraces.MustOpenCsvFile(tc.InputFName)
				modCt = csvtraces.MustOpenCsvFile(tc.ModuleFName)
			)

			cmp := wizard.Compile(func(build *wizard.Builder) {
				var limbs [common.NbLimbU128]ifaces.Column
				for i := 0; i < common.NbLimbU128; i++ {
					limbs[i] = inpCt.GetCommit(build, fmt.Sprintf("LIMBS_%d", i))
				}

				inp = Input{
					IsModExpBase:     inpCt.GetCommit(build, "IS_MODEXP_BASE"),
					IsModExpExponent: inpCt.GetCommit(build, "IS_MODEXP_EXPONENT"),
					IsModExpModulus:  inpCt.GetCommit(build, "IS_MODEXP_MODULUS"),
					IsModExpResult:   inpCt.GetCommit(build, "IS_MODEXP_RESULT"),
					Limbs:            limbs,
					Settings:         Settings{MaxNbInstance256: 1, MaxNbInstance4096: 1},
				}

				mod = newModule(build.CompiledIOP, inp)
			}, dummy.Compile)

			proof := wizard.Prove(cmp, func(run *wizard.ProverRuntime) {

				var names []string
				for i := 0; i < common.NbLimbU128; i++ {
					names = append(names, fmt.Sprintf("LIMBS_%d", i))
				}

				names = append(names, "IS_MODEXP_BASE", "IS_MODEXP_EXPONENT", "IS_MODEXP_MODULUS", "IS_MODEXP_RESULT")

				inpCt.Assign(run, names...)
				mod.Assign(run)

				var moduleNames []string
				for i := 0; i < common.NbLimbU128; i++ {
					moduleNames = append(moduleNames, fmt.Sprintf("MODEXP_LIMBS_%d", i))
				}

				names = append(names, "MODEXP_IS_ACTIVE", "MODEXP_IS_SMALL", "MODEXP_IS_LARGE", "MODEXP_TO_SMALL_CIRC")

				modCt.CheckAssignment(run, moduleNames...)
			})

			if err := wizard.Verify(cmp, proof); err != nil {
				t.Fatal("proof failed", err)
			}

			t.Log("proof succeeded")
		})
	}
}
