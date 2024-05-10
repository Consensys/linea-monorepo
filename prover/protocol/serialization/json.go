package serialization

import (
	"encoding/json"
	"fmt"
	"reflect"
)

// serializeAnyWithJSONPkg serializes an interface{} object into JSON using the
// standard reflection-based [json] package. It will panic on failure
// and is meant to be used on data and types that controlled by the current
// package.
func serializeAnyWithJSONPkg(x any) json.RawMessage {
	res, err := json.Marshal(x)
	if err != nil {
		// that would be unexpected for primitive types
		panic(err)
	}
	return res
}

// deserializeAnyWithJSONPkg calls [json.Unmarshal] and wraps the error if any.
func deserializeAnyWithJSONPkg(data json.RawMessage, x any) error {
	if err := json.Unmarshal(data, x); err != nil {
		return fmt.Errorf("json.Unmarshal failed: %w", err)
	}
	return nil
}

// deserializeValueWithJSON packages attemps to deserialize `data` into the
// [reflect.Value] `v`. It will return an error if the value cannot be accessed
// through the [reflect] package or if it cannot be set.
func deserializeValueWithJSONPkg(data json.RawMessage, v reflect.Value) error {

	if !v.CanAddr() {
		return fmt.Errorf("deserializeValueWithJSONPkg cannot be used for type %v", v.Type())
	}

	if !v.CanInterface() {
		return fmt.Errorf("could not deserialize value of type `%s` because it's an unexported field", v.Type().String())
	}

	// just to ensure that the JSON package will get a pointer. Otherwise, it
	// will not accept to deserialize.
	if v.Kind() != reflect.Pointer && v.CanAddr() {
		v = v.Addr()
	}

	return deserializeAnyWithJSONPkg(data, v.Interface())
}

// serializeValueWithJSONPkg serializes a [reflect.Value] using the [json]
// package. It will return an error if the provided value is an unexported
// field.
func serializeValueWithJSONPkg(v reflect.Value) (json.RawMessage, error) {
	if !v.CanInterface() {
		return nil, fmt.Errorf("could not serialize value of type `%s` because it's an unexported field", v.Type().String())
	}
	return serializeAnyWithJSONPkg(v.Interface()), nil
}

// castAsString returns the string value of a [reflect.String] kind
// [reflect.Value]. It will return an error if the value does not have the right
// kind.
func castAsString(v reflect.Value) (string, error) {
	if v.Kind() != reflect.String {
		return "", fmt.Errorf("expected a string kind value: got %q", v.String())
	}
	return v.String(), nil
}
