package serialization_test

import (
	"crypto/rand"
	"math/big"
	"testing"

	"github.com/consensys/linea-monorepo/prover/maths/field"
	"github.com/consensys/linea-monorepo/prover/protocol/serialization"
)

func TestSerdeFE(t *testing.T) {
	newFieldElement := func(n int64) field.Element {
		var f field.Element
		f.SetBigInt(big.NewInt(n))
		return f
	}

	// Test cases for single field.Element
	singleTests := []struct {
		name string
		val  field.Element
	}{
		{"Zero", field.Element{0, 0, 0, 0}},
		{"Small", newFieldElement(42)},
		{"Large", newFieldElement(1 << 60)},
		{
			"ModulusMinusOne",
			func() field.Element {
				var f field.Element
				f.SetOne()
				f.Neg(&f) // modulus - 1
				return f
			}(),
		},
		{
			"Random",
			func() field.Element {
				var f field.Element
				b := make([]byte, 32)
				_, _ = rand.Read(b)
				f.SetBytes(b)
				return f
			}(),
		},
	}

	for _, tt := range singleTests {
		t.Run("Single_"+tt.name, func(t *testing.T) {
			bytes, err := serialization.Serialize(tt.val)
			if err != nil {
				t.Fatalf("serialize error: %v", err)
			}
			var result field.Element
			if err := serialization.Deserialize(bytes, &result); err != nil {
				t.Fatalf("deserialize error: %v", err)
			}
			if result != tt.val {
				t.Errorf("mismatch: got %v, want %v", result, tt.val)
			}
		})
	}

	// Test cases for []field.Element
	arrayTests := []struct {
		name string
		val  []field.Element
	}{
		{"Empty", []field.Element{}},
		{"Single", []field.Element{newFieldElement(42)}},
		{"Multiple", []field.Element{
			newFieldElement(0),
			newFieldElement(42),
			newFieldElement(1 << 60),
			func() field.Element {
				var f field.Element
				f.SetOne()
				f.Neg(&f)
				return f
			}(),
		}},
		{
			"Repeated",
			[]field.Element{
				newFieldElement(123),
				newFieldElement(123),
				newFieldElement(123),
			},
		},
		{
			"BoundaryNearModulus",
			[]field.Element{
				func() field.Element {
					var f field.Element
					f.SetOne()
					f.Neg(&f)
					return f
				}(),
				newFieldElement(1),
			},
		},
		{
			"LargeArray",
			func() []field.Element {
				arr := make([]field.Element, 1000)
				for i := 0; i < len(arr); i++ {
					arr[i] = newFieldElement(int64(i * i))
				}
				return arr
			}(),
		},
		{
			"RandomElements",
			func() []field.Element {
				arr := make([]field.Element, 100)
				for i := range arr {
					var f field.Element
					b := make([]byte, 32)
					_, _ = rand.Read(b)
					f.SetBytes(b)
					arr[i] = f
				}
				return arr
			}(),
		},
	}

	for _, tt := range arrayTests {
		t.Run("Array_"+tt.name, func(t *testing.T) {
			bytes, err := serialization.Serialize(tt.val)
			if err != nil {
				t.Fatalf("serialize error: %v", err)
			}
			var result []field.Element
			if err := serialization.Deserialize(bytes, &result); err != nil {
				t.Fatalf("deserialize error: %v", err)
			}
			if len(result) != len(tt.val) {
				t.Fatalf("length mismatch: got %d, want %d", len(result), len(tt.val))
			}
			for i := range tt.val {
				if result[i] != tt.val[i] {
					t.Errorf("element %d mismatch: got %v, want %v", i, result[i], tt.val[i])
				}
			}
		})
	}
}

func TestFieldElementLimbMismatch(t *testing.T) {
	// Construct a field.Element with known limb values
	original := [4]uint64{
		4432961018360255618, // limb 0
		1234567890123456789, // limb 1
		9876543210987654321, // limb 2
		1111111111111111111, // limb 3
	}

	// Serialize it
	bytes, err := serialization.Serialize(original)
	if err != nil {
		t.Fatalf("failed to serialize: %v", err)
	}

	// Deserialize it
	var result [4]uint64
	if err := serialization.Deserialize(bytes, &result); err != nil {
		t.Fatalf("failed to deserialize: %v", err)
	}

	// Compare limbs manually and print diff
	for i := 0; i < 4; i++ {
		if result[i] != original[i] {
			t.Errorf("field.Element limb mismatch at [%d]: got %d, want %d", i, result[i], original[i])
		}
	}
}

// TestFieldToSmallBigInt verifies that fieldToSmallBigInt
// returns exactly the expected *big.Int for various edge cases,
// and that feeding that big.Int back into a field.Element recovers the original.
func TestFieldToSmallBigInt(t *testing.T) {
	// Helper to build a field.Element from a signed int64
	fromInt64 := func(x int64) field.Element {
		var fe field.Element
		fe.SetBigInt(big.NewInt(x))
		return fe
	}

	// Helper to build “modulus − 1”
	modulusMinusOne := func() field.Element {
		var fe field.Element
		fe.SetOne() // fe = 1
		fe.Neg(&fe) // fe = −1 mod p  ⇒  p − 1
		return fe
	}

	// Helper to build a field.Element from any *big.Int (even > 64 bits)
	fromBigInt := func(b *big.Int) field.Element {
		var fe field.Element
		fe.SetBigInt(b)
		return fe
	}

	// A 200‐bit value: 1<<200
	bit200 := new(big.Int).Lsh(big.NewInt(1), 200)

	tests := []struct {
		name   string
		input  field.Element // v
		expect *big.Int      // what fieldToSmallBigInt(v) must return
	}{
		{
			name:   "Zero",
			input:  field.Element{0, 0, 0, 0},
			expect: big.NewInt(0),
		},
		{
			name:   "Small-Positive",
			input:  fromInt64(42),
			expect: big.NewInt(42),
		},
		{
			name:  "ModulusMinusOne",
			input: modulusMinusOne(),
			// v = p−1, so −v = 1; that fits in int64 ⇒ return −1
			expect: big.NewInt(-1),
		},
		{
			name: "JustOverMaxInt64",
			// Let v = 1<<63. 1<<63 fits into uint64 but > MaxInt64,
			// so we fall to “full BigInt(+v)”.
			input:  fromBigInt(new(big.Int).SetUint64(1 << 63)),
			expect: new(big.Int).SetUint64(1 << 63),
		},
		{
			name:   "Large200Bit",
			input:  fromBigInt(bit200),
			expect: new(big.Int).Set(bit200),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := serialization.FieldToSmallBigInt(tt.input)

			// 1) Compare got vs expect
			if got.Cmp(tt.expect) != 0 {
				t.Fatalf("FieldToSmallBigInt(%v) = %v, want %v",
					tt.input, got, tt.expect)
			}

			// 2) Feed `got` back into a fresh field.Element, ensure it matches `tt.input`.
			var roundtripped field.Element
			roundtripped.SetBigInt(got)
			if roundtripped != tt.input {
				t.Errorf("round-trip mismatch: SetBigInt(%v) gave %v, want %v",
					got, roundtripped, tt.input)
			}
		})
	}
}
