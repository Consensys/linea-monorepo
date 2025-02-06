package vortex

import (
	"github.com/consensys/linea-monorepo/prover/crypto/ringsis"
)

// Option to be passed to vortex
type VortexOp func(ctx *Ctx)

// Overrides the number of opened columns (should
// not be used in production)
func ForceNumOpenedColumns(nbCol int) VortexOp {
	return func(ctx *Ctx) {
		ctx.NumOpenedCol = nbCol
	}
}

// Allows passing a SIS instance
func WithSISParams(params *ringsis.Params) VortexOp {
	return func(ctx *Ctx) {
		ctx.SisParams = params
	}
}

// Allows skipping rounds when there are not many polynomials
func WithDryThreshold(dryThreshold int) VortexOp {
	return func(ctx *Ctx) {
		ctx.DryTreshold = dryThreshold
	}
}

// Replace SIS with a custom hasher
func ReplaceSisByMimc() VortexOp {
	return func(ctx *Ctx) {
		ctx.ReplaceSisByMimc = true
		ctx.SisParams = nil
	}
}

// PremarkAsSelfRecursed marks the ctx as selfrecursed. This is useful
// toward conglomerating the receiver comp but is not needed for
// self-recursion or full-recursion.
func PremarkAsSelfRecursed() VortexOp {
	return func(ctx *Ctx) {
		ctx.IsSelfrecursed = true
	}
}
