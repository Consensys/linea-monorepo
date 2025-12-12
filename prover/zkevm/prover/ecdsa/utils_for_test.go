package ecdsa

import (
	"fmt"
	"slices"

	"github.com/consensys/linea-monorepo/prover/crypto/keccak"
	"github.com/consensys/linea-monorepo/prover/maths/common/smartvectors"
	"github.com/consensys/linea-monorepo/prover/maths/common/vector"
	"github.com/consensys/linea-monorepo/prover/maths/field"
	"github.com/consensys/linea-monorepo/prover/protocol/ifaces"
	"github.com/consensys/linea-monorepo/prover/protocol/wizard"
	"github.com/consensys/linea-monorepo/prover/utils"
	"github.com/consensys/linea-monorepo/prover/zkevm/prover/common"
	"github.com/consensys/linea-monorepo/prover/zkevm/prover/hash/generic"
	"golang.org/x/crypto/sha3"
)

func commitEcRecTxnData(comp *wizard.CompiledIOP, size1 int, size int, ac *antichamber) (td *txnData, ecRec *EcRecover) {
	td = &txnData{
		Ct:       comp.InsertCommit(0, ifaces.ColIDf("txn_data.CT"), size1, true),
		User:     comp.InsertCommit(0, ifaces.ColIDf("txn_data.USER"), size1, true),
		Selector: comp.InsertCommit(0, ifaces.ColIDf("txn_data.SELECTOR"), size1, true),
	}

	for i := 0; i < common.NbLimbU256; i++ {
		td.From[i] = comp.InsertCommit(0, ifaces.ColID(fmt.Sprintf("txndata.From_%d", i)), size1, true)
	}

	ecRec = &EcRecover{
		EcRecoverIsRes: comp.InsertCommit(0, ifaces.ColIDf("ECRECOVER_ISRES"), size, true),
	}

	for i := 0; i < common.NbLimbU128; i++ {
		ecRec.Limb[i] = comp.InsertCommit(0, ifaces.ColIDf("ECRECOVER_LIMB_%d", i), size, true)
	}

	// Patch: the assignment of the ecRecLimbs assumes a big endian form but the
	// correct form expected by the module is little endian.
	slices.Reverse(ecRec.Limb[:])

	ac.IsActive = comp.InsertCommit(0, "AntiChamber_IsActive", size, true)
	return td, ecRec
}

func AssignEcRecTxnData(
	run *wizard.ProverRuntime,
	gbm generic.GenDataModule,
	nbEcRec, nbTxS int,
	sizeTxnData, size int,
	td *txnData, ecRec *EcRecover,
	ac *antichamber,
) {
	streams := gbm.ScanStreams(run)
	permTrace := keccak.GenerateTrace(streams)

	// now assign ecRecover.Limb and txn_data.From from the permutation trace.
	isEcRecRes := make([]field.Element, nbEcRec*nbRowsPerEcRec)
	ecRecLimb := make([][]field.Element, common.NbLimbU128)
	for i := range ecRecLimb {
		ecRecLimb[i] = make([]field.Element, nbEcRec*nbRowsPerEcRec)
	}

	nbRowsPerTxInTxnData := 9
	var ctWit []field.Element
	var userWit []field.Element
	var selectorWit []field.Element
	for i := 0; i < nbTxS; i++ {
		for j := 0; j < nbRowsPerTxInTxnData; j++ {
			if j == 0 {
				// First row: ct=0, User=1, Selector=1 (for isFrom condition)
				ctWit = append(ctWit, field.Zero())
				userWit = append(userWit, field.One())
				selectorWit = append(selectorWit, field.One())
			} else {
				// Other rows: ct=j, User=1, Selector=1
				ctWit = append(ctWit, field.NewElement(uint64(j)))
				userWit = append(userWit, field.One())
				selectorWit = append(selectorWit, field.One())
			}
		}
	}

	// Initialize an array of from limbs columns
	from := make([][]field.Element, 0, common.NbLimbU256)
	for i := 0; i < common.NbLimbU256; i++ {
		from = append(from, make([]field.Element, nbTxS*nbRowsPerTxInTxnData))
	}

	offSetEcRec := 0

	if nbEcRec+nbTxS != len(permTrace.HashOutPut) {
		utils.Panic("the number of generated hash %v should be %v + %v", len(permTrace.HashOutPut), nbEcRec, nbTxS)
	}

	for i, hashRes := range permTrace.HashOutPut {
		if i < nbEcRec {
			isEcRecRes[i*nbRowsPerEcRec+offSetEcRec] = field.One()
			isEcRecRes[i*nbRowsPerEcRec+offSetEcRec+1] = field.One()

			ecRecHiLimbs := common.SplitBytes(hashRes[addressTrimmedBytes:halfDigest])
			ecRecLoLimbs := common.SplitBytes(hashRes[halfDigest:])

			for j := 0; j < common.NbLimbU128; j++ {
				if j >= addressTrimmedColumns {
					ecRecLimb[j][i*nbRowsPerEcRec+offSetEcRec].SetBytes(ecRecHiLimbs[j-addressTrimmedColumns])
				}

				ecRecLimb[j][i*nbRowsPerEcRec+offSetEcRec+1].SetBytes(ecRecLoLimbs[j])
			}

			continue
		} else {
			fromLimbs := common.SplitBytes(hashRes[:])
			j := i - nbEcRec

			for k, limb := range fromLimbs {
				// Initialize limb values for each column of from
				from[k][j*nbRowsPerTxInTxnData].SetBytes(limb[:])
			}
		}
	}

	run.AssignColumn(ecRec.EcRecoverIsRes.GetColID(), smartvectors.RightZeroPadded(isEcRecRes, size))

	// Patch: the assignment of the ecRecLimbs assumes a big endian form but the
	// correct form expected by the module is little endian.
	slices.Reverse(ecRecLimb)
	for i := 0; i < common.NbLimbU128; i++ {
		run.AssignColumn(ecRec.Limb[i].GetColID(), smartvectors.RightZeroPadded(ecRecLimb[i], size))
	}

	// they are arithmetization columns, so LeftZeroPad
	for i := 0; i < common.NbLimbU256; i++ {
		run.AssignColumn(td.From[i].GetColID(), smartvectors.LeftZeroPadded(from[i], sizeTxnData))
	}

	run.AssignColumn(td.Ct.GetColID(), smartvectors.LeftZeroPadded(ctWit, sizeTxnData))
	run.AssignColumn(td.User.GetColID(), smartvectors.LeftZeroPadded(userWit, sizeTxnData))
	run.AssignColumn(td.Selector.GetColID(), smartvectors.LeftZeroPadded(selectorWit, sizeTxnData))

	effectiveSize := nbEcRec*nbRowsPerEcRec + nbTxS*nbRowsPerTxSign
	isActive := vector.Repeat(field.One(), effectiveSize)
	run.AssignColumn(ac.IsActive.GetColID(), smartvectors.RightZeroPadded(isActive, size))
}

// it estimates the required number of number of keccakF.
func (l *Settings) nbKeccakF(nbKeccakFPerTxn int) int {
	return l.MaxNbTx*nbKeccakFPerTxn + l.MaxNbEcRecover
}

func (l *Settings) sizeTxnData(nbRowsPerTxInTxnData int) int {
	return utils.NextPowerOfTwo(l.MaxNbTx * nbRowsPerTxInTxnData)
}

// It receives a set of public keys, and assigns the txn_data
func (td *txnData) assignTxnDataFromPK(
	run *wizard.ProverRuntime,
	ac *antichamber,
	rlpTxnHashes [][32]byte,
	nbRowsPerTxInTxnData int,
) {
	var (
		hasher  = sha3.NewLegacyKeccak256()
		maxNbTx = ac.Inputs.Settings.MaxNbTx
	)
	// compute the hash of public keys
	pkHash := make([][]byte, 0, len(rlpTxnHashes))
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
	// populate the CT, User, and Selector columns
	var ctWit []field.Element
	var userWit []field.Element
	var selectorWit []field.Element
	for i := 0; i < maxNbTx; i++ {
		for j := 0; j < nbRowsPerTxInTxnData; j++ {
			if j == 0 {
				// First row: ct=0, User=1, Selector=1 (for isFrom condition)
				ctWit = append(ctWit, field.Zero())
				userWit = append(userWit, field.One())
				selectorWit = append(selectorWit, field.One())
			} else {
				// Other rows: ct=j, User=1, Selector=1
				ctWit = append(ctWit, field.NewElement(uint64(j)))
				userWit = append(userWit, field.One())
				selectorWit = append(selectorWit, field.One())
			}
		}
	}

	// populate the columns FromHi and FromLo
	from := make([][]field.Element, 0, common.NbLimbU256)
	for i := 0; i < common.NbLimbU256; i++ {
		from = append(from, make([]field.Element, maxNbTx*nbRowsPerTxInTxnData))
	}

	for i := 0; i < len(pkHash); i++ {
		fromLimbs := common.SplitBytes(pkHash[i])

		for j, limb := range fromLimbs {
			// Initialize limb values for each column of from
			from[j][i*nbRowsPerTxInTxnData].SetBytes(limb[:])
		}
	}

	// these are arithmetization columns, so LeftZeroPad
	for i := 0; i < common.NbLimbU256; i++ {
		run.AssignColumn(td.From[i].GetColID(), smartvectors.LeftZeroPadded(from[i], ac.Inputs.Settings.sizeTxnData(nbRowsPerTxInTxnData)))
	}

	run.AssignColumn(td.Ct.GetColID(), smartvectors.LeftZeroPadded(ctWit, ac.Inputs.Settings.sizeTxnData(nbRowsPerTxInTxnData)))
	run.AssignColumn(td.User.GetColID(), smartvectors.LeftZeroPadded(userWit, ac.Inputs.Settings.sizeTxnData(nbRowsPerTxInTxnData)))
	run.AssignColumn(td.Selector.GetColID(), smartvectors.LeftZeroPadded(selectorWit, ac.Inputs.Settings.sizeTxnData(nbRowsPerTxInTxnData)))
}

// it commits to the txn_data
func commitTxnData(comp *wizard.CompiledIOP, limits *Settings, nbRowsPerTxInTxnData int) (td *txnData) {
	size := limits.sizeTxnData(nbRowsPerTxInTxnData)

	td = &txnData{
		Ct:       comp.InsertCommit(0, ifaces.ColIDf("txn_data.CT"), size, true),
		User:     comp.InsertCommit(0, ifaces.ColIDf("txn_data.USER"), size, true),
		Selector: comp.InsertCommit(0, ifaces.ColIDf("txn_data.SELECTOR"), size, true),
	}

	for i := 0; i < common.NbLimbU256; i++ {
		td.From[i] = comp.InsertCommit(0, ifaces.ColIDf("txn_data.From_%d", i), size, true)
	}

	return td
}
