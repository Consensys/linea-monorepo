package recursion

import (
	"fmt"

	"github.com/LFDT-Lineth/lineth-monorepo/prover-ray/maths/koalabear/field"
	"github.com/LFDT-Lineth/lineth-monorepo/prover-ray/wiop"
)

// Witness columns of the p2 chip (besides the AIR's, see p2_air.go).
const (
	p2Cnt = "cnt" // the row's number of reads
)

// numPerms is the number of compressions the program computes: the rows of
// the P2 core.
func (m *Machine) numPerms() int {
	n := 0
	for _, d := range m.mds {
		n += len(d.blocks)
	}
	return n
}

// p2Chip: the P2 core, one Poseidon2 permutation per row (see p2_air.go).
//
// Witness: the AIR's columns (in0..15, …, out0..15) and cnt. Bus: writes the
// compression tuple (in0..15, c0..7), c_i = out_(8+i) + in_(8+i), with
// multiplicity cnt (its number of reads) on the p2 sum. The rows the program
// does not use, padding included, are the permutation of zero, read by none.
type p2Chip struct{ m *Machine }

func (c p2Chip) name() string { return p2Mod }

func (c p2Chip) rows() int { return c.m.numPerms() }

func (c p2Chip) setup(*cols) {}

func (c p2Chip) define(m *module, bus *Bus) error {
	defineP2(m)
	var in [P2Width]wiop.Expression
	for i := range in {
		in[i] = m.w(laneName(p2In, i))
	}
	var out [p2Half]wiop.Expression
	for i := range out {
		out[i] = wiop.Add(m.w(laneName(p2Out, p2Half+i)), in[p2Half+i])
	}
	bus.WriteP2(m, in, out, m.w(p2Cnt))
	return nil
}

// trace fills every row (padding included) with a permutation: the program's
// compressions, then the permutation of zero.
func (c p2Chip) trace(cs *cols, r *Run) error {
	if len(r.p2) != c.rows() {
		return fmt.Errorf("%d compressions recorded, want %d", len(r.p2), c.rows())
	}
	var zero [P2Width]field.Element
	padding := p2Trace(zero)
	for row := range cs.n {
		vals := padding
		if row < len(r.p2) {
			vals = p2Trace(r.p2[row])
			cs.set(p2Cnt, row, 1)
		}
		for name, v := range vals {
			cs.setElem(name, row, v)
		}
	}
	cs.declare(p2Cnt)
	return nil
}
