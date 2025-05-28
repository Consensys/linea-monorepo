package serialization

import (
	"fmt"
	"reflect"
	"sort"

	"github.com/consensys/linea-monorepo/prover/protocol/coin"
	"github.com/consensys/linea-monorepo/prover/protocol/column"
	"github.com/consensys/linea-monorepo/prover/protocol/ifaces"
	"github.com/consensys/linea-monorepo/prover/protocol/query"
	"github.com/consensys/linea-monorepo/prover/protocol/wizard"
	"github.com/consensys/linea-monorepo/prover/utils"
	"github.com/google/uuid"
)

type BackReferenceType int8

const (
	ColumnBackReference BackReferenceType = iota
	ColumnIDBackReference
	CoinBackReference
	CoinIDBackReference
	QueryBackReference
	QueryIDBackReference
	CompiledIOPBackReference
	StoreBackReference
)

var (
	TypeOfColumnNatural = reflect.TypeOf(column.Natural{})
	TypeOfColumnID      = reflect.TypeOf(ifaces.ColID(""))
	TypeOfCoin          = reflect.TypeOf(coin.Info{})
	TypeOfCoinID        = reflect.TypeOf(coin.Name(""))
	TypeOfQuery         = reflect.TypeOf((*ifaces.Query)(nil)).Elem()
	TypeOfQueryID       = reflect.TypeOf(ifaces.QueryID(""))
	TypeOfCompiledIOP   = reflect.TypeOf((*wizard.CompiledIOP)(nil))
	TypeOfStore         = reflect.TypeOf((*column.Store)(nil))
)

type BackReference int

type Serializer struct {
	PackedObject    *PackedObject
	typeMap         map[string]int
	structSchemaMap map[string]int
	coinMap         map[uuid.UUID]int
	coinIdMap       map[string]int
	columnMap       map[uuid.UUID]int
	columnIdMap     map[string]int
	queryMap        map[uuid.UUID]int
	queryIDMap      map[string]int
	compiledIOPs    map[*wizard.CompiledIOP]int
	Stores          map[*column.Store]int
}

type Deserializer struct {
	TypeMap               map[string]int
	StructSchemaMap       map[string]int
	Columns               []*column.Natural
	Coins                 []*coin.Info
	Queries               []*ifaces.Query
	CompiledIOP           []*wizard.CompiledIOP
	Store                 []*column.Store
	PreUnmarshalledObject *PackedObject
}

type PackedObject struct {
	Types        []string             `cbor:"t"`
	StructSchema []PackedStructSchema `cbor:"s"`
	ColumnIDs    []string             `cbor:"id"`
	Columns      []PackedStructObject `cbor:"c"`
	CoinIDs      []string             `cbor:"i"`
	Coins        []PackedCoin         `cbor:"o"`
	QueryIDs     []string             `cbor:"w"`
	Store        [][]any              `cbor:"st"`
	Queries      []any                `cbor:"q"`
	CompiledIOP  []any                `cbor:"a"`
	Payload      []byte               `cbor:"p"`
}

type PackedIFace struct {
	Type     int `cbor:"t"`
	Concrete any `cbor:"c"`
}

type PackedCoin struct {
	Type        int8   `cbor:"t"`
	Size        int    `cbor:"s,omitempty"`
	UpperBound  int32  `cbor:"u,omitempty"`
	Name        string `cbor:"n"`
	Round       int    `cbor:"r"`
	CompiledIOP int8   `cbor:"i"`
}

type PackedStructSchema struct {
	Type   string   `cbor:"t"`
	Fields []string `cbor:"f"`
}

type PackedStructObject []any

func (s *Serializer) PackValue(v reflect.Value) any {

	// This captures the case where the value is nil to begin with
	if !v.IsValid() || v.Interface() == nil {
		return nil
	}

	typeOfV := v.Type()

	switch {
	case typeOfV == TypeOfColumnNatural:
		return s.PackColumn(v.Interface().(column.Natural))
	case typeOfV == TypeOfColumnID:
		return s.PackColumnID(v.Interface().(ifaces.ColID))
	case typeOfV == TypeOfCoin:
		return s.PackCoin(v.Interface().(coin.Info))
	case typeOfV == TypeOfCoinID:
		return s.PackCoinID(v.Interface().(coin.Name))
	case typeOfV == TypeOfQuery:
		return s.PackQuery(v.Interface().(ifaces.Query))
	case typeOfV == TypeOfQueryID:
		return s.PackQueryID(v.Interface().(ifaces.QueryID))
	case typeOfV == TypeOfCompiledIOP:
		return s.PackCompiledIOP(v.Interface().(*wizard.CompiledIOP))
	case typeOfV == TypeOfStore:
		return s.PackStore(v.Interface().(*column.Store))
	}

	switch v.Kind() {
	case reflect.Bool, reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32,
		reflect.Int64, reflect.Uint, reflect.Uint8, reflect.Uint16,
		reflect.Uint32, reflect.Uint64, reflect.String:
		return v.Interface()
	case reflect.Array, reflect.Slice:
		return s.PackArrayOrSlice(v)
	case reflect.Interface:
		return s.PackInterface(v)
	case reflect.Map:
		return s.PackMap(v)
	case reflect.Pointer:
		return s.PackValue(v.Elem())
	case reflect.Struct:
		return s.PackStructObject(v)
	default:
		panic(fmt.Sprintf("unsupported type kind: %v", v.Kind()))
	}
}

func (de *Deserializer) UnpackValue(v any, t reflect.Type) (reflect.Value, error) {

	if v == nil {
		return reflect.Zero(t), nil
	}

	switch t.Kind() {
	case reflect.Bool, reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32,
		reflect.Int64, reflect.Uint, reflect.Uint8, reflect.Uint16,
		reflect.Uint32, reflect.Uint64, reflect.String:
		v := reflect.New(t).Elem()
		v.Set(reflect.ValueOf(v.Interface()))
		return v, nil
	case reflect.Array, reflect.Slice:
		return de.UnpackArrayOrSlice(v.([]any), t)
	case reflect.Map:
		return de.UnpackMap(v.(map[string]any), t)
	case reflect.Interface:
		panic("unimplemented")
	case reflect.Pointer:
		return de.UnpackPointer(v, t)
	case reflect.Struct:
		if v_, ok := v.(PackedStructObject); ok {
			v = []any(v_)
		}
		return de.UnpackStructObject(v.([]any), t)
	default:
		return reflect.Value{}, fmt.Errorf("unsupported type kind: %v", t.Kind())
	}

}

func (s *Serializer) PackColumn(c column.Natural) BackReference {

	cid := c.UUID()

	if _, ok := s.columnMap[cid]; !ok {
		packed := c.Pack()
		marshaled := s.PackStructObject(reflect.ValueOf(packed))
		s.PackedObject.Columns = append(s.PackedObject.Columns, marshaled)
		s.columnMap[cid] = len(s.PackedObject.Columns) - 1
	}

	return BackReference(
		s.columnMap[cid],
	)
}

func (s *Serializer) PackColumnID(c ifaces.ColID) BackReference {

	if _, ok := s.columnIdMap[string(c)]; !ok {
		s.PackedObject.ColumnIDs = append(s.PackedObject.ColumnIDs, string(c))
		s.columnIdMap[string(c)] = len(s.PackedObject.ColumnIDs) - 1
	}

	return BackReference(s.columnIdMap[string(c)])

}

func (de *Deserializer) UnpackColumnID(v BackReference) (reflect.Value, error) {

	if v < 0 || int(v) >= len(de.PreUnmarshalledObject.ColumnIDs) {
		return reflect.Value{}, fmt.Errorf("invalid column ID: %v", v)
	}

	res := ifaces.ColID(de.PreUnmarshalledObject.ColumnIDs[v])
	return reflect.ValueOf(res), nil
}

func (s *Serializer) PackCoin(c coin.Info) BackReference {

	if _, ok := s.coinMap[c.UUID()]; !ok {
		s.PackedObject.Coins = append(s.PackedObject.Coins, s.AsPackedCoin(c))
		s.coinMap[c.UUID()] = len(s.PackedObject.Coins) - 1
	}

	return BackReference(s.coinMap[c.UUID()])
}

func (s *Serializer) PackCoinID(c coin.Name) BackReference {

	if _, ok := s.coinIdMap[string(c)]; !ok {
		s.PackedObject.CoinIDs = append(s.PackedObject.CoinIDs, string(c))
		s.coinIdMap[string(c)] = len(s.PackedObject.CoinIDs) - 1
	}

	return BackReference(s.coinIdMap[string(c)])
}

func (s *Deserializer) UnpackedCoinID(v BackReference) (reflect.Value, error) {

	if v < 0 || int(v) >= len(s.PreUnmarshalledObject.CoinIDs) {
		return reflect.Value{}, fmt.Errorf("invalid coin ID: %v", v)
	}

	return reflect.ValueOf(s.PreUnmarshalledObject.CoinIDs[v]), nil
}

func (s *Serializer) PackQuery(q ifaces.Query) BackReference {

	if _, ok := s.queryMap[q.UUID()]; !ok {

		var obj any

		switch q := q.(type) {
		case query.UnivariateEval,
			query.FixedPermutation,
			query.MiMC,
			query.Range,
			query.GlobalConstraint,
			query.GrandProduct,
			query.LogDerivativeSum,
			query.InnerProduct,
			query.Inclusion,
			query.Permutation,
			query.LocalConstraint,
			query.LocalOpening:
			obj = s.PackStructObject(reflect.ValueOf(q))
		case *query.PlonkInWizard:
			obj = s.PackStructObject(reflect.ValueOf(*q))
		case *query.Horner:
			obj = s.PackStructObject(reflect.ValueOf(*q))
		}

		s.PackedObject.Queries = append(s.PackedObject.Queries, obj)
		s.queryMap[q.UUID()] = len(s.PackedObject.Queries) - 1
	}

	return BackReference(s.queryMap[q.UUID()])
}

func (s *Serializer) PackQueryID(q ifaces.QueryID) BackReference {

	if _, ok := s.queryIDMap[string(q)]; !ok {
		s.PackedObject.QueryIDs = append(s.PackedObject.QueryIDs, string(q))
		s.queryIDMap[string(q)] = len(s.PackedObject.QueryIDs) - 1
	}

	return BackReference(s.queryIDMap[string(q)])

}

func (s *Serializer) PackCompiledIOP(comp *wizard.CompiledIOP) any {

	if _, ok := s.compiledIOPs[comp]; !ok {
		obj := s.PackStructObject(reflect.ValueOf(*comp))
		s.PackedObject.CompiledIOP = append(s.PackedObject.CompiledIOP, obj)
		s.compiledIOPs[comp] = len(s.PackedObject.CompiledIOP) - 1
	}

	return BackReference(s.compiledIOPs[comp])

}

func (ser *Serializer) PackStore(s *column.Store) any {

	if _, ok := ser.Stores[s]; !ok {
		obj := ser.PackStructObject(reflect.ValueOf(*s))
		ser.PackedObject.Store = append(ser.PackedObject.Store, obj)
		ser.Stores[s] = len(ser.PackedObject.Store) - 1
	}

	return BackReference(ser.Stores[s])

}

// PackArrayOrSlice serializes arrays or slices by recursively serializing each element.
func (s *Serializer) PackArrayOrSlice(v reflect.Value) []any {
	res := make([]any, v.Len())
	for i := 0; i < v.Len(); i++ {
		res[i] = s.PackValue(v.Index(i))
	}
	return res
}

func (de *Deserializer) UnpackArrayOrSlice(v []any, t reflect.Type) (reflect.Value, error) {

	var res reflect.Value

	switch t.Kind() {
	case reflect.Array:
		res = reflect.New(t).Elem()
		if t.Len() != len(v) {
			return reflect.Value{}, fmt.Errorf("failed to deserialize to %q, size mismatch: %d != %d", t.Name(), len(v), t.Len())
		}
	case reflect.Slice:
		res = reflect.MakeSlice(t, len(v), len(v))
	default:
		return reflect.Value{}, fmt.Errorf("failed to deserialize to %q, expected array or slice", t.Name())
	}

	subType := t.Elem()
	for i := 0; i < len(v); i++ {
		subV, err := de.UnpackValue(v[i], subType)
		if err != nil {
			return reflect.Value{}, fmt.Errorf("failed to deserialize to %q, error in element %d: %w", t.Name(), i, err)
		}
		res.Index(i).Set(subV)
	}

	return res, nil
}

func (s *Serializer) AsPackedCoin(c coin.Info) PackedCoin {
	return PackedCoin{
		Type:       int8(c.Type),
		Size:       c.Size,
		UpperBound: int32(c.UpperBound),
		Name:       string(c.Name),
		Round:      c.Round,
	}
}

func (s *Serializer) PackStructSchema(t reflect.Type) (schema PackedStructSchema) {

	if t.Kind() != reflect.Struct {
		utils.Panic("s.Kind() != reflect.Struct, type=%v", t.String())
	}

	cleanTypeString := getPkgPathAndTypeName(s)

	if i, ok := s.structSchemaMap[cleanTypeString]; ok {
		return s.PackedObject.StructSchema[i]
	}

	schema = PackedStructSchema{
		Type:   cleanTypeString,
		Fields: make([]string, t.NumField()),
	}

	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		schema.Fields[i] = field.Name
	}

	positionOfType := len(s.PackedObject.StructSchema)
	s.structSchemaMap[cleanTypeString] = positionOfType
	s.PackedObject.StructSchema = append(s.PackedObject.StructSchema, schema)

	return schema
}

func (s *Serializer) PackInterface(v reflect.Value) any {

	var (
		concrete          = v.Elem()
		cleanConcreteType = getPkgPathAndTypeName(concrete.Interface())
	)

	if _, ok := s.typeMap[cleanConcreteType]; !ok {
		s.PackedObject.Types = append(s.PackedObject.Types, cleanConcreteType)
		s.typeMap[cleanConcreteType] = len(s.PackedObject.Types) - 1
	}

	return PackedIFace{
		Type:     s.typeMap[cleanConcreteType],
		Concrete: s.PackValue(concrete),
	}
}

func (s *Serializer) PackStructObject(obj reflect.Value) PackedStructObject {

	if obj.Kind() != reflect.Struct {
		utils.Panic("obj.Kind() != reflect.Struct, type=%v", obj.Type().String())
	}

	values := make([]any, obj.NumField())

	// Note, since we don't want to register the schema before going through
	// all the components, we have to rely on the fact that schema and this loop
	// declare the fields in the same order.
	for i := 0; i < obj.NumField(); i++ {
		field := obj.Field(i)
		values[i] = s.PackValue(field)
	}

	s.PackStructSchema(obj.Type())

	// Importantly, we want to be sure that all the component have been
	// converted before we convert the current type. That way, we can ensure
	// that all the information of the dependency is added prior to the
	// information on the dependant struct. This is necessary for seria-
	// lization.
	return values
}

func (de *Deserializer) UnpackStructObject(v PackedStructObject, t reflect.Type) (reflect.Value, error) {

	if t.Kind() != reflect.Struct {
		return reflect.Value{}, fmt.Errorf("invalid type: %v", t.String())
	}

	var (
		cleanType    = getPkgPathAndTypeName(t)
		schemaID, ok = de.StructSchemaMap[cleanType]
	)

	if !ok {
		return reflect.Value{}, fmt.Errorf("invalid type: %v, it was not found in the schema", t.String())
	}

	var (
		res    = reflect.New(t).Elem()
		schema = de.PreUnmarshalledObject.StructSchema[schemaID]
	)

	for i := range v {

		_, ok := t.FieldByName(schema.Fields[i])
		if !ok {
			return reflect.Value{}, fmt.Errorf("invalid field name: %v, it was not found in the struct", schema.Fields[i], t.String())
		}

		field := res.FieldByName(schema.Fields[i])
		value, err := de.UnpackValue(v[i], field.Type())

		if err != nil {
			return reflect.Value{}, fmt.Errorf("could not deserialize field %q type=%v, err=%w", schema.Fields[i], field.Type().String(), err)
		}

		field.Set(value)
	}

	return res, nil
}

func (s *Serializer) PackMap(obj reflect.Value) map[string]any {

	if obj.Kind() != reflect.Map {
		utils.Panic("obj.Kind() != reflect.Map, type=%v", obj.Type().String())
	}

	var (
		keys       = obj.MapKeys()
		keyStrings = make([]string, 0, len(keys))
		keyMap     = make(map[string]reflect.Value, len(keys))
	)

	for _, k := range keys {
		keyString, err := castAsString(k)
		if err != nil {
			utils.Panic("invalid map key type %q: %v", obj.Type().Key().String(), err)
		}
		keyStrings = append(keyStrings, keyString)
		keyMap[keyString] = k
	}

	// Needed for deterministic encoding
	sort.Strings(keyStrings)

	res := make(map[string]any, len(keys))
	for _, keyString := range keyStrings {
		k := keyMap[keyString]
		res[keyString] = s.PackValue(obj.MapIndex(k))
	}

	return res
}

func (de *Deserializer) UnpackMap(v map[string]any, t reflect.Type) (reflect.Value, error) {

	if t.Kind() != reflect.Map {
		return reflect.Value{}, fmt.Errorf("invalid type: %v", t.String())
	}

	var (
		typeOfKey   = t.Key()
		typeOfValue = t.Elem()
		res         = reflect.MakeMap(t)
	)

	for key, val := range v {
		k, err := de.UnpackValue(key, typeOfKey)
		if err != nil {
			return reflect.Value{}, fmt.Errorf("could not deserialize map key %q type=%v, err=%w", key, typeOfKey.String(), err)
		}

		v, err := de.UnpackValue(val, typeOfValue)
		if err != nil {
			return reflect.Value{}, fmt.Errorf("could not deserialize map value %q type=%v, err=%w", val, typeOfValue.String(), err)
		}

		res.SetMapIndex(k, v)
	}

	return res, nil
}

func (de *Deserializer) UnpackPointer(v any, t reflect.Type) (reflect.Value, error) {

	if t.Kind() != reflect.Ptr {
		return reflect.Value{}, fmt.Errorf("invalid type: %v, expected a pointer", t.String())
	}

	value, err := de.UnpackValue(v, t.Elem())
	if err != nil {
		return reflect.Value{}, fmt.Errorf("could not deserialize pointer type=%v, err=%w", t.Elem().String(), err)
	}

	return value.Addr(), nil
}
