//go:build !fuzzlight

package modexp

import (
	"fmt"
	"github.com/consensys/linea-monorepo/prover/protocol/ifaces"
	"testing"

	"github.com/consensys/linea-monorepo/prover/protocol/compiler/dummy"
	"github.com/consensys/linea-monorepo/prover/protocol/query"
	"github.com/consensys/linea-monorepo/prover/protocol/wizard"
	"github.com/consensys/linea-monorepo/prover/utils/csvtraces"
	"github.com/consensys/linea-monorepo/prover/zkevm/prover/common"
)

func TestModexpWithCircuit(t *testing.T) {

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
					Settings:         Settings{MaxNbInstance256: 1, MaxNbInstance4096: 1, NbInstancesPerCircuitModexp256: 1, NbInstancesPerCircuitModexp4096: 1},
					Limbs:            limbs,
				}

				mod = newModule(build.CompiledIOP, inp).
					WithCircuit(build.CompiledIOP, query.PlonkRangeCheckOption(21, 4, false))
			}, dummy.Compile)

			proof := wizard.Prove(cmp, func(run *wizard.ProverRuntime) {

				var names []string
				for i := 0; i < common.NbLimbU128; i++ {
					names = append(names, fmt.Sprintf("LIMBS_%d", i))
				}

				names = append(names, "IS_MODEXP_BASE", "IS_MODEXP_EXPONENT", "IS_MODEXP_MODULUS", "IS_MODEXP_RESULT")

				inpCt.Assign(run, names...)

				mod.Assign(run)
			})

			if err := wizard.Verify(cmp, proof); err != nil {
				t.Fatal("proof failed", err)
			}

			t.Log("proof succeeded")
		})
	}

}
