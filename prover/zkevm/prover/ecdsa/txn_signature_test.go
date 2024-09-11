package ecdsa

import (
	"testing"

	"github.com/consensys/linea-monorepo/prover/protocol/compiler/dummy"
	"github.com/consensys/linea-monorepo/prover/protocol/wizard"
	"github.com/consensys/linea-monorepo/prover/zkevm/prover/common"
	"github.com/consensys/linea-monorepo/prover/zkevm/prover/hash/generic/testdata"
	"github.com/consensys/linea-monorepo/prover/zkevm/prover/hash/keccak"
	"github.com/stretchr/testify/assert"
)

func TestTxnSignature(t *testing.T) {

	// for this test they are the actual values
	settings := &Settings{
		MaxNbEcRecover: 5,
		MaxNbTx:        5,
	}
	size := settings.sizeAntichamber()
	m := &keccak.KeccakSingleProvider{}
	var txSign *txSignature

	compiled := wizard.Compile(func(b *wizard.Builder) {
		var (
			comp      = b.CompiledIOP
			createCol = common.CreateColFn(comp, "TESTING_TxSignature", size)

			txSignInputs = txSignatureInputs{
				ac: &antichamber{
					IsFetching: createCol("Is_Fetching"),
					IsActive:   createCol("Is_Active"),
					Source:     createCol("Source"),
					size:       size,
					Inputs:     &antichamberInput{settings: settings},
				},
				RlpTxn: testdata.CreateGenDataModule(comp, "RLP_TXN", 128),
			}
		)

		txSign = newTxSignatures(comp, txSignInputs)

		// check the keccak consistency over the provider.
		provider := txSign.provider
		keccakInp := keccak.KeccakSingleProviderInput{
			Provider:      provider,
			MaxNumKeccakF: 16,
		}
		m = keccak.NewKeccakSingleProvider(comp, keccakInp)
	}, dummy.Compile)

	proof := wizard.Prove(compiled, func(run *wizard.ProverRuntime) {

		// assign txSignInputs
		txSign.assigntxSignInputs(run, rlpTxnTest)
		txSign.assignTxSignature(run, settings.MaxNbEcRecover)
		m.Run(run)
	})
	assert.NoError(t, wizard.Verify(compiled, proof))
}

var rlpTxnTest = makeTestCase{
	HashNum: []int{1, 1, 1, 1, 1, 1, 2, 2, 2, 2, 2, 2, 3, 3, 3, 3, 3, 3, 4, 4, 4, 4, 4, 4, 5, 5, 5, 5, 5, 5},
	ToHash:  []int{1, 1, 0, 1, 0, 0, 1, 1, 1, 1, 0, 0, 1, 0, 1, 1, 0, 0, 1, 1, 1, 1, 0, 0, 1, 0, 1, 1, 0, 0},
}

func (txSign txSignature) assigntxSignInputs(run *wizard.ProverRuntime, c makeTestCase) {

	var (
		nbEcRec    = txSign.Inputs.ac.Inputs.settings.MaxNbEcRecover
		nbTxn      = txSign.Inputs.ac.Inputs.settings.MaxNbTx
		isFetching = common.NewVectorBuilder(txSign.Inputs.ac.IsFetching)
		isActive   = common.NewVectorBuilder(txSign.Inputs.ac.IsActive)
		source     = common.NewVectorBuilder(txSign.Inputs.ac.Source)
	)

	// assign rlpTxn
	testdata.GenerateAndAssignGenDataModule(run, &txSign.Inputs.RlpTxn, c.HashNum, c.ToHash, true)

	for i := 0; i < nbEcRec; i++ {
		for j := 0; j < nbRowsPerEcRecFetching; j++ {
			isFetching.PushInt(1)
		}
		for j := nbRowsPerEcRecFetching; j < nbRowsPerEcRec; j++ {
			isFetching.PushInt(0)
		}
		for j := 0; j < nbRowsPerEcRec; j++ {
			source.PushInt(0)
			isActive.PushInt(1)
		}
	}

	for i := 0; i < nbTxn; i++ {
		for j := 0; j < nbRowsPerTxSign; j++ {
			source.PushInt(1)
			isActive.PushInt(1)
		}
		isFetching.PushInt(1)
		for j := 1; j < nbRowsPerTxSign; j++ {
			isFetching.PushInt(0)
		}
	}
	isFetching.PadAndAssign(run)
	source.PadAndAssign(run)
	isActive.PadAndAssign(run)
}
