package fiatshamir_koalabear

import (
	"math"

	"github.com/consensys/linea-monorepo/prover/crypto/poseidon2_koalabear"
	"github.com/consensys/linea-monorepo/prover/maths/common/smartvectors"
	"github.com/consensys/linea-monorepo/prover/maths/field"
	"github.com/consensys/linea-monorepo/prover/maths/field/fext"
	"github.com/consensys/linea-monorepo/prover/utils"
)

// https://blog.trailofbits.com/2022/04/18/the-frozen-heart-vulnerability-in-plonk/

type FS struct {
	h *poseidon2_koalabear.MDHasher
}

func NewFS() FS {
	return FS{h: poseidon2_koalabear.NewMDHasher()}
}

func (fs *FS) Update(vec ...field.Element) {
	fs.h.WriteElements(vec...)
}

func (fs *FS) UpdateExt(vec ...fext.Element) {
	pad := make([]field.Element, 4*len(vec))
	for i := 0; i < len(vec); i++ {
		pad[4*i].Set(&vec[i].B0.A0)
		pad[4*i+1].Set(&vec[i].B0.A1)
		pad[4*i+2].Set(&vec[i].B1.A0)
		pad[4*i+3].Set(&vec[i].B1.A1)
	}
	fs.h.WriteElements(pad...)
}
func (fs *FS) UpdateGeneric(vec ...fext.GenericFieldElem) {
	if len(vec) == 0 {
		return
	}

	// Marshal the elements in a vector of bytes
	for _, f := range vec {
		if f.GetIsBase() {
			fs.Update(f.Base)
		} else {
			fs.UpdateExt(f.Ext)
		}
	}
}
func (fs *FS) UpdateVec(vecs ...[]field.Element) {
	for i := range vecs {
		fs.Update(vecs[i]...)
	}
}

// UpdateVec updates the Fiat-Shamir state by passing one of more slices of
// field elements.
func (fs *FS) UpdateVecExt(vecs ...[]fext.Element) {
	for i := range vecs {
		fs.UpdateExt(vecs[i]...)
	}
}

// UpdateSV updates the FS state with a smart-vector. No-op if the smart-vector
// has a length of zero.
func (fs *FS) UpdateSV(sv smartvectors.SmartVector) {
	if sv.Len() == 0 {
		return
	}

	vec := make([]field.Element, sv.Len())
	sv.WriteInSlice(vec)
	fs.Update(vec...)
}

func (fs *FS) RandomField() field.Octuplet {
	res := fs.h.SumElement()
	fs.safeguardUpdate()
	return res
}

func (fs *FS) RandomFext() fext.Element {
	s := fs.h.SumElement()
	var res fext.Element
	res.B0.A0 = s[0]
	res.B0.A1 = s[1]
	res.B1.A0 = s[2]
	res.B1.A1 = s[3]

	fs.UpdateExt(fext.NewFromUint(0, 0, 0, 0)) // safefuard update
	return res
}

func (fs *FS) RandomManyIntegers(num, upperBound int) []int {

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

	defer fs.safeguardUpdate()

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
		digest := fs.h.Sum(nil)
		buffer := NewBitReader(digest[:], field.Bits-1)

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
		// twice in a row when exiting the func (fs *FS)tion. Recall that we have a
		// defer of the safeguard update in the func (fs *FS)tion. This handles the
		// edge-case where the number of requested field elements is a multiple
		// of the number of challenges we can generate with a single field
		// element.
		if len(res) >= num {
			return res
		}

		// This updates ensures that for the next iterations of the loop and the
		// next randomness comsumption uses a fresh randomness.
		fs.safeguardUpdate()
	}
}

func (fs *FS) safeguardUpdate() {
	fs.Update(field.Zero())
}
