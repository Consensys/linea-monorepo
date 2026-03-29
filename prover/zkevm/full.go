package zkevm

import (
	"sync"

	"github.com/consensys/go-corset/pkg/ir/mir"
	"github.com/consensys/linea-monorepo/prover/config"
	"github.com/consensys/linea-monorepo/prover/crypto/ringsis"
	"github.com/consensys/linea-monorepo/prover/protocol/compiler"
	"github.com/consensys/linea-monorepo/prover/protocol/compiler/cleanup"
	"github.com/consensys/linea-monorepo/prover/protocol/compiler/dummy"
	"github.com/consensys/linea-monorepo/prover/protocol/compiler/plonkinwizard"
	"github.com/consensys/linea-monorepo/prover/protocol/compiler/poseidon2"
	"github.com/consensys/linea-monorepo/prover/protocol/compiler/selfrecursion"
	"github.com/consensys/linea-monorepo/prover/protocol/compiler/vortex"
	"github.com/consensys/linea-monorepo/prover/protocol/wizard"
	"github.com/consensys/linea-monorepo/prover/zkevm/arithmetization"
)

const (
	NbInputPerInstancesEcAdd           = 28
	NbInputPerInstancesEcMul           = 6
	NbInputPerInstanceEcPairMillerLoop = 1
	NbInputPerInstanceEcPairFinalExp   = 1
	NbInputPerInstanceEcPairG2Check    = 1
	NbInputPerInstanceSha2Block        = 3
	NbInputPerInstanceEcdsa            = 4

	NbInputPerInstanceBLSG1Add            = 16
	NbInputPerInstanceBLSG2Add            = 16
	NbInputPerInstanceBLSG1Msm            = 3
	NbInputPerInstanceBLSG2Msm            = 2
	NbInputPerInstanceBLSMillerLoop       = 1
	NbInputPerInstanceBLSFinalExp         = 1
	NbInputPerInstanceBLSG1Membership     = 6
	NbInputPerInstanceBLSG2Membership     = 6
	NbInputPerInstanceBLSG1Map            = 8
	NbInputPerInstanceBLSG2Map            = 2
	NbInputPerInstanceBLSC1Membership     = 16
	NbInputPerInstanceBLSC2Membership     = 16
	NbInputPerInstanceBLSPointEval        = 1
	NbInputPerInstanceBLSPointEvalFailure = 1

	NbInputPerInstanceP256Verify = 2
)

var (
	fullZkEvm              *ZkEvm
	fullZkEvmLarge         *ZkEvm
	fullZkEvmCheckOnly     *ZkEvm
	onceFullZkEvm          = sync.Once{}
	onceFullZkEvmLarge     = sync.Once{}
	onceFullZkEvmCheckOnly = sync.Once{}

	// This is the SIS instance, that has been found to minimize the overhead of
	// recursion. It is changed w.r.t to the estimated because the estimated one
	// allows for 10 bits limbs instead of just 8. But since the current state
	// of the self-recursion currently relies on the number of limbs to be a
	// power of two, we go with this one although it overshoots our security
	// level target.
	sisInstance = ringsis.Params{LogTwoBound: 16, LogTwoDegree: 6}

	dummyCompilationSuite = CompilationSuite{dummy.Compile}

	// This is the compilation suite in use for the full prover
	fullInitialCompilationSuite = CompilationSuite{
		// logdata.Log("initial-wizard"),
		poseidon2.CompilePoseidon2,
		plonkinwizard.Compile,
		compiler.Arcane(
			compiler.WithStitcherMinSize(16),
			compiler.WithTargetColSize(1<<19),
			// compiler.WithDebugMode("initial-compiler-step-0"),
			// compiler.GenCSVAfterExpansion("zkevm_first_compilation.csv"),
		),
		vortex.Compile(
			2, false,
			vortex.ForceNumOpenedColumns(256),
			vortex.WithSISParams(&sisInstance),
		),
		// logdata.Log("pre-recursion.post-vortex-1"),

		// First round of self-recursion
		selfrecursion.SelfRecurse,
		// logdata.Log("pre-recursion.post-selfrecursion-1"),
		cleanup.CleanUp,
		poseidon2.CompilePoseidon2,
		compiler.Arcane(
			compiler.WithTargetColSize(1<<17),
			compiler.WithStitcherMinSize(16),
			// compiler.WithDebugMode("initial-compiler-step-1"),
		),
		vortex.Compile(
			8, false,
			vortex.ForceNumOpenedColumns(86),
			vortex.WithSISParams(&sisInstance),
		),
		// logdata.Log("pre-recursion.post-vortex-2"),

		// Second round of self-recursion
		selfrecursion.SelfRecurse,
		// logdata.Log("pre-recursion.post-selfrecursion-2"),
		cleanup.CleanUp,
		poseidon2.CompilePoseidon2,
		compiler.Arcane(
			compiler.WithTargetColSize(1<<15),
			compiler.WithStitcherMinSize(16),
			// compiler.WithDebugMode("initial-compiler-step-2"),
		),
		vortex.Compile(
			16, false,
			vortex.ForceNumOpenedColumns(64),
			vortex.WithSISParams(&sisInstance),
		),

		// Fourth round of self-recursion
		// logdata.Log("pre-recursion.post-vortex-3"),
		selfrecursion.SelfRecurse,
		// logdata.Log("pre-recursion.post-selfrecursion-3"),
		cleanup.CleanUp,
		poseidon2.CompilePoseidon2,
		compiler.Arcane(
			compiler.WithTargetColSize(1<<14),
			compiler.WithStitcherMinSize(16),
			// compiler.WithDebugMode("initial-compiler-step-3"),
		),
		vortex.Compile(
			16, false,
			vortex.ForceNumOpenedColumns(64),
			vortex.WithOptionalSISHashingThreshold(1<<20),
			vortex.PremarkAsSelfRecursed(),
		),
		// logdata.Log("pre-recursion.post-vortex-4"),
	}

	// This is the compilation suite in use for the full prover (large variant)
	// Increases TargetColSize to fit within 2^27 SRS (134M constraints)
	// Stage 1 doubled from 1<<19 to 1<<20 (must be power of two)
	fullInitialCompilationSuiteLarge = CompilationSuite{
		// logdata.Log("initial-wizard"),
		poseidon2.CompilePoseidon2,
		plonkinwizard.Compile,
		compiler.Arcane(
			compiler.WithStitcherMinSize(16),
			compiler.WithTargetColSize(1<<20), // doubled from 1<<19
			// compiler.WithDebugMode("initial-compiler-step-0"),
			// compiler.GenCSVAfterExpansion("zkevm_first_compilation.csv"),
		),
		vortex.Compile(
			2, false,
			vortex.ForceNumOpenedColumns(256),
			vortex.WithSISParams(&sisInstance),
		),
		// logdata.Log("pre-recursion.post-vortex-1"),

		// First round of self-recursion
		selfrecursion.SelfRecurse,
		// logdata.Log("pre-recursion.post-selfrecursion-1"),
		cleanup.CleanUp,
		poseidon2.CompilePoseidon2,
		compiler.Arcane(
			compiler.WithTargetColSize(1<<17),
			compiler.WithStitcherMinSize(16),
			// compiler.WithDebugMode("initial-compiler-step-1"),
		),
		vortex.Compile(
			8, false,
			vortex.ForceNumOpenedColumns(86),
			vortex.WithSISParams(&sisInstance),
		),
		// logdata.Log("pre-recursion.post-vortex-2"),

		// Second round of self-recursion
		selfrecursion.SelfRecurse,
		// logdata.Log("pre-recursion.post-selfrecursion-2"),
		cleanup.CleanUp,
		poseidon2.CompilePoseidon2,
		compiler.Arcane(
			compiler.WithTargetColSize(1<<15),
			compiler.WithStitcherMinSize(16),
			// compiler.WithDebugMode("initial-compiler-step-2"),
		),
		vortex.Compile(
			16, false,
			vortex.ForceNumOpenedColumns(64),
			vortex.WithSISParams(&sisInstance),
		),

		// Fourth round of self-recursion
		// logdata.Log("pre-recursion.post-vortex-3"),
		selfrecursion.SelfRecurse,
		// logdata.Log("pre-recursion.post-selfrecursion-3"),
		cleanup.CleanUp,
		poseidon2.CompilePoseidon2,
		compiler.Arcane(
			compiler.WithTargetColSize(1<<14),
			compiler.WithStitcherMinSize(16),
			// compiler.WithDebugMode("initial-compiler-step-3"),
		),
		vortex.Compile(
			16, false,
			vortex.ForceNumOpenedColumns(64),
			vortex.WithOptionalSISHashingThreshold(1<<20),
			vortex.PremarkAsSelfRecursed(),
		),
		// logdata.Log("pre-recursion.post-vortex-4"),
	}

	// This is the compilation suite in use for the full prover *after* the
	// recursion step.
	fullSecondCompilationSuite = CompilationSuite{
		cleanup.CleanUp,
		poseidon2.CompilePoseidon2,
		compiler.Arcane(
			compiler.WithTargetColSize(1<<22),
			compiler.WithStitcherMinSize(16),
			// compiler.WithDebugMode("final-compiler-step-0"),
		),
		vortex.Compile(
			2, false,
			vortex.ForceNumOpenedColumns(256),
			vortex.WithSISParams(&sisInstance),
		),
		// logdata.Log("post-recursion.post-vortex-2"),

		// Second round of self-recursion
		selfrecursion.SelfRecurse,
		// logdata.Log("post-recursion.post-selfrecursion-2"),
		cleanup.CleanUp,
		poseidon2.CompilePoseidon2,
		compiler.Arcane(
			compiler.WithTargetColSize(1<<17),
			compiler.WithStitcherMinSize(16),
			// compiler.WithDebugMode("final-compiler-step-1"),
		),
		vortex.Compile(
			16, false,
			vortex.ForceNumOpenedColumns(64),
			vortex.WithSISParams(&sisInstance),
		),

		// Fourth round of self-recursion
		// logdata.Log("post-recursion.post-vortex-3"),
		selfrecursion.SelfRecurse,
		// logdata.Log("post-recursion.post-selfrecursion-3"),
		cleanup.CleanUp,
		poseidon2.CompilePoseidon2,
		compiler.Arcane(
			compiler.WithTargetColSize(1<<12),
			compiler.WithStitcherMinSize(16),
			// compiler.WithDebugMode("final-compiler-step-2"),
		),
		vortex.Compile(16, true, vortex.ForceNumOpenedColumns(64)),
	}
)

// FullZkEvm compiles the full prover zkEVM. It memoizes the results and
// returns it for all the subsequent calls. That is, it should not be called
// twice with different configuration parameters as it will always return the
// instance compiled with the parameters it received the first time. This
// behavior is motivated by the fact that the compilation process takes time
// and we don't want to spend the compilation time twice, plus in practice we
// won't need to call it with different configuration parameters.
func FullZkEvm(tl *config.TracesLimits, cfg *config.Config) *ZkEvm {

	onceFullZkEvm.Do(func() {
		// Initialize the Full zkEVM arithmetization
		fullZkEvm = FullZKEVMWithSuite(tl, cfg, fullInitialCompilationSuite, &fullSecondCompilationSuite)
	})

	return fullZkEvm
}

// FullZkEvmLarge is similar to FullZkEvm but uses the large compilation suite
// with doubled TargetColSize values to reduce constraint count.
func FullZkEvmLarge(tl *config.TracesLimits, cfg *config.Config) *ZkEvm {

	onceFullZkEvmLarge.Do(func() {
		// Initialize the Full zkEVM arithmetization with large compilation suite
		fullZkEvmLarge = FullZKEVMWithSuite(tl, cfg, fullInitialCompilationSuiteLarge, &fullSecondCompilationSuite)
	})

	return fullZkEvmLarge
}

func FullZkEVMCheckOnly(tl *config.TracesLimits, cfg *config.Config) *ZkEvm {

	onceFullZkEvmCheckOnly.Do(func() {
		// Initialize the Full zkEVM arithmetization
		fullZkEvmCheckOnly = FullZKEVMWithSuite(tl, cfg, dummyCompilationSuite, nil)
	})

	return fullZkEvmCheckOnly
}

// FullZKEVMWithSuite returns a compiled zkEVM with the given compilation suite.
// It can be used to benchmark the compilation time of the zkEVM and helps with
// performance optimization.
func FullZKEVMWithSuite(
	tl *config.TracesLimits,
	cfg *config.Config,
	preRecursionSuite CompilationSuite,
	postRecursionSuite *CompilationSuite,
) *ZkEvm {

	// @Alex: only set mandatory parameters here. aka, the one that are not
	// actually feature-gated.
	settings := Settings{
		PreRecursionCompilationSuite:  preRecursionSuite,
		PostRecursionCompilationSuite: postRecursionSuite,
		Arithmetization: arithmetization.Settings{
			Limits:                   tl,
			OptimisationLevel:        &mir.DEFAULT_OPTIMISATION_LEVEL,
			IgnoreCompatibilityCheck: &cfg.Execution.IgnoreCompatibilityCheck,
		},
		Metadata: wizard.VersionMetadata{
			Title:   "linea/evm-execution/full",
			Version: "beta-v1",
		},
	}

	// Initialize the Full zkEVM arithmetization
	fullZkEvm = NewZkEVM(settings)
	return fullZkEvm
}
