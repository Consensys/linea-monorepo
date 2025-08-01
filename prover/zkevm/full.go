package zkevm

import (
	"sync"

	"github.com/consensys/go-corset/pkg/mir"
	"github.com/consensys/linea-monorepo/prover/config"
	"github.com/consensys/linea-monorepo/prover/crypto/ringsis"
	"github.com/consensys/linea-monorepo/prover/protocol/compiler"
	"github.com/consensys/linea-monorepo/prover/protocol/compiler/cleanup"
	"github.com/consensys/linea-monorepo/prover/protocol/compiler/dummy"
	"github.com/consensys/linea-monorepo/prover/protocol/compiler/mimc"
	"github.com/consensys/linea-monorepo/prover/protocol/compiler/plonkinwizard"
	"github.com/consensys/linea-monorepo/prover/protocol/compiler/selfrecursion"
	"github.com/consensys/linea-monorepo/prover/protocol/compiler/vortex"
	"github.com/consensys/linea-monorepo/prover/protocol/wizard"
	"github.com/consensys/linea-monorepo/prover/utils"
	"github.com/consensys/linea-monorepo/prover/zkevm/arithmetization"
	"github.com/consensys/linea-monorepo/prover/zkevm/prover/ecarith"
	"github.com/consensys/linea-monorepo/prover/zkevm/prover/ecdsa"
	"github.com/consensys/linea-monorepo/prover/zkevm/prover/ecpair"
	"github.com/consensys/linea-monorepo/prover/zkevm/prover/hash/keccak"
	"github.com/consensys/linea-monorepo/prover/zkevm/prover/hash/sha2"
	"github.com/consensys/linea-monorepo/prover/zkevm/prover/modexp"
	"github.com/consensys/linea-monorepo/prover/zkevm/prover/statemanager"
	"github.com/consensys/linea-monorepo/prover/zkevm/prover/statemanager/accumulator"
)

const (
	NbInputPerInstancesEcAdd           = 28
	NbInputPerInstancesEcMul           = 6
	NbInputPerInstanceEcPairMillerLoop = 1
	NbInputPerInstanceEcPairFinalExp   = 1
	NbInputPerInstanceEcPairG2Check    = 1
	NbInputPerInstanceSha2Block        = 3
	NbInputPerInstanceModexp256        = 10
	NbInputPerInstanceModexp4096       = 1
	NbInputPerInstanceEcdsa            = 4
)

var (
	fullZkEvm               *ZkEvm
	fullZkEvmCheckOnly      *ZkEvm
	fullZkEvmSetup          *ZkEvm
	fullZkEvmSetupLarge     *ZkEvm
	onceFullZkEvm           = sync.Once{}
	onceFullZkEvmCheckOnly  = sync.Once{}
	onceFullZkEvmSetup      = sync.Once{}
	onceFullZkEvmSetupLarge = sync.Once{}

	// This is the SIS instance, that has been found to minimize the overhead of
	// recursion. It is changed w.r.t to the estimated because the estimated one
	// allows for 10 bits limbs instead of just 8. But since the current state
	// of the self-recursion currently relies on the number of limbs to be a
	// power of two, we go with this one although it overshoots our security
	// level target.
	sisInstance = ringsis.Params{LogTwoBound: 16, LogTwoDegree: 6}

	dummyCompilationSuite = CompilationSuite{dummy.CompileAtProverLvl()}

	// This is the compilation suite in use for the full prover
	fullCompilationSuite = CompilationSuite{
		// logdata.Log("initial-wizard"),
		mimc.CompileMiMC,
		plonkinwizard.Compile,
		compiler.Arcane(compiler.WithTargetColSize(1 << 19)),
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
		compiler.Arcane(compiler.WithTargetColSize(1 << 18)),
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
		compiler.Arcane(compiler.WithTargetColSize(1 << 16)),
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
		compiler.Arcane(compiler.WithTargetColSize(1 << 13)),
		vortex.Compile(
			8,
			vortex.ForceNumOpenedColumns(64),
			vortex.WithOptionalSISHashingThreshold(1<<20),
		),
		// logdata.Log("post-vortex-4"),
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
		fullZkEvm = FullZKEVMWithSuite(tl, fullCompilationSuite, cfg)
	})

	return fullZkEvm
}

func FullZkEVMCheckOnly(tl *config.TracesLimits, cfg *config.Config) *ZkEvm {

	onceFullZkEvmCheckOnly.Do(func() {
		// Initialize the Full zkEVM arithmetization
		fullZkEvmCheckOnly = FullZKEVMWithSuite(tl, dummyCompilationSuite, cfg)
	})

	return fullZkEvmCheckOnly
}

func FullZkEvmSetup(tl *config.TracesLimits, cfg *config.Config) *ZkEvm {
	onceFullZkEvmSetup.Do(func() {
		fullZkEvmSetup = FullZKEVMWithSuite(tl, fullCompilationSuite, cfg)
	})
	return fullZkEvmSetup
}

func FullZkEvmSetupLarge(tl *config.TracesLimits, cfg *config.Config) *ZkEvm {
	onceFullZkEvmSetupLarge.Do(func() {
		fullZkEvmSetupLarge = FullZKEVMWithSuite(tl, fullCompilationSuite, cfg)
	})
	return fullZkEvmSetupLarge
}

// FullZKEVMWithSuite returns a compiled zkEVM with the given compilation suite.
// It can be used to benchmark the compilation time of the zkEVM and helps with
// performance optimization.
func FullZKEVMWithSuite(tl *config.TracesLimits, suite CompilationSuite, cfg *config.Config) *ZkEvm {

	// @Alex: only set mandatory parameters here. aka, the one that are not
	// actually feature-gated.
	settings := Settings{
		CompilationSuite: suite,
		Arithmetization: arithmetization.Settings{
			Limits:                   tl,
			OptimisationLevel:        &mir.DEFAULT_OPTIMISATION_LEVEL,
			IgnoreCompatibilityCheck: &cfg.Execution.IgnoreCompatibilityCheck,
		},
		Statemanager: statemanager.Settings{
			AccSettings: accumulator.Settings{
				MaxNumProofs:    tl.ShomeiMerkleProofs,
				Name:            "SM_ACCUMULATOR",
				MerkleTreeDepth: 40,
			},
			MiMCCodeHashSize: tl.Rom,
		},
		Metadata: wizard.VersionMetadata{
			Title:   "linea/evm-execution/full",
			Version: "beta-v1",
		},
		Keccak: keccak.Settings{
			MaxNumKeccakf: tl.BlockKeccak,
		},
		Ecdsa: ecdsa.Settings{
			MaxNbEcRecover:     tl.PrecompileEcrecoverEffectiveCalls,
			MaxNbTx:            tl.BlockTransactions,
			NbInputInstance:    NbInputPerInstanceEcdsa,
			NbCircuitInstances: utils.DivCeil(tl.PrecompileEcrecoverEffectiveCalls+tl.BlockTransactions, NbInputPerInstanceEcdsa),
		},
		Modexp: modexp.Settings{
			MaxNbInstance256:                tl.PrecompileModexpEffectiveCalls,
			MaxNbInstance4096:               tl.PrecompileModexpEffectiveCalls4096,
			NbInstancesPerCircuitModexp256:  NbInputPerInstanceModexp256,
			NbInstancesPerCircuitModexp4096: NbInputPerInstanceModexp4096,
		},
		Ecadd: ecarith.Limits{
			// 14 was found the right number to have just under 2^19 constraints
			// per circuit.
			NbCircuitInstances: utils.DivCeil(tl.PrecompileEcaddEffectiveCalls, NbInputPerInstancesEcAdd),
			NbInputInstances:   NbInputPerInstancesEcAdd,
		},
		Ecmul: ecarith.Limits{
			NbCircuitInstances: utils.DivCeil(tl.PrecompileEcmulEffectiveCalls, NbInputPerInstancesEcMul),
			NbInputInstances:   NbInputPerInstancesEcMul,
		},
		Ecpair: ecpair.Limits{
			NbMillerLoopInputInstances:   NbInputPerInstanceEcPairMillerLoop,
			NbMillerLoopCircuits:         utils.DivCeil(tl.PrecompileEcpairingMillerLoops, NbInputPerInstanceEcPairMillerLoop),
			NbFinalExpInputInstances:     NbInputPerInstanceEcPairFinalExp,
			NbFinalExpCircuits:           utils.DivCeil(tl.PrecompileEcpairingEffectiveCalls, NbInputPerInstanceEcPairFinalExp),
			NbG2MembershipInputInstances: NbInputPerInstanceEcPairG2Check,
			NbG2MembershipCircuits:       utils.DivCeil(tl.PrecompileEcpairingG2MembershipCalls, NbInputPerInstanceEcPairG2Check),
		},
		Sha2: sha2.Settings{
			MaxNumSha2F:                    tl.PrecompileSha2Blocks,
			NbInstancesPerCircuitSha2Block: NbInputPerInstanceSha2Block,
		},
	}

	// Initialize the Full zkEVM arithmetization
	fullZkEvm = NewZkEVM(settings)
	return fullZkEvm
}
