package keccakfkoalabear

import "github.com/consensys/linea-monorepo/prover/protocol/ifaces"

const (
	// Number of 64bits lanes in a keccak block
	numLanesInBlock = 17
)

// each lane is 64 bits, represented as 8 bytes.
type lane [8]ifaces.Column

// keccakf state is a 5x5 matrix of lanes.
type state [5][5]lane

// state after each base conversion, each lane is decomposed into 16 slices of 4 bits each.
type stateBaseConversion [5][5][16]ifaces.Column

// Wizard module responsible for proving a sequence of keccakf permutation
type Module struct {
	// maximal number of Keccakf permutation that the module can handle
	maxNumKeccakf int

	// the State of the keccakf before starting a new round.
	// Note : unlike the original keccakf where the initial State is zero,
	// the initial State here is the first block of the message.
	state state

	// the blocks of the message to be absorbed.
	// first blocks of messages are located in positions 0 mod 24 and are represented in base clean 12,
	// other blocks of message are located in positions 23 mod 24 and are represented in base clean 11.
	// otherwise the blocks are zero.
	blocks [numLanesInBlock]lane

	// it is 1 over the effective part of the module,
	// indicating the rows of the module occupied by the witness.
	isActive ifaces.Column

	// theta module, responsible for updating the state in the theta step of keccakf
	theta theta
}
