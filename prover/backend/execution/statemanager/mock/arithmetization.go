package mock

import (
	"math/big"
	"sort"

	eth "github.com/consensys/zkevm-monorepo/prover/backend/execution/statemanager"
	"github.com/consensys/zkevm-monorepo/prover/maths/field"
	"github.com/consensys/zkevm-monorepo/prover/utils"
	"github.com/consensys/zkevm-monorepo/prover/utils/types"
)

// StateManagerVectors contains all the arithmetization columns needed for the state management at the account and storage levels
// (all these columns belong to the HUB module in the arithmetization)
// these slices will be at a later point be transformed into corresponding ifaces.Column() structs
type StateManagerVectors struct {
	// account data
	address                                              []eth.Address // helper column
	addressHI, addressLO                                 []field.Element
	nonce, nonceNew                                      []field.Element
	mimcCodeHash, mimcCodeHashNew                        []field.Element
	codeHashHI, codeHashLO, codeHashHINew, codeHashLONew []field.Element
	codeSizeOld, codeSizeNew                             []field.Element
	balanceOld, balanceNew                               []field.Element
	// storage data
	keyHI, keyLO                                       []field.Element
	valueHICurr, valueLOCurr, valueHINext, valueLONext []field.Element
	// helper numbers
	deploymentNumber, deploymentNumberInf []field.Element
	blockNumber                           []field.Element
	// helper columns
	exists, existsNew []field.Element
	peekAtAccount     []field.Element
	peekAtStorage     []field.Element
	// first and last marker columns
	firstAOC, lastAOC []field.Element
	firstKOC, lastKOC []field.Element
}

// NewStateManagerVectors will initialize a StateManagerVectors struct, where the containing slices are initialized but are empty
func NewStateManagerVectors() *StateManagerVectors {
	return &StateManagerVectors{
		// account info
		address:         []eth.Address{},
		addressHI:       []field.Element{},
		addressLO:       []field.Element{},
		nonce:           []field.Element{},
		nonceNew:        []field.Element{},
		mimcCodeHash:    []field.Element{},
		mimcCodeHashNew: []field.Element{},
		codeHashHI:      []field.Element{},
		codeHashLO:      []field.Element{},
		codeHashHINew:   []field.Element{},
		codeHashLONew:   []field.Element{},
		codeSizeOld:     []field.Element{},
		codeSizeNew:     []field.Element{},
		balanceOld:      []field.Element{},
		balanceNew:      []field.Element{},
		// storage data
		keyHI:       []field.Element{},
		keyLO:       []field.Element{},
		valueHICurr: []field.Element{},
		valueLOCurr: []field.Element{},
		valueHINext: []field.Element{},
		valueLONext: []field.Element{},
		// helper numbers
		deploymentNumber:    []field.Element{},
		deploymentNumberInf: []field.Element{},
		blockNumber:         []field.Element{},
		// helper columns
		exists:        []field.Element{},
		existsNew:     []field.Element{},
		peekAtAccount: []field.Element{},
		peekAtStorage: []field.Element{},
		// first and last marker columns
		firstAOC: []field.Element{},
		lastAOC:  []field.Element{},
		firstKOC: []field.Element{},
		lastKOC:  []field.Element{},
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
}

// StorageHistory contains column segments that describe the storage evolution for a specific key
// the keyHI, keyLO values are always the same and will be used for stitching
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

// The Stitcher constructs the vector slices of field.Element that will correspond to the HUB columns
// it will store every AccountHistory which appears in the columns, and manages that using a map
type Stitcher struct {
	accountData map[eth.Address]*AccountHistory
}

// NewAccountHistory initializes a new account history where each column exists but is empty
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
func (accHistory *AccountHistory) InitializeNewRow(currentBlock int) *AccountHistory {
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
		// prevBlockNumber = (*accHistory.blockNumber)[lastIndex] // use previous block number
	} else {
		// when the previous values do not exist yet, initialize them with 0s
		prevAddress, _ = types.AddressFromHex("0x")
		prevAddressHI.SetZero()
		prevAddressLO.SetZero()
		prevNonce.SetZero()
		prevMimcCodeHash.SetZero()
		prevCodeHashHI.SetZero()
		prevCodeHashLO.SetZero()
		prevCodeSize.SetZero()
		prevBalance.SetZero()
		prevDeploymentNumber.SetZero() // when the state is initialized, we start with deploymentNumber = 0
		// prevBlockNumber.SetInt64(int64(currentBlock))
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

// InitializeFromState will populate the column fragments inside the Stitcher struct (the column fragments are stored in the included AccountHistory and StorageHistory)
// since the State does not specify a block number, this has to be specified
// the current implementation sets the deployment number on this data to be 0
func (stitcher *Stitcher) InitializeFromState(currentBlock int, state State) *Stitcher {

	stitcher.accountData = map[eth.Address]*AccountHistory{}

	for accountAddress := range state {
		accHistory := NewAccountHistory()
		accHistory.currentExistence = 1 // when we initialize from state we will assume that the account exists
		accHistory.InitializeNewRow(currentBlock)
		// account information
		accHistory.address[0] = accountAddress
		accHistory.addressHI[0].SetBytes(accountAddress[:4])
		accHistory.addressLO[0].SetBytes(accountAddress[4:20])
		accHistory.nonce[0].SetZero() // Old column initialized to 0
		accHistory.nonceNew[0].SetInt64(state.GetNonce(accountAddress))
		// hashes
		mimcCodeHash := state.GetMimcCodeHash(accountAddress)
		accHistory.mimcCodeHash[0].SetZero() // Old column initialized to 0
		accHistory.mimcCodeHashNew[0].SetBytes(mimcCodeHash[:])
		accHistory.codeHashHI[0].SetZero() // Old column initialized to 0
		accHistory.codeHashLO[0].SetZero() // Old column initialized to 0
		codeHash := state.GetCodeHash(accountAddress)
		accHistory.codeHashHINew[0].SetBytes(codeHash[:16])
		accHistory.codeHashLONew[0].SetBytes(codeHash[16:])
		accHistory.codeSizeOld[0].SetZero() // Old column initialized to 0
		accHistory.codeSizeNew[0].SetInt64(state.GetCodeSize(accountAddress))
		// balance
		accHistory.balanceOld[0].SetZero() // Old column initialized to 0
		accHistory.balanceNew[0].SetBigInt(state.GetBalance(accountAddress))

		storageMap := state[accountAddress].Storage
		for storageKey := range storageMap {
			stoHist := NewStorageHistory()
			stoHist.address = append(stoHist.address, accountAddress)
			stoHist.addressHI = append(stoHist.addressHI, *(&field.Element{}).SetBytes(accountAddress[:4]))
			stoHist.addressLO = append(stoHist.addressLO, *(&field.Element{}).SetBytes(accountAddress[4:]))

			stoHist.keyHI = append(stoHist.keyHI, *(&field.Element{}).SetBytes(storageKey[:16]))
			stoHist.keyLO = append(stoHist.keyLO, *(&field.Element{}).SetBytes(storageKey[16:]))

			oldValue := field.Zero() // no old value on the initialization step
			oldValueSlice := oldValue.Bytes()
			stoHist.valueHICurr = append(stoHist.valueHICurr, *(&field.Element{}).SetBytes(oldValueSlice[:16]))
			stoHist.valueLOCurr = append(stoHist.valueLOCurr, *(&field.Element{}).SetBytes(oldValueSlice[16:]))

			nextValue := storageMap[storageKey]
			stoHist.valueHINext = append(stoHist.valueHINext, *(&field.Element{}).SetBytes(nextValue[:16]))
			stoHist.valueLONext = append(stoHist.valueLONext, *(&field.Element{}).SetBytes(nextValue[16:]))
			stoHist.deploymentNumber = append(stoHist.deploymentNumber, field.NewElement(0)) // current deployment number is 0
			stoHist.blockNumber = append(stoHist.blockNumber, field.NewElement(uint64(currentBlock)))

			stoHist.exists = append(stoHist.exists, field.NewElement(1))       // existence initialized as 1, can be extended later
			stoHist.existsNew = append(stoHist.existsNew, field.NewElement(1)) // exists new initialized to 1, can be extended later

			accHistory.storageData[storageKey] = stoHist
		}

		stitcher.accountData[accountAddress] = accHistory
	}
	return stitcher

}

// PadToNearestPowerOf2 will pad all the vectors with zeroes so that their length becomes a power of two and they can be transformed into iColumn structs
func (stateManagerVectors *StateManagerVectors) PadToNearestPowerOf2() {
	// first compute the necessary length for padding
	fullLength := utils.NextPowerOfTwo(len(stateManagerVectors.address))
	dummyVectorAddress := make([]eth.Address, fullLength-len(stateManagerVectors.address))
	dummyVector := make([]field.Element, fullLength-len(stateManagerVectors.address))
	// pad account columns
	stateManagerVectors.address = append(stateManagerVectors.address, dummyVectorAddress...)
	stateManagerVectors.addressHI = append(stateManagerVectors.addressHI, dummyVector...)
	stateManagerVectors.addressLO = append(stateManagerVectors.addressLO, dummyVector...)
	stateManagerVectors.nonce = append(stateManagerVectors.nonce, dummyVector...)
	stateManagerVectors.nonceNew = append(stateManagerVectors.nonceNew, dummyVector...)
	stateManagerVectors.mimcCodeHash = append(stateManagerVectors.mimcCodeHash, dummyVector...)
	stateManagerVectors.mimcCodeHashNew = append(stateManagerVectors.mimcCodeHashNew, dummyVector...)
	stateManagerVectors.codeHashHI = append(stateManagerVectors.codeHashHI, dummyVector...)
	stateManagerVectors.codeHashLO = append(stateManagerVectors.codeHashLO, dummyVector...)
	stateManagerVectors.codeHashHINew = append(stateManagerVectors.codeHashHINew, dummyVector...)
	stateManagerVectors.codeHashLONew = append(stateManagerVectors.codeHashLONew, dummyVector...)
	stateManagerVectors.codeSizeOld = append(stateManagerVectors.codeSizeOld, dummyVector...)
	stateManagerVectors.codeSizeNew = append(stateManagerVectors.codeSizeNew, dummyVector...)
	stateManagerVectors.balanceOld = append(stateManagerVectors.balanceOld, dummyVector...)
	stateManagerVectors.balanceNew = append(stateManagerVectors.balanceNew, dummyVector...)
	stateManagerVectors.deploymentNumber = append(stateManagerVectors.deploymentNumber, dummyVector...)
	stateManagerVectors.blockNumber = append(stateManagerVectors.blockNumber, dummyVector...)
	stateManagerVectors.exists = append(stateManagerVectors.exists, dummyVector...)
	stateManagerVectors.existsNew = append(stateManagerVectors.existsNew, dummyVector...)
	// pad storage columns
	stateManagerVectors.keyHI = append(stateManagerVectors.keyHI, dummyVector...)
	stateManagerVectors.keyLO = append(stateManagerVectors.keyLO, dummyVector...)
	stateManagerVectors.valueHICurr = append(stateManagerVectors.valueHICurr, dummyVector...)
	stateManagerVectors.valueLOCurr = append(stateManagerVectors.valueLOCurr, dummyVector...)
	stateManagerVectors.valueHINext = append(stateManagerVectors.valueHINext, dummyVector...)
	stateManagerVectors.valueLONext = append(stateManagerVectors.valueLONext, dummyVector...)

	// pad peekAtStorage
	stateManagerVectors.peekAtStorage = append(stateManagerVectors.peekAtStorage, dummyVector...)

	// pad peekAtAccount
	stateManagerVectors.peekAtAccount = append(stateManagerVectors.peekAtAccount, dummyVector...)

	// pad first AOC and last AOC
	stateManagerVectors.firstAOC = append(stateManagerVectors.firstAOC, dummyVector...)
	stateManagerVectors.lastAOC = append(stateManagerVectors.lastAOC, dummyVector...)

	// pad first KOC and last KOC
	stateManagerVectors.firstKOC = append(stateManagerVectors.firstKOC, dummyVector...)
	stateManagerVectors.lastKOC = append(stateManagerVectors.lastKOC, dummyVector...)

	// pad deployment Number Infinity
	stateManagerVectors.deploymentNumberInf = append(stateManagerVectors.deploymentNumberInf, dummyVector...)
}

// Finalize uses the information in the AccountHistory structs that are stored inside the Stitcher and concatenates these vectors
// AccountHistory contains multiple StorageHistory, one for each storage key that ever gets accessed.
// Finalize also adds the information in StorageHistory to the concatenated vectors.
// The final result is padded so that the length of the columns is a power of 2.
func (stitcher *Stitcher) Finalize() *StateManagerVectors {
	sortedAccountAddresses := MapKeysToSlice(stitcher.accountData)
	sort.Slice(sortedAccountAddresses, func(i, j int) bool {
		s1 := sortedAccountAddresses[i].Hex()
		s2 := sortedAccountAddresses[j].Hex()
		return s1 < s2
	})

	stateManagerVectors := NewStateManagerVectors()
	// stitch account columns
	for _, accountAddress := range sortedAccountAddresses {
		accHist := stitcher.accountData[accountAddress]
		stateManagerVectors.address = append(stateManagerVectors.address, accHist.address...)
		stateManagerVectors.addressHI = append(stateManagerVectors.addressHI, accHist.addressHI...)
		stateManagerVectors.addressLO = append(stateManagerVectors.addressLO, accHist.addressLO...)
		stateManagerVectors.nonce = append(stateManagerVectors.nonce, accHist.nonce...)
		stateManagerVectors.nonceNew = append(stateManagerVectors.nonceNew, accHist.nonceNew...)
		stateManagerVectors.mimcCodeHash = append(stateManagerVectors.mimcCodeHash, accHist.mimcCodeHash...)
		stateManagerVectors.mimcCodeHashNew = append(stateManagerVectors.mimcCodeHashNew, accHist.mimcCodeHashNew...)
		stateManagerVectors.codeHashHI = append(stateManagerVectors.codeHashHI, accHist.codeHashHI...)
		stateManagerVectors.codeHashLO = append(stateManagerVectors.codeHashLO, accHist.codeHashLO...)
		stateManagerVectors.codeHashHINew = append(stateManagerVectors.codeHashHINew, accHist.codeHashHINew...)
		stateManagerVectors.codeHashLONew = append(stateManagerVectors.codeHashLONew, accHist.codeHashLONew...)
		stateManagerVectors.codeSizeOld = append(stateManagerVectors.codeSizeOld, accHist.codeSizeOld...)
		stateManagerVectors.codeSizeNew = append(stateManagerVectors.codeSizeNew, accHist.codeSizeNew...)
		stateManagerVectors.balanceOld = append(stateManagerVectors.balanceOld, accHist.balanceOld...)
		stateManagerVectors.balanceNew = append(stateManagerVectors.balanceNew, accHist.balanceNew...)
		stateManagerVectors.deploymentNumber = append(stateManagerVectors.deploymentNumber, accHist.deploymentNumber...)
		stateManagerVectors.blockNumber = append(stateManagerVectors.blockNumber, accHist.blockNumber...)
		stateManagerVectors.exists = append(stateManagerVectors.exists, accHist.exists...)
		stateManagerVectors.existsNew = append(stateManagerVectors.existsNew, accHist.existsNew...)

		// add dummy rows
		dummyVector := make([]field.Element, len(accHist.addressHI))
		stateManagerVectors.keyHI = append(stateManagerVectors.keyHI, dummyVector...)
		stateManagerVectors.keyLO = append(stateManagerVectors.keyLO, dummyVector...)
		stateManagerVectors.valueHICurr = append(stateManagerVectors.valueHICurr, dummyVector...)
		stateManagerVectors.valueLOCurr = append(stateManagerVectors.valueLOCurr, dummyVector...)
		stateManagerVectors.valueHINext = append(stateManagerVectors.valueHINext, dummyVector...)
		stateManagerVectors.valueLONext = append(stateManagerVectors.valueLONext, dummyVector...)

		// add dummy helper column, all zeros
		stateManagerVectors.peekAtStorage = append(stateManagerVectors.peekAtStorage, dummyVector...)

		// initialize the peek at account column and then add it
		peekAtAccountFragment := make([]field.Element, len(accHist.addressHI))
		for index := range peekAtAccountFragment {
			peekAtAccountFragment[index].SetOne()
		}
		stateManagerVectors.peekAtAccount = append(stateManagerVectors.peekAtAccount, peekAtAccountFragment...)

		// first AOC and last AOC
		firstAOCFragment := make([]field.Element, len(accHist.addressHI))
		firstAOCFragment[0].SetOne()
		lastAOCFragment := make([]field.Element, len(accHist.addressHI))
		lastAOCFragment[len(lastAOCFragment)-1].SetOne()
		stateManagerVectors.firstAOC = append(stateManagerVectors.firstAOC, firstAOCFragment...)
		stateManagerVectors.lastAOC = append(stateManagerVectors.lastAOC, lastAOCFragment...)

		// first KOC and last KOC
		stateManagerVectors.firstKOC = append(stateManagerVectors.firstKOC, dummyVector...)
		stateManagerVectors.lastKOC = append(stateManagerVectors.lastKOC, dummyVector...)

		// deployment Number Infinity
		deploymentInfFragment := make([]field.Element, len(accHist.addressHI))
		for index := range deploymentInfFragment {
			deploymentInfFragment[index].SetInt64(int64(accHist.currentDeploymentNo))
		}
		stateManagerVectors.deploymentNumberInf = append(stateManagerVectors.deploymentNumberInf, deploymentInfFragment...)
	}

	// add storage keys
	for _, accountAddress := range sortedAccountAddresses {
		accHist := stitcher.accountData[accountAddress]
		storageHistMap := accHist.storageData
		sortedStorageKeys := MapKeysToSlice(storageHistMap)
		for _, key := range sortedStorageKeys {
			stoHist := storageHistMap[key]
			stateManagerVectors.address = append(stateManagerVectors.address, stoHist.address...)
			stateManagerVectors.addressHI = append(stateManagerVectors.addressHI, stoHist.addressHI...)
			stateManagerVectors.addressLO = append(stateManagerVectors.addressLO, stoHist.addressLO...)

			stateManagerVectors.deploymentNumber = append(stateManagerVectors.deploymentNumber, stoHist.deploymentNumber...)
			stateManagerVectors.blockNumber = append(stateManagerVectors.blockNumber, stoHist.blockNumber...)
			stateManagerVectors.exists = append(stateManagerVectors.exists, stoHist.exists...)
			stateManagerVectors.existsNew = append(stateManagerVectors.existsNew, stoHist.existsNew...)
			// add keys and values
			stateManagerVectors.keyHI = append(stateManagerVectors.keyHI, stoHist.keyHI...)
			stateManagerVectors.keyLO = append(stateManagerVectors.keyLO, stoHist.keyLO...)
			stateManagerVectors.valueHICurr = append(stateManagerVectors.valueHICurr, stoHist.valueHICurr...)
			stateManagerVectors.valueLOCurr = append(stateManagerVectors.valueLOCurr, stoHist.valueLOCurr...)
			stateManagerVectors.valueHINext = append(stateManagerVectors.valueHINext, stoHist.valueHINext...)
			stateManagerVectors.valueLONext = append(stateManagerVectors.valueLONext, stoHist.valueLONext...)
			// add dummy rows
			dummyVector := make([]field.Element, len(stoHist.addressHI))
			stateManagerVectors.nonce = append(stateManagerVectors.nonce, dummyVector...)
			stateManagerVectors.nonceNew = append(stateManagerVectors.nonceNew, dummyVector...)

			stateManagerVectors.mimcCodeHash = append(stateManagerVectors.mimcCodeHash, dummyVector...)
			stateManagerVectors.mimcCodeHashNew = append(stateManagerVectors.mimcCodeHashNew, dummyVector...)

			stateManagerVectors.codeHashHI = append(stateManagerVectors.codeHashHI, dummyVector...)
			stateManagerVectors.codeHashLO = append(stateManagerVectors.codeHashLO, dummyVector...)

			stateManagerVectors.codeHashHINew = append(stateManagerVectors.codeHashHINew, dummyVector...)
			stateManagerVectors.codeHashLONew = append(stateManagerVectors.codeHashLONew, dummyVector...)

			stateManagerVectors.codeSizeOld = append(stateManagerVectors.codeSizeOld, dummyVector...)
			stateManagerVectors.codeSizeNew = append(stateManagerVectors.codeSizeNew, dummyVector...)

			stateManagerVectors.balanceOld = append(stateManagerVectors.balanceOld, dummyVector...)
			stateManagerVectors.balanceNew = append(stateManagerVectors.balanceNew, dummyVector...)

			// add dummy helper column, all zeros
			stateManagerVectors.peekAtAccount = append(stateManagerVectors.peekAtAccount, dummyVector...)

			// initialize the peek at storage column segment and then add it
			peekAtStorageFragment := make([]field.Element, len(stoHist.addressHI))
			for index := range peekAtStorageFragment {
				peekAtStorageFragment[index].SetOne()
			}
			stateManagerVectors.peekAtStorage = append(stateManagerVectors.peekAtStorage, peekAtStorageFragment...)

			// firstAOC and lastAOC columns
			stateManagerVectors.firstAOC = append(stateManagerVectors.firstAOC, dummyVector...)
			stateManagerVectors.lastAOC = append(stateManagerVectors.lastAOC, dummyVector...)

			// first KOC and last KOC initializing and adding
			firstKOCFragment := make([]field.Element, len(stoHist.addressHI))
			firstKOCFragment[0].SetOne()
			lastKOCFragment := make([]field.Element, len(stoHist.addressHI))
			lastKOCFragment[len(lastKOCFragment)-1].SetOne()
			stateManagerVectors.firstKOC = append(stateManagerVectors.firstKOC, firstKOCFragment...)
			stateManagerVectors.lastKOC = append(stateManagerVectors.lastKOC, lastKOCFragment...)

			// deployment Number Infinity
			deploymentInfFragment := make([]field.Element, len(stoHist.addressHI))
			for index := range deploymentInfFragment {
				deploymentInfFragment[index].SetInt64(int64(accHist.currentDeploymentNo))
			}
			stateManagerVectors.deploymentNumberInf = append(stateManagerVectors.deploymentNumberInf, deploymentInfFragment...)
		}
	}
	stateManagerVectors.PadToNearestPowerOf2()
	return stateManagerVectors
}

// AddFrame uses the information inside a frame to add one more row in AccountHistory
// by appending one entry to each slice contained in AccountHistory
func (accHistory *AccountHistory) AddFrame(frame StateAccessLog) *AccountHistory {
	accHistory.InitializeNewRow(frame.Block) // this sets the block number as well
	lastIndex := len(accHistory.addressHI) - 1
	accHistory.address[lastIndex] = frame.Address
	accHistory.addressHI[lastIndex].SetBytes(frame.Address[:4])
	accHistory.addressLO[lastIndex].SetBytes(frame.Address[4:])
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
		accHistory.codeHashHINew[lastIndex].SetBytes(codeHash[16:])
	case AccountErasal:
		if accHistory.currentExistence == 0 {
			panic("Attempted to delete an account which does not exist.")
		}
		accHistory.currentExistence = 0
		accHistory.exists[lastIndex] = field.One()     // it exists when delete is invoked
		accHistory.existsNew[lastIndex] = field.Zero() // it does not exist afterward
	case AccountInit:
		if accHistory.currentExistence == 1 {
			panic("Attempted to initialize an account which already exists.")
		}
		accHistory.currentDeploymentNo++
		accHistory.currentExistence = 1
		accHistory.deploymentNumber[lastIndex].SetUint64(uint64(accHistory.currentDeploymentNo))
		accHistory.exists[lastIndex] = field.Zero()   // it does not exist when create is invoked
		accHistory.existsNew[lastIndex] = field.One() // it exists afterwards
	}
	return accHistory
}

// AddFrame uses the information inside a frame to create a new row in the StorageHistory, by adding one entry to each column
// currentExistence is derived from the accountHistory and is used to compute the exists and existsNew columns
// currentDeploymentNumber is derived from the accountHistory and will be used to compute the deployment number
func (stoHistory *StorageHistory) AddFrame(frame StateAccessLog, currentDeploymentNo int, currentExistence int) *StorageHistory {

	stoHistory.address = append(stoHistory.address, frame.Address)
	stoHistory.addressHI = append(stoHistory.addressHI, *(&field.Element{}).SetBytes(frame.Address[:4]))
	stoHistory.addressLO = append(stoHistory.addressLO, *(&field.Element{}).SetBytes(frame.Address[4:]))

	stoHistory.keyHI = append(stoHistory.keyHI, *(&field.Element{}).SetBytes(frame.Key[:16]))
	stoHistory.keyLO = append(stoHistory.keyLO, *(&field.Element{}).SetBytes(frame.Key[16:]))

	var oldValue types.FullBytes32
	if frame.OldValue == nil {
		zeroElem := field.Zero()
		oldValue = zeroElem.Bytes()
	} else {
		oldValue = frame.OldValue.(types.FullBytes32)
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
	accountHistory, keyIsPresent := stitcher.accountData[frame.Address]
	if !keyIsPresent {
		accountHistory = NewAccountHistory()
	}
	stitcher.accountData[frame.Address] = accountHistory.AddFrame(frame)
}

// MapKeysToSlice is an utility function that will output the key set as a slice
// this is necessary because we need to list the keys in the columns in a sorted manner. and in order to use the sort function we need a slice
func MapKeysToSlice[K comparable, V any](accountSet map[K]V) []K {
	// obtain the slice of key values from a map
	var accounts []K
	for addressIterator := range accountSet {
		accounts = append(accounts, addressIterator)
	}
	return accounts
}
