package fiatshamir

import (
	"github.com/consensys/gnark-crypto/hash"
	"github.com/consensys/linea-monorepo/prover/maths/common/smartvectors"
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
			panic(err)
		}
	}
}

func UpdateExt(h hash.StateStorer, vec ...fext.Element) {
	for _, f := range vec {
		bytes := fext.Bytes(&f)
		_, err := h.Write(bytes[:])
		if err != nil {
			panic(err)
		}
	}
}

func UpdateVec(h hash.StateStorer, vecs ...[]field.Element) {
	for i := range vecs {
		Update(h, vecs[i]...)
	}
}

// UpdateVec updates the Fiat-Shamir state by passing one of more slices of
// field elements.
func UpdateVecExt(h hash.StateStorer, vecs ...[]fext.Element) {
	for i := range vecs {
		UpdateExt(h, vecs[i]...)
	}
}

// UpdateSV updates the FS state with a smart-vector. No-op if the smart-vector
// has a length of zero.
func UpdateSV(h hash.StateStorer, sv smartvectors.SmartVector) {
	if sv.Len() == 0 {
		return
	}

	vec := make([]field.Element, sv.Len())
	sv.WriteInSlice(vec)
	Update(h, vec...)
}

// RandomField generates and returns a single field element from the Fiat-Shamir
// transcript.
func RandomField(h hash.StateStorer) field.Element {
	s := h.Sum(nil)
	var res field.Element
	res.SetBytes(s)
	safeguardUpdate(h)
	return res
}

func RandomFieldFromSeed(h hash.StateStorer, seed field.Element, name string) field.Element {

	// map name to a field elmt
	tmpFr := field.Element{}
	nameBytes := []byte(name)
	hasher, _ := blake2b.New256(nil)
	hasher.Write(nameBytes)
	nameBytes = hasher.Sum(nil)

	// This ensures that the name is hashed into a field element
	tmpFr.SetBytes(nameBytes)
	nameBytes_ := tmpFr.Bytes()
	nameBytes = nameBytes_[:]

	// The seed is then obtained by calling the compression function over
	// the seed and the encoded name.
	oldState := h.State()
	defer h.SetState(oldState)

	h.SetState(seed.Marshal())
	if _, err := h.Write(nameBytes); err != nil {
		panic(err)
	}
	challBytes := h.Sum(nil)
	res := new(field.Element).SetBytes(challBytes)

	return *res
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
	UpdateExt(h, fext.NewElement(0, 0, 0, 0)) // safefuard update
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

	logTwoUpperBound := -1
	tmp := upperBound
	for tmp != 0 {
		logTwoUpperBound++
		tmp >>= 1
	}

	numChallengesPerDigest := (field.Bits - 1) / int(logTwoUpperBound)

	res := make([]int, num)
	for {
		digest := h.Sum(nil)
		buffer := NewBitReader(digest, field.Bits-1)

		for i := 0; i < numChallengesPerDigest; i++ {
			// Stopping condition, we computed enough challenges
			if len(res) >= num {
				return res
			}

			curChallenge, err := buffer.ReadInt(int(logTwoUpperBound))
			if err != nil {
				utils.Panic("could not instantiate the buffer for a single field element")
			}
			res = append(res, int(curChallenge)%upperBound)
		}

		// This is guarded by the condition to prevent the [State] updating
		// twice in a row when exiting the function. Recall that we have a
		// defer of the safeguard update in the function. This handles the
		// edge-case where the number of requested field elements is a multiple
		// of the number of challenges we can generate with a single field
		// element.
		if len(res) >= num {
			return res
		}

		// This updates ensures that for the next iterations of the loop and the
		// next randomness comsumption uses a fresh randomness.
		safeguardUpdate(h)
	}
}

func safeguardUpdate(h hash.StateStorer) {
	Update(h, field.NewElement(0))
}
