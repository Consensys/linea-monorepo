package ecdsa

import (
	"github.com/consensys/linea-monorepo/prover/maths/common/smartvectors"
	"github.com/consensys/linea-monorepo/prover/maths/field"
	"github.com/consensys/linea-monorepo/prover/protocol/dedicated/plonk"
	"github.com/consensys/linea-monorepo/prover/protocol/ifaces"
	"github.com/consensys/linea-monorepo/prover/protocol/wizard"
	"github.com/consensys/linea-monorepo/prover/utils"
	"github.com/consensys/linea-monorepo/prover/zkevm/prover/hash/generic"
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

type antichamberInput struct {
	ecSource     *ecDataSource
	txSource     *txnData
	rlpTxn       generic.GenDataModule
	settings     *Settings
	plonkOptions []any
}

type antichamber struct {
	Inputs     *antichamberInput
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

	// providers for keccak, Providers contain the inputs and outputs of keccak hash.
	Providers []generic.GenericByteModule
}

type Settings struct {
	MaxNbEcRecover     int
	MaxNbTx            int
	NbInputInstance    int
	NbCircuitInstances int
}

func (l *Settings) sizeAntichamber() int {
	return utils.NextPowerOfTwo(l.MaxNbEcRecover*nbRowsPerEcRec + l.MaxNbTx*nbRowsPerTxSign)
}

func newAntichamber(comp *wizard.CompiledIOP, inputs *antichamberInput) *antichamber {

	settings := inputs.settings
	if settings.MaxNbEcRecover+settings.MaxNbTx > settings.NbInputInstance*settings.NbCircuitInstances {
		utils.Panic("the number of supported instances %v should be %v + %v", settings.NbInputInstance*settings.NbCircuitInstances, settings.MaxNbEcRecover, settings.MaxNbTx)
	}
	size := inputs.settings.sizeAntichamber()
	createCol := createColFn(comp, NAME_ANTICHAMBER, size)

	// declare the native columns
	res := &antichamber{

		IsActive:   createCol("IS_ACTIVE"),
		ID:         createCol("ID"),
		IsPushing:  createCol("IS_PUSHING"),
		IsFetching: createCol("IS_FETCHING"),
		Source:     createCol("SOURCE"),
		Inputs:     inputs,

		size: size,
	}

	// declare submodules
	txSignInputs := txSignatureInputs{
		RlpTxn: inputs.rlpTxn,
		ac:     res,
	}
	res.txSignature = newTxSignatures(comp, txSignInputs)
	res.EcRecover = newEcRecover(comp, inputs.settings, inputs.ecSource)
	res.UnalignedGnarkData = newUnalignedGnarkData(comp, size, res.unalignedGnarkDataSource())
	res.Addresses = newAddress(comp, size, res.EcRecover, res, inputs.txSource)
	toAlign := &plonk.CircuitAlignmentInput{
		Name:               NAME_GNARK_DATA,
		Round:              ROUND_NR,
		DataToCircuit:      res.UnalignedGnarkData.GnarkData,
		DataToCircuitMask:  res.IsPushing,
		Circuit:            newMultiEcRecoverCircuit(settings.NbInputInstance),
		InputFiller:        inputFiller,
		PlonkOptions:       inputs.plonkOptions,
		NbCircuitInstances: settings.NbCircuitInstances,
	}
	res.AlignedGnarkData = plonk.DefineAlignment(comp, toAlign)

	// root module constraints
	res.csIsActiveActivation(comp)
	res.csZeroWhenInactive(comp)
	res.csConsistentPushingFetching(comp)
	res.csIDSequential(comp)
	res.csSource(comp)
	res.csTransitions(comp)

	// consistency with submodules
	// ecrecover
	res.EcRecover.csConstrainAuxProjectionMaskConsistency(comp, res.Source, res.IsFetching)

	// assign keccak providers
	res.Providers = append([]generic.GenericByteModule{res.Addresses.provider}, res.txSignature.provider)

	return res
}

// assign assings all values in the antichamber. This includes assignment of all
// the submodules:
//   - EcRecover
//   - TxSignature
//   - Addresses
//   - UnalignedGnarkData
//   - AlignedGnarkData
//
// As the initial data is copied from the EC_DATA arithmetization module, then
// it has to be provided as an input.
func (ac *antichamber) assign(run *wizard.ProverRuntime, txGet TxSignatureGetter, nbTx int) {
	var (
		ecSrc             = ac.Inputs.ecSource
		txSource          = ac.Inputs.txSource
		nbActualEcRecover = ecSrc.nbActualInstances(run)
	)
	ac.assignAntichamber(run, nbActualEcRecover, nbTx)
	ac.EcRecover.Assign(run, ecSrc)
	ac.txSignature.assignTxSignature(run, nbActualEcRecover)
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
func (ac *antichamber) assignAntichamber(run *wizard.ProverRuntime, nbEcRecInstances, nbTxInstances int) {

	var (
		maxNbEcRecover = ac.Inputs.settings.MaxNbEcRecover
		maxNbTx        = ac.Inputs.settings.MaxNbTx
	)

	if nbRowsPerEcRec*maxNbEcRecover+nbRowsPerTxSign*maxNbTx > ac.size {
		utils.Panic("not enough space in antichamber to store all the data. Need %d, got %d", 24*maxNbEcRecover+15*maxNbTx, ac.size)
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

	for i := 0; i < nbTxInstances; i++ {
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
