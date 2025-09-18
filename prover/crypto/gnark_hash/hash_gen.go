// Package gnark_hash provides an interface that hash functions (as gadget) should implement.
// Copy from https://github.com/Consensys/gnark/blob/master/std/hash/hash.go
// TODO put this file in gnark ?

package gnark_hash

import (
	"github.com/consensys/gnark/frontend"
	"github.com/consensys/gnark/std/math/uints"
	"github.com/consensys/linea-monorepo/prover/protocol/zk"
)

type FieldHasher[T zk.Element] interface {
	Sum() T
	Write(data ...T)

	// Reset set the intermediate state to zero.
	Reset()
}

// StateStorer allows to store and retrieve the state of a hash function.
type StateStorer[T zk.Element] interface {
	FieldHasher[T]
	State() []T
	SetState(state []T) error
}

// BinaryHasher hashes inputs into a short digest. It takes as inputs bytes and
// outputs byte array whose length depends on the underlying hash function. For
// SNARK-native hash functions use [FieldHasher].
type BinaryHasher interface {
	Sum() []uints.U8

	Write([]uints.U8)

	// Size returns the number of bytes output by Sum.
	Size() int
}

type BinaryFixedLengthHasher[T zk.Element] interface {
	BinaryHasher
	// FixedLengthSum returns digest of the first length bytes. See the
	// [WithMinimalLength] option for setting lower bound on length.
	FixedLengthSum(length T) []uints.U8
}

type HasherConfig struct {
	MinimalLength int
}

// Option allows configuring the hash functions.
type Option func(*HasherConfig) error

// WithMinimalLength hints the minimal length of the input to the hash function.
// Default minimal length is 0.
func WithMinimalLength(minimalLength int) Option {
	return func(cfg *HasherConfig) error {
		cfg.MinimalLength = minimalLength
		return nil
	}
}

// 2 to 1 compression function
type Compressor[T zk.Element] interface {
	Compress(T, T) T
}

type merkleDamgardHasher[T zk.Element] struct {
	state T
	iv    T
	f     Compressor[T]
	api   zk.APIGen[T]
}

// NewMerkleDamgardHasher transforms a 2-1 one-way function into a hash
// initialState is a value whose preimage is not known
func NewMerkleDamgardHasher[T zk.Element](api frontend.API, f Compressor[T], initialState T) FieldHasher[T] {
	mApi, err := zk.NewApi[T](api)
	if err != nil { // TODO handle panic
		panic(err)
	}
	return &merkleDamgardHasher[T]{
		state: initialState,
		iv:    initialState,
		f:     f,
		api:   mApi,
	}
}

func (h *merkleDamgardHasher[T]) Reset() {
	h.state = h.iv
}

func (h *merkleDamgardHasher[T]) Write(data ...T) {
	for _, d := range data {
		h.state = h.f.Compress(h.state, d)
	}
}

func (h *merkleDamgardHasher[T]) Sum() T {
	return h.state
}
