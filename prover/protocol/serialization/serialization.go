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
type mode int

const (
	nilString = "null"
)

const (
	DeclarationMode mode = iota
	ReferenceMode
	// pureExprMode is meant to serialize symbolic expression. All references to
	// a compiled IOP are stripped out from the serialized expression. It allows
	// deserializing without the complexity of the underlying CompiledIOP if
	// there is one. It is used for generating/reading test-case expressions.
	pureExprMode
)

// Types that necessitate special handling by the de/serializer
var (
	columnType   = reflect.TypeOf((*ifaces.Column)(nil)).Elem()
	queryType    = reflect.TypeOf((*ifaces.Query)(nil)).Elem()
	naturalType  = reflect.TypeOf(column.Natural{})
	metadataType = reflect.TypeOf((*symbolic.Metadata)(nil)).Elem()
)

// SerializeValue recursively serializes `v` into JSON. This function is
// specially crafted to handle the special types occurring in the `protocol`
// directory. It can be used in `DeclarationMode` where the [column.Natural]
// are serialized as [serializableColumnDecl] objects or in the [referenceMode]
// where the [ifaces.Query] and [column.Natural] objects are serialized using
// only their ID.
func SerializeValue(v reflect.Value, mode mode) (json.RawMessage, error) {

	// If v resolves to nil, then encode it as such
	if v.Interface() == nil {
		return json.RawMessage(nilString), nil
	}

	switch v.Kind() {
	// e.g. all the primitive types. Just defer to the normal JSON encoder even
	// if this is not optimal as this doubles the reflection overheads.
	case reflect.Bool, reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32,
		reflect.Int64, reflect.Uint, reflect.Uint8, reflect.Uint16,
		reflect.Uint32, reflect.Uint64, reflect.String:
		return serializeAnyWithCborPkg(v.Interface())

	case reflect.Array, reflect.Slice:

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

	case reflect.Interface:

		if mode == pureExprMode && v.Type() == metadataType {

			var (
				m              = v.Interface().(symbolic.Metadata)
				mString        = m.String()
				stringVar      = symbolic.StringVar(mString)
				stringVarValue = reflect.ValueOf(stringVar)
			)

			return SerializeValue(stringVarValue, mode)
		}

		if mode == DeclarationMode && v.Type() == columnType {
			// Only natural columns can be expected in this case.
			col := v.Interface().(column.Natural)
			raw := intoSerializableColDecl(&col)

			// Note that the mode does not really matter here.
			return SerializeValue(reflect.ValueOf(raw), mode)
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

	case reflect.Map:

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

	case reflect.Pointer:

		return SerializeValue(v.Elem(), mode)

	case reflect.Struct:
		var (
			typeOfV   = v.Type()
			numFields = typeOfV.NumField()
			raw       = map[string]json.RawMessage{}
		)

		// If we just want to encode a reference to a column that has been
		// explicitly registered in the store. It suffices to provide a
		// reference to the column. From that, we can load it in the store,
		// that's why there is no need to serialize the full object; the ID
		// is enough.
		if mode == ReferenceMode && typeOfV == naturalType {
			colID := v.Interface().(column.Natural).ID
			return serializeAnyWithCborPkg(colID)
		}

		// Note that this is the same handling as in the interface-level
		// handling. This can be triggered when passing the Column as initial
		// input to the serializer.
		if mode == DeclarationMode && typeOfV == naturalType {
			// Only natural columns can be expected in this case.
			col := v.Interface().(column.Natural)
			raw := intoSerializableColDecl(&col)

			// Note that the mode does not really matter here.
			return SerializeValue(reflect.ValueOf(raw), mode)
		}

		// Likewise for if the type implements query, we can just deserialize
		// looking up the query by name in the compiled IOP. So there is no need
		// to include it in the serialized object. The ID is enough.
		if mode == ReferenceMode && typeOfV.Implements(queryType) {
			queryID := v.Interface().(ifaces.Query).Name()
			return serializeAnyWithCborPkg(queryID)
		}

		// If 'v' implements query or column, it may be the case that its
		// components refer to other queries or columns. In that case, we should
		// ensure that the serializer enters in reference mode.
		if typeOfV.Implements(queryType) || typeOfV.Implements(columnType) {
			mode = ReferenceMode
		}

		for i := 0; i < numFields; i++ {
			var (
				fieldName    = typeOfV.Field(i).Name
				rawFieldName = strcase.ToLowerCamel(fieldName)
				fieldValue   = v.Field(i)
			)

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

	panic(fmt.Sprintf("unreachable: v=%++v", v))
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
func DeserializeValue(data json.RawMessage, mode mode, t reflect.Type, comp *wizard.CompiledIOP) (reflect.Value, error) {

	// If the string is <nil> then we know we are deserializing a nil object.
	if bytes.Equal(data, []byte(nilString)) {
		return reflect.New(t), nil
	}

	switch t.Kind() {

	default:
		return reflect.Value{}, fmt.Errorf("unsupported kind: %v", t.Kind().String())

	// e.g. all the primitive types. Just defer to the normal JSON encoder even
	// if this is not optimal as this doubles the reflection overheads.
	case reflect.Bool, reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32,
		reflect.Int64, reflect.Uint, reflect.Uint8, reflect.Uint16,
		reflect.Uint32, reflect.Uint64, reflect.String:

		v := reflect.New(t).Elem()
		return v, deserializeAnyWithCborPkg(data, v.Addr().Interface())

	case reflect.Array, reflect.Slice:

		var (
			raw     = []json.RawMessage{}
			v       reflect.Value
			subType = t.Elem()
		)

		if err := deserializeAnyWithCborPkg(data, &raw); err != nil {
			return reflect.Value{}, fmt.Errorf("failed to deserialize to a `%v` value: %w", t.Name(), err)
		}

		if t.Kind() == reflect.Array {
			v = reflect.New(t).Elem()
			if t.Len() != len(raw) {
				return reflect.Value{}, fmt.Errorf("failed to deserialize to `%v`, size mismatch: %d != %d", t.Name(), len(raw), v.Len())
			}
		}

		if t.Kind() == reflect.Slice {
			v = reflect.MakeSlice(t, len(raw), len(raw))
		}

		// sanity-check: the clause above must be exhaustive
		if v == (reflect.Value{}) {
			panic("v cannot be empty")
		}

		for i := range raw {
			subV, err := DeserializeValue(raw[i], mode, subType, comp)
			if err != nil {
				return reflect.Value{}, fmt.Errorf("failed to deserialize to `%v`: could not deserialize entry %v: %w", t.Name(), i, err)
			}
			v.Index(i).Set(subV)
		}

		return v, nil

	case reflect.Map:
		var (
			raw       = map[string]json.RawMessage{}
			keyType   = t.Key()
			valueType = t.Elem()
			v         = reflect.MakeMap(t)
		)

		if keyType.Kind() != reflect.String {
			// That's a JSON limitation
			return reflect.Value{}, fmt.Errorf("cannot deserialize a map whose keys are not string-like: `%v`", t.Name())
		}

		if err := deserializeAnyWithCborPkg(data, &raw); err != nil {
			return reflect.Value{}, fmt.Errorf("failed to deserialize to a `%v` value: %w", t.Name(), err)
		}

		for keyRaw, valRaw := range raw {
			val, err := DeserializeValue(valRaw, mode, valueType, comp)
			if err != nil {
				return reflect.Value{}, fmt.Errorf("failed to deserialize map field %v of type %v: %w", keyRaw, valueType.Name(), err)
			}

			// We are guaranteed at this point that the "kind" of `key`` is
			// String but the type could be different (for instance, it could be
			// [ifaces.ColID]). If that happens, then we need to convert `key`
			// into the
			key := reflect.ValueOf(keyRaw).Convert(keyType)
			v.SetMapIndex(key, val)
		}

		return v, nil

	case reflect.Interface:

		// Reminder; here the important thing is to ensure that the returned
		// Value actually bears the requested interface type and not the
		// concrete type.
		ifaceValue := reflect.New(t).Elem()

		if mode == pureExprMode && t == metadataType {
			var stringVar symbolic.StringVar
			return DeserializeValue(data, mode, reflect.TypeOf(stringVar), comp)
		}

		if mode == DeclarationMode && t == columnType {
			// Only natural columns can be expected in this case.
			rawType := reflect.TypeOf(&serializableColumnDecl{})

			// Note that the mode does not really matter here.
			v, err := DeserializeValue(data, mode, rawType, comp)
			if err != nil {
				return reflect.Value{}, fmt.Errorf("could not deserialize column interface declaration: %w", err)
			}

			// This mutates the comp object. This is required because this is
			// the only way to instantiate a Natural column cleanly.
			nat := v.Interface().(*serializableColumnDecl).intoNaturalAndRegister(comp)
			ifaceValue.Set(reflect.ValueOf(nat))
			return ifaceValue, nil
		}

		var (
			// Reminder: we serialize interfaces as
			// {
			// 		"type": "<pkg/path>.(<TypeName>)"
			// 		"value": <rawJson of the concrete type>
			// }
			raw = struct {
				Type  string          `json:"type"`
				Value json.RawMessage `json:"value"`
			}{}
		)

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

	case reflect.Pointer:

		v, err := DeserializeValue(data, mode, t.Elem(), comp)
		if err != nil {
			return reflect.Value{}, fmt.Errorf("failed to deserialize pointer-type `%v`: %w", t.Name(), err)
		}

		return v.Addr(), nil

	case reflect.Struct:

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

			// This panics if the column is not found. We are fine with because
			// this is a bug in the serializer if that happens.
			nat := comp.Columns.GetHandle(colID)
			return reflect.ValueOf(nat), nil
		}

		// Likewise for if the type implements query, we can just deserialize
		// looking up the query by name in the compiled IOP. So there is no need
		// to include it in the serialized object. The ID is enough.
		if mode == ReferenceMode && t.Implements(queryType) {
			var queryID ifaces.QueryID
			if err := deserializeAnyWithCborPkg(data, &queryID); err != nil {
				return reflect.Value{}, fmt.Errorf("could not deserialize column ID: %w", err)
			}

			// Here, we would need to carefully inspect the type of the query to
			// know where in where to look it for in compiled IOP. Instead, we
			// go the easy way and just try in both QueriesParams and
			// QueriesNoParams.
			if comp.QueriesParams.Exists(queryID) {
				q := comp.QueriesParams.Data(queryID)
				return reflect.ValueOf(q), nil
			}

			if comp.QueriesNoParams.Exists(queryID) {
				q := comp.QueriesNoParams.Data(queryID)
				return reflect.ValueOf(q), nil
			}

			// This means that the requested column ID is nowhere to be found
			// in the compiled IOP and this is likely a bug in the serializer.
			// If it happens, it might mean that the serializer is attempting
			// to read the current reference before it read its declaration.
			// This is the caller's responsibility to ensure that this is not
			// the case.
			utils.Panic("could not find requested column ID: %v", queryID)
		}

		if mode == DeclarationMode && t == naturalType {
			// Only natural columns can be expected in this case.
			rawType := reflect.TypeOf(&serializableColumnDecl{})

			// Note that the mode does not really matter here.
			v, err := DeserializeValue(data, mode, rawType, comp)
			if err != nil {
				return reflect.Value{}, fmt.Errorf("could not deserialize column interface declaration: %w", err)
			}

			// This mutates the comp object. This is required because this is
			// the only way to instantiate a Natural column cleanly.
			nat := v.Interface().(*serializableColumnDecl).intoNaturalAndRegister(comp)
			return reflect.ValueOf(nat), nil
		}

		// If 'v' implements query or column, it may be the case that its
		// components refer to other queries or columns. In that case, we should
		// ensure that the deserializer enters in reference mode in the deeper
		// steps of the recursion.
		if t.Implements(queryType) || t.Implements(columnType) {
			mode = ReferenceMode
		}

		var (
			numFields = t.NumField()
			raw       = map[string]json.RawMessage{}
			v         = reflect.New(t).Elem()
		)

		if err := deserializeAnyWithCborPkg(data, &raw); err != nil {
			return reflect.Value{}, fmt.Errorf("could note deserialize struct type `%v` : %w", t.Name(), err)
		}

		for i := 0; i < numFields; i++ {
			var (
				fieldName    = t.Field(i).Name
				rawFieldName = strcase.ToLowerCamel(fieldName)
				fieldType    = t.Field(i).Type
			)

			fieldRaw, ok := raw[rawFieldName]

			if !ok {
				// We don't have an `omitempty` feature in the current serializer,
				// and we don't think that it will happen. So we expect that
				// every field end up in the serialized object. It would be ok
				// to ignore unrecognized fields however.
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
}
