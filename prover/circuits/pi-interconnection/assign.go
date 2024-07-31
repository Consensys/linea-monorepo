package pi_interconnection

import (
	"bytes"
	"encoding/base64"
	"errors"
	"hash"

	"github.com/consensys/gnark-crypto/ecc/bls12-381/fr"
	"github.com/consensys/linea-monorepo/prover/backend/blobsubmission"
	"github.com/consensys/linea-monorepo/prover/circuits/aggregation"
	decompression "github.com/consensys/linea-monorepo/prover/circuits/blobdecompression/v1"
	"github.com/consensys/linea-monorepo/prover/circuits/execution"
	"github.com/consensys/linea-monorepo/prover/circuits/internal"
	"github.com/consensys/linea-monorepo/prover/circuits/pi-interconnection/keccak"
	"github.com/consensys/linea-monorepo/prover/lib/compressor/blob"
	public_input "github.com/consensys/linea-monorepo/prover/public-input"
	"github.com/consensys/linea-monorepo/prover/utils"
	"golang.org/x/crypto/sha3"
)

type ExecutionRequest struct {
	L2MsgHashes            [][32]byte
	FinalStateRootHash     [32]byte
	FinalBlockNumber       uint64
	FinalBlockTimestamp    uint64
	FinalRollingHash       [32]byte
	FinalRollingHashNumber uint64
}

type Request struct {
	DecompDict     []byte
	Decompressions []blobsubmission.Response
	Executions     []ExecutionRequest
	Aggregation    public_input.Aggregation
}

func (c *Compiled) Assign(r Request) (a Circuit, err error) {
	internal.RegisterHints()
	keccak.RegisterHints()

	// TODO there is data duplication in the request. Check consistency

	// infer config
	config := c.getConfig()
	a = config.allocateCircuit()

	if len(r.Decompressions) > config.MaxNbDecompression {
		err = errors.New("number of decompression proofs exceeds maximum")
		return
	}
	if len(r.Executions) > config.MaxNbExecution {
		err = errors.New("number of execution proofs exceeds maximum")
		return
	}
	if len(r.Decompressions)+len(r.Executions) > config.MaxNbCircuits && config.MaxNbCircuits > 0 {
		err = errors.New("total number of circuits exceeds maximum")
		return
	}

	dict, err := blob.GetDict() // TODO look up dict based on checksum
	if err != nil {
		return
	}

	// For Shnarfs and Merkle Roots
	hshK := c.Keccak.GetHasher()

	prevShnarf, err := utils.HexDecodeString(r.Aggregation.ParentAggregationFinalShnarf)
	if err != nil {
		return
	}
	internal.Copy(a.ParentShnarf[:], prevShnarf)

	execDataChecksums := make([][]byte, 0, len(r.Executions))
	shnarfs := make([][]byte, config.MaxNbDecompression)
	// Decompression FPI
	for i, p := range r.Decompressions {
		var blobData [1024 * 128]byte
		if b, err := base64.StdEncoding.DecodeString(p.CompressedData); err != nil {
			return a, err
		} else {
			copy(blobData[:], b)
		}

		var (
			x [32]byte
			y fr.Element
		)
		var b []byte
		if b, err = utils.HexDecodeString(p.ExpectedX); err != nil { // TODO this is reduced. find how to get the unreduced value
			return
		} else {
			copy(x[:], b)
		}
		if _, err = y.SetString(p.ExpectedY); err != nil {
			return
		}
		if shnarfs[i], err = utils.HexDecodeString(p.ExpectedShnarf); err != nil {
			return
		}

		// TODO this recomputes much of the data in p; check consistency
		var (
			fpi  decompression.FunctionalPublicInput
			sfpi decompression.FunctionalPublicInputSnark
		)
		if fpi, err = decompression.AssignFPI(blobData[:], dict, p.Eip4844Enabled, x, y); err != nil {
			return
		}
		execDataChecksums = append(execDataChecksums, fpi.BatchSums...) // len(execDataChecksums) = index of the first execution associated with the next blob
		if sfpi, err = fpi.ToSnarkType(); err != nil {
			return
		}
		a.DecompressionFPIQ[i] = sfpi.FunctionalPublicInputQSnark
		if a.DecompressionPublicInput[i], err = fpi.Sum(); err != nil {
			return
		}

		// recompute shnarf
		shnarf := blobsubmission.Shnarf{
			OldShnarf:        prevShnarf,
			SnarkHash:        fpi.SnarkHash,
			NewStateRootHash: r.Executions[len(execDataChecksums)-1].FinalStateRootHash[:],
			X:                fpi.X[:],
			Y:                y,
			Hash:             &hshK,
		}

		if prevShnarf = shnarf.Compute(); !bytes.Equal(prevShnarf, shnarfs[i]) {
			err = errors.New("shnarf mismatch")
			return
		}
	}
	if len(execDataChecksums) != len(r.Executions) {
		err = errors.New("number of execution circuits does not match the number of batches in decompression circuits")
		return
	}
	var zero [32]byte
	for i := len(r.Decompressions); i < len(a.DecompressionFPIQ); i++ {
		shnarf := blobsubmission.Shnarf{
			OldShnarf:        prevShnarf,
			SnarkHash:        zero[:],
			NewStateRootHash: r.Executions[len(execDataChecksums)-1].FinalStateRootHash[:],
			X:                zero[:],
			Hash:             &hshK,
		}
		prevShnarf = shnarf.Compute()
		shnarfs[i] = prevShnarf

		fpi := decompression.FunctionalPublicInput{
			SnarkHash: zero[:],
		}

		if fpis, err := fpi.ToSnarkType(); err != nil {
			return a, nil
		} else {
			a.DecompressionFPIQ[i] = fpis.FunctionalPublicInputQSnark
		}
		internal.Copy(a.DecompressionFPIQ[i].X[:], zero[:])
		if a.DecompressionPublicInput[i], err = fpi.Sum(decompression.WithBatchesSum(zero[:])); err != nil { // TODO zero batches sum is probably incorrect
			return
		}
	}

	// Aggregation FPI

	aggregationFPI, err := aggregation.NewFunctionalPublicInput(&r.Aggregation)
	if err != nil {
		return
	}
	if !bytes.Equal(shnarfs[len(r.Decompressions)-1], aggregationFPI.FinalShnarf[:]) {
		err = errors.New("mismatch between decompression/aggregation-supplied shnarfs")
		return
	}
	aggregationFPI.NbDecompression = uint64(len(r.Decompressions))
	a.FunctionalPublicInputQSnark = aggregationFPI.ToSnarkType().FunctionalPublicInputQSnark

	merkleNbLeaves := 1 << config.L2MsgMerkleDepth
	maxNbL2MessageHashes := config.L2MessageMaxNbMerkle * merkleNbLeaves
	l2MessageHashes := make([][32]byte, 0, maxNbL2MessageHashes)
	// Execution FPI
	executionFPI := execution.FunctionalPublicInput{
		FinalStateRootHash:     aggregationFPI.InitialStateRootHash,
		FinalBlockNumber:       aggregationFPI.InitialBlockNumber,
		FinalBlockTimestamp:    aggregationFPI.InitialBlockTimestamp,
		FinalRollingHash:       aggregationFPI.InitialRollingHash,
		FinalRollingHashNumber: aggregationFPI.InitialRollingHashNumber,
		L2MessageServiceAddr:   aggregationFPI.L2MessageServiceAddr,
		ChainID:                aggregationFPI.ChainID,
		MaxNbL2MessageHashes:   config.MaxNbMsgPerExecution,
	}
	for i := range a.ExecutionFPIQ {
		executionFPI.InitialRollingHash = executionFPI.FinalRollingHash
		executionFPI.InitialBlockNumber = executionFPI.FinalBlockNumber
		executionFPI.InitialBlockTimestamp = executionFPI.FinalBlockTimestamp
		executionFPI.InitialRollingHash = executionFPI.FinalRollingHash
		executionFPI.InitialRollingHashNumber = executionFPI.FinalRollingHashNumber
		executionFPI.InitialStateRootHash = executionFPI.FinalStateRootHash
		executionFPI.L2MessageHashes = nil

		if i < len(r.Executions) {
			executionFPI.FinalRollingHash = r.Executions[i].FinalRollingHash
			executionFPI.FinalBlockNumber = r.Executions[i].FinalBlockNumber
			executionFPI.FinalBlockTimestamp = r.Executions[i].FinalBlockTimestamp
			executionFPI.FinalRollingHash = r.Executions[i].FinalRollingHash
			executionFPI.FinalRollingHashNumber = r.Executions[i].FinalRollingHashNumber
			executionFPI.FinalStateRootHash = r.Executions[i].FinalStateRootHash

			copy(executionFPI.DataChecksum[:], execDataChecksums[i])
			executionFPI.L2MessageHashes = r.Executions[i].L2MsgHashes

			l2MessageHashes = append(l2MessageHashes, r.Executions[i].L2MsgHashes...)
		}

		a.ExecutionPublicInput[i] = executionFPI.Sum()
		a.ExecutionFPIQ[i] = executionFPI.ToSnarkType().FunctionalPublicInputQSnark
	}
	// consistency check
	if executionFPI.FinalBlockTimestamp != aggregationFPI.FinalBlockTimestamp ||
		executionFPI.FinalBlockNumber != aggregationFPI.FinalBlockNumber ||
		executionFPI.FinalRollingHash != aggregationFPI.FinalRollingHash ||
		executionFPI.FinalRollingHashNumber != aggregationFPI.FinalRollingHashNumber {
		err = errors.New("final execution values not matching final aggregation values")
		return
	}
	if len(l2MessageHashes) > maxNbL2MessageHashes {
		err = errors.New("too many L2 messages")
		return
	}

	for i := range r.Aggregation.L2MsgRootHashes {
		var expectedRoot []byte
		if expectedRoot, err = utils.HexDecodeString(r.Aggregation.L2MsgRootHashes[i]); err != nil {
			return
		}
		computedRoot := MerkleRoot(&hshK, merkleNbLeaves, l2MessageHashes[i*merkleNbLeaves:min((i+1)*merkleNbLeaves, len(l2MessageHashes))])
		if !bytes.Equal(expectedRoot[:], computedRoot[:]) {
			err = errors.New("merkle root mismatch")
			return
		}
	}
	// padding merkle root hashes
	emptyTree := make([][]byte, config.L2MsgMerkleDepth+1)
	emptyTree[0] = make([]byte, 64)
	hsh := sha3.NewLegacyKeccak256()
	for i := 1; i < len(emptyTree); i++ {
		hsh.Reset()
		hsh.Write(emptyTree[i-1])
		emptyTree[i] = hsh.Sum(nil)
		emptyTree[i] = append(emptyTree[i], emptyTree[i]...)
	}

	// pad the merkle roots
	if len(r.Aggregation.L2MsgRootHashes) > config.L2MessageMaxNbMerkle {
		err = errors.New("more merkle trees than there is capacity")
		return
	}

	{
		roots := internal.CloneSlice(r.Aggregation.L2MsgRootHashes, config.L2MessageMaxNbMerkle)
		emptyRootHex := utils.HexEncodeToString(emptyTree[len(emptyTree)-1][:32])

		for i := len(r.Aggregation.L2MsgRootHashes); i < config.L2MessageMaxNbMerkle; i++ {
			for depth := config.L2MsgMerkleDepth; depth > 0; depth-- {
				for j := 0; j < 1<<(depth-1); j++ {
					hshK.Skip(emptyTree[config.L2MsgMerkleDepth-depth])
				}
			}
			roots = append(roots, emptyRootHex)
		}

		aggrPi := r.Aggregation
		aggrPi.L2MsgRootHashes = roots
		a.AggregationPublicInput = aggrPi.Sum(&hshK)
	}

	a.Keccak, err = hshK.Assign()

	return
}

// MerkleRoot computes the merkle root of data using the given hasher.
// TODO modify aggregation.PackInMiniTrees to optionally take a hasher instead of reimplementing
func MerkleRoot(hsh hash.Hash, treeNbLeaves int, data [][32]byte) [32]byte {
	if len(data) == 0 || len(data) > treeNbLeaves {
		panic("unacceptable tree size")
	}

	// duplicate; pad if necessary
	b := make([][32]byte, treeNbLeaves)
	copy(b, data)

	for len(b) != 1 {
		n := len(b) / 2
		for i := 0; i < n; i++ {
			hsh.Reset()
			hsh.Write(b[2*i][:])
			hsh.Write(b[2*i+1][:])
			copy(b[i][:], hsh.Sum(nil))
		}
		b = b[:n]
	}

	return b[0]
}
