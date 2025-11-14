package poseidon2_bls12377

import (
	"github.com/consensys/gnark/frontend"
	"github.com/consensys/gnark/std/permutation/poseidon2"
)

// GnarkMDHasher Merkle Damgard implementation using poseidon2 as compression function with width 16
// The hashing process goes as follow:
// newState := compress(old state, buf) where buf is an octuplet, left padded with zeroes if needed
// type GnarkMDHasher hash.FieldHasher

// GnarkMDHasher Merkle Damgard implementation using poseidon2 as compression function with width 16
// The hashing process goes as follow:
// newState := compress(old state, buf) where buf is an octuplet, left padded with zeroes if needed
type GnarkMDHasher struct {
	// api frontend.API

	// Sponge construction state
	state frontend.Variable

	// data to hash
	buffer []frontend.Variable

	compressor *poseidon2.Permutation
}

// NewGnarkMDHasher returns a new Octuplet
func NewGnarkMDHasher(api frontend.API) (GnarkMDHasher, error) {
	var res GnarkMDHasher
	res.state = 0
	var err error

	// default parameters
	res.compressor, err = poseidon2.NewPoseidon2FromParameters(api, 2, 6, 26)
	if err != nil {
		return res, err
	}
	return res, nil
}

func (h *GnarkMDHasher) Reset() {
	h.buffer = h.buffer[:0]
	h.state = 0
}

func (h *GnarkMDHasher) Write(data ...frontend.Variable) {
	h.buffer = append(h.buffer, data...)
}

func (h *GnarkMDHasher) SetState(state frontend.Variable) {
	h.state = state
}

func (h *GnarkMDHasher) State() frontend.Variable {
	return h.state
}

func (h *GnarkMDHasher) Sum() frontend.Variable {
	for i := 0; i < len(h.buffer); i++ {
		h.state = h.compressor.Compress(h.state, h.buffer[i])
	}
	h.buffer = h.buffer[:0]
	return h.state
}

// NewGnarkMDHasher returns a new Octuplet
// func NewGnarkMDHasher(api frontend.API) (GnarkMDHasher, error) {
// 	f, err := poseidon2.NewPoseidon2(api)
// 	if err != nil {
// 		return nil, fmt.Errorf("could not create poseidon2 hasher: %w", err)
// 	}
// 	return hash.NewMerkleDamgardHasher(api, f, 0), nil
// }
