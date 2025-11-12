package poseidon2_interface

import (
	"github.com/consensys/linea-monorepo/prover/maths/field"
)

// MDHasher interface for a Merkle Damgard hasher whose underlying permutation is
// poseidon2, either over koalabear or over bls12377 (with conversion bls12377 <-> koalabear under the hood)
type MDHasher interface {
	Reset()
	WriteElements(elmts []field.Element)
	SumElement() field.Octuplet
}
