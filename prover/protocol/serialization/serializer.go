package serialization

import (
	"errors"
	"fmt"
	"math/big"
	"reflect"
	"sort"

	"github.com/consensys/linea-monorepo/prover/maths/field"
	"github.com/consensys/linea-monorepo/prover/protocol/coin"
	"github.com/consensys/linea-monorepo/prover/protocol/column"
	"github.com/consensys/linea-monorepo/prover/protocol/ifaces"
	"github.com/consensys/linea-monorepo/prover/protocol/wizard"
	"github.com/consensys/linea-monorepo/prover/symbolic"
	"github.com/consensys/linea-monorepo/prover/utils"
	"github.com/google/uuid"
)

// Global type constants for reflection-based type checking.
// These define the reflect.Type of key protocol-specific types, used to identify
// special types during serialization and deserialization.
var (
	TypeOfColumnNatural    = reflect.TypeOf(column.Natural{})
	TypeOfColumnID         = reflect.TypeOf(ifaces.ColID(""))
	TypeOfCoin             = reflect.TypeOf(coin.Info{})
	TypeOfCoinID           = reflect.TypeOf(coin.Name(""))
	TypeOfQuery            = reflect.TypeOf((*ifaces.Query)(nil)).Elem()
	TypeOfQueryID          = reflect.TypeOf(ifaces.QueryID(""))
	TypeOfCompiledIOP      = reflect.TypeOf(&wizard.CompiledIOP{})
	TypeOfStore            = reflect.TypeOf(&column.Store{})
	TypeOfPackedColumn     = reflect.TypeOf(column.PackedNatural{})
	TypeOfPackedStore      = reflect.TypeOf(column.PackedStore{})
	TypeOfVariableMetadata = reflect.TypeOf((*symbolic.Metadata)(nil)).Elem()
	TypeOfArrayOfExpr      = reflect.TypeOf([]*symbolic.Expression{})
	TypeOfExpression       = reflect.TypeOf(&symbolic.Expression{})
	TypeOfArrayOfInt       = reflect.TypeOf([]int{})
	TypeOfFieldElement     = reflect.TypeOf(field.Element{})
	TypeOfBigInt           = reflect.TypeOf(&big.Int{})
)

// BackReference represents an integer index into PackedObject arrays (e.g., Columns, Coins).
// It enables efficient reuse of shared objects, avoiding redundant serialization.
type BackReference int

// Serializer manages the serialization process, packing objects into a PackedObject.
// It tracks references to objects (e.g., columns, coins) and collects warnings for non-fatal issues.
type Serializer struct {
	PackedObject    *PackedObject               // The output structure containing serialized data.
	typeMap         map[string]int              // Maps type names to indices in PackedObject.Types.
	structSchemaMap map[string]int              // Maps struct type names to indices in PackedObject.StructSchema.
	coinMap         map[uuid.UUID]int           // Maps coin UUIDs to indices in PackedObject.Coins.
	coinIdMap       map[string]int              // Maps coin IDs to indices in PackedObject.CoinIDs.
	columnMap       map[uuid.UUID]int           // Maps column UUIDs to indices in PackedObject.Columns.
	columnIdMap     map[string]int              // Maps column IDs to indices in PackedObject.ColumnIDs.
	queryMap        map[uuid.UUID]int           // Maps query UUIDs to indices in PackedObject.Queries.
	queryIDMap      map[string]int              // Maps query IDs to indices in PackedObject.QueryIDs.
	compiledIOPs    map[*wizard.CompiledIOP]int // Maps CompiledIOP pointers to indices in PackedObject.CompiledIOP.
	Stores          map[*column.Store]int       // Maps Store pointers to indices in PackedObject.Store.
	Warnings        []string                    // Collects warnings (e.g., unexported fields) for debugging.
}

// Deserializer manages the deserialization process, reconstructing objects from a PackedObject.
// It caches reconstructed objects to resolve back-references and collects warnings.
type Deserializer struct {
	StructSchemaMap map[string]int        // Maps struct type names to indices in PackedObject.StructSchema.
	Columns         []*column.Natural     // Cache of deserialized columns, indexed by BackReference.
	Coins           []*coin.Info          // Cache of deserialized coins.
	Queries         []*ifaces.Query       // Cache of deserialized queries.
	CompiledIOPs    []*wizard.CompiledIOP // Cache of deserialized CompiledIOPs.
	Stores          []*column.Store       // Cache of deserialized stores.
	PackedObject    *PackedObject         // The input structure to deserialize.
	Warnings        []string              // Collects warnings for debugging.
}

// PackedObject is the serialized representation of data, designed for CBOR encoding.
// It stores type metadata, objects, and a payload for the root serialized value.
type PackedObject struct {
	Types        []string             `cbor:"t"`  // Type names for interfaces.
	StructSchema []PackedStructSchema `cbor:"s"`  // Schemas for structs (type and field names).
	ColumnIDs    []string             `cbor:"id"` // String IDs for columns.
	Columns      []PackedStructObject `cbor:"c"`  // Serialized columns (as PackedStructObject).
	CoinIDs      []string             `cbor:"i"`  // String IDs for coins.
	Coins        []PackedCoin         `cbor:"o"`  // Serialized coins.
	QueryIDs     []string             `cbor:"w"`  // String IDs for queries.
	Queries      []PackedStructObject `cbor:"q"`  // Serialized queries.
	Store        [][]any              `cbor:"st"` // Serialized stores (as arrays).
	CompiledIOP  []PackedStructObject `cbor:"a"`  // Serialized CompiledIOPs.
	Payload      []byte               `cbor:"p"`  // CBOR-encoded root value.
}

// PackedIFace serializes an interface value, storing its type index and concrete value.
type PackedIFace struct {
	Type     int `cbor:"t"` // Index into PackedObject.Types.
	Concrete any `cbor:"c"` // Serialized concrete value.
}

// PackedCoin is a compact representation of coin.Info, optimized for CBOR encoding.
type PackedCoin struct {
	Type        int8   `cbor:"t"`           // Coin type (e.g., Random, Fixed).
	Size        int    `cbor:"s,omitempty"` // Coin size (optional).
	UpperBound  int32  `cbor:"u,omitempty"` // Upper bound for coin (optional).
	Name        string `cbor:"n"`           // Coin name.
	Round       int    `cbor:"r"`           // Round number.
	CompiledIOP int8   `cbor:"i"`           // Unused (placeholder for CompiledIOP index).
}

// PackedStructSchema defines a struct’s type and field names for deserialization.
type PackedStructSchema struct {
	Type   string   `cbor:"t"` // Type name (e.g., "pkg.Type").
	Fields []string `cbor:"f"` // Field names in declaration order.
}

// PackedStructObject is a slice of serialized field values for a struct.
type PackedStructObject []any

// Serialize is the entry point for serializing any value into CBOR-encoded bytes.
// It packs the value into a PackedObject.Payload, encodes the PackedObject, and returns the result.
// Warnings are printed for debugging.
func Serialize(v any) ([]byte, error) {
	// Initialize a new Serializer with empty maps and a PackedObject.
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

	// Pack the input value.
	payload, err := ser.PackValue(reflect.ValueOf(v))
	if err != nil {
		return nil, fmt.Errorf("could not pack the value: %w", err)
	}

	// Print any warnings (e.g., unexported fields).
	for i := range ser.Warnings {
		fmt.Println(ser.Warnings[i])
	}

	// CBOR-encode the payload.
	bytesOfPayload, err := encodeWithCBOR(payload)
	if err != nil {
		return nil, fmt.Errorf("could not encode the payload with CBOR: %v", err)
	}

	// Store the encoded payload in PackedObject.
	packedObject := ser.PackedObject
	packedObject.Payload = bytesOfPayload

	// CBOR-encode the entire PackedObject.
	bytesOfV, err := encodeWithCBOR(packedObject)
	if err != nil {
		return nil, fmt.Errorf("could not encode the packedObject with CBOR: %v", err)
	}

	return bytesOfV, nil
}

// Deserialize is the entry point for deserializing CBOR-encoded bytes into a pointer.
// It decodes the bytes into a PackedObject, unpacks the Payload, and sets the result into v.
// Warnings are printed for debugging.
func Deserialize(bytes []byte, v any) error {
	// Create a new PackedObject to decode into.
	packedObject := &PackedObject{}

	// Ensure v is a pointer.
	if reflect.TypeOf(v).Kind() != reflect.Ptr {
		return fmt.Errorf("invalid type: %v, expected a pointer", reflect.TypeOf(v).String())
	}

	// Decode the bytes into PackedObject.
	if err := decodeWithCBOR(bytes, packedObject); err != nil {
		return fmt.Errorf("the serialized object does not have the [PackedObject] format, err=%w", err)
	}

	// Initialize a Deserializer with pre-allocated caches.
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
		payloadType = reflect.TypeOf(v).Elem() // Target type (dereferenced).
	)

	// Decode the Payload into a root value.
	if err := decodeWithCBOR(packedObject.Payload, &payloadRoot); err != nil {
		return fmt.Errorf("could not deserialize the payload, err=%w", err)
	}

	// Register struct schemas in StructSchemaMap.
	for i, t := range packedObject.StructSchema {
		deser.StructSchemaMap[t.Type] = i
	}

	// Unpack the root value into the target type.
	res, err := deser.UnpackValue(payloadRoot, payloadType)
	if err != nil {
		return fmt.Errorf("could not deserialize the payload, err=%w", err)
	}

	// Print warnings.
	for i := range deser.Warnings {
		fmt.Println(deser.Warnings[i])
	}

	// Debugging output (remove in production).
	fmt.Printf("value of res = %v, value of v = %T\n", res.Type(), v)

	// Set the result into the provided pointer.
	valueOfV := reflect.ValueOf(v)
	valueOfV.Elem().Set(res)
	return nil
}

// PackValue recursively serializes a reflect.Value into a serializable form.
// It handles protocol-specific types (e.g., columns, coins) and generic types (e.g., structs, slices).
// Returns the serialized value or an error.
func (s *Serializer) PackValue(v reflect.Value) (any, error) {
	// This captures the case where the value is nil to begin with
	if !v.IsValid() || v.Interface() == nil {
		return nil, nil
	}

	typeOfV := v.Type()

	fmt.Printf("type of v = %v\n", typeOfV.String())

	// Identify custom codexes
	if codex, ok := CustomCodexes[typeOfV]; ok {
		return codex.Ser(s, v)
	}

	// Handle protocol-specific types.
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

	// Handle generic Go types.
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
		// Panic for unsupported types
		panic(fmt.Sprintf("unsupported type kind: %v", v.Kind()))
	}
}

// UnpackValue recursively deserializes a value into a target reflect.Type.
// It resolves back-references for protocol-specific types and handles generic types.
// Returns the deserialized reflect.Value or an error.
func (de *Deserializer) UnpackValue(v any, t reflect.Type) (r reflect.Value, e error) {
	if v == nil {
		return reflect.Zero(t), nil
	}

	// Identify custom codexes
	if codex, ok := CustomCodexes[t]; ok {
		return codex.Des(de, v, t)
	}

	// Handle protocol-specific types.
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

	// Handle generic Go types.
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

// PackColumn serializes a column.Natural, returning a BackReference to its index in PackedObject.Columns.
// It ensures columns are serialized once, using UUIDs for deduplication.
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

// UnpackColumn deserializes a column.Natural from a BackReference.
// It caches the result to maintain object identity.
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

// PackColumnID serializes an ifaces.ColID (string), returning a BackReference to its index in PackedObject.ColumnIDs.
func (s *Serializer) PackColumnID(c ifaces.ColID) (BackReference, error) {
	if _, ok := s.columnIdMap[string(c)]; !ok {
		s.PackedObject.ColumnIDs = append(s.PackedObject.ColumnIDs, string(c))
		s.columnIdMap[string(c)] = len(s.PackedObject.ColumnIDs) - 1
	}

	return BackReference(s.columnIdMap[string(c)]), nil
}

// UnpackColumnID deserializes an ifaces.ColID from a BackReference.
func (de *Deserializer) UnpackColumnID(v BackReference) (reflect.Value, error) {
	if v < 0 || int(v) >= len(de.PackedObject.ColumnIDs) {
		return reflect.Value{}, fmt.Errorf("invalid column ID: %v", v)
	}

	res := ifaces.ColID(de.PackedObject.ColumnIDs[v])
	return reflect.ValueOf(res), nil
}

// PackCoin serializes a coin.Info, returning a BackReference to its index in PackedObject.Coins.
func (s *Serializer) PackCoin(c coin.Info) (BackReference, error) {
	if _, ok := s.coinMap[c.UUID()]; !ok {
		s.PackedObject.Coins = append(s.PackedObject.Coins, s.AsPackedCoin(c))
		s.coinMap[c.UUID()] = len(s.PackedObject.Coins) - 1
	}

	return BackReference(s.coinMap[c.UUID()]), nil
}

// UnpackCoin deserializes a coin.Info from a BackReference, caching the result.
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

// AsPackedCoin converts a coin.Info to a PackedCoin for serialization.
func (s *Serializer) AsPackedCoin(c coin.Info) PackedCoin {
	return PackedCoin{
		Type:       int8(c.Type),
		Size:       c.Size,
		UpperBound: int32(c.UpperBound),
		Name:       string(c.Name),
		Round:      c.Round,
	}
}

// PackCoinID serializes a coin.Name (string), returning a BackReference to its index in PackedObject.CoinIDs.
func (s *Serializer) PackCoinID(c coin.Name) (BackReference, error) {
	if _, ok := s.coinIdMap[string(c)]; !ok {
		s.PackedObject.CoinIDs = append(s.PackedObject.CoinIDs, string(c))
		s.coinIdMap[string(c)] = len(s.PackedObject.CoinIDs) - 1
	}

	return BackReference(s.coinIdMap[string(c)]), nil
}

// UnpackCoinID deserializes a coin.Name from a BackReference.
func (s *Deserializer) UnpackCoinID(v BackReference) (reflect.Value, error) {
	if v < 0 || int(v) >= len(s.PackedObject.CoinIDs) {
		return reflect.Value{}, fmt.Errorf("invalid coin ID: %v", v)
	}

	res := coin.Name(s.PackedObject.CoinIDs[v])
	return reflect.ValueOf(res), nil
}

// PackQuery serializes an ifaces.Query, returning a BackReference to its index in PackedObject.Queries.
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

// UnpackQuery deserializes an ifaces.Query from a BackReference, ensuring it matches the target type.
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

// PackQueryID serializes an ifaces.QueryID (string), returning a BackReference to its index in PackedObject.QueryIDs.
func (s *Serializer) PackQueryID(q ifaces.QueryID) (BackReference, error) {
	if _, ok := s.queryIDMap[string(q)]; !ok {
		s.PackedObject.QueryIDs = append(s.PackedObject.QueryIDs, string(q))
		s.queryIDMap[string(q)] = len(s.PackedObject.QueryIDs) - 1
	}

	return BackReference(s.queryIDMap[string(q)]), nil
}

// UnpackQueryID deserializes an ifaces.QueryID from a BackReference.
func (de *Deserializer) UnpackQueryID(v BackReference) (reflect.Value, error) {
	if v < 0 || int(v) >= len(de.PackedObject.QueryIDs) {
		return reflect.Value{}, fmt.Errorf("invalid query ID: %v", v)
	}

	res := ifaces.QueryID(de.PackedObject.QueryIDs[v])
	return reflect.ValueOf(res), nil
}

// PackCompiledIOP serializes a wizard.CompiledIOP, returning a BackReference to its index in PackedObject.CompiledIOP.
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

// UnpackCompiledIOP deserializes a wizard.CompiledIOP from a BackReference, caching the result.
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

// PackStore serializes a column.Store, returning a BackReference to its index in PackedObject.Store.
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

// UnpackStore deserializes a column.Store from a BackReference, caching the result.
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
// It collects errors for all elements and returns a combined error if any fail.
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

// UnpackArrayOrSlice deserializes arrays or slices, reconstructing elements into the target type.
// It collects errors for all elements and returns a combined error if any fail.
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

// PackStructSchema creates a PackedStructSchema for a struct type, registering it in PackedObject.StructSchema.
// It returns the schema or an error if the type is not a struct.
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

// PackInterface serializes an interface value, storing its type index and concrete value in a PackedIFace.
// It ensures the concrete type is registered and returns an error if not.
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

// UnpackInterface deserializes an interface value from a map, resolving the concrete type and value.
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

// PackStructObject serializes a struct as a PackedStructObject (slice of field values).
// It skips unexported fields, logs warnings, and registers the schema.
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

// UnpackStructObject deserializes a PackedStructObject into a struct, using the schema to map fields.
// It skips unexported fields, logs warnings, and collects errors.
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

// PackMap serializes a map with string keys, returning a map[any]any.
// It sorts keys for deterministic encoding and collects errors.
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

// UnpackMap deserializes a map[any]any into a map of the target type.
// It collects errors for keys and values.
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

// UnpackPointer deserializes a pointer value, ensuring the result is addressable.
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

// UnpackPrimitive converts a primitive value to the target type using reflection.
func (de *Deserializer) UnpackPrimitive(v any, t reflect.Type) reflect.Value {
	return reflect.ValueOf(v).Convert(t)
}

// warnf logs a warning message to the Serializer’s Warnings slice.
func (ser *Serializer) warnf(warning string) {
	ser.Warnings = append(ser.Warnings, warning)
}

// warnf logs a warning message to the Deserializer’s Warnings slice.
func (de *Deserializer) warnf(warning string) {
	de.Warnings = append(de.Warnings, warning)
}

// backReferenceFromCBORInt converts a CBOR-decoded uint64 to a BackReference.
// It assumes the input is a valid index.
func backReferenceFromCBORInt(n any) BackReference {
	return BackReference(n.(uint64))
}
