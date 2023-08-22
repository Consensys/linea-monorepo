package fiatshamir

import (
	"math"

	"github.com/consensys/accelerated-crypto-monorepo/crypto/mimc/gkrmimc"
	"github.com/consensys/accelerated-crypto-monorepo/maths/field"
	"github.com/consensys/accelerated-crypto-monorepo/utils"
	"github.com/consensys/gnark/frontend"
	"github.com/consensys/gnark/std/hash"
)

/*
GnarkFiatShamir mirrors `fiatshamire.State` in a gnark circuit.
*/
type GnarkFiatShamir struct {
	hasher hash.FieldHasher
	/*
		pointer to the gnark-API (also passed to the hasher but
		behind an interface)
	*/
	api frontend.API
}

/*
Creates a new fiat-shamir state that mirrors `fiatshamir.State`
in a gnark circuit.
*/
func NewGnarkFiatShamir(api frontend.API, factory *gkrmimc.HasherFactory) *GnarkFiatShamir {
	hasher := factory.NewHasher()

	return &GnarkFiatShamir{
		hasher: &hasher,
		api:    api,
	}
}

// Update the Fiat-Shamir state with a vector of field element
func (fs *GnarkFiatShamir) Update(vec ...frontend.Variable) {
	// Safeguard against nil
	for _, x := range vec {
		if x == nil {
			panic("gnark fiat-shamir updated with a nil frontend variable")
		}
	}
	fs.hasher.Write(vec...)
}

// Update the Fiat-Shamir state with a matrix of field element
func (fs *GnarkFiatShamir) UpdateVec(mat ...[]frontend.Variable) {
	for i := range mat {
		fs.Update(mat[i]...)
	}
}

// Returns a single valued fiat-shamir hash
func (fs *GnarkFiatShamir) RandomField() frontend.Variable {
	defer fs.safeguardUpdate()
	return fs.hasher.Sum()
}

// Returns a vector of variable that will contain small integers
func (fs *GnarkFiatShamir) RandomManyIntegers(num, upperBound int) []frontend.Variable {
	defer fs.safeguardUpdate()

	if !utils.IsPowerOfTwo(upperBound) {
		utils.Panic("expected a power of two but got %v", upperBound)
	}

	// Compute the number of bytes to generate the challenges in bits
	challsBitSize := int(math.Ceil(math.Log2(float64(upperBound))))

	/*
		The function is done in a way that minimizes the number
		of calls to the hash function. We try to extract as many
		"ranged integer challenges" for each digest

		Number of challenges computable with one call to hash
		(implicitly, round it down)
	*/
	maxNumChallsPerDigest := field.Bits / int(challsBitSize)

	res := make([]frontend.Variable, 0, num)

	challCount := 0

	for {
		digest := fs.hasher.Sum()
		digestBits := fs.api.ToBinary(digest)

		for i := 0; i < maxNumChallsPerDigest; i++ {
			// Stopping condition, we computed enough challenges
			if challCount >= num {
				return res
			}

			/*
				Drains the first `challsBitSize` of the digestBits into
				a new challenge to be returned.
			*/
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

// Update the stae as a safeguard. This way, if we query another
// vector of integers just after. We do not get the same result.
func (fs *GnarkFiatShamir) safeguardUpdate() {
	fs.Update(0)
}
