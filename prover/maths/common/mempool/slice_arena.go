package mempool

import (
	"github.com/consensys/linea-monorepo/prover/maths/field"
	"github.com/consensys/linea-monorepo/prover/maths/field/fext"
)

// SliceArena is a simple not-threadsafe arena implementation that uses a
// mempool to carry its allocation. It will only put back free memory in the
// the parent pool when TearDown is called.
type SliceArena struct {
	freesBase []*[]field.Element
	freesExt  []*[]fext.Element
	parent    MemPool
}

func WrapsWithMemCache(pool MemPool) *SliceArena {
	return &SliceArena{
		freesBase: make([]*[]field.Element, 0, 1<<7),
		freesExt:  make([]*[]fext.Element, 0, 1<<7),
		parent:    pool,
	}
}

func (m *SliceArena) Prewarm(nbPrewarm int) MemPool {
	m.parent.Prewarm(nbPrewarm)
	return m
}

func (m *SliceArena) Alloc() *[]field.Element {

	if len(m.freesBase) == 0 {
		return m.parent.Alloc()
	}

	last := m.freesBase[len(m.freesBase)-1]
	m.freesBase = m.freesBase[:len(m.freesBase)-1]
	return last
}

func (m *SliceArena) AllocBase() *[]field.Element {

	if len(m.freesBase) == 0 {
		return m.parent.AllocBase()
	}

	last := m.freesBase[len(m.freesBase)-1]
	m.freesBase = m.freesBase[:len(m.freesBase)-1]
	return last
}

func (m *SliceArena) AllocExt() *[]fext.Element {

	if len(m.freesExt) == 0 {
		return m.parent.AllocExt()
	}

	last := m.freesExt[len(m.freesBase)-1]
	m.freesBase = m.freesBase[:len(m.freesBase)-1]
	return last
}

func (m *SliceArena) Free(v *[]field.Element) error {
	m.freesBase = append(m.freesBase, v)
	return nil
}

func (m *SliceArena) FreeBase(v *[]field.Element) error {
	m.freesBase = append(m.freesBase, v)
	return nil
}

func (m *SliceArena) FreeExt(v *[]fext.Element) error {
	m.freesExt = append(m.freesExt, v)
	return nil
}
func (m *SliceArena) Size() int {
	return m.parent.Size()
}

func (m *SliceArena) TearDown() {
	// free the base vectors
	for i := range m.freesBase {
		m.parent.FreeBase(m.freesBase[i])
	}
	// free the extension vectors
	for i := range m.freesExt {
		m.parent.FreeExt(m.freesExt[i])
	}
}
