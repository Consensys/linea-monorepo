package poseidon2_koalabear

import (
	"github.com/consensys/gnark/frontend"
	"github.com/consensys/linea-monorepo/prover/maths/field/koalagnark"
)

// same code as in poseidon2_circuit.go, but the variables are koalagnark.Var, to follow the
// same api as poseidon2_bls12377.GnarkMDHasher. Poseidon2 on koala is compiled only on koala, so
// we convert the koalagnark.Var to frontendVariable and use GnarkMDHasher.

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

func toOctuplet(v koalagnark.Octuplet) Octuplet {
	var res Octuplet
	for j := 0; j < 8; j++ {
		res[j] = v[j].Native()
	}
	return res
}

func toWVOctuplet(v Octuplet) koalagnark.Octuplet {
	var res koalagnark.Octuplet
	for j := 0; j < 8; j++ {
		res[j] = koalagnark.WrapFrontendVariable(v[j])
	}
	return res
}

func (h *GnarkMDHasherWV) Write(data ...koalagnark.Element) {
	buffer := make([]frontend.Variable, len(data))
	for i := 0; i < len(buffer); i++ {
		buffer[i] = data[i].Native()
	}
	h.gnarkMDHasher.Write(buffer...)
}

func (h *GnarkMDHasherWV) WriteOctuplet(data ...koalagnark.Octuplet) {
	var buf Octuplet
	for i := 0; i < len(data); i++ {
		buf = toOctuplet(data[i])
		h.gnarkMDHasher.WriteOctuplet(buf)
	}
}

func (h *GnarkMDHasherWV) SetState(state koalagnark.Octuplet) {
	_state := toOctuplet(state)
	h.gnarkMDHasher.SetState(_state)
}

func (h *GnarkMDHasherWV) State() koalagnark.Octuplet {
	return toWVOctuplet(h.gnarkMDHasher.State())
}

func (h *GnarkMDHasherWV) Sum() koalagnark.Octuplet {
	s := h.gnarkMDHasher.Sum()
	return toWVOctuplet(s)
}
