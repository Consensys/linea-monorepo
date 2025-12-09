package serialization

import (
	"bytes"
	"math/big"
	"reflect"

	"github.com/consensys/gnark-crypto/utils/unsafe"
	"github.com/consensys/linea-monorepo/prover/maths/common/smartvectors"
	"github.com/consensys/linea-monorepo/prover/maths/common/smartvectorsext"
	"github.com/consensys/linea-monorepo/prover/maths/field/fext"

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

	// extension kinds (new)
	regularExt
	pooledExt
	constantExt
	paddedCircularWindowExt
)

const cborTagFextElementsPacked uint64 = 60002 // private-use CBOR tag for packed fext.Element bytes

// marshalSmartVector is the CustomSerde-style marshaler:
//
//	ser  : serializer context (keeps helpers & maps available)
//	sv   : the SmartVector interface value to serialize
//
// returns an any that is CBOR-serializable (map/primitive/tag) and optional error.
// Corrected marshalSmartVector: uses pointer concrete types

// NOTE: add these imports to the top of the file (adjust fext import path to your repo):
//
//   smartvectorsext "github.com/consensys/linea-monorepo/prover/maths/common/smartvectorsext"
//   fext "github.com/consensys/linea-monorepo/prover/maths/common/fext" // <- REPLACE with real path
//
// (Other imports already present in file remain unchanged.)

// marshalSmartVector: packs smartvectors into a CBOR-friendly wire form.
// Supports both base (field.Element) vectors and extension (fext.Element) vectors.
func marshalSmartVector(ser *Serializer, sv smartvectors.SmartVector) (any, *serdeError) {
	if sv == nil {
		return nil, nil
	}

	// Fast-path common concrete types (pointer receiver implementations)
	switch concrete := sv.(type) {
	case *smartvectors.Regular:
		var buf bytes.Buffer
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
			"x": bi, // big.Int handled by CustomCodex for big.Int
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
	}

	// --- Extension types (do NOT access unexported fields directly) ---
	// We rely on the public SmartVector helpers (IntoRegVecSaveAllocExt(), Len(), etc.)
	// and on public constructors (NewConstantExt, NewRegularExt, NewPaddedCircularWindowExt).
	// Add import alias for the ext package: smartvectorsext

	switch concrete := sv.(type) {
	case *smartvectorsext.RegularExt:
		var buf bytes.Buffer
		// write the []fext.Element into a byte blob
		if err := unsafe.WriteSlice(&buf, []fext.Element(*concrete)); err != nil {
			return nil, newSerdeErrorf("could not marshal RegularExt: %w", err)
		}
		return map[string]any{
			"t": regularExt,
			"b": cbor.Tag{Number: cborTagFextElementsPacked, Content: buf.Bytes()},
			"l": len(*concrete),
		}, nil

	case *smartvectorsext.PooledExt:
		var buf bytes.Buffer
		if err := unsafe.WriteSlice(&buf, []fext.Element(concrete.RegularExt)); err != nil {
			return nil, newSerdeErrorf("could not marshal PooledExt: %w", err)
		}
		return map[string]any{
			"t": pooledExt,
			"b": cbor.Tag{Number: cborTagFextElementsPacked, Content: buf.Bytes()},
			"l": len(concrete.RegularExt),
		}, nil

	case *smartvectorsext.ConstantExt:
		// safe way: ask the SmartVector for its ext backing slice (single element)
		// If the ConstantExt exposes an accessor you can use it; otherwise we derive:
		if exter, ok := sv.(interface{ IntoRegVecSaveAllocExt() []fext.Element }); ok {
			sl := exter.IntoRegVecSaveAllocExt()
			if len(sl) == 0 {
				return nil, newSerdeErrorf("ConstantExt produced empty backing slice")
			}
			var buf bytes.Buffer
			if err := unsafe.WriteSlice(&buf, sl); err != nil {
				return nil, newSerdeErrorf("could not marshal ConstantExt fext element: %w", err)
			}
			return map[string]any{
				"t": constantExt,
				"b": cbor.Tag{Number: cborTagFextElementsPacked, Content: buf.Bytes()},
				"l": sv.Len(),
			}, nil
		}
		// fallthrough to error if no ext helper present
		return nil, newSerdeErrorf("unsupported SmartVector concrete ext-type (ConstantExt) without IntoRegVecSaveAllocExt: %T", sv)

	case *smartvectorsext.PaddedCircularWindowExt:
		// Prefer to use exported accessors if the ext package provides them.
		// Try to detect accessor interface:
		type paddedExtAccessor interface {
			Window() []fext.Element
			PaddingVal() fext.Element
			TotLen() int
			Offset() int
		}
		if acc, ok := sv.(paddedExtAccessor); ok {
			var buf bytes.Buffer
			if err := unsafe.WriteSlice(&buf, acc.Window()); err != nil {
				return nil, newSerdeErrorf("could not marshal PaddedCircularWindowExt window: %w", err)
			}
			// pack pad as single-element fext blob
			var padBuf bytes.Buffer
			if err := unsafe.WriteSlice(&padBuf, []fext.Element{acc.PaddingVal()}); err != nil {
				return nil, newSerdeErrorf("could not marshal PaddedCircularWindowExt pad: %w", err)
			}
			return map[string]any{
				"t":   paddedCircularWindowExt,
				"b":   cbor.Tag{Number: cborTagFextElementsPacked, Content: buf.Bytes()},
				"pad": cbor.Tag{Number: cborTagFextElementsPacked, Content: padBuf.Bytes()},
				"tot": acc.TotLen(),
				"off": acc.Offset(),
				"w":   len(acc.Window()),
			}, nil
		}
		// Fallback: if no accessors, try generic IntoRegVecSaveAllocExt and encode full vector as a regularExt (lossy: loses padding/offset encoding)
		if exter, ok := sv.(interface{ IntoRegVecSaveAllocExt() []fext.Element }); ok {
			sl := exter.IntoRegVecSaveAllocExt()
			var buf bytes.Buffer
			if err := unsafe.WriteSlice(&buf, sl); err != nil {
				return nil, newSerdeErrorf("could not marshal PaddedCircularWindowExt fallback: %w", err)
			}
			// encode as regularExt (best-effort fallback when accessors are not exported)
			return map[string]any{
				"t": regularExt,
				"b": cbor.Tag{Number: cborTagFextElementsPacked, Content: buf.Bytes()},
				"l": len(sl),
			}, nil
		}
		return nil, newSerdeErrorf("unsupported SmartVector concrete ext-type (PaddedCircularWindowExt) without accessors or IntoRegVecSaveAllocExt: %T", sv)
	}

	// If we arrived here, we didn't recognize the concrete type
	return nil, newSerdeErrorf("unsupported SmartVector concrete type: %T", sv)
}

// unmarshalSmartVector: mirror of marshalSmartVector. Accepts both old string form and numeric form.
func unmarshalSmartVector(des *Deserializer, val any) (smartvectors.SmartVector, *serdeError) {
	if val == nil {
		return nil, nil
	}

	// Accept either map[string]any or map[interface{}]interface{}
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

	// two allowed encodings for 't':
	//  - string name (backwards compatible)
	//  - numeric int kind
	var kind int
	switch t := rawType.(type) {
	case string:
		switch t {
		case "regular":
			kind = regular
		case "pooled":
			kind = pooled
		case "constant":
			kind = constant
		case "padded":
			kind = paddedCircularWindow
		case "regularExt":
			kind = regularExt
		case "pooledExt":
			kind = pooledExt
		case "constantExt":
			kind = constantExt
		case "paddedExt":
			kind = paddedCircularWindowExt
		default:
			return nil, newSerdeErrorf("unknown smartvector kind string: %s", t)
		}
	case int:
		kind = t
	case int64:
		kind = int(t)
	case uint64:
		kind = int(t)
	default:
		return nil, newSerdeErrorf("smartvector type tag not numeric nor string: %T", rawType)
	}

	// helper to extract packed bytes (either field.Element or fext.Element)
	getPackedBytes := func(key string) ([]byte, *serdeError) {
		bAny, ok := get(key)
		if !ok {
			return nil, newSerdeErrorf("smartvector missing packed bytes '%s'", key)
		}
		switch x := bAny.(type) {
		case cbor.Tag:
			// allow either field tag or fext tag; caller will ensure tag is correct
			if content, ok := x.Content.([]byte); ok {
				return content, nil
			}
			return nil, newSerdeErrorf("tagged content not []byte for %s, got %T", key, x.Content)
		case []byte:
			return x, nil
		default:
			return nil, newSerdeErrorf("unexpected type for packed bytes %s: %T", key, bAny)
		}
	}

	switch kind {
	// field.Element kinds
	case regular, pooled:
		raw, err := getPackedBytes("b")
		if err != nil {
			return nil, err
		}
		r := bytes.NewReader(raw)
		slice, _, rerr := unsafe.ReadSlice[[]field.Element](r)
		if rerr != nil {
			return nil, newSerdeErrorf("could not read smartvector field elements: %w", rerr)
		}
		if kind == regular {
			rv := smartvectors.Regular(slice)
			return &rv, nil
		}
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
		raw, err := getPackedBytes("b")
		if err != nil {
			return nil, err
		}
		r := bytes.NewReader(raw)
		window, _, rerr := unsafe.ReadSlice[[]field.Element](r)
		if rerr != nil {
			return nil, newSerdeErrorf("could not read padded window elements: %w", rerr)
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

	// extension kinds: use fext reading
	case regularExt, pooledExt:
		raw, err := getPackedBytes("b")
		if err != nil {
			return nil, err
		}
		r := bytes.NewReader(raw)
		sliceExt, _, rerr := unsafe.ReadSlice[[]fext.Element](r)
		if rerr != nil {
			return nil, newSerdeErrorf("could not read smartvector fext elements: %w", rerr)
		}
		if kind == regularExt {
			rv := smartvectorsext.RegularExt(sliceExt)
			return &rv, nil
		}
		pv := smartvectorsext.PooledExt{RegularExt: smartvectorsext.RegularExt(sliceExt)}
		return &pv, nil

	case constantExt:
		// packed single-element fext blob
		raw, err := getPackedBytes("b")
		if err != nil {
			return nil, err
		}
		r := bytes.NewReader(raw)
		sliceExt, _, rerr := unsafe.ReadSlice[[]fext.Element](r)
		if rerr != nil {
			return nil, newSerdeErrorf("could not read constantExt fext element: %w", rerr)
		}
		if len(sliceExt) == 0 {
			return nil, newSerdeErrorf("empty fext slice for constantExt")
		}
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
		// use the public constructor
		ce := smartvectorsext.NewConstantExt(sliceExt[0], length)
		return ce, nil

	case paddedCircularWindowExt:
		raw, err := getPackedBytes("b")
		if err != nil {
			return nil, err
		}
		r := bytes.NewReader(raw)
		window, _, rerr := unsafe.ReadSlice[[]fext.Element](r)
		if rerr != nil {
			return nil, newSerdeErrorf("could not read padded ext window: %w", rerr)
		}

		// read pad tag which is a single-element packed fext blob
		var pad fext.Element
		if padAny, ok := get("pad"); ok && padAny != nil {
			switch x := padAny.(type) {
			case cbor.Tag:
				if bb, ok := x.Content.([]byte); ok {
					pr := bytes.NewReader(bb)
					sl, _, rerr := unsafe.ReadSlice[[]fext.Element](pr)
					if rerr != nil {
						return nil, newSerdeErrorf("could not read padded pad element: %w", rerr)
					}
					if len(sl) > 0 {
						pad = sl[0]
					}
				} else {
					return nil, newSerdeErrorf("padded pad content not []byte, got %T", x.Content)
				}
			case []byte:
				pr := bytes.NewReader(x)
				sl, _, rerr := unsafe.ReadSlice[[]fext.Element](pr)
				if rerr != nil {
					return nil, newSerdeErrorf("could not read padded pad element: %w", rerr)
				}
				if len(sl) > 0 {
					pad = sl[0]
				}
			default:
				return nil, newSerdeErrorf("unexpected type for padded pad: %T", padAny)
			}
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

		// Construct via public constructor
		pw := smartvectorsext.NewPaddedCircularWindowExt(window, pad, off, tot)
		return pw, nil

	default:
		return nil, newSerdeErrorf("unknown smartvector kind: %v", kind)
	}
}

// unmarshalSmartVector is the CustomSerde-style unmarshaler:
//
//	des : deserializer context
//	val : decoded CBOR payload for the SmartVector (map/primitive/tag)
//
// returns a concrete SmartVector implementation.

/*
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
} */
