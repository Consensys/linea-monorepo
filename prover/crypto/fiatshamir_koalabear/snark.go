package fiatshamir_koalabear

import (
	"math"

	"github.com/consensys/gnark/frontend"
	"github.com/consensys/linea-monorepo/prover/crypto/poseidon2_koalabear"

	"github.com/consensys/linea-monorepo/prover/maths/field"
	"github.com/consensys/linea-monorepo/prover/maths/field/gnarkfext"
	"github.com/consensys/linea-monorepo/prover/maths/zk"
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

// Update updates the Fiat-Shamir state with a vector of frontend.Variable
// representing field element each.
func (fs *GnarkFS) Update(vec ...frontend.Variable) {
	fs.hasher.Write(vec...)
}
func (fs *GnarkFS) UpdateExt(vec ...gnarkfext.E4Gen) {
	for i := 0; i < len(vec); i++ {
		fs.hasher.Write(vec[i].B0.A0.AsNative())
		fs.hasher.Write(vec[i].B0.A1.AsNative())
		fs.hasher.Write(vec[i].B1.A0.AsNative())
		fs.hasher.Write(vec[i].B1.A1.AsNative())
	}
}

// UpdateVec updates the Fiat-Shamir state with a matrix of field element.
func (fs *GnarkFS) UpdateVec(mat ...[]frontend.Variable) {
	for i := range mat {
		fs.Update(mat[i]...)
	}
}

// RandomField returns a single valued fiat-shamir hash
func (fs *GnarkFS) RandomFieldExt() gnarkfext.E4Gen {
	defer fs.safeguardUpdate()

	s := fs.hasher.Sum()
	var res gnarkfext.E4Gen
	res.B0.A0 = zk.WrapFrontendVariable(s[0])
	res.B0.A1 = zk.WrapFrontendVariable(s[1])
	res.B1.A0 = zk.WrapFrontendVariable(s[2])
	res.B1.A1 = zk.WrapFrontendVariable(s[3])
	return res
}

// RandomManyIntegers returns a vector of variable that will contain small integers
func (fs *GnarkFS) RandomManyIntegers(num, upperBound int) []frontend.Variable {

	// Even `1` would be wierd, there would be only one acceptable coin value.
	if upperBound < 1 {
		utils.Panic("UpperBound was %v", upperBound)
	}

	if !utils.IsPowerOfTwo(upperBound) {
		utils.Panic("Expected a power of two but got %v", upperBound)
	}

	if num == 0 {
		return []frontend.Variable{}
	}

	defer fs.safeguardUpdate()

	var (
		// Compute the number of bytes to generate the challenges in bits
		challsBitSize = int(math.Ceil(math.Log2(float64(upperBound))))
		// Number of challenges computable with one call to hash (implicitly, the
		// the division is rounded-down)
		maxNumChallsPerDigest = (field.Bits - 1) / int(challsBitSize)
		// res stores the function result
		res = make([]frontend.Variable, 0, num)
		// challCount stores the number of generated small integers
		challCount = 0
	)

	for {
		digest := fs.hasher.Sum()
		digestBits := fs.api.ToBinary(digest)

		for i := 0; i < maxNumChallsPerDigest; i++ {
			// Stopping condition, we computed enough challenges
			if challCount >= num {
				return res
			}

			// Drains the first `challsBitSize` of the digestBits into
			// a new challenge to be returned.
			newChall := fs.api.FromBinary(digestBits[:challsBitSize]...)
			digestBits = digestBits[challsBitSize:]
			res = append(res, newChall)
			challCount++
		}

		/*
			This test prevents reupdating fs just before returning
			and thus, safeguardupdating once more because of the defer.
		*/
		if challCount >= num {
			return res
		}

		fs.safeguardUpdate()
	}
}

// safeguardUpdate updates the state as a safeguard by appending a field element
// representing a "0". This is used every time a field element is consumed from
// the hasher to ensure that the next field element will have a different
// value.
func (fs *GnarkFS) safeguardUpdate() {
	fs.hasher.Write(0)
}
