package fiatshamir

import (
	"hash"
	"math"

	"github.com/consensys/accelerated-crypto-monorepo/crypto/mimc"
	"github.com/consensys/accelerated-crypto-monorepo/maths/common/smartvectors"
	"github.com/consensys/accelerated-crypto-monorepo/maths/common/vector"
	"github.com/consensys/accelerated-crypto-monorepo/maths/field"
	"github.com/consensys/accelerated-crypto-monorepo/utils"
	"github.com/consensys/accelerated-crypto-monorepo/utils/profiling"
)

// Holds a fiat-shamir state
type State struct {
	hasher           hash.Hash
	TranscriptSize   int
	NumCoinGenerated int
}

// Construct a fresh FS state
func NewMiMCFiatShamir() *State {
	return &State{
		hasher: mimc.NewMiMC(),
	}
}

// Update the FS state with a smart-vector
func (fs *State) UpdateSV(sv smartvectors.SmartVector) {
	vec := make([]field.Element, sv.Len())
	sv.WriteInSlice(vec)
	fs.Update(vec...)
}

// Update the Fiat-Shamir state with a vector of field element
func (fs *State) Update(vec ...field.Element) {

	// Marshal the elements in a vector of bytes
	bytes := vector.Marshal(vec)

	_, err := fs.hasher.Write(bytes)
	if err != nil {
		panic("Hashing is not supposed to fail")
	}

	// Increase the transcript counter
	fs.TranscriptSize += len(vec)
}

// Update the Fiat-Shamir state with a matrix of field element
func (fs *State) UpdateVec(mat ...[]field.Element) {
	for i := range mat {
		fs.Update(mat[i]...)
	}
}

// Return one field element
func (fs *State) RandomField() field.Element {
	defer fs.safeguardUpdate()
	challBytes := fs.hasher.Sum(nil)
	var res field.Element
	res.SetBytes(challBytes)

	// increase the counter by one
	fs.NumCoinGenerated++
	return res
}

/*
Returns as a challenge, a list of positive integers
in the given range. The upperbound is strict and it
is restricted to powers of two
*/
func (fs *State) RandomManyIntegers(num, upperBound int) []int {

	stoptimer := profiling.LogTimer("FS - random bounded integers %v (bound %v)", num, upperBound)
	defer stoptimer()

	// Even `1` would be wierd, there would be only one acceptable coin value.
	if upperBound < 1 {
		utils.Panic("upperBound was %v", upperBound)
	}

	defer fs.safeguardUpdate()

	if !utils.IsPowerOfTwo(upperBound) {
		utils.Panic("expected a power of two but got %v", upperBound)
	}

	// Compute the number of bytes to generate the challenges in bits
	challsBitSize := math.Ceil(math.Log2(float64(upperBound)))

	// The function is done in a way that minimizes the number
	// of calls to the hash function. We try to extract as many
	// "ranged integer challenges" for each digest

	// Number of challenges computable with one call to hash
	// (implicitly, round it down)
	maxNumChallsPerDigest := field.Bits / int(challsBitSize)

	res := make([]int, 0, num)

	challCount := 0

	for {
		digest := fs.hasher.Sum(nil)
		buffer := NewBitReader(digest, field.Bits)

		// Increase the counter
		fs.NumCoinGenerated++

		for i := 0; i < maxNumChallsPerDigest; i++ {
			// Stopping condition, we computed enough challenges
			if challCount >= num {
				return res
			}

			newChall := buffer.ReadInt(int(challsBitSize))
			res = append(res, int(newChall)%upperBound)
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
func (fs *State) safeguardUpdate() {
	fs.Update(field.NewElement(0))
}
