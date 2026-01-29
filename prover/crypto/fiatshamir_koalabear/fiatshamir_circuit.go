package fiatshamir_koalabear

import (
	"math/bits"

	"github.com/consensys/gnark/frontend"
	"github.com/consensys/linea-monorepo/prover/crypto/poseidon2_koalabear"
	"github.com/consensys/linea-monorepo/prover/maths/field/koalagnark"
	"github.com/consensys/linea-monorepo/prover/utils"
)

type GnarkFS struct {
	hasher poseidon2_koalabear.GnarkMDHasher
	// pointer to the gnark-API (also passed to the hasher but behind an
	// interface). This is needed to perform bit-decomposition.
	api frontend.API
}

func NewGnarkFS(api frontend.API) *GnarkFS {
	hasher, _ := poseidon2_koalabear.NewGnarkMDHasher(api)
	return &GnarkFS{
		hasher: hasher,
		api:    api,
	}
}

// Update updates the Fiat-Shamir state with a vector of frontend.Variable
// representing field element each.
func (fs *GnarkFS) Update(vec ...frontend.Variable) {
	fs.hasher.Write(vec...)
}

func (fs *GnarkFS) UpdateExt(vec ...koalagnark.Ext) {
	for i := 0; i < len(vec); i++ {
		fs.hasher.Write(vec[i].B0.A0.Native())
		fs.hasher.Write(vec[i].B0.A1.Native())
		fs.hasher.Write(vec[i].B1.A0.Native())
		fs.hasher.Write(vec[i].B1.A1.Native())
	}
}

// UpdateVec updates the Fiat-Shamir state with a matrix of field element.
func (fs *GnarkFS) UpdateVec(mat ...[]frontend.Variable) {
	for i := range mat {
		fs.Update(mat[i]...)
	}
}

func (fs *GnarkFS) RandomField() poseidon2_koalabear.Octuplet {
	defer fs.safeguardUpdate()
	return fs.hasher.Sum()
}

// RandomField returns a single valued fiat-shamir hash
func (fs *GnarkFS) RandomFieldExt() koalagnark.Ext {
	defer fs.safeguardUpdate()

	s := fs.hasher.Sum()
	var res koalagnark.Ext
	res.B0.A0 = koalagnark.WrapFrontendVariable(s[0])
	res.B0.A1 = koalagnark.WrapFrontendVariable(s[1])
	res.B1.A0 = koalagnark.WrapFrontendVariable(s[2])
	res.B1.A1 = koalagnark.WrapFrontendVariable(s[3])
	return res
}

func (fs *GnarkFS) RandomManyIntegers(num, upperBound int) []frontend.Variable {
	n := utils.NextPowerOfTwo(upperBound)
	nbBits := bits.TrailingZeros(uint(n))
	i := 0
	res := make([]frontend.Variable, num)
	for i < num {
		// thake the remainder mod n of each limb
		c := fs.RandomField() // already calls safeguardUpdate()
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

// SetState mutates the fiat-shamir state of
func (fs *GnarkFS) SetState(state poseidon2_koalabear.Octuplet) {
	fs.hasher.Reset()
	fs.hasher.SetState(state)
}

// State mutates returns the state of the fiat-shamir hasher. The
// function will also updates its own state with unprocessed inputs.
func (fs *GnarkFS) State() poseidon2_koalabear.Octuplet {
	return fs.hasher.State()
}

// safeguardUpdate updates the state as a safeguard by appending a field element
// representing a "0". This is used every time a field element is consumed from
// the hasher to ensure that the next field element will have a different
// value.
func (fs *GnarkFS) safeguardUpdate() {
	fs.hasher.Write(0)
}
