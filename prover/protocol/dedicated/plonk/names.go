package plonk

import (
	"fmt"

	"github.com/consensys/accelerated-crypto-monorepo/protocol/ifaces"
)

// Returns the complete name of a gate columns
func (ctx *Ctx) colIDf(name string, args ...any) ifaces.ColID {
	fmtted := fmt.Sprintf(name, args...)
	return ifaces.ColIDf("%v_%v", ctx.name, fmtted)
}

// Returns a queryID
func (ctx *Ctx) queryIDf(name string, args ...any) ifaces.QueryID {
	fmtted := fmt.Sprintf(name, args...)
	return ifaces.QueryIDf("%v_%v", ctx.name, fmtted)
}

// Returns a labeled string
func (ctx *Ctx) Sprintf(name string, args ...any) string {
	fmtted := fmt.Sprintf(name, args...)
	return fmt.Sprintf("%v_%v", ctx.name, fmtted)
}
