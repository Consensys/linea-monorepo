package ecdsa

import (
	"testing"

	"github.com/consensys/zkevm-monorepo/prover/protocol/compiler/dummy"
	"github.com/consensys/zkevm-monorepo/prover/protocol/wizard"
	"github.com/consensys/zkevm-monorepo/prover/utils"
	"github.com/consensys/zkevm-monorepo/prover/zkevm/prover/hash/generic"
	"github.com/consensys/zkevm-monorepo/prover/zkevm/prover/hash/generic/testdata"
	"github.com/consensys/zkevm-monorepo/prover/zkevm/prover/hash/keccak"
	"github.com/stretchr/testify/assert"
)

func TestAddress(t *testing.T) {
	c := testCases
	limits := &Settings{
		MaxNbEcRecover: 1,
		MaxNbTx:        4,
	}
	ac := &antichamber{Inputs: &antichamberInput{settings: limits}}
	var addr *Addresses
	var uaGnark *UnalignedGnarkData
	var ecRec *EcRecover
	var td *txnData
	gbmGnark := generic.GenDataModule{}
	m := &keccak.KeccakSingleProvider{}

	size := limits.sizeAntichamber()

	nbKeccakF := ac.Inputs.settings.nbKeccakF(8) // if each txn has 8 blocks
	nbRowsPerTxInTxnData := 9

	sizeTxnData := utils.NextPowerOfTwo(limits.MaxNbTx * nbRowsPerTxInTxnData)

	compiled := wizard.Compile(func(b *wizard.Builder) {
		comp := b.CompiledIOP

		// generate a gbm and use it to represent gnark-columns
		gbmGnark = testdata.CreateGenDataModule(comp, "UnGNARK", size)
		ac = &antichamber{
			Inputs: &antichamberInput{settings: limits},
			ID:     gbmGnark.HashNum,
		}
		uaGnark = &UnalignedGnarkData{
			GnarkData:           gbmGnark.Limb,
			GnarkPublicKeyIndex: gbmGnark.Index,
			IsPublicKey:         gbmGnark.ToHash,
		}
		ac.UnalignedGnarkData = uaGnark

		// commit to txnData and ecRecover
		td, ecRec = commitEcRecTxnData(comp, sizeTxnData, size, ac)

		// native columns and  constraints
		addr = newAddress(comp, size, ecRec, ac, td)

		// define keccak (columns and constraints)
		keccakInp := keccak.KeccakSingleProviderInput{
			Provider:      addr.provider,
			MaxNumKeccakF: nbKeccakF,
		}
		m = keccak.NewKeccakSingleProvider(comp, keccakInp)

	}, dummy.Compile)

	proof := wizard.Prove(compiled, func(run *wizard.ProverRuntime) {

		testdata.GenerateAndAssignGenDataModule(run, &gbmGnark, c.HashNum, c.ToHash, false)
		// it assign mock data to EcRec and txn_data
		AssignEcRecTxnData(run, gbmGnark, limits.MaxNbEcRecover, limits.MaxNbTx, sizeTxnData, size, td, ecRec, ac)

		// assign address columns
		addr.assignAddress(run, limits.MaxNbEcRecover, size, ac, ecRec, uaGnark, td)

		// assign keccak columns via provider that is embedded in the receiver
		m.Run(run)
	})
	assert.NoError(t, wizard.Verify(compiled, proof))
}

type makeTestCase struct {
	HashNum []int
	ToHash  []int
}

var testCases = makeTestCase{
	HashNum: []int{1, 1, 1, 1, 1, 1, 2, 2, 2, 2, 2, 2, 3, 3, 3, 3, 3, 3, 4, 4, 4, 4, 4, 4, 5, 5, 5, 5, 5, 5},
	ToHash:  []int{1, 1, 1, 1, 1, 0, 1, 1, 1, 1, 0, 0, 1, 1, 1, 1, 0, 0, 1, 1, 1, 1, 0, 0, 1, 1, 1, 1, 0, 0},
}
