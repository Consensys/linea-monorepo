package recursion

// Setup columns of the witness chip.
const (
	witnessAddr = "addr"
	witnessMult = "mult"
)

// Witness columns of the witness chip.
const (
	witnessV = "v" // lanes
)

// witnessChip: one row per cell the prover writes freely: the inputs, and the
// hints, cells computed during execution whose readers check them.
//
// Setup: addr, mult. Witness: v0..7. Bus: writes [addr ‖ v] with
// multiplicity mult (the cell's read count). No constraint: the readers check
// the values.
type witnessChip struct{ m *Machine }

func (c witnessChip) name() string { return witnessMod }

func (c witnessChip) cells() []int {
	return append(append([]int(nil), c.m.inputs...), c.m.hints...)
}

func (c witnessChip) rows() int { return len(c.cells()) }

func (c witnessChip) setup(cs *cols) {
	cs.declare(witnessAddr, witnessMult)
	for row, a := range c.cells() {
		cs.set(witnessAddr, row, uint64(a))
		cs.set(witnessMult, row, uint64(c.m.reads[a]))
	}
}

func (c witnessChip) define(m *module, bus *Bus) error {
	bus.Write(m, m.s(witnessAddr), cellLanes(m.lanes(witnessV, CellWidth)), m.s(witnessMult))
	return nil
}

func (c witnessChip) trace(cs *cols, r *Run) error {
	cs.declareLanes(witnessV, CellWidth)
	for row, a := range c.cells() {
		cs.setCell(witnessV, row, r.values[a])
	}
	return nil
}
