package fiatshamir_bls12377

import (
	"github.com/consensys/gnark-crypto/ecc/bls12-377/fr"
	"github.com/consensys/linea-monorepo/prover/crypto/encoding"
	"github.com/consensys/linea-monorepo/prover/crypto/poseidon2_bls12377"
	"github.com/consensys/linea-monorepo/prover/maths/field"
)

// https://blog.trailofbits.com/2022/04/18/the-frozen-heart-vulnerability-in-plonk/

type FS struct {
	h *poseidon2_bls12377.MDHasher
}

func NewFS() FS {
	return FS{h: poseidon2_bls12377.NewMDHasher()}
}

// --------------------------------------------------
// List of methods to updae the FS state with frElmts
func (fs *FS) UpdateFrElmt(vec ...fr.Element) {
	fs.h.WriteElements(vec...)
}

func (fs *FS) UpdateVecFrElmt(vecs ...[]fr.Element) {
	for i := range vecs {
		fs.UpdateFrElmt(vecs[i]...)
	}
}

func (fs *FS) RandomFieldFrElmt() fr.Element {
	res := fs.h.SumElement()
	fs.safeguardUpdate()
	return res
}

// ------------------------------------------------------
// List of methods to updae the FS state with koala elmts

func (fs *FS) UpdateElmts(vec ...field.Element) {
	fs.h.WriteKoalabearElements(vec...)
}

func (fs *FS) UpdateVecElmts(vec ...[]field.Element) {
	v := make([]field.Element, len(vec))
	for _, _v := range vec {
		v = append(v, _v...)
	}
	fs.UpdateElmts(v...)
}

func (fs *FS) RandomField() field.Octuplet {
	r := fs.RandomFieldFrElmt() // the safeguard update is called
	res := encoding.EncodeFrElementToOctuplet(r)
	return res
}

func (fs *FS) RandomManyIntegers(n int) []field.Octuplet {
	res := make([]field.Octuplet, n)
	for i := 0; i < n; i++ {
		r := fs.RandomFieldFrElmt()
		res[i] = encoding.EncodeFrElementToOctuplet(r)
	}
	return res
}

// func RandomFext(h *poseidon2_bls12377.MDHasher) fext.Element {
// 	s := h.SumElement()
// 	var res fext.Element
// 	res.B0.A0 = s[0]
// 	res.B0.A1 = s[1]
// 	res.B1.A0 = s[2]
// 	res.B1.A1 = s[3]

// 	UpdateExt(h, fext.NewFromUint(0, 0, 0, 0)) // safefuard update
// 	return res
// }

// func RandomManyIntegers(h *poseidon2_bls12377.MDHasher, num, upperBound int) []int {

// 	// Even `1` would be wierd, there would be only one acceptable coin value.
// 	if upperBound < 1 {
// 		utils.Panic("UpperBound was %v", upperBound)
// 	}

// 	if !utils.IsPowerOfTwo(upperBound) {
// 		utils.Panic("Expected a power of two but got %v", upperBound)
// 	}

// 	if num == 0 {
// 		return []int{}
// 	}

// 	defer safeguardUpdate(h)

// 	var (
// 		// challsBitSize stores the number of bits required instantiate each
// 		// small integer.
// 		challsBitSize = math.Ceil(math.Log2(float64(upperBound)))
// 		// Number of challenges computable with one call to hash (implicitly,
// 		// the division is rounded down). The "-1" corresponds to the fact that
// 		// the most significant bit of a field element cannot be assumed to
// 		// contain exactly 1 bit of entropy since the modulus of the field is
// 		// never a power of 2.
// 		maxNumChallsPerDigest = (field.Bits - 1) / int(challsBitSize)
// 		// res stores the preallocated result slice, to which all generated
// 		// small integers will be appended.
// 		res = make([]int, 0, num)
// 	)

// 	for {
// 		digest := h.Sum(nil)
// 		buffer := NewBitReader(digest[:], field.Bits-1)

// 		// Increase the counter

// 		for i := 0; i < maxNumChallsPerDigest; i++ {
// 			// Stopping condition, we computed enough challenges
// 			if len(res) >= num {
// 				return res
// 			}

// 			newChall, err := buffer.ReadInt(int(challsBitSize))
// 			if err != nil {
// 				utils.Panic("could not instantiate the buffer for a single field element")
// 			}
// 			res = append(res, int(newChall)%upperBound)
// 		}

// 		// This is guarded by the condition to prevent the [State] updating
// 		// twice in a row when exiting the function. Recall that we have a
// 		// defer of the safeguard update in the function. This handles the
// 		// edge-case where the number of requested field elements is a multiple
// 		// of the number of challenges we can generate with a single field
// 		// element.
// 		if len(res) >= num {
// 			return res
// 		}

// 		// This updates ensures that for the next iterations of the loop and the
// 		// next randomness comsumption uses a fresh randomness.
// 		safeguardUpdate(h)
// 	}
// }

func (fs *FS) safeguardUpdate() {
	fs.UpdateFrElmt(fr.Element{})
}
