package recursion

import (
	"math/rand/v2"
	"testing"

	"github.com/LFDT-Lineth/lineth-monorepo/prover-ray/maths/koalabear/field"
	"github.com/stretchr/testify/require"
)

var packKinds = []int{KindScalar, KindE6, KindDigest, KindE6, KindScalar}

// packProgram packs items of every kind, at offsets that do and do not cross
// a cell (stream offsets 0, 1, 7 (crossing), 15 (crossing) and 21), the last
// one a constant; and takes two Lanes of the digest, lanes 2..7 and lane 7.
// With copies, the stream and the first lanes are copied, so that they are
// read outside the windows too; without, only the windows check them.
func packProgram(copies bool) (m *Machine, stream []int, length int, lanes [2]int) {
	m = &Machine{}
	cp := withCopy(m)
	items := []int{m.InputScalar(), m.InputE6(), m.InputDigest(), m.InputE6(), m.Const(ScalarCell(field.NewElement(3)))}
	stream, length = m.Pack(items, packKinds)
	lanes = [2]int{m.Lanes(items[2], 2, KindE6), m.Lanes(items[2], 7, KindScalar)}
	if copies {
		for _, c := range stream {
			cp.Copy(c)
		}
		cp.Copy(lanes[0])
	}
	return m, stream, length, lanes
}

// packInputs returns the input items of packProgram: a scalar, an E6 value,
// a digest and an E6 value.
func packInputs(seed uint64) []Cell {
	r := rand.New(rand.NewPCG(seed, 6))
	in := make([]Cell, 4)
	for i, k := range packKinds[:4] {
		for j := range k {
			in[i][j].SetUint64(r.Uint64())
		}
	}
	return in
}

// TestWindowChip: the stream holds the items, in order and zero padded to
// whole cells, the lanes hold the digest's, and both prove, for two inputs.
func TestWindowChip(t *testing.T) {
	m, stream, length, lanes := packProgram(true)
	require.Equal(t, 22, length)
	require.Len(t, stream, 3)
	// the window rows write the Lanes outputs: the witness chip only writes
	// the inputs and the stream
	require.Equal(t, len(m.inputs)+len(stream), witnessChip{m}.rows())
	p, err := m.Compile()
	require.NoError(t, err)
	for _, seed := range []uint64{1, 2} {
		in := packInputs(seed)
		r, err := p.Execute(in)
		require.NoError(t, err)

		var xs []field.Element
		for i, k := range packKinds[:4] {
			xs = append(xs, in[i][:k]...)
		}
		xs = append(xs, field.NewElement(3))
		for c, want := range PackElements(xs) {
			require.Equal(t, want, r.Value(stream[c]), "stream cell %d", c)
		}
		var want0, want1 Cell
		copy(want0[:], in[2][2:8])
		want1[0] = in[2][7]
		require.Equal(t, want0, r.Value(lanes[0]), "lanes 2..7")
		require.Equal(t, want1, r.Value(lanes[1]), "lane 7")
		require.NoError(t, proveAndVerify(p, r), "seed %d", seed)
	}
}

// TestWindowChipRejects: a wrong stream cell (a hint) or Lanes output (written
// by its window row), consistent on the bus, is rejected by the windows,
// beyond a Lanes output's kind included. Nothing but the windows reads the
// stream and the lanes. The stream's padding lanes are left to its hasher
// (see Pack).
func TestWindowChipRejects(t *testing.T) {
	for name, tamper := range map[string]struct {
		cell func(stream []int, lanes [2]int) int
		lane int
	}{
		// the scalar item, at offset 0
		"scalar item": {func(s []int, _ [2]int) int { return s[0] }, 0},
		// lane 3 of the E6 item, which starts at offset 1 of the first cell
		"item lane": {func(s []int, _ [2]int) int { return s[0] }, 4},
		// the digest crosses from the first cell (lane 7) into the second
		"crossing item, first cell":  {func(s []int, _ [2]int) int { return s[0] }, 7},
		"crossing item, second cell": {func(s []int, _ [2]int) int { return s[1] }, 2},
		// the constant, the stream's last item
		"constant item":     {func(s []int, _ [2]int) int { return s[2] }, 5},
		"lanes output, E6":  {func(_ []int, l [2]int) int { return l[0] }, 3},
		"lanes output, one": {func(_ []int, l [2]int) int { return l[1] }, 0},
		// beyond the output's kind, zero by its row's strict form
		"lanes output, E6, lane 6":     {func(_ []int, l [2]int) int { return l[0] }, 6},
		"lanes output, scalar, lane 3": {func(_ []int, l [2]int) int { return l[1] }, 3},
	} {
		t.Run(name, func(t *testing.T) {
			m, stream, _, lanes := packProgram(false)
			p, err := m.Compile()
			require.NoError(t, err)
			r, err := p.Execute(packInputs(1))
			require.NoError(t, err)
			c := tamper.cell(stream, lanes)
			r.values[c][tamper.lane].SetUint64(5)
			require.NoError(t, p.trace(r))
			require.Error(t, proveAndVerify(p, r), "wrong cell accepted")
		})
	}
}

// TestPackItemKind: an E6 cell packed as a scalar proves only if its lanes
// 1..5 are zero: its window reads it zero beyond the packed kind, since the
// stream (and its hash) holds lane 0 only.
func TestPackItemKind(t *testing.T) {
	m := &Machine{}
	e := m.InputE6()
	m.Pack([]int{e}, []int{KindScalar})
	p, err := m.Compile()
	require.NoError(t, err)

	var x Cell
	x[0].SetUint64(3)
	r, err := p.Execute([]Cell{x})
	require.NoError(t, err)
	require.NoError(t, proveAndVerify(p, r), "an E6 value in the base field")

	x[1].SetUint64(4)
	r, err = p.Execute([]Cell{x})
	require.NoError(t, err)
	require.Error(t, proveAndVerify(p, r), "lane 1 of an E6 cell packed as a scalar")
}
