package serialization

import (
	"bytes"
	"math/big"
	"reflect"

	"github.com/consensys/gnark-crypto/utils/unsafe"
	"github.com/consensys/linea-monorepo/prover/maths/common/smartvectors"
	"github.com/consensys/linea-monorepo/prover/maths/field"
	"github.com/fxamacker/cbor/v2"
)

var CustomSerde = map[reflect.Type]struct{}{}

func init() {
	CustomSerde[TypeOfModuleWitnessGLPtr] = struct{}{}
	CustomSerde[TypeOfModuleWitnessLPPPtr] = struct{}{}
	CustomSerde[TypeOfSmartVector] = struct{}{}
}

type PackModuleWitness struct {
	ModuleName           string
	ModuleIndex          int
	SegmentModuleIndex   int
	TotalSegmentCount    []int
	Columns              [][]byte
	ReceivedValuesGlobal cbor.Tag
	VkMerkleRoot         *big.Int
}

/*
func marshalModuleWitnessGL(ser *Serializer, mod *distributed.ModuleWitnessGL) ([]byte, *serdeError) {
	packedModule := &PackModuleWitness{
		ModuleName:         string(mod.ModuleName),
		ModuleIndex:        mod.ModuleIndex,
		SegmentModuleIndex: mod.SegmentModuleIndex,
		TotalSegmentCount:  mod.TotalSegmentCount,
		VkMerkleRoot:       fieldToSmallBigInt(mod.VkMerkleRoot),
	}

	var buf = &bytes.Buffer{}
	if err := unsafe.WriteSlice(buf, mod.ReceivedValuesGlobal); err != nil {
		return nil, newSerdeErrorf("could not marshal array of field element: %w", err)
	}
	packedModule.ReceivedValuesGlobal = cbor.Tag{
		Number:  cborTagFieldElementsPacked,
		Content: buf.Bytes(),
	}

}

func unmarshalModuleWitnessGL(des *Deserializer, payload []byte) (*distributed.ModuleWitnessGL, *serdeError) {
} */

const (
	regular = iota
	pooled
	constant
	paddedCircularWindow
)

// marshalSmartVector is the CustomSerde-style marshaler:
//
//	ser  : serializer context (keeps helpers & maps available)
//	sv   : the SmartVector interface value to serialize
//
// returns an any that is CBOR-serializable (map/primitive/tag) and optional error.
// Corrected marshalSmartVector: uses pointer concrete types
func marshalSmartVector(ser *Serializer, sv smartvectors.SmartVector) (any, *serdeError) {
	if sv == nil {
		return nil, nil
	}

	switch concrete := sv.(type) {
	// pointer cases because methods have pointer receivers
	case *smartvectors.Regular:
		var buf bytes.Buffer
		// concrete is *smartvectors.Regular -> convert to []field.Element
		if err := unsafe.WriteSlice(&buf, []field.Element(*concrete)); err != nil {
			return nil, newSerdeErrorf("could not marshal Regular smartvector: %w", err)
		}
		return map[string]any{
			"t": regular,
			"b": cbor.Tag{Number: cborTagFieldElementsPacked, Content: buf.Bytes()},
			"l": len(*concrete),
		}, nil

	case *smartvectors.Pooled:
		var buf bytes.Buffer
		if err := unsafe.WriteSlice(&buf, []field.Element(concrete.Regular)); err != nil {
			return nil, newSerdeErrorf("could not marshal Pooled smartvector: %w", err)
		}
		return map[string]any{
			"t": pooled,
			"b": cbor.Tag{Number: cborTagFieldElementsPacked, Content: buf.Bytes()},
			"l": len(concrete.Regular),
		}, nil

	case *smartvectors.Constant:
		bi := fieldToSmallBigInt(concrete.Value)
		return map[string]any{
			"t": constant,
			"x": bi, // big.Int handled by your big.Int codex
			"l": concrete.Length,
		}, nil

	case *smartvectors.PaddedCircularWindow:
		var buf bytes.Buffer
		if err := unsafe.WriteSlice(&buf, concrete.Window_); err != nil {
			return nil, newSerdeErrorf("could not marshal PaddedCircularWindow window: %w", err)
		}
		padBig := fieldToSmallBigInt(concrete.PaddingVal_)
		return map[string]any{
			"t":   paddedCircularWindow,
			"b":   cbor.Tag{Number: cborTagFieldElementsPacked, Content: buf.Bytes()},
			"pad": padBig,
			"tot": concrete.TotLen_,
			"off": concrete.Offset_,
			"w":   len(concrete.Window_),
		}, nil

	default:
		return nil, newSerdeErrorf("unsupported SmartVector concrete type: %T", sv)
	}
}

// unmarshalSmartVector is the CustomSerde-style unmarshaler:
//
//	des : deserializer context
//	val : decoded CBOR payload for the SmartVector (map/primitive/tag)
//
// returns a concrete SmartVector implementation.
func unmarshalSmartVector(des *Deserializer, val any) (smartvectors.SmartVector, *serdeError) {
	if val == nil {
		return nil, nil
	}

	// Accept both map[string]any and map[interface{}]interface{}
	var m map[interface{}]interface{}
	switch t := val.(type) {
	case map[string]any:
		m = make(map[interface{}]interface{}, len(t))
		for k, v := range t {
			m[k] = v
		}
	case map[interface{}]interface{}:
		m = t
	default:
		return nil, newSerdeErrorf("invalid smartvector payload type: %T", val)
	}

	get := func(k string) (interface{}, bool) {
		if v, ok := m[k]; ok {
			return v, true
		}
		v2, ok2 := m[interface{}(k)]
		return v2, ok2
	}

	rawType, ok := get("t")
	if !ok {
		return nil, newSerdeErrorf("smartvector missing type tag 't'")
	}

	var typInt int
	switch v := rawType.(type) {
	case int:
		typInt = v
	case int64:
		typInt = int(v)
	case uint64:
		typInt = int(v)
	default:
		return nil, newSerdeErrorf("smartvector type tag not numeric: %T", rawType)
	}

	switch typInt {
	case regular, pooled:
		bAny, ok := get("b")
		if !ok {
			return nil, newSerdeErrorf("smartvector missing packed bytes 'b'")
		}

		var raw []byte
		switch x := bAny.(type) {
		case cbor.Tag:
			if x.Number != cborTagFieldElementsPacked {
				return nil, newSerdeErrorf("unexpected tag for smartvector packed elements: %d", x.Number)
			}
			bb, ok := x.Content.([]byte)
			if !ok {
				return nil, newSerdeErrorf("tagged smartvector content not []byte, got %T", x.Content)
			}
			raw = bb
		case []byte:
			raw = x
		default:
			return nil, newSerdeErrorf("unexpected type for smartvector packed bytes: %T", bAny)
		}

		r := bytes.NewReader(raw)
		slice, _, err := unsafe.ReadSlice[[]field.Element](r)
		if err != nil {
			return nil, newSerdeErrorf("could not read smartvector elements: %w", err)
		}

		if typInt == regular {
			// need pointer to Regular because pointer type implements SmartVector
			rv := smartvectors.Regular(slice)
			return &rv, nil
		}
		// pooled
		pv := smartvectors.Pooled{Regular: slice}
		return &pv, nil

	case constant:
		xAny, ok := get("x")
		if !ok {
			return nil, newSerdeErrorf("constant smartvector missing element 'x'")
		}
		bv, err := unmarshalBigInt(des, xAny, TypeOfBigInt)
		if err != nil {
			return nil, err.wrapPath("(constant.x)")
		}
		bi := bv.Interface().(*big.Int)
		var fe field.Element
		fe.SetBigInt(bi)

		length := 0
		if lAny, ok := get("l"); ok && lAny != nil {
			switch n := lAny.(type) {
			case uint64:
				length = int(n)
			case int:
				length = n
			case int64:
				length = int(n)
			}
		}

		cv := smartvectors.Constant{Value: fe, Length: length}
		return &cv, nil

	case paddedCircularWindow:
		bAny, ok := get("b")
		if !ok {
			return nil, newSerdeErrorf("padded smartvector missing 'b'")
		}
		var raw []byte
		switch x := bAny.(type) {
		case cbor.Tag:
			if x.Number != cborTagFieldElementsPacked {
				return nil, newSerdeErrorf("unexpected tag for padded smartvector: %d", x.Number)
			}
			bb, ok := x.Content.([]byte)
			if !ok {
				return nil, newSerdeErrorf("tagged padded content not []byte: %T", x.Content)
			}
			raw = bb
		case []byte:
			raw = x
		default:
			return nil, newSerdeErrorf("unexpected type for padded smartvector packed bytes: %T", bAny)
		}

		r := bytes.NewReader(raw)
		window, _, err := unsafe.ReadSlice[[]field.Element](r)
		if err != nil {
			return nil, newSerdeErrorf("could not read padded smartvector window: %w", err)
		}

		var pad field.Element
		if padAny, ok := get("pad"); ok && padAny != nil {
			bv, err := unmarshalBigInt(des, padAny, TypeOfBigInt)
			if err != nil {
				return nil, err.wrapPath("(padded.pad)")
			}
			pad.SetBigInt(bv.Interface().(*big.Int))
		}

		tot := 0
		off := 0
		if totAny, ok := get("tot"); ok && totAny != nil {
			switch n := totAny.(type) {
			case uint64:
				tot = int(n)
			case int:
				tot = n
			case int64:
				tot = int(n)
			}
		}
		if offAny, ok := get("off"); ok && offAny != nil {
			switch n := offAny.(type) {
			case uint64:
				off = int(n)
			case int:
				off = n
			case int64:
				off = int(n)
			}
		}

		pw := smartvectors.PaddedCircularWindow{
			Window_:     window,
			PaddingVal_: pad,
			TotLen_:     tot,
			Offset_:     off,
		}
		return &pw, nil

	default:
		return nil, newSerdeErrorf("unknown smartvector kind: %v", typInt)
	}
}
