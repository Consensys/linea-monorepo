package serialization

import (
	"encoding/json"
	"fmt"
	"reflect"
	"sync"

	"github.com/ugorji/go/codec"
)

// Serializable defines an interface for types with custom serialization.
type Serializable interface {
	Serialize() (json.RawMessage, error)
	Deserialize(data json.RawMessage) error
}

// Size limits matching original configuration.
const (
	MaxArrayElements = 134217728 // 1 << 27
	MaxMapPairs      = 134217728 // 1 << 27
)

// typeCache caches CBOR encoding/decoding metadata for types.
var (
	typeCacheMu sync.RWMutex
	typeCache   = make(map[reflect.Type]*codec.CborHandle)
)

// getCborHandle returns a cached CBOR handle for a type.
func getCborHandle(t reflect.Type) *codec.CborHandle {
	typeCacheMu.RLock()
	if handle, ok := typeCache[t]; ok {
		typeCacheMu.RUnlock()
		return handle
	}
	typeCacheMu.RUnlock()

	// Configure CborHandle with available options
	handle := &codec.CborHandle{}
	handle.Canonical = true   // Ensure deterministic encoding
	handle.RawToString = true // Convert raw bytes to strings
	handle.MapType = reflect.TypeOf(map[string]interface{}(nil))

	typeCacheMu.Lock()
	typeCache[t] = handle
	typeCacheMu.Unlock()
	return handle
}

// validateSize checks if arrays or maps exceed size limits.
func validateSize(v any) error {
	// Handle slices/arrays
	if reflect.TypeOf(v).Kind() == reflect.Slice || reflect.TypeOf(v).Kind() == reflect.Array {
		if reflect.ValueOf(v).Len() > MaxArrayElements {
			return fmt.Errorf("array size %d exceeds limit %d", reflect.ValueOf(v).Len(), MaxArrayElements)
		}
	}

	// Handle maps
	if reflect.TypeOf(v).Kind() == reflect.Map {
		if reflect.ValueOf(v).Len() > MaxMapPairs {
			return fmt.Errorf("map size %d exceeds limit %d", reflect.ValueOf(v).Len(), MaxMapPairs)
		}
	}

	// Optionally recurse into nested structures (simplified for performance)
	return nil
}

// serializeAnyWithCborPkg serializes a value to CBOR, preferring Serializable interface.
func serializeAnyWithCborPkg(v any) (json.RawMessage, error) {
	if s, ok := v.(Serializable); ok {
		return s.Serialize()
	}

	// Validate size before encoding
	if err := validateSize(v); err != nil {
		return nil, fmt.Errorf("size validation failed for type %T: %w", v, err)
	}

	var b []byte
	enc := codec.NewEncoderBytes(&b, getCborHandle(reflect.TypeOf(v)))
	if err := enc.Encode(v); err != nil {
		return nil, fmt.Errorf("failed to encode value of type %T: %w", v, err)
	}
	return b, nil
}

// deserializeAnyWithCborPkg deserializes CBOR into a value, preferring Serializable interface.
func deserializeAnyWithCborPkg(data []byte, v any) error {
	if s, ok := v.(Serializable); ok {
		return s.Deserialize(data)
	}

	// Decode into a temporary interface{} to validate size
	var temp interface{}
	dec := codec.NewDecoderBytes(data, getCborHandle(reflect.TypeOf(v)))
	if err := dec.Decode(&temp); err != nil {
		return fmt.Errorf("failed to decode into temporary type: %w", err)
	}

	// Validate size of decoded data
	if err := validateSize(temp); err != nil {
		return err
	}

	// Decode into target value
	dec = codec.NewDecoderBytes(data, getCborHandle(reflect.TypeOf(v)))
	if err := dec.Decode(v); err != nil {
		return fmt.Errorf("failed to decode into type %T: %w", v, err)
	}
	return nil
}
