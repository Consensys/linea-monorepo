package plonkinternal

import (
	"fmt"

	"github.com/consensys/linea-monorepo/prover/protocol/ifaces"
)

// Returns the complete Name of a gate columns
func (ctx *CompilationCtx) colIDf(Name string, args ...any) ifaces.ColID {
	fmtted := fmt.Sprintf(Name, args...)
	return ifaces.ColIDf("%v_%v", ctx.Name, fmtted)
}

// Returns a queryID
func (ctx *CompilationCtx) queryIDf(Name string, args ...any) ifaces.QueryID {
	fmtted := fmt.Sprintf(Name, args...)
	return ifaces.QueryIDf("%v_%v", ctx.Name, fmtted)
}

// Returns a labeled string
func (ctx *CompilationCtx) Sprintf(Name string, args ...any) string {
	fmtted := fmt.Sprintf(Name, args...)
	return fmt.Sprintf("%v_%v", ctx.Name, fmtted)
}
