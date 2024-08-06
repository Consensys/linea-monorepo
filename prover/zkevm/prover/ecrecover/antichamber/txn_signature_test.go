package antichamber

import (
	"testing"

	"github.com/consensys/zkevm-monorepo/prover/protocol/compiler/dummy"
	"github.com/consensys/zkevm-monorepo/prover/protocol/wizard"
	"github.com/consensys/zkevm-monorepo/prover/zkevm/prover/hash/generic"
	"github.com/consensys/zkevm-monorepo/prover/zkevm/prover/hash/generic/testdata"
	"github.com/consensys/zkevm-monorepo/prover/zkevm/prover/hash/keccak"
	"github.com/stretchr/testify/assert"
)

func TestTxnSignature(t *testing.T) {
	c := RLP_TXN_test
	limits := &Limits{
		MaxNbEcRecover: 5,
		MaxNbTx:        5,
	}
	gbmSize := 256
	size := limits.sizeAntichamber()
	m := &keccak.KeccakSingleProvider{}

	ac := Antichamber{Limits: limits}
	var txSign *txSignature
	rlpTxn := generic.GenDataModule{}

	nbKeccakF := ac.nbKeccakF(3)

	compiled := wizard.Compile(func(b *wizard.Builder) {
		comp := b.CompiledIOP
		rlpTxn = testdata.CreateGenDataModule(comp, "RLP_TXN", gbmSize)
		txSign = newTxSignatures(comp, rlpTxn, size)
		provider := txSign.GetProvider(comp, rlpTxn)

		keccakInp := keccak.KeccakSingleProviderInput{
			Provider:      provider,
			MaxNumKeccakF: nbKeccakF,
		}
		m = keccak.NewKeccakSingleProvider(comp, keccakInp)
	}, dummy.Compile)

	proof := wizard.Prove(compiled, func(run *wizard.ProverRuntime) {
		testdata.GenerateAndAssignGenDataModule(run, &rlpTxn, c.HashNum, c.ToHash, true)
		txSign.assignTxSignature(run, rlpTxn, limits.MaxNbEcRecover, size)
		m.Run(run)
	})
	assert.NoError(t, wizard.Verify(compiled, proof))
}

var RLP_TXN_test = makeTestCase{
	HashNum: []int{1, 1, 1, 1, 1, 1, 2, 2, 2, 2, 2, 2, 3, 3, 3, 3, 3, 3, 4, 4, 4, 4, 4, 4, 5, 5, 5, 5, 5, 5},
	ToHash:  []int{1, 1, 0, 1, 0, 0, 1, 1, 1, 1, 0, 0, 1, 0, 1, 1, 0, 0, 1, 1, 1, 1, 0, 0, 1, 0, 1, 1, 0, 0},
}
