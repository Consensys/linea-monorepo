package prover

import (
	"sync"

	"github.com/consensys/accelerated-crypto-monorepo/crypto/ringsis"
	"github.com/consensys/accelerated-crypto-monorepo/protocol/compiler"
	"github.com/consensys/accelerated-crypto-monorepo/protocol/compiler/cleanup"
	"github.com/consensys/accelerated-crypto-monorepo/protocol/compiler/dummy"
	"github.com/consensys/accelerated-crypto-monorepo/protocol/compiler/logdata"
	"github.com/consensys/accelerated-crypto-monorepo/protocol/compiler/mimc"
	"github.com/consensys/accelerated-crypto-monorepo/protocol/compiler/selfrecursion/selfrecursionwithmerkle"
	"github.com/consensys/accelerated-crypto-monorepo/protocol/compiler/vortex"
	"github.com/consensys/accelerated-crypto-monorepo/protocol/wizard"
	"github.com/consensys/accelerated-crypto-monorepo/zkevm"
	"github.com/consensys/accelerated-crypto-monorepo/zkevm/define"
)

// Identifying the prover's modes
const (
	NOTRACE    string = "notrace"
	LIGHT      string = "light"
	FULL       string = "full"
	FULL_LARGE string = "full-large"
	CHECKER    string = "checker"
)

// The compiled circuit to use for everything. The compilation happens
// when we load the runtime.
var devLightIOP *wizard.CompiledIOP
var onceDevLight sync.Once = sync.Once{}

// The complete polynomial-IOP
var fatIOP *wizard.CompiledIOP
var onceFat sync.Once = sync.Once{}

// The checker (only verifies the constraints)
var checker *wizard.CompiledIOP
var onceChecker sync.Once = sync.Once{}

// Initializes the dev-light iop (only once) and returns the result
func getDevLightIOP() *wizard.CompiledIOP {
	onceDevLight.Do(func() {
		devLightIOP = wizard.Compile(
			zkevm.WrapDefine(define.ZkEVMDefine, zkevm.NumColLimit(10)),
			compiler.Arcane(1<<16, 1<<17),
			vortex.Compile(2, vortex.WithDryThreshold(16)),
		)
	})
	return devLightIOP
}

// Initializes the full iop (only once) and returns the result
func GetFullIOP(po *ProverOptions) *wizard.CompiledIOP {
	onceFat.Do(func() {
		// This is the lattice instance that we found to be optimal
		// it changed w.r.t to the estimated because the estimated
		// one allows for 10 bits limbs instead of just 8. But since
		// the current state of the self-recursion currently relies
		// on the number of limbs to be a power of two, we go with
		// this one instead.
		sisInstance := ringsis.Params{
			LogTwoBound_: 8,
			LogTwoDegree: 6,
		}

		fatIOP = wizard.Compile(
			// Registering the  inittial WIZARD protocol representing the zkEVM
			zkevm.WrapDefine(define.ZkEVMDefine, po.AppendToZkEvmOptions()...),
			logdata.Log("initial-wizard"),
			mimc.CompileMiMC,
			compiler.Arcane(1<<10, 1<<19, false),
			vortex.Compile(
				2,
				vortex.ForceNumOpenedColumns(256),
				vortex.WithSISParams(&sisInstance),
				vortex.MerkleMode,
			),
			logdata.Log("post-vortex-1"),

			// First round of self-recursion
			selfrecursionwithmerkle.SelfRecurse,
			logdata.Log("post-selfrecursion-1"),
			cleanup.CleanUp,
			mimc.CompileMiMC,
			compiler.Arcane(1<<10, 1<<18, false),
			vortex.Compile(
				2,
				vortex.ForceNumOpenedColumns(256),
				vortex.WithSISParams(&sisInstance),
				vortex.MerkleMode,
			),
			logdata.Log("post-vortex-2"),

			// Second round of self-recursion
			selfrecursionwithmerkle.SelfRecurse,
			logdata.Log("post-selfrecursion-2"),
			cleanup.CleanUp,
			mimc.CompileMiMC,
			compiler.Arcane(1<<10, 1<<16, false),
			vortex.Compile(
				8,
				vortex.ForceNumOpenedColumns(64),
				vortex.WithSISParams(&sisInstance),
				vortex.MerkleMode,
			),

			// Fourth round of self-recursion
			logdata.Log("post-vortex-3"),
			selfrecursionwithmerkle.SelfRecurse,
			logdata.Log("post-selfrecursion-3"),
			cleanup.CleanUp,
			mimc.CompileMiMC,
			compiler.Arcane(1<<10, 1<<13, false),
			vortex.Compile(
				8,
				vortex.ForceNumOpenedColumns(64),
				vortex.ReplaceSisByMimc(),
				vortex.MerkleMode,
			),
			logdata.Log("post-vortex-4"),
		)
	})
	return fatIOP
}

// Initializes the checker (only once) and returns the result
func getChecker(po *ProverOptions) *wizard.CompiledIOP {
	onceChecker.Do(func() {
		checker = wizard.Compile(
			zkevm.WrapDefine(define.ZkEVMDefine, po.AppendToZkEvmOptions()...),
			dummy.Compile,
		)
	})
	return checker
}

// Return an IOP tailored for the quick check of the
// constraints without any proof computation
func CompiledCheckerIOP(po *ProverOptions) *wizard.CompiledIOP {
	return getChecker(po)
}
