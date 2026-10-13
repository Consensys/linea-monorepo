// Package recursion is a machine to run the ray verifier inside ray: a
// write-once memory of 8-lane cells, chips (one wiop module each) and a bus
// (one log-derivative sum) through which the chips read and write the memory.
//
// A program is built once (the inputs, constants and instructions), compiled
// into a wiop.System whose setup (precomputed) columns hold the program —
// every address, multiplicity and opcode — and executed on any number of
// statements, each giving the chips' witness columns.
package recursion

import (
	"github.com/LFDT-Lineth/lineth-monorepo/prover-ray/maths/koalabear/field"
)

// CellWidth is the number of base elements of a memory cell: a digest, an E6
// value in lanes 0..5, or a scalar in lane 0.
const CellWidth = 8

// Cell is the value of a memory cell.
type Cell [CellWidth]field.Element

// Cell layouts: an E6 value in lanes 0..5 (B0.A0, B0.A1, B1.A0, B1.A1, B2.A0,
// B2.A1), a scalar in lane 0. The other lanes are zero, which every reader
// enforces through the constants of its tuple. A scalar cell is also the E6
// cell of its base-field embedding.

// E6Cell returns the cell of x.
func E6Cell(x field.Ext) Cell {
	return Cell{x.B0.A0, x.B0.A1, x.B1.A0, x.B1.A1, x.B2.A0, x.B2.A1}
}

// CellE6 returns the E6 value of a cell (lanes 0..5).
func CellE6(c Cell) field.Ext {
	var x field.Ext
	x.B0.A0, x.B0.A1, x.B1.A0, x.B1.A1, x.B2.A0, x.B2.A1 = c[0], c[1], c[2], c[3], c[4], c[5]
	return x
}

// ScalarCell returns the cell of v.
func ScalarCell(v field.Element) Cell { return Cell{v} }

// Item kinds: the number of lanes of an item; these are the usual ones.
const (
	KindScalar = 1
	KindE6     = 6
	KindDigest = 8
)

const (
	roleNone = iota
	roleRead
	roleWrite
)

// slot is one memory access of a chip row: a read of an address, or the
// write of a freshly allocated one.
type slot struct {
	addr, role int
}

func (m *Machine) rd(addr int) slot { m.read(addr); return slot{addr, roleRead} }

func (m *Machine) wr() slot { return slot{m.alloc(), roleWrite} }

// mult is the signed bus multiplicity of s: −1 for a read, the read count for
// a write, 0 for no access.
func (m *Machine) mult(s slot) int64 {
	switch s.role {
	case roleRead:
		return -1
	case roleWrite:
		return int64(m.reads[s.addr])
	}
	return 0
}
