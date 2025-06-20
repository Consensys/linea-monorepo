package fiatshamir

import (
	"github.com/consensys/gnark-crypto/hash"
	"github.com/consensys/linea-monorepo/prover/maths/field"
	"github.com/consensys/linea-monorepo/prover/maths/field/fext"
	"github.com/consensys/linea-monorepo/prover/utils"
	"golang.org/x/crypto/blake2b"
)

// State holds a Fiat-Shamir state. The Fiat-Shamir state can be updated by
// providing field elements and can be consumed to generate either field
// elements or sequences of small integers. The Fiat-Shamir instantiation relies
// on the MiMC hash function and uses the following strategy for generating
// random public coins:
//
//   - The messages are appended to the hasher as it is received by the Fiat-
//     Shamir state, field element by field element
//   - When requested by the caller, the State can generates a field element by
//     hashing the transcript. Recall that we are using a SNARK-friendly hash
//     function and that its output should be interpreted as a field element.
//   - Everytime a field element has been generated, the hasher is updated by
//     appending an artificial message: field.Element(0). This is to ensure
//     that all generated field elements are independants.
//
// To be safely used, the Fiat-Shamir State should initialized by passing an
// initial field element summarizing the protocol in its full-extent. This is
// to prevent rogue protocol attack as in the Frozen Heart vulnerability.
//
// https://blog.trailofbits.com/2022/04/18/the-frozen-heart-vulnerability-in-plonk/

func Update(h hash.StateStorer, vec ...field.Element) {
	for _, f := range vec {
		bytes := f.Bytes()
		_, err := h.Write(bytes[:])
		if err != nil {
			panic("Hashing is not supposed to fail")
		}
	}
}

func UpdateExt(h hash.StateStorer, vec ...fext.Element) {
	for _, f := range vec {
		bytes := fext.Bytes(&f)
		_, err := h.Write(bytes[:])
		if err != nil {
			panic("Hashing is not supposed to fail")
		}
	}
}

func UpdateVec(h hash.StateStorer, vecs ...[]field.Element) {
	for i := range vecs {
		Update(h, vecs[i]...)
	}
}

// RandomField generates and returns a single field element from the Fiat-Shamir
// transcript.
func RandomField(h hash.StateStorer) field.Element {
	s := h.Sum(nil)
	var res field.Element
	res.SetBytes(s)
	UpdateExt(h, fext.NewElement(0, 0, 0, 0)) // ??
	return res
}

func RandomFext(h hash.StateStorer) fext.Element {
	// TODO @thomas according the size of s, run several hashes to fit in an fext elmt
	s := h.Sum(nil)
	var res fext.Element
	if len(s) > 4 {
		res = fext.SetBytes(s)
	} else {
		res.B0.A0.SetBytes(s)
	}
	UpdateExt(h, fext.NewElement(0, 0, 0, 0)) // ??
	return res
}

func RandomFromSeed(h hash.StateStorer, seed field.Element, name string) field.Element {

	bName := []byte(name)
	hasher, _ := blake2b.New256(nil) //?? TODO @Thomas use proper hash to field
	hasher.Write(bName)
	bName = hasher.Sum(nil)

	var bnToFr field.Element // reduction mod r...
	bnToFr.SetBytes(bName)
	bName = bnToFr.Marshal()

	// seed = compress(name || seed)
	backupState := h.State() // ??
	h.SetState(seed.Marshal())
	if _, err := h.Write(bName); err != nil {
		panic(err)
	}
	bRes := h.Sum(nil)
	var res field.Element
	res.SetBytes(bRes)

	err := h.SetState(backupState)
	if err != nil {
		panic(err)
	}

	return res
}

// TODO@yao: RandomFextFromSeed generates and returns a single fext element from the seed and the given name.
func RandomManyIntegers(h hash.StateStorer, num, upperBound int) []int {

	if upperBound < 1 {
		utils.Panic("UpperBound was %v", upperBound)
	}

	if !utils.IsPowerOfTwo(upperBound) {
		utils.Panic("Expected a power of two but got %v", upperBound)
	}

	if num == 0 {
		return []int{}
	}

	logTwoUpperBound := -1
	for upperBound != 0 {
		logTwoUpperBound++
		upperBound >>= 1
	}

	// number of field elmts per hash
	maxNumChallsPerDigest := (field.Bits - 1) / int(logTwoUpperBound)

	res := make([]int, 0, num)

	for {
		digest := h.Sum(nil)
		buffer := NewBitReader(digest, field.Bits-1)

		for i := 0; i < maxNumChallsPerDigest; i++ {
			// Stopping condition, we computed enough challenges
			if len(res) >= num {
				return res
			}

			newChall, err := buffer.ReadInt(int(logTwoUpperBound))
			if err != nil {
				utils.Panic("could not instantiate the buffer for a single field element")
			}
			res = append(res, int(newChall)%upperBound)
		}

		if len(res) >= num {
			return res
		}

		UpdateExt(h, fext.NewElement(0, 0, 0, 0)) // ??
	}
}
