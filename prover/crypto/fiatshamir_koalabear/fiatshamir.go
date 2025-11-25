package fiatshamir_koalabear

import (
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
	n := utils.NextPowerOfTwo(upperBound)
	mask := n - 1
	i := 0
	res := make([]int, num)
	for i < num {
		// thake the remainder mod n of each limb
		c := fs.RandomField()
		for j := 0; j < 8; j++ {
			b := c[j].Bits()
			res[i] = int(b[0]) & mask
			i++
			fs.safeguardUpdate()
			if i >= num {
				break
			}
		}
	}
	return res
}

func (fs *FS) safeguardUpdate() {
	fs.Update(field.Zero())
}
