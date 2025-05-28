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

type BackReference struct {
	Type BackReferenceType `cbor:"t"`
	ID   int               `cbor:"i"`
}

type Serializer struct {
	PreMarshaledObject *PackedObject
	TypeMap            map[string]int
	StructSchemaMap    map[string]int
	CoinMap            map[string]int
	ColumnMap          map[string]int
	ColumnIdMap        map[string]int
	QueryMap           map[string]int
	QueryIDMap         map[string]int
	CompiledIOPs       map[*wizard.CompiledIOP]int
	Stores             map[*column.Store]int
	CurrentCompiledIOP int
}

type Deserializer struct {
	Columns               []*column.Natural
	Coins                 []*coin.Info
	Queries               []*ifaces.Query
	CompiledIOP           []*wizard.CompiledIOP
	CurrentCompiledIOP    *wizard.CompiledIOP
	PreUnmarshalledObject *PackedObject
}

type PackedObject struct {
	Types        []string                    `cbor:"t"`
	StructSchema []MarshalizableStructSchema `cbor:"s"`
	ColumnIDs    []string                    `cbor:"id"`
	Columns      []MarshalizableStructObject `cbor:"c"`
	CoinIDs      []string                    `cbor:"i"`
	Coins        []PackedCoin                `cbor:"o"`
	QueryIDs     []string                    `cbor:"w"`
	Store        [][]any                     `cbor:"st"`
	Queries      []any                       `cbor:"q"`
	CompiledIOP  []any                       `cbor:"a"`
	Payload      []byte                      `cbor:"p"`
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

type MarshalizableStructSchema struct {
	Type   string   `cbor:"t"`
	Fields []string `cbor:"f"`
}

type MarshalizableStructObject []any

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

func (s *Serializer) PackColumn(c column.Natural) BackReference {

	cid := c.PackedIdentifier()

	if _, ok := s.ColumnMap[string(cid)]; !ok {
		packed := c.Pack()
		marshaled := s.PackStructObject(reflect.ValueOf(packed))
		s.PreMarshaledObject.Columns = append(s.PreMarshaledObject.Columns, marshaled)
		s.ColumnMap[string(cid)] = len(s.PreMarshaledObject.Columns) - 1
	}

	return BackReference{
		Type: ColumnBackReference,
		ID:   s.ColumnMap[string(cid)],
	}
}

func (s *Serializer) PackColumnID(c ifaces.ColID) BackReference {

	if _, ok := s.ColumnIdMap[string(c)]; !ok {
		s.PreMarshaledObject.ColumnIDs = append(s.PreMarshaledObject.ColumnIDs, string(c))
		s.ColumnIdMap[string(c)] = len(s.PreMarshaledObject.ColumnIDs) - 1
	}

	return BackReference{
		Type: ColumnIDBackReference,
		ID:   s.ColumnIdMap[string(c)],
	}
}

func (s *Serializer) PackCoin(c coin.Info) BackReference {

	if _, ok := s.CoinMap[string(c.Name)]; !ok {
		s.PreMarshaledObject.Coins = append(s.PreMarshaledObject.Coins, s.AsPackedCoin(c))
		s.CoinMap[string(c.Name)] = len(s.PreMarshaledObject.Coins) - 1
	}

	return BackReference{
		Type: CoinBackReference,
		ID:   s.CoinMap[string(c.Name)],
	}
}

func (s *Serializer) PackCoinID(c coin.Name) BackReference {

	if _, ok := s.CoinMap[string(c)]; !ok {
		utils.Panic("serializing a coin ID for an unknown coin: %v", c)
	}

	return BackReference{
		Type: CoinIDBackReference,
		ID:   s.CoinMap[string(c)],
	}
}

func (s *Serializer) PackQuery(q ifaces.Query) BackReference {

	if _, ok := s.QueryMap[string(q.Name())]; !ok {

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

		s.PreMarshaledObject.Queries = append(s.PreMarshaledObject.Queries, obj)
		s.QueryMap[string(q.Name())] = len(s.PreMarshaledObject.Queries) - 1
	}

	return BackReference{
		Type: QueryBackReference,
		ID:   s.QueryMap[string(q.Name())],
	}
}

func (s *Serializer) PackQueryID(q ifaces.QueryID) BackReference {

	if _, ok := s.QueryMap[string(q)]; !ok {
		s.PreMarshaledObject.QueryIDs = append(s.PreMarshaledObject.QueryIDs, string(q))
		s.QueryIDMap[string(q)] = len(s.PreMarshaledObject.QueryIDs) - 1
	}

	return BackReference{
		Type: QueryIDBackReference,
		ID:   s.QueryMap[string(q)],
	}
}

func (s *Serializer) PackCompiledIOP(comp *wizard.CompiledIOP) any {

	if _, ok := s.CompiledIOPs[comp]; !ok {
		obj := s.PackStructObject(reflect.ValueOf(*comp))
		s.PreMarshaledObject.CompiledIOP = append(s.PreMarshaledObject.CompiledIOP, obj)
		s.CompiledIOPs[comp] = len(s.PreMarshaledObject.CompiledIOP) - 1
	}

	return BackReference{
		Type: CompiledIOPBackReference,
		ID:   s.CompiledIOPs[comp],
	}
}

func (ser *Serializer) PackStore(s *column.Store) any {

	if _, ok := ser.Stores[s]; !ok {
		obj := ser.PackStructObject(reflect.ValueOf(*s))
		ser.PreMarshaledObject.Store = append(ser.PreMarshaledObject.Store, obj)
		ser.Stores[s] = len(ser.PreMarshaledObject.Store) - 1
	}

	return BackReference{
		Type: StoreBackReference,
		ID:   ser.Stores[s],
	}
}

// PackArrayOrSlice serializes arrays or slices by recursively serializing each element.
func (s *Serializer) PackArrayOrSlice(v reflect.Value) []any {
	res := make([]any, v.Len())
	for i := 0; i < v.Len(); i++ {
		res[i] = s.PackValue(v.Index(i))
	}
	return res
}

func (s *Serializer) AsPackedCoin(c coin.Info) PackedCoin {
	return PackedCoin{
		Type:        int8(c.Type),
		Size:        c.Size,
		UpperBound:  int32(c.UpperBound),
		Name:        string(c.Name),
		Round:       c.Round,
		CompiledIOP: int8(s.CurrentCompiledIOP),
	}
}

func (s *Serializer) PackStructSchema(t reflect.Type) (marshalizable MarshalizableStructSchema, index int) {

	if t.Kind() != reflect.Struct {
		utils.Panic("s.Kind() != reflect.Struct, type=%v", t.String())
	}

	cleanTypeString := getPkgPathAndTypeName(s)

	if i, ok := s.StructSchemaMap[cleanTypeString]; ok {
		return s.PreMarshaledObject.StructSchema[i], i
	}

	marshalizable = MarshalizableStructSchema{
		Type:   cleanTypeString,
		Fields: make([]string, t.NumField()),
	}

	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		marshalizable.Fields[i] = field.Name
	}

	positionOfType := len(s.PreMarshaledObject.StructSchema)
	s.StructSchemaMap[cleanTypeString] = positionOfType
	s.PreMarshaledObject.StructSchema = append(s.PreMarshaledObject.StructSchema, marshalizable)

	return marshalizable, positionOfType
}

func (s *Serializer) PackInterface(v reflect.Value) any {

	var (
		concrete          = v.Elem()
		cleanConcreteType = getPkgPathAndTypeName(concrete.Interface())
	)

	if _, ok := s.TypeMap[cleanConcreteType]; !ok {
		s.PreMarshaledObject.Types = append(s.PreMarshaledObject.Types, cleanConcreteType)
		s.TypeMap[cleanConcreteType] = len(s.PreMarshaledObject.Types) - 1
	}

	return PackedIFace{
		Type:     s.TypeMap[cleanConcreteType],
		Concrete: s.PackValue(concrete),
	}
}

func (s *Serializer) PackStructObject(obj reflect.Value) MarshalizableStructObject {

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

// UnpackBackReference unpacks a backreference.
func UnpackBackReference(v []any) (BackReference, error) {

	if len(v) != 2 {
		return BackReference{}, fmt.Errorf("invalid backreference: %v", v)
	}

	return BackReference{
		Type: v[0].(BackReferenceType),
		ID:   v[1].(int),
	}, nil
}

// UnmarshalColumn unpacks a column. It expects v to represent a backreference.
func (s *Deserializer) UnmarshalColumn(v []any) (column.Natural, error) {

	backRef, err := UnpackBackReference(v)
	if err != nil {
		return column.Natural{}, fmt.Errorf("could not deserialize column, %v", err)
	}

	if backRef.Type != ColumnBackReference {
		return column.Natural{}, fmt.Errorf("invalid backreference type: %v", backRef.Type)
	}

	if backRef.ID >= len(s.Columns) {
		return column.Natural{}, fmt.Errorf("invalid backreference ID: %v, #columns: %v", backRef.ID, len(s.Columns))
	}

	if s.Columns[backRef.ID] != nil {
		return *s.Columns[backRef.ID], nil
	}

	// Here, it is the first time we encounter the column. We need to unpack it
	// from the pre-unmarshalled object.
	panic("unimplemented")

}
