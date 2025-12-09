package serialization

import (
	"testing"

	"github.com/fxamacker/cbor/v2"
	"github.com/stretchr/testify/require"

	"github.com/consensys/linea-monorepo/prover/maths/common/smartvectors"
	"github.com/consensys/linea-monorepo/prover/maths/common/smartvectorsext"
	"github.com/consensys/linea-monorepo/prover/maths/field"
	"github.com/consensys/linea-monorepo/prover/maths/field/fext"
)

// helper to create a small regular slice of n elements filled with start..start+n-1
func makeFieldSlice(start, n int) []field.Element {
	s := make([]field.Element, n)
	for i := 0; i < n; i++ {
		s[i] = field.NewElement(uint64(start + i))
	}
	return s
}

// helper to create a small regular slice of n fext.Elements filled with start..start+n-1
func makeExtSlice(start, n int) []fext.Element {
	s := make([]fext.Element, n)
	for i := 0; i < n; i++ {
		// create a simple element; the precise constructor may differ in your codebase
		s[i] = fext.NewElement(uint64(start+i), uint64(start+i))
	}
	return s
}

func TestMarshalUnmarshalSmartVector_Roundtrip(t *testing.T) {
	// Regular (pointer)
	origR := smartvectors.Regular(makeFieldSlice(1, 4))
	rPtr := &origR

	// Pooled (pointer)
	origP := smartvectors.Pooled{Regular: makeFieldSlice(10, 3)}
	pPtr := &origP

	// Constant (pointer)
	origC := smartvectors.Constant{Value: field.NewElement(7), Length: 5}
	cPtr := &origC

	// PaddedCircularWindow (pointer)
	window := makeFieldSlice(20, 3)
	origW := smartvectors.PaddedCircularWindow{
		Window_:     window,
		PaddingVal_: field.NewElement(99),
		TotLen_:     16,
		Offset_:     2,
	}
	wPtr := &origW

	cases := []struct {
		name  string
		in    smartvectors.SmartVector
		check func(t *testing.T, got smartvectors.SmartVector)
	}{
		{
			"regular",
			rPtr,
			func(t *testing.T, got smartvectors.SmartVector) {
				r, ok := got.(*smartvectors.Regular)
				require.True(t, ok, "expected *Regular, got %T", got)
				require.Equal(t, len(*rPtr), len(*r))
				for i := range *r {
					require.Equal(t, (*rPtr)[i].String(), (*r)[i].String())
				}
			},
		},
		{
			"pooled",
			pPtr,
			func(t *testing.T, got smartvectors.SmartVector) {
				p, ok := got.(*smartvectors.Pooled)
				require.True(t, ok, "expected *Pooled, got %T", got)
				require.Equal(t, len(origP.Regular), len(p.Regular))
				for i := range p.Regular {
					require.Equal(t, origP.Regular[i].String(), p.Regular[i].String())
				}
			},
		},
		{
			"constant",
			cPtr,
			func(t *testing.T, got smartvectors.SmartVector) {
				c, ok := got.(*smartvectors.Constant)
				require.True(t, ok, "expected *Constant, got %T", got)
				require.Equal(t, origC.Length, c.Length)
				require.Equal(t, origC.Value.String(), c.Value.String())
			},
		},
		{
			"padded",
			wPtr,
			func(t *testing.T, got smartvectors.SmartVector) {
				pw, ok := got.(*smartvectors.PaddedCircularWindow)
				require.True(t, ok, "expected *PaddedCircularWindow, got %T", got)
				require.Equal(t, origW.TotLen_, pw.TotLen_)
				require.Equal(t, origW.Offset_, pw.Offset_)
				require.Equal(t, origW.PaddingVal_.String(), pw.PaddingVal_.String())
				require.Equal(t, len(origW.Window_), len(pw.Window_))
				for i := range pw.Window_ {
					require.Equal(t, origW.Window_[i].String(), pw.Window_[i].String())
				}
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			ser := NewSerializer()

			// 1) Marshal into a CBOR-serializable Go value
			packed, mErr := marshalSmartVector(ser, tc.in)
			require.True(t, mErr == nil, "unexpected serde error: %#v", mErr)

			// 2) Full CBOR encode/decode cycle to emulate real wire path
			enc, err := encodeWithCBOR(packed)
			require.NoError(t, err)

			var decoded any
			// use the registered CBOR decoder you have; decode into an empty interface
			require.NoError(t, decodeWithCBOR(enc, &decoded))

			// 3) Unmarshal from the decoded CBOR value using a fresh Deserializer.
			// The Deserializer is lightweight for this test; we don't need a full PackedObject.
			des := NewDeserializer(&PackedObject{})
			got, sErr := unmarshalSmartVector(des, decoded)
			require.True(t, sErr == nil, "unexpected serde error: %#v", sErr)
			require.NotNil(t, got)

			// 4) semantic check
			tc.check(t, got)
		})
	}
}

func TestMarshalUnmarshalSmartVector_ExtRoundtrip(t *testing.T) {
	// Build extension testcases (use constructors existing in your smartvectors package)
	testCases := []struct {
		name  string
		in    smartvectors.SmartVector
		check func(t *testing.T, got smartvectors.SmartVector)
	}{
		{
			"ConstantExt-zero",
			smartvectorsext.NewConstantExt(fext.Zero(), 4),
			func(t *testing.T, got smartvectors.SmartVector) {
				_, ok := got.(*smartvectorsext.ConstantExt)
				require.True(t, ok, "expected *ConstantExt, got %T", got)
			},
		},
		{
			"ConstantExt-nonzero",
			smartvectorsext.NewConstantExt(fext.NewElement(42, 43), 4),
			func(t *testing.T, got smartvectors.SmartVector) {
				_, ok := got.(*smartvectorsext.ConstantExt)
				require.True(t, ok, "expected *ConstantExt, got %T", got)
			},
		},
		{
			"PaddedCircularWindowExt",
			smartvectorsext.NewPaddedCircularWindowExt(makeExtSlice(2, 3), fext.NewElement(7, 8), 1, 8),
			func(t *testing.T, got smartvectors.SmartVector) {
				_, ok := got.(*smartvectorsext.PaddedCircularWindowExt)
				require.True(t, ok, "expected *PaddedCircularWindowExt, got %T", got)
			},
		},

		// 	{
		// 		"RegularExt",
		// 		(&smartvectorsext.RegularExt{}).FromSlice(makeExtSlice(1, 5)), // or: smartvectors.RegularExt(makeExtSlice(...)) depending on API
		// 		func(t *testing.T, got smartvectors.SmartVector) {
		// 			r, ok := got.(*smartvectorsext.RegularExt)
		// 			require.True(t, ok, "expected *RegularExt, got %T", got)
		// 			require.Equal(t, 5, len(*r))
		// 			for i := range *r {
		// 				require.Equal(t, makeExtSlice(1, 5)[i].String(), (*r)[i].String())
		// 			}
		// 		},
		// 	},

		// 	{
		// 		"RotatedExt",
		// 		smartvectorsext.NewRotatedExt(&smartvectorsext.PooledExt{RegularExt: smartvectorsext.RegularExt(makeExtSlice(4, 4))}, 1),
		// 		func(t *testing.T, got smartvectors.SmartVector) {
		// 			_, ok := got.(*smartvectorsext.RotatedExt)
		// 			require.True(t, ok, "expected *RotatedExt, got %T", got)
		// 		},
		// 	},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ser := NewSerializer()

			// 1) Marshal into a CBOR-serializable Go value
			packed, mErr := marshalSmartVector(ser, tc.in)
			if mErr != nil {
				t.Fatalf("marshal error: %s", mErr.Error())
			}

			// 2) Full CBOR encode/decode cycle to emulate real wire path
			enc, encErr := encodeWithCBOR(packed)
			require.NoError(t, encErr)

			var decoded any
			require.NoError(t, decodeWithCBOR(enc, &decoded))

			// 3) Unmarshal from the decoded CBOR value using a fresh Deserializer.
			des := NewDeserializer(&PackedObject{})
			got, uErr := unmarshalSmartVector(des, decoded)
			if uErr != nil {
				t.Fatalf("unmarshal error: %s", uErr.Error())
			}
			require.NotNil(t, got)

			// 4) semantic check
			tc.check(t, got)
		})
	}
}

// quick sanity test to ensure cbor.Tag roundtrip works as expected (optional)
func TestCBORTag_Roundtrip(t *testing.T) {
	tag := cbor.Tag{Number: cborTagFieldElementsPacked, Content: []byte{1, 2, 3}}
	bs, err := encodeWithCBOR(tag)
	require.NoError(t, err)
	var out any
	require.NoError(t, decodeWithCBOR(bs, &out))
	// out should be decoded back as cbor.Tag or a map depending on decoder; we just assert no error
	_ = out
}
