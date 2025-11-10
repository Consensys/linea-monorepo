package poseidon2_bls12377

import (
	"fmt"

	"github.com/consensys/gnark/frontend"
	"github.com/consensys/gnark/std/hash"
	"github.com/consensys/gnark/std/permutation/poseidon2"
)

// MerkleDamgardHasher Merkle Damgard implementation using poseidon2 as compression function with width 16
// The hashing process goes as follow:
// newState := compress(old state, buf) where buf is an octuplet, left padded with zeroes if needed
type MerkleDamgardHasher hash.FieldHasher

// NewMerkleDamgardHasher returns a new Octuplet
func NewMerkleDamgardHasher(api frontend.API) (MerkleDamgardHasher, error) {
	f, err := poseidon2.NewPoseidon2(api)
	if err != nil {
		return nil, fmt.Errorf("could not create poseidon2 hasher: %w", err)
	}
	return hash.NewMerkleDamgardHasher(api, f, 0), nil
}
