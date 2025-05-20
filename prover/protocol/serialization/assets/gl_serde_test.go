package assets

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

var gl = dw.GLs[0]

// rawModuleGL represents the serialized form of ModuleGL.
type rawModuleGL struct {
	CompiledIOP              json.RawMessage `json:"compiledIOP"`
	Disc                     json.RawMessage `json:"disc"`
	DefinitionInput          json.RawMessage `json:"definitionInput"`
	IsFirst                  json.RawMessage `json:"isFirst"`
	IsLast                   json.RawMessage `json:"isLast"`
	SentValuesGlobal         json.RawMessage `json:"sentValuesGlobal"`
	SentValuesGlobalHash     json.RawMessage `json:"sentValuesGlobalHash"`
	SentValuesGlobalMap      json.RawMessage `json:"sentValuesGlobalMap"`
	ReceivedValuesGlobal     json.RawMessage `json:"receivedValuesGlobal"`
	ReceivedValuesGlobalAccs json.RawMessage `json:"receivedValuesGlobalAccs"`
	ReceivedValuesGlobalHash json.RawMessage `json:"receivedValuesGlobalHash"`
	ReceivedValuesGlobalMap  json.RawMessage `json:"receivedValuesGlobalMap"`
}

// TestSerdeGL tests full serialization and deserialization of a ModuleGL.
func TestSerdeGL(t *testing.T) {
	serializedData, err := SerializeModuleGL(gl)
	if err != nil {
		t.Fatalf("Failed to serialize ModuleGL: %v", err)
	}

	deserializedGL, err := DeserializeModuleGL(serializedData)
	if err != nil {
		t.Fatalf("Failed to deserialize ModuleGL: %v", err)
	}

	if !test_utils.CompareExportedFields(gl, deserializedGL) {
		t.Errorf("Mismatch in exported fields after full ModuleGL serde")
	}
}

// SerializeModuleGL serializes a single ModuleGL instance field-by-field.
func SerializeModuleGL(gl *distributed.ModuleGL) ([]byte, error) {
	if gl == nil {
		return []byte(serialization.NilString), nil
	}

	raw := &rawModuleGL{}

	// Serialize CompiledIOP
	comp := gl.GetModuleTranslator().Wiop
	if comp == nil {
		return nil, fmt.Errorf("ModuleGL has nil CompiledIOP")
	}

	compSer, err := serialization.SerializeCompiledIOP(comp)
	if err != nil {
		return nil, fmt.Errorf("failed to serialize CompiledIOP: %w", err)
	}
	raw.CompiledIOP = compSer

	// Serialize Disc
	disc := gl.GetModuleTranslator().Disc
	serComp, err := SerializeDisc(disc)
	if err != nil {
		return nil, fmt.Errorf("failed to serialize GL module discoverer: %w", err)
	}
	raw.Disc = serComp

	// Serialize IsFirst
	if gl.IsFirst != nil {
		isFirstSer, err := serialization.SerializeValue(reflect.ValueOf(gl.IsFirst), serialization.DeclarationMode)
		if err != nil {
			return nil, fmt.Errorf("failed to serialize IsFirst: %w", err)
		}
		raw.IsFirst = isFirstSer
	} else {
		raw.IsFirst = []byte(serialization.NilString)
	}

	// Serialize IsLast
	if gl.IsLast != nil {
		isLastSer, err := serialization.SerializeValue(reflect.ValueOf(gl.IsLast), serialization.DeclarationMode)
		if err != nil {
			return nil, fmt.Errorf("failed to serialize IsLast: %w", err)
		}
		raw.IsLast = isLastSer
	} else {
		raw.IsLast = []byte(serialization.NilString)
	}

	// Serialize SentValuesGlobal
	if len(gl.SentValuesGlobal) > 0 {
		svgSer, err := serialization.SerializeValue(reflect.ValueOf(gl.SentValuesGlobal), serialization.DeclarationMode)
		if err != nil {
			return nil, fmt.Errorf("failed to serialize SentValuesGlobal: %w", err)
		}
		raw.SentValuesGlobal = svgSer
	} else {
		raw.SentValuesGlobal = []byte(serialization.NilString)
	}

	// Serialize SentValuesGlobalHash
	if gl.SentValuesGlobalHash != nil {
		svgHashSer, err := serialization.SerializeValue(reflect.ValueOf(gl.SentValuesGlobalHash), serialization.DeclarationMode)
		if err != nil {
			return nil, fmt.Errorf("failed to serialize SentValuesGlobalHash: %w", err)
		}
		raw.SentValuesGlobalHash = svgHashSer
	} else {
		raw.SentValuesGlobalHash = []byte(serialization.NilString)
	}

	// Serialize SentValuesGlobalMap
	if len(gl.SentValuesGlobalMap) > 0 {
		svgMapSer, err := serialization.SerializeValue(reflect.ValueOf(gl.SentValuesGlobalMap), serialization.DeclarationMode)
		if err != nil {
			return nil, fmt.Errorf("failed to serialize SentValuesGlobalMap: %w", err)
		}
		raw.SentValuesGlobalMap = svgMapSer
	} else {
		raw.SentValuesGlobalMap = []byte(serialization.NilString)
	}

	// Serialize ReceivedValuesGlobal
	if gl.ReceivedValuesGlobal != nil {
		rvgSer, err := serialization.SerializeValue(reflect.ValueOf(gl.ReceivedValuesGlobal), serialization.DeclarationMode)
		if err != nil {
			return nil, fmt.Errorf("failed to serialize ReceivedValuesGlobal: %w", err)
		}
		raw.ReceivedValuesGlobal = rvgSer
	} else {
		raw.ReceivedValuesGlobal = []byte(serialization.NilString)
	}

	// Serialize ReceivedValuesGlobalAccs
	if len(gl.ReceivedValuesGlobalAccs) > 0 {
		rvgAccsSer, err := serialization.SerializeValue(reflect.ValueOf(gl.ReceivedValuesGlobalAccs), serialization.DeclarationMode)
		if err != nil {
			return nil, fmt.Errorf("failed to serialize ReceivedValuesGlobalAccs: %w", err)
		}
		raw.ReceivedValuesGlobalAccs = rvgAccsSer
	} else {
		raw.ReceivedValuesGlobalAccs = []byte(serialization.NilString)
	}

	// Serialize ReceivedValuesGlobalHash
	if gl.ReceivedValuesGlobalHash != nil {
		rvgHashSer, err := serialization.SerializeValue(reflect.ValueOf(gl.ReceivedValuesGlobalHash), serialization.DeclarationMode)
		if err != nil {
			return nil, fmt.Errorf("failed to serialize ReceivedValuesGlobalHash: %w", err)
		}
		raw.ReceivedValuesGlobalHash = rvgHashSer
	} else {
		raw.ReceivedValuesGlobalHash = []byte(serialization.NilString)
	}

	// Serialize ReceivedValuesGlobalMap
	if len(gl.ReceivedValuesGlobalMap) > 0 {
		rvgMapSer, err := serialization.SerializeValue(reflect.ValueOf(gl.ReceivedValuesGlobalMap), serialization.DeclarationMode)
		if err != nil {
			return nil, fmt.Errorf("failed to serialize ReceivedValuesGlobalMap: %w", err)
		}
		raw.ReceivedValuesGlobalMap = rvgMapSer
	} else {
		raw.ReceivedValuesGlobalMap = []byte(serialization.NilString)
	}

	return serialization.SerializeAnyWithCborPkg(raw)
}

// DeserializeModuleGL deserializes a single ModuleGL instance field-by-field.
func DeserializeModuleGL(data []byte) (*distributed.ModuleGL, error) {
	if bytes.Equal(data, []byte(serialization.NilString)) {
		return nil, nil
	}

	var raw rawModuleGL
	if err := serialization.DeserializeAnyWithCborPkg(data, &raw); err != nil {
		return nil, fmt.Errorf("failed to deserialize ModuleGL raw data: %w", err)
	}

	// Initialize ModuleGL
	gl := &distributed.ModuleGL{}

	// Deserialize CompiledIOP
	comp, err := serialization.DeserializeCompiledIOP(raw.CompiledIOP)
	if err != nil {
		return nil, fmt.Errorf("failed to deserialize CompiledIOP: %w", err)
	}

	disc, err := DeserializeDisc(raw.Disc)
	if err != nil {
		return nil, fmt.Errorf("failed to deserialize GL module discoverer: %w", err)
	}
	gl.SetModuleTranslator(comp, disc)

	// Deserialize IsFirst
	if !bytes.Equal(raw.IsFirst, []byte(serialization.NilString)) {
		isFirstVal, err := serialization.DeserializeValue(raw.IsFirst, serialization.DeclarationMode, reflect.TypeOf(column.Natural{}), comp)
		if err != nil {
			return nil, fmt.Errorf("failed to deserialize IsFirst: %w", err)
		}
		gl.IsFirst = isFirstVal.Interface().(ifaces.Column)
	}

	// Deserialize IsLast
	if !bytes.Equal(raw.IsLast, []byte(serialization.NilString)) {
		isLastVal, err := serialization.DeserializeValue(raw.IsLast, serialization.DeclarationMode, reflect.TypeOf(column.Natural{}), comp)
		if err != nil {
			return nil, fmt.Errorf("failed to deserialize IsLast: %w", err)
		}
		gl.IsLast = isLastVal.Interface().(ifaces.Column)
	}

	// Deserialize SentValuesGlobal
	if !bytes.Equal(raw.SentValuesGlobal, []byte(serialization.NilString)) {
		svgVal, err := serialization.DeserializeValue(raw.SentValuesGlobal, serialization.DeclarationMode, reflect.TypeOf([]query.LocalOpening{}), comp)
		if err != nil {
			return nil, fmt.Errorf("failed to deserialize SentValuesGlobal: %w", err)
		}
		gl.SentValuesGlobal = svgVal.Interface().([]query.LocalOpening)
	}

	// Deserialize SentValuesGlobalHash
	if !bytes.Equal(raw.SentValuesGlobalHash, []byte(serialization.NilString)) {
		svgHashVal, err := serialization.DeserializeValue(raw.SentValuesGlobalHash, serialization.DeclarationMode, reflect.TypeOf(column.Natural{}), comp)
		if err != nil {
			return nil, fmt.Errorf("failed to deserialize SentValuesGlobalHash: %w", err)
		}
		gl.SentValuesGlobalHash = svgHashVal.Interface().(ifaces.Column)
	}

	// Deserialize SentValuesGlobalMap
	if !bytes.Equal(raw.SentValuesGlobalMap, []byte(serialization.NilString)) {
		svgMapVal, err := serialization.DeserializeValue(raw.SentValuesGlobalMap, serialization.DeclarationMode, reflect.TypeOf(map[string]int{}), comp)
		if err != nil {
			return nil, fmt.Errorf("failed to deserialize SentValuesGlobalMap: %w", err)
		}
		gl.SentValuesGlobalMap = svgMapVal.Interface().(map[string]int)
	}

	// Deserialize ReceivedValuesGlobal
	if !bytes.Equal(raw.ReceivedValuesGlobal, []byte(serialization.NilString)) {
		rvgVal, err := serialization.DeserializeValue(raw.ReceivedValuesGlobal, serialization.DeclarationMode, reflect.TypeOf(column.Natural{}), comp)
		if err != nil {
			return nil, fmt.Errorf("failed to deserialize ReceivedValuesGlobal: %w", err)
		}
		gl.ReceivedValuesGlobal = rvgVal.Interface().(ifaces.Column)
	}

	// Deserialize ReceivedValuesGlobalAccs
	if !bytes.Equal(raw.ReceivedValuesGlobalAccs, []byte(serialization.NilString)) {
		rvgAccsVal, err := serialization.DeserializeValue(raw.ReceivedValuesGlobalAccs, serialization.DeclarationMode, reflect.TypeOf([]ifaces.Accessor{}), comp)
		if err != nil {
			return nil, fmt.Errorf("failed to deserialize ReceivedValuesGlobalAccs: %w", err)
		}
		gl.ReceivedValuesGlobalAccs = rvgAccsVal.Interface().([]ifaces.Accessor)
	}

	// Deserialize ReceivedValuesGlobalHash
	if !bytes.Equal(raw.ReceivedValuesGlobalHash, []byte(serialization.NilString)) {
		rvgHashVal, err := serialization.DeserializeValue(raw.ReceivedValuesGlobalHash, serialization.DeclarationMode, reflect.TypeOf(column.Natural{}), comp)
		if err != nil {
			return nil, fmt.Errorf("failed to deserialize ReceivedValuesGlobalHash: %w", err)
		}
		gl.ReceivedValuesGlobalHash = rvgHashVal.Interface().(ifaces.Column)
	}

	// Deserialize ReceivedValuesGlobalMap
	if !bytes.Equal(raw.ReceivedValuesGlobalMap, []byte(serialization.NilString)) {
		rvgMapVal, err := serialization.DeserializeValue(raw.ReceivedValuesGlobalMap, serialization.DeclarationMode, reflect.TypeOf(map[string]int{}), comp)
		if err != nil {
			return nil, fmt.Errorf("failed to deserialize ReceivedValuesGlobalMap: %w", err)
		}
		gl.ReceivedValuesGlobalMap = rvgMapVal.Interface().(map[string]int)
	}

	return gl, nil
}

// SerializeModuleGLs serializes a slice of ModuleGL instances.
func SerializeModuleGLs(gls []*distributed.ModuleGL) ([]byte, error) {
	rawGLs := make([]json.RawMessage, len(gls))
	for i, gl := range gls {
		glSer, err := SerializeModuleGL(gl)
		if err != nil {
			return nil, fmt.Errorf("failed to serialize ModuleGL at index %d: %w", i, err)
		}
		rawGLs[i] = glSer
	}
	return serialization.SerializeAnyWithCborPkg(rawGLs)
}

// DeserializeModuleGLs deserializes a slice of ModuleGL instances.
func DeserializeModuleGLs(data []byte) ([]*distributed.ModuleGL, error) {
	var rawGLs []json.RawMessage
	if err := serialization.DeserializeAnyWithCborPkg(data, &rawGLs); err != nil {
		return nil, fmt.Errorf("failed to deserialize GLs raw slice: %w", err)
	}

	gls := make([]*distributed.ModuleGL, len(rawGLs))
	for i, raw := range rawGLs {
		gl, err := DeserializeModuleGL(raw)
		if err != nil {
			return nil, fmt.Errorf("failed to deserialize ModuleGL at index %d: %w", i, err)
		}
		gls[i] = gl
	}
	return gls, nil
}
