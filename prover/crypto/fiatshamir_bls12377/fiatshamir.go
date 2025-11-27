package fiatshamir_bls12377

import (
	"github.com/consensys/gnark-crypto/ecc/bls12-377/fr"
	"github.com/consensys/linea-monorepo/prover/crypto/encoding"
	"github.com/consensys/linea-monorepo/prover/crypto/poseidon2_bls12377"
	"github.com/consensys/linea-monorepo/prover/maths/common/smartvectors"
	"github.com/consensys/linea-monorepo/prover/maths/field"
	"github.com/consensys/linea-monorepo/prover/maths/field/fext"
	"github.com/consensys/linea-monorepo/prover/utils"
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

func (fs *FS) Update(vec ...field.Element) {
	// fmt.Printf("non-circuit: Updating FS with field elements %v \n", vector.Prettify(vec))
	fs.h.WriteKoalabearElements(vec...)
}

func (fs *FS) UpdateExt(vec ...fext.Element) {
	// fmt.Printf("non-circuit: Updating FS with fext elements %v \n", vectorext.Prettify(vec))

	for i := 0; i < len(vec); i++ {
		fs.h.WriteKoalabearElements(vec[i].B0.A0)
		fs.h.WriteKoalabearElements(vec[i].B0.A1)
		fs.h.WriteKoalabearElements(vec[i].B1.A0)
		fs.h.WriteKoalabearElements(vec[i].B1.A1)
	}
}

func (fs *FS) UpdateVec(vec ...[]field.Element) {
	v := make([]field.Element, len(vec))
	for _, _v := range vec {
		v = append(v, _v...)
	}
	fs.Update(v...)
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
	if smartvectors.IsBase(sv) {
		vec := make([]field.Element, sv.Len())
		sv.WriteInSlice(vec)
		fs.Update(vec...)
	} else {
		vec := make([]fext.Element, sv.Len())
		sv.WriteInSliceExt(vec)
		fs.UpdateExt(vec...)
	}
}

func (fs *FS) RandomField() field.Octuplet {
	r := fs.RandomFieldFrElmt() // the safeguard update is called
	res := encoding.EncodeFrElementToOctuplet(r)
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

func (fs *FS) RandomFext() fext.Element {
	s := fs.RandomField()
	var res fext.Element
	res.B0.A0 = s[0]
	res.B0.A1 = s[1]
	res.B1.A0 = s[2]
	res.B1.A1 = s[3]

	fs.safeguardUpdate()
	return res
}

func (fs *FS) SetState(s field.Octuplet) {
	state := encoding.EncodeKoalabearOctupletToFrElement(s)
	bState := state.Bytes()
	fs.h.SetState(bState[:])
}

func (fs *FS) State() field.Octuplet {
	bState := fs.h.State()
	var state fr.Element
	state.SetBytes(bState)
	return encoding.EncodeFrElementToOctuplet(state)
}

func (fs *FS) safeguardUpdate() {
	fs.UpdateFrElmt(fr.Element{})
}
