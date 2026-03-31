package fiatshamir_koalabear

import (
	"unsafe"

	"github.com/consensys/linea-monorepo/prover/crypto/poseidon2_koalabear"
	"github.com/consensys/linea-monorepo/prover/maths/koalabear/field"
	"github.com/consensys/linea-monorepo/prover/utils"
	"github.com/consensys/linea-monorepo/prover/utils/types"
	"golang.org/x/crypto/blake2b"
)

// https://blog.trailofbits.com/2022/04/18/the-frozen-heart-vulnerability-in-plonk/

type FS struct {
	h *poseidon2_koalabear.MDHasher
}

func NewFS() *FS {
	return &FS{h: poseidon2_koalabear.NewMDHasher()}
}

func (fs *FS) Update(vec ...field.Element) {
	fs.h.WriteElements(vec...)
}

func (fs *FS) UpdateExt(vec ...field.Ext) {
	if len(vec) == 0 {
		return
	}
	vElems := unsafe.Slice((*field.Element)(unsafe.Pointer(&vec[0])), 4*len(vec))
	fs.h.WriteElements(vElems...)
}

func (fs *FS) UpdateGeneric(vec ...field.FieldElem) {
	if len(vec) == 0 {
		return
	}

	// Marshal the elements in a vector of bytes
	for _, f := range vec {
		if f.IsBase() {
			fs.Update(f.AsBase())
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
func (fs *FS) UpdateVecExt(vecs ...[]field.Ext) {
	for i := range vecs {
		fs.UpdateExt(vecs[i]...)
	}
}

// UpdateSV updates the FS state with a smart-vector. No-op if the smart-vector
// has a length of zero.
func (fs *FS) UpdateSV(sv field.FieldVec) {
	if sv.Len() == 0 {
		return
	}

	if sv.IsBase() {
		vec := sv.AsBase()
		fs.Update(vec...)
	} else {
		vec := sv.AsExt()
		fs.UpdateExt(vec...)
	}
}

func (fs *FS) RandomField() field.Octuplet {
	defer fs.safeguardUpdate()
	return fs.h.SumElement()
}

func (fs *FS) RandomFext() field.Ext {
	s := fs.RandomField() // already calls safeguardUpdate()
	var res field.Ext
	res.B0.A0 = s[0]
	res.B0.A1 = s[1]
	res.B1.A0 = s[2]
	res.B1.A1 = s[3]

	return res
}

func (fs *FS) RandomFieldFromSeed(seed field.Octuplet, name string) field.Ext {

	// The first step encodes the 'name' into a single field element. The
	// field element is obtained by hashing and taking the modulo of the
	// result to fit into a field element.
	nameBytes := []byte(name)
	hasher, _ := blake2b.New256(nil)
	hasher.Write(nameBytes)
	nameBytes = hasher.Sum(nil)
	nameOctuplet := types.BytesToKoalaOctupletLoose(nameBytes)

	// The seed is then obtained by calling the compression function over
	// the seed and the encoded name.
	oldState := fs.State()
	defer fs.SetState(oldState)

	fs.SetState(seed)
	fs.h.WriteElements(nameOctuplet[:]...)
	res := fs.RandomFext()
	return res
}

func (fs *FS) RandomManyIntegers(num, upperBound int) []int {
	n := utils.NextPowerOfTwo(upperBound)
	mask := n - 1
	i := 0
	res := make([]int, num)
	for i < num {
		// take the remainder mod n of each limb
		c := fs.RandomField() // already calls safeguardUpdate()
		for j := 0; j < 8; j++ {
			b := c[j].Bits()
			res[i] = int(b[0]) & mask
			i++
			if i >= num {
				break
			}
		}
	}
	return res
}

func (fs *FS) SetState(s field.Octuplet) {
	fs.h.SetStateOctuplet(s)
}

func (fs *FS) State() field.Octuplet {
	return fs.h.GetStateOctuplet()
}

func (fs *FS) safeguardUpdate() {
	fs.Update(field.Zero())
}
