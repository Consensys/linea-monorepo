package jsonutil

import (
	"fmt"
	"strconv"

	"github.com/consensys/accelerated-crypto-monorepo/backend/statemanager/eth"
	"github.com/consensys/accelerated-crypto-monorepo/utils"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/valyala/fastjson"
)

// Convenient type aliases
type (
	EthAddress = eth.Address
	Bytes32    = eth.FullBytes32
)

func TryGetInt(v fastjson.Value, fields ...string) (int, error) {
	got := v.Get(fields...)
	if got == nil {
		return 0, fmt.Errorf("%v not found (%v)", fields, v.String())
	}

	res, err := got.Int()
	if err != nil {
		err = fmt.Errorf("error : %v, json : %v, path : %v", err, v.String(), fields)
	}
	return res, err
}

// Parse an integer in hexadecimal format (e.g "0xab" -> int(171))
func TryGetHexInt(v fastjson.Value, fields ...string) (int, error) {
	got := v.Get(fields...)
	if got == nil {
		return 0, fmt.Errorf("%v not found (%v)", fields, v.String())
	}

	parsedStr, err := TryGetString(*got)
	if err != nil {
		return 0, fmt.Errorf("error : %v, json : %v, path : %v", err, v.String(), fields)
	}

	// Clear the 0x prefix if any
	if len(parsedStr) >= 2 && parsedStr[0:2] == "0x" {
		parsedStr = parsedStr[2:]
	}

	// Parse th
	parsedInt, err := strconv.ParseInt(parsedStr, 16, 64)
	if err != nil {
		return 0, fmt.Errorf("error : %v, json : %v, path : %v", err, v.String(), fields)
	}

	return int(parsedInt), err
}

func TryGetInt64(v fastjson.Value, field ...string) (int64, error) {
	got := v.Get(field...)
	if got == nil {
		return 0, fmt.Errorf("%v not found (%v)", field, v.String())
	}
	return got.Int64()
}

func TryGetString(v fastjson.Value, fields ...string) (string, error) {
	got := v.Get(fields...)

	if got == nil {
		return "", fmt.Errorf("%v not found (%v)", fields, v.String())
	}

	// type-check just to be sure we are manipulating a string
	if got.Type() != fastjson.TypeString {
		return "", fmt.Errorf("%v not a string", fields)
	}

	res := got.GetStringBytes()
	if res == nil {
		return "", fmt.Errorf("%v not a valid string", fields)
	}

	return string(res), nil
}

func TryGetArray(v fastjson.Value, field ...string) ([]fastjson.Value, error) {

	// Test the type of the underlying array
	if v.Get(field...).Type() != fastjson.TypeArray {
		return nil, fmt.Errorf("path %v for %v : expected an array", field, v.String())
	}

	array := v.GetArray(field...)

	// It can be an empty array
	if array == nil {
		return []fastjson.Value{}, nil
	}

	// check for nil entries
	for i := range array {
		if array[i] == nil {
			return nil, fmt.Errorf("%v[%v]", field, i)
		}
	}

	// it's clean, we can dereference and return
	res := make([]fastjson.Value, len(array))
	for i := range array {
		res[i] = *array[i]
	}

	return res, nil
}

func TryGetHexBytes(v fastjson.Value, field ...string) ([]byte, error) {
	s, err := TryGetString(v, field...)
	if err != nil {
		return nil, err
	}

	// Then attempt to parse the string as a byte32
	b, err := hexutil.Decode(s)
	if err != nil {
		return nil, fmt.Errorf("invalid hex string (%v): %v", s, err)
	}

	return b, nil
}

func TryGetBytes32(v fastjson.Value, field ...string) (Bytes32, error) {
	// Parse an hex string byte slice first
	b, err := TryGetHexBytes(v, field...)
	if err != nil {
		return Bytes32{}, err
	}

	// Check the length of the bytes, which should equal 32
	if len(b) != 32 {
		return Bytes32{}, fmt.Errorf("invalid length. should be 32. was %v (%v = 0x%x)", len(b), field, b)
	}

	res := Bytes32{}
	copy(res[:], b)
	return res, nil
}

func TryGetDigest(v fastjson.Value, field ...string) (eth.Digest, error) {
	res, err := TryGetBytes32(v, field...)
	return eth.Digest(res), err
}

func TryGetAddress(v fastjson.Value, field string) (EthAddress, error) {
	// Parse an hex string byte slice first
	b, err := TryGetHexBytes(v, field)
	if err != nil {
		return EthAddress{}, err
	}

	// Check the length of the bytes, which should equal 20. In practice they can sometime
	// hold over
	if len(b) != 20 {
		return EthAddress{}, fmt.Errorf("invalid length. should be 20. was %v (%v=0x%x)", len(b), field, b)
	}

	res := EthAddress{}
	copy(res[:], b)
	return res, nil
}

// Bytes32IsSet returns true if the given json value is set and
// v corresponds to a Bytes32. Returns an error, if the value is
// set but not a Bytes32.
func Bytes32IsSet(v *fastjson.Value) bool {

	// The value is missing
	if v == nil {
		return false
	}

	// Test the type
	switch v.Type() {
	case fastjson.TypeNull:
		return false
	case fastjson.TypeString:
		// NO-OP
	default:
		utils.Panic("unexpected type %v", v.Type())
	}

	// Now, we should be able to parse it as a string
	s, err := TryGetString(*v)
	if err != nil {
		panic("unreachable")
	}

	// Case where we got "0x" or "". In this case, we say
	// the field is not set.
	if len(s) == 0 || s == "0x" {
		return false
	}

	// Try to parse the string as a byte32
	_, err = TryGetBytes32(*v)
	if err != nil {
		utils.Panic("The string is not a valid bytes32 : %v", err)
	}

	return true
}

// Bytes32IsSet returns true if the given json value is set and
// v corresponds to a Bytes32. Returns an error, if the value is
// set but not a Bytes32.
func ArrayIsSet(v *fastjson.Value) bool {

	// The value is missing
	if v == nil {
		return false
	}

	// Test the type
	switch v.Type() {
	case fastjson.TypeNull:
		return false
	case fastjson.TypeArray:
		// NO-OP
	default:
		utils.Panic("unexpected type %v", v.Type())
	}

	// Now, we should be able to parse it as a string
	s, err := TryGetArray(*v)
	if err != nil {
		panic("unreachable")
	}

	// Passed an empty array. Equivalent to a missing field.
	if len(s) == 0 {
		return false
	}

	return true
}
