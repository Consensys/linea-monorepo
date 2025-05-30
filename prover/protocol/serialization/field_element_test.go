package serialization

import (
	"math/big"
	"testing"

	"github.com/consensys/linea-monorepo/prover/maths/field"
)

// TestFieldElementSerialization tests round-trip serialization of field.Element and []field.Element.
func TestSerdeFE(t *testing.T) {
	// Helper to create a field.Element from a big.Int
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
		{
			name: "Zero",
			val:  field.Element{0, 0, 0, 0},
		},
		{
			name: "Small",
			val:  newFieldElement(42),
		},
		{
			name: "Large",
			val:  newFieldElement(1 << 60),
		},
		{
			name: "ModulusMinusOne",
			val: func() field.Element {
				var f field.Element
				f.SetOne()
				f.Neg(&f) // modulus - 1
				return f
			}(),
		},
	}

	// Test single field.Element
	for _, tt := range singleTests {
		t.Run("Single_"+tt.name, func(t *testing.T) {
			// Serialize
			bytes, err := Serialize(tt.val)
			if err != nil {
				t.Errorf("failed to serialize field.Element: %v", err)
				return
			}
			// Deserialize
			var result field.Element
			err = Deserialize(bytes, &result)
			if err != nil {
				t.Errorf("failed to deserialize field.Element: %v", err)
				return
			}
			// Compare
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
		{
			name: "Empty",
			val:  []field.Element{},
		},
		{
			name: "Single",
			val:  []field.Element{newFieldElement(42)},
		},
		{
			name: "Multiple",
			val: []field.Element{
				newFieldElement(0),
				newFieldElement(42),
				newFieldElement(1 << 60),
				func() field.Element {
					var f field.Element
					f.SetOne()
					f.Neg(&f)
					return f
				}(),
			},
		},
	}

	// Test []field.Element
	for _, tt := range arrayTests {
		t.Run("Array_"+tt.name, func(t *testing.T) {
			// Serialize
			bytes, err := Serialize(tt.val)
			if err != nil {
				t.Errorf("failed to serialize []field.Element: %v", err)
				return
			}
			// Deserialize
			var result []field.Element
			err = Deserialize(bytes, &result)
			if err != nil {
				t.Errorf("failed to deserialize []field.Element: %v", err)
				return
			}
			// Compare lengths
			if len(result) != len(tt.val) {
				t.Errorf("length mismatch: got %d, want %d", len(result), len(tt.val))
				return
			}
			// Compare elements
			for i := range tt.val {
				if result[i] != tt.val[i] {
					t.Errorf("element %d mismatch: got %v, want %v", i, result[i], tt.val[i])
				}
			}
		})
	}
}
