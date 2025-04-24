package serialization

import (
	"bytes"
	"encoding/json"
	"fmt"
	"reflect"
	"unicode"

	"github.com/consensys/linea-monorepo/prover/protocol/column"
	"github.com/consensys/linea-monorepo/prover/protocol/ifaces"
	"github.com/consensys/linea-monorepo/prover/protocol/wizard"
	"github.com/consensys/linea-monorepo/prover/symbolic"
	"github.com/consensys/linea-monorepo/prover/utils"
	"github.com/iancoleman/strcase"
)

// Mode indicates whether the serializer is currently running over a part of
// the wizard in which is declaring columns, coins or queries or one that is
// doing references to them.
type Mode int

const (
	nilString = "null"
)

const (
	DeclarationMode Mode = iota
	ReferenceMode
	pureExprMode // pureExprMode is meant to serialize symbolic expression.
)

// Types that necessitate special handling by the de/serializer
var (
	columnType   = reflect.TypeOf((*ifaces.Column)(nil)).Elem()
	queryType    = reflect.TypeOf((*ifaces.Query)(nil)).Elem()
	naturalType  = reflect.TypeOf(column.Natural{})
	metadataType = reflect.TypeOf((*symbolic.Metadata)(nil)).Elem()
)

// SerializeValue recursively serializes `v` into JSON.
func SerializeValue(v reflect.Value, mode Mode) (json.RawMessage, error) {
	if v.Interface() == nil {
		return json.RawMessage(nilString), nil
	}

	switch v.Kind() {
	case reflect.Bool, reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32,
		reflect.Int64, reflect.Uint, reflect.Uint8, reflect.Uint16,
		reflect.Uint32, reflect.Uint64, reflect.String:
		return serializeAnyWithCborPkg(v.Interface())

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
		panic(fmt.Sprintf("unreachable: v=%++v", v))
	}
}

func serializeArrayOrSlice(v reflect.Value, mode Mode) (json.RawMessage, error) {
	length := v.Len()
	raw := make([]json.RawMessage, length)
	for i := 0; i < length; i++ {
		r, err := SerializeValue(v.Index(i), mode)
		if err != nil {
			return nil, fmt.Errorf("could not serialize position %d of type %q: %w", i, v.Type().String(), err)
		}
		raw[i] = r
	}
	return serializeAnyWithCborPkg(raw)
}

func serializeInterface(v reflect.Value, mode Mode) (json.RawMessage, error) {
	if mode == pureExprMode && v.Type() == metadataType {
		return serializeMetadata(v, mode)
	}

	if mode == DeclarationMode && v.Type() == columnType {
		return serializeColumn(v, mode)
	}

	concrete := v.Elem()
	rawValue, err := SerializeValue(concrete, mode)
	if err != nil {
		return nil, fmt.Errorf("could not serialize interface value of type `%v`: %w", v.Type().String(), err)
	}

	raw := map[string]any{
		"type":  getPkgPathAndTypeNameIndirect(concrete.Interface()),
		"value": rawValue,
	}

	return serializeAnyWithCborPkg(raw)
}

func serializeMetadata(v reflect.Value, mode Mode) (json.RawMessage, error) {
	m := v.Interface().(symbolic.Metadata)
	mString := m.String()
	stringVar := symbolic.StringVar(mString)
	stringVarValue := reflect.ValueOf(stringVar)
	return SerializeValue(stringVarValue, mode)
}

func serializeColumn(v reflect.Value, mode Mode) (json.RawMessage, error) {
	col := v.Interface().(column.Natural)
	raw := intoSerializableColDecl(&col)
	return SerializeValue(reflect.ValueOf(raw), mode)
}

func serializeMap(v reflect.Value, mode Mode) (json.RawMessage, error) {
	raw := map[string]json.RawMessage{}
	keys := v.MapKeys()
	for _, k := range keys {
		keyString, err := castAsString(k)
		if err != nil {
			return nil, fmt.Errorf("invalid map key type: `%v`: %w", v.Type().Key().String(), err)
		}
		value := v.MapIndex(k)
		r, err := SerializeValue(value, mode)
		if err != nil {
			return nil, fmt.Errorf("could not serialize map key `%v`: %w", keyString, err)
		}
		raw[keyString] = r
	}
	return serializeAnyWithCborPkg(raw)
}

func serializeStruct(v reflect.Value, mode Mode) (json.RawMessage, error) {
	typeOfV := v.Type()
	numFields := typeOfV.NumField()
	raw := map[string]json.RawMessage{}

	if mode == ReferenceMode && typeOfV == naturalType {
		colID := v.Interface().(column.Natural).ID
		return serializeAnyWithCborPkg(colID)
	}

	if mode == DeclarationMode && typeOfV == naturalType {
		col := v.Interface().(column.Natural)
		raw := intoSerializableColDecl(&col)
		return SerializeValue(reflect.ValueOf(raw), mode)
	}

	if mode == ReferenceMode && typeOfV.Implements(queryType) {
		queryID := v.Interface().(ifaces.Query).Name()
		return serializeAnyWithCborPkg(queryID)
	}

	if typeOfV.Implements(queryType) || typeOfV.Implements(columnType) {
		mode = ReferenceMode
	}

	for i := 0; i < numFields; i++ {
		fieldName := typeOfV.Field(i).Name
		rawFieldName := strcase.ToLowerCamel(fieldName)
		fieldValue := v.Field(i)

		if unicode.IsLower(rune(fieldName[0])) {
			utils.Panic("unexported field: struct=%v name=%v type=%v", typeOfV.String(), fieldName, fieldValue.Type().String())
		}

		r, err := SerializeValue(fieldValue, mode)
		if err != nil {
			return nil, fmt.Errorf("could not serialize struct field (%v).%v: %w", typeOfV.String(), rawFieldName, err)
		}
		raw[rawFieldName] = r
	}

	return serializeAnyWithCborPkg(raw)
}

// DeserializeValue recursively unmarshals `data` into a [reflect.Value] of
// type matching the caller's target type `t`.
func DeserializeValue(data json.RawMessage, mode Mode, t reflect.Type, comp *wizard.CompiledIOP) (reflect.Value, error) {
	if bytes.Equal(data, []byte(nilString)) {
		return reflect.New(t), nil
	}

	switch t.Kind() {
	default:
		return reflect.Value{}, fmt.Errorf("unsupported kind: %v", t.Kind().String())

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
	}
}

func deserializePrimitive(data json.RawMessage, t reflect.Type) (reflect.Value, error) {
	v := reflect.New(t).Elem()
	return v, deserializeAnyWithCborPkg(data, v.Addr().Interface())
}

func deserializeArrayOrSlice(data json.RawMessage, mode Mode, t reflect.Type, comp *wizard.CompiledIOP) (reflect.Value, error) {
	raw := []json.RawMessage{}
	if err := deserializeAnyWithCborPkg(data, &raw); err != nil {
		return reflect.Value{}, fmt.Errorf("failed to deserialize to a `%v` value: %w", t.Name(), err)
	}

	v := reflect.New(t).Elem()
	if t.Kind() == reflect.Array && t.Len() != len(raw) {
		return reflect.Value{}, fmt.Errorf("failed to deserialize to `%v`, size mismatch: %d != %d", t.Name(), len(raw), v.Len())
	}

	if t.Kind() == reflect.Slice {
		v = reflect.MakeSlice(t, len(raw), len(raw))
	}

	for i := range raw {
		subV, err := DeserializeValue(raw[i], mode, t.Elem(), comp)
		if err != nil {
			return reflect.Value{}, fmt.Errorf("failed to deserialize to `%v`: could not deserialize entry %v: %w", t.Name(), i, err)
		}
		v.Index(i).Set(subV)
	}

	return v, nil
}

func deserializeMap(data json.RawMessage, mode Mode, t reflect.Type, comp *wizard.CompiledIOP) (reflect.Value, error) {
	raw := map[string]json.RawMessage{}
	if err := deserializeAnyWithCborPkg(data, &raw); err != nil {
		return reflect.Value{}, fmt.Errorf("failed to deserialize to a `%v` value: %w", t.Name(), err)
	}

	v := reflect.MakeMap(t)
	for keyRaw, valRaw := range raw {
		val, err := DeserializeValue(valRaw, mode, t.Elem(), comp)
		if err != nil {
			return reflect.Value{}, fmt.Errorf("failed to deserialize map field %v of type %v: %w", keyRaw, t.Elem().Name(), err)
		}
		key := reflect.ValueOf(keyRaw).Convert(t.Key())
		v.SetMapIndex(key, val)
	}

	return v, nil
}

func deserializeInterface(data json.RawMessage, mode Mode, t reflect.Type, comp *wizard.CompiledIOP) (reflect.Value, error) {
	ifaceValue := reflect.New(t).Elem()

	if mode == pureExprMode && t == metadataType {
		var stringVar symbolic.StringVar
		return DeserializeValue(data, mode, reflect.TypeOf(stringVar), comp)
	}

	if mode == DeclarationMode && t == columnType {
		return deserializeColumn(data, mode, comp)
	}

	raw := struct {
		Type  string          `json:"type"`
		Value json.RawMessage `json:"value"`
	}{}

	if err := deserializeAnyWithCborPkg(data, &raw); err != nil {
		return reflect.Value{}, fmt.Errorf("failed to deserialize interface `%v`: %w", t.Name(), err)
	}

	concreteType, err := findRegisteredImplementation(raw.Type)
	if err != nil {
		return reflect.Value{}, fmt.Errorf("failed to deserialize interface `%v`: %w", t.Name(), err)
	}

	v, err := DeserializeValue(raw.Value, mode, concreteType, comp)
	if err != nil {
		return reflect.Value{}, fmt.Errorf("failed to deserialize interface `%v`: %w", t.Name(), err)
	}

	ifaceValue.Set(v)
	return ifaceValue, nil
}

func deserializeColumn(data json.RawMessage, mode Mode, comp *wizard.CompiledIOP) (reflect.Value, error) {
	rawType := reflect.TypeOf(&serializableColumnDecl{})
	v, err := DeserializeValue(data, mode, rawType, comp)
	if err != nil {
		return reflect.Value{}, fmt.Errorf("could not deserialize column interface declaration: %w", err)
	}
	nat := v.Interface().(*serializableColumnDecl).intoNaturalAndRegister(comp)
	ifaceValue := reflect.New(columnType).Elem()
	ifaceValue.Set(reflect.ValueOf(nat))
	return ifaceValue, nil
}

func deserializePointer(data json.RawMessage, mode Mode, t reflect.Type, comp *wizard.CompiledIOP) (reflect.Value, error) {
	v, err := DeserializeValue(data, mode, t.Elem(), comp)
	if err != nil {
		return reflect.Value{}, fmt.Errorf("failed to deserialize pointer-type `%v`: %w", t.Name(), err)
	}
	return v.Addr(), nil
}

func deserializeStruct(data json.RawMessage, mode Mode, t reflect.Type, comp *wizard.CompiledIOP) (reflect.Value, error) {
	if mode == ReferenceMode && t == naturalType {
		return deserializeNaturalColumn(data, comp)
	}

	if mode == ReferenceMode && t.Implements(queryType) {
		return deserializeQuery(data, comp)
	}

	if mode == DeclarationMode && t == naturalType {
		return deserializeColumn(data, mode, comp)
	}

	if t.Implements(queryType) || t.Implements(columnType) {
		mode = ReferenceMode
	}

	raw := map[string]json.RawMessage{}
	if err := deserializeAnyWithCborPkg(data, &raw); err != nil {
		return reflect.Value{}, fmt.Errorf("could note deserialize struct type `%v` : %w", t.Name(), err)
	}

	v := reflect.New(t).Elem()
	for i := 0; i < t.NumField(); i++ {
		fieldName := t.Field(i).Name
		rawFieldName := strcase.ToLowerCamel(fieldName)
		fieldType := t.Field(i).Type

		fieldRaw, ok := raw[rawFieldName]
		if !ok {
			utils.Panic("missing struct field `%v.%v` of type `%v`, the provided sub-JSON is %v", t.String(), fieldName, fieldType.String(), raw)
		}

		fieldValue, err := DeserializeValue(fieldRaw, mode, fieldType, comp)
		if err != nil {
			utils.Panic("could not deserialize struct field %v.%v of type %v: %v", t.String(), fieldName, fieldName, err)
		}

		v.Field(i).Set(fieldValue)
	}

	return v, nil
}

func deserializeNaturalColumn(data json.RawMessage, comp *wizard.CompiledIOP) (reflect.Value, error) {
	var colID ifaces.ColID
	if err := deserializeAnyWithCborPkg(data, &colID); err != nil {
		return reflect.Value{}, fmt.Errorf("could not deserialize column ID: %w", err)
	}
	nat := comp.Columns.GetHandle(colID)
	return reflect.ValueOf(nat), nil
}

func deserializeQuery(data json.RawMessage, comp *wizard.CompiledIOP) (reflect.Value, error) {
	var queryID ifaces.QueryID
	if err := deserializeAnyWithCborPkg(data, &queryID); err != nil {
		return reflect.Value{}, fmt.Errorf("could not deserialize column ID: %w", err)
	}

	if comp.QueriesParams.Exists(queryID) {
		q := comp.QueriesParams.Data(queryID)
		return reflect.ValueOf(q), nil
	}

	if comp.QueriesNoParams.Exists(queryID) {
		q := comp.QueriesNoParams.Data(queryID)
		return reflect.ValueOf(q), nil
	}

	utils.Panic("could not find requested column ID: %v", queryID)
	return reflect.Value{}, nil
}
