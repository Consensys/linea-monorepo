package distributed

import (
	"testing"

	"github.com/consensys/linea-monorepo/prover/protocol/column"
)

func TestQueryBasedDiscoveryOnZkEVM(t *testing.T) {

	var (
		zkevm = GetZkEVM()
		disc  = &QueryBasedModuleDiscoverer{}
	)

	precompileInitialWizard(zkevm.WizardIOP, nil)

	// The test is to make sure that this function returns
	disc.Analyze(zkevm.WizardIOP)

	mapSize := map[ModuleName]int{}

	allCols := zkevm.WizardIOP.Columns.AllKeys()
	for _, colName := range allCols {
		col := zkevm.WizardIOP.Columns.GetHandle(colName)

		var (
			oldSize = col.Size()
			nat     = col.(column.Natural)
			newSize = disc.NewSizeOf(nat)
			module  = disc.ModuleOf(nat)
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

	for _, col := range zkevm.WizardIOP.Columns.AllKeys() {

		var (
			nat     = zkevm.WizardIOP.Columns.GetHandle(col).(column.Natural)
			modules = []ModuleName{}
		)

		for i := range disc.modules {
			mod := disc.modules[i]
			if mod.ds.Has(nat) {
				modules = append(modules, mod.moduleName)
			}
		}

		if len(modules) == 0 {
			t.Errorf("could not match any module for %v", col)
		}

		if len(modules) > 1 {
			t.Errorf("could match more than one module for %v: %v", col, modules)
		}
	}
}

func TestStandardDiscoveryOnZkEVM(t *testing.T) {

	var (
		zkevm = GetZkEVM()
		disc  = &StandardModuleDiscoverer{
			TargetWeight: 1 << 28,
			Affinities:   GetAffinities(zkevm),
		}
	)

	precompileInitialWizard(zkevm.WizardIOP, nil)

	// The test is to make sure that this function returns
	disc.Analyze(zkevm.WizardIOP)

	allCols := zkevm.WizardIOP.Columns.AllKeys()
	for _, colName := range allCols {
		col := zkevm.WizardIOP.Columns.GetHandle(colName)

		var (
			nat     = col.(column.Natural)
			newSize = disc.NewSizeOf(nat)
			module  = disc.ModuleOf(nat)
		)

		if module == "" {
			t.Errorf("module of %v is empty", colName)
		}

		if newSize == 0 {
			t.Errorf("new-size of %v is 0", colName)
		}
	}

	for _, col := range zkevm.WizardIOP.Columns.AllKeys() {

		var (
			nat     = zkevm.WizardIOP.Columns.GetHandle(col).(column.Natural)
			modules = []ModuleName{}
		)

		for i := range disc.modules {
			mod := disc.modules[i]
			for k := range mod.subModules {
				if mod.subModules[k].ds.Has(nat) {
					modules = append(modules, mod.moduleName)
				}
			}
		}

		if len(modules) == 0 {
			t.Errorf("could not match any module for %v", col)
		}

		if len(modules) > 1 {
			t.Errorf("could match more than one module for %v: %v", col, modules)
		}
	}

	t.Logf("totalNumber of columns: %v", len(zkevm.WizardIOP.Columns.AllKeys()))

	for _, mod := range disc.modules {
		t.Logf("module=%v weight=%v numcol=%v\n", mod.moduleName, mod.Weight(), disc.NumColumnOf(mod.moduleName))
	}
}
