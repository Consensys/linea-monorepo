package serialization

import (
	"bytes"
	"encoding/json"
	"fmt"
	"reflect"
	"sort"
	"sync"
	"unicode"

	"github.com/consensys/linea-monorepo/prover/protocol/column"
	"github.com/consensys/linea-monorepo/prover/protocol/dedicated"
	"github.com/consensys/linea-monorepo/prover/protocol/ifaces"
	"github.com/consensys/linea-monorepo/prover/protocol/wizard"
	"github.com/consensys/linea-monorepo/prover/symbolic"
	"github.com/consensys/linea-monorepo/prover/utils"
	"github.com/iancoleman/strcase"
	"github.com/sirupsen/logrus"
)

type Mode int

const (
	DeclarationMode Mode = iota
	ReferenceMode
	pureExprMode
)

const (
	NilString = "null"
)

var (
	columnType          = reflect.TypeOf((*ifaces.Column)(nil)).Elem()
	queryType           = reflect.TypeOf((*ifaces.Query)(nil)).Elem()
	naturalType         = reflect.TypeOf(column.Natural{})
	metadataType        = reflect.TypeOf((*symbolic.Metadata)(nil)).Elem()
	manuallyShiftedType = reflect.TypeOf(&dedicated.ManuallyShifted{})
)

type structFieldCache struct {
	fields []structField
}

type structField struct {
	name      string
	rawName   string
	fieldType reflect.Type
}

var (
	structCacheMu sync.RWMutex
	structCache   = make(map[reflect.Type]*structFieldCache)
)

func getStructFieldCache(t reflect.Type) *structFieldCache {
	structCacheMu.RLock()
	if cache, ok := structCache[t]; ok {
		structCacheMu.RUnlock()
		return cache
	}
	structCacheMu.RUnlock()

	numFields := t.NumField()
	fields := make([]structField, numFields)
	for i := 0; i < numFields; i++ {
		f := t.Field(i)
		fields[i] = structField{
			name:      f.Name,
			rawName:   strcase.ToLowerCamel(f.Name),
			fieldType: f.Type,
		}
	}
	sort.Slice(fields, func(i, j int) bool {
		return fields[i].rawName < fields[j].rawName
	})

	cache := &structFieldCache{fields: fields}
	structCacheMu.Lock()
	structCache[t] = cache
	structCacheMu.Unlock()
	return cache
}

func SerializeValue(v reflect.Value, mode Mode) (json.RawMessage, error) {
	if !v.IsValid() || v.Interface() == nil {
		return json.RawMessage(NilString), nil
	}

	switch v.Kind() {
	case reflect.Bool, reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32,
		reflect.Int64, reflect.Uint, reflect.Uint8, reflect.Uint16,
		reflect.Uint32, reflect.Uint64, reflect.String:
		return serializePrimitive(v)
	case reflect.Array, reflect.Slice:
		return serializeArrayOrSlice(v, mode)
	case reflect.Interface:
		return serializeInterface(v, mode)
	case reflect.Map:
		return serializeMap(v, mode)
	case reflect.Pointer:
		return SerializeValue(v.Elem(), mode)
	case reflect.Struct:
		return serializeStruct(v, mode)
	default:
		return nil, fmt.Errorf("unsupported type kind: %v", v.Kind())
	}
}

func DeserializeValue(data json.RawMessage, mode Mode, t reflect.Type, comp *wizard.CompiledIOP) (reflect.Value, error) {
	if bytes.Equal(data, []byte(NilString)) {
		return reflect.New(t), nil
	}

	switch t.Kind() {
	case reflect.Bool, reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32,
		reflect.Int64, reflect.Uint, reflect.Uint8, reflect.Uint16,
		reflect.Uint32, reflect.Uint64, reflect.String:
		return deserializePrimitive(data, t)
	case reflect.Array, reflect.Slice:
		return deserializeArrayOrSlice(data, mode, t, comp)
	case reflect.Map:
		return deserializeMap(data, mode, t, comp)
	case reflect.Interface:
		return deserializeInterface(data, mode, t, comp)
	case reflect.Pointer:
		return deserializePointer(data, mode, t, comp)
	case reflect.Struct:
		return deserializeStruct(data, mode, t, comp)
	default:
		return reflect.Value{}, fmt.Errorf("unsupported type kind: %v", t.Kind())
	}
}

func serializePrimitive(v reflect.Value) (json.RawMessage, error) {
	return serializeAnyWithCborPkg(v.Interface())
}

func serializeArrayOrSlice(v reflect.Value, mode Mode) (json.RawMessage, error) {
	length := v.Len()
	raw := make([]json.RawMessage, 0, length)
	const batchSize = 64

	for i := 0; i < length; i += batchSize {
		end := i + batchSize
		if end > length {
			end = length
		}
		for j := i; j < end; j++ {
			r, err := SerializeValue(v.Index(j), mode)
			if err != nil {
				return nil, fmt.Errorf("could not serialize position %d of type %q: %w", j, v.Type().String(), err)
			}
			raw = append(raw, r)
		}
	}
	return serializeAnyWithCborPkg(raw)
}

func serializeInterface(v reflect.Value, mode Mode) (json.RawMessage, error) {
	if mode == pureExprMode && v.Type() == metadataType {
		m := v.Interface().(symbolic.Metadata)
		stringVar := symbolic.StringVar(m.String())
		return SerializeValue(reflect.ValueOf(stringVar), mode)
	}

	if mode == DeclarationMode && v.Type() == columnType {
		return serializeColumnInterface(v, mode)
	}

	concrete := v.Elem()
	rawValue, err := SerializeValue(concrete, mode)
	if err != nil {
		return nil, fmt.Errorf("could not serialize interface value of type %q: %w", v.Type().String(), err)
	}

	// Useful for DEBUG purposes
	concreteType := getPkgPathAndTypeNameIndirect(concrete.Interface())
	_, err = findRegisteredImplementation(concreteType)
	if err != nil {
		logrus.Warnf("MISSING concrete type in implementation registry:%s \n", concreteType)
	}

	raw := map[string]interface{}{
		"type":  concreteType,
		"value": rawValue,
	}
	return serializeAnyWithCborPkg(raw)
}

// serializeColumnInterface handles serialization of column interfaces in DeclarationMode.
// Column types can be either Natural or ManuallyShifted
func serializeColumnInterface(v reflect.Value, mode Mode) (json.RawMessage, error) {
	concrete := v.Elem()

	var data json.RawMessage
	var err error
	switch concrete.Type() {
	case naturalType:
		col := v.Interface().(column.Natural)
		decl := intoSerializableColDecl(&col)
		data, err = SerializeValue(reflect.ValueOf(decl), mode)
		if err != nil {
			return nil, err
		}

	case manuallyShiftedType:
		shifted := v.Interface().(*dedicated.ManuallyShifted)
		decl := intoSerializableManuallyShifted(shifted)
		data, err = SerializeValue(reflect.ValueOf(decl), mode)
		if err != nil {
			return nil, err
		}

	default:
		return nil, fmt.Errorf("unsupported column type in DeclarationMode: %s", concrete.Type().String())
	}

	raw := map[string]interface{}{
		"type":  concrete.Type().String(),
		"value": data,
	}

	return serializeAnyWithCborPkg(raw)
}

func serializeMap(v reflect.Value, mode Mode) (json.RawMessage, error) {
	keys := v.MapKeys()
	keyStrings := make([]string, 0, len(keys))
	keyMap := make(map[string]reflect.Value, len(keys))

	for _, k := range keys {
		keyString, err := castAsString(k)
		if err != nil {
			return nil, fmt.Errorf("invalid map key type %q: %w", v.Type().Key().String(), err)
		}
		keyStrings = append(keyStrings, keyString)
		keyMap[keyString] = k
	}
	sort.Strings(keyStrings)

	raw := make(map[string]json.RawMessage, len(keys))
	for _, keyString := range keyStrings {
		k := keyMap[keyString]
		r, err := SerializeValue(v.MapIndex(k), mode)
		if err != nil {
			return nil, fmt.Errorf("could not serialize map key %q: %w", keyString, err)
		}
		raw[keyString] = r
	}
	return serializeAnyWithCborPkg(raw)
}

func serializeStruct(v reflect.Value, mode Mode) (json.RawMessage, error) {
	typeOfV := v.Type()

	if mode == ReferenceMode && typeOfV == naturalType {
		colID := v.Interface().(column.Natural).ID
		return serializeAnyWithCborPkg(colID)
	}

	if mode == DeclarationMode && typeOfV == naturalType {
		col := v.Interface().(column.Natural)
		decl := intoSerializableColDecl(&col)
		return SerializeValue(reflect.ValueOf(decl), mode)
	}

	if mode == DeclarationMode && typeOfV == manuallyShiftedType {
		shifted := v.Interface().(*dedicated.ManuallyShifted)
		decl := intoSerializableManuallyShifted(shifted)
		return SerializeValue(reflect.ValueOf(decl), mode)
	}

	if mode == ReferenceMode && typeOfV.Implements(queryType) {
		queryID := v.Interface().(ifaces.Query).Name()
		return serializeAnyWithCborPkg(queryID)
	}

	newMode := mode
	if typeOfV.Implements(queryType) || typeOfV.Implements(columnType) {
		newMode = ReferenceMode
	}

	cache := getStructFieldCache(typeOfV)
	raw := make(map[string]json.RawMessage, len(cache.fields))
	for _, f := range cache.fields {
		if unicode.IsLower(rune(f.name[0])) {
			continue // Skip unexported fields
		}

		r, err := SerializeValue(v.FieldByName(f.name), newMode)
		if err != nil {
			return nil, fmt.Errorf("could not serialize struct field %v.%v: %w", typeOfV.String(), f.rawName, err)
		}
		raw[f.rawName] = r
	}
	return serializeAnyWithCborPkg(raw)
}

func deserializePrimitive(data json.RawMessage, t reflect.Type) (reflect.Value, error) {
	v := reflect.New(t).Elem()
	if err := deserializeAnyWithCborPkg(data, v.Addr().Interface()); err != nil {
		return reflect.Value{}, err
	}
	return v, nil
}

func deserializeArrayOrSlice(data json.RawMessage, mode Mode, t reflect.Type, comp *wizard.CompiledIOP) (reflect.Value, error) {
	var raw []json.RawMessage
	if err := deserializeAnyWithCborPkg(data, &raw); err != nil {
		return reflect.Value{}, fmt.Errorf("failed to deserialize to %q: %w", t.Name(), err)
	}

	var v reflect.Value
	if t.Kind() == reflect.Array {
		v = reflect.New(t).Elem()
		if t.Len() != len(raw) {
			return reflect.Value{}, fmt.Errorf("failed to deserialize to %q, size mismatch: %d != %d", t.Name(), len(raw), v.Len())
		}
	} else {
		v = reflect.MakeSlice(t, len(raw), len(raw))
	}

	if v == (reflect.Value{}) {
		panic("slice value cannot be empty")
	}

	subType := t.Elem()
	const batchSize = 64
	for i := 0; i < len(raw); i += batchSize {
		end := i + batchSize
		if end > len(raw) {
			end = len(raw)
		}
		for j := i; j < end; j++ {
			subV, err := DeserializeValue(raw[j], mode, subType, comp)
			if err != nil {
				return reflect.Value{}, fmt.Errorf("failed to deserialize to %q: could not deserialize entry %v: %w", t.Name(), j, err)
			}
			v.Index(j).Set(subV)
		}
	}
	return v, nil
}

func deserializeMap(data json.RawMessage, mode Mode, t reflect.Type, comp *wizard.CompiledIOP) (reflect.Value, error) {
	keyType := t.Key()
	if keyType.Kind() != reflect.String {
		return reflect.Value{}, fmt.Errorf("cannot deserialize a map with non-string keys: %q", t.Name())
	}

	var raw map[string]json.RawMessage
	if err := deserializeAnyWithCborPkg(data, &raw); err != nil {
		return reflect.Value{}, fmt.Errorf("failed to deserialize to %q: %w", t.Name(), err)
	}

	v := reflect.MakeMap(t)
	valueType := t.Elem()
	for keyRaw, valRaw := range raw {
		val, err := DeserializeValue(valRaw, mode, valueType, comp)
		if err != nil {
			return reflect.Value{}, fmt.Errorf("failed to deserialize map field %v of type %v: %w", keyRaw, valueType.Name(), err)
		}
		key := reflect.ValueOf(keyRaw).Convert(keyType)
		v.SetMapIndex(key, val)
	}
	return v, nil
}

func deserializeInterface(data json.RawMessage, mode Mode, t reflect.Type, comp *wizard.CompiledIOP) (reflect.Value, error) {
	ifaceValue := reflect.New(t).Elem()

	if mode == pureExprMode && t == metadataType {
		var stringVar symbolic.StringVar
		result, err := DeserializeValue(data, mode, reflect.TypeOf(stringVar), comp)
		if err != nil {
			return reflect.Value{}, err
		}
		return result, nil
	}

	var raw struct {
		Type  string          `json:"type"`
		Value json.RawMessage `json:"value"`
	}
	if err := deserializeAnyWithCborPkg(data, &raw); err != nil {
		return reflect.Value{}, fmt.Errorf("failed to deserialize interface %q: %w", t.Name(), err)
	}

	switch raw.Type {
	case "column.Natural":
		rawType := reflect.TypeOf(&serializableColumnDecl{})
		v, err := DeserializeValue(raw.Value, mode, rawType, comp)
		if err != nil {
			return reflect.Value{}, fmt.Errorf("could not deserialize column interface declaration: %w", err)
		}
		nat := v.Interface().(*serializableColumnDecl).intoNaturalAndRegister(comp)
		ifaceValue.Set(reflect.ValueOf(nat))
		return ifaceValue, nil
	case "*dedicated.ManuallyShifted":
		rawType := reflect.TypeOf(&serializableManuallyShifted{})
		v, err := DeserializeValue(raw.Value, mode, rawType, comp)
		if err != nil {
			return reflect.Value{}, fmt.Errorf("could not deserialize ManuallyShifted: %w", err)
		}
		shifted := v.Interface().(*serializableManuallyShifted).intoManuallyShifted(comp)
		ifaceValue.Set(reflect.ValueOf(shifted))
		return ifaceValue, nil
	default:
		concreteType, err := findRegisteredImplementation(raw.Type)
		if err != nil {
			return reflect.Value{}, fmt.Errorf("failed to find concrete type and deserialize interface %q: %w", t.Name(), err)
		}

		v, err := DeserializeValue(raw.Value, mode, concreteType, comp)
		if err != nil {
			return reflect.Value{}, fmt.Errorf("failed to deserialize interface %q: %w", t.Name(), err)
		}
		ifaceValue.Set(v)
		return ifaceValue, nil
	}
}

func deserializePointer(data json.RawMessage, mode Mode, t reflect.Type, comp *wizard.CompiledIOP) (reflect.Value, error) {
	v, err := DeserializeValue(data, mode, t.Elem(), comp)
	if err != nil {
		return reflect.Value{}, fmt.Errorf("failed to deserialize pointer-type %q: %w", t.Name(), err)
	}
	return v.Addr(), nil
}

func deserializeStruct(data json.RawMessage, mode Mode, t reflect.Type, comp *wizard.CompiledIOP) (reflect.Value, error) {
	// Handle ReferenceMode for naturalType
	if mode == ReferenceMode && t == naturalType {
		var colID ifaces.ColID
		if err := deserializeAnyWithCborPkg(data, &colID); err != nil {
			return reflect.Value{}, fmt.Errorf("could not deserialize column ID: %w", err)
		}
		return reflect.ValueOf(comp.Columns.GetHandle(colID)), nil
	}

	// Handle DeclarationMode for naturalType
	if mode == DeclarationMode && t == naturalType {
		rawType := reflect.TypeOf(&serializableColumnDecl{})
		v, err := DeserializeValue(data, mode, rawType, comp)
		if err != nil {
			return reflect.Value{}, fmt.Errorf("could not deserialize column interface declaration: %w", err)
		}
		return reflect.ValueOf(v.Interface().(*serializableColumnDecl).intoNaturalAndRegister(comp)), nil
	}

	// Handle DeclarationMode for manuallyShiftedType
	if mode == DeclarationMode && t == manuallyShiftedType {
		rawType := reflect.TypeOf(&serializableManuallyShifted{})
		v, err := DeserializeValue(data, mode, rawType, comp)
		if err != nil {
			return reflect.Value{}, fmt.Errorf("could not deserialize column interface declaration: %w", err)
		}
		return reflect.ValueOf(v.Interface().(*serializableManuallyShifted).intoManuallyShifted(comp)), nil
	}

	// Handle ReferenceMode for queryType
	if mode == ReferenceMode && t.Implements(queryType) {
		var queryID ifaces.QueryID
		if err := deserializeAnyWithCborPkg(data, &queryID); err != nil {
			return reflect.Value{}, fmt.Errorf("could not deserialize query ID: %w", err)
		}
		if comp.QueriesParams.Exists(queryID) {
			return reflect.ValueOf(comp.QueriesParams.Data(queryID)), nil
		}
		if comp.QueriesNoParams.Exists(queryID) {
			return reflect.ValueOf(comp.QueriesNoParams.Data(queryID)), nil
		}
		utils.Panic("Could not find requested query ID: %v", queryID)
	}

	// Adjust mode for queryType or columnType
	newMode := mode
	if t.Implements(queryType) || t.Implements(columnType) {
		newMode = ReferenceMode
	}

	// Decode raw map
	var raw map[string]json.RawMessage
	if err := deserializeAnyWithCborPkg(data, &raw); err != nil {
		return reflect.Value{}, fmt.Errorf("could not deserialize struct type %q: %w", t.Name(), err)
	}

	// Deserialize fields
	v := reflect.New(t).Elem()
	cache := getStructFieldCache(t)

	for _, f := range cache.fields {
		if unicode.IsLower(rune(f.name[0])) {
			continue // Skip unexported fields
		}

		fieldRaw, ok := raw[f.rawName]
		if !ok {
			utils.Panic("Missing struct field %q.%v of type %q, provided sub-JSON: %v", t.String(), f.name, f.fieldType.String(), raw)
		}

		fieldValue, err := DeserializeValue(fieldRaw, newMode, f.fieldType, comp)
		if err != nil {
			utils.Panic("Could not deserialize struct field %q.%v of type %q: %v", t.String(), f.name, f.fieldType.String(), err)
		}

		// Handle pointer types
		if fieldValue.Type().Kind() == reflect.Ptr && fieldValue.Type().Elem() == f.fieldType {
			fieldValue = fieldValue.Elem()
		}

		v.FieldByName(f.name).Set(fieldValue)
	}

	return v, nil
}
