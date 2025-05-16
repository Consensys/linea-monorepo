package serdetests

import (
	"reflect"
	"testing"

	"github.com/consensys/linea-monorepo/prover/protocol/distributed"
	"github.com/consensys/linea-monorepo/prover/protocol/serialization"
	"github.com/consensys/linea-monorepo/prover/utils/test_utils"
)

var (
	zkevm      = test_utils.GetZkEVM()
	affinities = test_utils.GetAffinities(zkevm)
	discoverer = &distributed.StandardModuleDiscoverer{
		TargetWeight: 1 << 28,
		Affinities:   affinities,
		Predivision:  1,
	}
	dw = distributed.DistributeWizard(zkevm.WizardIOP, discoverer)
)

// TestSerdeDistWizard tests serialization and deserialization of DistributedWizard fields.
func TestSerdeDistWizard(t *testing.T) {
	// Run subtests for attributes
	t.Run("TestSerdeModuleNames", TestSerdeModuleNames)
	t.Run("TestSerdeLPPs", TestSerdeLPPs)
	t.Run("TestSerdeGLs", TestSerdeGLs)
	t.Run("TestSerdeDefaultModule", TestSerdeDefMods)
	t.Run("TestSerdeBootstrapper", TestSerdeBootstrapper)
	t.Run("TestSerdeModDisc", TestSerdeDWModDisc)
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
	serializedData, err := serializeModuleLPPs(dw.LPPs)
	if err != nil {
		t.Fatalf("failed to serialize LPPs: %v", err)
	}

	deserializedLPPs, err := deserializeModuleLPPs(serializedData)
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
