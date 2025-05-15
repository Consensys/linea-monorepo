package serdetests

import (
	"bytes"
	"encoding/json"
	"fmt"
	"reflect"
	"testing"

	"github.com/consensys/linea-monorepo/prover/protocol/column"
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
	Disc                   json.RawMessage `json:"disc"`
	DefinitionInputs       json.RawMessage `json:"definitionInputs"`
	InitialFiatShamirState json.RawMessage `json:"initialFiatShamirState"`
	N0Hash                 json.RawMessage `json:"n0Hash"`
	N1Hash                 json.RawMessage `json:"n1Hash"`
	LogDerivativeSum       json.RawMessage `json:"logDerivativeSum"`
	GrandProduct           json.RawMessage `json:"grandProduct"`
	Horner                 json.RawMessage `json:"horner"`
}

// TestSerdeLPP tests full serialization and deserialization of a ModuleLPP.
func TestSerdeLPP(t *testing.T) {
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

// serializeModuleLPP serializes a single ModuleLPP instance field-by-field.
func serializeModuleLPP(lpp *distributed.ModuleLPP) ([]byte, error) {
	if lpp == nil {
		return []byte(serialization.NilString), nil
	}

	raw := &rawModuleLPP{}

	// Serialize CompiledIOP first (includes Columns store)
	comp := lpp.GetModuleTranslator().Wiop
	if comp == nil {
		return nil, fmt.Errorf("ModuleLPP has nil CompiledIOP")
	}

	compSer, err := serialization.SerializeCompiledIOP(comp)
	if err != nil {
		return nil, fmt.Errorf("failed to serialize CompiledIOP: %w", err)
	}
	raw.CompiledIOP = compSer

	// Serialize disc
	disc := lpp.GetModuleTranslator().Disc
	serComp, err := SerializeDisc(disc)
	if err != nil {
		return nil, fmt.Errorf("failed to serialize LPP module discoverer:%w", err)
	}

	raw.Disc = serComp

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

	// Initialize ModuleLPP
	lpp := &distributed.ModuleLPP{}

	// Deserialize CompiledIOP first (includes Columns store)
	comp, err := serialization.DeserializeCompiledIOP(raw.CompiledIOP)
	if err != nil {
		return nil, fmt.Errorf("failed to deserialize CompiledIOP: %w", err)
	}

	disc, err := DeserializeDisc(raw.Disc)
	if err != nil {
		return nil, fmt.Errorf("failed to deserialize LPP module discoverer: %w", err)
	}
	lpp.SetModuleTranslator(comp, disc)

	// Deserialize InitialFiatShamirState (depends on Columns)
	if !bytes.Equal(raw.InitialFiatShamirState, []byte(serialization.NilString)) {
		ifsVal, err := serialization.DeserializeValue(raw.InitialFiatShamirState, serialization.DeclarationMode, reflect.TypeOf(column.Natural{}), comp)
		if err != nil {
			return nil, fmt.Errorf("failed to deserialize InitialFiatShamirState: %w", err)
		}
		lpp.InitialFiatShamirState = ifsVal.Interface().(ifaces.Column)
	}

	// Deserialize N0Hash (depends on Columns)
	if !bytes.Equal(raw.N0Hash, []byte(serialization.NilString)) {
		n0Val, err := serialization.DeserializeValue(raw.N0Hash, serialization.DeclarationMode, reflect.TypeOf(column.Natural{}), comp)
		if err != nil {
			return nil, fmt.Errorf("failed to deserialize N0Hash: %w", err)
		}
		lpp.N0Hash = n0Val.Interface().(ifaces.Column)
	}

	// Deserialize N1Hash (depends on Columns)
	if !bytes.Equal(raw.N1Hash, []byte(serialization.NilString)) {
		n1Val, err := serialization.DeserializeValue(raw.N1Hash, serialization.DeclarationMode, reflect.TypeOf(column.Natural{}), comp)
		if err != nil {
			return nil, fmt.Errorf("failed to deserialize N1Hash: %w", err)
		}
		lpp.N1Hash = n1Val.Interface().(ifaces.Column)
	}

	// Deserialize LogDerivativeSum
	if !bytes.Equal(raw.LogDerivativeSum, []byte(serialization.NilString)) {
		// comp = lpp.GetModuleTranslator().Wiop
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
// testField handles serialization, deserialization, and comparison for a single field.
func testField(t *testing.T, fieldName string, fieldValue interface{}, fieldType reflect.Type) {
	if fieldValue == nil {
		t.Fatalf("%s field is nil", fieldName)
	}

	comp := serialization.NewEmptyCompiledIOP()
	serializedData, err := serializeValue(fieldValue)
	if err != nil {
		t.Fatalf("Failed to serialize %s: %v", fieldName, err)
	}

	deserializedVal, err := serialization.DeserializeValue(serializedData, serialization.DeclarationMode, fieldType, comp)
	if err != nil {
		t.Fatalf("Failed to deserialize %s: %v", fieldName, err)
	}

	if !test_utils.CompareExportedFields(fieldValue, deserializedVal.Interface()) {
		t.Errorf("Mismatch in exported fields for %s", fieldName)
	}
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
	if lpp.InitialFiatShamirState == nil {
		t.Fatalf("InitialFiatShamirState is nil")
	}
	fmt.Printf("Concrete type of InitialFiatShamirState: %T\n", lpp.InitialFiatShamirState)
	testField(t, "InitialFiatShamirState", lpp.InitialFiatShamirState, reflect.TypeOf(column.Natural{}))
}

// TestLPPN0Hash tests serialization and deserialization of the N0Hash field.
func TestLPPN0Hash(t *testing.T) {
	testField(t, "N0Hash", lpp.N0Hash, reflect.TypeOf(column.Natural{}))
}

// TestLPPN1Hash tests serialization and deserialization of the N1Hash field.
func TestLPPN1Hash(t *testing.T) {
	testField(t, "N1Hash", lpp.N1Hash, reflect.TypeOf(column.Natural{}))
}

// TestLPPLogDerivativeSum tests serialization and deserialization of the LogDerivativeSum field.
func TestLPPLogDerivativeSum(t *testing.T) {
	testField(t, "LogDerivativeSum", lpp.LogDerivativeSum, reflect.TypeOf(query.LogDerivativeSum{}))
}

// func TestLPPLogDerivativeSum(t *testing.T) {
// 	originalComp := lpp.GetModuleTranslator().Wiop
// 	if originalComp == nil {
// 		t.Fatal("CompiledIOP is nil in lpp")
// 	}
// 	testField(t, "LogDerivativeSum", lpp.LogDerivativeSum, reflect.TypeOf(query.LogDerivativeSum{}), originalComp)
//}

// TestLPPGrandProduct tests serialization and deserialization of the GrandProduct field.
func TestLPPGrandProduct(t *testing.T) {
	testField(t, "GrandProduct", lpp.GrandProduct, reflect.TypeOf(query.GrandProduct{}))
}

// TestLPPHorner tests serialization and deserialization of the Horner field.
func TestLPPHorner(t *testing.T) {
	testField(t, "Horner", lpp.Horner, reflect.TypeOf(query.Horner{}))
}

*/
