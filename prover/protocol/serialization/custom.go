package serialization

import (
	"bytes"
	"fmt"
	"hash"
	"io"
	"math"
	"math/big"
	"reflect"
	"strings"
	"sync"

	"github.com/consensys/gnark/frontend"
	"github.com/consensys/linea-monorepo/prover/crypto/mimc"
	"github.com/consensys/linea-monorepo/prover/crypto/ringsis"
	"github.com/consensys/linea-monorepo/prover/crypto/state-management/hashtypes"
	"github.com/consensys/linea-monorepo/prover/maths/field"
	"github.com/consensys/linea-monorepo/prover/zkevm/arithmetization"
	"github.com/fxamacker/cbor/v2"
	"github.com/pierrec/lz4/v4"
)

// CustomCodex represents an optional behavior for a specific type
type CustomCodex struct {
	Type reflect.Type
	Ser  func(path string, ser *Serializer, val reflect.Value) (any, error)
	Des  func(path string, des *Deserializer, val any, t reflect.Type) (reflect.Value, error)
}

var CustomCodexes = map[reflect.Type]CustomCodex{}

func init() {

	CustomCodexes[TypeOfBigInt] = CustomCodex{
		Type: TypeOfBigInt,
		Ser:  marshalBigInt,
		Des:  unmarshalBigInt,
	}

	CustomCodexes[TypeOfFieldElement] = CustomCodex{
		Type: TypeOfFieldElement,
		Ser:  marshalFieldElement,
		Des:  unmarshalFieldElement,
	}

	CustomCodexes[TypeOfArrOfFieldElement] = CustomCodex{
		Type: TypeOfArrOfFieldElement,
		Ser:  marshalArrayOfFieldElement,
		Des:  unmarshalArrayOfFieldElement,
	}

	CustomCodexes[TypeOfArithmetization] = CustomCodex{
		Type: TypeOfArithmetization,
		Ser:  marshalArithmetization,
		Des:  unmarshalArithmetization,
	}

	CustomCodexes[TypeOfFrontendVariable] = CustomCodex{
		Type: TypeOfFrontendVariable,
		Ser:  marshalFrontendVariable,
		Des:  unmarshalFrontendVariable,
	}

	CustomCodexes[TypeOfHashFuncGenerator] = CustomCodex{
		Type: TypeOfHashFuncGenerator,
		Ser:  marshalHashGenerator,
		Des:  unmarshalHashGenerator,
	}

	CustomCodexes[TypeOfHashTypeHasher] = CustomCodex{
		Type: TypeOfHashTypeHasher,
		Ser:  marshalHashTypeHasher,
		Des:  unmarshalHashTypeHasher,
	}

	CustomCodexes[reflect.TypeOf(sync.Mutex{})] = CustomCodex{
		Type: reflect.TypeOf(sync.Mutex{}),
		Ser:  marshalAsNil,
		Des:  unmarshalAsZero,
	}

	CustomCodexes[TypeOfRingSisKeyPtr] = CustomCodex{
		Type: TypeOfRingSisKeyPtr,
		Ser:  marshalRingSisKey,
		Des:  unmarshalRingSisKey,
	}
}

func marshalRingSisKey(path string, ser *Serializer, val reflect.Value) (any, error) {
	key := val.Interface().(*ringsis.Key)
	keyGenParams := key.KeyGen
	res, err := ser.PackStructObject(path, reflect.ValueOf(*keyGenParams))
	if err != nil {
		return nil, fmt.Errorf("could not marshal ring-sis key: %w", err)
	}
	return res, nil
}

func unmarshalRingSisKey(path string, des *Deserializer, val any, _ reflect.Type) (reflect.Value, error) {

	if v_, ok := val.(PackedStructObject); ok {
		val = []any(v_)
	}

	res, err := des.UnpackStructObject(path, val.([]any), TypeofRingSisKeyGenParam)
	if err != nil {
		return reflect.Value{}, fmt.Errorf("could not unpack struct object for ring-sis key: %w", err)
	}

	keyGenParams, ok := res.Interface().(ringsis.KeyGen)
	if !ok {
		return reflect.Value{}, fmt.Errorf("could not cast to ringsis.KeyGen: %w", err)
	}
	innerParams := keyGenParams.Params
	ringSiskey := ringsis.GenerateKey(*innerParams, keyGenParams.MaxNumFieldToHash)
	return reflect.ValueOf(ringSiskey), nil
}

func marshalFieldElement(path string, _ *Serializer, val reflect.Value) (any, error) {
	f := val.Interface().(field.Element)
	bi := fieldToSmallBigInt(f)
	f.BigInt(bi)
	return marshalBigInt(path, nil, reflect.ValueOf(bi))
}

func unmarshalFieldElement(path string, _ *Deserializer, val any, _ reflect.Type) (reflect.Value, error) {
	f, err := unmarshalBigInt(path, nil, val, TypeOfBigInt)
	if err != nil {
		return reflect.Value{}, err
	}

	var fe field.Element
	fe.SetBigInt(f.Interface().(*big.Int))
	return reflect.ValueOf(fe), nil
}

func marshalBigInt(path string, _ *Serializer, val reflect.Value) (any, error) {
	return val.Interface().(*big.Int), nil
}

func unmarshalBigInt(path string, _ *Deserializer, val any, _ reflect.Type) (reflect.Value, error) {
	switch v := val.(type) {
	case big.Int:
		return reflect.ValueOf(&v), nil
	case int64:
		return reflect.ValueOf(big.NewInt(v)), nil
	case uint64:
		return reflect.ValueOf(new(big.Int).SetUint64(v)), nil
	default:
		return reflect.Value{}, fmt.Errorf("invalid type: %T, value: %++v", val, val)
	}
}

func marshalArrayOfFieldElement(path string, _ *Serializer, val reflect.Value) (any, error) {

	var (
		v           = val.Interface().([]field.Element)
		buffer      = &bytes.Buffer{}
		lz4Encoder  = lz4.NewWriter(buffer)
		cborEncoder = cbor.NewEncoder(lz4Encoder)
	)

	// The header of the array is its size
	if err := cborEncoder.Encode(uint64(len(v))); err != nil {
		return nil, err
	}

	for i := 0; i < len(v); i++ {
		bi := fieldToSmallBigInt(v[i])
		if err := cborEncoder.Encode(bi); err != nil {
			return nil, err
		}
	}

	if err := lz4Encoder.Close(); err != nil {
		return nil, err
	}

	return buffer.Bytes(), nil
}

func unmarshalArrayOfFieldElement(path string, _ *Deserializer, val any, _ reflect.Type) (reflect.Value, error) {

	var (
		buffer      = bytes.NewReader(val.([]byte))
		lz4Decoder  = lz4.NewReader(buffer)
		cborDecoder = cbor.NewDecoder(lz4Decoder)
	)

	// The header of the encoded array is its size as a uint64
	size := uint64(0)
	if err := cborDecoder.Decode(&size); err != nil {
		return reflect.Value{}, err
	}

	v := make([]field.Element, 0, size)
	for {
		var bi big.Int
		if err := cborDecoder.Decode(&bi); err != nil {
			if err == io.EOF {
				break
			}
			return reflect.Value{}, err
		}

		f := field.Element{}
		f.SetBigInt(&bi)
		v = append(v, f)
	}

	return reflect.ValueOf(v), nil
}

func marshalArithmetization(path string, ser *Serializer, val reflect.Value) (any, error) {

	res, err := ser.PackStructObject(path, val)
	if err != nil {
		return nil, fmt.Errorf("could not marshal arithmetization: %w", err)
	}

	return res, nil
}

func unmarshalArithmetization(path string, des *Deserializer, val any, _ reflect.Type) (reflect.Value, error) {
	if v_, ok := val.(PackedStructObject); ok {
		val = []any(v_)
	}

	res, err := des.UnpackStructObject(path, val.([]any), TypeOfArithmetization)
	if err != nil {
		return reflect.Value{}, fmt.Errorf("could not unmarshal arithmetization: %w", err)
	}

	arith := res.Interface().(arithmetization.Arithmetization)
	schema, meta, err := arithmetization.UnmarshalZkEVMBin(arith.ZkEVMBin, arith.Settings.OptimisationLevel)
	if err != nil {
		return reflect.Value{}, fmt.Errorf("could not unmarshal arithmetization: %w", err)
	}
	arith.Schema = schema
	arith.Metadata = meta
	return reflect.ValueOf(arith), nil
}

func marshalFrontendVariable(path string, ser *Serializer, val reflect.Value) (any, error) {

	var (
		variable = val.Interface().(frontend.Variable)
		bi       = &big.Int{}
	)

	switch v := variable.(type) {
	case int:
		bi.SetInt64(int64(v))
	case uint:
		bi.SetUint64(uint64(v))
	case int64:
		bi.SetInt64(int64(v))
	case uint64:
		bi.SetUint64(v)
	case int32:
		bi.SetInt64(int64(v))
	case uint32:
		bi.SetUint64(uint64(v))
	case int16:
		bi.SetInt64(int64(v))
	case uint16:
		bi.SetUint64(uint64(v))
	case int8:
		bi.SetInt64(int64(v))
	case uint8:
		bi.SetUint64(uint64(v))
	case field.Element:
		bi = fieldToSmallBigInt(v)
	case big.Int:
		*bi = v
	case *big.Int:
		bi = v
	case string:
		bi.SetString(v, 0)
	default:
		if variable == nil {
			return nil, nil
		}

		// The check cannot be done on val.Type as it would return
		// [frontend.Variable]. The check is somewhat fragile as it rely on
		// the type name and the package name. We return nil in that case, be
		// -cause it signifies that the variable belongs to a circuit that has
		// been compiled by gnark. That information is not relevant to
		// serialize.
		if strings.Contains(reflect.TypeOf(variable).String(), "expr.Term") {
			return nil, nil
		}

		return nil, fmt.Errorf("invalid type for a frontend variable: %T, value: %++v, type.string=%v, path=%v", variable, variable, reflect.TypeOf(variable).String(), path)
	}

	res, err := marshalBigInt(path, ser, reflect.ValueOf(bi))
	if err != nil {
		return nil, fmt.Errorf("could not marshal frontend variable: %w, path=%v", err, path)
	}

	return res, nil
}

func unmarshalFrontendVariable(path string, des *Deserializer, val any, _ reflect.Type) (reflect.Value, error) {

	bi, err := unmarshalBigInt(path, des, val, TypeOfBigInt)
	if err != nil {
		return reflect.Value{}, fmt.Errorf("could not unmarshal frontend variable: %w", err)
	}

	v := reflect.New(TypeOfFrontendVariable).Elem()
	v.Set(reflect.ValueOf(bi))
	return v, nil
}

func marshalHashGenerator(path string, ser *Serializer, val reflect.Value) (any, error) {
	return nil, nil
}

func unmarshalHashGenerator(path string, des *Deserializer, val any, _ reflect.Type) (reflect.Value, error) {
	f := func() hash.Hash {
		return mimc.NewMiMC()
	}
	return reflect.ValueOf(f), nil
}

func marshalHashTypeHasher(path string, ser *Serializer, val reflect.Value) (any, error) {
	return nil, nil
}

func unmarshalHashTypeHasher(path string, des *Deserializer, val any, _ reflect.Type) (reflect.Value, error) {
	return reflect.ValueOf(hashtypes.MiMC), nil
}

// marshalAsNil is a custom serialization function that marshals the given value to nil.
// It is used for types that are not meant to be serialized, such as functions.
func marshalAsNil(_ string, _ *Serializer, _ reflect.Value) (any, error) {
	return nil, nil
}

// unmarshalAsZero is a custom deserialization function that unmarshals the
// given value to zero. It is meant for the type that are not intended to be
// serialized.
func unmarshalAsZero(path string, des *Deserializer, val any, t reflect.Type) (reflect.Value, error) {
	return reflect.Zero(t), nil
}

// This converts the field.Element to a smaller big.Int. This is done to
// reduce the size of the CBOR encoding. The backward conversion is automatically
// done [field.SetBigInt] as it handles negative values.
func fieldToSmallBigInt(v field.Element) *big.Int {
	neg := new(field.Element).Neg(&v)
	if neg.IsUint64() {
		n := neg.Uint64()
		unsafe := n > math.MaxInt64
		if !unsafe {
			return new(big.Int).SetInt64(-int64(n))
		}
	}

	bi := &big.Int{}
	v.BigInt(bi)
	return bi
}
