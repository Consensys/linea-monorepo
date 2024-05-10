package zkevm

import (
	"sync"

	"github.com/consensys/zkevm-monorepo/prover/config"
	"github.com/consensys/zkevm-monorepo/prover/protocol/compiler/dummy"
	"github.com/consensys/zkevm-monorepo/prover/zkevm/arithmetization"
)

var (
	checkerZkEvm     *ZkEvm
	onceCheckerZkEvm = sync.Once{}

	checkerCompilationSuite = compilationSuite{
		// The dummy compiler returns the witness as the proof and manually
		// checks it as the verifier. It is essentially the trivial proof
		// system.
		dummy.Compile,
	}
)

// The checker is not meant to generate proofs, it is meant to be used to check
// that the provided prover inputs are correct. It typically is used to audit
// the traces of the arithmetization. Currently, it does not include the keccaks
// nor does it include the state-management checks.
func CheckerZkEvm(tl *config.TracesLimits) *ZkEvm {
	onceCheckerZkEvm.Do(func() {
		settings := Settings{
			Arithmetization: arithmetization.Settings{
				Traces: tl,
			},
		}
		checkerZkEvm = NewZkEVM(settings).Compile(checkerCompilationSuite)
	})
	return checkerZkEvm
}
