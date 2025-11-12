package publicInput

import (
	"fmt"

	"github.com/consensys/linea-monorepo/prover/protocol/accessors"
	"github.com/consensys/linea-monorepo/prover/protocol/ifaces"
	"github.com/consensys/linea-monorepo/prover/protocol/query"
	"github.com/consensys/linea-monorepo/prover/protocol/wizard"
	"github.com/consensys/linea-monorepo/prover/utils"
	"github.com/consensys/linea-monorepo/prover/utils/types"
	pcommon "github.com/consensys/linea-monorepo/prover/zkevm/prover/common"
	"github.com/consensys/linea-monorepo/prover/zkevm/prover/hash/generic"
	"github.com/consensys/linea-monorepo/prover/zkevm/prover/hash/importpad"
	pack "github.com/consensys/linea-monorepo/prover/zkevm/prover/hash/packing"
	arith "github.com/consensys/linea-monorepo/prover/zkevm/prover/publicInput/arith_struct"
	edc "github.com/consensys/linea-monorepo/prover/zkevm/prover/publicInput/execution_data_collector"
	fetch "github.com/consensys/linea-monorepo/prover/zkevm/prover/publicInput/fetchers_arithmetization"
	"github.com/consensys/linea-monorepo/prover/zkevm/prover/publicInput/logs"
	"github.com/consensys/linea-monorepo/prover/zkevm/prover/statemanager/statesummary"
	"github.com/ethereum/go-ethereum/common"
)

const (
	DataNbBytes                  = "DataNbBytes"
	DataChecksum                 = "DataChecksum"
	L2MessageHash                = "L2MessageHash"
	InitialStateRootHash         = "InitialStateRootHash"
	FinalStateRootHash           = "FinalStateRootHash"
	InitialBlockNumber           = "InitialBlockNumber"
	FinalBlockNumber             = "FinalBlockNumber"
	InitialBlockTimestamp        = "InitialBlockTimestamp"
	FinalBlockTimestamp          = "FinalBlockTimestamp"
	FirstRollingHashUpdate       = "FirstRollingHashUpdate"
	LastRollingHashUpdate        = "LastRollingHashUpdate"
	FirstRollingHashUpdateNumber = "FirstRollingHashUpdateNumber"
	LastRollingHashNumberUpdate  = "LastRollingHashNumberUpdate"
	ChainID                      = "ChainID"
	NBytesChainID                = "NBytesChainID"
	L2MessageServiceAddr         = "L2MessageServiceAddr"
)

// PublicInput collects a number of submodules responsible for collecting the
// wizard witness data holding the public inputs of the execution circuit.
type PublicInput struct {
	Inputs             InputModules
	Aux                AuxiliaryModules
	TimestampFetcher   *fetch.TimestampFetcher
	RootHashFetcher    *fetch.RootHashFetcher
	RollingHashFetcher *logs.RollingSelector
	LogHasher          logs.LogHasher
	ExecMiMCHasher     edc.MIMCHasher
	DataNbBytes        ifaces.Column
	ChainID            [pcommon.NbLimbU128]ifaces.Column
	ChainIDNBytes      ifaces.Column
	Extractor          FunctionalInputExtractor
}

// AuxiliaryModules are intermediary modules needed to assign the data in the PublicInput
type AuxiliaryModules struct {
	FetchedL2L1, FetchedRollingMsg, FetchedRollingHash logs.ExtractedData
	LogSelectors                                       logs.Selectors
	BlockTxnMetadata                                   fetch.BlockTxnMetadata
	TxnDataFetcher                                     fetch.TxnDataFetcher
	RlpTxnFetcher                                      fetch.RlpTxnFetcher
	ExecDataCollector                                  *edc.ExecutionDataCollector
	ExecDataCollectorPadding                           wizard.ProverAction
	ExecDataCollectorPacking                           pack.Packing
}

// Settings contains options for proving and verifying that the public inputs are computed properly.
type Settings struct {
	Name string
}

// InputModules groups several arithmetization modules needed to compute the public input.
type InputModules struct {
	BlockData    *arith.BlockDataCols
	TxnData      *arith.TxnData
	RlpTxn       *arith.RlpTxn
	LogCols      logs.LogColumns
	StateSummary *statesummary.Module
}

// NewPublicInputZkEVM constructs and returns a [PublicInput] module using the
// right columns from the arithmetization by their names.
func NewPublicInputZkEVM(comp *wizard.CompiledIOP, settings *Settings, ss *statesummary.Module) PublicInput {

	getCol := func(s string) ifaces.Column {
		return comp.Columns.GetHandle(ifaces.ColID(s))
	}

	settings.Name = "PUBLIC_INPUT"

	inputModules := newPublicInput(
		comp,
		&InputModules{
			BlockData: &arith.BlockDataCols{
				RelBlock:   getCol("blockdata.REL_BLOCK"),
				Inst:       getCol("blockdata.INST"),
				Ct:         getCol("blockdata.CT"),
				DataHi:     getCol("blockdata.DATA_HI"),
				DataLo:     getCol("blockdata.DATA_LO"),
				FirstBlock: getCol("blockdata.FIRST_BLOCK_NUMBER"),
			},
			TxnData: &arith.TxnData{
				AbsTxNum:        getCol("txndata.USER_TXN_NUMBER"),
				AbsTxNumMax:     getCol("txndata.prover___USER_TXN_NUMBER_MAX"),
				Ct:              getCol("txndata.CT"),
				FromHi:          getCol("txndata.hubFROM_ADDRESS_HI_xor_rlpCHAIN_ID"),
				FromLo:          getCol("txndata.computationARG_1_LO_xor_hubFROM_ADDRESS_LO_xor_rlpTO_ADDRESS_LO"),
				IsLastTxOfBlock: getCol("txndata.prover___IS_LAST_USER_TXN_OF_BLOCK"),
				RelBlock:        getCol("txndata.BLK_NUMBER"),
				RelTxNum:        getCol("txndata.prover___RELATIVE_USER_TXN_NUMBER"),
				RelTxNumMax:     getCol("txndata.prover___RELATIVE_USER_TXN_NUMBER_MAX"),
				USER:            getCol("txndata.USER"),
				Selector:        getCol("txndata.HUB"),
				SYSI:            getCol("txndata.SYSI"),
				SYSF:            getCol("txndata.SYSF"),
			},
			RlpTxn: &arith.RlpTxn{
				AbsTxNum:       getCol("rlptxn.USER_TXN_NUMBER"),
				AbsTxNumMax:    getCol("rlptxn.prover___USER_TXN_NUMBER_MAX"),
				ToHashByProver: getCol("rlptxn.TO_HASH_BY_PROVER"),
				Limb:           getCol("rlptxn.cmpLIMB"),
				NBytes:         getCol("rlptxn.cmpLIMB_SIZE"),
				TxnPerspective: getCol("rlptxn.TXN"),
				ChainID:        getCol("rlptxn.txnCHAIN_ID"),
			},
			LogCols: logs.LogColumns{
				IsLog0:       getCol("loginfo.IS_LOG_X_0"),
				IsLog1:       getCol("loginfo.IS_LOG_X_1"),
				IsLog2:       getCol("loginfo.IS_LOG_X_2"),
				IsLog3:       getCol("loginfo.IS_LOG_X_3"),
				IsLog4:       getCol("loginfo.IS_LOG_X_4"),
				AbsLogNum:    getCol("loginfo.ABS_LOG_NUM"),
				AbsLogNumMax: getCol("loginfo.ABS_LOG_NUM_MAX"),
				Ct:           getCol("loginfo.CT"),
				DataHi:       getCol("loginfo.DATA_HI"),
				DataLo:       getCol("loginfo.DATA_LO"),
				TxEmitsLogs:  getCol("loginfo.TXN_EMITS_LOGS"),
			},
			StateSummary: ss,
			TxnData: &arith.TxnData{
				AbsTxNum:        getCol("txndata.ABS_TX_NUM"),
				AbsTxNumMax:     getCol("txndata.ABS_TX_NUM_MAX"),
				Ct:              getCol("txndata.CT"),
				IsLastTxOfBlock: getCol("txndata.IS_LAST_TX_OF_BLOCK"),
				RelBlock:        getCol("txndata.REL_BLOCK"),
				RelTxNum:        getCol("txndata.REL_TX_NUM"),
				RelTxNumMax:     getCol("txndata.REL_TX_NUM_MAX"),
			},
			RlpTxn: &arith.RlpTxn{
				AbsTxNum:       getCol("rlptxn.ABS_TX_NUM"),
				AbsTxNumMax:    getCol("rlptxn.ABS_TX_NUM_INFINY"),
				ToHashByProver: getCol("rlptxn.TO_HASH_BY_PROVER"),
				NBytes:         getCol("rlptxn.nBYTES"),
				Done:           getCol("rlptxn.DONE"),
				IsPhaseChainID: getCol("rlptxn.IS_PHASE_CHAIN_ID"),
			},
			LogCols: logs.LogColumns{
				IsLog0:       getCol("loginfo.IS_LOG_X_0"),
				IsLog1:       getCol("loginfo.IS_LOG_X_1"),
				IsLog2:       getCol("loginfo.IS_LOG_X_2"),
				IsLog3:       getCol("loginfo.IS_LOG_X_3"),
				IsLog4:       getCol("loginfo.IS_LOG_X_4"),
				AbsLogNum:    getCol("loginfo.ABS_LOG_NUM"),
				AbsLogNumMax: getCol("loginfo.ABS_LOG_NUM_MAX"),
				Ct:           getCol("loginfo.CT"),
				TxEmitsLogs:  getCol("loginfo.TXN_EMITS_LOGS"),
			},
			StateSummary: ss,
		},
	)

	for i := range inputModules.RlpTxn.Limbs {
		inputModules.RlpTxn.Limbs[i] = getCol(fmt.Sprintf("rlptxn.LIMB_%d", i))
	}

	for i := range inputModules.BlockData.FirstBlock {
		inputModules.BlockData.FirstBlock[i] = getCol(fmt.Sprintf("blockdata.FIRST_BLOCK_NUMBER_%d", i))
	}

	for i := range pcommon.NbLimbU256 {
		inputModules.BlockData.Data[i] = getCol(fmt.Sprintf("blockdata.DATA_%d", i))
		inputModules.LogCols.Data[i] = getCol(fmt.Sprintf("loginfo.DATA_%d", i))
	}

	for i := range inputModules.TxnData.From {
		inputModules.TxnData.From[i] = getCol(fmt.Sprintf("txndata.FROM_%d", i))
	}

	return newPublicInput(comp, inputModules, *settings)
}

// newPublicInput receives as input a series of modules and returns a *PublicInput and
// an *AuxiliaryModules struct. The AuxiliaryModules are intermediary modules needed to
// both define and assign the PublicInput.
func newPublicInput(
	comp *wizard.CompiledIOP,
	inp *InputModules,
	settings Settings,
) PublicInput {

	if len(settings.Name) == 0 {
		utils.Panic("no name was provided in settings: %++v", settings)
	}

	// Timestamps
	timestampFetcher := fetch.NewTimestampFetcher(comp, "PUBLIC_INPUT_TIMESTAMP_FETCHER", inp.BlockData)
	fetch.DefineTimestampFetcher(comp, timestampFetcher, "PUBLIC_INPUT_TIMESTAMP_FETCHER", inp.BlockData)

	// Logs: Fetchers, Selectors and Hasher
	fetchedL2L1 := logs.NewExtractedData(comp, inp.LogCols.Ct.Size(), "PUBLIC_INPUT_L2L1LOGS")
	fetchedRollingMsg := logs.NewExtractedData(comp, inp.LogCols.Ct.Size(), "PUBLIC_INPUT_ROLLING_MSG")
	fetchedRollingHash := logs.NewExtractedData(comp, inp.LogCols.Ct.Size(), "PUBLIC_INPUT_ROLLING_HASH")
	logSelectors := logs.NewSelectorColumns(comp, inp.LogCols)
	logHasherL2l1 := logs.NewLogHasher(comp, inp.LogCols.Ct.Size(), "PUBLIC_INPUT_L2L1LOGS")
	rollingSelector := logs.NewRollingSelector(comp, "PUBLIC_INPUT_ROLLING_SEL", fetchedRollingHash.Data[0].Size())

	// Define Logs: Fetchers, Selectors and Hasher
	logs.DefineExtractedData(comp, inp.LogCols, logSelectors, fetchedL2L1, logs.L2L1)
	logs.DefineExtractedData(comp, inp.LogCols, logSelectors, fetchedRollingMsg, logs.RollingMsgNo)
	logs.DefineExtractedData(comp, inp.LogCols, logSelectors, fetchedRollingHash, logs.RollingHash)
	logs.DefineHasher(comp, logHasherL2l1, "PUBLIC_INPUT_L2L1LOGS", fetchedL2L1)
	logs.DefineRollingSelector(comp, rollingSelector, "PUBLIC_INPUT_ROLLING_SEL", fetchedRollingHash, fetchedRollingMsg)

	// RootHash fetcher from the StateSummary
	rootHashFetcher := fetch.NewRootHashFetcher(comp, "PUBLIC_INPUT_ROOT_HASH_FETCHER", inp.StateSummary.IsActive.Size())
	fetch.DefineRootHashFetcher(comp, rootHashFetcher, "PUBLIC_INPUT_ROOT_HASH_FETCHER", *inp.StateSummary)

	// Metadata fetcher
	blockTxnMeta := fetch.NewBlockTxnMetadata(comp, "BLOCK_TX_METADATA", inp.TxnData)
	fetch.DefineBlockTxnMetaData(comp, &blockTxnMeta, "BLOCK_TX_METADATA", inp.TxnData)

	// TxnData fetcher
	txnDataFetcher := fetch.NewTxnDataFetcher(comp, "PUBLIC_INPUT_TXN_DATA_FETCHER", inp.TxnData)
	fetch.DefineTxnDataFetcher(comp, &txnDataFetcher, "PUBLIC_INPUT_TXN_DATA_FETCHER", inp.TxnData)

	// RlpTxn fetcher
	rlpFetcher := fetch.NewRlpTxnFetcher(comp, "PUBLIC_INPUT_RLP_TXN_FETCHER", inp.RlpTxn)
	fetch.DefineRlpTxnFetcher(comp, &rlpFetcher, "PUBLIC_INPUT_RLP_TXN_FETCHER", inp.RlpTxn)

	// ExecutionDataCollector
	limbColSize := edc.GetSummarySize(inp.TxnData, inp.RlpTxn)
	limbColSize = 4 * limbColSize // we need to artificially blow up the column size by 2, or padding will fail
	execDataCollector := edc.NewExecutionDataCollector(comp, "EXECUTION_DATA_COLLECTOR", limbColSize)
	edc.DefineExecutionDataCollector(comp, execDataCollector, "EXECUTION_DATA_COLLECTOR", timestampFetcher, blockTxnMeta, txnDataFetcher, rlpFetcher)

	// ExecutionDataCollector: Padding
	importInp := importpad.ImportAndPadInputs{
		Name: settings.Name,
		Src: generic.GenericByteModule{Data: generic.GenDataModule{
			HashNum: execDataCollector.HashNum,
			Index:   execDataCollector.Ct,
			ToHash:  execDataCollector.IsActive,
			NBytes:  execDataCollector.NoBytes,
			Limbs:   execDataCollector.Limbs[:],
		}},
		PaddingStrategy: generic.MiMCUsecase,
	}
	padding := importpad.ImportAndPad(comp, importInp, limbColSize)

	// ExecutionDataCollector: Packing
	packingInp := pack.PackingInput{
		MaxNumBlocks: execDataCollector.BlockID.Size(),
		PackingParam: generic.MiMCUsecase,
		Imported: pack.Importation{
			Limb:      padding.Limbs,
			NByte:     padding.NBytes,
			IsNewHash: padding.IsNewHash,
			IsActive:  padding.IsActive,
		},
		Name: "EXECUTION_DATA_MIMC",
	}
	packingMod := pack.NewPack(comp, packingInp)

	// ExecutionDataCollector: Hashing
	mimcHasher := edc.NewMIMCHasher(comp, packingMod.Repacked.Lanes, packingMod.Repacked.IsLaneActive, "MIMC_HASHER")
	mimcHasher.DefineHasher(comp, "EXECUTION_DATA_COLLECTOR_MIMC_HASHER")

	publicInput := PublicInput{
		TimestampFetcher:   timestampFetcher,
		RootHashFetcher:    rootHashFetcher,
		RollingHashFetcher: rollingSelector,
		LogHasher:          logHasherL2l1,
		ExecMiMCHasher:     *mimcHasher,
		DataNbBytes:        execDataCollector.FinalTotalBytesCounter,
		ChainID:            rlpFetcher.ChainID,
		ChainIDNBytes:      rlpFetcher.NBytesChainID,
		Inputs:             *inp,
		Aux: AuxiliaryModules{
			FetchedL2L1:              fetchedL2L1,
			FetchedRollingMsg:        fetchedRollingMsg,
			FetchedRollingHash:       fetchedRollingHash,
			LogSelectors:             logSelectors,
			BlockTxnMetadata:         blockTxnMeta,
			TxnDataFetcher:           txnDataFetcher,
			RlpTxnFetcher:            rlpFetcher,
			ExecDataCollector:        execDataCollector,
			ExecDataCollectorPadding: padding,
			ExecDataCollectorPacking: *packingMod,
		},
	}

	publicInput.generateExtractor(comp)

	return publicInput
}

// Assign both a PublicInput and AuxiliaryModules using data from InputModules.
// The AuxiliaryModules are intermediary modules needed to both define and assign the PublicInput.
func (pub *PublicInput) Assign(run *wizard.ProverRuntime, l2BridgeAddress common.Address, blockHashList []types.FullBytes32) {

	var (
		inp = pub.Inputs
		aux = pub.Aux
	)

	// assign the timestamp module
	fetch.AssignTimestampFetcher(run, pub.TimestampFetcher, inp.BlockData)
	// assign the log modules
	aux.LogSelectors.Assign(run, l2BridgeAddress)
	logs.AssignExtractedData(run, inp.LogCols, aux.LogSelectors, aux.FetchedL2L1, logs.L2L1)
	logs.AssignExtractedData(run, inp.LogCols, aux.LogSelectors, aux.FetchedRollingMsg, logs.RollingMsgNo)
	logs.AssignExtractedData(run, inp.LogCols, aux.LogSelectors, aux.FetchedRollingHash, logs.RollingHash)
	logs.AssignHasher(run, pub.LogHasher, aux.FetchedL2L1)
	logs.AssignRollingSelector(run, pub.RollingHashFetcher, aux.FetchedRollingHash, aux.FetchedRollingMsg)
	// assign the root hash fetcher
	fetch.AssignRootHashFetcher(run, pub.RootHashFetcher, *inp.StateSummary)
	// assign the execution data collector's necessary fetchers
	fetch.AssignBlockTxnMetadata(run, aux.BlockTxnMetadata, inp.TxnData)
	fetch.AssignTxnDataFetcher(run, aux.TxnDataFetcher, inp.TxnData)
	fetch.AssignRlpTxnFetcher(run, &aux.RlpTxnFetcher, inp.RlpTxn)
	// assign the ExecutionDataCollector
	edc.AssignExecutionDataCollector(run, aux.ExecDataCollector, pub.TimestampFetcher, aux.BlockTxnMetadata, aux.TxnDataFetcher, aux.RlpTxnFetcher, blockHashList)
	aux.ExecDataCollectorPadding.Run(run)
	aux.ExecDataCollectorPacking.Run(run)
	pub.ExecMiMCHasher.AssignHasher(run)
	pub.Extractor.Run(run)
}

// GetExtractor returns [FunctionalInputExtractor] giving access to the totality
// of the public inputs recovered by the public input module.
func (pi *PublicInput) generateExtractor(comp *wizard.CompiledIOP) {

	createNewLocalOpening := func(col ifaces.Column) query.LocalOpening {
		return comp.InsertLocalOpening(0, ifaces.QueryIDf("%s_%s", "PUBLIC_INPUT_LOCAL_OPENING", col.GetColID()), col)
	}

	pi.Extractor = FunctionalInputExtractor{
		DataNbBytes:   createNewLocalOpening(pi.DataNbBytes),
		NBytesChainID: createNewLocalOpening(pi.ChainIDNBytes),
	}

	comp.PublicInputs = append(comp.PublicInputs,
		wizard.PublicInput{Name: DataNbBytes, Acc: accessors.NewLocalOpeningAccessor(pi.Extractor.DataNbBytes, 0)},
		wizard.PublicInput{Name: NBytesChainID, Acc: accessors.NewLocalOpeningAccessor(pi.Extractor.NBytesChainID, 0)},
	)

	for i := range pcommon.NbLimbU256 {
		pi.Extractor.InitialStateRootHash[i] = createNewLocalOpening(pi.RootHashFetcher.First[i])
		pi.Extractor.FinalStateRootHash[i] = createNewLocalOpening(pi.RootHashFetcher.Last[i])
		pi.Extractor.FirstRollingHashUpdate[i] = createNewLocalOpening(pi.RollingHashFetcher.First[i])
		pi.Extractor.LastRollingHashUpdate[i] = createNewLocalOpening(pi.RollingHashFetcher.Last[i])
		pi.Extractor.L2MessageHash[i] = createNewLocalOpening(pi.LogHasher.HashFinal[i])
		pi.Extractor.DataChecksum[i] = createNewLocalOpening(pi.ExecMiMCHasher.HashFinal[i])

		comp.PublicInputs = append(comp.PublicInputs,
			wizard.PublicInput{
				Name: fmt.Sprintf("%s_%d", InitialStateRootHash, i),
				Acc:  accessors.NewLocalOpeningAccessor(pi.Extractor.InitialStateRootHash[i], 0),
			},
			wizard.PublicInput{
				Name: fmt.Sprintf("%s_%d", FinalStateRootHash, i),
				Acc:  accessors.NewLocalOpeningAccessor(pi.Extractor.FinalStateRootHash[i], 0),
			},
			wizard.PublicInput{
				Name: fmt.Sprintf("%s_%d", FirstRollingHashUpdate, i),
				Acc:  accessors.NewLocalOpeningAccessor(pi.Extractor.FirstRollingHashUpdate[i], 0),
			},
			wizard.PublicInput{
				Name: fmt.Sprintf("%s_%d", LastRollingHashUpdate, i),
				Acc:  accessors.NewLocalOpeningAccessor(pi.Extractor.LastRollingHashUpdate[i], 0),
			},
			wizard.PublicInput{
				Name: fmt.Sprintf("%s_%d", L2MessageHash, i),
				Acc:  accessors.NewLocalOpeningAccessor(pi.Extractor.L2MessageHash[i], 0),
			},
			wizard.PublicInput{
				Name: DataChecksum,
				Acc:  accessors.NewLocalOpeningAccessor(pi.Extractor.DataChecksum[i], 0),
			},
		)
	}

	for i := range pcommon.NbLimbEthAddress {
		pi.Extractor.L2MessageServiceAddr[i] = createNewLocalOpening(pi.Aux.LogSelectors.L2BridgeAddressCol[i])

		comp.PublicInputs = append(comp.PublicInputs,
			wizard.PublicInput{
				Name: fmt.Sprintf("%s_%d", L2MessageServiceAddr, i),
				Acc:  accessors.NewLocalOpeningAccessor(pi.Extractor.L2MessageServiceAddr[i], 0),
			},
		)
	}

	for i := range pcommon.NbLimbU128 {
		pi.Extractor.InitialBlockTimestamp[i] = createNewLocalOpening(pi.TimestampFetcher.First[i])
		pi.Extractor.FinalBlockTimestamp[i] = createNewLocalOpening(pi.TimestampFetcher.Last[i])
		pi.Extractor.FirstRollingHashUpdateNumber[i] = createNewLocalOpening(pi.RollingHashFetcher.FirstMessageNo[i])
		pi.Extractor.LastRollingHashUpdateNumber[i] = createNewLocalOpening(pi.RollingHashFetcher.LastMessageNo[i])
		pi.Extractor.ChainID[i] = createNewLocalOpening(pi.ChainID[i])

		comp.PublicInputs = append(comp.PublicInputs,
			wizard.PublicInput{
				Name: fmt.Sprintf("%s_%d", InitialBlockTimestamp, i),
				Acc:  accessors.NewLocalOpeningAccessor(pi.Extractor.InitialBlockTimestamp[i], 0),
			},
			wizard.PublicInput{
				Name: fmt.Sprintf("%s_%d", FinalBlockTimestamp, i),
				Acc:  accessors.NewLocalOpeningAccessor(pi.Extractor.FinalBlockTimestamp[i], 0),
			},
			wizard.PublicInput{
				Name: fmt.Sprintf("%s_%d", FirstRollingHashUpdateNumber, i),
				Acc:  accessors.NewLocalOpeningAccessor(pi.Extractor.FirstRollingHashUpdateNumber[i], 0),
			},
			wizard.PublicInput{
				Name: fmt.Sprintf("%s_%d", LastRollingHashNumberUpdate, i),
				Acc:  accessors.NewLocalOpeningAccessor(pi.Extractor.LastRollingHashUpdateNumber[i], 0),
			},
			wizard.PublicInput{
				Name: fmt.Sprintf("%s_%d", ChainID, i),
				Acc:  accessors.NewLocalOpeningAccessor(pi.Extractor.ChainID[i], 0)},
		)
	}

	for i := range pcommon.NbLimbU48 {
		pi.Extractor.InitialBlockNumber[i] = createNewLocalOpening(pi.TimestampFetcher.FirstBlockID[i])
		pi.Extractor.FinalBlockNumber[i] = createNewLocalOpening(pi.TimestampFetcher.LastBlockID[i])

		comp.PublicInputs = append(comp.PublicInputs,
			wizard.PublicInput{
				Name: fmt.Sprintf("%s_%d", InitialBlockNumber, i),
				Acc:  accessors.NewLocalOpeningAccessor(pi.Extractor.InitialBlockNumber[i], 0),
			},
			wizard.PublicInput{
				Name: fmt.Sprintf("%s_%d", FinalBlockNumber, i),
				Acc:  accessors.NewLocalOpeningAccessor(pi.Extractor.FinalBlockNumber[i], 0),
			},
		)
	}
}
