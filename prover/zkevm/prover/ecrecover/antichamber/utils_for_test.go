package antichamber

import (
	"github.com/consensys/zkevm-monorepo/prover/crypto/keccak"
	"github.com/consensys/zkevm-monorepo/prover/maths/common/smartvectors"
	"github.com/consensys/zkevm-monorepo/prover/maths/common/vector"
	"github.com/consensys/zkevm-monorepo/prover/maths/field"
	"github.com/consensys/zkevm-monorepo/prover/protocol/ifaces"
	"github.com/consensys/zkevm-monorepo/prover/protocol/wizard"
	"github.com/consensys/zkevm-monorepo/prover/utils"
	"github.com/consensys/zkevm-monorepo/prover/zkevm/prover/hash/generic"

	"golang.org/x/crypto/sha3"
)

func commitEcRecTxnData(comp *wizard.CompiledIOP, size1 int, size int, ac *Antichamber) (td *txnData, ecRec *EcRecover) {
	td = &txnData{
		fromHi: comp.InsertCommit(0, ifaces.ColIDf("txn_data.FromHi"), size1),
		fromLo: comp.InsertCommit(0, ifaces.ColIDf("txn_data.FromLo"), size1),
		ct:     comp.InsertCommit(0, ifaces.ColIDf("txn_data.CT"), size1),
	}

	ecRec = &EcRecover{
		Limb:           comp.InsertCommit(0, ifaces.ColIDf("ECRECOVER_LIMB"), size),
		EcRecoverIsRes: comp.InsertCommit(0, ifaces.ColIDf("ECRECOVER_ISRES"), size),
	}
	ac.IsActive = comp.InsertCommit(0, ifaces.ColID("AntiChamber_IsActive"), size)
	return td, ecRec
}

func AssignEcRecTxnData(
	run *wizard.ProverRuntime,
	gbm generic.GenDataModule,
	nbEcRec, nbTxS int,
	sizeTxnData, size int,
	td *txnData, ecRec *EcRecover,
	ac *Antichamber,
) {

	permTrace := keccak.GenerateTrace(gbm.ScanStreams(run))

	// now assign ecRecover.Limb and txn_data.From from the permutation trace.
	isEcRecRes := make([]field.Element, nbEcRec*nbRowsPerEcRec)
	ecRecLimb := make([]field.Element, nbEcRec*nbRowsPerEcRec)

	nbRowsPerTxInTxnData := 9
	var ctWit []field.Element
	for i := 0; i < nbTxS; i++ {
		for j := 0; j < nbRowsPerTxInTxnData; j++ {
			ctWit = append(ctWit, field.NewElement(uint64(j+1)))
		}
	}

	fromHi := make([]field.Element, nbTxS*nbRowsPerTxInTxnData)
	fromLo := make([]field.Element, nbTxS*nbRowsPerTxInTxnData)
	offSetEcRec := 0

	if nbEcRec+nbTxS != len(permTrace.HashOutPut) {
		utils.Panic("the number of generated hash %v should be %v + %v", len(permTrace.HashOutPut), nbEcRec, nbTxS)
	}
	for i, hashRes := range permTrace.HashOutPut {

		if i < nbEcRec {
			isEcRecRes[i*nbRowsPerEcRec+offSetEcRec] = field.One()
			isEcRecRes[i*nbRowsPerEcRec+offSetEcRec+1] = field.One()

			ecRecLimb[i*nbRowsPerEcRec+offSetEcRec].SetBytes(hashRes[halfDigest-trimmingSize : halfDigest])
			ecRecLimb[i*nbRowsPerEcRec+offSetEcRec+1].SetBytes(hashRes[halfDigest:])
		} else {
			j := i - nbEcRec

			fromHi[j*nbRowsPerTxInTxnData].SetBytes(hashRes[halfDigest-trimmingSize : halfDigest])
			fromLo[j*nbRowsPerTxInTxnData].SetBytes(hashRes[halfDigest:])
		}
	}

	run.AssignColumn(ecRec.EcRecoverIsRes.GetColID(), smartvectors.RightZeroPadded(isEcRecRes, size))
	run.AssignColumn(ecRec.Limb.GetColID(), smartvectors.RightZeroPadded(ecRecLimb, size))

	// they are arithmetization columns, so LeftZeroPad
	run.AssignColumn(td.fromHi.GetColID(), smartvectors.LeftZeroPadded(fromHi, sizeTxnData))
	run.AssignColumn(td.fromLo.GetColID(), smartvectors.LeftZeroPadded(fromLo, sizeTxnData))
	run.AssignColumn(td.ct.GetColID(), smartvectors.LeftZeroPadded(ctWit, sizeTxnData))

	effectiveSize := nbEcRec*nbRowsPerEcRec + nbTxS*nbRowsPerTxSign
	isActive := vector.Repeat(field.One(), effectiveSize)
	run.AssignColumn(ac.IsActive.GetColID(), smartvectors.RightZeroPadded(isActive, size))
}

// it estimates the required number of number of keccakF.
func (l *Limits) nbKeccakF(nbKeccakFPerTxn int) int {
	return l.MaxNbTx*nbKeccakFPerTxn + l.MaxNbEcRecover
}

func (l *Limits) sizeTxnData(nbRowsPerTxInTxnData int) int {
	return utils.NextPowerOfTwo(l.MaxNbTx * nbRowsPerTxInTxnData)
}

// It receives a set of public keys, and assigns the txn_data
func (td *txnData) assignTxnDataFromPK(
	run *wizard.ProverRuntime,
	ac *Antichamber,
	rlpTxnHashes [][32]byte,
	nbRowsPerTxInTxnData int,
) {

	// compute the hash of public keys
	var pkHash [][]byte
	hasher := sha3.NewLegacyKeccak256()
	for i := range rlpTxnHashes {
		pk, _, _, _, err := generateDeterministicSignature(rlpTxnHashes[i][:])
		if err != nil {
			utils.Panic("error generating signature")
		}
		buf := pk.A.RawBytes()
		hasher.Write(buf[:])
		res := hasher.Sum(nil)
		hasher.Reset()
		pkHash = append(pkHash, res)
	}

	// now assign  txn_data from the hash results.
	// populate the CT column
	var ctWit []field.Element
	for i := 0; i < ac.MaxNbTx; i++ {
		for j := 0; j < nbRowsPerTxInTxnData; j++ {
			ctWit = append(ctWit, field.NewElement(uint64(j+1)))
		}
	}

	// populate the columns FromHi and FromLo
	fromHi := make([]field.Element, ac.MaxNbTx*nbRowsPerTxInTxnData)
	fromLo := make([]field.Element, ac.MaxNbTx*nbRowsPerTxInTxnData)

	for i := 0; i < len(pkHash); i++ {
		fromHi[i*nbRowsPerTxInTxnData].SetBytes(pkHash[i][halfDigest-trimmingSize : halfDigest])
		fromLo[i*nbRowsPerTxInTxnData].SetBytes(pkHash[i][halfDigest:])

	}

	// these are arithmetization columns, so LeftZeroPad
	run.AssignColumn(td.fromHi.GetColID(), smartvectors.LeftZeroPadded(fromHi, ac.sizeTxnData(nbRowsPerTxInTxnData)))
	run.AssignColumn(td.fromLo.GetColID(), smartvectors.LeftZeroPadded(fromLo, ac.sizeTxnData(nbRowsPerTxInTxnData)))
	run.AssignColumn(td.ct.GetColID(), smartvectors.LeftZeroPadded(ctWit, ac.sizeTxnData(nbRowsPerTxInTxnData)))
}

// it commits to the txn_data
func commitTxnData(comp *wizard.CompiledIOP, limits *Limits, nbRowsPerTxInTxnData int) (td *txnData) {
	size := limits.sizeTxnData(nbRowsPerTxInTxnData)
	td = &txnData{
		fromHi: comp.InsertCommit(0, ifaces.ColIDf("txn_data.FromHi"), size),
		fromLo: comp.InsertCommit(0, ifaces.ColIDf("txn_data.FromLo"), size),
		ct:     comp.InsertCommit(0, ifaces.ColIDf("txn_data.CT"), size),
	}
	return td
}
