package ecdsa

import (
	"fmt"

	"github.com/consensys/linea-monorepo/prover/crypto/keccak"
	"github.com/consensys/linea-monorepo/prover/maths/common/smartvectors"
	"github.com/consensys/linea-monorepo/prover/maths/common/vector"
	"github.com/consensys/linea-monorepo/prover/maths/field"
	"github.com/consensys/linea-monorepo/prover/protocol/column"
	"github.com/consensys/linea-monorepo/prover/protocol/column/verifiercol"
	"github.com/consensys/linea-monorepo/prover/protocol/dedicated"
	"github.com/consensys/linea-monorepo/prover/protocol/ifaces"
	"github.com/consensys/linea-monorepo/prover/protocol/query"
	"github.com/consensys/linea-monorepo/prover/protocol/wizard"
	sym "github.com/consensys/linea-monorepo/prover/symbolic"
	"github.com/consensys/linea-monorepo/prover/zkevm/prover/common"
	commoncs "github.com/consensys/linea-monorepo/prover/zkevm/prover/common/common_constraints"
	"github.com/consensys/linea-monorepo/prover/zkevm/prover/hash/generic"
)

const (
	// addressHiBytes is the size of the leftover from trimmed addressHi part (in bytes).
	addressHiBytes = 4
	// addressHiColumns is the number of addressHi columns.
	addressHiColumns = addressHiBytes / common.LimbBytes
	// addressTrimmedBytes size of the trimmed address part (in bytes).
	addressTrimmedBytes = common.NbLimbU256 - addressHiBytes
	// addressTrimmedColumns number of columns that represent the trimmed address part.
	addressTrimmedColumns = addressTrimmedBytes / common.LimbBytes
)

// Addresses submodule is responsible for the columns holding the address of the sender,
// and checking their consistency with the claimed public key
// (since address is the truncated hash of public key).
//
// The addresses comes from two arithmetization modules txn-data and ec-data.
//
// The public-key comes from Gnark-Data.
type Addresses struct {

	// main columns
	AddressUntrimmed [common.NbLimbU256]ifaces.Column
	Address          [common.NbLimbEthAddress]ifaces.Column

	// filters over address columns
	IsAddress            ifaces.Column
	IsAddressFromEcRec   ifaces.Column
	IsAddressFromTxnData ifaces.Column

	// helper columns for intermediate computations/proofs

	// filter over ecRecover; indicating only the AddressHi limbs from EcRecoverIsRes
	// we need this columns just because projection query does not support expressions
	// as filter. The shifted version of this column by 1 indicates the addressLo from
	// ecrec.
	IsAddressHiEcRec ifaces.Column

	// IsAddressLoEcRec is a column that is explicitly constructed as the shifted version
	// of IsAddressLoEcRec. The reason we need to build this column is because projection
	// queries cannot be distributed if their selector is a shifted column. (There would
	// be a need to deal with the segment boundaries and it has not been figured out.)
	// So, to simplify, we create a column directly and leave the business of managing
	// the boundary values to a global constraint to which the feature is avaiable.
	IsAddressLoEcRec *dedicated.ManuallyShifted

	// a column of all 16 indicating that all 16 bytes of public key should be hashed.
	Col16 ifaces.Column

	// used as the hashID for hashing by keccak.
	HashNum ifaces.Column

	// providers for keccak, Providers contain the inputs and outputs of keccak hash.
	Provider generic.GenericByteModule
}

// newAddress creates an Address struct, declaring native columns and the constraints among them.
func newAddress(comp *wizard.CompiledIOP, size int, ecRec *EcRecover, ac *antichamber, td *txnData) *Addresses {
	createCol := createColFn(comp, NAME_ADDRESSES, size)
	ecRecSize := ecRec.EcRecoverIsRes.Size()

	// Create columns for every 16 bit limb of address
	address := [common.NbLimbEthAddress]ifaces.Column{}
	for i := 0; i < common.NbLimbEthAddress; i++ {
		address[i] = createCol(fmt.Sprintf("ADDRESS_%d", i))
	}

	// Create columns for every 16 bit limb of addressUntrimmed
	addressUntrimmed := [common.NbLimbU256]ifaces.Column{}
	for i := 0; i < common.NbLimbU256; i++ {
		addressUntrimmed[i] = createCol(fmt.Sprintf("ADRESS_UNTRIMMED_%d", i))
	}

	// declare the native columns
	addr := &Addresses{
		Address:              address,
		AddressUntrimmed:     addressUntrimmed,
		IsAddress:            createCol("IS_ADDRESS"),
		Col16:                verifiercol.NewConstantCol(field.NewElement(16), size, "ecdsa-col16"),
		IsAddressHiEcRec:     comp.InsertCommit(0, ifaces.ColIDf("ISADRESS_HI_ECREC"), ecRecSize),
		IsAddressFromEcRec:   createCol("ISADRESS_FROM_ECREC"),
		IsAddressFromTxnData: createCol("ISADRESS_FROM_TXNDATA"),
		HashNum:              createCol("HASH_NUM"),
	}

	addr.IsAddressLoEcRec = dedicated.ManuallyShift(comp, addr.IsAddressHiEcRec, -1, "ISADRESS_LO_ECREC")

	td.csTxnData(comp)

	// addresses are fetched from two arithmetization modules (ecRecover and txn-data)
	// IsAddress = IsAdressFromEcRec + IsAdressFromTxnData
	comp.InsertGlobal(0, ifaces.QueryIDf("Format_IsAddress"),
		sym.Sub(addr.IsAddress, sym.Add(addr.IsAddressFromEcRec, addr.IsAddressFromTxnData)))

	commoncs.MustBeBinary(comp, addr.IsAddress)
	commoncs.MustBeBinary(comp, addr.IsAddressFromEcRec)
	commoncs.MustBeBinary(comp, addr.IsAddressFromTxnData)
	commoncs.MustZeroWhenInactive(comp, ac.IsActive,
		addr.IsAddress,
		addr.HashNum,
	)

	// check the trimming of address limb columns to addressUntrimmed limb columns
	addr.csAddressTrimming(comp)

	// check that IsAddressHiEcRec is well-formed
	addr.csIsAddressHiEcRec(comp, ecRec)

	// projection from ecRecover to address columns
	// ecdata is already projected over our ecRecover. Thus, we only project from our ecrecover.

	// Check that first 6 elements (trimmed 12 bytes) of address higher part are all 0
	for i := 0; i < addressTrimmedColumns; i++ {
		comp.InsertGlobal(0, ifaces.QueryIDf("Trimmed_Bytes_Zeros_%d", i),
			sym.Mul(addr.IsAddressHiEcRec, addr.IsAddressFromEcRec, ecRec.Limb[i]),
		)
	}

	comp.InsertProjection(ifaces.QueryIDf("Project_AddressHi_EcRec"),
		query.ProjectionInput{
			ColumnA: ecRec.Limb[addressTrimmedColumns:],
			ColumnB: addr.Address[:addressHiColumns],
			FilterA: addr.IsAddressHiEcRec,
			FilterB: addr.IsAddressFromEcRec,
		},
	)

	comp.InsertProjection(ifaces.QueryIDf("Project_AddressLo_EcRec"),
		query.ProjectionInput{
			ColumnA: ecRec.Limb[:],
			ColumnB: addr.Address[addressHiColumns:],
			FilterA: column.Shift(addr.IsAddressHiEcRec, -1),
			FilterB: addr.IsAddressFromEcRec,
		},
	)

	// projection from txn-data to address columns
	comp.InsertProjection(ifaces.QueryIDf("Project_AddressUntrimmed_TxnData"),
		query.ProjectionInput{
			ColumnA: td.From[:],
			ColumnB: addressUntrimmed[:],
			FilterA: td.IsFrom,
			FilterB: addr.IsAddressFromTxnData,
		},
	)

	// impose that hashNum = ac.ID + 1
	comp.InsertGlobal(0, ifaces.QueryIDf("Hash_NUM_IS_ID"),
		sym.Mul(ac.IsActive,
			sym.Sub(addr.HashNum, ac.ID, 1)),
	)

	// assign the keccak provider
	addr.Provider = addr.GetProvider(comp, addr.HashNum, ac.UnalignedGnarkData)

	return addr
}

// It checks the well-forming of IsAddressHiEcRec
func (addr *Addresses) csIsAddressHiEcRec(comp *wizard.CompiledIOP, ecRec *EcRecover) {
	// if EcRecoverIsRes[i] == 1 and EcRecover[i+1] == 1 ---> isAddressHiEcRec[i] = 1
	comp.InsertGlobal(0, ifaces.QueryIDf("Is_AddressHi_EcRec_1"),
		sym.Mul(ecRec.EcRecoverIsRes, column.Shift(ecRec.EcRecoverIsRes, 1),
			sym.Sub(1, addr.IsAddressHiEcRec)))

	// if EcRecoverIsRes[i] == 0  ---> isAddressHiEcRec[i] = 0
	comp.InsertGlobal(0, ifaces.QueryIDf("Is_AddressHi_EcRec_2"),
		sym.Mul(sym.Sub(1, ecRec.EcRecoverIsRes), addr.IsAddressHiEcRec))

	// if EcRecoverIsRes[i] == 1 and EcRecover[i+1] == 0 ---> isAddressHiEcRec[i] = 0
	comp.InsertGlobal(0, ifaces.QueryIDf("Is_AddressHi_EcRec_3"),
		sym.Mul(ecRec.EcRecoverIsRes, sym.Sub(1, column.Shift(ecRec.EcRecoverIsRes, 1)),
			addr.IsAddressHiEcRec))
}

// The constraints for trimming the addressUntrimmed to address
func (addr *Addresses) csAddressTrimming(comp *wizard.CompiledIOP) {
	for i := 0; i < common.NbLimbEthAddress; i++ {
		comp.InsertGlobal(0, ifaces.QueryIDf("Address_Trimming_%d", i), sym.Sub(addr.Address[i], addr.AddressUntrimmed[i+addressTrimmedColumns]))
	}
}

// It builds a provider from  public key extracted from Gnark-Data (as hash input) and addresses (as output).
// the consistency check is then deferred to the keccak module.
func (addr *Addresses) GetProvider(comp *wizard.CompiledIOP, id ifaces.Column, uaGnark *UnalignedGnarkData) generic.GenericByteModule {
	// generate a generic byte Module as keccak provider.
	provider := addr.buildGenericModule(id, uaGnark)
	return provider
}

// It builds a GenericByteModule from Address columns and Public-Key/GnarkData columns.
func (addr *Addresses) buildGenericModule(id ifaces.Column, uaGnark *UnalignedGnarkData) (pkModule generic.GenericByteModule) {
	pkModule.Data = generic.GenDataModule{
		HashNum: id,
		Limbs:   uaGnark.GnarkData[:],

		// a column of all 16, since all the bytes of public key are used in hashing
		NBytes: addr.Col16,
		Index:  uaGnark.GnarkPublicKeyIndex,
		ToHash: uaGnark.IsPublicKey,
	}

	pkModule.Info = generic.GenInfoModule{
<<<<<<< HEAD
		HashHi:   addr.Address[:addressHiColumns],
		HashLo:   addr.Address[addressHiColumns:],
		IsHashHi: addr.IsAddress,
		IsHashLo: addr.IsAddress,
=======
		Hash:   addr.address[:],
		IsHash: addr.isAddress,
>>>>>>> 4a6dab3e (Implement "glue" for sha2 module over koalabear field)
	}
	return pkModule
}

// It assigns the native columns specific to the submodule.
func (addr *Addresses) assignAddress(
	run *wizard.ProverRuntime,
	nbEcRecover, size int,
	ac *antichamber,
	ecRec *EcRecover,
	uaGnark *UnalignedGnarkData,
	td *txnData,
) {
	// assign td.isFrom
	td.Pa_IsZero.Run(run)

	// assign HashNum
	var (
		one      = field.One()
		id       = ac.ID.GetColAssignment(run).IntoRegVecSaveAlloc()
		isActive = ac.IsActive.GetColAssignment(run).IntoRegVecSaveAlloc()
		hashNum  = common.NewVectorBuilder(addr.HashNum)
	)

	for row := range id {
		if isActive[row].IsOne() {
			f := *new(field.Element).Add(&id[row], &one)
			hashNum.PushField(f)
		} else {
			hashNum.PushInt(0)
		}
	}

	hashNum.PadAndAssign(run)
	addr.assignMainColumns(run, nbEcRecover, size, uaGnark)
	addr.assignHelperColumns(run, ecRec)
}

// It assigns the main columns
func (addr *Addresses) assignMainColumns(
	run *wizard.ProverRuntime,
	nbEcRecover, size int,
	uaGnark *UnalignedGnarkData,
) {
	pkModule := addr.buildGenericModule(addr.HashNum, uaGnark)

	split := splitAt(nbEcRecover)
	n := nbRowsPerEcRec

	streams := pkModule.Data.ScanStreams(run)
	permTrace := keccak.GenerateTrace(streams)

	// Initialize an array of addressUntrimmedColumns limbs columns
	addressUntrimmedColumns := make([][]field.Element, 0, common.NbLimbU256)
	for i := 0; i < common.NbLimbU256; i++ {
		addressUntrimmedColumns = append(addressUntrimmedColumns, make([]field.Element, 0, len(permTrace.HashOutPut)))
	}

	// Initialize an array of address limbs columns
	addressColumns := make([][]field.Element, 0, common.NbLimbEthAddress)
	for i := 0; i < common.NbLimbEthAddress; i++ {
		addressColumns = append(addressColumns, make([]field.Element, 0, len(permTrace.HashOutPut)))
	}

	isHash := make([]field.Element, 0, len(permTrace.HashOutPut))
	for _, digest := range permTrace.HashOutPut {
		if len(addressColumns[len(addressColumns)-1]) == split {
			n = nbRowsPerTxSign
		}

		// Initialize limb values for each column of addressUntrimmed
		addressUntrimmed := common.SplitBytes(digest[:])
		for j, limb := range addressUntrimmed {
			var element field.Element
			element.SetBytes(limb[:])

			repeat := vector.Repeat(element, n)
			addressUntrimmedColumns[j] = append(addressUntrimmedColumns[j], repeat...)
		}

		// Initialize limb values for each column of address
		address := common.SplitBytes(digest[addressTrimmedBytes:])
		for j, limb := range address {
			var element field.Element
			element.SetBytes(limb[:])

			repeat := vector.Repeat(element, n)
			addressColumns[j] = append(addressColumns[j], repeat...)
		}

		repeatIsTxHash := vector.Repeat(field.Zero(), n-1)

		isHash = append(isHash, field.One())
		isHash = append(isHash, repeatIsTxHash...)
	}

	isFromEcRec := isHash[:split]
	isFromTxnData := vector.Repeat(field.Zero(), split)
	isFromTxnData = append(isFromTxnData, isHash[split:]...)

	// Assign values to columns
	// We do a reverse of address columns since the address was storing in big-endian format.
	for i := 0; i < common.NbLimbU256; i++ {
		run.AssignColumn(addr.AddressUntrimmed[i].GetColID(), smartvectors.RightZeroPadded(addressUntrimmedColumns[i], size))
	}

	for i := 0; i < common.NbLimbEthAddress; i++ {
		run.AssignColumn(addr.Address[i].GetColID(), smartvectors.RightZeroPadded(addressColumns[i], size))
	}

	run.AssignColumn(addr.IsAddress.GetColID(), smartvectors.RightZeroPadded(isHash, size))
	run.AssignColumn(addr.IsAddressFromEcRec.GetColID(), smartvectors.RightZeroPadded(isFromEcRec, size))
	run.AssignColumn(addr.IsAddressFromTxnData.GetColID(), smartvectors.RightZeroPadded(isFromTxnData, size))
}

// It assigns the helper columns
func (addr *Addresses) assignHelperColumns(run *wizard.ProverRuntime, ecRec *EcRecover) {

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

	run.AssignColumn(addr.IsAddressHiEcRec.GetColID(), smartvectors.NewRegular(col))

	// this calls the auto-assign feature of the
	_ = addr.IsAddressLoEcRec.GetColAssignment(run)
}

// It indicates the row where ecrecover and txSignature are split.
func splitAt(nbEcRecover int) int {
	return nbEcRecover * nbRowsPerEcRec
}

func (td *txnData) csTxnData(comp *wizard.CompiledIOP) {

	//  isFrom == 1 iff ct==0 && User==1 && HubSelector==1
	//  (Ct-0)^2 + (User-1)^2 + (HubSelector-1)^2 = 0
	condition := sym.Add(
		sym.Square(sym.Sub(td.Ct, 0)),
		sym.Square(sym.Sub(td.User, 1)),
		sym.Square(sym.Sub(td.Selector, 1)),
	)

	td.IsFrom, td.Pa_IsZero = dedicated.IsZero(comp, condition).GetColumnAndProverAction()

}

// txndata represents the txn_data module from the arithmetization side.
type txnData struct {
	From [common.NbLimbU256]ifaces.Column
	Ct   ifaces.Column

	// helper column
	IsFrom    ifaces.Column
	Pa_IsZero wizard.ProverAction

	User     ifaces.Column
	Selector ifaces.Column
}
