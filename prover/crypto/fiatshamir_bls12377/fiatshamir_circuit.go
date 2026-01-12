package fiatshamir_bls12377

import (
	"math/bits"

	"github.com/consensys/gnark/frontend"
	"github.com/consensys/linea-monorepo/prover/crypto/encoding"
	"github.com/consensys/linea-monorepo/prover/crypto/poseidon2_bls12377"
	"github.com/consensys/linea-monorepo/prover/crypto/poseidon2_koalabear"
	"github.com/consensys/linea-monorepo/prover/utils"

	"github.com/consensys/linea-monorepo/prover/maths/field/fext"
	"github.com/consensys/linea-monorepo/prover/maths/field/gnarkfext"
	"github.com/consensys/linea-monorepo/prover/maths/zk"
)

type GnarkFS struct {
	koalaBuf []frontend.Variable
	hasher   poseidon2_bls12377.GnarkMDHasher
	// pointer to the gnark-API (also passed to the hasher but behind an
	// interface). This is needed to perform bit-decomposition.
	api frontend.API
}

func NewGnarkFS(api frontend.API, verboseMode ...bool) *GnarkFS {
	hasher, err := poseidon2_bls12377.NewGnarkMDHasher(api, verboseMode...)
	if err != nil {
		panic(err)
	}
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
	fs.flushKoala()
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
	fs.flushKoala()
	defer fs.safeguardUpdate()
	return fs.hasher.Sum()
}

// ------------------------------------------------------
// List of methods to updae the FS state with koala elmts

func (fs *GnarkFS) Update(vec ...frontend.Variable) {
	// fs.hasher.WriteWVs(vec...)
	fs.koalaBuf = append(fs.koalaBuf, vec...)
}

func (fs *GnarkFS) UpdateExt(vec ...gnarkfext.E4Gen) {
	// ext4, _ := gnarkfext.NewExt4(fs.api)
	for i := 0; i < len(vec); i++ {
		fs.koalaBuf = append(fs.koalaBuf, vec[i].B0.A0.AsNative())
		fs.koalaBuf = append(fs.koalaBuf, vec[i].B0.A1.AsNative())
		fs.koalaBuf = append(fs.koalaBuf, vec[i].B1.A0.AsNative())
		fs.koalaBuf = append(fs.koalaBuf, vec[i].B1.A1.AsNative())
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

func (fs *GnarkFS) UpdateVec(vec ...[]frontend.Variable) {
	fs.Update(vec)
}

func (fs *GnarkFS) RandomField() poseidon2_koalabear.Octuplet {
	r := fs.RandomFrElmt() // the safeguard update is called
	res := encoding.EncodeFVTo8WVs(fs.api, r)
	return res
}

func (fs *GnarkFS) RandomFieldExt() gnarkfext.E4Gen {
	r := fs.RandomField() // the safeguard update is called
	res := gnarkfext.E4Gen{}
	res.B0.A0 = zk.WrapFrontendVariable(r[0])
	res.B0.A1 = zk.WrapFrontendVariable(r[1])
	res.B1.A0 = zk.WrapFrontendVariable(r[2])
	res.B1.A1 = zk.WrapFrontendVariable(r[3])
	return res
}
func (fs *GnarkFS) RandomManyIntegers(num, upperBound int) []frontend.Variable {

	n := utils.NextPowerOfTwo(upperBound)
	nbBits := bits.TrailingZeros(uint(n))
	i := 0
	res := make([]frontend.Variable, num)
	for i < num {
		// take the remainder mod n of each limb
		c := fs.RandomField() // already calls safeguardUpdate() once
		for j := 0; j < 8; j++ {
			b := fs.api.ToBinary(c[j])
			res[i] = fs.api.FromBinary(b[:nbBits]...)
			i++
			if i >= num {
				break
			}
		}
	}
	return res
}

func (fs *GnarkFS) SetState(s poseidon2_koalabear.Octuplet) {
	state := encoding.Encode8WVsToFV(fs.api, s)
	fs.hasher.SetState(state)
}

func (fs *GnarkFS) State() poseidon2_koalabear.Octuplet {
	state := fs.hasher.State()
	return encoding.EncodeFVTo8WVs(fs.api, state)
}

// safeguardUpdate updates the state as a safeguard by appending a field element
// representing a "0". This is used every time a field element is consumed from
// the hasher to ensure that the next field element will have a different
// value.
func (fs *GnarkFS) safeguardUpdate() {
	fs.UpdateFrElmt(0)
}

func (fs *GnarkFS) flushKoala() {
	if len(fs.koalaBuf) == 0 {
		return
	}
	fs.hasher.WriteKoala(fs.koalaBuf...)
	fs.koalaBuf = fs.koalaBuf[:0]
}
