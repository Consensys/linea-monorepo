package recursion

import (
	"fmt"

	"github.com/LFDT-Lineth/lineth-monorepo/prover-ray/maths/koalabear/field"
	"github.com/LFDT-Lineth/lineth-monorepo/prover-ray/wiop"
)

// Setup columns of the window chip.
const (
	windowActive = "active"
	windowAddrP  = "addr_p"
	windowAddrQ  = "addr_q"
	windowSelQ   = "sel_q"
	windowAddrI  = "addr_i"
	windowOff    = "off" // lanes
	windowLen    = "len" // lanes 1..7
)

// Witness columns of the window chip.
const (
	windowP = "p" // lanes
	windowQ = "q" // lanes
	windowI = "i" // lanes
)

// Streams. A stream is a sequence of elements laid out CellWidth per cell,
// zero padded to whole cells: hashing it absorbs one cell per block (ray's
// Merkle–Damgård hash absorbs 8-element blocks). The padding lanes of a
// packed stream are not checked by its windows: its hasher must ignore them
// (or fix them to zero), knowing the stream's length.

// streamCells is the number of cells of a stream of length elements.
func streamCells(length int) int { return (length + CellWidth - 1) / CellWidth }

// PackElements returns the cells of a stream of elements.
func PackElements(xs []field.Element) []Cell {
	res := make([]Cell, streamCells(len(xs)))
	for c := range res {
		for j := range CellWidth {
			if p := c*CellWidth + j; p < len(xs) {
				res[c][j] = xs[p]
			}
		}
	}
	return res
}

// InputStream allocates the input cells of a stream of length elements (see
// PackElements for their values) and returns their addresses.
func (m *Machine) InputStream(length int) []int {
	res := make([]int, streamCells(length))
	for i := range res {
		res[i] = m.Input()
	}
	return res
}

type packInst struct {
	items, kinds []int
	stream       []int // the stream cells, written by the witness chip
}

type lanesInst struct {
	src, off, kind, out int
}

type windowInst struct {
	p, q, item int // q < 0 when the item fits in p
	off, kind  int
}

// Pack lays the items (cells of the given kinds: their first kind lanes) out
// contiguously as a stream, written by the witness chip as hints, checks every
// item against its window of the stream, and returns the stream cells and its
// length in elements. Hashing the stream's first length elements hashes the
// items.
func (m *Machine) Pack(items, kinds []int) ([]int, int) {
	if len(items) != len(kinds) {
		panic(fmt.Sprintf("Pack: %d items, %d kinds", len(items), len(kinds)))
	}
	length := 0
	for _, k := range kinds {
		checkKind(k)
		length += k
	}
	stream := make([]int, streamCells(length))
	for i := range stream {
		stream[i] = m.alloc()
	}
	m.hints = append(m.hints, stream...)
	m.packs = append(m.packs, packInst{items: items, kinds: kinds, stream: stream})
	m.record(instrPack, len(m.packs)-1)
	g := 0
	for i, it := range items {
		m.Window(stream, g, kinds[i], it)
		g += kinds[i]
	}
	return stream, length
}

func (m *Machine) execPack(p packInst, r *Run) {
	var xs []field.Element
	for i, it := range p.items {
		xs = append(xs, r.values[it][:p.kinds[i]]...)
	}
	for c, v := range PackElements(xs) {
		r.values[p.stream[c]] = v
	}
}

// Lanes returns a cell holding lanes off..off+kind−1 of the cell src, in its
// lanes 0..kind−1 and zero elsewhere: for instance the E6 value of a
// challenge digest (Lanes(d, 0, KindE6)), or one of its lanes as a scalar.
// The witness chip writes it as a hint, and a window checks it against src.
func (m *Machine) Lanes(src, off, kind int) int {
	checkKind(kind)
	if off < 0 || off+kind > CellWidth {
		panic(fmt.Sprintf("Lanes: lanes %d..%d out of a cell", off, off+kind-1))
	}
	out := m.alloc()
	m.hints = append(m.hints, out)
	m.lanes = append(m.lanes, lanesInst{src, off, kind, out})
	m.record(instrLanes, len(m.lanes)-1)
	m.Window([]int{src}, off, kind, out)
	return out
}

func (m *Machine) execLanes(l lanesInst, r *Run) {
	var c Cell
	copy(c[:l.kind], r.values[l.src][l.off:l.off+l.kind])
	r.values[l.out] = c
}

// Window checks that item is the kind elements of the stream starting at
// element offset g.
func (m *Machine) Window(stream []int, g, kind, item int) {
	checkKind(kind)
	c, off := g/CellWidth, g%CellWidth
	w := windowInst{p: stream[c], q: -1, item: item, off: off, kind: kind}
	m.read(w.p)
	if off+kind > CellWidth {
		w.q = stream[c+1]
		m.read(w.q)
	}
	m.read(item)
	m.windows = append(m.windows, w)
}

func checkKind(k int) {
	if k < 1 || k > CellWidth {
		panic(fmt.Sprintf("item kind %d", k))
	}
}

// windowChip: one row per item of a packed stream (and per Lanes).
//
// Witness: p0..7, q0..7 (two consecutive stream cells, the window W = p‖q),
// i0..7 (the item). Setup: active, addr_p, addr_q, sel_q (q is read), addr_i,
// off0..7 (one-hot offset of the item in p), and len1..7 (len_k = 1 when the
// item has more than k lanes). For each lane k of the item,
//
//	len_k·(i_k − Σ_o off_o·W_(o+k)) = 0,
//
// with len_0 = 1. The item's other lanes are left to its readers: their
// tuples fix them to zero, or they do not use them. Bus: reads p, q (if
// sel_q) and the item.
type windowChip struct{ m *Machine }

func (c windowChip) name() string { return windowMod }

func (c windowChip) rows() int { return len(c.m.windows) }

func (c windowChip) setup(cs *cols) {
	cs.declare(windowActive, windowAddrP, windowAddrQ, windowSelQ, windowAddrI)
	cs.declareLanes(windowOff, CellWidth)
	for k := 1; k < CellWidth; k++ {
		cs.declare(laneName(windowLen, k))
	}
	for row, w := range c.m.windows {
		cs.set(windowActive, row, 1)
		cs.set(windowAddrP, row, uint64(w.p))
		if w.q >= 0 {
			cs.set(windowAddrQ, row, uint64(w.q))
			cs.set(windowSelQ, row, 1)
		}
		cs.set(windowAddrI, row, uint64(w.item))
		cs.set(laneName(windowOff, w.off), row, 1)
		for k := 1; k < w.kind; k++ {
			cs.set(laneName(windowLen, k), row, 1)
		}
	}
}

func (c windowChip) define(m *module, bus *Bus) error {
	active := m.s(windowActive)
	bus.Read(m, m.s(windowAddrP), cellLanes(m.lanes(windowP, CellWidth)), active)
	bus.Read(m, m.s(windowAddrQ), cellLanes(m.lanes(windowQ, CellWidth)), m.s(windowSelQ))
	bus.Read(m, m.s(windowAddrI), cellLanes(m.lanes(windowI, CellWidth)), active)

	window := func(j int) wiop.Expression {
		if j < CellWidth {
			return m.w(laneName(windowP, j))
		}
		return m.w(laneName(windowQ, j-CellWidth))
	}
	for k := range CellWidth {
		terms := make([]wiop.Expression, CellWidth)
		for o := range CellWidth {
			terms[o] = wiop.Mul(m.s(laneName(windowOff, o)), window(o+k))
		}
		diff := wiop.Sub(m.w(laneName(windowI, k)), sum(terms...))
		if k > 0 {
			diff = wiop.Mul(m.s(laneName(windowLen, k)), diff)
		}
		m.assertZero(diff)
	}
	return nil
}

func (c windowChip) trace(cs *cols, r *Run) error {
	for _, x := range []string{windowP, windowQ, windowI} {
		cs.declareLanes(x, CellWidth)
	}
	for row, w := range c.m.windows {
		cs.setCell(windowP, row, r.values[w.p])
		if w.q >= 0 {
			cs.setCell(windowQ, row, r.values[w.q])
		}
		cs.setCell(windowI, row, r.values[w.item])
	}
	return nil
}
