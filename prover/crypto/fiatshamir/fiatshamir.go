package fiatshamir

import (
	"math"

	"github.com/consensys/linea-monorepo/prover/crypto/state-management/hashtypes"
	"github.com/consensys/linea-monorepo/prover/maths/common/smartvectors"
	"github.com/consensys/linea-monorepo/prover/maths/field"
	"github.com/consensys/linea-monorepo/prover/maths/field/fext"
	"github.com/consensys/linea-monorepo/prover/utils"
)

// State holds a Fiat-Shamir state. The Fiat-Shamir state can be updated by
// providing field elements and can be consumed to generate either field
// elements or sequences of small integers. The Fiat-Shamir instantiation relies
// on the MiMC hash function and uses the following strategy for generating
// random public coins:
//
//   - The messages are appended to the hasher as it is received by the Fiat-
//     Shamir state, field element by field element
//   - When requested by the caller, the State can generates a field element by
//     hashing the transcript. Recall that we are using a SNARK-friendly hash
//     function and that its output should be interpreted as a field element.
//   - Everytime a field element has been generated, the hasher is updated by
//     appending an artificial message: field.Element(0). This is to ensure
//     that all generated field elements are independants.
//
// To be safely used, the Fiat-Shamir State should initialized by passing an
// initial field element summarizing the protocol in its full-extent. This is
// to prevent rogue protocol attack as in the Frozen Heart vulnerability.
//
// https://blog.trailofbits.com/2022/04/18/the-frozen-heart-vulnerability-in-plonk/

func Update(h *hashtypes.Poseidon2FieldHasherDigest, vec ...field.Element) {
	h.WriteElements(vec)
}

func UpdateExt(h *hashtypes.Poseidon2FieldHasherDigest, vec ...fext.Element) {
	for _, f := range vec {
		h.WriteElement(f.B0.A0)
		h.WriteElement(f.B0.A1)
		h.WriteElement(f.B1.A0)
		h.WriteElement(f.B1.A1)
	}
}
func UpdateGeneric(h *hashtypes.Poseidon2FieldHasherDigest, vec ...fext.GenericFieldElem) {
	if len(vec) == 0 {
		return
	}

	// Marshal the elements in a vector of bytes
	for _, f := range vec {
		if f.GetIsBase() {
			Update(h, f.Base)
		} else {
			UpdateExt(h, f.Ext)
		}
	}
}
func UpdateVec(h *hashtypes.Poseidon2FieldHasherDigest, vecs ...[]field.Element) {
	for i := range vecs {
		Update(h, vecs[i]...)
	}
}

// UpdateVec updates the Fiat-Shamir state by passing one of more slices of
// field elements.
func UpdateVecExt(h *hashtypes.Poseidon2FieldHasherDigest, vecs ...[]fext.Element) {
	for i := range vecs {
		UpdateExt(h, vecs[i]...)
	}
}

// UpdateSV updates the FS state with a smart-vector. No-op if the smart-vector
// has a length of zero.
func UpdateSV(h *hashtypes.Poseidon2FieldHasherDigest, sv smartvectors.SmartVector) {
	if sv.Len() == 0 {
		return
	}

	vec := make([]field.Element, sv.Len())
	sv.WriteInSlice(vec)
	Update(h, vec...)
}

func RandomFext(h *hashtypes.Poseidon2FieldHasherDigest) fext.Element {
	s := h.SumElements(nil)
	var res fext.Element
	res.B0.A0 = s[0]
	res.B0.A1 = s[1]
	res.B1.A0 = s[2]
	res.B1.A1 = s[3]

	UpdateExt(h, fext.NewFromUint(0, 0, 0, 0)) // safefuard update
	return res
}

func RandomManyIntegers(h *hashtypes.Poseidon2FieldHasherDigest, num, upperBound int) []int {

	// Even `1` would be wierd, there would be only one acceptable coin value.
	if upperBound < 1 {
		utils.Panic("UpperBound was %v", upperBound)
	}

	if !utils.IsPowerOfTwo(upperBound) {
		utils.Panic("Expected a power of two but got %v", upperBound)
	}

	if num == 0 {
		return []int{}
	}

	defer safeguardUpdate(h)

	var (
		// challsBitSize stores the number of bits required instantiate each
		// small integer.
		challsBitSize = math.Ceil(math.Log2(float64(upperBound)))
		// Number of challenges computable with one call to hash (implicitly,
		// the division is rounded down). The "-1" corresponds to the fact that
		// the most significant bit of a field element cannot be assumed to
		// contain exactly 1 bit of entropy since the modulus of the field is
		// never a power of 2.
		maxNumChallsPerDigest = (field.Bits - 1) / int(challsBitSize)
		// res stores the preallocated result slice, to which all generated
		// small integers will be appended.
		res = make([]int, 0, num)
	)

	for {
		digest := h.Sum(nil)
		buffer := NewBitReader(digest, field.Bits-1)

		// Increase the counter

		for i := 0; i < maxNumChallsPerDigest; i++ {
			// Stopping condition, we computed enough challenges
			if len(res) >= num {
				return res
			}

			newChall, err := buffer.ReadInt(int(challsBitSize))
			if err != nil {
				utils.Panic("could not instantiate the buffer for a single field element")
			}
			res = append(res, int(newChall)%upperBound)
		}

		// This is guarded by the condition to prevent the [State] updating
		// twice in a row when exiting the function. Recall that we have a
		// defer of the safeguard update in the function. This handles the
		// edge-case where the number of requested field elements is a multiple
		// of the number of challenges we can generate with a single field
		// element.
		if len(res) >= num {
			return res
		}

		// This updates ensures that for the next iterations of the loop and the
		// next randomness comsumption uses a fresh randomness.
		safeguardUpdate(h)
	}
}

func safeguardUpdate(h *hashtypes.Poseidon2FieldHasherDigest) {
	Update(h, field.NewElement(0))
}
