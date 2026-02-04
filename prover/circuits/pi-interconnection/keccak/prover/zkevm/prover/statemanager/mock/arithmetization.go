package mock

import (
	eth "github.com/consensys/linea-monorepo/prover/circuits/pi-interconnection/keccak/prover/backend/execution/statemanager"
	"github.com/consensys/linea-monorepo/prover/circuits/pi-interconnection/keccak/prover/maths/field"
	"github.com/consensys/linea-monorepo/prover/circuits/pi-interconnection/keccak/prover/utils/types"
)

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
