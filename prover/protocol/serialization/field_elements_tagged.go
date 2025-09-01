package serialization

import (
	"bytes"

	"reflect"

	"github.com/consensys/gnark-crypto/utils/unsafe"
	"github.com/consensys/linea-monorepo/prover/maths/common/smartvectors"
	"github.com/consensys/linea-monorepo/prover/maths/field"
	"github.com/fxamacker/cbor/v2"
)

// Use a vendor-specific tag number for homogeneous field elements packed as bytes.
// RFC 8746 reserves ranges for typed arrays but doesn't define this 377-field explicitly.
// Pick a private-use tag in the high range to avoid collisions.
const cborTagFieldElementsPacked uint64 = 60001

// marshalArrayOfFieldElement: keep zero-copy-ish packing, add CBOR tag wrapper.
func marshalArrayOfFieldElement(_ *Serializer, val reflect.Value) (any, *serdeError) {
	var buf = &bytes.Buffer{}

	v, ok := val.Interface().([]field.Element)
	if !ok {
		v = []field.Element(val.Interface().(smartvectors.Regular))
	}
	if err := unsafe.WriteSlice(buf, v); err != nil {
		return nil, newSerdeErrorf("could not marshal array of field element: %w", err)
	}

	// Wrap in cbor.Tag so decoders know this is a homogeneous packed vector.
	return cbor.Tag{
		Number:  cborTagFieldElementsPacked,
		Content: buf.Bytes(),
	}, nil
}

// unmarshalArrayOfFieldElement: accept either tagged content or raw []byte for backward compatibility.
func unmarshalArrayOfFieldElement(_ *Deserializer, val any, t reflect.Type) (reflect.Value, *serdeError) {
	var raw []byte

	switch x := val.(type) {
	case cbor.Tag:
		// Accept our tag and extract the bytes content.
		if x.Number != cborTagFieldElementsPacked {
			return reflect.Value{}, newSerdeErrorf("unexpected CBOR tag for field elements: %d", x.Number)
		}
		b, ok := x.Content.([]byte)
		if !ok {
			return reflect.Value{}, newSerdeErrorf("tagged field elements not []byte content, got %T", x.Content)
		}
		raw = b
	case []byte:
		// Older payloads encoded raw []byte without a tag.
		raw = x
	default:
		return reflect.Value{}, newSerdeErrorf("invalid type for field elements: %T", val)
	}

	r := bytes.NewReader(raw)
	v, _, err := unsafe.ReadSlice[[]field.Element](r)
	if err != nil {
		return reflect.Value{}, newSerdeErrorf("could not unmarshal array of field element: %w", err)
	}

	if t == reflect.TypeOf(smartvectors.Regular{}) {
		return reflect.ValueOf(smartvectors.Regular(v)), nil
	}
	return reflect.ValueOf(v), nil
}
