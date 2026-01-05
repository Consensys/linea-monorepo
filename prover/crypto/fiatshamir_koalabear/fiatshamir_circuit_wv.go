package fiatshamir_koalabear

import (
	"math/bits"

	"github.com/consensys/gnark/frontend"
	"github.com/consensys/linea-monorepo/prover/crypto/poseidon2_koalabear"

	"github.com/consensys/linea-monorepo/prover/maths/field/gnarkfext"
	"github.com/consensys/linea-monorepo/prover/maths/zk"
	"github.com/consensys/linea-monorepo/prover/utils"
)

type GnarkFSWV struct {
	hasher poseidon2_koalabear.GnarkMDHasher
	// pointer to the gnark-API (also passed to the hasher but behind an
	// interface). This is needed to perform bit-decomposition.
	api frontend.API
}

func NewGnarkFSWV(api frontend.API) *GnarkFSWV {
	hasher, _ := poseidon2_koalabear.NewGnarkMDHasher(api)
	return &GnarkFSWV{
		hasher: hasher,
		api:    api,
	}
}

// Update updates the Fiat-Shamir state with a vector of frontend.Variable
// representing field element each.
func (fs *GnarkFSWV) Update(vec ...zk.WrappedVariable) {
	_vec := wvTofv(vec)
	fs.hasher.Write(_vec...)
}
func (fs *GnarkFSWV) UpdateExt(vec ...gnarkfext.E4Gen) {
	for i := 0; i < len(vec); i++ {
		fs.hasher.Write(vec[i].B0.A0.AsNative())
		fs.hasher.Write(vec[i].B0.A1.AsNative())
		fs.hasher.Write(vec[i].B1.A0.AsNative())
		fs.hasher.Write(vec[i].B1.A1.AsNative())
	}
}

func wvTofv(v []zk.WrappedVariable) []frontend.Variable {
	buf := make([]frontend.Variable, len(v))
	for i := 0; i < len(v); i++ {
		buf[i] = v[i].AsNative()
	}
	return buf
}

func octupletToZkoctuplet(v poseidon2_koalabear.Octuplet) zk.Octuplet {
	var res zk.Octuplet
	for i := 0; i < 8; i++ {
		res[i] = zk.WrapFrontendVariable(v[i])
	}
	return res
}

func zkoctupletTooctuplet(v zk.Octuplet) poseidon2_koalabear.Octuplet {
	var res poseidon2_koalabear.Octuplet
	for i := 0; i < 8; i++ {
		res[i] = v[i].AsNative()
	}
	return res
}

// UpdateVec updates the Fiat-Shamir state with a matrix of field element.
func (fs *GnarkFSWV) UpdateVec(mat ...[]zk.WrappedVariable) {
	for i := range mat {
		fs.Update(mat[i]...)
	}
}

func (fs *GnarkFSWV) RandomField() zk.Octuplet {
	res := fs.hasher.Sum()
	fs.safeguardUpdate()
	return octupletToZkoctuplet(res)
}

func (fs *GnarkFSWV) randomFieldNative() poseidon2_koalabear.Octuplet {
	res := fs.hasher.Sum()
	fs.safeguardUpdate()
	return res
}

// RandomField returns a single valued fiat-shamir hash
func (fs *GnarkFSWV) RandomFieldExt() gnarkfext.E4Gen {
	defer fs.safeguardUpdate()

	s := fs.RandomField()
	var res gnarkfext.E4Gen
	res.B0.A0 = s[0]
	res.B0.A1 = s[1]
	res.B1.A0 = s[2]
	res.B1.A1 = s[3]

	return res
}

func (fs *GnarkFSWV) RandomManyIntegers(num, upperBound int) []frontend.Variable {
	n := utils.NextPowerOfTwo(upperBound)
	nbBits := bits.TrailingZeros(uint(n))
	i := 0
	res := make([]frontend.Variable, num)
	for i < num {
		// thake the remainder mod n of each limb
		c := fs.randomFieldNative()
		for j := 0; j < 8; j++ {
			b := fs.api.ToBinary(c[j])
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

// SetState mutates the fiat-shamir state of
func (fs *GnarkFSWV) SetState(state zk.Octuplet) {
	fs.hasher.Reset()
	_state := zkoctupletTooctuplet(state)
	fs.hasher.SetState(_state)
}

// State mutates returns the state of the fiat-shamir hasher. The
// function will also updates its own state with unprocessed inputs.
func (fs *GnarkFSWV) State() zk.Octuplet {
	return octupletToZkoctuplet(fs.hasher.State())
}

// safeguardUpdate updates the state as a safeguard by appending a field element
// representing a "0". This is used every time a field element is consumed from
// the hasher to ensure that the next field element will have a different
// value.
func (fs *GnarkFSWV) safeguardUpdate() {
	fs.hasher.Write(0)
}
