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

	// snapshot map: Aids in debugging to find missing concrete type implementation
	snapshot = make(map[string]struct{})
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

// SerializeValue recursively serializes `v` into JSON. This function is
// specially crafted to handle the special types occurring in the `protocol`
// directory. It can be used in `DeclarationMode` where the [column.Natural]
// are serialized as [serializableColumnDecl] objects or in the [referenceMode]
// where the [ifaces.Query] and [column.Natural] objects are serialized using
// only their ID.
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
	case reflect.Func:
		logrus.Debugf("Ignoring func type at value: %v", v)
		return json.RawMessage(NilString), nil
	default:
		return nil, fmt.Errorf("unsupported type kind: %v", v.Kind())
	}
}

// DeserializeValue recursively unmarshals `data` into a [reflect.Value] of
// type matching the caller's target type `t`. It reverses the work of
// [serializeValue]. It comes with the following subtleties:
//
//  1. `comp` of the function is mutated when run in [Declaration] mode over something
//     that contains column type as it will register the columns within the provided
//     `comp` object.
//  2. If run with [ReferenceMode], it will expect all the potential references
//     it finds to be references of objects that have been already unmarshalled
//     in declaration mode.
func DeserializeValue(data json.RawMessage, mode Mode, t reflect.Type, comp *wizard.CompiledIOP) (reflect.Value, error) {
	if bytes.Equal(data, []byte(NilString)) {
		return reflect.Zero(t), nil
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

	// fmt.Printf("Serializing Interface \n")
	typeOfV := v.Type()

	// Return nil string for types that are to be purposefully ignored.
	// For ex: Ignore Gnark circuit-related params
	if IsIgnoreableType(typeOfV) {
		return json.RawMessage(NilString), nil
	}

	if mode == pureExprMode && typeOfV == metadataType {
		m := v.Interface().(symbolic.Metadata)
		stringVar := symbolic.StringVar(m.String())
		return SerializeValue(reflect.ValueOf(stringVar), mode)
	}

	if mode == DeclarationMode && typeOfV == columnType {
		return serializeColumnInterface(v, mode)
	}

	// Explicit handling of nil pointers
	if v.IsNil() {
		return json.RawMessage(NilString), nil
	}

	concrete := v.Elem()
	rawValue, err := SerializeValue(concrete, mode)
	if err != nil {
		return nil, fmt.Errorf("could not serialize interface value of type %q: %w", v.Type().String(), err)
	}

	concreteTypeStr := getPkgPathAndTypeNameIndirect(concrete.Interface())
	concreteType, err := findRegisteredImplementation(concreteTypeStr)

	if err != nil && !IsIgnoreableType(concreteType) {
		// Check and warn if the concrete type is missing in the implementation registry
		if _, exists := snapshot[concreteTypeStr]; !exists {
			logrus.Warnf("MISSING concrete type in implementation registry: %s\n", concreteTypeStr)
			snapshot[concreteTypeStr] = struct{}{} // This is to stop recurrent warnings
		}
	}

	raw := map[string]interface{}{
		"type":  concreteTypeStr,
		"value": rawValue,
	}
	return serializeAnyWithCborPkg(raw)
}

// serializeStruct serializes a struct into JSON, ignoring certain fields based on the mode and type.
// Includes special handling for *wizard.CompiledIOP fields.
func serializeStruct(v reflect.Value, mode Mode) (json.RawMessage, error) {
	typeOfV := v.Type()

	// Skip serialization for ignorable types
	if IsIgnoreableType(typeOfV) {
		logrus.Warnf("Skipping serialization of Ignorable type: %v\n", typeOfV)
		return json.RawMessage(NilString), nil
	}

	// If we just want to encode a reference to a column that has been
	// explicitly registered in the store. It suffices to provide a
	// reference to the column. From that, we can load it in the store,
	// that's why there is no need to serialize the full object; the ID
	// is enough.
	if mode == ReferenceMode && typeOfV == naturalType {
		// logrus.Println("Ser. columns in ref. mode")
		colID := v.Interface().(column.Natural).ID
		return serializeAnyWithCborPkg(colID)
	}

	// Note that this is the same handling as in the interface-level
	// handling. This can be triggered when passing the Column as initial
	// input to the serializer.
	if mode == DeclarationMode && typeOfV == naturalType {
		// logrus.Println("Ser. columns in declaration mode")
		col := v.Interface().(column.Natural)
		decl := intoSerializableColDecl(&col)
		return SerializeValue(reflect.ValueOf(decl), mode)
	}

	if mode == DeclarationMode && typeOfV == manuallyShiftedType {
		shifted := v.Interface().(*dedicated.ManuallyShifted)
		decl := intoSerializableManuallyShifted(shifted)
		return SerializeValue(reflect.ValueOf(decl), mode)
	}

	// Likewise for if the type implements query, we can just deserialize
	// looking up the query by name in the compiled IOP. So there is no need
	// to include it in the serialized object. The ID is enough.
	if mode == ReferenceMode && typeOfV.Implements(queryType) {
		//logrus.Println("Ser. query in ref. mode")
		queryID := v.Interface().(ifaces.Query).Name()
		return serializeAnyWithCborPkg(queryID)
	}

	// If 'v' implements query or column, it may be the case that its
	// components refer to other queries or columns. In that case, we should
	// ensure that the serializer enters in reference mode.
	newMode := mode
	if typeOfV.Implements(queryType) || typeOfV.Implements(columnType) {
		// logrus.Printf("Setting newmode columns from mode:%d in ref. mode \n", newMode)
		newMode = ReferenceMode
	}

	// Serialize fields
	cache := getStructFieldCache(typeOfV)
	raw := make(map[string]json.RawMessage, len(cache.fields))
	for _, f := range cache.fields {
		if unicode.IsLower(rune(f.name[0])) {
			continue // Skip unexported fields
		}

		fieldType := f.fieldType
		// if fieldType == compiledIOPType {
		// 	logrus.Infof("Ignoring a struct field name:%s of type *wizard.CompiledIOP in ref. mode\n", f.name)
		// 	continue
		// }

		fieldValue := v.FieldByName(f.name)

		// Special handling for *wizard.CompiledIOP fields
		if fieldType == compiledIOPType {
			if fieldValue.IsNil() {
				raw[f.rawName] = []byte(NilString)
			} else {
				logrus.Infof("Ser. a field name:%s of type *wizard.CompiledIOP embedded within a struct \n", f.name)
				iopSer, err := SerializeCompiledIOP(fieldValue.Interface().(*wizard.CompiledIOP))
				if err != nil {
					return nil, fmt.Errorf("failed to serialize *wizard.CompiledIOP field %s: %w", f.name, err)
				}
				raw[f.rawName] = iopSer
			}
			continue
		}

		// Serialize other fields as usual
		r, err := SerializeValue(fieldValue, newMode)
		if err != nil {
			return nil, fmt.Errorf("could not serialize struct field %v.%v: %w", typeOfV.String(), f.rawName, err)
		}
		raw[f.rawName] = r
	}
	return serializeAnyWithCborPkg(raw)
}

// serializeMap serializes a map to CBOR-encoded data.
func serializeMap(v reflect.Value, mode Mode) (json.RawMessage, error) {
	keys := v.MapKeys()
	keyStrings := make([]string, 0, len(keys))
	keyMap := make(map[string]reflect.Value, len(keys))

	for _, k := range keys {
		keyString, err := castAsString(k)
		if err != nil {
			return nil, fmt.Errorf("invalid map key type %q: %w", v.Type().Key().Name(), err)
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

// deserializeMap deserializes a map from CBOR-encoded data, supporting non-string keys.
func deserializeMap(data json.RawMessage, mode Mode, t reflect.Type, comp *wizard.CompiledIOP) (reflect.Value, error) {
	var raw map[string]json.RawMessage
	if err := deserializeAnyWithCborPkg(data, &raw); err != nil {
		return reflect.Value{}, fmt.Errorf("failed to deserialize to %q: %w", t.Name(), err)
	}

	keyType := t.Key()
	v := reflect.MakeMap(t)
	valueType := t.Elem()

	for keyRaw, valRaw := range raw {
		// Convert the string key to the target key type
		key, err := convertKeyToType(keyRaw, keyType)
		if err != nil {
			return reflect.Value{}, fmt.Errorf("failed to convert map key %q to type %v: %w", keyRaw, keyType.Name(), err)
		}

		// Deserialize the value
		val, err := DeserializeValue(valRaw, mode, valueType, comp)
		if err != nil {
			return reflect.Value{}, fmt.Errorf("failed to deserialize map field %v of type %v: %w", keyRaw, valueType.Name(), err)
		}

		v.SetMapIndex(key, val)
	}
	return v, nil
}

func deserializeInterface(data json.RawMessage, mode Mode, t reflect.Type, comp *wizard.CompiledIOP) (reflect.Value, error) {

	if IsIgnoreableType(t) {
		if bytes.Equal(data, []byte(NilString)) {
			logrus.Warnf("Skipping deserialization of Ignorable type: %v", t)
			return reflect.Zero(t), nil
		}
		return reflect.Value{}, fmt.Errorf("expected null or empty struct for %v, got: %v", t, data)
	}

	// Special case for pureExprMode and metadataType
	if mode == pureExprMode && t == metadataType {
		return deserializeStringVar(data, mode, comp)
	}

	// Create a new reflect.Value for the interface type
	// Reminder; here the important thing is to ensure that the returned
	// Value actually bears the requested interface type and not the
	// concrete type.
	ifaceValue := reflect.New(t).Elem()
	if bytes.Equal(data, []byte(NilString)) {
		return ifaceValue, nil
	}

	// Deserialize the raw JSON into a structured format
	// Reminder: we serialize interfaces as
	// {
	// 		"type": "<pkg/path>.(<TypeName>)"
	// 		"value": <rawJson of the concrete type>
	// }
	var raw struct {
		Type  string          `json:"type"`
		Value json.RawMessage `json:"value"`
	}
	if err := deserializeAnyWithCborPkg(data, &raw); err != nil {
		return reflect.Value{}, fmt.Errorf("failed to deserialize interface %q: %w", t.Name(), err)
	}

	//fmt.Printf("Deser Interface:%s \n", raw.Type)
	// Handle specific types based on the "Type" field
	switch raw.Type {
	case "column.Natural":
		return deserializeColumnNatural(raw.Value, mode, comp, ifaceValue)
	case "column.Shifted":
		return deserializeStruct(raw.Value, mode, shiftedColType, comp)
	case "*column.FakeColumn":
		return deserializeFakeColumn(raw.Value, mode, comp, ifaceValue)
	case "verifiercol.ConstCol":
		return deserializeStruct(raw.Value, mode, vConstColType, comp)
	case "verifiercol.ExpandedVerifCol":
		return deserializeStruct(raw.Value, mode, vExpandedColType, comp)
	case "verifiercol.FromAccessors":
		return deserializeStruct(raw.Value, mode, vFromAccessorType, comp)
	case "verifiercol.RepeatedAccessor":
		return deserializeStruct(raw.Value, mode, vRepeatedAccessorType, comp)
	case "*dedicated.ManuallyShifted":
		return deserializeManuallyShifted(raw.Value, mode, comp, ifaceValue)
	default:
		return deserializeConcreteType(raw.Type, raw.Value, mode, comp, ifaceValue)
	}
}

// Helper function to deserialize concrete types
func deserializeConcreteType(typeName string, value json.RawMessage, mode Mode, comp *wizard.CompiledIOP, ifaceValue reflect.Value) (reflect.Value, error) {
	// Find the registered implementation for the type
	concreteType, err := findRegisteredImplementation(typeName)
	if err != nil {
		return reflect.Value{}, fmt.Errorf("failed to find concrete type for %q: %w", typeName, err)
	}

	// Determine if the concrete type is a pointer
	isPointer := concreteType.Kind() == reflect.Ptr
	targetType := concreteType
	if isPointer {
		// Dereference the elem if it is a pointer
		targetType = concreteType.Elem()
	}

	// Deserialize the value into the target type
	v, err := DeserializeValue(value, mode, targetType, comp)
	if err != nil {
		return reflect.Value{}, fmt.Errorf("failed to deserialize concrete type %q: %w", typeName, err)
	}

	// Handle pointer types: When the concrete type is a pointer, the interface must hold a pointer to the deserialized value
	// to match the registered type and ensure addressability. Without this, for nested interfaces or pointer fields within
	// the concrete type (e.g., a struct with a *[]int field), deserializePointer (which calls v.Addr()) would panic if v is
	// unaddressable (e.g., a temporary value). This is critical when deserializing interface values like ifaces.ColAssignment,
	// where the concrete type or its fields may be pointers. For non-pointer types, the deserialized value can be assigned
	// directly, as it is already in the correct form and typically addressable when stored in the interface.
	if isPointer {
		ptrValue := reflect.New(targetType) // returns a pointer value to new type
		if v.IsValid() && v.Type().AssignableTo(targetType) {
			ptrValue.Elem().Set(v)
		} else {
			return reflect.Value{}, fmt.Errorf("cannot assign deserialized value of type %v to pointer type %v", v.Type(), targetType)
		}
		ifaceValue.Set(ptrValue)
	} else {
		ifaceValue.Set(v)
	}

	return ifaceValue, nil
}

func deserializePointer(data json.RawMessage, mode Mode, t reflect.Type, comp *wizard.CompiledIOP) (reflect.Value, error) {
	v, err := DeserializeValue(data, mode, t.Elem(), comp)
	if err != nil {
		return reflect.Value{}, fmt.Errorf("failed to deserialize pointer-type %q: %w", t.Name(), err)
	}
	return v.Addr(), nil
}

// deserializeStruct deserializes JSON into a struct, with special handling for *wizard.CompiledIOP fields.
func deserializeStruct(data json.RawMessage, mode Mode, t reflect.Type, comp *wizard.CompiledIOP) (reflect.Value, error) {
	// Handle ignorable types
	if IsIgnoreableType(t) {
		if bytes.Equal(data, []byte(NilString)) {
			logrus.Warnf("Skipping deserialization of Ignorable type: %v", t)
			return reflect.Zero(t), nil
		}
		return reflect.Value{}, fmt.Errorf("expected null or empty struct for %v, got: %v", t, data)
	}

	// Handle specific type/mode combinations (e.g., naturalType, queryType)
	// When the type is a natural column in reference mode, the serializer
	// optimizes by just passing the column name. To deserialize, we need
	// to parse the name and look up the column into `comp`. This works
	// under the assumption that the column declaration are always
	// deserialized before we deserialize the references to these columns.
	if mode == ReferenceMode && t == naturalType {
		var colID ifaces.ColID
		if err := deserializeAnyWithCborPkg(data, &colID); err != nil {
			return reflect.Value{}, fmt.Errorf("could not deserialize column ID: %w", err)
		}
		return reflect.ValueOf(comp.Columns.GetHandle(colID)), nil
	}
	if mode == DeclarationMode && t == naturalType {
		rawType := reflect.TypeOf(&serializableColumnDecl{})
		v, err := DeserializeValue(data, mode, rawType, comp)
		if err != nil {
			return reflect.Value{}, fmt.Errorf("could not deserialize column interface declaration: %w", err)
		}
		return reflect.ValueOf(v.Interface().(*serializableColumnDecl).intoNaturalAndRegister(comp)), nil
	}
	if mode == DeclarationMode && t == manuallyShiftedType {
		rawType := reflect.TypeOf(&serializableManuallyShifted{})
		v, err := DeserializeValue(data, mode, rawType, comp)
		if err != nil {
			return reflect.Value{}, fmt.Errorf("could not deserialize column interface declaration: %w", err)
		}
		return reflect.ValueOf(v.Interface().(*serializableManuallyShifted).intoManuallyShifted(comp)), nil
	}

	// Likewise for if the type implements query, we can just deserialize
	// looking up the query by name in the compiled IOP. So there is no need
	// to include it in the serialized object. The ID is enough.
	if mode == ReferenceMode && t.Implements(queryType) {
		var queryID ifaces.QueryID
		if err := deserializeAnyWithCborPkg(data, &queryID); err != nil {
			return reflect.Value{}, fmt.Errorf("could not deserialize query ID: %w", err)
		}
		// Here, we would need to carefully inspect the type of the query to
		// know where in where to look it for in compiled IOP. Instead, we
		// go the easy way and just try in both QueriesParams and
		// QueriesNoParams.
		if comp.QueriesParams.Exists(queryID) {
			return reflect.ValueOf(comp.QueriesParams.Data(queryID)), nil
		}
		if comp.QueriesNoParams.Exists(queryID) {
			return reflect.ValueOf(comp.QueriesNoParams.Data(queryID)), nil
		}

		// This means that the requested column ID is nowhere to be found
		// in the compiled IOP and this is likely a bug in the serializer.
		// If it happens, it might mean that the serializer is attempting
		// to read the current reference before it read its declaration.
		// This is the caller's responsibility to ensure that this is not
		// the case.
		utils.Panic("Could not find requested column ID: %v", queryID)
	}

	// If 'v' implements query or column, it may be the case that its
	// components refer to other queries or columns. In that case, we should
	// ensure that the deserializer enters in reference mode in the deeper
	// steps of the recursion.
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

		fieldType := f.fieldType
		// if fieldType == compiledIOPType {
		// 	logrus.Infof("Ignoring Deser. a struct field name:%s of type *wizard.CompiledIOP  in ref. mode\n", f.name)
		// 	// newMode = ReferenceMode
		// 	continue
		// }

		fieldRaw, ok := raw[f.rawName]
		if !ok {
			utils.Panic("Missing struct field %q.%v of type %q, provided sub-JSON: %v", t.String(), f.name, f.fieldType.String(), raw)
		}

		// Special handling for *wizard.CompiledIOP fields => ex: RecursionComp
		if fieldType == compiledIOPType {
			if bytes.Equal(fieldRaw, []byte(NilString)) {
				v.FieldByName(f.name).Set(reflect.Zero(fieldType))
			} else {
				logrus.Infof("Deser. field name:%s of type *wizard.CompiledIOP embedded within a struct\n", f.name)
				iop, err := DeserializeCompiledIOP(fieldRaw)
				if err != nil {
					return reflect.Value{}, fmt.Errorf("failed to deserialize *wizard.CompiledIOP field %s: %w", f.name, err)
				}
				v.FieldByName(f.name).Set(reflect.ValueOf(iop))
			}
			continue
		}

		// Deserialize other fields as usual
		fieldValue, err := DeserializeValue(fieldRaw, newMode, fieldType, comp)
		if err != nil {
			utils.Panic("Could not deserialize struct field %q.%v of type %q: %v", t.String(), f.name, fieldType.String(), err)
		}

		// Handle zero (invalid) values
		if !fieldValue.IsValid() {
			v.FieldByName(f.name).Set(reflect.Zero(fieldType))
			continue
		}

		// Handle pointer types
		if fieldValue.Type().Kind() == reflect.Ptr && fieldValue.Type().Elem() == fieldType {
			fieldValue = fieldValue.Elem()
		}

		v.FieldByName(f.name).Set(fieldValue)
	}

	return v, nil
}
