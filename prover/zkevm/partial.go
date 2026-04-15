package zkevm

import (
	"sync"

	"github.com/consensys/linea-monorepo/prover/protocol/compiler"
	"github.com/consensys/linea-monorepo/prover/protocol/compiler/vortex"
)

var (
	partialZkEvm     *ZkEvm
	oncePartialZkEvm = sync.Once{}

	partialCompilationSuite = CompilationSuite{
		compiler.Arcane(compiler.WithTargetColSize(1 << 17)),
		vortex.Compile(2, false, vortex.WithOptionalSISHashingThreshold(16)),
	}
)
