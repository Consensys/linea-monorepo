package serialization

import (
	"encoding/json"
	"fmt"
	"reflect"
	"unicode"

	"github.com/consensys/linea-monorepo/prover/protocol/coin"
	"github.com/consensys/linea-monorepo/prover/protocol/column"
	"github.com/consensys/linea-monorepo/prover/protocol/dedicated"
	"github.com/consensys/linea-monorepo/prover/protocol/ifaces"
	"github.com/consensys/linea-monorepo/prover/protocol/wizard"
	"github.com/consensys/linea-monorepo/prover/utils"
	"github.com/fxamacker/cbor/v2"
	"github.com/sirupsen/logrus"
)

type objectType int8

const (
	typeObjectType objectType = iota
	columnObjectType
	columnName
	coinObjectType
	queriesObjectType
)

// Serializer is a serializer for wizard-IOPs
type Serializer struct {
	typeMap    map[string]int
	Types      []string `cbor:"types,omitempty"`
	columnMap  map[ifaces.ColID]int
	Columns    [][]byte `cbor:"columns,omitempty"`
	coinMap    map[coin.Name]int
	Coin       [][]byte `cbor:"coin,omitempty"`
	queryMap   map[ifaces.QueryID]int
	Queries    []byte `cbor:"queries,omitempty"`
	MainObject any    `cbor:"main_object,omitempty"`
}

type BackReference struct {
	what objectType `cbor:"w"`
	id   int        `cbor:"i"`
}

type SerializableInterface struct {
	T string `cbor:"t"`
	V []byte `cbor:"v"`
}

func (ser *Serializer) MarshalValue(v reflect.Value) ([]byte, error) {

	if !v.IsValid() || v.Interface() == nil {
		return json.RawMessage(NilString), nil
	}

	typeOfV := v.Type()

	if IsIgnoreableType(typeOfV) {
		return nil, nil
	}

	if typeOfV == naturalType {
		col := v.Interface().(column.Natural)
		return ser.MarshalColumn(col)
	}

	if typeOfV == coinType {
		coin := v.Interface().(coin.Info)
		return ser.MarshalCoin(coin)
	}

	switch v.Kind() {
	case reflect.Bool, reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32,
		reflect.Int64, reflect.Uint, reflect.Uint8, reflect.Uint16,
		reflect.Uint32, reflect.Uint64, reflect.String:
		return serializePrimitive(v)
	case reflect.Array, reflect.Slice:
		return serializeArrayOrSlice(v)
	case reflect.Interface:
		return serializeInterface(v)
	case reflect.Map:
		return serializeMap(v)
	case reflect.Pointer:
		return SerializeValue(v.Elem())
	case reflect.Struct:
		return serializeStruct(v)
	case reflect.Func:
		logrus.Debugf("Ignoring func type at value: %v", v)
		return json.RawMessage(NilString), nil
	default:
		return nil, fmt.Errorf("unsupported type kind: %v", v.Kind())
	}
}

func (ser *Serializer) MarshalInterface(v reflect.Value) ([]byte, error) {

	var (
		concrete          = v.Elem()
		concreteTypeStr   = getPkgPathAndTypeNameIndirect(concrete.Interface())
		concreteType, err = findRegisteredImplementation(concreteTypeStr)
	)

	if err != nil {
		return nil, fmt.Errorf("could not find registered type for %q: %w", concreteTypeStr, err)
	}

	if IsIgnoreableType(concreteType) {
		return nil, nil
	}

	marshaledV, err := ser.MarshalValue(concrete)
	if err != nil {
		return nil, fmt.Errorf("could not marshal value, interface-type=%v concrete-type=%v: %w", v.Type().String(), concreteTypeStr, err)
	}

	posType, ok := ser.typeMap[concreteTypeStr]

	if !ok {
		posType = len(ser.Types)
		ser.typeMap[concreteTypeStr] = posType
		ser.Types = append(ser.Types, concreteTypeStr)
	}

	unmarshaled := SerializableInterface{
		T: ser.Types[posType],
		V: marshaledV,
	}

	marshaled, err := cbor.Marshal(unmarshaled)
	if err != nil {
		return nil, fmt.Errorf("could not marshal interface value, interface-type=%v concrete-type=%v: %w", v.Type().String(), concreteTypeStr, err)
	}

	return marshaled, nil
}

func (ser *Serializer) MarshalColumn(col column.Natural) ([]byte, error) {

	if _, ok := ser.columnMap[col.ID]; !ok {

		newPos := len(ser.Columns)
		ser.columnMap[col.ID] = newPos

		serColumn := serializableColumn{
			Name:     string(col.GetColID()),
			Round:    int8(col.Round()),
			Status:   int8(col.Status()),
			Log2Size: int8(utils.Log2Floor(col.Size())),
		}

		marshaled, err := cbor.Marshal(serColumn)
		if err != nil {
			return nil, fmt.Errorf("could not marshal column, n=%q err=%v", col.ID, err)
		}

		ser.Columns = append(ser.Columns, marshaled)
	}

	pos := ser.columnMap[col.ID]

	backRef := &BackReference{
		what: columnObjectType,
		id:   pos,
	}

	res, err := cbor.Marshal(backRef)
	if err != nil {
		return nil, fmt.Errorf("could not marshal backreference, n=%q err=%v", col.ID, err)
	}

	return res, nil
}

// serializableColumnDecl is used to represent a "natural" column, meaning a
// column that is explicitly registered as part of the scheme. This is in
// opposition with [serializableColumnRef] where such columns are encoded
// by just citing their names.
//
// Concretely, we need this because [column.Natural] has a complex structure
// that is deeply nested within [column.Store]. And this prevents directly
// applying the generic reflection-based serialization logic to it.
type serializableColumn struct {
	Name     string `cbor:"n"`
	Round    int8   `cbor:"r"`
	Status   int8   `cbor:"s"`
	Log2Size int8   `cbor:"l"`
}

// serializableCoin
type serializableCoin struct {
	Name  string `cbor:"n"`
	Type  int8   `cbor:"t"`
	Size  int16  `cbor:"s,omitempty"`
	Upper int16  `cbor:"u,omitempty"`
	Round int8   `cbor:"r"`
}

// serializeStruct serializes a struct into JSON, ignoring certain fields based on the mode and type.
// Includes special handling for *wizard.CompiledIOP fields.
func (ser *Serializer) serializeStruct(v reflect.Value, mode Mode) (json.RawMessage, error) {

	typeOfV := v.Type()

	// Skip serialization for ignorable types
	if IsIgnoreableType(typeOfV) {
		return nil, nil
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
