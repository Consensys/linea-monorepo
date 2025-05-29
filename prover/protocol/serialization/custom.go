package serialization

import (
	"errors"
	"fmt"
	"math/big"
	"reflect"

	"github.com/consensys/linea-monorepo/prover/maths/field"
	"github.com/consensys/linea-monorepo/prover/symbolic"
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

var CustomCodexes = []CustomCodex{
	{
		Type: TypeOfBigInt,
		Ser:  marshalBigInt,
		Des:  unmarshalBigInt,
	},
	{
		Type: TypeOfFieldElement,
		Ser:  marshalFieldElement,
		Des:  unmarshalFieldElement,
	},
	{
		Type: TypeOfExpression,
		Ser:  marshalExpression,
		Des:  unmarshalExpression,
	},
}

func marshalFieldElement(_ *Serializer, val reflect.Value) (any, error) {
	f := val.Interface().(field.Element)
	res := &big.Int{}
	f.BigInt(res)
	return f, nil
}

func unmarshalFieldElement(_ *Deserializer, val any, _ reflect.Type) (reflect.Value, error) {
	f := val.(*big.Int)
	var fe field.Element
	fe.SetBigInt(f)
	return reflect.ValueOf(f), nil
}

func marshalBigInt(_ *Serializer, val reflect.Value) (any, error) {
	return val.Interface().(*big.Int), nil
}

func unmarshalBigInt(_ *Deserializer, val any, _ reflect.Type) (reflect.Value, error) {
	return reflect.ValueOf(val.(*big.Int)), nil
}

func marshalExpression(ser *Serializer, val reflect.Value) (any, error) {

	e := val.Interface().(*symbolic.Expression)

	switch op := e.Operator.(type) {

	case symbolic.Variable:
		m, em := ser.PackArrayOrSlice(reflect.ValueOf(op.Metadata))
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
