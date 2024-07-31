package antichamber

import (
	"testing"

	"github.com/consensys/zkevm-monorepo/prover/protocol/compiler/dummy"
	"github.com/consensys/zkevm-monorepo/prover/protocol/wizard"
	"github.com/consensys/zkevm-monorepo/prover/utils"
	"github.com/consensys/zkevm-monorepo/prover/zkevm/prover/hash/datatransfer/acc_module"
	"github.com/consensys/zkevm-monorepo/prover/zkevm/prover/hash/generic"
	"github.com/consensys/zkevm-monorepo/prover/zkevm/prover/hash/keccak"
	"github.com/stretchr/testify/assert"
)

func TestAddress(t *testing.T) {

	m := keccak.Module{}
	limits := &Limits{
		MaxNbEcRecover: 1,
		MaxNbTx:        4,
	}
	ac := &Antichamber{Limits: limits}
	var addr *Addresses
	var uaGnark *UnalignedGnarkData
	var ecRec *EcRecover
	var td *txnData
	gbmGnark := generic.GenericByteModule{}

	size := limits.sizeAntichamber()

	nbKeccakF := ac.nbKeccakF(8) // if each txn has 8 blocks
	nbRowsPerTxInTxnData := 9

	sizeTxnData := utils.NextPowerOfTwo(limits.MaxNbTx * nbRowsPerTxInTxnData)

	compiled := wizard.Compile(func(b *wizard.Builder) {
		comp := b.CompiledIOP

		// generate a gbm and use it to represent gnark-columns
		gbmGnark = acc_module.CommitGBM(comp, 0, generic.SHAKIRA, size)
		ac = &Antichamber{
			Limits: limits,
			ID:     gbmGnark.Data.HashNum,
		}
		uaGnark = &UnalignedGnarkData{
			GnarkData:           gbmGnark.Data.Limb,
			GnarkPublicKeyIndex: gbmGnark.Data.Index,
			IsPublicKey:         gbmGnark.Data.TO_HASH,
		}

		// commit to txnData and ecRecover
		td, ecRec = commitEcRecTxnData(comp, sizeTxnData, size, ac)

		// native columns and  constraints
		addr = newAddress(comp, size, ecRec, ac, td)

		// prepare the provider for keccak
		provider := addr.GetProvider(comp, ac, uaGnark)

		// define keccak (columns and constraints)
		m.Define(comp, []generic.GenericByteModule{provider}, nbKeccakF)
	}, dummy.Compile)

	proof := wizard.Prove(compiled, func(run *wizard.ProverRuntime) {

		// assign GnarkColumns via gbmGnark
		acc_module.AssignGBMfromTable(run, &gbmGnark, size, limits.MaxNbEcRecover+limits.MaxNbTx, false)
		// it assign mock data to EcRec and txn_data
		AssignEcRecTxnData(run, gbmGnark, limits.MaxNbEcRecover, limits.MaxNbTx, sizeTxnData, size, td, ecRec, ac)

		// assign address columns
		addr.assignAddress(run, limits.MaxNbEcRecover, size, ac, ecRec, uaGnark, td)

		// assign keccak columns via provider that is embedded in the receiver
		m.AssignKeccak(run)
	})
	assert.NoError(t, wizard.Verify(compiled, proof))
}
