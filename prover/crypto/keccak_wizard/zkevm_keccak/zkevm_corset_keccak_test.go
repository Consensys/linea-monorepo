//go:build !nocorset

package zkevm_keccak_test

import (
	"path"
	"testing"

	"github.com/consensys/accelerated-crypto-monorepo/backend/files"
	"github.com/consensys/accelerated-crypto-monorepo/crypto/keccak_wizard/zkevm_keccak"
	"github.com/consensys/accelerated-crypto-monorepo/maths/common/smartvectors"
	"github.com/consensys/accelerated-crypto-monorepo/maths/field"
	"github.com/consensys/accelerated-crypto-monorepo/protocol/compiler/dummy"
	"github.com/consensys/accelerated-crypto-monorepo/protocol/wizard"
	"github.com/consensys/accelerated-crypto-monorepo/utils"
	"github.com/consensys/accelerated-crypto-monorepo/zkevm"
	"github.com/stretchr/testify/require"
)

const INPUT_FILEPATH = "./tr_1687504256.gz"

//const INPUT_FILEPATH = "./tr_1687854804.gz"

func TestFetchDataTxRLP(t *testing.T) {
	//logrus.SetLevel(logrus.TraceLevel)

	/*
		numPerm : the number of permutation supported by KeccakFModule
			size of the KeccakFModule is numPerm*32
	*/

	test := []struct {
		sizeZkModule int
		numPerm      int
		shouldPass   bool
	}{
		{65536, 1024, true}, //should pass
		//{8192, 32, false}, // should fail, //Ignored, since it is too computation intensive for CI

	}

	for _, tcase := range test {

		sizeZkModule := tcase.sizeZkModule
		numPerm := tcase.numPerm

		define := func(build *wizard.Builder) {
			RegisterCommitModule(build, sizeZkModule, zkevm_keccak.AttTXRLP, zkevm_keccak.TXRLP)
			zkevm_keccak.RegisterKeccak(build.CompiledIOP, 0, numPerm)
		}
		compiled := wizard.Compile(
			define,
			dummy.Compile,
		)

		runtest := func() {

			filepath := path.Join(INPUT_FILEPATH)
			f := files.MustReadCompressed(filepath)

			// assignment from corset
			proof := wizard.Prove(compiled, func(run *wizard.ProverRuntime) {
				zkevm.AssignFromCorset(f, run)
			})
			err := wizard.Verify(compiled, proof)

			if err != nil {
				utils.Panic("the proof did not pass : %v", err)
			}

			f.Close()
		}

		if tcase.shouldPass {
			runtest()
		} else {
			require.Panics(t, runtest)
		}
	}

}

func RegisterCommitModule(build *wizard.Builder, size int, module zkevm_keccak.ZkModuleAtt, moduleName zkevm_keccak.Module) {
	build.RegisterCommit(module.HashNum, size)
	build.RegisterCommit(module.INDEX, size)
	build.RegisterCommit(module.LIMB, size)
	build.RegisterCommit(module.Nbytes, size)
	if moduleName == zkevm_keccak.TXRLP {
		build.RegisterCommit(module.LX, size)
		build.RegisterCommit(module.LC, size)
	}
}

// assignment from a random table
func AssignZkModuleCol(run *wizard.ProverRuntime, size int, module zkevm_keccak.ZkModuleAtt) {
	table := TableForTest(size)
	limb := table.InputTable.Limb
	u := make([]field.Element, len(limb))

	for i := range limb {
		u[i].SetBytes(limb[i][:])

	}
	v := smartvectors.NewRegular(u)
	run.AssignColumn(module.HashNum, smartvectors.ForTest(table.InputTable.HashNum...))
	run.AssignColumn(module.INDEX, smartvectors.ForTest(table.InputTable.Index...))
	run.AssignColumn(module.LIMB, v.DeepCopy())
	run.AssignColumn(module.Nbytes, smartvectors.ForTest(table.InputTable.Nbytes...))

}
