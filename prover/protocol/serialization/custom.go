package serialization

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"math/big"
	"reflect"

	"github.com/consensys/linea-monorepo/prover/maths/field"
	"github.com/consensys/linea-monorepo/prover/symbolic"
	"github.com/fxamacker/cbor/v2"
	"github.com/pierrec/lz4/v4"
)

const (
	variableExpressionOpType uint64 = iota
	constantExpressionOpType
	linearCombinationExpressionOpType
	productExpressionOpType
	polyEvalExpressionOpType
)

// CustomCodex represents an optional behavior for a specific type
type CustomCodex struct {
	Type reflect.Type
	Ser  func(ser *Serializer, val reflect.Value) (any, error)
	Des  func(des *Deserializer, val any, t reflect.Type) (reflect.Value, error)
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

	CustomCodexes[TypeOfExpression] = CustomCodex{
		Type: TypeOfExpression,
		Ser:  marshalExpression,
		Des:  unmarshalExpression,
	}

	CustomCodexes[TypeOfArrOfFieldElement] = CustomCodex{
		Type: TypeOfArrOfFieldElement,
		Ser:  marshalArrayOfFieldElement,
		Des:  unmarshalArrayOfFieldElement,
	}
}

func marshalFieldElement(_ *Serializer, val reflect.Value) (any, error) {
	f := val.Interface().(field.Element)
	bi := fieldToSmallBigInt(f)
	f.BigInt(bi)
	return marshalBigInt(nil, reflect.ValueOf(bi))
}

func unmarshalFieldElement(_ *Deserializer, val any, _ reflect.Type) (reflect.Value, error) {

	f, err := unmarshalBigInt(nil, val, TypeOfBigInt)
	if err != nil {
		return reflect.Value{}, err
	}
	var fe field.Element
	fe.SetBigInt(f.Interface().(*big.Int))
	return reflect.ValueOf(fe), nil
}

func marshalBigInt(_ *Serializer, val reflect.Value) (any, error) {
	return val.Interface().(*big.Int), nil
}

func unmarshalBigInt(_ *Deserializer, val any, _ reflect.Type) (reflect.Value, error) {
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

func marshalExpression(ser *Serializer, val reflect.Value) (any, error) {

	e := val.Interface().(*symbolic.Expression)

	// This possible when the expression is nil but its pointer is an initialized
	// value. This can happen for an non-used pointer field in a struct. In that
	// case, we just return nil. This won't be caught by the first condition of
	// PackValue because the pointer itself does exists, it's the pointed value
	// that does not.
	if e == nil {
		return nil, nil
	}

	switch op := e.Operator.(type) {

	case symbolic.Variable:

		metaValue := reflect.ValueOf(op.Metadata).Convert(TypeOfVariableMetadata)
		m, em := ser.PackInterface(metaValue)
		if em != nil {
			return nil, fmt.Errorf("could not marshal expression variable metadata: %w", em)
		}

		return map[any]any{
			"t": variableExpressionOpType,
			"m": m,
		}, nil

	case symbolic.Constant:
		v, ev := marshalFieldElement(ser, reflect.ValueOf(op.Val))
		if ev != nil {
			return nil, fmt.Errorf("could not marshal expression constant value: %w", ev)
		}

		return map[any]any{
			"t": constantExpressionOpType,
			"v": v,
		}, nil

	case symbolic.LinComb:
		var (
			c, ec = ser.PackArrayOrSlice(reflect.ValueOf(op.Coeffs))
			v, ev = ser.PackArrayOrSlice(reflect.ValueOf(e.Children))
		)

		if ec != nil || ev != nil {
			return nil, fmt.Errorf("counld not marshal linear-combination expression: %w", errors.Join(ec, ev))
		}

		return map[any]any{
			"t": linearCombinationExpressionOpType,
			"c": c,
			"v": v,
		}, nil

	case symbolic.Product:
		var (
			x, ex = ser.PackArrayOrSlice(reflect.ValueOf(op.Exponents))
			v, ev = ser.PackArrayOrSlice(reflect.ValueOf(e.Children))
		)

		if ex != nil || ev != nil {
			return nil, fmt.Errorf("counld not marshal product expression: %w", errors.Join(ex, ev))
		}

		return map[any]any{
			"t": productExpressionOpType,
			"c": x,
			"v": v,
		}, nil

	case symbolic.PolyEval:
		v, ev := ser.PackArrayOrSlice(reflect.ValueOf(e.Children))
		if ev != nil {
			return nil, fmt.Errorf("counld not marshal polynomial evaluation expression: %w", ev)
		}

		return map[any]any{
			"t": polyEvalExpressionOpType,
			"v": v,
		}, nil

	default:
		return nil, fmt.Errorf("unsupported expression operator: %T", e.Operator)
	}
}

func unmarshalExpression(des *Deserializer, val any, _ reflect.Type) (reflect.Value, error) {

	eMap := val.(map[any]any)
	opType := eMap["t"].(uint64)

	switch opType {

	case variableExpressionOpType:
		mPacked := eMap["m"].(map[any]any)
		m, em := des.UnpackInterface(mPacked, TypeOfVariableMetadata)
		if em != nil {
			return reflect.Value{}, fmt.Errorf("could not unmarshal expression variable metadata: %w", em)
		}

		new := symbolic.NewVariable(m.Interface().(symbolic.Metadata))
		return reflect.ValueOf(new), nil

	case constantExpressionOpType:
		v, ev := unmarshalFieldElement(des, eMap["v"], nil)
		if ev != nil {
			return reflect.Value{}, fmt.Errorf("could not unmarshal expression constant value: %w", ev)
		}

		f := v.Interface().(field.Element)
		new := symbolic.NewConstant(f)
		return reflect.ValueOf(new), nil

	case linearCombinationExpressionOpType, productExpressionOpType:
		c, ec := des.UnpackArrayOrSlice(eMap["c"].([]any), TypeOfArrayOfInt)
		if ec != nil {
			return reflect.Value{}, fmt.Errorf("could not unmarshal linear-combination expression coeffs: %w", ec)
		}

		v, ev := des.UnpackArrayOrSlice(eMap["v"].([]any), TypeOfArrayOfExpr)
		if ev != nil {
			return reflect.Value{}, fmt.Errorf("could not unmarshal linear-combination expression children: %w", ev)
		}

		var e *symbolic.Expression

		if opType == linearCombinationExpressionOpType {
			e = symbolic.NewLinComb(
				v.Interface().([]*symbolic.Expression),
				c.Interface().([]int),
			)
		}

		if opType == productExpressionOpType {
			e = symbolic.NewProduct(
				v.Interface().([]*symbolic.Expression),
				c.Interface().([]int),
			)
		}

		return reflect.ValueOf(e), nil

	case polyEvalExpressionOpType:

		v, ev := des.UnpackArrayOrSlice(eMap["v"].([]any), TypeOfArrayOfExpr)
		if ev != nil {
			return reflect.Value{}, fmt.Errorf("could not unmarshal polynomial evaluation expression children: %w", ev)
		}

		args := v.Interface().([]*symbolic.Expression)
		new := symbolic.NewPolyEval(args[0], args[1:])
		return reflect.ValueOf(new), nil

	default:
		return reflect.Value{}, fmt.Errorf("unsupported expression operator: %v", opType)

	}
}

func marshalArrayOfFieldElement(_ *Serializer, val reflect.Value) (any, error) {

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

func unmarshalArrayOfFieldElement(_ *Deserializer, val any, _ reflect.Type) (reflect.Value, error) {

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

// This converts the field.Element to a smaller big.Int. This is done to
// reduce the size of the CBOR encoding. The backward conversion is automatically
// done [field.SetBigInt] as it handles negative values.
func fieldToSmallBigInt(v field.Element) *big.Int {

	neg := new(field.Element).Neg(&v)
	if neg.IsUint64() {
		n := neg.Uint64()
		return new(big.Int).SetInt64(int64(n))
	}

	bi := &big.Int{}
	v.BigInt(bi)
	return bi
}
