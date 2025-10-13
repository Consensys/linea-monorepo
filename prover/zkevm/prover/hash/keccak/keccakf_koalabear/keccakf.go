package keccakfkoalabear

import "github.com/consensys/linea-monorepo/prover/protocol/ifaces"

const (
	// Number of 64bits lanes in a keccak block
	numLanesInBlock = 17
)

// Wizard module responsible for proving a sequence of keccakf permutation
type Module struct {
	// Maximal number of Keccakf permutation that the module can handle
	MaxNumKeccakf int

	// The State of the keccakf before starting a new round.
	// Note : unlike the original keccakf where the initial State is zero,
	// the initial State here is the first block of the message.
	State [5][5]ifaces.Column

	// Columns representing the messages blocks hashed with keccak. More
	// Given our implementation it is more efficient to do the
	// xoring in base B (so at the end of the function) and the "initial" XOR
	// cannot be done at this moment. Fortunately, the initial XOR operation is
	// straightforward to handle as an edge-case since it is done against a
	// zero-state. Thus,  the entries of position 23 mod 24 are in Base B
	// and they are zero of the corresponding round is the last round of the
	// last call to the sponge function for a given hash (i.e. there are no more
	// blocks to XORIN and what remains is only to recover the result of the
	// hash). The first block are in Base A located at positions 0 mod 24.
	// At any other position block is zero.
	//  The Keccakf module trusts these columns to be well-formed.
	Blocks [numLanesInBlock]ifaces.Column

	// It is 1 over the effective part of the module,
	// indicating the rows of the module occupied by the witness.
	IsActive ifaces.Column

	Theta theta
}
