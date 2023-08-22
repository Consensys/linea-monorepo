package main

import (
	"github.com/consensys/accelerated-crypto-monorepo/protocol/compiler"
	"github.com/consensys/accelerated-crypto-monorepo/protocol/compiler/logdata"
	"github.com/consensys/accelerated-crypto-monorepo/protocol/compiler/selfrecursion"
	"github.com/consensys/accelerated-crypto-monorepo/protocol/compiler/vortex"
	"github.com/consensys/accelerated-crypto-monorepo/protocol/wizard"
	"github.com/consensys/accelerated-crypto-monorepo/zkevm"
	"github.com/consensys/accelerated-crypto-monorepo/zkevm/define"
)

func main() {
	runMeasure(1<<17, 1<<17, 1<<16, 1<<16, 1<<16)
}

func runMeasure(splitSizes ...int) {

	suite := []func(*wizard.CompiledIOP){
		logdata.Log("pre-compilation"),
	}

	for i := range splitSizes {
		if i > 0 {
			suite = append(suite,
				selfrecursion.SelfRecurse)
		}
		suite = append(suite,
			compiler.Arcane(splitSizes[i]/4, splitSizes[i], true),
			logdata.Log("post-arcane"),
			vortex.Compile(2),
		)
	}

	suite = append(suite, logdata.Log("FINAL"))

	_ = wizard.Compile(zkevm.WrapDefine(define.ZkEVMDefine), suite...)

}
