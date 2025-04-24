package serialization

import (
	"encoding/json"
	"fmt"
	"reflect"
	"sync"

	"github.com/ugorji/go/codec"
)

// Size limits matching original configuration.
const (
	MaxArrayElements = 1 << 27 // 134217728
	MaxMapPairs      = 1 << 27 // 134217728
	defaultBufferCap = 256     // Preallocated buffer size
)

var (
	// Shared pool for encoders/decoders
	encoderPool = sync.Pool{
		New: func() interface{} {
			return codec.NewEncoderBytes(nil, newOptimizedHandle())
		},
	}

	decoderPool = sync.Pool{
		New: func() interface{} {
			return codec.NewDecoderBytes(nil, newOptimizedHandle())
		},
	}
)

// newOptimizedHandle creates a CBOR handle with performance-focused settings
func newOptimizedHandle() *codec.CborHandle {
	// Configure CborHandle with available options
	handle := &codec.CborHandle{}

	// Ensure deterministic encoding
	handle.Canonical = true

	// Disable Convert raw bytes to strings
	handle.RawToString = false
	handle.MapType = reflect.TypeOf(map[string]interface{}(nil))

	handle.SkipUnexpectedTags = true
	return handle
}

// validateSize optimized with type switches
func validateSize(v any) error {
	switch tv := v.(type) {
	case interface{ Len() int }:
		if tv.Len() > MaxArrayElements {
			return fmt.Errorf("size %d exceeds limit %d", tv.Len(), MaxArrayElements)
		}
	case map[string]interface{}:
		if len(tv) > MaxMapPairs {
			return fmt.Errorf("map size %d exceeds limit %d", len(tv), MaxMapPairs)
		}
	}
	return nil
}

// serializeAnyWithCborPkg with pooling and buffer preallocation
func serializeAnyWithCborPkg(v any) (json.RawMessage, error) {
	if s, ok := v.(Serializable); ok {
		return s.Serialize()
	}

	if err := validateSize(v); err != nil {
		return nil, fmt.Errorf("size validation failed for %T: %w", v, err)
	}

	// Get encoder from pool
	enc := encoderPool.Get().(*codec.Encoder)
	defer encoderPool.Put(enc)

	// Preallocate buffer
	buf := make([]byte, 0, defaultBufferCap)
	enc.ResetBytes(&buf)

	if err := enc.Encode(v); err != nil {
		return nil, fmt.Errorf("encode failed for %T: %w", v, err)
	}
	return buf, nil
}

// deserializeAnyWithCborPkg with pooled decoder
func deserializeAnyWithCborPkg(data []byte, v any) error {
	if s, ok := v.(Serializable); ok {
		return s.Deserialize(data)
	}

	// Get decoder from pool
	dec := decoderPool.Get().(*codec.Decoder)
	defer decoderPool.Put(dec)

	dec.ResetBytes(data)
	return dec.Decode(v)
}
