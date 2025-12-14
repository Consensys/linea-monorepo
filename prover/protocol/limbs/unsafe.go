package limbs

import "github.com/consensys/linea-monorepo/prover/protocol/ifaces"

func NewLimbsFromRawUnsafe[E Endianness](name ifaces.ColID, limbs []ifaces.Column) Limbs[E] {
	return Limbs[E]{c: limbs, name: name}
}
