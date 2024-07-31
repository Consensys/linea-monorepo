package antichamber

import (
	"github.com/consensys/zkevm-monorepo/prover/maths/common/smartvectors"
	"github.com/consensys/zkevm-monorepo/prover/maths/field"
	"github.com/consensys/zkevm-monorepo/prover/protocol/dedicated/plonk"
	"github.com/consensys/zkevm-monorepo/prover/protocol/ifaces"
	"github.com/consensys/zkevm-monorepo/prover/protocol/wizard"
	"github.com/consensys/zkevm-monorepo/prover/utils"
	"github.com/consensys/zkevm-monorepo/prover/zkevm/prover/ecrecover"
)

const (
	ROUND_NR                 = 0
	NAME_ANTICHAMBER         = "ECDSA_ANTICHAMBER"
	NAME_ECRECOVER           = "ECDSA_ANTICHAMBER_ECRECOVER"
	NAME_ADDRESSES           = "ECDSA_ANTICHAMBER_ADDRESSES"
	NAME_TXSIGNATURE         = "ECDSA_ANTICHAMBER_TXSIGNATURE"
	NAME_UNALIGNED_GNARKDATA = "ECDSA_ANTICHAMBER_UNALIGNED_GNARK_DATA"
	NAME_GNARK_DATA          = "ECDSA_ANTICHAMBER_GNARK_DATA"
)

const (
	// keccak digest is on size 32bytes
	halfDigest = 16

	// number of public inputs gnark circuit takes as a public witness
	nbRowsPerGnarkPushing = 14

	// 10 rows for fetching from arithmetization-columns
	nbRowsPerEcRecFetching = 10
	nbRowsPerEcRec         = nbRowsPerEcRecFetching + nbRowsPerGnarkPushing

	// 1 row for fetching from RLP
	nbRowsPerTxSignFetching = 1
	nbRowsPerTxSign         = nbRowsPerTxSignFetching + nbRowsPerGnarkPushing
)

func createColFn(comp *wizard.CompiledIOP, rootName string, size int) func(name string) ifaces.Column {
	return func(name string) ifaces.Column {
		return comp.InsertCommit(ROUND_NR, ifaces.ColIDf("%s_%s", rootName, name), size)
	}
}

type Antichamber struct {
	IsActive   ifaces.Column
	ID         ifaces.Column
	IsPushing  ifaces.Column
	IsFetching ifaces.Column
	Source     ifaces.Column // source of the data. Source=0 <> ECRecover, Source=1 <> TxSignature

	*EcRecover
	*Addresses
	*txSignature
	*UnalignedGnarkData
	AlignedGnarkData *plonk.Alignment

	// size of AntiChamber
	size int
	*Limits
}

type GnarkData struct {
	IdPerm         ifaces.Column
	GnarkIndexPerm ifaces.Column
	DataPerm       ifaces.Column
}

type Limits struct {
	MaxNbEcRecover     int
	MaxNbTx            int
	NbInputInstance    int
	NbCircuitInstances int
}

func (l *Limits) sizeAntichamber() int {
	return utils.NextPowerOfTwo(l.MaxNbEcRecover*nbRowsPerEcRec + l.MaxNbTx*nbRowsPerTxSign)
}

func NewAntichamber(comp *wizard.CompiledIOP, limits *Limits, ecSource *ecDataSource, txSource *txnData, plonkOptions []plonk.Option) *Antichamber {
	if limits.MaxNbEcRecover+limits.MaxNbTx != limits.NbInputInstance*limits.NbCircuitInstances {
		utils.Panic("the number of supported instances %v should be %v + %v", limits.NbInputInstance*limits.NbCircuitInstances, limits.MaxNbEcRecover, limits.MaxNbTx)
	}
	size := limits.sizeAntichamber()
	createCol := createColFn(comp, NAME_ANTICHAMBER, size)

	// declare the native columns
	res := &Antichamber{
		IsActive:   createCol("IS_ACTIVE"),
		ID:         createCol("ID"),
		IsPushing:  createCol("IS_PUSHING"),
		IsFetching: createCol("IS_FETCHING"),
		Source:     createCol("SOURCE"),

		size:   size,
		Limits: limits,
	}

	// declare submodules
	res.txSignature = newTxSignatures(comp, size)
	res.EcRecover = newEcRecover(comp, limits, ecSource)
	res.UnalignedGnarkData = newUnalignedGnarkData(comp, size, res.unalignedGnarkDataSource())
	res.Addresses = newAddress(comp, size, res.EcRecover, res, txSource)
	toAlign := &plonk.CircuitAlignmentInput{
		Name:               NAME_GNARK_DATA,
		Round:              ROUND_NR,
		DataToCircuit:      res.UnalignedGnarkData.GnarkData,
		DataToCircuitMask:  res.IsPushing,
		Circuit:            ecrecover.NewMultiEcRecoverCircuit(limits.NbInputInstance),
		InputFiller:        ecrecover.InputFiller,
		PlonkOptions:       plonkOptions,
		NbCircuitInstances: limits.NbCircuitInstances,
	}
	res.AlignedGnarkData = plonk.DefineAlignment(comp, toAlign)

	// root module constraints
	res.csIsActive(comp)
	res.csZeroWhenInactive(comp)
	res.csConsistentPushingFetching(comp)
	res.csIDSequential(comp)
	res.csSource(comp)
	res.csTransitions(comp)

	// consistency with submodules
	// ecrecover
	res.EcRecover.csConstrainAuxProjectionMaskConsistency(comp, res.Source, res.IsFetching)

	return res
}

// Assign assings all values in the antichamber. This includes assignment of all
// the submodules:
//   - EcRecover
//   - TxSignature
//   - Addresses
//   - UnalignedGnarkData
//   - AlignedGnarkData
//
// As the initial data is copied from the EC_DATA arithmetization module, then
// it has to be provided as an input.
func (ac *Antichamber) Assign(run *wizard.ProverRuntime, ecSrc *ecDataSource, txSource *txnData, txGet TxSignatureGetter) {
	nbActualEcRecover := ecSrc.nbActualInstances(run)
	ac.assignAntichamber(run, nbActualEcRecover)
	ac.EcRecover.Assign(run, ecSrc)
	ac.txSignature.assignTxSignature(run, nbActualEcRecover, ac.size)
	ac.UnalignedGnarkData.Assign(run, ac.unalignedGnarkDataSource(), txGet)
	ac.Addresses.assignAddress(run, nbActualEcRecover, ac.size, ac, ac.EcRecover, ac.UnalignedGnarkData, txSource)
	ac.AlignedGnarkData.Assign(run)
}

// assignAntichamber assigns the values values in the main part of the antichamber, namely columns:
//   - IsActive
//   - ID
//   - IsPushing
//   - IsFetching
//   - Source
//
// The assignment depends on the number of defined EcRecover and TxSignature instances.
func (ac *Antichamber) assignAntichamber(run *wizard.ProverRuntime, nbEcRecInstances int) {
	if nbRowsPerEcRec*ac.MaxNbEcRecover+nbRowsPerTxSign*ac.MaxNbTx > ac.size {
		utils.Panic("not enough space in antichamber to store all the data. Need %d, got %d", 24*ac.MaxNbEcRecover+15*ac.MaxNbTx, ac.size)
	}
	// prepare root module columns
	// for ecrecover case we need 10+14 rows (fetchin and pushing). For TX we need 1+14

	// allocate the columns for preparing the assignment
	resIsActive := make([]field.Element, ac.size)
	resID := make([]field.Element, ac.size)
	resIsPushing := make([]field.Element, ac.size)
	resIsFetching := make([]field.Element, ac.size)
	resSource := make([]field.Element, ac.size)

	var idxInstance uint64
	for i := 0; i < nbEcRecInstances; i++ {
		for j := 0; j < nbRowsPerEcRec; j++ {
			resIsActive[i*nbRowsPerEcRec+j] = field.NewElement(1)
			resID[i*nbRowsPerEcRec+j] = field.NewElement(idxInstance)
			resSource[i*nbRowsPerEcRec+j].Set(&SOURCE_ECRECOVER)
		}
		for j := 0; j < nbRowsPerEcRecFetching; j++ {
			resIsFetching[i*nbRowsPerEcRec+j] = field.NewElement(1)
		}
		for j := 0; j < nbRowsPerGnarkPushing; j++ {
			resIsPushing[i*nbRowsPerEcRec+nbRowsPerEcRecFetching+j] = field.NewElement(1)
		}
		idxInstance++
	}

	for i := 0; i < ac.MaxNbTx; i++ {
		for j := 0; j < nbRowsPerTxSign; j++ {
			resIsActive[nbRowsPerEcRec*nbEcRecInstances+i*nbRowsPerTxSign+j] = field.NewElement(1)
			resID[nbRowsPerEcRec*nbEcRecInstances+i*nbRowsPerTxSign+j] = field.NewElement(idxInstance)
			resSource[nbRowsPerEcRec*nbEcRecInstances+i*nbRowsPerTxSign+j].Set(&SOURCE_TX)
		}
		resIsFetching[nbRowsPerEcRec*nbEcRecInstances+i*nbRowsPerTxSign] = field.NewElement(1)
		for j := 0; j < nbRowsPerGnarkPushing; j++ {
			resIsPushing[nbRowsPerEcRec*nbEcRecInstances+i*nbRowsPerTxSign+nbRowsPerTxSignFetching+j] = field.NewElement(1)
		}
		idxInstance++
	}

	run.AssignColumn(ac.IsActive.GetColID(), smartvectors.NewRegular(resIsActive))
	run.AssignColumn(ac.ID.GetColID(), smartvectors.NewRegular(resID))
	run.AssignColumn(ac.IsPushing.GetColID(), smartvectors.NewRegular(resIsPushing))
	run.AssignColumn(ac.IsFetching.GetColID(), smartvectors.NewRegular(resIsFetching))
	run.AssignColumn(ac.Source.GetColID(), smartvectors.NewRegular(resSource))
}
