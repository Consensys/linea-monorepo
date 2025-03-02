package experiment

import (
	"fmt"
	"testing"

	"github.com/consensys/linea-monorepo/prover/backend/files"
)

func TestQueryBasedDiscoveryOnZkEVM(t *testing.T) {

	var (
		zkevm = GetZkEVM()
		disc  = &QueryBasedModuleDiscoverer{}
	)

	precompileInitialWizard(zkevm.WizardIOP)

	// The test is to make sure that this function returns
	disc.Analyze(zkevm.WizardIOP)

	mapSize := map[ModuleName]int{}

	allCols := zkevm.WizardIOP.Columns.AllKeys()
	for _, colName := range allCols {
		col := zkevm.WizardIOP.Columns.GetHandle(colName)

		var (
			oldSize = col.Size()
			newSize = disc.NewSizeOf(col)
			module  = disc.ModuleOf(col)
		)

		if module == "" {
			t.Errorf("module of %v is empty", colName)
		}

		if newSize == 0 {
			t.Errorf("new-size of %v is 0", colName)
		}

		if oldSize != newSize {
			t.Errorf("new-size of %v is %v but expected %v", colName, newSize, oldSize)
		}

		if _, ok := mapSize[module]; !ok {
			mapSize[module] = oldSize
		}

		if mapSize[module] != oldSize {
			t.Errorf("size of %v is %v but expected %v", module, oldSize, mapSize[module])
		}
	}

}

func TestStandardDiscoveryOnZkEVM(t *testing.T) {

	var (
		zkevm = GetZkEVM()
		disc  = &StandardModuleDiscoverer{
			TargetWeight: 1 << 27,
		}
	)

	precompileInitialWizard(zkevm.WizardIOP)

	// The test is to make sure that this function returns
	disc.Analyze(zkevm.WizardIOP)

	csvFile := files.MustOverwrite("modules.csv")
	fmt.Fprintf(csvFile, "%v, %v, %v, %v\n", "column", "size-old", "size-new", "module")

	allCols := zkevm.WizardIOP.Columns.AllKeys()
	for _, colName := range allCols {
		col := zkevm.WizardIOP.Columns.GetHandle(colName)

		var (
			oldSize = col.Size()
			newSize = disc.NewSizeOf(col)
			module  = disc.ModuleOf(col)
		)

		if module == "" {
			t.Errorf("module of %v is empty", colName)
		}

		if newSize == 0 {
			t.Errorf("new-size of %v is 0", colName)
		}

		fmt.Fprintf(csvFile, "%v, %v, %v, %v\n", colName, oldSize, newSize, module)
	}
}
