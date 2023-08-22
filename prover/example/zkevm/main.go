package main

import (
	"path"

	"github.com/consensys/accelerated-crypto-monorepo/backend/files"
	"github.com/consensys/accelerated-crypto-monorepo/protocol/compiler"
	"github.com/consensys/accelerated-crypto-monorepo/protocol/compiler/logdata"
	"github.com/consensys/accelerated-crypto-monorepo/protocol/compiler/vortex"
	"github.com/consensys/accelerated-crypto-monorepo/protocol/wizard"
	"github.com/consensys/accelerated-crypto-monorepo/utils/profiling"

	"github.com/consensys/accelerated-crypto-monorepo/zkevm"
	"github.com/consensys/accelerated-crypto-monorepo/zkevm/define"
)

const INPUT_FILEPATH = "./backend/testing/a/raw-trace.gz"

func runTest() {

	profiling.SKIP_PROFILING = true
	filepath := path.Join(INPUT_FILEPATH)
	f := files.MustReadCompressed(filepath)

	compiled := wizard.Compile(
		zkevm.WrapDefine(define.ZkEVMDefine),
		compiler.Arcane(1<<15, 1<<17),
		vortex.Compile(2, vortex.WithDryThreshold(16)),
		logdata.Log("post-vortex-1"),
	)

	var proof wizard.Proof
	profiling.ProfileTrace("zkevm-example-prover", true, false, func() {
		proof = wizard.Prove(compiled, func(run *wizard.ProverRuntime) { zkevm.AssignFromCorset(f, run) })
	})

	err := wizard.Verify(compiled, proof)
	if err != nil {
		panic(err)
	}
}

func main() {
	runTest()
}
