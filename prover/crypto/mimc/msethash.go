package mimc

import (
	"github.com/consensys/gnark/frontend"
	"github.com/consensys/linea-monorepo/prover/maths/field"
)

const (
	// The dimension 2 is obtained by simulating the CPW attack on this
	// instantiation of SIS and finding it is resistant enough.
	MSetHashSize = 2
)

// MSetHash represents a multisets hash (LtHash) instantiated using the MiMC
// hash function. The zero value of this type is a valid multisets hash for
// the empty set.
type MSetHash [MSetHashSize]field.Element

// MSetHashGnark is a multisets hash (LtHash) instantiated using the MiMC
// hash function. The zero value of this type is a valid multisets hash for
// the empty set.
type MSetHashGnark [MSetHashSize]frontend.Variable

// Insert adds the given messages to the multisets hash. The message can be an
// array of field elements of any size. The function panics if given an empty
// msg.
func (m *MSetHash) Insert(msgs ...field.Element) {
	m.update(false, msgs...)
}

// Remove removes the given messages from the multisets hash. The message can be
// an array of field elements of any size. The function panics if given an empty
// msg.
func (m *MSetHash) Remove(msgs ...field.Element) {
	m.update(true, msgs...)
}

// Add combines the two multisets hashes into a single multisets hash.
func (m *MSetHash) Add(other MSetHash) {
	for i := 0; i < MSetHashSize; i++ {
		m[i].Add(&m[i], &other[i])
	}
}

// Sub substracts the multiset "other" from "m"
func (m *MSetHash) Sub(other MSetHash) {
	for i := 0; i < MSetHashSize; i++ {
		m[i].Sub(&m[i], &other[i])
	}
}

// update adds or removes an element from the multisets hash.
func (m *MSetHash) update(rem bool, msgs ...field.Element) {

	var state field.Element

	if len(msgs) == 0 {
		panic("got provided an empty message")
	}

	for _, msg := range msgs {
		state = BlockCompression(state, msg)
	}

	for i := 0; i < MSetHashSize; i++ {
		if i > 0 {
			state = BlockCompression(state, field.Zero())
		}

		if rem {
			m[i].Sub(&m[i], &state)
		} else {
			m[i].Add(&m[i], &state)
		}
	}
}

// updateGnark updates the multisets hash using the gnark library.
func (m *MSetHashGnark) update(api frontend.API, rem bool, msgs []frontend.Variable) {

	state := frontend.Variable(0)

	if len(msgs) == 0 {
		panic("got provided an empty message")
	}

	for _, msg := range msgs {
		state = GnarkBlockCompression(api, state, msg)
	}

	for i := 0; i < MSetHashSize; i++ {
		if i > 0 {
			state = GnarkBlockCompression(api, state, 0)
		}

		if rem {
			m[i] = api.Sub(m[i], state)
		} else {
			m[i] = api.Add(m[i], state)
		}
	}
}

// Add combines two multisets in a gnark library
func (m *MSetHashGnark) Add(api frontend.API, other MSetHashGnark) {
	for i := 0; i < MSetHashSize; i++ {
		m[i] = api.Add(m[i], other[i])
	}
}
