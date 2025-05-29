package serialization

import (
	"errors"
	"fmt"
	"reflect"
	"sort"

	"github.com/consensys/linea-monorepo/prover/protocol/coin"
	"github.com/consensys/linea-monorepo/prover/protocol/column"
	"github.com/consensys/linea-monorepo/prover/protocol/ifaces"
	"github.com/consensys/linea-monorepo/prover/protocol/wizard"
	"github.com/consensys/linea-monorepo/prover/utils"
	"github.com/google/uuid"
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
	TypeOfPackedColumn  = reflect.TypeOf(column.PackedNatural{})
	TypeOfPackedStore   = reflect.TypeOf(column.PackedStore{})
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
	Warnings        []string
}

type Deserializer struct {
	StructSchemaMap map[string]int
	Columns         []*column.Natural
	Coins           []*coin.Info
	Queries         []*ifaces.Query
	CompiledIOPs    []*wizard.CompiledIOP
	Stores          []*column.Store
	PackedObject    *PackedObject
	Warnings        []string
}

type PackedObject struct {
	Types        []string             `cbor:"t"`
	StructSchema []PackedStructSchema `cbor:"s"`
	ColumnIDs    []string             `cbor:"id"`
	Columns      []PackedStructObject `cbor:"c"`
	CoinIDs      []string             `cbor:"i"`
	Coins        []PackedCoin         `cbor:"o"`
	QueryIDs     []string             `cbor:"w"`
	Queries      []PackedStructObject `cbor:"q"`
	Store        [][]any              `cbor:"st"`
	CompiledIOP  []PackedStructObject `cbor:"a"`
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

// Serialize generically serializes a value.
func Serialize(v any) ([]byte, error) {

	ser := &Serializer{
		PackedObject:    &PackedObject{},
		typeMap:         map[string]int{},
		structSchemaMap: map[string]int{},
		coinMap:         map[uuid.UUID]int{},
		coinIdMap:       map[string]int{},
		columnMap:       map[uuid.UUID]int{},
		columnIdMap:     map[string]int{},
		queryMap:        map[uuid.UUID]int{},
		queryIDMap:      map[string]int{},
		compiledIOPs:    map[*wizard.CompiledIOP]int{},
		Stores:          map[*column.Store]int{},
	}

	payload, err := ser.PackValue(reflect.ValueOf(v))
	if err != nil {
		return nil, fmt.Errorf("could not pack the value: %w", err)
	}

	for i := range ser.Warnings {
		fmt.Println(ser.Warnings[i])
	}

	bytesOfPayload, err := encodeWithCBOR(payload)
	if err != nil {
		return nil, fmt.Errorf("could not encode the payload with CBOR: %v", err)
	}

	packedObject := ser.PackedObject
	packedObject.Payload = bytesOfPayload

	bytesOfV, err := encodeWithCBOR(packedObject)
	if err != nil {
		return nil, fmt.Errorf("could not encode the packedObject with CBOR: %v", err)
	}

	return bytesOfV, nil
}

func Deserialize(bytes []byte, v any) error {

	packedObject := &PackedObject{}

	if reflect.TypeOf(v).Kind() != reflect.Ptr {
		return fmt.Errorf("invalid type: %v, expected a pointer", reflect.TypeOf(v).String())
	}

	if err := decodeWithCBOR(bytes, packedObject); err != nil {
		return fmt.Errorf("the serialized object does not have the [PackedObject] format, err=%w", err)
	}

	deser := &Deserializer{
		StructSchemaMap: make(map[string]int, len(packedObject.StructSchema)),
		Columns:         make([]*column.Natural, len(packedObject.Columns)),
		Coins:           make([]*coin.Info, len(packedObject.Coins)),
		Queries:         make([]*ifaces.Query, len(packedObject.Queries)),
		CompiledIOPs:    make([]*wizard.CompiledIOP, len(packedObject.CompiledIOP)),
		Stores:          make([]*column.Store, len(packedObject.Store)),
		PackedObject:    packedObject,
	}

	var (
		payloadRoot any
		payloadType = reflect.TypeOf(v).Elem()
	)

	if err := decodeWithCBOR(packedObject.Payload, &payloadRoot); err != nil {
		return fmt.Errorf("could not deserialize the payload, err=%w", err)
	}

	// Before decoding, we need to register all the types
	for i, t := range packedObject.StructSchema {
		deser.StructSchemaMap[t.Type] = i
	}

	res, err := deser.UnpackValue(payloadRoot, payloadType)
	if err != nil {
		return fmt.Errorf("could not deserialize the payload, err=%w", err)
	}

	for i := range deser.Warnings {
		fmt.Println(deser.Warnings[i])
	}

	fmt.Printf("value of res = %v, value of v = %T\n", res.Type(), v)

	valueOfV := reflect.ValueOf(v)
	valueOfV.Elem().Set(res)
	return nil
}

func (s *Serializer) PackValue(v reflect.Value) (any, error) {

	// This captures the case where the value is nil to begin with
	if !v.IsValid() || v.Interface() == nil {
		return nil, nil
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
	case typeOfV.Implements(TypeOfQuery) && typeOfV.Kind() != reflect.Interface:
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
		return v.Interface(), nil
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

func (de *Deserializer) UnpackValue(v any, t reflect.Type) (r reflect.Value, e error) {

	if v == nil {
		return reflect.Zero(t), nil
	}

	switch {
	case t == TypeOfColumnNatural:
		return de.UnpackColumn(backReferenceFromCBORInt(v))
	case t == TypeOfColumnID:
		return de.UnpackColumnID(backReferenceFromCBORInt(v))
	case t == TypeOfCoin:
		return de.UnpackCoin(backReferenceFromCBORInt(v))
	case t == TypeOfCoinID:
		return de.UnpackCoinID(backReferenceFromCBORInt(v))
	case t.Implements(TypeOfQuery) && t.Kind() != reflect.Interface:
		return de.UnpackQuery(backReferenceFromCBORInt(v), t)
	case t == TypeOfQueryID:
		return de.UnpackQueryID(backReferenceFromCBORInt(v))
	case t == TypeOfCompiledIOP:
		return de.UnpackCompiledIOP(backReferenceFromCBORInt(v))
	case t == TypeOfStore:
		return de.UnpackStore(backReferenceFromCBORInt(v))
	}

	switch t.Kind() {
	case reflect.Bool, reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32,
		reflect.Int64, reflect.Uint, reflect.Uint8, reflect.Uint16,
		reflect.Uint32, reflect.Uint64, reflect.String:
		return de.UnpackPrimitive(v, t), nil
	case reflect.Array, reflect.Slice:
		return de.UnpackArrayOrSlice(v.([]any), t)
	case reflect.Map:
		v := v.(map[any]any)
		return de.UnpackMap(v, t)
	case reflect.Interface:
		v := v.(map[interface{}]any)
		return de.UnpackInterface(v, t)
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

func (s *Serializer) PackColumn(c column.Natural) (BackReference, error) {

	cid := c.UUID()

	if _, ok := s.columnMap[cid]; !ok {
		packed := c.Pack()
		marshaled, err := s.PackStructObject(reflect.ValueOf(packed))
		if err != nil {
			return 0, fmt.Errorf("could not marshal column %q: %v", cid, err)
		}
		s.PackedObject.Columns = append(s.PackedObject.Columns, marshaled)
		s.columnMap[cid] = len(s.PackedObject.Columns) - 1
	}

	return BackReference(
		s.columnMap[cid],
	), nil
}

func (de *Deserializer) UnpackColumn(v BackReference) (reflect.Value, error) {

	if v < 0 || int(v) >= len(de.PackedObject.Columns) {
		return reflect.Value{}, fmt.Errorf("invalid column: %v", v)
	}

	// It's the first time that 'd' sees the column: it unpacks it from the
	// pre-unmarshalled object
	if de.Columns[v] == nil {

		packedStruct := de.PackedObject.Columns[v]
		packedNatVal, err := de.UnpackStructObject(packedStruct, TypeOfPackedColumn)
		if err != nil {
			return reflect.Value{}, fmt.Errorf("could not scan the unpacked column: %w", err)
		}

		packedNat := packedNatVal.Interface().(column.PackedNatural)
		nat := packedNat.Unpack()
		de.Columns[v] = &nat
	}

	return reflect.ValueOf(*de.Columns[v]), nil
}

func (s *Serializer) PackColumnID(c ifaces.ColID) (BackReference, error) {

	if _, ok := s.columnIdMap[string(c)]; !ok {
		s.PackedObject.ColumnIDs = append(s.PackedObject.ColumnIDs, string(c))
		s.columnIdMap[string(c)] = len(s.PackedObject.ColumnIDs) - 1
	}

	return BackReference(s.columnIdMap[string(c)]), nil
}

func (de *Deserializer) UnpackColumnID(v BackReference) (reflect.Value, error) {

	if v < 0 || int(v) >= len(de.PackedObject.ColumnIDs) {
		return reflect.Value{}, fmt.Errorf("invalid column ID: %v", v)
	}

	res := ifaces.ColID(de.PackedObject.ColumnIDs[v])
	return reflect.ValueOf(res), nil
}

func (s *Serializer) PackCoin(c coin.Info) (BackReference, error) {

	if _, ok := s.coinMap[c.UUID()]; !ok {
		s.PackedObject.Coins = append(s.PackedObject.Coins, s.AsPackedCoin(c))
		s.coinMap[c.UUID()] = len(s.PackedObject.Coins) - 1
	}

	return BackReference(s.coinMap[c.UUID()]), nil
}

func (de *Deserializer) UnpackCoin(v BackReference) (reflect.Value, error) {

	if v < 0 || int(v) >= len(de.PackedObject.Coins) {
		return reflect.Value{}, fmt.Errorf("invalid coin: %v", v)
	}

	if de.Coins[v] == nil {

		packedCoin := de.PackedObject.Coins[v]

		sizes := []int{}

		if packedCoin.Size > 0 {
			sizes = append(sizes, packedCoin.Size)
		}

		if packedCoin.UpperBound > 0 {
			sizes = append(sizes, int(packedCoin.UpperBound))
		}

		unpacked := coin.NewInfo(
			coin.Name(packedCoin.Name),
			coin.Type(packedCoin.Type),
			packedCoin.Round,
			sizes...,
		)

		de.Coins[v] = &unpacked
	}

	return reflect.ValueOf(*de.Coins[v]), nil
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

func (s *Serializer) PackCoinID(c coin.Name) (BackReference, error) {

	if _, ok := s.coinIdMap[string(c)]; !ok {
		s.PackedObject.CoinIDs = append(s.PackedObject.CoinIDs, string(c))
		s.coinIdMap[string(c)] = len(s.PackedObject.CoinIDs) - 1
	}

	return BackReference(s.coinIdMap[string(c)]), nil
}

func (s *Deserializer) UnpackCoinID(v BackReference) (reflect.Value, error) {

	if v < 0 || int(v) >= len(s.PackedObject.CoinIDs) {
		return reflect.Value{}, fmt.Errorf("invalid coin ID: %v", v)
	}

	res := coin.Name(s.PackedObject.CoinIDs[v])
	return reflect.ValueOf(res), nil
}

func (s *Serializer) PackQuery(q ifaces.Query) (BackReference, error) {

	if _, ok := s.queryMap[q.UUID()]; !ok {

		valueOfQ := reflect.ValueOf(q)
		if valueOfQ.Type().Kind() == reflect.Ptr {
			valueOfQ = valueOfQ.Elem()
		}

		obj, err := s.PackStructObject(valueOfQ)
		if err != nil {
			return 0, fmt.Errorf("could not pack query, type=%T : %w", q, err)
		}
		s.PackedObject.Queries = append(s.PackedObject.Queries, obj)
		s.queryMap[q.UUID()] = len(s.PackedObject.Queries) - 1
	}

	return BackReference(s.queryMap[q.UUID()]), nil
}

func (de *Deserializer) UnpackQuery(v BackReference, t reflect.Type) (reflect.Value, error) {

	typeConcrete := t
	if t.Kind() == reflect.Ptr {
		typeConcrete = t.Elem()
	}

	if !t.Implements(TypeOfQuery) || typeConcrete.Kind() != reflect.Struct {
		return reflect.Value{}, fmt.Errorf("invalid query type: %v", t.String())
	}

	if v < 0 || int(v) >= len(de.PackedObject.Queries) {
		return reflect.Value{}, fmt.Errorf("invalid query: %v", v)
	}

	if de.Queries[v] == nil {

		packedQuery := de.PackedObject.Queries[v]
		query, err := de.UnpackStructObject(packedQuery, typeConcrete)
		if err != nil {
			return reflect.Value{}, fmt.Errorf("could not scan the unpacked query: %w", err)
		}

		if t.Kind() == reflect.Ptr {
			query = query.Addr()
		}

		q := query.Interface().(ifaces.Query)
		de.Queries[v] = &q
	}

	var (
		qIfaces = *de.Queries[v]
		qValue  = reflect.ValueOf(qIfaces)
	)

	if qValue.Type() != t {
		return reflect.Value{}, fmt.Errorf("the deserialized query does not have the expected type, %v != %v", t.String(), qValue.Type().String())
	}

	return qValue, nil
}

func (s *Serializer) PackQueryID(q ifaces.QueryID) (BackReference, error) {

	if _, ok := s.queryIDMap[string(q)]; !ok {
		s.PackedObject.QueryIDs = append(s.PackedObject.QueryIDs, string(q))
		s.queryIDMap[string(q)] = len(s.PackedObject.QueryIDs) - 1
	}

	return BackReference(s.queryIDMap[string(q)]), nil
}

func (de *Deserializer) UnpackQueryID(v BackReference) (reflect.Value, error) {

	if v < 0 || int(v) >= len(de.PackedObject.QueryIDs) {
		return reflect.Value{}, fmt.Errorf("invalid query ID: %v", v)
	}

	res := ifaces.QueryID(de.PackedObject.QueryIDs[v])
	return reflect.ValueOf(res), nil
}

func (s *Serializer) PackCompiledIOP(comp *wizard.CompiledIOP) (any, error) {

	if _, ok := s.compiledIOPs[comp]; !ok {
		obj, err := s.PackStructObject(reflect.ValueOf(*comp))
		if err != nil {
			return nil, fmt.Errorf("could not pack compiled IOP: %w", err)
		}
		s.PackedObject.CompiledIOP = append(s.PackedObject.CompiledIOP, obj)
		s.compiledIOPs[comp] = len(s.PackedObject.CompiledIOP) - 1
	}

	return BackReference(s.compiledIOPs[comp]), nil
}

func (s *Deserializer) UnpackCompiledIOP(v BackReference) (reflect.Value, error) {

	if v < 0 || int(v) >= len(s.PackedObject.CompiledIOP) {
		return reflect.Value{}, fmt.Errorf("invalid compiled IOP: %v", v)
	}

	if s.CompiledIOPs[v] == nil {

		packedCompiledIOP := s.PackedObject.CompiledIOP[v]
		compiledIOP, err := s.UnpackStructObject(packedCompiledIOP, TypeOfCompiledIOP)
		if err != nil {
			return reflect.Value{}, fmt.Errorf("could not scan the unpacked compiled IOP: %w", err)
		}

		c := compiledIOP.Addr().Interface().(*wizard.CompiledIOP)
		s.CompiledIOPs[v] = c
	}

	return reflect.ValueOf(s.CompiledIOPs[v]), nil
}

func (ser *Serializer) PackStore(s *column.Store) (BackReference, error) {

	if _, ok := ser.Stores[s]; !ok {
		packedStore := s.Pack()
		obj, err := ser.PackArrayOrSlice(reflect.ValueOf(packedStore))
		if err != nil {
			return 0, fmt.Errorf("could not pack store: %w", err)
		}
		ser.PackedObject.Store = append(ser.PackedObject.Store, obj)
		ser.Stores[s] = len(ser.PackedObject.Store) - 1
	}

	return BackReference(ser.Stores[s]), nil
}

func (de *Deserializer) UnpackStore(v BackReference) (reflect.Value, error) {

	if v < 0 || int(v) >= len(de.PackedObject.Store) {
		return reflect.Value{}, fmt.Errorf("invalid store: %v", v)
	}

	if de.Stores[v] == nil {

		preStore := de.PackedObject.Store[v]
		storeArr, err := de.UnpackArrayOrSlice(preStore, TypeOfPackedStore)
		if err != nil {
			return reflect.Value{}, fmt.Errorf("could not scan the unpacked store: %w", err)
		}

		pStore := storeArr.Interface().(column.PackedStore)
		de.Stores[v] = pStore.Unpack()
	}

	return reflect.ValueOf(de.Stores[v]), nil
}

// PackArrayOrSlice serializes arrays or slices by recursively serializing each element.
func (s *Serializer) PackArrayOrSlice(v reflect.Value) ([]any, error) {

	res := make([]any, v.Len())
	var globalErr error

	for i := 0; i < v.Len(); i++ {
		ri, err := s.PackValue(v.Index(i))
		if err != nil {
			globalErr = errors.Join(globalErr, fmt.Errorf("position %d: %w", i, err))
		}
		res[i] = ri
	}

	if globalErr != nil {
		return nil, fmt.Errorf("failed to serialize array or slice: %w", globalErr)
	}

	return res, nil
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

	var globalErr error

	subType := t.Elem()
	for i := 0; i < len(v); i++ {
		subV, err := de.UnpackValue(v[i], subType)
		if err != nil {
			err := fmt.Errorf("failed to deserialize to %q, error in element %d: %w", t.Name(), i, err)
			globalErr = errors.Join(globalErr, err)
			continue
		}
		res.Index(i).Set(subV)
	}

	if globalErr != nil {
		return reflect.Value{}, globalErr
	}

	return res, nil
}

func (s *Serializer) PackStructSchema(t reflect.Type) (schema PackedStructSchema, err error) {

	if t.Kind() != reflect.Struct {
		return PackedStructSchema{}, fmt.Errorf("s.Kind() != reflect.Struct, type=%v", t.String())
	}

	cleanTypeString := getPkgPathAndTypeName(t)

	if i, ok := s.structSchemaMap[cleanTypeString]; ok {
		return s.PackedObject.StructSchema[i], nil
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

	return schema, nil
}

func (s *Serializer) PackInterface(v reflect.Value) (any, error) {

	var (
		concrete          = v.Elem()
		cleanConcreteType = getPkgPathAndTypeNameIndirect(concrete.Interface())
	)

	if _, err := findRegisteredImplementation(cleanConcreteType); err != nil {
		return nil, fmt.Errorf("attempted to serialize unregistered type repr=%q type=%v: %w", cleanConcreteType, concrete.Type().String(), err)
	}

	if _, ok := s.typeMap[cleanConcreteType]; !ok {
		s.PackedObject.Types = append(s.PackedObject.Types, cleanConcreteType)
		s.typeMap[cleanConcreteType] = len(s.PackedObject.Types) - 1
	}

	packedConcrete, err := s.PackValue(concrete)
	if err != nil {
		return nil, fmt.Errorf("could not marshal concrete interface value, type=%v: %w", concrete.Type().String(), err)
	}

	return PackedIFace{
		Type:     s.typeMap[cleanConcreteType],
		Concrete: packedConcrete,
	}, nil
}

func (de *Deserializer) UnpackInterface(pi map[interface{}]interface{}, t reflect.Type) (reflect.Value, error) {

	var (
		ctype, ok = pi["t"].(uint64)
		concrete  = pi["c"]
	)

	if !ok || int(ctype) >= len(de.PackedObject.Types) {
		return reflect.Value{}, fmt.Errorf("invalid type: %v", ctype)
	}

	cleanConcreteType := de.PackedObject.Types[ctype]
	refType, err := findRegisteredImplementation(cleanConcreteType)
	if err != nil {
		return reflect.Value{}, fmt.Errorf("unregistered type %q: %w", cleanConcreteType, err)
	}

	if !refType.Implements(t) {
		return reflect.Value{}, fmt.Errorf("the resolved type does not implement the target interface, %v ~ %v", refType.String(), t.String())
	}

	cres, err := de.UnpackValue(concrete, refType)
	if err != nil {
		return reflect.Value{}, fmt.Errorf("could not deserialize interface object, type=%v, err=%w", t.String(), err)
	}

	// Create a new reflect.Value for the interface type
	// Reminder; here the important thing is to ensure that the returned
	// Value actually bears the requested interface type and not the
	// concrete type.
	ifaceValue := reflect.New(t).Elem()
	ifaceValue.Set(cres)

	return ifaceValue, nil
}

func (s *Serializer) PackStructObject(obj reflect.Value) (PackedStructObject, error) {

	if obj.Kind() != reflect.Struct {
		utils.Panic("obj.Kind() != reflect.Struct, type=%v", obj.Type().String())
	}

	values := make([]any, obj.NumField())
	var globalErr error

	// Note, since we don't want to register the schema before going through
	// all the components, we have to rely on the fact that schema and this loop
	// declare the fields in the same order.
	for i := 0; i < obj.NumField(); i++ {
		field := obj.Field(i)

		if !obj.Type().Field(i).IsExported() {
			s.warnf(fmt.Sprintf("field %v.%v is not exported", obj.Type().String(), obj.Type().Field(i).Name))
			continue
		}

		vi, err := s.PackValue(field)
		if err != nil {
			globalErr = errors.Join(globalErr, fmt.Errorf("field name=%s type=%v: %w", obj.Type().Field(i).Name, field.Type().String(), err))
		}
		values[i] = vi
	}

	if globalErr != nil {
		return PackedStructObject{}, fmt.Errorf("failed to pack struct object, type=%v: %w", obj.Type().String(), globalErr)
	}

	if _, err := s.PackStructSchema(obj.Type()); err != nil {
		return PackedStructObject{}, fmt.Errorf("failed to pack struct schema: %w", err)
	}

	// Importantly, we want to be sure that all the component have been
	// converted before we convert the current type. That way, we can ensure
	// that all the information of the dependency is added prior to the
	// information on the dependant struct. This is necessary for seria-
	// lization.
	return values, nil
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
		schema = de.PackedObject.StructSchema[schemaID]
	)

	// To ease debugging, all the errors for all the fields are joined and
	// wrapped in a single error.
	var globalErr error

	for i := range v {

		structField, ok := t.FieldByName(schema.Fields[i])
		if !ok {
			return reflect.Value{}, fmt.Errorf("invalid field name: %v, it was not found in the struct %v", schema.Fields[i], t.String())
		}

		if !structField.IsExported() {
			de.warnf(fmt.Sprintf("field %v.%v is not exported", t.String(), structField.Name))
			continue
		}

		field := res.FieldByName(schema.Fields[i])
		value, err := de.UnpackValue(v[i], field.Type())

		if err != nil {
			err = fmt.Errorf("field %q type=%v, err=%w", schema.Fields[i], field.Type().String(), err)
			globalErr = errors.Join(globalErr, err)
			continue
		}

		field.Set(value)
	}

	if globalErr != nil {
		return reflect.Value{}, fmt.Errorf("failed to deserialize struct: %w", globalErr)
	}

	return res, nil
}

func (s *Serializer) PackMap(obj reflect.Value) (map[any]any, error) {

	if obj.Kind() != reflect.Map {
		return nil, fmt.Errorf("obj.Kind() != reflect.Map, type=%v", obj.Type().String())
	}

	var (
		keys       = obj.MapKeys()
		keyStrings = make([]string, 0, len(keys))
		keyMap     = make(map[string]reflect.Value, len(keys))
	)

	for _, k := range keys {
		keyString := k.String()
		keyStrings = append(keyStrings, keyString)
		keyMap[keyString] = k
	}

	// Needed for deterministic encoding
	sort.Strings(keyStrings)

	var (
		res       = make(map[any]any, len(keys))
		globalErr error
	)

	for _, keyString := range keyStrings {
		k := keyMap[keyString]
		rk, err := s.PackValue(obj.MapIndex(k))
		if err != nil {
			globalErr = errors.Join(globalErr, fmt.Errorf("key %q type=%v: %w", keyString, k.Type().String(), err))
			continue
		}
		res[keyString] = rk
	}

	if globalErr != nil {
		return nil, fmt.Errorf("failed to pack map, type=%v: %w", obj.Type().String(), globalErr)
	}

	return res, nil
}

func (de *Deserializer) UnpackMap(v map[any]any, t reflect.Type) (reflect.Value, error) {

	if t.Kind() != reflect.Map {
		return reflect.Value{}, fmt.Errorf("invalid type: %v", t.String())
	}

	var (
		typeOfKey   = t.Key()
		typeOfValue = t.Elem()
		res         = reflect.MakeMap(t)
		globalErr   error
	)

	for key, val := range v {
		k, err := de.UnpackValue(key, typeOfKey)
		if err != nil {
			return reflect.Value{}, fmt.Errorf("could not deserialize map key %q type=%v, err=%w", key, typeOfKey.String(), err)
		}

		v, err := de.UnpackValue(val, typeOfValue)
		if err != nil {
			err = fmt.Errorf("could not deserialize map value %q type=%v, err=%w", val, typeOfValue.String(), err)
			globalErr = errors.Join(globalErr, err)
			continue
		}

		res.SetMapIndex(k, v)
	}

	if globalErr != nil {
		return reflect.Value{}, fmt.Errorf("failed to deserialize map: %w", globalErr)
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

	if value.CanAddr() {
		return value.Addr(), nil
	}

	// If the value cannot be addressed, it means that it is a temp value. And
	// we can attempt to fix this by resetting the value to what we wanted.
	// Note that reflect.New returns a pointer to the provided type, which is
	// what we wanted.
	res := reflect.New(t.Elem())
	res.Elem().Set(value)
	return res, nil
}

func (de *Deserializer) UnpackPrimitive(v any, t reflect.Type) reflect.Value {
	return reflect.ValueOf(v).Convert(t)
}

func (ser *Serializer) warnf(warning string) {
	ser.Warnings = append(ser.Warnings, warning)
}

func (de *Deserializer) warnf(warning string) {
	de.Warnings = append(de.Warnings, warning)
}

func backReferenceFromCBORInt(n any) BackReference {
	return BackReference(n.(uint64))
}
