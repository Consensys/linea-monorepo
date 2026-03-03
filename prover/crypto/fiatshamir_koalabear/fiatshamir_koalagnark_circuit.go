package fiatshamir_koalabear

import (
	"math/bits"

	"github.com/consensys/gnark/frontend"
	"github.com/consensys/linea-monorepo/prover/crypto/poseidon2_koalabear"
	"github.com/consensys/linea-monorepo/prover/maths/field/koalagnark"
	"github.com/consensys/linea-monorepo/prover/utils"
	"github.com/consensys/linea-monorepo/prover/utils/types"
	"golang.org/x/crypto/blake2b"
)

// GnarkFSKoalagnark is a Fiat-Shamir implementation that uses
// KoalagnarkMDHasher (koalagnark.API-based) instead of GnarkMDHasher
// (frontend.Variable-based). This correctly handles both native KoalaBear
// circuits and emulated circuits (e.g., BN254) where KoalaBear arithmetic
// must be emulated rather than performed natively.
type GnarkFSKoalagnark struct {
	hasher   *poseidon2_koalabear.KoalagnarkMDHasher
	api      frontend.API
	koalaAPI *koalagnark.API
}

// NewGnarkFSKoalagnark creates a Fiat-Shamir instance using KoalagnarkMDHasher.
// This works in both native KoalaBear circuits and emulated circuits.
func NewGnarkFSKoalagnark(api frontend.API) *GnarkFSKoalagnark {
	hasher := poseidon2_koalabear.NewKoalagnarkMDHasher(api)
	return &GnarkFSKoalagnark{
		hasher:   hasher,
		api:      api,
		koalaAPI: koalagnark.NewAPI(api),
	}
}

// Update updates the Fiat-Shamir state with koalagnark.Element values.
func (fs *GnarkFSKoalagnark) Update(vec ...koalagnark.Element) {
	fs.hasher.Write(vec...)
}

// UpdateExt updates the Fiat-Shamir state with extension field elements.
func (fs *GnarkFSKoalagnark) UpdateExt(vec ...koalagnark.Ext) {
	for i := 0; i < len(vec); i++ {
		fs.hasher.Write(vec[i].B0.A0)
		fs.hasher.Write(vec[i].B0.A1)
		fs.hasher.Write(vec[i].B1.A0)
		fs.hasher.Write(vec[i].B1.A1)
	}
}

// UpdateVec updates the Fiat-Shamir state with a matrix of field elements.
func (fs *GnarkFSKoalagnark) UpdateVec(mat ...[]koalagnark.Element) {
	for i := range mat {
		fs.Update(mat[i]...)
	}
}

// RandomField returns an octuplet of random field elements from the FS state.
func (fs *GnarkFSKoalagnark) RandomField() koalagnark.Octuplet {
	res := fs.hasher.Sum()
	defer fs.safeguardUpdate()
	return koalagnarkOctupletToZkOctuplet(res)
}

// randomFieldNative returns the raw KoalagnarkOctuplet for internal use.
func (fs *GnarkFSKoalagnark) randomFieldNative() poseidon2_koalabear.KoalagnarkOctuplet {
	defer fs.safeguardUpdate()
	return fs.hasher.Sum()
}

// RandomFieldExt returns a single extension field element from the FS state.
func (fs *GnarkFSKoalagnark) RandomFieldExt() koalagnark.Ext {
	s := fs.RandomField() // already calls safeguardUpdate()
	var res koalagnark.Ext
	res.B0.A0 = s[0]
	res.B0.A1 = s[1]
	res.B1.A0 = s[2]
	res.B1.A1 = s[3]
	return res
}

// RandomManyIntegers returns num random integers in [0, upperBound).
// The integers are produced from the FS state by bit-decomposition.
func (fs *GnarkFSKoalagnark) RandomManyIntegers(num, upperBound int) []frontend.Variable {
	n := utils.NextPowerOfTwo(upperBound)
	nbBits := bits.TrailingZeros(uint(n))
	i := 0
	res := make([]frontend.Variable, num)
	for i < num {
		c := fs.randomFieldNative() // already calls safeguardUpdate()
		for j := 0; j < 8; j++ {
			// Use koalaAPI.ToBinary for correct bit decomposition in emulated mode
			b := fs.koalaAPI.ToBinary(koalagnarkElemToZkElem(c[j]))
			res[i] = fs.api.FromBinary(b[:nbBits]...)
			i++
			if i >= num {
				break
			}
		}
	}
	return res
}

// RandomFieldFromSeed returns a random extension field element derived from
// a seed and a name string.
func (fs *GnarkFSKoalagnark) RandomFieldFromSeed(seed koalagnark.Octuplet, name string) koalagnark.Ext {
	// Encode the name into a single octuplet using blake2b
	nameBytes := []byte(name)
	hasher, _ := blake2b.New256(nil)
	hasher.Write(nameBytes)
	nameBytes = hasher.Sum(nil)
	nameOctuplet := types.BytesToKoalaOctupletLoose(nameBytes)
	var nameKoalaOctuplet [8]koalagnark.Element
	for i := 0; i < 8; i++ {
		nameKoalaOctuplet[i] = koalagnark.NewElementFromKoala(nameOctuplet[i])
	}

	// Save/restore state around seed-based derivation
	oldState := fs.State()
	defer fs.SetState(oldState)

	fs.SetState(seed)
	fs.hasher.Write(nameKoalaOctuplet[:]...)

	return fs.RandomFieldExt()
}

// SetState sets the Fiat-Shamir state.
func (fs *GnarkFSKoalagnark) SetState(state koalagnark.Octuplet) {
	fs.hasher.Reset()
	_state := zkOctupletToKoalagnarkOctuplet(state)
	fs.hasher.SetState(_state)
}

// State returns the current Fiat-Shamir state.
func (fs *GnarkFSKoalagnark) State() koalagnark.Octuplet {
	return koalagnarkOctupletToZkOctuplet(fs.hasher.State())
}

// safeguardUpdate appends a zero element to the FS state after each consumption
// to ensure the next generated value differs.
func (fs *GnarkFSKoalagnark) safeguardUpdate() {
	fs.hasher.Write(fs.koalaAPI.Zero())
}

// --- Conversion helpers ---

func koalagnarkOctupletToZkOctuplet(v poseidon2_koalabear.KoalagnarkOctuplet) koalagnark.Octuplet {
	var res koalagnark.Octuplet
	for i := 0; i < 8; i++ {
		res[i] = v[i]
	}
	return res
}

func zkOctupletToKoalagnarkOctuplet(v koalagnark.Octuplet) poseidon2_koalabear.KoalagnarkOctuplet {
	var res poseidon2_koalabear.KoalagnarkOctuplet
	for i := 0; i < 8; i++ {
		res[i] = v[i]
	}
	return res
}

func koalagnarkElemToZkElem(v koalagnark.Element) koalagnark.Element {
	return v
}
