package serdetests

import (
	"bytes"
	"encoding/json"
	"fmt"
	"reflect"
	"testing"

	"github.com/consensys/linea-monorepo/prover/protocol/distributed"
	"github.com/consensys/linea-monorepo/prover/protocol/ifaces"
	"github.com/consensys/linea-monorepo/prover/protocol/query"
	"github.com/consensys/linea-monorepo/prover/protocol/serialization"
	"github.com/consensys/linea-monorepo/prover/utils/test_utils"
)

var lpp = dw.LPPs[0]

// rawModuleLPP represents the serialized form of ModuleLPP.
type rawModuleLPP struct {
	CompiledIOP            json.RawMessage `json:"compiledIOP"`
	DefinitionInputs       json.RawMessage `json:"definitionInputs"`
	InitialFiatShamirState json.RawMessage `json:"initialFiatShamirState"`
	N0Hash                 json.RawMessage `json:"n0Hash"`
	N1Hash                 json.RawMessage `json:"n1Hash"`
	LogDerivativeSum       json.RawMessage `json:"logDerivativeSum"`
	GrandProduct           json.RawMessage `json:"grandProduct"`
	Horner                 json.RawMessage `json:"horner"`
}

// TestSerdeLPPs tests serialization and deserialization of the LPPs field with individual subtests.
func TestSerdeLPPs(t *testing.T) {
	if dw == nil {
		t.Fatal("GetDistributedWizard returned nil")
	}

	// Run individual subtests
	// Updated t.Run calls
	t.Run("CompiledIOP", TestLPPIOP)
	t.Run("InitialFiatShamirState", TestLPPFSState)
	t.Run("N0Hash", TestLPPN0Hash)
	t.Run("N1Hash", TestLPPN1Hash)
	t.Run("LogDerivativeSum", TestLPPLogDerivativeSum)
	t.Run("GrandProduct", TestLPPGrandProduct)
	t.Run("Horner", TestLPPHorner)
	t.Run("FullModuleLPP", TestFullModuleLPP)
	t.Run("FullLPPsSlice", TestFullLPPsSlice)
}

// TestLPPIOP tests serialization and deserialization of the CompiledIOP field.
func TestLPPIOP(t *testing.T) {
	comp := lpp.GetModuleTranslator().Wiop
	if comp == nil {
		t.Fatal("CompiledIOP is nil")
	}

	serializedData, err := serialization.SerializeCompiledIOP(comp)
	if err != nil {
		t.Fatalf("Failed to serialize CompiledIOP: %v", err)
	}

	deserializedComp, err := serialization.DeserializeCompiledIOP(serializedData)
	if err != nil {
		t.Fatalf("Failed to deserialize CompiledIOP: %v", err)
	}

	if !test_utils.CompareExportedFields(comp, deserializedComp) {
		t.Errorf("Mismatch in exported fields for CompiledIOP")
	}
}

// TestLPPFSState tests serialization and deserialization of the InitialFiatShamirState field.
func TestLPPFSState(t *testing.T) {
	testField(t, "InitialFiatShamirState", lpp.InitialFiatShamirState, reflect.TypeOf((*ifaces.Column)(nil)).Elem())
}

// TestLPPN0Hash tests serialization and deserialization of the N0Hash field.
func TestLPPN0Hash(t *testing.T) {
	testField(t, "N0Hash", lpp.N0Hash, reflect.TypeOf((*ifaces.Column)(nil)).Elem())
}

// TestLPPN1Hash tests serialization and deserialization of the N1Hash field.
func TestLPPN1Hash(t *testing.T) {
	testField(t, "N1Hash", lpp.N1Hash, reflect.TypeOf((*ifaces.Column)(nil)).Elem())
}

// TestLPPLogDerivativeSum tests serialization and deserialization of the LogDerivativeSum field.
func TestLPPLogDerivativeSum(t *testing.T) {
	testField(t, "LogDerivativeSum", lpp.LogDerivativeSum, reflect.TypeOf(query.LogDerivativeSum{}))
}

// TestLPPGrandProduct tests serialization and deserialization of the GrandProduct field.
func TestLPPGrandProduct(t *testing.T) {
	testField(t, "GrandProduct", lpp.GrandProduct, reflect.TypeOf(query.GrandProduct{}))
}

// TestLPPHorner tests serialization and deserialization of the Horner field.
func TestLPPHorner(t *testing.T) {
	testField(t, "Horner", lpp.Horner, reflect.TypeOf(query.Horner{}))
}

// TestFullModuleLPP tests full serialization and deserialization of a ModuleLPP.
func TestFullModuleLPP(t *testing.T) {
	serializedData, err := serializeModuleLPP(lpp)
	if err != nil {
		t.Fatalf("Failed to serialize ModuleLPP: %v", err)
	}

	deserializedLPP, err := deserializeModuleLPP(serializedData)
	if err != nil {
		t.Fatalf("Failed to deserialize ModuleLPP: %v", err)
	}

	if !test_utils.CompareExportedFields(lpp, deserializedLPP) {
		t.Errorf("Mismatch in exported fields after full ModuleLPP serde")
	}
}

// TestFullLPPsSlice tests full serialization and deserialization of the LPPs slice.
func TestFullLPPsSlice(t *testing.T) {
	serializedData, err := serializeModuleLPPs(dw.LPPs)
	if err != nil {
		t.Fatalf("Failed to serialize LPPs: %v", err)
	}

	deserializedLPPs, err := deserializeModuleLPPs(serializedData)
	if err != nil {
		t.Fatalf("Failed to deserialize LPPs: %v", err)
	}

	if !test_utils.CompareExportedFields(dw.LPPs, deserializedLPPs) {
		t.Errorf("Mismatch in exported fields after full LPPs slice serde")
	}
}

// testField handles serialization, deserialization, and comparison for a single field.
func testField(t *testing.T, fieldName string, fieldValue interface{}, fieldType reflect.Type) {
	if fieldValue == nil {
		t.Fatalf("%s field is nil", fieldName)
	}

	serializedData, err := serializeValue(fieldValue)
	if err != nil {
		t.Fatalf("Failed to serialize %s: %v", fieldName, err)
	}

	comp := serialization.NewEmptyCompiledIOP()

	deserializedVal, err := serialization.DeserializeValue(serializedData, serialization.DeclarationMode, fieldType, comp)
	if err != nil {
		t.Fatalf("Failed to deserialize %s: %v", fieldName, err)
	}

	if !test_utils.CompareExportedFields(fieldValue, deserializedVal.Interface()) {
		t.Errorf("Mismatch in exported fields for %s", fieldName)
	}
}

// serializeModuleLPP serializes a single ModuleLPP instance field-by-field.
func serializeModuleLPP(lpp *distributed.ModuleLPP) ([]byte, error) {
	if lpp == nil {
		return []byte(serialization.NilString), nil
	}

	comp := lpp.GetModuleTranslator().Wiop
	if comp == nil {
		return nil, fmt.Errorf("ModuleLPP has nil CompiledIOP")
	}

	raw := &rawModuleLPP{}

	// Serialize CompiledIOP first (includes Columns store)
	compSer, err := serialization.SerializeCompiledIOP(comp)
	if err != nil {
		return nil, fmt.Errorf("failed to serialize CompiledIOP: %w", err)
	}
	raw.CompiledIOP = compSer

	// Serialize definitionInputs
	// if len(lpp.definitionInputs) > 0 {
	// 	defSer, err := serialization.SerializeValue(reflect.ValueOf(lpp.definitionInputs), serialization.DeclarationMode)
	// 	if err != nil {
	// 		return nil, fmt.Errorf("failed to serialize definitionInputs: %w", err)
	// 	}
	// 	raw.DefinitionInputs = defSer
	// } else {
	// 	raw.DefinitionInputs = []byte(serialization.NilString)
	// }

	// Serialize InitialFiatShamirState
	if lpp.InitialFiatShamirState != nil {
		ifsSer, err := serialization.SerializeValue(reflect.ValueOf(lpp.InitialFiatShamirState), serialization.DeclarationMode)
		if err != nil {
			return nil, fmt.Errorf("failed to serialize InitialFiatShamirState: %w", err)
		}
		raw.InitialFiatShamirState = ifsSer
	} else {
		raw.InitialFiatShamirState = []byte(serialization.NilString)
	}

	// Serialize N0Hash
	if lpp.N0Hash != nil {
		n0Ser, err := serialization.SerializeValue(reflect.ValueOf(lpp.N0Hash), serialization.DeclarationMode)
		if err != nil {
			return nil, fmt.Errorf("failed to serialize N0Hash: %w", err)
		}
		raw.N0Hash = n0Ser
	} else {
		raw.N0Hash = []byte(serialization.NilString)
	}

	// Serialize N1Hash
	if lpp.N1Hash != nil {
		n1Ser, err := serialization.SerializeValue(reflect.ValueOf(lpp.N1Hash), serialization.DeclarationMode)
		if err != nil {
			return nil, fmt.Errorf("failed to serialize N1Hash: %w", err)
		}
		raw.N1Hash = n1Ser
	} else {
		raw.N1Hash = []byte(serialization.NilString)
	}

	// Serialize LogDerivativeSum
	if !reflect.ValueOf(lpp.LogDerivativeSum).IsZero() {
		ldsSer, err := serialization.SerializeValue(reflect.ValueOf(lpp.LogDerivativeSum), serialization.DeclarationMode)
		if err != nil {
			return nil, fmt.Errorf("failed to serialize LogDerivativeSum: %w", err)
		}
		raw.LogDerivativeSum = ldsSer
	} else {
		raw.LogDerivativeSum = []byte(serialization.NilString)
	}

	// Serialize GrandProduct
	if !reflect.ValueOf(lpp.GrandProduct).IsZero() {
		gpSer, err := serialization.SerializeValue(reflect.ValueOf(lpp.GrandProduct), serialization.DeclarationMode)
		if err != nil {
			return nil, fmt.Errorf("failed to serialize GrandProduct: %w", err)
		}
		raw.GrandProduct = gpSer
	} else {
		raw.GrandProduct = []byte(serialization.NilString)
	}

	// Serialize Horner
	if !reflect.ValueOf(lpp.Horner).IsZero() {
		hSer, err := serialization.SerializeValue(reflect.ValueOf(lpp.Horner), serialization.DeclarationMode)
		if err != nil {
			return nil, fmt.Errorf("failed to serialize Horner: %w", err)
		}
		raw.Horner = hSer
	} else {
		raw.Horner = []byte(serialization.NilString)
	}

	return serialization.SerializeAnyWithCborPkg(raw)
}

// deserializeModuleLPP deserializes a single ModuleLPP instance field-by-field.
func deserializeModuleLPP(data []byte) (*distributed.ModuleLPP, error) {
	if bytes.Equal(data, []byte(serialization.NilString)) {
		return nil, nil
	}

	var raw rawModuleLPP
	if err := serialization.DeserializeAnyWithCborPkg(data, &raw); err != nil {
		return nil, fmt.Errorf("failed to deserialize ModuleLPP raw data: %w", err)
	}

	// Deserialize CompiledIOP first (includes Columns store)
	if raw.CompiledIOP == nil {
		return nil, fmt.Errorf("missing CompiledIOP data in serialized ModuleLPP")
	}
	comp, err := serialization.DeserializeCompiledIOP(raw.CompiledIOP)
	if err != nil {
		return nil, fmt.Errorf("failed to deserialize CompiledIOP: %w", err)
	}

	// Initialize ModuleLPP
	lpp := &distributed.ModuleLPP{}
	lpp.SetModuleTranslatorIOP(comp)

	// Deserialize definitionInputs
	// if !bytes.Equal(raw.DefinitionInputs, []byte(serialization.NilString)) {
	// 	defVal, err := serialization.DeserializeValue(raw.DefinitionInputs, serialization.DeclarationMode, reflect.TypeOf([]FilteredModuleInputs{}), comp)
	// 	if err != nil {
	// 		return nil, fmt.Errorf("failed to deserialize definitionInputs: %w", err)
	// 	}
	// 	lpp.definitionInputs = defVal.Interface().([]FilteredModuleInputs)
	// }

	// Deserialize InitialFiatShamirState (depends on Columns)
	if !bytes.Equal(raw.InitialFiatShamirState, []byte(serialization.NilString)) {
		ifsVal, err := serialization.DeserializeValue(raw.InitialFiatShamirState, serialization.DeclarationMode, reflect.TypeOf((*ifaces.Column)(nil)).Elem(), comp)
		if err != nil {
			return nil, fmt.Errorf("failed to deserialize InitialFiatShamirState: %w", err)
		}
		lpp.InitialFiatShamirState = ifsVal.Interface().(ifaces.Column)
	}

	// Deserialize N0Hash (depends on Columns)
	if !bytes.Equal(raw.N0Hash, []byte(serialization.NilString)) {
		n0Val, err := serialization.DeserializeValue(raw.N0Hash, serialization.DeclarationMode, reflect.TypeOf((*ifaces.Column)(nil)).Elem(), comp)
		if err != nil {
			return nil, fmt.Errorf("failed to deserialize N0Hash: %w", err)
		}
		lpp.N0Hash = n0Val.Interface().(ifaces.Column)
	}

	// Deserialize N1Hash (depends on Columns)
	if !bytes.Equal(raw.N1Hash, []byte(serialization.NilString)) {
		n1Val, err := serialization.DeserializeValue(raw.N1Hash, serialization.DeclarationMode, reflect.TypeOf((*ifaces.Column)(nil)).Elem(), comp)
		if err != nil {
			return nil, fmt.Errorf("failed to deserialize N1Hash: %w", err)
		}
		lpp.N1Hash = n1Val.Interface().(ifaces.Column)
	}

	// Deserialize LogDerivativeSum
	if !bytes.Equal(raw.LogDerivativeSum, []byte(serialization.NilString)) {
		ldsVal, err := serialization.DeserializeValue(raw.LogDerivativeSum, serialization.DeclarationMode, reflect.TypeOf(query.LogDerivativeSum{}), comp)
		if err != nil {
			return nil, fmt.Errorf("failed to deserialize LogDerivativeSum: %w", err)
		}
		lpp.LogDerivativeSum = ldsVal.Interface().(query.LogDerivativeSum)
	}

	// Deserialize GrandProduct
	if !bytes.Equal(raw.GrandProduct, []byte(serialization.NilString)) {
		gpVal, err := serialization.DeserializeValue(raw.GrandProduct, serialization.DeclarationMode, reflect.TypeOf(query.GrandProduct{}), comp)
		if err != nil {
			return nil, fmt.Errorf("failed to deserialize GrandProduct: %w", err)
		}
		lpp.GrandProduct = gpVal.Interface().(query.GrandProduct)
	}

	// Deserialize Horner
	if !bytes.Equal(raw.Horner, []byte(serialization.NilString)) {
		hVal, err := serialization.DeserializeValue(raw.Horner, serialization.DeclarationMode, reflect.TypeOf(query.Horner{}), comp)
		if err != nil {
			return nil, fmt.Errorf("failed to deserialize Horner: %w", err)
		}
		lpp.Horner = hVal.Interface().(query.Horner)
	}

	return lpp, nil
}

// serializeModuleLPPs serializes a slice of ModuleLPP instances.
func serializeModuleLPPs(lpps []*distributed.ModuleLPP) ([]byte, error) {
	rawLPPs := make([]json.RawMessage, len(lpps))
	for i, lpp := range lpps {
		lppSer, err := serializeModuleLPP(lpp)
		if err != nil {
			return nil, fmt.Errorf("failed to serialize ModuleLPP at index %d: %w", i, err)
		}
		rawLPPs[i] = lppSer
	}
	return serialization.SerializeAnyWithCborPkg(rawLPPs)
}

// deserializeModuleLPPs deserializes a slice of ModuleLPP instances.
func deserializeModuleLPPs(data []byte) ([]*distributed.ModuleLPP, error) {
	var rawLPPs []json.RawMessage
	if err := serialization.DeserializeAnyWithCborPkg(data, &rawLPPs); err != nil {
		return nil, fmt.Errorf("failed to deserialize LPPs raw slice: %w", err)
	}

	lpps := make([]*distributed.ModuleLPP, len(rawLPPs))
	for i, raw := range rawLPPs {
		lpp, err := deserializeModuleLPP(raw)
		if err != nil {
			return nil, fmt.Errorf("failed to deserialize ModuleLPP at index %d: %w", i, err)
		}
		lpps[i] = lpp
	}
	return lpps, nil
}

/*
// TestSerdeLPPs tests serialization and deserialization of the LPPs field with subtests for each field.
func TestSerdeLPPs(t *testing.T) {
	if dw == nil {
		t.Fatal("GetDistributedWizard returned nil")
	}
	if len(dw.LPPs) == 0 {
		t.Skip("No LPP modules to test")
	}

	lpp := dw.LPPs[0]
	if lpp == nil {
		t.Fatal("LPP at index 0 is nil")
	}

	// Run subtests for each field
	t.Run("CompiledIOP", func(t *testing.T) { testField(t, lpp, "CompiledIOP") })
	t.Run("DefinitionInputs", func(t *testing.T) { testField(t, lpp, "DefinitionInputs") })
	t.Run("InitialFiatShamirState", func(t *testing.T) { testField(t, lpp, "InitialFiatShamirState") })
	t.Run("N0Hash", func(t *testing.T) { testField(t, lpp, "N0Hash") })
	t.Run("N1Hash", func(t *testing.T) { testField(t, lpp, "N1Hash") })
	t.Run("LogDerivativeSum", func(t *testing.T) { testField(t, lpp, "LogDerivativeSum") })
	t.Run("GrandProduct", func(t *testing.T) { testField(t, lpp, "GrandProduct") })
	t.Run("Horner", func(t *testing.T) { testField(t, lpp, "Horner") })
	t.Run("FullModuleLPP", func(t *testing.T) {
		ser, err := serializeModuleLPP(lpp)
		if err != nil {
			t.Fatalf("Failed to serialize ModuleLPP: %v", err)
		}
		deserLPP, err := deserializeModuleLPP(ser)
		if err != nil {
			t.Fatalf("Failed to deserialize ModuleLPP: %v", err)
		}
		if !test_utils.CompareExportedFields(lpp, deserLPP) {
			t.Errorf("Mismatch in exported fields after full ModuleLPP serde")
		}
	})
	t.Run("FullLPPsSlice", func(t *testing.T) {
		ser, err := serializeModuleLPPs(dw.LPPs)
		if err != nil {
			t.Fatalf("Failed to serialize LPPs: %v", err)
		}
		deserLPPs, err := deserializeModuleLPPs(ser)
		if err != nil {
			t.Fatalf("Failed to deserialize LPPs: %v", err)
		}
		if !test_utils.CompareExportedFields(dw.LPPs, deserLPPs) {
			t.Errorf("Mismatch in exported fields after full LPPs slice serde")
		}
	})
}

// testField tests serialization and deserialization of a single field of ModuleLPP.
func testField(t *testing.T, lpp *distributed.ModuleLPP, fieldName string) {
	comp := lpp.GetModuleTranslator().Wiop
	if comp == nil {
		t.Fatal("CompiledIOP is nil")
	}

	// Define field metadata
	type fieldInfo struct {
		value     interface{}
		typ       reflect.Type
		compare   func(t *testing.T, orig, deser interface{})
		serialize func() ([]byte, error)
	}

	fields := map[string]fieldInfo{
		"CompiledIOP": {
			value: comp,
			typ:   reflect.TypeOf(&wizard.CompiledIOP{}),
			compare: func(t *testing.T, orig, deser interface{}) {
				if !test_utils.CompareExportedFields(orig, deser) {
					t.Errorf("CompiledIOP mismatch in exported fields")
				}
			},
			serialize: func() ([]byte, error) { return serialization.SerializeCompiledIOP(comp) },
		},
		// "DefinitionInputs": {
		// 	value: lpp.definitionInputs,
		// 	typ:   reflect.TypeOf([]FilteredModuleInputs{}),
		// 	compare: func(t *testing.T, orig, deser interface{}) {
		// 		if !reflect.DeepEqual(orig, deser) {
		// 			t.Errorf("DefinitionInputs mismatch: got %+v, want %+v", deser, orig)
		// 		}
		// 	},
		// },
		"InitialFiatShamirState": {
			value: lpp.InitialFiatShamirState,
			typ:   reflect.TypeOf((*ifaces.Column)(nil)).Elem(),
			compare: func(t *testing.T, orig, deser interface{}) {
				if orig == nil && deser == nil {
					return
				}
				if orig == nil || deser == nil {
					t.Errorf("InitialFiatShamirState mismatch: one is nil")
					return
				}
				origCol := orig.(ifaces.Column)
				deserCol := deser.(ifaces.Column)
				if origCol.GetColID() != deserCol.GetColID() {
					t.Errorf("InitialFiatShamirState ColID mismatch: got %v, want %v", deserCol.GetColID(), origCol.GetColID())
				}
			},
		},
		"N0Hash": {
			value: lpp.N0Hash,
			typ:   reflect.TypeOf((*ifaces.Column)(nil)).Elem(),
			compare: func(t *testing.T, orig, deser interface{}) {
				if orig == nil && deser == nil {
					return
				}
				if orig == nil || deser == nil {
					t.Errorf("N0Hash mismatch: one is nil")
					return
				}
				origCol := orig.(ifaces.Column)
				deserCol := deser.(ifaces.Column)
				if origCol.GetColID() != deserCol.GetColID() {
					t.Errorf("N0Hash ColID mismatch: got %v, want %v", deserCol.GetColID(), origCol.GetColID())
				}
			},
		},
		"N1Hash": {
			value: lpp.N1Hash,
			typ:   reflect.TypeOf((*ifaces.Column)(nil)).Elem(),
			compare: func(t *testing.T, orig, deser interface{}) {
				if orig == nil && deser == nil {
					return
				}
				if orig == nil || deser == nil {
					t.Errorf("N1Hash mismatch: one is nil")
					return
				}
				origCol := orig.(ifaces.Column)
				deserCol := deser.(ifaces.Column)
				if origCol.GetColID() != deserCol.GetColID() {
					t.Errorf("N1Hash ColID mismatch: got %v, want %v", deserCol.GetColID(), origCol.GetColID())
				}
			},
		},
		"LogDerivativeSum": {
			value: lpp.LogDerivativeSum,
			typ:   reflect.TypeOf(query.LogDerivativeSum{}),
			compare: func(t *testing.T, orig, deser interface{}) {
				if !reflect.DeepEqual(orig, deser) {
					t.Errorf("LogDerivativeSum mismatch: got %+v, want %+v", deser, orig)
				}
			},
		},
		"GrandProduct": {
			value: lpp.GrandProduct,
			typ:   reflect.TypeOf(query.GrandProduct{}),
			compare: func(t *testing.T, orig, deser interface{}) {
				if !reflect.DeepEqual(orig, deser) {
					t.Errorf("GrandProduct mismatch: got %+v, want %+v", deser, orig)
				}
			},
		},
		"Horner": {
			value: lpp.Horner,
			typ:   reflect.TypeOf(query.Horner{}),
			compare: func(t *testing.T, orig, deser interface{}) {
				if !reflect.DeepEqual(orig, deser) {
					t.Errorf("Horner mismatch: got %+v, want %+v", deser, orig)
				}
			},
		},
	}

	info, ok := fields[fieldName]
	if !ok {
		t.Fatalf("Unknown field: %s", fieldName)
	}

	// Serialize the field
	var ser []byte
	if info.serialize != nil {
		var err error
		ser, err = info.serialize()
		if err != nil {
			t.Fatalf("Failed to serialize %s: %v", fieldName, err)
		}
	} else if info.value == nil || reflect.ValueOf(info.value).IsZero() {
		ser = []byte(serialization.NilString)
	} else {
		var err error
		ser, err = serialization.SerializeValue(reflect.ValueOf(info.value), serialization.DeclarationMode)
		if err != nil {
			t.Fatalf("Failed to serialize %s: %v", fieldName, err)
		}
	}

	// Deserialize the field
	var deserVal reflect.Value
	if bytes.Equal(ser, []byte(serialization.NilString)) {
		deserVal = reflect.Zero(info.typ)
	} else {
		var err error
		deserVal, err = serialization.DeserializeValue(ser, serialization.DeclarationMode, info.typ, comp)
		if err != nil {
			t.Fatalf("Failed to deserialize %s: %v", fieldName, err)
		}
	}

	// Compare original and deserialized values
	info.compare(t, info.value, deserVal.Interface())
} */
