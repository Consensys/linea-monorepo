package poseidon2_bls12377

import (
	"fmt"

	"github.com/consensys/gnark/frontend"
	"github.com/consensys/gnark/std/permutation/poseidon2"
	"github.com/consensys/linea-monorepo/prover/crypto/encoding"
	"github.com/consensys/linea-monorepo/prover/maths/zk"
)

// GnarkMDHasher Merkle Damgard implementation using poseidon2 as compression function with width 16
// The hashing process goes as follow:
// newState := compress(old state, buf) where buf is an octuplet, left padded with zeroes if needed
type GnarkMDHasher struct {
	Api frontend.API

	// Sponge construction Sstate
	Sstate frontend.Variable

	// data to hash
	Bbuffer []frontend.Variable

	compressor *poseidon2.Permutation

	verbose bool
}

// NewGnarkMDHasher returns a new Octuplet
func NewGnarkMDHasher(api frontend.API, verbose ...bool) (GnarkMDHasher, error) {
	var res GnarkMDHasher
	res.Sstate = 0
	res.Api = api
	var err error

	if len(verbose) > 0 {
		res.verbose = verbose[0]
	}

	// default parameters
	res.compressor, err = poseidon2.NewPoseidon2FromParameters(api, 2, 6, 26)
	if err != nil {
		return res, err
	}
	return res, nil
}

func (h *GnarkMDHasher) Reset() {
	h.Bbuffer = h.Bbuffer[:0]
	h.Sstate = 0
}

func (h *GnarkMDHasher) Write(data ...frontend.Variable) {
	quo := (len(h.Bbuffer) + len(data)) / maxSizeBuf
	rem := (len(h.Bbuffer) + len(data)) % maxSizeBuf
	off := len(h.Bbuffer)

	for i := 0; i < quo; i++ {
		h.Bbuffer = append(h.Bbuffer, data[:maxSizeBuf-off]...)

		h.Api.Println("d.Buffer: \n", h.Bbuffer[:2])

		sum := h.Sum()
		h.Api.Println("After SumElement: \n", sum)
		h.Bbuffer = h.Bbuffer[:0] // flush the buffer once maxSizeBuf is reached
		off = len(h.Bbuffer)
		data = data[maxSizeBuf-off:] // Update data to the remaining part

	}

	h.Bbuffer = append(h.Bbuffer, data[:rem-off]...)
}

func (h *GnarkMDHasher) WriteWVs(data ...zk.WrappedVariable) {
	_data := encoding.EncodeWVsToFVs(h.Api, data)

	h.Bbuffer = append(h.Bbuffer, _data...)
}

func (h *GnarkMDHasher) SetState(state frontend.Variable) {
	h.Bbuffer = h.Bbuffer[:0]
	h.Sstate = state
}

func (h *GnarkMDHasher) State() frontend.Variable {
	// If the buffer is clean, we can short-path the execution and directly
	if len(h.Bbuffer) == 0 {
		return h.Sstate
	}

	// If the buffer is not clean, we cannot clean it locally as it would modify
	// the state of the hasher locally. Instead, we clone the buffer and flush
	// the buffer on the clone.
	clone, _ := NewGnarkMDHasher(h.Api)
	clone.Bbuffer = make([]frontend.Variable, len(h.Bbuffer))
	copy(clone.Bbuffer, h.Bbuffer)
	clone.Sstate = h.Sstate
	_ = clone.Sum()
	return clone.Sstate
}

func (h *GnarkMDHasher) Sum() frontend.Variable {

	fmt.Printf("len(h.Bbuffer)=%d\n", len(h.Bbuffer))
	if h.Bbuffer != nil {
		fmt.Printf("h.Sstate==nil=%v\n", h.compressor == nil)
		for i := 0; i < len(h.Bbuffer); i++ {
			h.Sstate = h.compressor.Compress(h.Sstate, h.Bbuffer[i])
		}
	}
	h.Bbuffer = h.Bbuffer[:0]

	return h.Sstate
}
