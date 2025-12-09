package serialization

import (
	"bytes"
	"fmt"
	"hash"
	"math/big"
	"reflect"
	"strings"
	"sync"

	"github.com/consensys/gnark-crypto/ecc/bls12-377/fr/fft"
	"github.com/consensys/gnark-crypto/utils/unsafe"
	"github.com/consensys/gnark/frontend"
	"github.com/consensys/linea-monorepo/prover/crypto/mimc"
	"github.com/consensys/linea-monorepo/prover/crypto/ringsis"
	"github.com/consensys/linea-monorepo/prover/crypto/state-management/hashtypes"
	"github.com/consensys/linea-monorepo/prover/maths/common/smartvectors"
	"github.com/consensys/linea-monorepo/prover/maths/field"
	"github.com/consensys/linea-monorepo/prover/protocol/distributed"
	"github.com/consensys/linea-monorepo/prover/protocol/ifaces"
	"github.com/consensys/linea-monorepo/prover/zkevm/arithmetization"
	"github.com/fxamacker/cbor/v2"
)

// Use a vendor-specific tag number for homogeneous field elements packed as bytes.
// RFC 8746 reserves ranges for typed arrays but doesn't define this 377-field explicitly.
// Pick a private-use tag in the high range to avoid collisions.
const cborTagFieldElementsPacked uint64 = 60001

// CustomCodex represents an optional behavior for a specific type
type CustomCodex struct {
	Type reflect.Type
	Ser  func(ser *Serializer, val reflect.Value) (any, *serdeError)
	Des  func(des *Deserializer, val any, t reflect.Type) (reflect.Value, *serdeError)
}

var (
	CustomCodexes = map[reflect.Type]CustomCodex{}
)

func init() {

	CustomCodexes[TypeOfSmartVector] = CustomCodex{
		Type: TypeOfSmartVector,
		Ser: func(ser *Serializer, v reflect.Value) (any, *serdeError) {
			if v.IsNil() {
				return nil, nil
			}
			sv := v.Interface().(smartvectors.SmartVector)
			return marshalSmartVector(ser, sv)
		},
		Des: func(des *Deserializer, val any, t reflect.Type) (reflect.Value, *serdeError) {
			sv, err := unmarshalSmartVector(des, val)
			if err != nil {
				return reflect.Value{}, err
			}
			if sv == nil {
				return reflect.Zero(t), nil
			}

			// Create a reflect.Value of the exact target type (the interface type),
			// then set the concrete value into it so its Type() == t.
			rv := reflect.New(t).Elem()
			rv.Set(reflect.ValueOf(sv))
			return rv, nil
			// return reflect.ValueOf(sv), nil
		},
	}

	CustomCodexes[TypeOfModuleWitnessGLPtr] = CustomCodex{
		Type: TypeOfModuleWitnessGLPtr,
		Ser: func(ser *Serializer, v reflect.Value) (any, *serdeError) {
			if v.IsNil() {
				return nil, nil
			}
			mod := v.Interface().(*distributed.ModuleWitnessGL)
			b, err := marshalModuleWitnessGL(ser, mod)
			if err != nil {
				return nil, err
			}
			// raw bytes stored directly in CBOR
			return b, nil
		},
		Des: func(des *Deserializer, val any, t reflect.Type) (reflect.Value, *serdeError) {
			if val == nil {
				return reflect.Zero(t), nil
			}
			raw, ok := val.([]byte)
			if !ok {
				return reflect.Value{}, newSerdeErrorf("ModuleWitnessGL expects bytes, got %T", val)
			}
			mod, err := unmarshalModuleWitnessGL(des, raw)
			if err != nil {
				return reflect.Value{}, err
			}
			return reflect.ValueOf(mod), nil
		},
	}

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

	CustomCodexes[TypeOfArrOfFieldElement] = CustomCodex{
		Type: TypeOfArrOfFieldElement,
		Ser:  marshalArrayOfFieldElement,
		Des:  unmarshalArrayOfFieldElement,
	}

	CustomCodexes[TypeOfArithmetization] = CustomCodex{
		Type: TypeOfArithmetization,
		Ser:  marshalArithmetization,
		Des:  unmarshalArithmetization,
	}

	CustomCodexes[TypeOfFrontendVariable] = CustomCodex{
		Type: TypeOfFrontendVariable,
		Ser:  marshalFrontendVariable,
		Des:  unmarshalFrontendVariable,
	}

	CustomCodexes[TypeOfHashFuncGenerator] = CustomCodex{
		Type: TypeOfHashFuncGenerator,
		Ser:  marshalAsEmptyStruct,
		Des:  unmarshalHashGenerator,
	}

	CustomCodexes[TypeOfHashTypeHasher] = CustomCodex{
		Type: TypeOfHashTypeHasher,
		Ser:  marshalHashTypeHasher,
		Des:  unmarshalHashTypeHasher,
	}

	CustomCodexes[reflect.TypeOf(sync.Mutex{})] = CustomCodex{
		Type: reflect.TypeOf(sync.Mutex{}),
		Ser:  marshalAsNil,
		Des:  unmarshalAsZero,
	}

	CustomCodexes[TypeOfRingSisKeyPtr] = CustomCodex{
		Type: TypeOfRingSisKeyPtr,
		Ser:  marshalRingSisKey,
		Des:  unmarshalRingSisKey,
	}

	CustomCodexes[TypeOfGnarkFFTDomainPtr] = CustomCodex{
		Type: TypeOfGnarkFFTDomainPtr,
		Ser:  marshalGnarkFFTDomain,
		Des:  unmarshalGnarkFFtDomain,
	}

	CustomCodexes[reflect.TypeOf(smartvectors.Regular{})] = CustomCodex{
		Type: reflect.TypeOf(smartvectors.Regular{}),
		Ser:  marshalArrayOfFieldElement,
		Des:  unmarshalArrayOfFieldElement,
	}

	CustomCodexes[TypeOfMutexPtr] = CustomCodex{
		Type: TypeOfMutexPtr,
		Ser:  marshalAsEmptyStruct,
		Des:  makeNewObject,
	}
}

func marshalRingSisKey(ser *Serializer, val reflect.Value) (any, *serdeError) {
	key, ok := val.Interface().(*ringsis.Key)
	if !ok {
		return nil, newSerdeErrorf("illegal cast of val of type %T to %v", val, TypeOfRingSisKeyPtr)
	}
	return key.KeyGen.MaxNumFieldToHash, nil
}

func unmarshalRingSisKey(des *Deserializer, val any, _ reflect.Type) (reflect.Value, *serdeError) {
	maxNumFieldToHash, ok := val.(uint64)
	if !ok {
		return reflect.Value{}, newSerdeErrorf("illegal cast of val of type %T to int", val)
	}
	ringSiskey := ringsis.GenerateKey(ringsis.StdParams, int(maxNumFieldToHash))
	return reflect.ValueOf(ringSiskey), nil
}

func marshalGnarkFFTDomain(ser *Serializer, val reflect.Value) (any, *serdeError) {
	domain := val.Interface().(*fft.Domain)

	if domain == nil {
		return nil, nil
	}

	var buf bytes.Buffer
	if _, err := domain.WriteTo(&buf); err != nil {
		return nil, newSerdeErrorf("could not marshal fft.Domain: %w", err)
	}
	return buf.Bytes(), nil
}

func unmarshalGnarkFFtDomain(des *Deserializer, val any, _ reflect.Type) (reflect.Value, *serdeError) {

	// nil case
	if val == nil {
		return reflect.Zero(TypeOfGnarkFFTDomainPtr), nil
	}

	// Expect a []byte coming from CBOR decoding
	var b []byte
	switch v := val.(type) {
	case []byte:
		b = v
	default:
		// defensive: CBOR typically decodes bytes to []byte, but return a helpful error if not.
		return reflect.Value{}, newSerdeErrorf("expected []byte for fft.Domain deserialization, got %T", val)
	}

	d := &fft.Domain{}
	if _, err := d.ReadFrom(bytes.NewReader(b)); err != nil {
		return reflect.Value{}, newSerdeErrorf("could not unmarshal fft.Domain: %w", err)
	}

	return reflect.ValueOf(d), nil
}

func marshalFieldElement(_ *Serializer, val reflect.Value) (any, *serdeError) {
	f := val.Interface().(field.Element)
	bi := fieldToSmallBigInt(f)
	f.BigInt(bi)
	return marshalBigInt(nil, reflect.ValueOf(bi))
}

func unmarshalFieldElement(_ *Deserializer, val any, _ reflect.Type) (reflect.Value, *serdeError) {

	f, err := unmarshalBigInt(nil, val, TypeOfBigInt)
	if err != nil {
		return reflect.Value{}, err
	}

	var fe field.Element
	fe.SetBigInt(f.Interface().(*big.Int))
	return reflect.ValueOf(fe), nil
}

func marshalBigInt(_ *Serializer, val reflect.Value) (any, *serdeError) {
	return val.Interface().(*big.Int), nil
}

func unmarshalBigInt(_ *Deserializer, val any, _ reflect.Type) (reflect.Value, *serdeError) {
	switch v := val.(type) {
	case big.Int:
		return reflect.ValueOf(&v), nil
	case int64:
		return reflect.ValueOf(big.NewInt(v)), nil
	case uint64:
		return reflect.ValueOf(new(big.Int).SetUint64(v)), nil
	default:
		return reflect.Value{}, newSerdeErrorf("invalid type: %T, value: %++v", val, val)
	}
}

// marshalArrayOfFieldElement: add CBOR tag wrapper.
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
	// Packing field elements as a single tagged byte string avoids element-by-element CBOR encoding/decoding,
	// cutting per-element reflection, encoder work, intermediate allocations, and per-item headers;
	// The optimization replaces a CBOR array of N field.Element items with a single tagged byte string whose content is the N elements
	// serialized contiguously in native limb form, then wrapped once with a private tag (e.g., 60001) indicating “packed field elements.”
	// This changes the wire shape from O(N) separate CBOR items to one item containing O(N) bytes. Saves about 55GiB of runtime memory.
	return cbor.Tag{
		Number:  cborTagFieldElementsPacked,
		Content: buf.Bytes(),
	}, nil
}

// unmarshalArrayOfFieldElement: accept either tagged content or raw []byte for backward compatibility.
func unmarshalArrayOfFieldElement(_ *Deserializer, val any, t reflect.Type) (reflect.Value, *serdeError) {
	var raw []byte

	switch x := val.(type) {
	// The tagged byte string path simply extracts the []byte content and reconstructs []field.Element using unsafe.ReadSlice
	// in one pass, instead of driving the decoder through N element decodes and reflection-based assignments.
	// This is a single decode step on the CBOR side plus a single contiguous read on the application side.
	// It avoids per-element CBOR encode/decode and reflection, replacing N items with a single tag+byte-string and a
	// single-pass binary read.
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

func marshalArithmetization(ser *Serializer, val reflect.Value) (any, *serdeError) {

	res, err := ser.PackStructObject(val)
	if err != nil {
		return nil, newSerdeErrorf("could not marshal arithmetization: %w", err)
	}

	return res, nil
}

func unmarshalArithmetization(des *Deserializer, val any, _ reflect.Type) (reflect.Value, *serdeError) {
	var errA error
	//
	if v_, ok := val.(PackedStructObject); ok {
		val = []any(v_)
	}
	res, err := des.UnpackStructObject(val.([]any), TypeOfArithmetization)
	if err != nil {
		return reflect.Value{}, newSerdeErrorf("could not unmarshal arithmetization: %w", err)
	}
	arith := res.Interface().(arithmetization.Arithmetization)
	// Parse binary file
	arith.BinaryFile, arith.Metadata, errA = arithmetization.UnmarshalZkEVMBin(arith.ZkEVMBin)
	if errA != nil {
		return reflect.Value{}, newSerdeErrorf("could not unmarshal arithmetization: %w", err)
	}
	// Compile binary file into an air.Schema
	arith.AirSchema, arith.LimbMapping = arithmetization.CompileZkevmBin(arith.BinaryFile, arith.Settings.OptimisationLevel)
	// Done
	return reflect.ValueOf(arith), nil
}

func marshalFrontendVariable(ser *Serializer, val reflect.Value) (any, *serdeError) {

	var (
		variable = val.Interface().(frontend.Variable)
		bi       = &big.Int{}
	)

	switch v := variable.(type) {
	case int:
		bi.SetInt64(int64(v))
	case uint:
		bi.SetUint64(uint64(v))
	case int64:
		bi.SetInt64(int64(v))
	case uint64:
		bi.SetUint64(v)
	case int32:
		bi.SetInt64(int64(v))
	case uint32:
		bi.SetUint64(uint64(v))
	case int16:
		bi.SetInt64(int64(v))
	case uint16:
		bi.SetUint64(uint64(v))
	case int8:
		bi.SetInt64(int64(v))
	case uint8:
		bi.SetUint64(uint64(v))
	case field.Element:
		bi = fieldToSmallBigInt(v)
	case big.Int:
		*bi = v
	case *big.Int:
		bi = v
	case string:
		bi.SetString(v, 0)
	default:
		if variable == nil {
			return nil, nil
		}

		// The check cannot be done on val.Type as it would return
		// [frontend.Variable]. The check is somewhat fragile as it rely on
		// the type name and the package name. We return nil in that case, be
		// -cause it signifies that the variable belongs to a circuit that has
		// been compiled by gnark. That information is not relevant to
		// serialize.
		if strings.Contains(reflect.TypeOf(variable).String(), "expr.Term") {
			return nil, nil
		}

		return nil, newSerdeErrorf("invalid type for a frontend variable: %T, value: %++v, type.string=%v", variable, variable, reflect.TypeOf(variable).String())
	}

	res, err := marshalBigInt(ser, reflect.ValueOf(bi))
	if err != nil {
		return nil, newSerdeErrorf("could not marshal frontend variable: %w", err)
	}

	return res, nil
}

func unmarshalFrontendVariable(des *Deserializer, val any, _ reflect.Type) (reflect.Value, *serdeError) {

	bi, err := unmarshalBigInt(des, val, TypeOfBigInt)
	if err != nil {
		return reflect.Value{}, newSerdeErrorf("could not unmarshal frontend variable: %w", err)
	}

	v := reflect.New(TypeOfFrontendVariable).Elem()
	v.Set(reflect.ValueOf(bi))
	return v, nil
}

func unmarshalHashGenerator(des *Deserializer, val any, _ reflect.Type) (reflect.Value, *serdeError) {
	f := func() hash.Hash {
		return mimc.NewMiMC()
	}
	return reflect.ValueOf(f), nil
}

func marshalHashTypeHasher(ser *Serializer, val reflect.Value) (any, *serdeError) {
	return nil, nil
}

func unmarshalHashTypeHasher(des *Deserializer, val any, _ reflect.Type) (reflect.Value, *serdeError) {
	return reflect.ValueOf(hashtypes.MiMC), nil
}

// marshalAsNil is a custom serialization function that marshals the given value
// to nil. It is used for types that are not meant to be serialized, such as
// functions.
func marshalAsNil(_ *Serializer, _ reflect.Value) (any, *serdeError) {
	return nil, nil
}

// unmarshalAsZero is a custom deserialization function that unmarshals the
// given value to zero. It is meant for the type that are not intended to be
// serialized.
func unmarshalAsZero(des *Deserializer, val any, t reflect.Type) (reflect.Value, *serdeError) {
	return reflect.Zero(t), nil
}

// This converts the field.Element to a smaller big.Int. This is done to
// reduce the size of the CBOR encoding. The backward conversion is automatically
// done [field.SetBigInt] as it handles negative values.
func fieldToSmallBigInt(v field.Element) *big.Int {
	neg := new(field.Element).Neg(&v)
	if neg.IsUint64() {
		n := neg.Uint64()
		return new(big.Int).SetInt64(-int64(n))
	}

	bi := &big.Int{}
	v.BigInt(bi)
	return bi
}

// marshalAsEmptyStruct is a custom serialization function that marshals the
// given value to an empty struct. It is used for types that are not meant to be
// serialized, such as functions.
func marshalAsEmptyStruct(_ *Serializer, _ reflect.Value) (any, *serdeError) {
	return struct{}{}, nil
}

// makeNewPtr creates an object using reflect.New and is indicated for pointer
// types as it creates a pointer to zero object rather than returning a nil
// pointer. The provided type must be a pointer type.
func makeNewObject(_ *Deserializer, _ any, t reflect.Type) (reflect.Value, *serdeError) {
	if t.Kind() != reflect.Ptr {
		return reflect.Value{}, newSerdeErrorf("type %v is not a pointer type", t.String())
	}
	return reflect.New(t.Elem()), nil
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

func marshalModuleWitnessGL(ser *Serializer, mod *distributed.ModuleWitnessGL) ([]byte, *serdeError) {
	if mod == nil {
		return nil, nil
	}

	// base fields
	packedModule := &PackModuleWitness{
		ModuleName:         string(mod.ModuleName),
		ModuleIndex:        mod.ModuleIndex,
		SegmentModuleIndex: mod.SegmentModuleIndex,
		TotalSegmentCount:  mod.TotalSegmentCount,
		VkMerkleRoot:       fieldToSmallBigInt(mod.VkMerkleRoot),
	}

	// ReceivedValuesGlobal -> pack as tagged packed bytes
	{
		var buf = &bytes.Buffer{}
		if err := unsafe.WriteSlice(buf, mod.ReceivedValuesGlobal); err != nil {
			return nil, newSerdeErrorf("could not marshal array of field element: %w", err)
		}
		packedModule.ReceivedValuesGlobal = cbor.Tag{
			Number:  cborTagFieldElementsPacked,
			Content: buf.Bytes(),
		}
	}

	// Columns: we want Columns to be a slice indexed by column backreference.
	// For each column in mod.Columns:
	//  - call ser.PackColumnID to ensure ColumnIDs contains it and get backref index
	//  - marshal the smartvector into a CBOR-serializable Go value (marshalSmartVector)
	//  - encode that value with CBOR and store the resulting []byte at the backref index
	colBlobs := map[int][]byte{} // temporary store idx -> encoded bytes
	var maxIdx int
	for colID, sv := range mod.Columns {
		br, err := ser.PackColumnID(colID)
		if err != nil {
			return nil, err.wrapPath(fmt.Sprintf("PackColumnID(%s)", string(colID)))
		}
		idx := int(br)
		if idx > maxIdx {
			maxIdx = idx
		}

		// marshal SmartVector -> any
		packedAny, mErr := marshalSmartVector(ser, sv)
		if mErr != nil {
			return nil, mErr.wrapPath(fmt.Sprintf("column(%s)", string(colID)))
		}

		// CBOR encode the marshaled representation into a []byte
		enc, encErr := encodeWithCBOR(packedAny)
		if encErr != nil {
			return nil, newSerdeErrorf("could not encode smartvector for column %s: %w", string(colID), encErr)
		}
		colBlobs[idx] = enc
	}

	// Now build the final slice sized to current number of ColumnIDs.
	// Note: ser.PackedObject.ColumnIDs may have grown during PackColumnID calls,
	// so we use its final length so indexes line up with UnpackColumnID.
	nCols := len(ser.PackedObject.ColumnIDs)
	columnsSlice := make([][]byte, nCols)
	// fill known entries
	for idx, blob := range colBlobs {
		if idx < 0 || idx >= nCols {
			// shouldn't happen; defensive
			return nil, newSerdeErrorf("internal: computed column index out of range: %d (nCols=%d)", idx, nCols)
		}
		columnsSlice[idx] = blob
	}
	// absent entries remain nil (will be encoded as nil/empty when CBOR-encoded by encodeWithCBOR)

	packedModule.Columns = columnsSlice

	// Final: CBOR-encode entire packedModule
	out, err := encodeWithCBOR(packedModule)
	if err != nil {
		return nil, newSerdeErrorf("could not CBOR-encode PackModuleWitness: %w", err)
	}
	return out, nil
}

func unmarshalModuleWitnessGL(des *Deserializer, payload []byte) (*distributed.ModuleWitnessGL, *serdeError) {
	if len(payload) == 0 {
		return nil, nil
	}

	var packed PackModuleWitness
	if err := decodeWithCBOR(payload, &packed); err != nil {
		return nil, newSerdeErrorf("could not decode PackModuleWitness: %w", err)
	}

	res := &distributed.ModuleWitnessGL{
		ModuleName:         distributed.ModuleName(packed.ModuleName),
		ModuleIndex:        packed.ModuleIndex,
		SegmentModuleIndex: packed.SegmentModuleIndex,
		TotalSegmentCount:  packed.TotalSegmentCount,
		Columns:            make(map[ifaces.ColID]smartvectors.SmartVector),
	}

	// VkMerkleRoot
	if packed.VkMerkleRoot != nil {
		var fe field.Element
		fe.SetBigInt(packed.VkMerkleRoot)
		res.VkMerkleRoot = fe
	}

	// ReceivedValuesGlobal: expect a cbor.Tag with packed field elements
	if packed.ReceivedValuesGlobal.Number != 0 || packed.ReceivedValuesGlobal.Content != nil {
		if packed.ReceivedValuesGlobal.Number != cborTagFieldElementsPacked {
			return nil, newSerdeErrorf("unexpected tag for ReceivedValuesGlobal: %d", packed.ReceivedValuesGlobal.Number)
		}
		raw, ok := packed.ReceivedValuesGlobal.Content.([]byte)
		if !ok {
			return nil, newSerdeErrorf("ReceivedValuesGlobal tag content not []byte but %T", packed.ReceivedValuesGlobal.Content)
		}
		r := bytes.NewReader(raw)
		slice, _, err := unsafe.ReadSlice[[]field.Element](r)
		if err != nil {
			return nil, newSerdeErrorf("could not read ReceivedValuesGlobal elements: %w", err)
		}
		res.ReceivedValuesGlobal = slice
	}

	// Columns: packed.Columns is a slice aligned with des.PackedObject.ColumnIDs
	// We must ensure that des.PackedObject.ColumnIDs exists and has enough entries.
	if len(packed.Columns) > len(des.PackedObject.ColumnIDs) {
		// It's possible (corrupt input) - fail to be safe
		return nil, newSerdeErrorf("mismatch: packed.Columns length %d > des.PackedObject.ColumnIDs length %d", len(packed.Columns), len(des.PackedObject.ColumnIDs))
	}

	for idx, blob := range packed.Columns {
		// if there is no registered ColumnID for this index, skip
		if idx >= len(des.PackedObject.ColumnIDs) {
			// defensive: ignore trailing blobs with no column id
			continue
		}
		colName := des.PackedObject.ColumnIDs[idx]
		colID := ifaces.ColID(colName)

		// If blob is nil (wasn't present), then value is nil smartvector
		if len(blob) == 0 {
			res.Columns[colID] = nil
			continue
		}

		// decode CBOR blob into `any`
		var decoded any
		if err := decodeWithCBOR(blob, &decoded); err != nil {
			return nil, newSerdeErrorf("could not decode smartvector for column %s: %w", colName, err)
		}

		// unmarshal SmartVector
		sv, sErr := unmarshalSmartVector(des, decoded)
		if sErr != nil {
			return nil, sErr.wrapPath(fmt.Sprintf("column(%s)", colName))
		}
		res.Columns[colID] = sv
	}

	return res, nil
}

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
