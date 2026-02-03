package mock

import (
	"math/big"
	"sort"

	eth "github.com/consensys/linea-monorepo/prover/circuits/pi-interconnection/keccak/prover/backend/execution/statemanager"
	"github.com/consensys/linea-monorepo/prover/circuits/pi-interconnection/keccak/prover/crypto/keccak"
	"github.com/consensys/linea-monorepo/prover/circuits/pi-interconnection/keccak/prover/maths/field"
	"github.com/consensys/linea-monorepo/prover/circuits/pi-interconnection/keccak/prover/utils"
	"github.com/consensys/linea-monorepo/prover/circuits/pi-interconnection/keccak/prover/utils/types"
)

// StateManagerVectors contains all the arithmetization columns needed for the state management at the account and storage levels
// (all these columns belong to the HUB module in the arithmetization)
// these slices will be at a later point be transformed into corresponding ifaces.Column() structs
type StateManagerVectors struct {
	// account data
	Address                                              []eth.Address // helper column
	AddressHI, AddressLO                                 []field.Element
	Nonce, NonceNew                                      []field.Element
	MimcCodeHash, MimcCodeHashNew                        []field.Element
	CodeHashHI, CodeHashLO, CodeHashHINew, CodeHashLONew []field.Element
	CodeSizeOld, CodeSizeNew                             []field.Element
	BalanceOld, BalanceNew                               []field.Element
	// storage data
	KeyHI, KeyLO                                       []field.Element
	ValueHICurr, ValueLOCurr, ValueHINext, ValueLONext []field.Element
	// helper numbers
	DeploymentNumber, DeploymentNumberInf []field.Element
	BlockNumber                           []field.Element
	// helper columns
	Exists, ExistsNew []field.Element
	PeekAtAccount     []field.Element
	PeekAtStorage     []field.Element
	// first and last marker columns
	FirstAOC, LastAOC []field.Element
	FirstKOC, LastKOC []field.Element
	// block markers
	FirstAOCBlock, LastAOCBlock            []field.Element
	FirstKOCBlock, LastKOCBlock            []field.Element
	MinDeploymentBlock, MaxDeploymentBlock []field.Element
}

func (smv *StateManagerVectors) Size() int {
	return len(smv.AddressHI)
}

// NewStateManagerVectors will initialize a StateManagerVectors struct, where the containing slices are initialized but are empty
func NewStateManagerVectors() *StateManagerVectors {
	return &StateManagerVectors{
		// account info
		Address:         []eth.Address{},
		AddressHI:       []field.Element{},
		AddressLO:       []field.Element{},
		Nonce:           []field.Element{},
		NonceNew:        []field.Element{},
		MimcCodeHash:    []field.Element{},
		MimcCodeHashNew: []field.Element{},
		CodeHashHI:      []field.Element{},
		CodeHashLO:      []field.Element{},
		CodeHashHINew:   []field.Element{},
		CodeHashLONew:   []field.Element{},
		CodeSizeOld:     []field.Element{},
		CodeSizeNew:     []field.Element{},
		BalanceOld:      []field.Element{},
		BalanceNew:      []field.Element{},
		// storage data
		KeyHI:       []field.Element{},
		KeyLO:       []field.Element{},
		ValueHICurr: []field.Element{},
		ValueLOCurr: []field.Element{},
		ValueHINext: []field.Element{},
		ValueLONext: []field.Element{},
		// helper numbers
		DeploymentNumber:    []field.Element{},
		DeploymentNumberInf: []field.Element{},
		BlockNumber:         []field.Element{},
		// helper columns
		Exists:        []field.Element{},
		ExistsNew:     []field.Element{},
		PeekAtAccount: []field.Element{},
		PeekAtStorage: []field.Element{},
		// first and last marker columns
		FirstAOC: []field.Element{},
		LastAOC:  []field.Element{},
		FirstKOC: []field.Element{},
		LastKOC:  []field.Element{},
		// block markers
		FirstAOCBlock:      []field.Element{},
		LastAOCBlock:       []field.Element{},
		FirstKOCBlock:      []field.Element{},
		LastKOCBlock:       []field.Element{},
		MinDeploymentBlock: []field.Element{},
		MaxDeploymentBlock: []field.Element{},
	}
}

// AccountHistory contains the segments that describe the evolution of the Account
// the data is kept as field elements, and will be stitched together
// to form the columns of a StateManagerAccountColumns struct
type AccountHistory struct {
	address                                              []eth.Address   // helper column
	addressHI, addressLO                                 []field.Element // these two vectors will contain the same element
	nonce, nonceNew                                      []field.Element
	mimcCodeHash, mimcCodeHashNew                        []field.Element
	codeHashHI, codeHashLO, codeHashHINew, codeHashLONew []field.Element
	codeSizeOld, codeSizeNew                             []field.Element
	balanceOld, balanceNew                               []field.Element
	deploymentNumber, deploymentNumberInf                []field.Element
	blockNumber                                          []field.Element
	currentDeploymentNo                                  int
	exists, existsNew                                    []field.Element
	currentExistence                                     int // it is a bool but we model it as int for ease of field conversions
	storageData                                          map[types.FullBytes32]*StorageHistory
	minDeplBlock                                         map[int]int
	maxDeplBlock                                         map[int]int
}

// StorageHistory contains column segments that describe the storage evolution for a specific key
// the KeyHI, KeyLO values are always the same and will be used for stitching
type StorageHistory struct {
	address                               []eth.Address
	addressHI, addressLO                  []field.Element // these two vectors will contain the same element
	keyHI, keyLO                          []field.Element // this column will contain the same value of the key everywhere
	valueHICurr, valueLOCurr              []field.Element
	valueHINext, valueLONext              []field.Element
	deploymentNumber, deploymentNumberInf []field.Element
	blockNumber                           []field.Element
	exists, existsNew                     []field.Element
}

const (
	GENERATE_ACP_SAMPLE int = 0
	GENERATE_SCP_SAMPLE int = 1
)

func SampleTypeToString(sampleType int) string {
	switch sampleType {
	case GENERATE_ACP_SAMPLE:
		return "ACP"
	case GENERATE_SCP_SAMPLE:
		return "SCP"
	default:
		panic("Wrong arithmetization sample type used")
	}

}

// The Stitcher constructs the vector slices of field.Element that will correspond to the HUB columns
// it will store every AccountHistory which appears in the columns, and manages that using a map
type Stitcher struct {
	accountData    map[eth.Address]*AccountHistory
	accountTouched map[eth.Address]struct{}
	initialState   *State
	currentBlock   int
}

// NewAccountHistory initializes a new account history where each column Exists but is empty
func NewAccountHistory() *AccountHistory {
	return &AccountHistory{
		address:             []eth.Address{},
		addressHI:           []field.Element{},
		addressLO:           []field.Element{},
		nonce:               []field.Element{},
		nonceNew:            []field.Element{},
		mimcCodeHash:        []field.Element{},
		mimcCodeHashNew:     []field.Element{},
		codeHashHI:          []field.Element{},
		codeHashLO:          []field.Element{},
		codeHashHINew:       []field.Element{},
		codeHashLONew:       []field.Element{},
		codeSizeOld:         []field.Element{},
		codeSizeNew:         []field.Element{},
		balanceOld:          []field.Element{},
		balanceNew:          []field.Element{},
		deploymentNumber:    []field.Element{},
		deploymentNumberInf: []field.Element{},
		blockNumber:         []field.Element{},
		exists:              []field.Element{},
		existsNew:           []field.Element{},
		currentDeploymentNo: 0,
		currentExistence:    0,
		minDeplBlock:        map[int]int{},
		maxDeplBlock:        map[int]int{},
		storageData:         map[types.FullBytes32]*StorageHistory{},
	}
}

// NewStorageHistory contains the history of storage accesses, where each row corresponds to a storage access
func NewStorageHistory() *StorageHistory {
	return &StorageHistory{
		address:             []eth.Address{},
		addressHI:           []field.Element{},
		addressLO:           []field.Element{},
		keyHI:               []field.Element{},
		keyLO:               []field.Element{},
		valueHICurr:         []field.Element{},
		valueLOCurr:         []field.Element{},
		valueHINext:         []field.Element{},
		valueLONext:         []field.Element{},
		deploymentNumber:    []field.Element{},
		deploymentNumberInf: []field.Element{},
		blockNumber:         []field.Element{},
		exists:              []field.Element{},
		existsNew:           []field.Element{},
	}
}

// InitializeNewRow creates a new cell in each column, which is initialized either with special values if this is the start of the AccountHistory
// otherwise the current row is initialized with the values of the previous row
// this simplified the process of adding a row, as most information is duplicated (for example during reads, or when writing only modifies a single value and the others remain unchanged)
// so we will duplicate and only make the changes for the few cells that must be changed
func (accHistory *AccountHistory) InitializeNewRow(currentBlock int, initialState *State, address eth.Address) *AccountHistory {
	var prevAddress eth.Address
	var prevAddressHI, prevAddressLO, prevNonce, prevMimcCodeHash, prevCodeHashHI, prevCodeHashLO, prevCodeSize, prevBalance, prevDeploymentNumber field.Element

	if len(accHistory.addressLO) > 0 {
		// store the information from the previous row in the following values
		// we pick the information in [...]New columns, as these have the most up-to-date information
		lastIndex := len(accHistory.addressLO) - 1
		prevAddress = accHistory.address[lastIndex]
		prevAddressHI = accHistory.addressHI[lastIndex]
		prevAddressLO = accHistory.addressLO[lastIndex]
		prevNonce = accHistory.nonceNew[lastIndex]
		prevMimcCodeHash = accHistory.mimcCodeHashNew[lastIndex]
		prevCodeHashHI = accHistory.codeHashHINew[lastIndex]
		prevCodeHashLO = accHistory.codeHashLONew[lastIndex]
		prevCodeSize = accHistory.codeSizeNew[lastIndex]
		prevBalance = accHistory.balanceNew[lastIndex]
		prevDeploymentNumber = accHistory.deploymentNumber[lastIndex]
		// prevBlockNumber = (*accHistory.BlockNumber)[lastIndex] // use previous block number
	} else {
		// search whether the previous values can be found in the initial state
		_, isPresent := (*initialState)[address]
		if !isPresent {
			// when the previous values do not exist yet in the initial state, initialize them with 0s
			prevAddress = address
			prevAddressHI.SetBytes(address[:4])
			prevAddressLO.SetBytes(address[4:])
			prevNonce.SetZero()
			prevMimcCodeHash.SetZero()
			codeHashEmpty := keccak.Hash([]byte{})
			prevCodeHashHI.SetBytes(codeHashEmpty[:16])
			prevCodeHashLO.SetBytes(codeHashEmpty[16:])
			prevCodeSize.SetZero()
			prevBalance.SetZero()
			prevDeploymentNumber.SetZero() // when the state is initialized, we start with DeploymentNumber = 0 MISTAKE HERE
			// prevBlockNumber.SetInt64(int64(currentBlock))
		} else {
			// initialize them with values from the state
			accHistory.minDeplBlock[currentBlock] = 0 // minimum deployment number per block is 0 when we initialize from state
			accHistory.currentExistence = 1           // when we initialize from state we will assume that the account Exists
			// account information
			prevAddress = address
			prevAddressHI.SetBytes(address[:4])
			prevAddressLO.SetBytes(address[4:20])
			prevNonce.SetInt64(initialState.GetNonce(address))
			// hashes
			stateMimcHash := initialState.GetMimcCodeHash(address)
			prevMimcCodeHash.SetBytes(stateMimcHash[:])
			stateCodeHash := initialState.GetCodeHash(address)
			prevCodeHashHI.SetBytes(stateCodeHash[:16])
			prevCodeHashLO.SetBytes(stateCodeHash[16:])
			prevCodeSize.SetInt64(initialState.GetCodeSize(address))

			prevBalance.SetBigInt(initialState.GetBalance(address))
			prevDeploymentNumber.SetZero() // correct here, initialized from state
		}

	}
	// account-related information
	accHistory.address = append(accHistory.address, prevAddress)
	accHistory.addressHI = append(accHistory.addressHI, prevAddressHI)
	accHistory.addressLO = append(accHistory.addressLO, prevAddressLO)
	accHistory.nonce = append(accHistory.nonce, prevNonce)
	accHistory.nonceNew = append(accHistory.nonceNew, prevNonce)
	accHistory.mimcCodeHash = append(accHistory.mimcCodeHash, prevMimcCodeHash)
	accHistory.mimcCodeHashNew = append(accHistory.mimcCodeHashNew, prevMimcCodeHash)
	accHistory.codeHashHI = append(accHistory.codeHashHI, prevCodeHashHI)
	accHistory.codeHashLO = append(accHistory.codeHashLO, prevCodeHashLO)
	accHistory.codeHashHINew = append(accHistory.codeHashHINew, prevCodeHashHI)
	accHistory.codeHashLONew = append(accHistory.codeHashLONew, prevCodeHashLO)
	accHistory.codeSizeOld = append(accHistory.codeSizeOld, prevCodeSize)
	accHistory.codeSizeNew = append(accHistory.codeSizeNew, prevCodeSize)
	accHistory.balanceOld = append(accHistory.balanceOld, prevBalance)
	accHistory.balanceNew = append(accHistory.balanceNew, prevBalance)

	// deployment number
	accHistory.deploymentNumber = append(accHistory.deploymentNumber, prevDeploymentNumber)

	// block number
	var prevBlockNumber field.Element
	accHistory.blockNumber = append(accHistory.blockNumber, *prevBlockNumber.SetInt64(int64(currentBlock)))

	// exist and existNew columns
	var existsElem, existsNewElem field.Element
	accHistory.exists = append(accHistory.exists, *existsElem.SetInt64(int64(accHistory.currentExistence)))
	accHistory.existsNew = append(accHistory.existsNew, *existsNewElem.SetInt64(int64(accHistory.currentExistence)))

	return accHistory
}

func (stitcher *Stitcher) Initialize(currentBlock int, state State) *Stitcher {
	stitcher.accountData = map[eth.Address]*AccountHistory{}
	stitcher.accountTouched = map[eth.Address]struct{}{}
	stateCopy := state.DeepCopy()
	stitcher.initialState = &stateCopy
	stitcher.currentBlock = currentBlock
	return stitcher
}

// PrependDummyVector is used inside PadToNearestPowerOf2 to prepend the slices with zeroes
// in order for their length to be a power of two
func PrependDummyVector[T any](vector []T, fullLength int, currentLength int) []T {
	dummyVector := make([]T, fullLength-currentLength)
	dummyVector = append(dummyVector, vector...)
	return dummyVector
}

// PadToNearestPowerOf2 will pad all the vectors with zeroes so that their length becomes a power of two and they can be transformed into iColumn structs
func (stateManagerVectors *StateManagerVectors) PadToNearestPowerOf2() {
	// first compute the necessary length for padding
	fullLength := utils.NextPowerOfTwo(len(stateManagerVectors.Address))
	currentLength := len(stateManagerVectors.Address)

	// pad account columns
	stateManagerVectors.Address = PrependDummyVector(stateManagerVectors.Address, fullLength, currentLength)
	stateManagerVectors.AddressHI = PrependDummyVector(stateManagerVectors.AddressHI, fullLength, currentLength)
	stateManagerVectors.AddressLO = PrependDummyVector(stateManagerVectors.AddressLO, fullLength, currentLength)
	stateManagerVectors.Nonce = PrependDummyVector(stateManagerVectors.Nonce, fullLength, currentLength)
	stateManagerVectors.NonceNew = PrependDummyVector(stateManagerVectors.NonceNew, fullLength, currentLength)
	stateManagerVectors.MimcCodeHash = PrependDummyVector(stateManagerVectors.MimcCodeHash, fullLength, currentLength)
	stateManagerVectors.MimcCodeHashNew = PrependDummyVector(stateManagerVectors.MimcCodeHashNew, fullLength, currentLength)
	stateManagerVectors.CodeHashHI = PrependDummyVector(stateManagerVectors.CodeHashHI, fullLength, currentLength)
	stateManagerVectors.CodeHashLO = PrependDummyVector(stateManagerVectors.CodeHashLO, fullLength, currentLength)
	stateManagerVectors.CodeHashHINew = PrependDummyVector(stateManagerVectors.CodeHashHINew, fullLength, currentLength)
	stateManagerVectors.CodeHashLONew = PrependDummyVector(stateManagerVectors.CodeHashLONew, fullLength, currentLength)
	stateManagerVectors.CodeSizeOld = PrependDummyVector(stateManagerVectors.CodeSizeOld, fullLength, currentLength)
	stateManagerVectors.CodeSizeNew = PrependDummyVector(stateManagerVectors.CodeSizeNew, fullLength, currentLength)
	stateManagerVectors.BalanceOld = PrependDummyVector(stateManagerVectors.BalanceOld, fullLength, currentLength)
	stateManagerVectors.BalanceNew = PrependDummyVector(stateManagerVectors.BalanceNew, fullLength, currentLength)

	stateManagerVectors.DeploymentNumber = PrependDummyVector(stateManagerVectors.DeploymentNumber, fullLength, currentLength)
	stateManagerVectors.BlockNumber = PrependDummyVector(stateManagerVectors.BlockNumber, fullLength, currentLength)
	stateManagerVectors.Exists = PrependDummyVector(stateManagerVectors.Exists, fullLength, currentLength)
	stateManagerVectors.ExistsNew = PrependDummyVector(stateManagerVectors.ExistsNew, fullLength, currentLength)

	// pad storage columns
	stateManagerVectors.KeyHI = PrependDummyVector(stateManagerVectors.KeyHI, fullLength, currentLength)
	stateManagerVectors.KeyLO = PrependDummyVector(stateManagerVectors.KeyLO, fullLength, currentLength)
	stateManagerVectors.ValueHICurr = PrependDummyVector(stateManagerVectors.ValueHICurr, fullLength, currentLength)
	stateManagerVectors.ValueLOCurr = PrependDummyVector(stateManagerVectors.ValueLOCurr, fullLength, currentLength)
	stateManagerVectors.ValueHINext = PrependDummyVector(stateManagerVectors.ValueHINext, fullLength, currentLength)
	stateManagerVectors.ValueLONext = PrependDummyVector(stateManagerVectors.ValueLONext, fullLength, currentLength)

	// pad PeekAtStorage
	stateManagerVectors.PeekAtStorage = PrependDummyVector(stateManagerVectors.PeekAtStorage, fullLength, currentLength)

	// pad PeekAtAccount
	stateManagerVectors.PeekAtAccount = PrependDummyVector(stateManagerVectors.PeekAtAccount, fullLength, currentLength)

	// pad first AOC and last AOC
	stateManagerVectors.FirstAOC = PrependDummyVector(stateManagerVectors.FirstAOC, fullLength, currentLength)
	stateManagerVectors.LastAOC = PrependDummyVector(stateManagerVectors.LastAOC, fullLength, currentLength)

	// pad first KOC and last KOC
	stateManagerVectors.FirstKOC = PrependDummyVector(stateManagerVectors.FirstKOC, fullLength, currentLength)
	stateManagerVectors.LastKOC = PrependDummyVector(stateManagerVectors.LastKOC, fullLength, currentLength)

	// pad deployment Number Infinity
	stateManagerVectors.DeploymentNumberInf = PrependDummyVector(stateManagerVectors.DeploymentNumberInf, fullLength, currentLength)

	// pad block markers
	stateManagerVectors.FirstAOCBlock = PrependDummyVector(stateManagerVectors.FirstAOCBlock, fullLength, currentLength)
	stateManagerVectors.LastAOCBlock = PrependDummyVector(stateManagerVectors.LastAOCBlock, fullLength, currentLength)
	stateManagerVectors.FirstKOCBlock = PrependDummyVector(stateManagerVectors.FirstKOCBlock, fullLength, currentLength)
	stateManagerVectors.LastKOCBlock = PrependDummyVector(stateManagerVectors.LastKOCBlock, fullLength, currentLength)
	stateManagerVectors.MinDeploymentBlock = PrependDummyVector(stateManagerVectors.MinDeploymentBlock, fullLength, currentLength)
	stateManagerVectors.MaxDeploymentBlock = PrependDummyVector(stateManagerVectors.MaxDeploymentBlock, fullLength, currentLength)
}

// Finalize uses the information in the AccountHistory structs that are stored inside the Stitcher and concatenates these vectors
// AccountHistory contains multiple StorageHistory, one for each storage key that ever gets accessed.
// Finalize also adds the information in StorageHistory to the concatenated vectors.
// The final result is padded so that the length of the columns is a power of 2.
func (stitcher *Stitcher) Finalize(sampleType int) *StateManagerVectors {
	sortedAccountAddresses := MapKeysToSlice(stitcher.accountData)
	sort.Slice(sortedAccountAddresses, func(i, j int) bool {
		s1 := sortedAccountAddresses[i].Hex()
		s2 := sortedAccountAddresses[j].Hex()
		return s1 < s2
	})

	stateManagerVectors := NewStateManagerVectors()

	// stitch account columns
	StitchAccountColumns := func() {
		for _, accountAddress := range sortedAccountAddresses {
			_, accountTouched := stitcher.accountTouched[accountAddress]
			if accountTouched {
				accHist := stitcher.accountData[accountAddress]
				stateManagerVectors.Address = append(stateManagerVectors.Address, accHist.address...)
				stateManagerVectors.AddressHI = append(stateManagerVectors.AddressHI, accHist.addressHI...)
				stateManagerVectors.AddressLO = append(stateManagerVectors.AddressLO, accHist.addressLO...)
				stateManagerVectors.Nonce = append(stateManagerVectors.Nonce, accHist.nonce...)
				stateManagerVectors.NonceNew = append(stateManagerVectors.NonceNew, accHist.nonceNew...)
				stateManagerVectors.MimcCodeHash = append(stateManagerVectors.MimcCodeHash, accHist.mimcCodeHash...)
				stateManagerVectors.MimcCodeHashNew = append(stateManagerVectors.MimcCodeHashNew, accHist.mimcCodeHashNew...)
				stateManagerVectors.CodeHashHI = append(stateManagerVectors.CodeHashHI, accHist.codeHashHI...)
				stateManagerVectors.CodeHashLO = append(stateManagerVectors.CodeHashLO, accHist.codeHashLO...)
				stateManagerVectors.CodeHashHINew = append(stateManagerVectors.CodeHashHINew, accHist.codeHashHINew...)
				stateManagerVectors.CodeHashLONew = append(stateManagerVectors.CodeHashLONew, accHist.codeHashLONew...)
				stateManagerVectors.CodeSizeOld = append(stateManagerVectors.CodeSizeOld, accHist.codeSizeOld...)
				stateManagerVectors.CodeSizeNew = append(stateManagerVectors.CodeSizeNew, accHist.codeSizeNew...)
				stateManagerVectors.BalanceOld = append(stateManagerVectors.BalanceOld, accHist.balanceOld...)
				stateManagerVectors.BalanceNew = append(stateManagerVectors.BalanceNew, accHist.balanceNew...)
				stateManagerVectors.DeploymentNumber = append(stateManagerVectors.DeploymentNumber, accHist.deploymentNumber...)
				stateManagerVectors.BlockNumber = append(stateManagerVectors.BlockNumber, accHist.blockNumber...)
				stateManagerVectors.Exists = append(stateManagerVectors.Exists, accHist.exists...)
				stateManagerVectors.ExistsNew = append(stateManagerVectors.ExistsNew, accHist.existsNew...)

				// add dummy rows
				dummyVector := make([]field.Element, len(accHist.addressHI))
				stateManagerVectors.KeyHI = append(stateManagerVectors.KeyHI, dummyVector...)
				stateManagerVectors.KeyLO = append(stateManagerVectors.KeyLO, dummyVector...)
				stateManagerVectors.ValueHICurr = append(stateManagerVectors.ValueHICurr, dummyVector...)
				stateManagerVectors.ValueLOCurr = append(stateManagerVectors.ValueLOCurr, dummyVector...)
				stateManagerVectors.ValueHINext = append(stateManagerVectors.ValueHINext, dummyVector...)
				stateManagerVectors.ValueLONext = append(stateManagerVectors.ValueLONext, dummyVector...)

				// add dummy helper column, all zeros
				stateManagerVectors.PeekAtStorage = append(stateManagerVectors.PeekAtStorage, dummyVector...)

				// initialize the peek at account column and then add it
				peekAtAccountFragment := make([]field.Element, len(accHist.addressHI))
				for index := range peekAtAccountFragment {
					peekAtAccountFragment[index].SetOne()
				}
				stateManagerVectors.PeekAtAccount = append(stateManagerVectors.PeekAtAccount, peekAtAccountFragment...)

				// first AOC and last AOC
				firstAOCFragment := make([]field.Element, len(accHist.addressHI))
				firstAOCFragment[0].SetOne()
				lastAOCFragment := make([]field.Element, len(accHist.addressHI))
				lastAOCFragment[len(lastAOCFragment)-1].SetOne()
				stateManagerVectors.FirstAOC = append(stateManagerVectors.FirstAOC, firstAOCFragment...)
				stateManagerVectors.LastAOC = append(stateManagerVectors.LastAOC, lastAOCFragment...)

				// first KOC and last KOC
				stateManagerVectors.FirstKOC = append(stateManagerVectors.FirstKOC, dummyVector...)
				stateManagerVectors.LastKOC = append(stateManagerVectors.LastKOC, dummyVector...)

				// deployment Number Infinity
				deploymentInfFragment := make([]field.Element, len(accHist.addressHI))
				for index := range deploymentInfFragment {
					deploymentInfFragment[index].SetInt64(int64(accHist.currentDeploymentNo))
				}
				stateManagerVectors.DeploymentNumberInf = append(stateManagerVectors.DeploymentNumberInf, deploymentInfFragment...)
				// FirstAOCBlock
				firstAOCBlockFragment := make([]field.Element, len(accHist.addressHI))
				firstAOCBlockFragment[0].SetOne()
				for index := 1; index < len(firstAOCBlockFragment); index++ {
					if !accHist.blockNumber[index].Equal(&accHist.blockNumber[index-1]) {
						firstAOCBlockFragment[index].SetOne()
					}
				}
				stateManagerVectors.FirstAOCBlock = append(stateManagerVectors.FirstAOCBlock, firstAOCBlockFragment...)
				// LastAOCBlock
				lastAOCBlockFragment := make([]field.Element, len(accHist.addressHI))
				lastAOCBlockFragment[len(lastAOCBlockFragment)-1].SetOne()
				for index := 1; index < len(lastAOCBlockFragment); index++ {
					if !accHist.blockNumber[index].Equal(&accHist.blockNumber[index-1]) {
						lastAOCBlockFragment[index-1].SetOne()
					}
				}
				stateManagerVectors.LastAOCBlock = append(stateManagerVectors.LastAOCBlock, lastAOCBlockFragment...)
				// FirstKOCBlock, LastKOCBlock and MinDeploymentBlock and MaxDeploymentBlock, all zeros.
				stateManagerVectors.FirstKOCBlock = append(stateManagerVectors.FirstKOCBlock, dummyVector...)
				stateManagerVectors.LastKOCBlock = append(stateManagerVectors.LastKOCBlock, dummyVector...)
				stateManagerVectors.MinDeploymentBlock = append(stateManagerVectors.MinDeploymentBlock, dummyVector...)
				stateManagerVectors.MaxDeploymentBlock = append(stateManagerVectors.MaxDeploymentBlock, dummyVector...)
			}
		}
	}

	// Stitch the segments corresponding to storage keys
	StitchStorageColumns := func() {
		for _, accountAddress := range sortedAccountAddresses {
			_, accountTouched := stitcher.accountTouched[accountAddress]
			if accountTouched {
				accHist := stitcher.accountData[accountAddress]
				storageHistMap := accHist.storageData
				sortedStorageKeys := MapKeysToSlice(storageHistMap)
				sort.Slice(sortedStorageKeys, func(i, j int) bool {
					s1 := sortedStorageKeys[i].Hex()
					s2 := sortedStorageKeys[j].Hex()
					return s1 < s2
				})
				for _, key := range sortedStorageKeys {
					stoHist := storageHistMap[key]
					stateManagerVectors.Address = append(stateManagerVectors.Address, stoHist.address...)
					stateManagerVectors.AddressHI = append(stateManagerVectors.AddressHI, stoHist.addressHI...)
					stateManagerVectors.AddressLO = append(stateManagerVectors.AddressLO, stoHist.addressLO...)

					stateManagerVectors.DeploymentNumber = append(stateManagerVectors.DeploymentNumber, stoHist.deploymentNumber...)
					stateManagerVectors.BlockNumber = append(stateManagerVectors.BlockNumber, stoHist.blockNumber...)
					stateManagerVectors.Exists = append(stateManagerVectors.Exists, stoHist.exists...)
					stateManagerVectors.ExistsNew = append(stateManagerVectors.ExistsNew, stoHist.existsNew...)
					// add keys and values
					stateManagerVectors.KeyHI = append(stateManagerVectors.KeyHI, stoHist.keyHI...)
					stateManagerVectors.KeyLO = append(stateManagerVectors.KeyLO, stoHist.keyLO...)
					stateManagerVectors.ValueHICurr = append(stateManagerVectors.ValueHICurr, stoHist.valueHICurr...)
					stateManagerVectors.ValueLOCurr = append(stateManagerVectors.ValueLOCurr, stoHist.valueLOCurr...)
					stateManagerVectors.ValueHINext = append(stateManagerVectors.ValueHINext, stoHist.valueHINext...)
					stateManagerVectors.ValueLONext = append(stateManagerVectors.ValueLONext, stoHist.valueLONext...)
					// add dummy rows
					dummyVector := make([]field.Element, len(stoHist.addressHI))
					stateManagerVectors.Nonce = append(stateManagerVectors.Nonce, dummyVector...)
					stateManagerVectors.NonceNew = append(stateManagerVectors.NonceNew, dummyVector...)

					stateManagerVectors.MimcCodeHash = append(stateManagerVectors.MimcCodeHash, dummyVector...)
					stateManagerVectors.MimcCodeHashNew = append(stateManagerVectors.MimcCodeHashNew, dummyVector...)

					stateManagerVectors.CodeHashHI = append(stateManagerVectors.CodeHashHI, dummyVector...)
					stateManagerVectors.CodeHashLO = append(stateManagerVectors.CodeHashLO, dummyVector...)

					stateManagerVectors.CodeHashHINew = append(stateManagerVectors.CodeHashHINew, dummyVector...)
					stateManagerVectors.CodeHashLONew = append(stateManagerVectors.CodeHashLONew, dummyVector...)

					stateManagerVectors.CodeSizeOld = append(stateManagerVectors.CodeSizeOld, dummyVector...)
					stateManagerVectors.CodeSizeNew = append(stateManagerVectors.CodeSizeNew, dummyVector...)

					stateManagerVectors.BalanceOld = append(stateManagerVectors.BalanceOld, dummyVector...)
					stateManagerVectors.BalanceNew = append(stateManagerVectors.BalanceNew, dummyVector...)

					// add dummy helper column, all zeros
					stateManagerVectors.PeekAtAccount = append(stateManagerVectors.PeekAtAccount, dummyVector...)

					// initialize the peek at storage column segment and then add it
					peekAtStorageFragment := make([]field.Element, len(stoHist.addressHI))
					for index := range peekAtStorageFragment {
						peekAtStorageFragment[index].SetOne()
					}
					stateManagerVectors.PeekAtStorage = append(stateManagerVectors.PeekAtStorage, peekAtStorageFragment...)

					// FirstAOC and LastAOC columns
					stateManagerVectors.FirstAOC = append(stateManagerVectors.FirstAOC, dummyVector...)
					stateManagerVectors.LastAOC = append(stateManagerVectors.LastAOC, dummyVector...)

					// first KOC and last KOC initializing and adding
					firstKOCFragment := make([]field.Element, len(stoHist.addressHI))
					firstKOCFragment[0].SetOne()
					lastKOCFragment := make([]field.Element, len(stoHist.addressHI))
					lastKOCFragment[len(lastKOCFragment)-1].SetOne()
					stateManagerVectors.FirstKOC = append(stateManagerVectors.FirstKOC, firstKOCFragment...)
					stateManagerVectors.LastKOC = append(stateManagerVectors.LastKOC, lastKOCFragment...)

					// deployment Number Infinity
					deploymentInfFragment := make([]field.Element, len(stoHist.addressHI))
					for index := range deploymentInfFragment {
						deploymentInfFragment[index].SetInt64(int64(accHist.currentDeploymentNo))
					}
					stateManagerVectors.DeploymentNumberInf = append(stateManagerVectors.DeploymentNumberInf, deploymentInfFragment...)

					// FirstKOCBlock
					firstKOCBlockFragment := make([]field.Element, len(stoHist.addressHI))
					firstKOCBlockFragment[0].SetOne()
					for index := 1; index < len(firstKOCBlockFragment); index++ {
						if !stoHist.blockNumber[index].Equal(&stoHist.blockNumber[index-1]) {
							firstKOCBlockFragment[index].SetOne()
						}
					}
					stateManagerVectors.FirstKOCBlock = append(stateManagerVectors.FirstKOCBlock, firstKOCBlockFragment...)
					// LastAOCBlock
					lastKOCBlockFragment := make([]field.Element, len(stoHist.addressHI))
					lastKOCBlockFragment[len(lastKOCBlockFragment)-1].SetOne()
					for index := 1; index < len(lastKOCBlockFragment); index++ {
						if !stoHist.blockNumber[index].Equal(&stoHist.blockNumber[index-1]) {
							lastKOCBlockFragment[index-1].SetOne()
						}
					}
					stateManagerVectors.LastKOCBlock = append(stateManagerVectors.LastKOCBlock, lastKOCBlockFragment...)
					// FirstAOCBlock, LastAOCBlock and MinDeploymentBlock and MaxDeploymentBlock, all zeros.
					stateManagerVectors.FirstAOCBlock = append(stateManagerVectors.FirstAOCBlock, dummyVector...)
					stateManagerVectors.LastAOCBlock = append(stateManagerVectors.LastAOCBlock, dummyVector...)
					minDeplBlockFragment := make([]field.Element, len(stoHist.addressHI))
					for index := range minDeplBlockFragment {
						minDeplBlock := accHist.minDeplBlock[field.ToInt(&stoHist.blockNumber[index])]
						minDeplBlockFragment[index].SetInt64(int64(minDeplBlock))
					}
					stateManagerVectors.MinDeploymentBlock = append(stateManagerVectors.MinDeploymentBlock, minDeplBlockFragment...)
					maxDeplBlockFragment := make([]field.Element, len(stoHist.addressHI))
					for index := range maxDeplBlockFragment {
						maxDeplBlockFragment[index].SetInt64(int64(accHist.maxDeplBlock[field.ToInt(&stoHist.blockNumber[index])]))
					}
					stateManagerVectors.MaxDeploymentBlock = append(stateManagerVectors.MaxDeploymentBlock, maxDeplBlockFragment...)
				}
			}
		}
	}

	switch sampleType {
	case GENERATE_ACP_SAMPLE:
		StitchStorageColumns()
		StitchAccountColumns()
	case GENERATE_SCP_SAMPLE:
		StitchAccountColumns()
		StitchStorageColumns()
	}

	stateManagerVectors.PadToNearestPowerOf2()
	return stateManagerVectors
}

// AddFrame uses the information inside a frame to add one more row in AccountHistory
// by appending one entry to each slice contained in AccountHistory
func (accHistory *AccountHistory) AddFrame(frame StateAccessLog, initialState *State, address eth.Address) *AccountHistory {
	accHistory.InitializeNewRow(frame.Block, initialState, address) // this sets the block number as well
	lastIndex := len(accHistory.addressHI) - 1
	accHistory.address[lastIndex] = frame.Address
	accHistory.addressHI[lastIndex].SetBytes(frame.Address[:4])
	accHistory.addressLO[lastIndex].SetBytes(frame.Address[4:])

	/*
		the following line writes maxDeplBlock even when the account might not exist
		the storage keys on those rows must still be checked
	*/
	accHistory.maxDeplBlock[frame.Block] = accHistory.currentDeploymentNo
	_, exists := accHistory.minDeplBlock[frame.Block]
	if !exists {
		// only update minDeplBlock if
		accHistory.minDeplBlock[frame.Block] = accHistory.currentDeploymentNo
	}

	switch frame.Type {
	case Storage:
		storageMap := accHistory.storageData
		stoHist, storageExists := storageMap[frame.Key]
		if !storageExists {
			stoHist = NewStorageHistory()
			storageMap[frame.Key] = stoHist
		}
		stoHist.AddFrame(frame, accHistory.currentDeploymentNo, accHistory.currentExistence)
	case Nonce:
		accHistory.nonceNew[lastIndex].SetInt64((frame.Value).(int64))
	case Balance:
		accHistory.balanceNew[lastIndex].SetBigInt((frame.Value).(*big.Int))
	case Codesize: // only for reads
		accHistory.codeSizeNew[lastIndex].SetInt64((frame.Value).(int64))
	case Codehash:
		// this should also update the MIMC code hash, but the frame does not contain this value
		codeHash := frame.Value.(types.FullBytes32)
		accHistory.codeHashHINew[lastIndex].SetBytes(codeHash[:16])
		accHistory.codeHashLONew[lastIndex].SetBytes(codeHash[16:])
	case AccountErasal:
		// we do not delete the storage history, as it remains persistent over multiple deployments
		if accHistory.currentExistence == 0 {
			panic("Attempted to delete an account which does not exist.")
		}
		accHistory.currentExistence = 0
		accHistory.exists[lastIndex] = field.One()     // it Exists when delete is invoked
		accHistory.existsNew[lastIndex] = field.Zero() // it does not exist afterward
		accHistory.nonceNew[lastIndex].SetZero()
		accHistory.balanceNew[lastIndex].SetZero()
		accHistory.codeSizeNew[lastIndex].SetZero()
		codeHashEmpty := keccak.Hash([]byte{})
		accHistory.codeHashHINew[lastIndex].SetBytes(codeHashEmpty[:16])
		accHistory.codeHashLONew[lastIndex].SetBytes(codeHashEmpty[16:])
	case AccountInit:
		if accHistory.currentExistence == 1 {
			panic("Attempted to initialize an account which already Exists.")
		}
		accHistory.currentDeploymentNo++
		accHistory.currentExistence = 1
		accHistory.deploymentNumber[lastIndex].SetUint64(uint64(accHistory.currentDeploymentNo))
		accHistory.exists[lastIndex] = field.Zero()   // it does not exist when create is invoked
		accHistory.existsNew[lastIndex] = field.One() // it Exists afterwards
		frameValues := frame.Value.([]any)
		accHistory.codeSizeNew[lastIndex].SetUint64(uint64(frameValues[0].(int64)))

		// previous codeHash is emptu before init
		codeHashEmpty := keccak.Hash([]byte{})
		accHistory.codeHashHI[lastIndex].SetBytes(codeHashEmpty[:16])
		accHistory.codeHashLO[lastIndex].SetBytes(codeHashEmpty[16:])

		// initialize new codehash with the corresponding values inside the account
		codeHashBytes := frameValues[1].(types.FullBytes32)
		accHistory.codeHashHINew[lastIndex].SetBytes(codeHashBytes[:16])
		accHistory.codeHashLONew[lastIndex].SetBytes(codeHashBytes[16:])
		accHistory.nonceNew[lastIndex].SetZero()
		accHistory.balanceNew[lastIndex].SetZero()

		_, exists := accHistory.minDeplBlock[frame.Block]
		if !exists {
			accHistory.minDeplBlock[frame.Block] = accHistory.currentDeploymentNo
		}
		accHistory.maxDeplBlock[frame.Block] = accHistory.currentDeploymentNo
	}
	return accHistory
}

// AddFrame uses the information inside a frame to create a new row in the StorageHistory, by adding one entry to each column
// currentExistence is derived from the accountHistory and is used to compute the Exists and ExistsNew columns
// currentDeploymentNumber is derived from the accountHistory and will be used to compute the deployment number
func (stoHistory *StorageHistory) AddFrame(frame StateAccessLog, currentDeploymentNo int, currentExistence int) *StorageHistory {

	stoHistory.address = append(stoHistory.address, frame.Address)
	stoHistory.addressHI = append(stoHistory.addressHI, *(&field.Element{}).SetBytes(frame.Address[:4]))
	stoHistory.addressLO = append(stoHistory.addressLO, *(&field.Element{}).SetBytes(frame.Address[4:]))

	stoHistory.keyHI = append(stoHistory.keyHI, *(&field.Element{}).SetBytes(frame.Key[:16]))
	stoHistory.keyLO = append(stoHistory.keyLO, *(&field.Element{}).SetBytes(frame.Key[16:]))

	var oldValue types.FullBytes32
	if frame.IsWrite {
		oldValue = frame.OldValue.(types.FullBytes32)
	} else {
		// it is a Read operation, therefore we can just populate the old value with the new value
		oldValue = frame.Value.(types.FullBytes32)
	}

	stoHistory.valueHICurr = append(stoHistory.valueHICurr, *(&field.Element{}).SetBytes(oldValue[:16]))
	stoHistory.valueLOCurr = append(stoHistory.valueLOCurr, *(&field.Element{}).SetBytes(oldValue[16:]))

	nextValue := frame.Value.(types.FullBytes32)
	stoHistory.valueHINext = append(stoHistory.valueHINext, *(&field.Element{}).SetBytes(nextValue[:16]))
	stoHistory.valueLONext = append(stoHistory.valueLONext, *(&field.Element{}).SetBytes(nextValue[16:]))
	stoHistory.deploymentNumber = append(stoHistory.deploymentNumber, field.NewElement(uint64(currentDeploymentNo)))
	stoHistory.blockNumber = append(stoHistory.blockNumber, field.NewElement(uint64(frame.Block)))

	stoHistory.exists = append(stoHistory.exists, field.NewElement(uint64(currentExistence)))
	stoHistory.existsNew = append(stoHistory.existsNew, field.NewElement(uint64(currentExistence)))
	return stoHistory
}

func (stitcher *Stitcher) AddFrame(frame StateAccessLog) {
	stitcher.accountTouched[frame.Address] = struct{}{}
	accountHistory, keyIsPresent := stitcher.accountData[frame.Address]
	if !keyIsPresent {
		accountHistory = NewAccountHistory()
	}
	stitcher.accountData[frame.Address] = accountHistory.AddFrame(frame, stitcher.initialState, frame.Address)
}

// MapKeysToSlice is an utility function that will output the key set as a slice
// this is necessary because we need to list the keys in the columns in a sorted manner. and in order to use the sort function we need a slice
func MapKeysToSlice[K comparable, V any](accountSet map[K]V) []K {
	// obtain the slice of key values from a map
	accounts := make([]K, 0, len(accountSet))
	for addressIterator := range accountSet {
		accounts = append(accounts, addressIterator)
	}
	return accounts
}
