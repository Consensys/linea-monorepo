package assets

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/consensys/linea-monorepo/prover/protocol/distributed"
	"github.com/consensys/linea-monorepo/prover/protocol/serialization"
	"github.com/consensys/linea-monorepo/prover/utils/test_utils"
)

var (
	zkEVM      = test_utils.GetZkEVM()
	affinities = test_utils.GetAffinities(zkEVM)
	discoverer = &distributed.StandardModuleDiscoverer{
		TargetWeight: 1 << 28,
		Affinities:   affinities,
		Predivision:  1,
	}
	// dw = distributed.DistributeWizard(zkEVM.WizardIOP, discoverer)
	dw = distributed.DistributeWizard(zkEVM.WizardIOP, discoverer).CompileSegments()
)

// TestSerdeDistWizardFull tests full serialization and deserialization of DistributedWizard.
func TestSerdeDistWizardFull(t *testing.T) {
	if dw == nil {
		t.Fatal("DistributedWizard is nil")
	}

	// Serialize the DistributedWizard
	serData, err := SerializeDistWizard(dw)
	if err != nil {
		t.Fatalf("Failed to serialize DistributedWizard: %v", err)
	}

	// Deserialize the DistributedWizard
	deserializedDW, err := DeserializeDistWizard(serData)
	if err != nil {
		t.Fatalf("Failed to deserialize DistributedWizard: %v", err)
	}

	// Compare exported fields
	if !test_utils.CompareExportedFields(dw, deserializedDW) {
		t.Errorf("Mismatch in exported fields after full DistributedWizard serde")
	}
}

// TestSerdeDistWizard tests serialization and deserialization of DistributedWizard fields.
func TestSerdeDistWizard(t *testing.T) {
	// Run subtests for attributes
	t.Run("TestSerdeModuleNames", TestSerdeModuleNames)
	t.Run("TestSerdeLPPs", TestSerdeLPPs)
	t.Run("TestSerdeGLs", TestSerdeGLs)
	t.Run("TestSerdeDefaultModule", TestSerdeDefMods)
	t.Run("TestSerdeBootstrapper", TestSerdeBootstrapper)
	t.Run("TestSerdeModDisc", TestSerdeDWModDisc)
	t.Run("TestSerdeCompiledGLs", TestSerdeCompiledGLs)
	t.Run("TestSerdeCompiledLPPs", TestSerdeCompiledLPPs)
	t.Run("TestSerdeDWCompiledDef", TestSerdeDWCompiledDef)
	t.Run("TestSerdeDWCompiledConglomeration", TestSerdeDWCompiledCong)
}

// TestSerdeModuleNames tests serialization and deserialization of the ModuleNames field.
func TestSerdeModuleNames(t *testing.T) {
	if dw == nil {
		t.Fatal("GetDistributedWizard returned nil")
	}
	if dw.ModuleNames == nil {
		t.Fatal("ModuleNames field is nil")
	}

	// Serialize the original ModuleNames
	moduleNamesSer, err := serializeValue(dw.ModuleNames)
	if err != nil {
		t.Fatalf("Failed to serialize ModuleNames: %v", err)
	}

	// Create a new empty CompiledIOP for deserialization
	comp := serialization.NewEmptyCompiledIOP()

	// Deserialize into a new ModuleNames
	deserializedVal, err := serialization.DeserializeValue(moduleNamesSer, serialization.DeclarationMode, reflect.TypeOf([]distributed.ModuleName{}), comp)
	if err != nil {
		t.Fatalf("Failed to deserialize ModuleNames: %v", err)
	}
	deserialized, ok := deserializedVal.Interface().([]distributed.ModuleName)
	if !ok {
		t.Fatalf("Deserialized value is not []ModuleName: got %T", deserializedVal.Interface())
	}

	// Compare structs while ignoring unexported fields
	if !test_utils.CompareExportedFields(dw.ModuleNames, deserialized) {
		t.Fatalf("Mis-matched fields after serde ModuleNames (ignoring unexported fields)")
	}
}

// TestSerdeLPPS tests full serialization and deserialization of the LPPs modules.
func TestSerdeLPPs(t *testing.T) {
	serializedData, err := SerializeModuleLPPs(dw.LPPs)
	if err != nil {
		t.Fatalf("failed to serialize LPPs: %v", err)
	}

	deserializedLPPs, err := DeserializeModuleLPPs(serializedData)
	if err != nil {
		t.Fatalf("failed to deserialize LPPs: %v", err)
	}

	if !test_utils.CompareExportedFields(dw.LPPs, deserializedLPPs) {
		t.Errorf("mismatch in exported fields after LPP mods serde")
	}
}

// TestSerdeGLs tests full serialization and deserialization of the GL modules.
func TestSerdeGLs(t *testing.T) {
	serializedData, err := SerializeModuleGLs(dw.GLs)
	if err != nil {
		t.Fatalf("failed to serialize LPPs: %v", err)
	}

	deserializedGLs, err := DeserializeModuleGLs(serializedData)
	if err != nil {
		t.Fatalf("failed to deserialize LPPs: %v", err)
	}

	if !test_utils.CompareExportedFields(dw.GLs, deserializedGLs) {
		t.Errorf("mismatch in exported fields after GL mods serde")
	}
}

func TestSerdeDefMods(t *testing.T) {
	if dw == nil {
		t.Fatal("distributed wizard is nil")
	}

	if dw.DefaultModule == nil {
		t.Fatal("Dist. Wizard default module i nil")
	}

	serDM, err := SerializeDWDefMods(dw.DefaultModule)
	if err != nil {
		t.Fatalf("error during serializing distributed wizard default module:%s \n", err.Error())
	}

	deSerDM, err := DeserializeDWDefMods(serDM)
	if err != nil {
		t.Fatalf("error during de-serializing distributed wizard default module:%s \n", err.Error())
	}

	if !test_utils.CompareExportedFields(dw.DefaultModule, deSerDM) {
		t.Errorf("mismatch in exported fields after DW Def.Mods serde")
	}
}

func TestSerdeBootstrapper(t *testing.T) {
	if dw == nil {
		t.Fatal("distributed wizard is nil")
	}

	if dw.Bootstrapper == nil {
		t.Fatal("Dist. Wizard default module i nil")
	}

	serBootstrap, err := serialization.SerializeCompiledIOP(dw.Bootstrapper)
	if err != nil {
		t.Fatalf("error during serializing distributed wizard bootstrapper:%s \n", err.Error())
	}

	deSerBootstrap, err := serialization.DeserializeCompiledIOP(serBootstrap)
	if err != nil {
		t.Fatalf("error during deserializing distributed wizard bootstrapper:%s \n", err.Error())
	}

	if !test_utils.CompareExportedFields(dw.Bootstrapper, deSerBootstrap) {
		t.Errorf("mismatch in exported fields after DW Def.Mods serde")
	}
}

func TestSerdeCompiledGLs(t *testing.T) {

	fmt.Printf("Length CompiledGLs:%d\n", len(dw.CompiledGLs))
	if len(dw.CompiledGLs) == 0 {
		fmt.Println("Skipping TestSerdeCompiledGLs due to nil")
	}
}

func TestSerdeCompiledLPPs(t *testing.T) {
	fmt.Printf("Length CompiledLPPs:%d\n", len(dw.CompiledLPPs))
	if len(dw.CompiledLPPs) == 0 {
		fmt.Println("Skipping TestSerdeCompiledLPPs due to nil")
	}
}

func TestSerdeDWCompiledDef(t *testing.T) {
	compDef := dw.CompiledDefault
	if compDef == nil {
		t.Skipf("No need for serde test due to nil CompiledDefault")
	}

	compDefSer, err := SerializeRecursedSegmentCompilation(compDef)
	if err != nil {
		t.Errorf("error during serializing dist.wizard compiled default of type *distributed.RecursedSegmentCompilation")
	}

	deSerCompDef, err := DeserializeRecursedSegmentCompilation(compDefSer)
	if err != nil {
		t.Errorf("error during deserializing dist.wizard compiled default of type *distributed.RecursedSegmentCompilation")
	}

	if !test_utils.CompareExportedFields(compDef, deSerCompDef) {
		t.Errorf("mismatch in exported fields after DWCompiledDefault serde")
	}
}

func TestSerdeDWCompiledCong(t *testing.T) {
	compDef := dw.CompiledConglomeration
	if compDef == nil {
		t.Skipf("No need for serde test due to nil CompiledConglomeration")
	}
}
