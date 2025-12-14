package limbs

import (
	"github.com/consensys/linea-monorepo/prover/maths/field"
	"github.com/consensys/linea-monorepo/prover/protocol/ifaces"
)

func NewLimbsFromRawUnsafe[E Endianness](name ifaces.ColID, limbs []ifaces.Column) Limbs[E] {
	return Limbs[E]{c: limbs, name: name}
}

func (l Limbs[E]) ToRawUnsafe() []ifaces.Column {
	return l.c
}

func NewRowFromRawUnsafe[E Endianness](r []field.Element) row[E] {
	return row[E]{T: r}
}

func (r row[E]) ToRawUnsafe() []field.Element {
	return r.T
}
