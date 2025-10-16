package keccakfkoalabear

import (
	"github.com/consensys/linea-monorepo/prover/protocol/ifaces"
	sym "github.com/consensys/linea-monorepo/prover/symbolic"
)

// BaseRecompose recompose the given slice r in the given base in little-endian order.
func BaseRecompose(r []ifaces.Column, base int) *sym.Expression {
	// .. using the Horner method
	s := sym.NewConstant(0)
	for i := len(r) - 1; i >= 0; i-- {
		s = sym.Mul(s, base)
		s = sym.Add(s, r[i])
	}
	return s
}
