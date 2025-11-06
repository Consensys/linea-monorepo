package common

import (
	"testing"
	"github.com/consensys/linea-monorepo/prover/utils/csvtraces"
	"github.com/consensys/linea-monorepo/prover/protocol/wizard"
	"fmt"
	"github.com/consensys/linea-monorepo/prover/protocol/compiler/dummy"
	"github.com/consensys/linea-monorepo/prover/protocol/ifaces"
	"github.com/stretchr/testify/assert"
)

type testModuleSource struct {
	CsTest ifaces.Column
	Limbs  []ifaces.Column
}

func TestFlattenColumn(t *testing.T) {
	testCases := []struct {
		InputFName  string
		ModuleFName string
		NbLimbs     int
	}{
		{
			InputFName:  "testdata/flatten_column/default_input.csv",
			ModuleFName: "testdata/flatten_column/default_module.csv",
			NbLimbs:     8,
		},
		{
			InputFName:  "testdata/flatten_column/three_cols_input.csv",
			ModuleFName: "testdata/flatten_column/three_cols_module.csv",
			NbLimbs:     3,
		},
		{
			InputFName:  "testdata/flatten_column/alternating_mask_input.csv",
			ModuleFName: "testdata/flatten_column/alternating_mask_module.csv",
			NbLimbs:     8,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.InputFName, func(t *testing.T) {

			modCt := csvtraces.MustOpenCsvFile(tc.ModuleFName)
			inpCt := csvtraces.MustOpenCsvFile(tc.InputFName)

			var inp *testModuleSource
			var flattenColumn *FlattenColumn

			cmp := wizard.Compile(func(build *wizard.Builder) {
				comp := build.CompiledIOP

				inp = &testModuleSource{
					CsTest: inpCt.GetCommit(build, "CS_TEST"),
					Limbs:  make([]ifaces.Column, tc.NbLimbs),
				}

				for i := range tc.NbLimbs {
					inp.Limbs[i] = inpCt.GetCommit(build, fmt.Sprintf("LIMB_%d", i))
				}

				flattenColumn = NewFlattenColumn(comp, tc.NbLimbs, inp.Limbs[:], inp.CsTest)
				flattenColumn.CsFlattenProjection(comp)
			}, dummy.Compile)

			proof := wizard.Prove(cmp, func(run *wizard.ProverRuntime) {
				var limbsInpColsNames []string
				limbsInpColsNames = append(limbsInpColsNames, string(inp.CsTest.GetColID()))
				for i := 0; i < tc.NbLimbs; i++ {
					limbsInpColsNames = append(limbsInpColsNames, string(inp.Limbs[i].GetColID()))
				}

				inpCt.Assign(run, limbsInpColsNames...)

				flattenColumn.Run(run)

				modCt.CheckAssignment(run,
					string(flattenColumn.LimbsColID()),
					string(flattenColumn.MaskColID()),
				)
			})

			assert.NoError(t, wizard.Verify(cmp, proof), "proof failed")
		})
	}
}

func TestFlattenSharedColColumn(t *testing.T) {
	testCases := []struct {
		InputFName  string
		ModuleFName string
		NbLimbs     int
	}{
		{
			InputFName:  "testdata/flatten_column/shared_limbs_input.csv",
			ModuleFName: "testdata/flatten_column/shared_limbs_module.csv",
			NbLimbs:     8,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.InputFName, func(t *testing.T) {

			modCt := csvtraces.MustOpenCsvFile(tc.ModuleFName)
			inpCt := csvtraces.MustOpenCsvFile(tc.InputFName)

			var inpA, inpB *testModuleSource
			var flattenColumnA, flattenColumnB *FlattenColumn

			cmp := wizard.Compile(func(build *wizard.Builder) {
				comp := build.CompiledIOP

				inpA = &testModuleSource{
					CsTest: inpCt.GetCommit(build, "CS_MODULE_A"),
					Limbs:  make([]ifaces.Column, tc.NbLimbs),
				}

				inpB = &testModuleSource{
					CsTest: inpCt.GetCommit(build, "CS_MODULE_B"),
					Limbs:  make([]ifaces.Column, tc.NbLimbs),
				}

				for i := range tc.NbLimbs {
					limbCommit := inpCt.GetCommit(build, fmt.Sprintf("LIMB_%d", i))
					inpA.Limbs[i] = limbCommit
					inpB.Limbs[i] = limbCommit
				}

				flattenColumnA = NewFlattenColumn(comp, tc.NbLimbs, inpA.Limbs[:], inpA.CsTest)
				flattenColumnA.CsFlattenProjection(comp)

				flattenColumnB = NewFlattenColumn(comp, tc.NbLimbs, inpB.Limbs[:], inpB.CsTest)
				flattenColumnB.CsFlattenProjection(comp)
			}, dummy.Compile)

			proof := wizard.Prove(cmp, func(run *wizard.ProverRuntime) {
				var limbsInpColsNames []string
				limbsInpColsNames = append(limbsInpColsNames, string(inpA.CsTest.GetColID()))
				limbsInpColsNames = append(limbsInpColsNames, string(inpB.CsTest.GetColID()))
				for i := 0; i < tc.NbLimbs; i++ {
					limbsInpColsNames = append(limbsInpColsNames, string(inpA.Limbs[i].GetColID()))
				}

				inpCt.Assign(run, limbsInpColsNames...)

				flattenColumnA.Run(run)
				flattenColumnB.Run(run)

				modCt.CheckAssignment(run,
					string(flattenColumnA.LimbsColID()),
					string(flattenColumnA.MaskColID()),
					string(flattenColumnB.LimbsColID()),
					string(flattenColumnB.MaskColID()),
				)
			})

			assert.NoError(t, wizard.Verify(cmp, proof), "proof failed")
		})
	}
}
