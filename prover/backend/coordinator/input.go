package coordinator

import (
	"bytes"
	"fmt"
	"io"
	"path"

	"github.com/consensys/accelerated-crypto-monorepo/backend/ethereum"
	"github.com/consensys/accelerated-crypto-monorepo/backend/files"
	"github.com/consensys/accelerated-crypto-monorepo/backend/jsonutil"
	"github.com/consensys/accelerated-crypto-monorepo/backend/statemanager/eth"
	"github.com/consensys/accelerated-crypto-monorepo/backend/statemanager/json"
	"github.com/consensys/accelerated-crypto-monorepo/protocol/wizard"
	"github.com/consensys/accelerated-crypto-monorepo/utils"
	"github.com/consensys/accelerated-crypto-monorepo/zkevm"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/rlp"
	"github.com/sirupsen/logrus"
	"github.com/valyala/fastjson"
)

// The JSON fields names that we use
const (
	ZK_PARENT_STATE_ROOT_HASH       string = "zkParentStateRootHash"
	KECCAK_PARENT_STATE_ROOT_HASH   string = "keccakParentStateRootHash"
	ZK_STATE_MERKLE_PROOF           string = "zkStateMerkleProof"
	TX_INDEX                        string = "transactionIndex"
	TOPICS                          string = "topics"
	DATA                            string = "data"
	ADDRESS                         string = "address"
	BRIDGE_LOGS                     string = "bridgeLogs"
	BLOCKS_DATA                     string = "blocksData"
	CONFLATED_EXECUTION_TRACES_FILE string = "conflatedExecutionTracesFile"
	RLP                             string = "rlp"
)

/*
Fastjson schema used to get the prover output

// Request

	{
	    "zkParentStateRootHash": "0xaf0100...",
	    "conflatedExecutionTracesFile": "file name of traces serialized in JSON and GZIPed (e.g 13673-13675.conflated.v0.0.1.json.gz)",
	    "tracesEngineVersion": "0.2.3",
	    "type2StateManagerVersion": "0.3.4",
	    "zkStateMerkleProof": ["..."],
	    "blocksData": [
	        {
	            "rlp": "full block rlp encoded"
				"bridgeLogs": "array of logs"
	        }
	    ]
	}
*/
type ProverInput struct {
	v *fastjson.Value
}

// Read a ProverInput from a file
func (pi *ProverInput) Read(reader io.Reader) {
	buf := bytes.Buffer{}
	buf.ReadFrom(reader)

	v, err := fastjson.ParseBytes(buf.Bytes())
	if err != nil {
		utils.Panic("Could not parse the input JSON file : %v", err)
	}

	pi.v = v
}

// Returns the keccak state root hash
func (pi *ProverInput) KeccakParentStateRootHash() (eth.Digest, error) {
	return jsonutil.TryGetDigest(*pi.v, KECCAK_PARENT_STATE_ROOT_HASH)
}

// Returns the parsed parent state root hash. Panic if not set
func (pi *ProverInput) ZKParentStateRootHash() eth.Digest {
	res, err := jsonutil.TryGetDigest(*pi.v, ZK_PARENT_STATE_ROOT_HASH)
	if err != nil {
		utils.Panic("could not parse the parent root hash : %v", err)
	}
	return res
}

// Returns the parsed state-manager traces
func (pi *ProverInput) StateManagerTraces() [][]any {
	v := pi.v.Get(ZK_STATE_MERKLE_PROOF)
	if v == nil {
		utils.Panic("missing the state-manager traces, prover expected the JSON key %v", ZK_STATE_MERKLE_PROOF)
	}

	traces, err := json.ParseStateManagerTraces(*v)
	if err != nil {
		utils.Panic("got an error while parsing the traces : %v", err)
	}

	logrus.Tracef("length of the traces in the output of ParseStateManager %v", len(traces))
	return traces
}

// Returns the parsed block data
func (pi *ProverInput) Blocks() []types.Block {

	// Attempt to parse the field as an array
	blocksV, err := jsonutil.TryGetArray(*pi.v, BLOCKS_DATA)
	if err != nil {
		utils.Panic("error while parsing the blocksdata: %v", err)
	}

	// Allocate the result
	res := make([]types.Block, len(blocksV))

	for i := range blocksV {
		// Attempt to parse the block as an hexstring
		blockRLPBytes, err := jsonutil.TryGetHexBytes(blocksV[i], RLP)
		if err != nil {
			utils.Panic("error while parsing the block RLP #%v : %v", i, err)
		}

		buffer := bytes.NewReader(blockRLPBytes)
		block := types.Block{}

		// Attempt to parse the RLP
		err = rlp.Decode(buffer, &block)
		if err != nil {
			utils.Panic("Could not RLP decode the blockRLP 0x%x (block #%v)", blockRLPBytes, i)
		}

		res[i] = block
	}

	return res
}

// Returns the transactions RLP encoded
func RlpTransactions(block *types.Block) []string {
	res := []string{}
	for _, tx := range block.Transactions() {
		txRlp := ethereum.EncodeTxForSigning(tx)
		res = append(res, hexutil.Encode(txRlp))
	}
	logrus.Tracef("computed the RLP of #%v transactions", len(block.Transactions()))
	return res
}

// Returns the list of the From addresses for each
// transaction in the block
func FromAddresses(block *types.Block) []string {
	froms := []string{}
	for _, tx := range block.Transactions() {
		from := ethereum.GetFrom(tx)
		froms = append(froms, hexutil.Encode(from[:]))
	}
	return froms
}

// Returns the path to the conflatedExecutionTracesFile
func (pi *ProverInput) ConflatedExecutionTracesFile() string {
	res, err := jsonutil.TryGetString(*pi.v, CONFLATED_EXECUTION_TRACES_FILE)
	if err != nil {
		utils.Panic("could not get the conflatedTracesFile : %v", err)
	}
	return res
}

// Returns the array of logs
func (pi *ProverInput) LogsForBlock(i int) []types.Log {
	blockList, err := jsonutil.TryGetArray(*pi.v, BLOCKS_DATA)
	if err != nil {
		panic(err)
	}

	if i >= len(blockList) {
		utils.Panic("Out of bound : tried getting block #%v but there are %v blocks", i, len(blockList))
	}

	logs, err := jsonutil.TryGetArray(blockList[i], BRIDGE_LOGS)
	if err != nil {
		panic(err)
	}

	res := make([]types.Log, 0, len(logs))
	for k := range logs {
		parsed := parseLog(logs[k])
		res = append(res, parsed)
	}

	return res
}

// Returns a prover using the internal expander
func (pi *ProverInput) Exprover(traceDir string) func(run *wizard.ProverRuntime) {
	tracePath := path.Join(traceDir, pi.ConflatedExecutionTracesFile())
	return func(run *wizard.ProverRuntime) {
		f := files.MustReadCompressed(tracePath)
		zkevm.AssignFromCorset(f, run)
		f.Close()
	}
}

// Returns true if the state-manager traces are present
func (pi *ProverInput) HasStateManagerTraces() bool {
	parent := pi.v.Get(ZK_PARENT_STATE_ROOT_HASH)
	traces := pi.v.Get(ZK_STATE_MERKLE_PROOF)

	hasParent := jsonutil.Bytes32IsSet(parent)
	hasTraces := jsonutil.ArrayIsSet(traces)

	// Both must be either set or unset but not one and not the other
	if hasParent != hasTraces {
		panic(fmt.Sprintf("either both fields must be nil or both non-nil (trace=%v, parent=%v)", traces, parent))
	}

	return hasParent && hasTraces
}

// Parse the logs. We don't parse every field because we don't need all of them
func parseLog(v fastjson.Value) types.Log {
	res := types.Log{}
	var err error

	// Attempt parsing the emitting contract event
	address, err := jsonutil.TryGetAddress(v, ADDRESS)
	if err != nil {
		utils.Panic("%v", err)
	}
	res.Address = common.Address(address)

	logrus.Tracef("found address: %v", address.Hex())

	// Attempt parsing the data
	res.Data, err = jsonutil.TryGetHexBytes(v, DATA)
	if err != nil {
		utils.Panic("%v", err)
	}

	logrus.Tracef("found data: 0x%x", res.Data)

	// Attempt parsing the topics
	topics, err := jsonutil.TryGetArray(v, TOPICS)
	if err != nil {
		utils.Panic("%v", err)
	}

	logrus.Tracef("found topics: %v", topics)

	res.Topics = make([]common.Hash, len(topics))
	for i := range topics {
		top, err := jsonutil.TryGetBytes32(topics[i])
		if err != nil {
			utils.Panic("%v", err)
		}
		res.Topics[i] = common.Hash(top)
	}

	// Parse the tx index
	txIndex, err := jsonutil.TryGetHexInt(v, TX_INDEX)
	if err != nil {
		utils.Panic("%v", err)
	}
	res.TxIndex = uint(txIndex)

	logrus.Tracef("found txIndex: %v", topics)

	return res
}

// GetRawSignaturesVerification returns the raw signatures verification claims
func (pi *ProverInput) GetRawSignaturesVerificationInputs() (txHashes [][32]byte, pubKeys [][64]byte, signatures [][65]byte) {

	blocks := pi.Blocks()

	// initialize the return values
	txHashes = [][32]byte{}
	pubKeys = [][64]byte{}
	signatures = [][65]byte{}

	for _, block := range blocks {
		for _, tx := range block.Transactions() {

			// compute the verification claims from the transaction and its signature
			txhash := ethereum.GetTxHash(tx)
			sig := ethereum.GetJsonSignature(tx)
			pubkey, encodedSig, err := ethereum.RecoverPublicKey(txhash, sig)

			if err != nil {
				utils.Panic("error recovering public key from transaction: %v", err)
			}

			// append the claims to the return arguments
			txHashes = append(txHashes, txhash)
			signatures = append(signatures, encodedSig)
			pubKeys = append(pubKeys, pubkey)

		}
	}

	return txHashes, pubKeys, signatures
}
