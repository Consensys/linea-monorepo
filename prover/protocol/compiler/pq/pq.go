// Package pq implements a post-quantum (PQ) proof wrapper that takes a
// Vortex-compiled Wizard IOP and recursively applies self-recursion until the
// estimated opening-proof size falls below a caller-specified byte budget.
//
// # Background
//
// The Vortex polynomial commitment scheme produces an "opening proof" that can
// be several megabytes in size (the dominant term is the opened column
// fragments: numColsToOpen × numRows × 32 bytes per field element).
//
// To publish such a proof on Ethereum L1 – where calldata costs dominate – we
// need the proof to fit within a size budget of roughly 300 KB.
//
// The PQ wrapper achieves this purely with hash-based (post-quantum)
// primitives by chaining the following compilation steps:
//
//  1. [selfrecursion.SelfRecurse]  – compresses the Vortex proof into a
//     new, smaller Wizard IOP whose Vortex parameters are set automatically.
//  2. [mimc.CompileMiMC]           – compiles MiMC hash queries.
//  3. [compiler.Arcane]            – compiles the resulting PLONK-in-Wizard
//     IOP with a controlled target column size.
//  4. [vortex.Compile]             – applies Vortex again with the same SIS
//     and blow-up parameters.
//
// Each iteration produces a smaller Vortex proof.  After at most
// [maxIterations] iterations the loop checks whether [EstimateOpeningProofSizeBytes]
// is below the configured budget and terminates.
//
// # Size Reduction Properties
//
// Self-recursion is beneficial when the original circuit is large.  For
// small toy circuits the recursive overhead (constant-ish) can exceed the
// savings, leading to a larger post-recursion proof.  In practice, Linea's
// zkEVM circuits have millions of columns, so repeated self-recursion
// converges to a small, constant-size proof that is independent of the
// original circuit size.
//
// # Usage
//
//	import "github.com/consensys/linea-monorepo/prover/protocol/compiler/pq"
//
//	cfg := pq.DefaultConfig()
//	compiled := wizard.Compile(
//	    defineMyIOP,
//	    vortex.Compile(2, ...),        // initial Vortex compilation
//	    pq.Wrap(cfg),                  // PQ wrapper – reduces proof to < 300 KB
//	    dummy.Compile,                 // final step (or replace with your prover)
//	)
package pq

import (
	"fmt"

	"github.com/consensys/linea-monorepo/prover/crypto/ringsis"
	"github.com/consensys/linea-monorepo/prover/protocol/compiler"
	"github.com/consensys/linea-monorepo/prover/protocol/compiler/mimc"
	"github.com/consensys/linea-monorepo/prover/protocol/compiler/selfrecursion"
	"github.com/consensys/linea-monorepo/prover/protocol/compiler/vortex"
	"github.com/consensys/linea-monorepo/prover/protocol/wizard"
	"github.com/sirupsen/logrus"
)

const (
	// DefaultProofSizeBudgetBytes is the default maximum serialized proof size
	// that the PQ wrapper aims to achieve.  It corresponds to 300 KiB.
	DefaultProofSizeBudgetBytes = 300 * 1024

	// defaultMaxIterations caps the recursion depth to prevent infinite loops
	// when the target cannot be reached with the current parameters.
	defaultMaxIterations = 16

	// defaultBlowUpFactor is the Reed-Solomon blow-up factor used for the
	// Vortex re-compilation step inside each wrapper iteration.
	defaultBlowUpFactor = 2

	// defaultTargetColSize is the column size passed to the Arcane compiler
	// during each wrapper iteration.
	defaultTargetColSize = 1 << 13
)

// Config holds the tunable parameters of the PQ wrapper.
type Config struct {
	// ProofSizeBudgetBytes is the target maximum serialized proof size
	// (in bytes).  The wrapper iterates until the estimated Vortex opening
	// proof falls below this value.  Defaults to [DefaultProofSizeBudgetBytes].
	ProofSizeBudgetBytes int

	// MaxIterations is the maximum number of self-recursion rounds.  The
	// wrapper panics if this limit is reached without satisfying the budget.
	// Defaults to [defaultMaxIterations].
	MaxIterations int

	// BlowUpFactor is the Reed-Solomon blow-up factor used inside each
	// Vortex re-compilation step.  Must be a power of two ≥ 2.
	BlowUpFactor int

	// TargetColSize is the target column size passed to the Arcane compiler
	// during each wrapper iteration.
	TargetColSize int

	// SISParams are the Ring-SIS parameters used for the inner Vortex
	// invocations.  Defaults to [ringsis.StdParams].
	SISParams ringsis.Params

	// NumOpenedCols is the number of columns opened during each inner Vortex
	// invocation.  When 0 (default) the compiler selects the number
	// automatically based on the target security level.
	NumOpenedCols int
}

// DefaultConfig returns a [Config] populated with sensible defaults that
// target a proof size of 300 KB.
func DefaultConfig() Config {
	return Config{
		ProofSizeBudgetBytes: DefaultProofSizeBudgetBytes,
		MaxIterations:        defaultMaxIterations,
		BlowUpFactor:         defaultBlowUpFactor,
		TargetColSize:        defaultTargetColSize,
		SISParams:            ringsis.StdParams,
	}
}

// Option is a functional option that mutates a [Config].
type Option func(*Config)

// WithProofSizeBudget sets the target maximum serialized proof size in bytes.
func WithProofSizeBudget(bytes int) Option {
	return func(c *Config) {
		c.ProofSizeBudgetBytes = bytes
	}
}

// WithMaxIterations overrides the maximum number of self-recursion rounds.
func WithMaxIterations(n int) Option {
	return func(c *Config) {
		c.MaxIterations = n
	}
}

// WithBlowUpFactor sets the Reed-Solomon blow-up factor for the inner Vortex
// re-compilation.
func WithBlowUpFactor(rho int) Option {
	return func(c *Config) {
		c.BlowUpFactor = rho
	}
}

// WithTargetColSize sets the Arcane target column size used inside each
// wrapper iteration.
func WithTargetColSize(size int) Option {
	return func(c *Config) {
		c.TargetColSize = size
	}
}

// WithSISParams overrides the Ring-SIS parameters used for the inner Vortex
// invocations.
func WithSISParams(p ringsis.Params) Option {
	return func(c *Config) {
		c.SISParams = p
	}
}

// WithNumOpenedCols fixes the number of opened columns per Vortex query.
// When not set the compiler chooses automatically.
func WithNumOpenedCols(n int) Option {
	return func(c *Config) {
		c.NumOpenedCols = n
	}
}

// Wrap returns a wizard compiler function that applies the PQ wrapper to a
// previously Vortex-compiled [wizard.CompiledIOP].
//
// The returned function is intended to be passed directly to [wizard.Compile]
// or [wizard.ContinueCompilation]:
//
//	compiled := wizard.Compile(
//	    defineMyIOP,
//	    vortex.Compile(2, ...),
//	    pq.Wrap(pq.DefaultConfig(), pq.WithProofSizeBudget(200*1024)),
//	    dummy.Compile,
//	)
func Wrap(cfg Config, opts ...Option) func(*wizard.CompiledIOP) {
	for _, o := range opts {
		o(&cfg)
	}

	return func(comp *wizard.CompiledIOP) {
		logrus.Infof(
			"[pq-wrapper] starting with budget=%d bytes, maxIter=%d",
			cfg.ProofSizeBudgetBytes, cfg.MaxIterations,
		)

		vortexCtx := extractVortexCtx(comp)
		estimated := vortexCtx.EstimateOpeningProofSizeBytes()
		logrus.Infof("[pq-wrapper] initial estimated proof size: %d bytes", estimated)

		if estimated <= cfg.ProofSizeBudgetBytes {
			logrus.Infof("[pq-wrapper] proof already within budget, no recursion needed")
			// restore the context so that the next compiler step can use it
			comp.PcsCtxs = vortexCtx
			return
		}

		// Restore the vortex context so that SelfRecurse can consume it.
		comp.PcsCtxs = vortexCtx

		for i := 0; i < cfg.MaxIterations; i++ {
			logrus.Infof("[pq-wrapper] iteration %d: applying self-recursion", i+1)

			// Build the slice of inner Vortex options.
			innerVortexOpts := buildVortexOptions(cfg)

			// Apply one round: SelfRecurse -> MiMC -> Arcane -> Vortex.
			comp = wizard.ContinueCompilation(
				comp,
				selfrecursion.SelfRecurse,
				mimc.CompileMiMC,
				compiler.Arcane(compiler.WithTargetColSize(cfg.TargetColSize)),
				vortex.Compile(cfg.BlowUpFactor, innerVortexOpts...),
			)

			innerCtx := extractVortexCtx(comp)
			estimated = innerCtx.EstimateOpeningProofSizeBytes()
			logrus.Infof("[pq-wrapper] iteration %d: estimated proof size after recursion: %d bytes", i+1, estimated)

			if estimated <= cfg.ProofSizeBudgetBytes {
				logrus.Infof("[pq-wrapper] target reached after %d iteration(s)", i+1)
				// Restore the vortex context for the next compiler step.
				comp.PcsCtxs = innerCtx
				return
			}

			// Restore for the next round of self-recursion.
			comp.PcsCtxs = innerCtx
		}

		panic(fmt.Sprintf(
			"[pq-wrapper] failed to reduce proof size to %d bytes within %d iterations; "+
				"last estimated size was %d bytes. "+
				"Consider increasing MaxIterations or adjusting SIS/blow-up parameters.",
			cfg.ProofSizeBudgetBytes, cfg.MaxIterations, estimated,
		))
	}
}

// buildVortexOptions constructs the list of [vortex.VortexOp] options from a
// [Config].
func buildVortexOptions(cfg Config) []vortex.VortexOp {
	opts := []vortex.VortexOp{
		vortex.WithSISParams(&cfg.SISParams),
	}
	if cfg.NumOpenedCols > 0 {
		opts = append(opts, vortex.ForceNumOpenedColumns(cfg.NumOpenedCols))
	}
	return opts
}

// extractVortexCtx extracts the [vortex.Ctx] from the compiled IOP without
// consuming it (it leaves PcsCtxs intact so it can be restored by the caller).
func extractVortexCtx(comp *wizard.CompiledIOP) *vortex.Ctx {
	ctx := comp.PcsCtxs
	if ctx == nil {
		panic("[pq-wrapper] the compiled IOP does not have a Vortex context; " +
			"make sure to apply vortex.Compile before pq.Wrap")
	}
	vCtx, ok := ctx.(*vortex.Ctx)
	if !ok {
		panic(fmt.Sprintf("[pq-wrapper] unexpected PCS context type %T; expected *vortex.Ctx", ctx))
	}
	return vCtx
}
