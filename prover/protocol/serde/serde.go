package serde

import (
	"fmt"
	"reflect"
	"strings"
	"unsafe"

	"github.com/consensys/linea-monorepo/prover/maths/field"
)

var (
	serdeStructTag         = "serde"
	serdeStructTagOmit     = "omit"
	SerdeStructTagTestOmit = "test_omit"
)

// Serialize transforms a Go object into a binary-packed byte slice.
// The resulting slice is structured with a metadata header at the beginning,
// followed by a linearized heap of data.
func Serialize(v any) ([]byte, error) {
	enc := newEncoder()

	// Reserve space for the FileHeader at the start of the buffer (Offset 0).
	// We will come back and fill this in once we know the final PayloadOffset and DataSize.
	_ = enc.write(FileHeader{})

	// Start the recursive serialization of the object graph.
	rootOff, err := encode(enc, reflect.ValueOf(v))
	if err != nil {
		return nil, err
	}

	// Construct the final metadata. PayloadOff points to the "Root" object.
	finalHeader := FileHeader{
		Magic:       Magic,
		Version:     1,
		PayloadType: 0,
		PayloadOff:  int64(rootOff),
		DataSize:    enc.offset,
	}

	// Finalize the encoded buffer and "Patch" the reserved file header at Offset 0.
	// Using unsafe here allows us to overwrite the header in-place without re-allocating (Zero-Copy).
	// This essentially stores the final FileHeader at the start of our encoded buffer.
	b := enc.buf.Bytes()
	*(*FileHeader)(unsafe.Pointer(&b[0])) = finalHeader
	return b, nil
}

// Deserialize reconstructs a Go object from a byte slice (typically a memory-mapped file).
// IMPORTANT: This is a Zero-Copy operation. The reconstructed object (v) will contain
// pointers, slices, and strings that point directly into the input buffer 'b'.
// The caller MUST ensure that 'b' remains valid and unmodified for the lifetime of 'v'.
func Deserialize(b []byte, v any) error {
	// Sanity check: ensure the buffer is at least as large as our header.
	if len(b) < int(SizeOf[FileHeader]()) {
		return fmt.Errorf("buffer too small to contain header")
	}

	// Map the header from the start of the buffer and verify the Magic identifier.
	header := (*FileHeader)(unsafe.Pointer(&b[0]))
	if header.Magic != Magic {
		return fmt.Errorf("invalid magic bytes: file is corrupted or not a valid archive")
	}

	// Handle the 'Null' payload case (e.g., if a nil pointer was serialized) by
	// setting the target to its zero-value (nil).
	if Ref(header.PayloadOff).IsNull() {
		val := reflect.ValueOf(v)
		if val.Kind() == reflect.Ptr {
			val.Elem().Set(reflect.Zero(val.Elem().Type()))
		}
		return nil
	}

	// Bound check: ensure the root object offset exists within the buffer.
	if Ref(header.PayloadOff) > Ref(len(b)) {
		return fmt.Errorf("root payload offset out of bounds")
	}

	// IMPORTANT: Ensure target 'v' is always a pointer so we can mutate it with the result.
	// This is hard-enforcement (found in any Go std. ser/de libs.) due to Go's pass-by-value semantics.
	// If otherwise, we wouldn't have a "hook" in the Go runtime to attach our reconstructed object graph to.
	// By requiring a pointer, we ensure the caller has provided a valid "landing pad" for the reconstructed data.
	val := reflect.ValueOf(v)
	if val.Kind() != reflect.Ptr {
		return fmt.Errorf("v must be a pointer to a struct or value")
	}

	// Init the decoder with the ptrMap to handle de-duplication of pointers
	// during the reconstruction process (preserving cycles).
	dec := &decoder{
		data:   b,
		ptrMap: make(map[int64]reflect.Value),
	}

	// Begin the "Pointer Swizzling" process starting from the PayloadOff.
	return dec.decode(val.Elem(), int64(header.PayloadOff))
}

// getBinarySize returns the number of bytes a value of type `T` will occupy
// in the serialized buffer according to serde layout rules.
//
// NOTE: This is NOT the same as Go's in-memory size. The size returned here reflects
// how the value is represented on disk:
//
// - Fixed-size, pointer-free values are inlined.
// - Variable-size or heap-backed values are replaced by an 8-byte Ref.
//
// This function MUST remain perfectly consistent with the actual write logic;
// any mismatch will result in corrupted offsets or incorrect deserialization.
func getBinarySize(t reflect.Type) int64 {

	// Indirect types (custom registries, ptrs, slices, strings etc) are types that have variable sizes
	// not known at compile time and hence their inline representation is a Ref (8-byte offset).
	// These values are written elsewhere and referenced inline by offset.
	if isIndirectType(t) {
		return 8 // Size of Ref (8-byte offset)
	}

	k := t.Kind()

	// Explicit handling of scalar types int/uint are platform-dependent in Go,
	// so we normalize them to 8 bytes in the serialized representation.
	if k == reflect.Int || k == reflect.Uint {
		return 8
	}

	// Structs are serialized inline by concatenating the serialized
	// representation of each exported, non-omitted field.
	if k == reflect.Struct {
		// Special-case POD types that are safe to inline as raw bytes.
		if t == reflect.TypeOf(field.Element{}) {
			return int64(t.Size())
		}

		var sum int64
		for i := 0; i < t.NumField(); i++ {
			f := t.Field(i)

			// Skip unexported fields.
			if !f.IsExported() {
				continue
			}

			// Skip fields explicitly omitted from serialization.
			if strings.Contains(f.Tag.Get(serdeStructTag), serdeStructTagOmit) {
				continue
			}

			sum += getBinarySize(f.Type)
		}
		return sum
	}

	// Arrays are fixed-size and serialized inline as repeated elements.
	if k == reflect.Array {
		// field.Element arrays are treated as POD blobs.
		if t == reflect.TypeOf(field.Element{}) {
			return int64(t.Size())
		}

		elemSize := getBinarySize(t.Elem())
		return elemSize * int64(t.Len())
	}

	// Fallback for other fixed-size, pointer-free types
	// (e.g. bool, int/uint8, int/uint16, int/uint32, int/uint64, float32/64).
	return int64(t.Size())
}

// isIndirectType returns true if the given type is an indirect type.
// Indirect types are types that have variable sizes not known at compile time.
// This includes pointers, slices, strings, interfaces, maps, and functions.
// Direct types are types that have fixed sizes known at compile time.
// This includes structs, arrays, and primitive types (bool,int/uint8, int/uint (normalized), etc).
// Types handled inside of the Custom Registry are also considered indirect.
func isIndirectType(t reflect.Type) bool {
	if _, ok := customRegistry[t]; ok {
		return true
	}

	// Indirect types are types that have variable sizes - not known at compile time
	k := t.Kind()
	return k == reflect.Ptr || k == reflect.Slice || k == reflect.String ||
		k == reflect.Interface || k == reflect.Map || k == reflect.Func
}

// normalizeIntegerSize converts platform-dependent types (int, uint) to fixed-size
// equivalents (int64, uint64) - 64 bit values. This ensures that the binary representation
// is consistent across different CPU architectures (32-bit vs 64-bit).
func normalizeIntegerSize(v reflect.Value) any {
	switch v.Kind() {
	case reflect.Int:
		return int64(v.Int())
	case reflect.Uint:
		return uint64(v.Uint())
	default:
		// For all other types (int64, float64, structs, etc.),
		// return the interface as-is.
		return v.Interface()
	}
}
