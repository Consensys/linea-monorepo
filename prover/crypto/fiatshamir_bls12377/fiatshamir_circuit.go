package fiatshamir_bls12377

import (
	"math/bits"

	"github.com/consensys/gnark/frontend"
	"github.com/consensys/linea-monorepo/prover/crypto/encoding"
	"github.com/consensys/linea-monorepo/prover/crypto/poseidon2_bls12377"
	"github.com/consensys/linea-monorepo/prover/utils"

	"github.com/consensys/linea-monorepo/prover/maths/field/fext"
	"github.com/consensys/linea-monorepo/prover/maths/field/gnarkfext"
	"github.com/consensys/linea-monorepo/prover/maths/zk"
)

type GnarkFS struct {
	hasher poseidon2_bls12377.GnarkMDHasher
	// pointer to the gnark-API (also passed to the hasher but behind an
	// interface). This is needed to perform bit-decomposition.
	api frontend.API
}

func NewGnarkFS(api frontend.API) *GnarkFS {
	hasher, _ := poseidon2_bls12377.NewGnarkMDHasher(api)
	return &GnarkFS{
		hasher: hasher,
		api:    api,
	}
}

// SetState mutates the fiat-shamir state of
func (fs *GnarkFS) SetStateFrElmt(state frontend.Variable) {
	fs.hasher.Reset()
	fs.hasher.SetState(state)
}

// State mutates returns the state of the fiat-shamir hasher. The
// function will also updates its own state with unprocessed inputs.
func (fs *GnarkFS) StateFrElmt() frontend.Variable {
	return fs.hasher.State()
}

// Update updates the Fiat-Shamir state with a vector of frontend.Variable
// representing field element each.
func (fs *GnarkFS) UpdateFrElmt(vec ...frontend.Variable) {
	fs.hasher.Write(vec...)
}

// UpdateVec updates the Fiat-Shamir state with a matrix of field element.
func (fs *GnarkFS) UpdateVecFrElmt(mat ...[]frontend.Variable) {
	for i := range mat {
		fs.UpdateFrElmt(mat[i]...)
	}
}

// RandomField returns a single valued fiat-shamir hash
func (fs *GnarkFS) RandomFrElmt() frontend.Variable {
	defer fs.safeguardUpdate()
	return fs.hasher.Sum()
}

// ------------------------------------------------------
// List of methods to updae the FS state with koala elmts

func (fs *GnarkFS) Update(vec ...zk.WrappedVariable) {

	fs.hasher.WriteWVs(vec...)
}
func (fs *GnarkFS) UpdateExt(vec ...gnarkfext.E4Gen) {
	// ext4, _ := gnarkfext.NewExt4(fs.api)
	// ext4.Println(vec...) // DEBUG
	for i := 0; i < len(vec); i++ {
		fs.hasher.WriteWVs(vec[i].B0.A0)
		fs.hasher.WriteWVs(vec[i].B0.A1)
		fs.hasher.WriteWVs(vec[i].B1.A0)
		fs.hasher.WriteWVs(vec[i].B1.A1)
	}
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
func (fs *GnarkFS) UpdateVec(vec ...[]zk.WrappedVariable) {
	v := make([]zk.WrappedVariable, 0, len(vec))

	for _, _v := range vec {
		v = append(v, _v...)
	}
	// apiGen, _ := zk.NewGenericApi(fs.api)
	// apiGen.Println(vec[0]...) // DEBUG

	fs.Update(v...)
}

func (fs *GnarkFS) RandomField() zk.Octuplet {
	r := fs.RandomFrElmt() // the safeguard update is called
	res := encoding.EncodeFVTo8WVs(fs.api, r)
	return res
}
func (fs *GnarkFS) RandomFieldExt() gnarkfext.E4Gen {
	r := fs.RandomField() // the safeguard update is called
	res := gnarkfext.E4Gen{}
	res.B0.A0 = r[0]
	res.B0.A1 = r[1]
	res.B1.A0 = r[2]
	res.B1.A1 = r[3]
	return res
}
func (fs *GnarkFS) RandomManyIntegers(num, upperBound int) []frontend.Variable {
	apiGen, err := zk.NewGenericApi(fs.api)
	if err != nil {
		panic(err)
	}
	n := utils.NextPowerOfTwo(upperBound)
	nbBits := bits.TrailingZeros(uint(n))
	i := 0
	res := make([]frontend.Variable, num)
	for i < num {
		// thake the remainder mod n of each limb
		c := fs.RandomField()
		for j := 0; j < 8; j++ {
			b := apiGen.ToBinary(c[j])
			res[i] = fs.api.FromBinary(b[:nbBits]...)
			i++
			fs.safeguardUpdate()
			if i >= num {
				break
			}
		}
	}
	return res
}

func (fs *GnarkFS) SetState(s zk.Octuplet) {
	state := encoding.Encode8WVsToFV(fs.api, s)
	fs.hasher.SetState(state)
}

func (fs *GnarkFS) State() zk.Octuplet {
	state := fs.hasher.State()
	return encoding.EncodeFVTo8WVs(fs.api, state)
}

// safeguardUpdate updates the state as a safeguard by appending a field element
// representing a "0". This is used every time a field element is consumed from
// the hasher to ensure that the next field element will have a different
// value.
func (fs *GnarkFS) safeguardUpdate() {
	fs.hasher.Write(0)
}
