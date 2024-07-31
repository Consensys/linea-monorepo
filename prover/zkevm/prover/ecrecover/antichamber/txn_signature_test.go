package antichamber

import (
	"testing"

	"github.com/consensys/linea-monorepo/prover/protocol/compiler/dummy"
	"github.com/consensys/linea-monorepo/prover/protocol/wizard"
	"github.com/consensys/linea-monorepo/prover/zkevm/prover/hash/datatransfer/acc_module"
	"github.com/consensys/linea-monorepo/prover/zkevm/prover/hash/generic"
	"github.com/consensys/linea-monorepo/prover/zkevm/prover/hash/keccak"
	"github.com/stretchr/testify/assert"
)

func TestTxnSignature(t *testing.T) {
	limits := &Limits{
		MaxNbEcRecover: 5,
		MaxNbTx:        5,
	}
	gbmSize := 256
	size := limits.sizeAntichamber()

	m := keccak.Module{}
	ac := Antichamber{Limits: limits}
	var txSign *txSignature
	gbm := generic.GenericByteModule{}

	nbKeccakF := ac.nbKeccakF(3)

	compiled := wizard.Compile(func(b *wizard.Builder) {
		comp := b.CompiledIOP
		gbm = acc_module.CommitGBM(comp, 0, generic.RLP_TXN, gbmSize)
		txSign = newTxSignatures(comp, size)
		provider := txSign.GetProvider(comp)
		m.Define(comp, []generic.GenericByteModule{provider}, nbKeccakF)
	}, dummy.Compile)

	proof := wizard.Prove(compiled, func(run *wizard.ProverRuntime) {
		witSize := gbmSize - gbmSize/15
		acc_module.AssignGBMfromTable(run, &gbm, witSize, limits.MaxNbTx)
		txSign.assignTxSignature(run, limits.MaxNbEcRecover, size)
		m.AssignKeccak(run)
	})
	assert.NoError(t, wizard.Verify(compiled, proof))
}
