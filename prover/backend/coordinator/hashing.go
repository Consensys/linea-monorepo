package coordinator

import (
	"bytes"
	"math/big"

	"github.com/consensys/gnark-crypto/ecc/bn254/fr"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/sirupsen/logrus"
)

func (j *ProverOutput) ComputeProofInput() {

	debugData := &j.DebugData

	// 1 - For each block
	for i := range j.BlocksData {
		blockInputs := &j.BlocksData[i]
		blockDebug := &PerBlockDebugData{}

		// - Hash the rlp transactions one by one
		txHashes := make([]string, len(blockInputs.RlpEncodedTransactions))
		for i := range txHashes {
			txHashes[i] = HexHashHex(blockInputs.RlpEncodedTransactions[i])
		}
		blockDebug.TxHashes = txHashes
		// - Hash of the transaction hashes concatenated altogether
		blockDebug.HashOfTxHashes = HexHashHex(blockDebug.TxHashes...)
		// - Hash of the log hashes concatenated altogether
		blockDebug.HashOfLogHashes = HexHashHex(blockInputs.L2ToL1MsgHashes...)
		// - Hash of the positions (encoded as uint256)
		{
			batchReceptionIndices := blockInputs.BatchReceptionIndices
			// encode each of them as a uint256
			casted := []string{}
			for i := range batchReceptionIndices {
				casted = append(casted, FmtIntHex(int(batchReceptionIndices[i])))
			}
			blockDebug.HashOfPositions = HexHashHex(casted...)
		}

		// Hash the from addresses : concatenated the addresses and hash the result.
		blockDebug.HashOfFromAddresses = HexHashHex(blockInputs.FromAddresses)

		// f - Get the final hashes
		blockDebug.HashForBlock = HexHashHex(
			blockDebug.HashOfTxHashes,
			blockDebug.HashOfLogHashes,
			blockDebug.HashOfPositions,
			blockDebug.HashOfFromAddresses,
		)

		debugData.Blocks = append(debugData.Blocks, *blockDebug)
	}

	// 2 - Hash for all the blocks altogether
	debugData.HashForAllBlocks = j.HashForAllBlocks()

	// 3 - Hash of the time stamps
	debugData.TimeStampsHash = j.TimeStampHashes()

	// 4 - Hash of the log hashes
	debugData.HashOfRootHashes = j.HashOfRootHashes()

	// 5 - Finally accumulate the first and last block number
	debugData.FinalHash = HexHashHex(
		debugData.HashForAllBlocks,
		FmtIntHex(j.FirstBlockNumber),
		debugData.TimeStampsHash,
		debugData.HashOfRootHashes,
	)

	// and apply the modulus
	debugData.FinalHash = ApplyModulus(debugData.FinalHash)
}

func (j *ProverOutput) TimeStampHashes() string {
	// Collect the timestamps
	timestamps := []uint64{}
	for _, b := range j.BlocksData {
		timestamps = append(timestamps, b.TimeStamp)
	}
	// Then, returns the hash
	return HexHashUint64(timestamps...)
}

func (j *ProverOutput) HashOfRootHashes() string {
	// Collect the root hashes
	rootHashes := []string{j.ParentStateRootHash}
	for _, b := range j.BlocksData {
		rootHashes = append(rootHashes, b.RootHash)
	}
	// Then, returns the hash
	return HexHashHex(rootHashes...)
}

func (j *ProverOutput) HashForAllBlocks() string {
	hashesForBlocks := []string{}
	// Concatenation of the hashes for each block
	for _, b := range j.DebugData.Blocks {
		hashesForBlocks = append(hashesForBlocks, b.HashForBlock)
	}
	// Then, return the hash
	return HexHashHex(hashesForBlocks...)
}

// Parse one or more hex string into a byte array, hash it and
// return the result as an hexstring. If several hex string are
// passed, what is hashed is the concatenation of the strings and
// the hasher is implictly updated only once.
func HexHashHex(v ...string) string {
	buffer := bytes.Buffer{}
	for i := range v {
		decoded, err := hexutil.Decode(v[i])
		if err != nil {
			logrus.Errorf("could not decode %v, because %v. This can happen when"+
				"the state-manager option is activated but no zk-merkleProof were found", v[i], err)
		}
		buffer.Write(decoded)
	}
	res := crypto.Keccak256(buffer.Bytes())
	return hexutil.Encode(res)
}

// Concatenate hex strings
func HexConcat(v ...string) string {
	buffer := bytes.Buffer{}
	for i := range v {
		decoded := hexutil.MustDecode(v[i])
		buffer.Write(decoded)
	}
	return hexutil.Encode(buffer.Bytes())
}

// Encode the uint64 into an hexstring representing it as a u256 in bigendian form
func HexHashUint64(v ...uint64) string {
	buffer := bytes.Buffer{}
	for i := range v {
		bytes := big.NewInt(int64(v[i])).Bytes()
		bytes = append(make([]byte, 32-len(bytes)), bytes...)
		buffer.Write(bytes)
	}
	res := crypto.Keccak256(buffer.Bytes())
	return hexutil.Encode(res)
}

// Format an integer as a 32 bytes hex string
func FmtIntHex(v int) string {
	bytes := big.NewInt(int64(v)).Bytes()
	bytes = append(make([]byte, 32-len(bytes)), bytes...)
	return hexutil.Encode(bytes)
}

// Apply the modulus
func ApplyModulus(b string) string {
	var f fr.Element
	f.SetString(b)
	fbytes := f.Bytes()
	return hexutil.Encode(fbytes[:])
}
