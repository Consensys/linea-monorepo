package poseidon2_koalabear

import (
	"github.com/consensys/gnark/frontend"
	"github.com/consensys/linea-monorepo/prover/maths/zk"
)

// same code as in poseidon2_circuit.go, but the variables are zk.WrappedVariable, to follow the
// same api as poseidon2_bls12377.GnarkMDHasher. Poseidon2 on koala is compiled only on koala, so
// we convert the zk.WrappedVariable to frontendVariable and use GnarkMDHasher.

// GnarkMDHasher Merkle Damgard implementation using poseidon2 as compression function with width 16
// The hashing process goes as follow:
// newState := compress(old state, buf) where buf is an octuplet, left padded with zeroes if needed
type GnarkMDHasherWV struct {
	gnarkMDHasher GnarkMDHasher
}

// NewGnarkMDHasher returns a new Octuplet
func NewGnarkMDHasherWV(api frontend.API) (GnarkMDHasherWV, error) {
	var res GnarkMDHasherWV
	var err error
	res.gnarkMDHasher, err = NewGnarkMDHasher(api)
	return res, err
}

func (h *GnarkMDHasherWV) Reset() {
	h.gnarkMDHasher.Reset()
}

func toOctuplet(v zk.Octuplet) Octuplet {
	var res Octuplet
	for j := 0; j < 8; j++ {
		res[j] = v[j].AsNative()
	}
	return res
}

func toWVOctuplet(v Octuplet) zk.Octuplet {
	var res zk.Octuplet
	for j := 0; j < 8; j++ {
		res[j] = zk.WrapFrontendVariable(v[j])
	}
	return res
}

func (h *GnarkMDHasherWV) Write(data ...zk.WrappedVariable) {
	buffer := make([]frontend.Variable, len(data))
	for i := 0; i < len(buffer); i++ {
		buffer[i] = data[i].AsNative()
	}
	h.gnarkMDHasher.Write(buffer...)
}

func (h *GnarkMDHasherWV) WriteOctuplet(data ...zk.Octuplet) {
	var buf Octuplet
	for i := 0; i < len(data); i++ {
		buf = toOctuplet(data[i])
		h.gnarkMDHasher.WriteOctuplet(buf)
	}
}

func (h *GnarkMDHasherWV) SetState(state zk.Octuplet) {
	_state := toOctuplet(state)
	h.gnarkMDHasher.SetState(_state)
}

func (h *GnarkMDHasherWV) State() zk.Octuplet {
	return toWVOctuplet(h.gnarkMDHasher.State())
}

func (h *GnarkMDHasherWV) Sum() zk.Octuplet {
	s := h.gnarkMDHasher.Sum()
	return toWVOctuplet(s)
}
