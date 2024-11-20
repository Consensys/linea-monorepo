package pi_interconnection

import (
	"bytes"
	"encoding/base64"
	"errors"
	"fmt"
	"github.com/consensys/linea-monorepo/prover/crypto/mimc"
	"hash"

	"github.com/consensys/gnark-crypto/ecc/bls12-381/fr"
	"github.com/consensys/linea-monorepo/prover/backend/blobsubmission"
	decompression "github.com/consensys/linea-monorepo/prover/circuits/blobdecompression/v1"
	"github.com/consensys/linea-monorepo/prover/circuits/execution"
	"github.com/consensys/linea-monorepo/prover/circuits/internal"
	"github.com/consensys/linea-monorepo/prover/circuits/pi-interconnection/keccak"
	"github.com/consensys/linea-monorepo/prover/lib/compressor/blob"
	public_input "github.com/consensys/linea-monorepo/prover/public-input"
	"github.com/consensys/linea-monorepo/prover/utils"
	"github.com/sirupsen/logrus"
	"golang.org/x/crypto/sha3"
)

type Request struct {
	Decompressions []blobsubmission.Response
	Executions     []public_input.Execution
	Aggregation    public_input.Aggregation
}

func (c *Compiled) Assign(r Request) (a Circuit, err error) {
	internal.RegisterHints()
	keccak.RegisterHints()
	utils.RegisterHints()

	// TODO there is data duplication in the request. Check consistency

	// infer config
	config, err := c.getConfig()
	if err != nil {
		return
	}
	a = allocateCircuit(config)

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
	utils.Copy(a.ParentShnarf[:], prevShnarf)

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
			err = fmt.Errorf("shnarf mismatch, i:%d, shnarf: %x, prevShnarf: %x, ", i, shnarfs[i], prevShnarf)
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
			OldShnarf: prevShnarf,
			SnarkHash: zero[:],
			X:         zero[:],
			Hash:      &hshK,
		}
		if len(r.Executions) == 0 { // edge case for integration testing
			if shnarf.NewStateRootHash, err = utils.HexDecodeString(r.Aggregation.ParentStateRootHash); err != nil {
				return
			}
		} else {
			shnarf.NewStateRootHash = r.Executions[len(execDataChecksums)-1].FinalStateRootHash[:]
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
		utils.Copy(a.DecompressionFPIQ[i].X[:], zero[:])

		a.DecompressionPublicInput[i] = 0
	}

	// Aggregation FPI
	aggregationFPI, err := public_input.NewAggregationFPI(&r.Aggregation)
	if err != nil {
		return
	}
	if len(r.Decompressions) != 0 && !bytes.Equal(shnarfs[len(r.Decompressions)-1], aggregationFPI.FinalShnarf[:]) { // first condition is an edge case for tests
		err = errors.New("mismatch between decompression/aggregation-supplied shnarfs")
		return
	}
	aggregationFPI.NbDecompression = uint64(len(r.Decompressions))
	a.AggregationFPIQSnark = aggregationFPI.ToSnarkType().AggregationFPIQSnark

	merkleNbLeaves := 1 << config.L2MsgMerkleDepth
	maxNbL2MessageHashes := config.L2MsgMaxNbMerkle * merkleNbLeaves
	l2MessageHashes := make([][32]byte, 0, maxNbL2MessageHashes)

	finalRollingHashNum, finalRollingHash := aggregationFPI.InitialRollingHashNumber, aggregationFPI.InitialRollingHash
	finalBlockTimestamp := aggregationFPI.LastFinalizedBlockTimestamp

	// Execution FPI
	executionFPI := execution.FunctionalPublicInput{
		FinalStateRootHash:   aggregationFPI.InitialStateRootHash,
		FinalBlockNumber:     aggregationFPI.LastFinalizedBlockNumber,
		FinalBlockTimestamp:  aggregationFPI.LastFinalizedBlockTimestamp,
		L2MessageServiceAddr: aggregationFPI.L2MessageServiceAddr,
		ChainID:              aggregationFPI.ChainID,
		MaxNbL2MessageHashes: config.ExecutionMaxNbMsg,
	}

	hshM := mimc.NewMiMC()
	for i := range a.ExecutionFPIQ {
		executionFPI.InitialRollingHash = [32]byte{}
		executionFPI.InitialRollingHashNumber = 0
		executionFPI.L2MessageHashes = nil

		// pad things correctly to make the circuit's life a bit easier
		executionFPI.InitialBlockNumber = executionFPI.FinalBlockNumber + 1
		executionFPI.InitialStateRootHash = executionFPI.FinalStateRootHash
		executionFPI.InitialBlockTimestamp = executionFPI.FinalBlockTimestamp + 1
		executionFPI.FinalBlockTimestamp = executionFPI.InitialBlockTimestamp

		executionFPI.L2MessageServiceAddr = r.Aggregation.L2MessageServiceAddr
		executionFPI.ChainID = r.Aggregation.ChainID

		a.ExecutionPublicInput[i] = 0 // unless...

		if i < len(r.Executions) {
			if initial, final := r.Executions[i].InitialBlockTimestamp, finalBlockTimestamp; initial <= final {
				err = fmt.Errorf("execution #%d. initial block timestamp is not after the final block timestamp %d≤%d", i, initial, final)
				return
			}
			if initial, final := r.Executions[i].InitialBlockTimestamp, r.Executions[i].FinalBlockTimestamp; initial > final {
				err = fmt.Errorf("execution #%d. initial block timestamp is after the final block timestamp %d>%d", i, initial, final)
				return
			}
			executionFPI.InitialBlockTimestamp = r.Executions[i].InitialBlockTimestamp
			executionFPI.FinalRollingHash = r.Executions[i].FinalRollingHash
			executionFPI.FinalBlockNumber = r.Executions[i].FinalBlockNumber
			executionFPI.FinalBlockTimestamp = r.Executions[i].FinalBlockTimestamp
			finalBlockTimestamp = r.Executions[i].FinalBlockTimestamp
			executionFPI.FinalRollingHash = r.Executions[i].FinalRollingHash
			executionFPI.FinalRollingHashNumber = r.Executions[i].FinalRollingHashNumber
			executionFPI.FinalStateRootHash = r.Executions[i].FinalStateRootHash

			copy(executionFPI.DataChecksum[:], execDataChecksums[i])
			executionFPI.L2MessageHashes = r.Executions[i].L2MsgHashes

			l2MessageHashes = append(l2MessageHashes, r.Executions[i].L2MsgHashes...)

			if got, want := &r.Executions[i].L2MessageServiceAddr, &r.Aggregation.L2MessageServiceAddr; *got != *want {
				err = fmt.Errorf("execution #%d. expected L2 service address %x, encountered %x", i, *want, *got)
				return
			}
			if got, want := executionFPI.ChainID, r.Aggregation.ChainID; got != want {
				err = fmt.Errorf("execution #%d. expected chain ID %x, encountered %x", i, want, got)
				return
			}

			if r.Executions[i].FinalRollingHashNumber != 0 { // if the rolling hash is being updated, record the change
				executionFPI.InitialRollingHash = finalRollingHash
				finalRollingHash = r.Executions[i].FinalRollingHash
				executionFPI.InitialRollingHashNumber = finalRollingHashNum
				finalRollingHashNum = r.Executions[i].FinalRollingHashNumber
			}

			a.ExecutionPublicInput[i] = executionFPI.Sum(hshM)
		}

		if snarkFPI, _err := executionFPI.ToSnarkType(); _err != nil {
			err = fmt.Errorf("execution #%d: %w", i, _err)
			return
		} else {
			a.ExecutionFPIQ[i] = snarkFPI.FunctionalPublicInputQSnark
		}
	}
	// consistency check
	if finalBlockTimestamp != aggregationFPI.FinalBlockTimestamp {
		err = fmt.Errorf("final block timestamps do not match: execution=%d, aggregation=%d",
			executionFPI.FinalBlockTimestamp, aggregationFPI.FinalBlockTimestamp)
		return
	}
	if executionFPI.FinalBlockNumber != aggregationFPI.FinalBlockNumber {
		err = fmt.Errorf("final block numbers do not match: execution=%d, aggregation=%d",
			executionFPI.FinalBlockNumber, aggregationFPI.FinalBlockNumber)
		return
	}

	if finalRollingHash != aggregationFPI.FinalRollingHash {
		err = fmt.Errorf("final rolling hashes do not match: execution=%x, aggregation=%x",
			executionFPI.FinalRollingHash, aggregationFPI.FinalRollingHash)
		return
	}

	if finalRollingHashNum != aggregationFPI.FinalRollingHashNumber {
		err = fmt.Errorf("final rolling hash numbers do not match: execution=%v, aggregation=%v",
			executionFPI.FinalRollingHashNumber, aggregationFPI.FinalRollingHashNumber)
		return
	}

	if len(l2MessageHashes) > maxNbL2MessageHashes {
		err = errors.New("too many L2 messages")
		return
	}

	if minNbRoots := (len(l2MessageHashes) + merkleNbLeaves - 1) / merkleNbLeaves; len(r.Aggregation.L2MsgRootHashes) < minNbRoots {
		err = fmt.Errorf("the %d merkle roots provided are too few to accommodate all %d execution messages. A minimum of %d is needed", len(r.Aggregation.L2MsgRootHashes), len(l2MessageHashes), minNbRoots)
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
	if len(r.Aggregation.L2MsgRootHashes) > config.L2MsgMaxNbMerkle {
		err = errors.New("more merkle trees than there is capacity")
		return
	}

	for i := len(r.Aggregation.L2MsgRootHashes); i < config.L2MsgMaxNbMerkle; i++ {
		for depth := config.L2MsgMerkleDepth; depth > 0; depth-- {
			for j := 0; j < 1<<(depth-1); j++ {
				hshK.Skip(emptyTree[config.L2MsgMerkleDepth-depth])
			}
		}
	}

	aggregationPI := r.Aggregation.Sum(&hshK)

	a.AggregationPublicInput[0] = aggregationPI[:16]
	a.AggregationPublicInput[1] = aggregationPI[16:]

	logrus.Infof("generating wizard proof for %d hashes from %d permutations", hshK.NbHashes(), hshK.MaxNbKeccakF())
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
