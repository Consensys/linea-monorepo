package tree

import (
	"github.com/consensys/linea-monorepo/prover/crypto/ringsis"
	"github.com/consensys/linea-monorepo/prover/protocol/compiler"
	"github.com/consensys/linea-monorepo/prover/protocol/compiler/cleanup"
	"github.com/consensys/linea-monorepo/prover/protocol/compiler/poseidon2"
	"github.com/consensys/linea-monorepo/prover/protocol/compiler/selfrecursion"
	"github.com/consensys/linea-monorepo/prover/protocol/compiler/vortex"
	"github.com/consensys/linea-monorepo/prover/protocol/wizard"
)

// CompilationSuite is an alias for a sequence of compiler passes.
type CompilationSuite = []func(*wizard.CompiledIOP)

// sisInstance matches the parameters used in zkevm/full.go for consistency.
var sisInstance = ringsis.Params{LogTwoBound: 16, LogTwoDegree: 6}

// TreeAggregationCompilationSuite returns the compilation suite for tree
// aggregation nodes. This is modeled after fullSecondCompilationSuite in
// zkevm/full.go but adapted for the tree aggregation context.
//
// The suite applies:
//  1. Cleanup + Poseidon2 compilation + Arcane optimization
//  2. Vortex commitment with Ring-SIS + Poseidon2 Merkle trees
//  3. Self-recursion to verify the Vortex proof within the protocol
//  4. Repeat with decreasing column sizes for proof compression
//  5. Final Vortex round with PremarkAsSelfRecursed for next-level compatibility
func TreeAggregationCompilationSuite() CompilationSuite {
	return CompilationSuite{
		cleanup.CleanUp,
		poseidon2.CompilePoseidon2,
		compiler.Arcane(
			compiler.WithTargetColSize(1<<20),
			compiler.WithStitcherMinSize(16),
		),
		vortex.Compile(
			2, false,
			vortex.ForceNumOpenedColumns(256),
			vortex.WithSISParams(&sisInstance),
		),

		// First round of self-recursion
		selfrecursion.SelfRecurse,
		cleanup.CleanUp,
		poseidon2.CompilePoseidon2,
		compiler.Arcane(
			compiler.WithTargetColSize(1<<16),
			compiler.WithStitcherMinSize(16),
		),
		vortex.Compile(
			16, false,
			vortex.ForceNumOpenedColumns(64),
			vortex.WithSISParams(&sisInstance),
		),

		// Second round of self-recursion
		selfrecursion.SelfRecurse,
		cleanup.CleanUp,
		poseidon2.CompilePoseidon2,
		compiler.Arcane(
			compiler.WithTargetColSize(1<<12),
			compiler.WithStitcherMinSize(16),
		),
		vortex.Compile(
			16, false,
			vortex.ForceNumOpenedColumns(64),
			vortex.WithOptionalSISHashingThreshold(1<<20),
			vortex.PremarkAsSelfRecursed(),
		),
	}
}

// TreeAggregationFinalCompilationSuite returns the compilation suite for
// the root (final) level of the tree. The last Vortex round uses IsLastRound=true
// since the output will be wrapped in BN254 rather than further recursed.
func TreeAggregationFinalCompilationSuite() CompilationSuite {
	return CompilationSuite{
		cleanup.CleanUp,
		poseidon2.CompilePoseidon2,
		compiler.Arcane(
			compiler.WithTargetColSize(1<<20),
			compiler.WithStitcherMinSize(16),
		),
		vortex.Compile(
			2, false,
			vortex.ForceNumOpenedColumns(256),
			vortex.WithSISParams(&sisInstance),
		),

		// First round of self-recursion
		selfrecursion.SelfRecurse,
		cleanup.CleanUp,
		poseidon2.CompilePoseidon2,
		compiler.Arcane(
			compiler.WithTargetColSize(1<<16),
			compiler.WithStitcherMinSize(16),
		),
		vortex.Compile(
			16, false,
			vortex.ForceNumOpenedColumns(64),
			vortex.WithSISParams(&sisInstance),
		),

		// Second round of self-recursion
		selfrecursion.SelfRecurse,
		cleanup.CleanUp,
		poseidon2.CompilePoseidon2,
		compiler.Arcane(
			compiler.WithTargetColSize(1<<12),
			compiler.WithStitcherMinSize(16),
		),
		// Final vortex round: IsLastRound=true since this will be wrapped
		// in BN254 rather than further recursed in KoalaBear
		vortex.Compile(
			16, true,
			vortex.ForceNumOpenedColumns(64),
			vortex.WithOptionalSISHashingThreshold(1<<20),
		),
	}
}
