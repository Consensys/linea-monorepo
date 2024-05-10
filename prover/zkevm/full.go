package zkevm

import (
	"sync"

	"github.com/consensys/zkevm-monorepo/prover/config"
	"github.com/consensys/zkevm-monorepo/prover/crypto/ringsis"
	"github.com/consensys/zkevm-monorepo/prover/protocol/compiler"
	"github.com/consensys/zkevm-monorepo/prover/protocol/compiler/cleanup"
	"github.com/consensys/zkevm-monorepo/prover/protocol/compiler/mimc"
	"github.com/consensys/zkevm-monorepo/prover/protocol/compiler/selfrecursion"
	"github.com/consensys/zkevm-monorepo/prover/protocol/compiler/vortex"
	"github.com/consensys/zkevm-monorepo/prover/zkevm/arithmetization"
	"github.com/consensys/zkevm-monorepo/prover/zkevm/prover/statemanager"
)

const (
	// @TODO: the keccak limits are hardcoded currently, in the future we should
	// instead take the limits from the trace limit file. Note, neither of these
	// limits are actually enforced by the prover as the keccak module is not
	// connected to the rest of the arithmetization. Thus, it is easy to just
	// ignore the overflowing keccaks and the state merkle-proofs.
	keccakLimit      = 1 << 13
	merkleProofLimit = 1 << 13
)

var (
	fullZkEvm     *ZkEvm
	onceFullZkEvm = sync.Once{}

	// This is the SIS instance, that has been found to minimize the overhead of
	// recursion. It is changed w.r.t to the estimated because the estimated one
	// allows for 10 bits limbs instead of just 8. But since the current state
	// of the self-recursion currently relies on the number of limbs to be a
	// power of two, we go with this one although it overshoots our security
	// level target.
	sisInstance = ringsis.Params{LogTwoBound: 8, LogTwoDegree: 6}

	// This is the compilation suite in use for the full prover
	fullCompilationSuite = compilationSuite{
		// logdata.Log("initial-wizard"),
		mimc.CompileMiMC,
		compiler.Arcane(1<<10, 1<<19, true),
		vortex.Compile(
			2,
			vortex.ForceNumOpenedColumns(256),
			vortex.WithSISParams(&sisInstance),
		),
		// logdata.Log("post-vortex-1"),

		// First round of self-recursion
		selfrecursion.SelfRecurse,
		// logdata.Log("post-selfrecursion-1"),
		cleanup.CleanUp,
		mimc.CompileMiMC,
		compiler.Arcane(1<<10, 1<<18, true),
		vortex.Compile(
			2,
			vortex.ForceNumOpenedColumns(256),
			vortex.WithSISParams(&sisInstance),
		),
		// logdata.Log("post-vortex-2"),

		// Second round of self-recursion
		selfrecursion.SelfRecurse,
		// logdata.Log("post-selfrecursion-2"),
		cleanup.CleanUp,
		mimc.CompileMiMC,
		compiler.Arcane(1<<10, 1<<16, true),
		vortex.Compile(
			8,
			vortex.ForceNumOpenedColumns(64),
			vortex.WithSISParams(&sisInstance),
		),

		// Fourth round of self-recursion
		// logdata.Log("post-vortex-3"),
		selfrecursion.SelfRecurse,
		// logdata.Log("post-selfrecursion-3"),
		cleanup.CleanUp,
		mimc.CompileMiMC,
		compiler.Arcane(1<<10, 1<<13, true),
		vortex.Compile(
			8,
			vortex.ForceNumOpenedColumns(64),
			vortex.ReplaceSisByMimc(),
		),
		// logdata.Log("post-vortex-4"),
	}
)

// FullZkEvm compiles the full prover zk-evm. It memoizes the results and
// returns it for all the subsequent calls. That is, it should not be called
// twice with different configuration parameters as it will always return the
// instance compiled with the parameters it received the first time. This
// behaviour is motivated by the fact that the compilation process takes time
// and we don't want to spend the compilation time twice, plus in practice we
// won't need to call it with different configuration parameters.
func FullZkEvm(feat *config.Features, tl *config.TracesLimits) *ZkEvm {

	onceFullZkEvm.Do(func() {

		// @Alex: only set mandatory parameters here. aka, the one that are not
		// actually feature-gated.
		settings := Settings{
			Arithmetization: arithmetization.Settings{
				Traces: tl,
			},
			Statemanager: statemanager.Settings{
				MaxMerkleProof: merkleProofLimit,
			},
			// The compilation suite itself is hard-coded and reflects the
			// actual full proof system.
			CompilationSuite: fullCompilationSuite,
		}

		// Keccak is feature-gated although we plan to make it mandatory in the
		// future.
		if feat.WithKeccak {
			settings.Keccak.Enabled = true
			settings.Keccak.MaxNumKeccakf = keccakLimit
		}

		// Initialize the Full zkEVM arithmetization
		fullZkEvm = NewZkEVM(settings).Compile(fullCompilationSuite)

	})

	return fullZkEvm
}
