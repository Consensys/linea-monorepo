package hashtypes

import (
	"hash"

	"github.com/consensys/accelerated-crypto-monorepo/maths/field"
	"github.com/consensys/gnark-crypto/ecc/bn254/fr/mimc"
	"github.com/ethereum/go-ethereum/crypto"
)

// Wrapper types for hasher which additionally provides a max value
type Hasher struct {
	hash.Hash        // the underlying hasher
	maxValue  Digest // the maximal value obtainable with that hasher
}

// Immutable accessor for the max value of the hasher
func (h Hasher) MaxDigest() Digest {
	return h.maxValue
}

// Create a Keccak hasher
func Keccak() Hasher {
	return Hasher{
		Hash: crypto.NewKeccakState(),
		maxValue: Digest{
			255, 255, 255, 255,
			255, 255, 255, 255,
			255, 255, 255, 255,
			255, 255, 255, 255,
			255, 255, 255, 255,
			255, 255, 255, 255,
			255, 255, 255, 255,
			255, 255, 255, 255,
		},
	}
}

// Create a new MiMC hasher
func MiMC() Hasher {
	maxVal := field.NewFromString("-1")
	return Hasher{
		Hash:     mimc.NewMiMC(),
		maxValue: maxVal.Bytes(),
	}
}
