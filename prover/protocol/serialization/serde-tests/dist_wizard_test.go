package serdetests

import (
	"reflect"
	"testing"

	"github.com/consensys/linea-monorepo/prover/protocol/distributed"
	"github.com/consensys/linea-monorepo/prover/protocol/serialization"
	"github.com/consensys/linea-monorepo/prover/utils/test_utils"
)

var (
	zkevm      = distributed.GetZkEVM()
	affinities = distributed.GetAffinities(zkevm)
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
	t.Run("ModuleNames", TestSerdeModuleNames)

	// t.Run("LPPs", TestSerdeLPPs)
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

/*
// TestSerdeLPPs tests serialization and deserialization of the LPPs field.
func TestSerdeLPPs(t *testing.T) {
	if dw == nil {
		t.Fatal("GetDistributedWizard returned nil")
	}
	if len(dw.LPPs) == 0 {
		fmt.Printf("No LPP modules at pre-compile time\n")
	}

	// Compiled IOPs
	// fmt.Println("DW LPPS compiled iop at index 0")
	// fmt.Println(dw.LPPs[0])

	// Serialize the original LPPs
	lppsSer, err := serializeModuleLPPs(dw.LPPs)
	if err != nil {
		t.Fatalf("Failed to serialize LPPs: %v", err)
	}

	// fmt.Printf("LPP modules serialized:%+v\n", lppsSer)

	// Create a new empty CompiledIOP for deserialization

	// Deserialize into a new LPPs
	deserializedLPPs, err := deserializeModuleLPPs(lppsSer)
	if err != nil {
		t.Fatalf("Failed to deserialize LPPs: %v", err)
	}

	// Compare structs while ignoring unexported fields
	if !test_utils.CompareExportedFields(dw.LPPs, deserializedLPPs) {
		t.Fatalf("Mis-matched fields after serde LPPs (ignoring unexported fields)")
	}
}

// serializeModuleLPPs serializes a slice of *ModuleLPP.
func serializeModuleLPPs(lpps []*distributed.ModuleLPP) ([]byte, error) {
	rawLPPs := make([]json.RawMessage, len(lpps))
	for i, lpp := range lpps {
		var lppSer json.RawMessage
		if lpp != nil {
			var err error
			lppSer, err = serialization.SerializeValue(reflect.ValueOf(lpp), serialization.DeclarationMode)
			if err != nil {
				return nil, fmt.Errorf("failed to serialize ModuleLPP at index %d: %w", i, err)
			}
		} else {
			lppSer = []byte(serialization.NilString)
		}
		rawLPPs[i] = lppSer
	}
	return serialization.SerializeAnyWithCborPkg(rawLPPs)
}

// deserializeModuleLPPs deserializes a slice of *ModuleLPP from CBOR-encoded data.
func deserializeModuleLPPs(data []byte) ([]*distributed.ModuleLPP, error) {
	var rawLPPs []json.RawMessage
	if err := serialization.DeserializeAnyWithCborPkg(data, &rawLPPs); err != nil {
		return nil, fmt.Errorf("failed to deserialize LPPs raw slice: %w", err)
	}

	lpps := make([]*distributed.ModuleLPP, len(rawLPPs))
	for i, raw := range rawLPPs {
		if bytes.Equal(raw, []byte(serialization.NilString)) {
			lpps[i] = nil
			continue
		}
		comp := serialization.NewEmptyCompiledIOP()
		deserializedVal, err := serialization.DeserializeValue(raw, serialization.DeclarationMode, reflect.TypeOf(&distributed.ModuleLPP{}), comp)
		if err != nil {
			return nil, fmt.Errorf("failed to deserialize ModuleLPP at index %d: %w", i, err)
		}

		lpp, ok := deserializedVal.Interface().(*distributed.ModuleLPP)
		if !ok {
			return nil, fmt.Errorf("deserialized value at index %d is not *distributed.ModuleLPP: got %T", i, deserializedVal.Interface())
		}
		lpps[i] = lpp
	}

	return lpps, nil
}


// serializeModuleLPP serializes a single ModuleLPP with its CompiledIOP's Columns store.
func serializeModuleLPP(lpp *distributed.ModuleLPP, comp *wizard.CompiledIOP) ([]byte, error) {
	if lpp == nil {
		return []byte(serialization.NilString), nil
	}

	if comp == nil {
		return nil, fmt.Errorf("ModuleLPP has nil CompiledIOP")
	}

	raw := make(map[string]json.RawMessage)

	// Serialize Columns from CompiledIOP
	columnsSer, err := serialization.SerializeValue(reflect.ValueOf(comp.Columns), serialization.DeclarationMode)
	if err != nil {
		return nil, fmt.Errorf("failed to serialize CompiledIOP Columns: %w", err)
	}
	raw["columns"] = columnsSer

	// Serialize the ModuleLPP itself
	lppSer, err := serialization.SerializeValue(reflect.ValueOf(lpp), serialization.DeclarationMode)
	if err != nil {
		return nil, fmt.Errorf("failed to serialize ModuleLPP: %w", err)
	}
	raw["lpp"] = lppSer

	return serialization.SerializeAnyWithCborPkg(raw)
}

// serializeModuleLPPs serializes a slice of ModuleLPPs, each with its own CompiledIOP.
func serializeModuleLPPs(lpps []*distributed.ModuleLPP) ([]byte, error) {
	rawLPPs := make([]json.RawMessage, len(lpps))
	for i, lpp := range lpps {
		var comp *wizard.CompiledIOP
		if lpp != nil {
			comp = lpp.GetModuleTranslator().Wiop
		}
		lppSer, err := serializeModuleLPP(lpp, comp)
		if err != nil {
			return nil, fmt.Errorf("failed to serialize ModuleLPP at index %d: %w", i, err)
		}
		rawLPPs[i] = lppSer
	}
	return serialization.SerializeAnyWithCborPkg(rawLPPs)
}

// deserializeModuleLPP deserializes a single ModuleLPP, restoring its CompiledIOP's Columns store.
func deserializeModuleLPP(data []byte, comp *wizard.CompiledIOP) (*distributed.ModuleLPP, error) {
	if bytes.Equal(data, []byte(serialization.NilString)) {
		return nil, nil
	}

	var raw map[string]json.RawMessage
	if err := serialization.DeserializeAnyWithCborPkg(data, &raw); err != nil {
		return nil, fmt.Errorf("failed to deserialize ModuleLPP raw data: %w", err)
	}

	// Deserialize Columns first
	columnsRaw, ok := raw["columns"]
	if !ok {
		return nil, fmt.Errorf("missing columns data in serialized ModuleLPP")
	}
	columnsVal, err := serialization.DeserializeValue(columnsRaw, serialization.DeclarationMode, reflect.TypeOf((*column.Store)(nil)).Elem(), comp)
	if err != nil {
		return nil, fmt.Errorf("failed to deserialize CompiledIOP Columns: %w", err)
	}
	columns, ok := columnsVal.Interface().(column.Store)
	if !ok {
		return nil, fmt.Errorf("deserialized columns is not *column.Store, got %T", columnsVal.Interface())
	}
	comp.Columns = columns

	// Deserialize the ModuleLPP
	lppRaw, ok := raw["lpp"]
	if !ok {
		return nil, fmt.Errorf("missing lpp data in serialized ModuleLPP")
	}
	deserializedVal, err := serialization.DeserializeValue(lppRaw, serialization.DeclarationMode, reflect.TypeOf(&distributed.ModuleLPP{}), comp)
	if err != nil {
		return nil, fmt.Errorf("failed to deserialize ModuleLPP: %w", err)
	}
	lpp, ok := deserializedVal.Interface().(*distributed.ModuleLPP)
	if !ok {
		return nil, fmt.Errorf("deserialized value is not *ModuleLPP, got %T", deserializedVal.Interface())
	}
	// Reattach the CompiledIOP to the moduleTranslator
	lpp.SetModuleTranslatorIOP(comp)
	return lpp, nil
}

// deserializeModuleLPPs deserializes a slice of ModuleLPPs, creating a new CompiledIOP for each.
func deserializeModuleLPPs(data []byte) ([]*distributed.ModuleLPP, error) {
	var rawLPPs []json.RawMessage
	if err := serialization.DeserializeAnyWithCborPkg(data, &rawLPPs); err != nil {
		return nil, fmt.Errorf("failed to deserialize LPPs raw slice: %w", err)
	}

	lpps := make([]*distributed.ModuleLPP, len(rawLPPs))
	for i, raw := range rawLPPs {
		comp := serialization.NewEmptyCompiledIOP()
		lpp, err := deserializeModuleLPP(raw, comp)
		if err != nil {
			return nil, fmt.Errorf("failed to deserialize ModuleLPP at index %d: %w", i, err)
		}
		lpps[i] = lpp
	}
	return lpps, nil
} */
