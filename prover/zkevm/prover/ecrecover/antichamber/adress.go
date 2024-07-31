package antichamber

import (
	"github.com/consensys/zkevm-monorepo/prover/crypto/keccak"
	"github.com/consensys/zkevm-monorepo/prover/maths/common/smartvectors"
	"github.com/consensys/zkevm-monorepo/prover/maths/common/vector"
	"github.com/consensys/zkevm-monorepo/prover/maths/field"
	"github.com/consensys/zkevm-monorepo/prover/protocol/column"
	"github.com/consensys/zkevm-monorepo/prover/protocol/dedicated/byte32cmp"
	"github.com/consensys/zkevm-monorepo/prover/protocol/dedicated/projection"
	"github.com/consensys/zkevm-monorepo/prover/protocol/ifaces"
	"github.com/consensys/zkevm-monorepo/prover/protocol/wizard"
	sym "github.com/consensys/zkevm-monorepo/prover/symbolic"
	"github.com/consensys/zkevm-monorepo/prover/zkevm/prover/hash/datatransfer/dedicated"
	"github.com/consensys/zkevm-monorepo/prover/zkevm/prover/hash/generic"
)

// Address submodule is responsible for the columns holding the address of the sender,
// and checking their consistency with the claimed public key
// (since address is the truncated hash of public key).
//
// The addresses comes from two arithmetization modules txn-data and ec-data.
//
// The public-key comes from Gnark-Data.
type Addresses struct {

	// main columns
	addressHiUntrimmed ifaces.Column
	addressHi          ifaces.Column
	addressLo          ifaces.Column

	// filters over address columns
	isAddress            ifaces.Column
	isAddressFromEcRec   ifaces.Column
	isAddressFromTxnData ifaces.Column

	// helper columns for intermediate computations/proofs

	// filter over ecRecover; indicating only the AddressHi from EcRecoverIsRes
	// we need this columns just because projection query does not support expressions as filter
	isAddressHiEcRec ifaces.Column
	// a column of all 16 indicating that all 16 bytes of public key should be hashed.
	col16 ifaces.Column

	// columns for decomposition and trimming the HashHi to AddressHi
	limbColumnsUntrimmed        byte32cmp.LimbColumns
	computeLimbColumnsUntrimmed wizard.ProverAction
}

// AddressHi is the trimming of HashHi, taking its last 4bytes.
const trimmingSize = 4

// newAddress creates an Address struct, declaring native columns and the constraints among them.
func newAddress(comp *wizard.CompiledIOP, size int, ecRec *EcRecover, ac *Antichamber, td *txnData) *Addresses {
	createCol := createColFn(comp, NAME_ADDRESSES, size)
	ecRecSize := ecRec.EcRecoverIsRes.Size()
	// declare the native columns
	addr := &Addresses{
		addressHi:          createCol("ADDRESS_HI"),
		addressLo:          createCol("ADDRESS_LO"),
		isAddress:          createCol("IS_ADDRESS"),
		addressHiUntrimmed: createCol("ADRESSHI_UNTRIMMED"),
		col16: comp.InsertPrecomputed(ifaces.ColIDf("ADDRESS_Col16"),
			smartvectors.NewRegular(vector.Repeat(field.NewElement(16), size))),
		isAddressHiEcRec:     comp.InsertCommit(0, ifaces.ColIDf("ISADRESS_HI_ECREC"), ecRecSize),
		isAddressFromEcRec:   createCol("ISADRESS_FROM_ECREC"),
		isAddressFromTxnData: createCol("ISADRESS_FROM_TXNDATA"),
	}

	// addresses are fetched from two arithmetization modules (ecRecover and txn-data)
	// IsAddress = IsAdressFromEcRec + IsAdressFromTxnData
	comp.InsertGlobal(0, ifaces.QueryIDf("Format_IsAddress"),
		sym.Sub(addr.isAddress, sym.Add(addr.isAddressFromEcRec, addr.isAddressFromTxnData)))

	mustBeBinary(comp, addr.isAddress)
	mustBeBinary(comp, addr.isAddressFromEcRec)
	mustBeBinary(comp, addr.isAddressFromTxnData)
	isZeroWhenInactive(comp, addr.isAddress, ac.IsActive)

	// check the  trimming of hashHi  to the addressHi
	addr.csAddressTrimming(comp)

	// check that IsAddressHiEcRec is well-formed
	addr.csIsAddressHiEcRec(comp, ecRec)

	// projection from ecRecover to address columns
	// ecdata is already projected over our ecRecover. Thus, we only project from our ecrecover.
	projection.InsertProjection(comp, ifaces.QueryIDf("Project_AddressHi_EcRec"),
		[]ifaces.Column{ecRec.Limb}, []ifaces.Column{addr.addressHi},
		addr.isAddressHiEcRec, addr.isAddressFromEcRec,
	)

	projection.InsertProjection(comp, ifaces.QueryIDf("Project_AddressLo_EcRec"),
		[]ifaces.Column{ecRec.Limb}, []ifaces.Column{addr.addressLo},
		column.Shift(addr.isAddressHiEcRec, -1), addr.isAddressFromEcRec,
	)
	td.csTxnData(comp)
	// projection from txn-data to address columns
	projection.InsertProjection(comp, ifaces.QueryIDf("Project_AddressHi_TxnData"),
		[]ifaces.Column{td.fromHi}, []ifaces.Column{addr.addressHi},
		td.isFrom, addr.isAddressFromTxnData,
	)

	projection.InsertProjection(comp, ifaces.QueryIDf("Project_AddressLO_TxnData"),
		[]ifaces.Column{td.fromLo}, []ifaces.Column{addr.addressLo},
		td.isFrom, addr.isAddressFromTxnData,
	)
	return addr
}

// It checks the well-forming of IsAddressHiEcRec
func (addr *Addresses) csIsAddressHiEcRec(comp *wizard.CompiledIOP, ecRec *EcRecover) {
	// if EcRecoverIsRes[i] == 1 and EcRecover[i+1] == 1 ---> isAddressHiEcRec[i] = 1
	comp.InsertGlobal(0, ifaces.QueryIDf("Is_AddressHi_EcRec_1"),
		sym.Mul(ecRec.EcRecoverIsRes, column.Shift(ecRec.EcRecoverIsRes, 1),
			sym.Sub(1, addr.isAddressHiEcRec)))

	// if EcRecoverIsRes[i] == 0  ---> isAddressHiEcRec[i] = 0
	comp.InsertGlobal(0, ifaces.QueryIDf("Is_AddressHi_EcRec_2"),
		sym.Mul(sym.Sub(1, ecRec.EcRecoverIsRes), addr.isAddressHiEcRec))

	// if EcRecoverIsRes[i] == 1 and EcRecover[i+1] == 0 ---> isAddressHiEcRec[i] = 0
	comp.InsertGlobal(0, ifaces.QueryIDf("Is_AddressHi_EcRec_3"),
		sym.Mul(ecRec.EcRecoverIsRes, sym.Sub(1, column.Shift(ecRec.EcRecoverIsRes, 1)),
			addr.isAddressHiEcRec))
}

// The constraints for trimming the HashHi to AddressHi
func (addr *Addresses) csAddressTrimming(comp *wizard.CompiledIOP) {

	bitPerLimbs := 16
	addr.limbColumnsUntrimmed, addr.computeLimbColumnsUntrimmed = byte32cmp.Decompose(comp, addr.addressHiUntrimmed, 8, bitPerLimbs)
	// addr.LimbColumnsTrimmed, addr.computeLimbColumnsTrimmed = byte32cmp.Decompose(comp, addr.AddressHi, 2, 16)

	// recompose two first limbs to get AddressHi
	// since decomposition is in little-endian, but address is in big-endian, we get the first limbs.
	a := addr.limbColumnsUntrimmed.Limbs[0]
	b := addr.limbColumnsUntrimmed.Limbs[1]
	pow2 := sym.NewConstant(1 << bitPerLimbs)
	expr := sym.Add(a, sym.Mul(pow2, b))

	comp.InsertGlobal(0, ifaces.QueryIDf("Address_Trimming"), sym.Sub(expr, addr.addressHi))
}

// It builds a provider from  public key extracted from Gnark-Data (as hash input) and addresses (as output).
// the consistency check is then deferred to the keccak module.
func (addr *Addresses) GetProvider(comp *wizard.CompiledIOP, ac *Antichamber, uaGnark *UnalignedGnarkData) generic.GenericByteModule {
	// generate a generic byte Module as keccak provider.
	provider := addr.buildGenericModule(ac, uaGnark)
	return provider
}

// It builds a GenericByteModule from Address columns and Public-Key/GnarkData columns.
func (addr *Addresses) buildGenericModule(ac *Antichamber, uaGnark *UnalignedGnarkData) (pkModule generic.GenericByteModule) {
	pkModule.Data = generic.GenDataModule{
		HashNum: ac.ID,
		Limb:    uaGnark.GnarkData,

		// a column of all 16, since all the bytes of public key are used in hashing
		NBytes:  addr.col16,
		Index:   uaGnark.GnarkPublicKeyIndex,
		TO_HASH: uaGnark.IsPublicKey,
	}

	pkModule.Info = generic.GenInfoModule{
		HashHi:   addr.addressHiUntrimmed,
		HashLo:   addr.addressLo,
		IsHashHi: addr.isAddress,
		IsHashLo: addr.isAddress,
	}
	return pkModule
}

// It assigns the native columns specific to the submodule.
func (addr *Addresses) assignAddress(
	run *wizard.ProverRuntime,
	nbEcRecover, size int,
	ac *Antichamber,
	ecRec *EcRecover,
	uaGnark *UnalignedGnarkData,
	td *txnData,
) {
	td.assignTxnData(run)
	addr.assignMainColumns(run, nbEcRecover, size, ac, uaGnark)
	addr.assignHelperColumns(run, ecRec)
}

// It assigns the main columns
func (addr *Addresses) assignMainColumns(
	run *wizard.ProverRuntime,
	nbEcRecover, size int,
	ac *Antichamber,
	uaGnark *UnalignedGnarkData,
) {
	pkModule := addr.buildGenericModule(ac, uaGnark)
	// since we use it just for trace generating, we have to turn-off the info-module (hash output).
	pkModule.Info = generic.GenInfoModule{}

	split := splitAt(nbEcRecover)
	n := nbRowsPerEcRec

	var (
		hashHi, hashLo, isHash, trimmedHi []field.Element
	)

	permTrace := keccak.PermTraces{}
	genTrace := generic.GenTrace{}
	pkModule.AppendTraces(run, &genTrace, &permTrace)

	var v, w, u field.Element
	for _, digest := range permTrace.HashOutPut {

		hi := digest[:halfDigest]
		lo := digest[halfDigest:]
		trimmed := hi[halfDigest-trimmingSize:]

		v.SetBytes(hi[:])
		w.SetBytes(lo[:])
		u.SetBytes(trimmed[:])

		if len(hashHi) == split {
			n = nbRowsPerTxSign
		}
		repeatLO := vector.Repeat(w, n)
		repeatHi := vector.Repeat(v, n)
		repeatTrimmedHi := vector.Repeat(u, n)
		repeatIsTxHash := vector.Repeat(field.Zero(), n-1)

		hashHi = append(hashHi, repeatHi...)
		hashLo = append(hashLo, repeatLO...)
		isHash = append(isHash, field.One())
		isHash = append(isHash, repeatIsTxHash...)
		trimmedHi = append(trimmedHi, repeatTrimmedHi...)
	}

	isFromEcRec := isHash[:split]
	isFromTxnData := vector.Repeat(field.Zero(), split)
	isFromTxnData = append(isFromTxnData, isHash[split:]...)

	run.AssignColumn(addr.addressHiUntrimmed.GetColID(), smartvectors.RightZeroPadded(hashHi, size))
	run.AssignColumn(addr.addressLo.GetColID(), smartvectors.RightZeroPadded(hashLo, size))
	run.AssignColumn(addr.isAddress.GetColID(), smartvectors.RightZeroPadded(isHash, size))
	run.AssignColumn(addr.addressHi.GetColID(), smartvectors.RightZeroPadded(trimmedHi, size))
	run.AssignColumn(addr.isAddressFromEcRec.GetColID(), smartvectors.RightZeroPadded(isFromEcRec, size))
	run.AssignColumn(addr.isAddressFromTxnData.GetColID(), smartvectors.RightZeroPadded(isFromTxnData, size))

}

// It assigns the helper columns
func (addr *Addresses) assignHelperColumns(run *wizard.ProverRuntime, ecRec *EcRecover) {

	// assign LimbColumns from decomposition via proverAction
	addr.computeLimbColumnsUntrimmed.Run(run)

	// assign isAddressHiEcRec
	isRes := ecRec.EcRecoverIsRes.GetColAssignment(run).IntoRegVecSaveAlloc()
	col := make([]field.Element, len(isRes))
	for i := 0; i < len(isRes); i++ {
		if i < len(isRes)-1 && isRes[i].IsOne() && isRes[i+1].IsOne() {
			col[i] = field.One()
			col[i+1] = field.Zero()
			i = i + 1
		}
	}
	run.AssignColumn(addr.isAddressHiEcRec.GetColID(), smartvectors.NewRegular(col))

}

// It indicates the row where ecrecover and txSignature are split.
func splitAt(nbEcRecover int) int {
	return nbEcRecover * nbRowsPerEcRec
}

func (td *txnData) csTxnData(comp *wizard.CompiledIOP) {
	td.isFrom = comp.InsertCommit(0, ifaces.ColIDf("%v_IsFrom", NAME_ADDRESSES), td.ct.Size())
	// check that isFrom == 1 iff ct==1
	dedicated.InsertIsTargetValue(comp, 0, ifaces.QueryIDf("IsFrom_IsCorrect"), field.One(), td.ct, td.isFrom)
}

func (td *txnData) assignTxnData(run *wizard.ProverRuntime) {
	// assign isFrom via CT
	var isFrom []field.Element
	ct := td.ct.GetColAssignment(run).IntoRegVecSaveAlloc()
	for j := range ct {
		if ct[j].IsOne() {
			isFrom = append(isFrom, field.One())
		} else {
			isFrom = append(isFrom, field.Zero())
		}
	}
	run.AssignColumn(td.isFrom.GetColID(), smartvectors.NewRegular(isFrom))
}

// txndata represents the txn_data module from the arithmetization side.
type txnData struct {
	fromHi ifaces.Column
	fromLo ifaces.Column
	ct     ifaces.Column

	// helper column
	isFrom ifaces.Column
}
