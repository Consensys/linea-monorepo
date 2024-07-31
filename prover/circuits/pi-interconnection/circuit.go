package pi_interconnection

import (
	"errors"
	"math/big"
	"slices"

	"github.com/consensys/gnark/frontend"
	"github.com/consensys/gnark/std/compress"
	"github.com/consensys/gnark/std/hash"
	"github.com/consensys/gnark/std/hash/mimc"
	"github.com/consensys/gnark/std/lookup/logderivlookup"
	"github.com/consensys/gnark/std/math/cmp"
	"github.com/consensys/zkevm-monorepo/prover/circuits/aggregation"
	decompression "github.com/consensys/zkevm-monorepo/prover/circuits/blobdecompression/v1"
	"github.com/consensys/zkevm-monorepo/prover/circuits/execution"
	"github.com/consensys/zkevm-monorepo/prover/circuits/internal"
	"github.com/consensys/zkevm-monorepo/prover/circuits/pi-interconnection/keccak"
	"github.com/consensys/zkevm-monorepo/prover/crypto/mimc/gkrmimc"
	"github.com/consensys/zkevm-monorepo/prover/protocol/wizard"
	"github.com/consensys/zkevm-monorepo/prover/utils"
)

type Circuit struct {
	AggregationPublicInput   frontend.Variable   `gnark:",public"` // the public input of the aggregation circuit
	DecompressionPublicInput []frontend.Variable `gnark:",public"`
	ExecutionPublicInput     []frontend.Variable `gnark:",public"`

	DecompressionFPIQ []decompression.FunctionalPublicInputQSnark
	ExecutionFPIQ     []execution.FunctionalPublicInputQSnark
	aggregation.FunctionalPublicInputQSnark

	Keccak keccak.StrictHasherCircuit

	// config
	L2MessageMerkleDepth int
	L2MessageMaxNbMerkle int

	MaxNbCircuits int
	UseGkrMimc    bool
}

func (c *Circuit) Define(api frontend.API) error {
	maxNbDecompression, maxNbExecution := len(c.DecompressionPublicInput), len(c.ExecutionPublicInput)
	if len(c.DecompressionFPIQ) != maxNbDecompression || len(c.ExecutionFPIQ) != maxNbExecution {
		return errors.New("public / functional public input length mismatch")
	}

	c.FunctionalPublicInputQSnark.RangeCheck(api)

	rDecompression := internal.NewRange(api, c.NbDecompression, len(c.DecompressionPublicInput))
	hshK := c.Keccak.NewHasher(api)

	// equivalently, nbBatchesSums[i] is the index of the first execution circuit associated with the i+1-st decompression circuit
	nbBatchesSums := rDecompression.PartialSumsF(func(i int) frontend.Variable { return c.DecompressionFPIQ[i].NbBatches })
	nbExecution := nbBatchesSums[len(nbBatchesSums)-1]

	if c.MaxNbCircuits > 0 {
		api.AssertIsLessOrEqual(api.Add(nbExecution, c.NbDecompression), c.MaxNbCircuits)
	}

	var (
		hshM hash.FieldHasher
	)
	if c.UseGkrMimc {
		hsh := gkrmimc.NewHasherFactory(api).NewHasher()
		hshM = &hsh
	} else {
		if hsh, err := mimc.NewMiMC(api); err != nil {
			return err
		} else {
			hshM = &hsh
		}
	}

	batchHashes := make([]frontend.Variable, len(c.ExecutionPublicInput))
	for i, pi := range c.ExecutionFPIQ {
		batchHashes[i] = pi.DataChecksum
	}

	finalStateRootHashes := logderivlookup.New(api)
	for _, pi := range c.ExecutionFPIQ {
		finalStateRootHashes.Insert(pi.FinalStateRootHash)
	}

	blobBatchHashes := internal.ChecksumSubSlices(api, hshM, batchHashes, internal.VarSlice{Values: nbBatchesSums, Length: c.NbDecompression})

	shnarfParams := make([]ShnarfIteration, len(c.DecompressionPublicInput))
	for i, piq := range c.DecompressionFPIQ {
		piq.RangeCheck(api)

		shnarfParams[i] = ShnarfIteration{ // prepare shnarf verification data
			BlobDataSnarkHash:    internal.ToBytes(api, piq.SnarkHash),
			NewStateRootHash:     internal.ToBytes(api, finalStateRootHashes.Lookup(api.Sub(nbBatchesSums[i], 1))[0]),
			EvaluationPointBytes: piq.X,
			EvaluationClaimBytes: fr377EncodedFr381ToBytes(api, piq.Y),
		}

		rDecompression.AssertEqualI(i, c.DecompressionPublicInput[i],
			piq.Sum(api, hshM, blobBatchHashes[i])) // "open" decompression circuit public input
	}

	shnarfs := ComputeShnarfs(&hshK, c.ParentShnarf, shnarfParams)

	initBlockNum, initHashNum, initHash := c.InitialBlockNumber, c.InitialRollingHashNumber, c.InitialRollingHash
	initBlockTime, initState := c.InitialBlockTimestamp, c.InitialStateRootHash
	var l2MessagesByByte [32][]internal.VarSlice

	execMaxNbL2Msg := len(c.ExecutionFPIQ[0].L2MessageHashes.Values)
	merkleNbLeaves := 1 << c.L2MessageMerkleDepth
	for j := range l2MessagesByByte {
		l2MessagesByByte[j] = make([]internal.VarSlice, maxNbExecution)
		for k := range l2MessagesByByte[j] {
			l2MessagesByByte[j][k] = internal.VarSlice{Values: make([]frontend.Variable, execMaxNbL2Msg)}
		}
	}

	comparator := cmp.NewBoundedComparator(api, new(big.Int).Lsh(big.NewInt(1), 64), false) // TODO does the "false" mean that the deltas are range checked?
	// TODO try using lookups or crumb decomposition to make comparisons more efficient
	for i, piq := range c.ExecutionFPIQ {
		piq.RangeCheck(api)

		comparator.IsLess(initBlockTime, piq.FinalBlockTimestamp)
		comparator.IsLess(initBlockNum, piq.FinalBlockNumber)
		comparator.IsLess(initHashNum, piq.FinalRollingHashNumber)

		pi := execution.FunctionalPublicInputSnark{
			FunctionalPublicInputQSnark: piq,
			InitialStateRootHash:        initState,
			InitialBlockNumber:          initBlockNum,
			InitialBlockTimestamp:       initBlockTime,
			InitialRollingHash:          initHash,
			InitialRollingHashNumber:    initHashNum,
			ChainID:                     c.ChainID,
			L2MessageServiceAddr:        c.L2MessageServiceAddr,
		}
		initBlockNum, initHashNum, initHash = pi.FinalBlockNumber, pi.FinalRollingHashNumber, pi.FinalRollingHash
		initBlockTime, initState = pi.FinalBlockTimestamp, pi.FinalStateRootHash

		api.AssertIsEqual(c.ExecutionPublicInput[i], pi.Sum(api, hshM)) // "open" execution circuit public input

		if len(pi.L2MessageHashes.Values) != execMaxNbL2Msg {
			return errors.New("number of L2 messages must be the same for all executions")
		}

		// "transpose" the L2 messages by byte for Concat -> Merkle
		for j := range l2MessagesByByte { // perf-TODO probably better to change all 32bytes into four uint64s instead.
			for k := range pi.L2MessageHashes.Values {
				l2MessagesByByte[j][i].Values[k] = pi.L2MessageHashes.Values[k][j]
			}
			l2MessagesByByte[j][i].Length = pi.L2MessageHashes.Length
		}
	}

	merkleLeavesConcat := internal.Slice[[32]frontend.Variable]{Values: make([][32]frontend.Variable, c.L2MessageMaxNbMerkle*merkleNbLeaves)}
	for i := 0; i < 32; i++ {
		ithBytes := internal.Concat(api, len(merkleLeavesConcat.Values), l2MessagesByByte[i]...)
		for j := range merkleLeavesConcat.Values {
			merkleLeavesConcat.Values[j][i] = ithBytes.Values[j]
		}
		merkleLeavesConcat.Length = ithBytes.Length // same value regardless of i
	}
	rExecution := internal.NewRange(api, nbExecution, maxNbExecution)

	pi := aggregation.FunctionalPublicInputSnark{
		FunctionalPublicInputQSnark: c.FunctionalPublicInputQSnark,
		NbL2Messages:                merkleLeavesConcat.Length,
		L2MsgMerkleTreeRoots:        make([][32]frontend.Variable, c.L2MessageMaxNbMerkle),
		FinalBlockNumber:            rExecution.LastF(func(i int) frontend.Variable { return c.ExecutionFPIQ[i].FinalBlockNumber }),
		FinalBlockTimestamp:         rExecution.LastF(func(i int) frontend.Variable { return c.ExecutionFPIQ[i].FinalBlockTimestamp }),
		FinalRollingHash:            rExecution.LastArray32F(func(i int) [32]frontend.Variable { return c.ExecutionFPIQ[i].FinalRollingHash }),
		FinalRollingHashNumber:      rExecution.LastF(func(i int) frontend.Variable { return c.ExecutionFPIQ[i].FinalRollingHashNumber }),
		FinalShnarf:                 rDecompression.LastArray32(shnarfs),
		L2MsgMerkleTreeDepth:        c.L2MessageMerkleDepth,
	}

	for i := range pi.L2MsgMerkleTreeRoots {
		pi.L2MsgMerkleTreeRoots[i] = merkleRoot(&hshK, merkleLeavesConcat.Values[i*merkleNbLeaves:(i+1)*merkleNbLeaves])
	}

	// "open" aggregation public input
	aggregationPIBytes := pi.Sum(api, &hshK)
	api.AssertIsEqual(c.AggregationPublicInput, compress.ReadNum(api, aggregationPIBytes[:], big.NewInt(256)))

	return hshK.Finalize()
}

func merkleRoot(hshK keccak.BlockHasher, leaves [][32]frontend.Variable) [32]frontend.Variable {

	values := slices.Clone(leaves)
	if !utils.IsPowerOfTwo(len(values)) {
		panic("number of leaves must be a perfect power of two")
	}
	for len(values) > 1 {
		for i := 0; i < len(values)/2; i++ {
			values[i] = hshK.Sum(nil, values[i*2], values[i*2+1])
		}
		values = values[:len(values)/2]
	}

	return values[0]
}

type Config struct {
	MaxNbDecompression   int
	MaxNbExecution       int
	MaxNbCircuits        int
	MaxNbKeccakF         int
	MaxNbMsgPerExecution int
	L2MsgMerkleDepth     int
	L2MessageMaxNbMerkle int // if not explicitly provided (i.e. non-positive) it will be set to maximum
}

type Compiled struct {
	Circuit *Circuit
	Keccak  keccak.CompiledStrictHasher
}

func (c Config) Compile(wizardCompilationOpts ...func(iop *wizard.CompiledIOP)) (*Compiled, error) {

	if c.L2MessageMaxNbMerkle <= 0 {
		merkleNbLeaves := 1 << c.L2MsgMerkleDepth
		c.L2MessageMaxNbMerkle = (c.MaxNbExecution*c.MaxNbMsgPerExecution + merkleNbLeaves - 1) / merkleNbLeaves
	}

	sh := c.newKeccakCompiler().Compile(c.MaxNbKeccakF, wizardCompilationOpts...)
	shc, err := sh.GetCircuit()
	if err != nil {
		return nil, err
	}

	circuit := c.allocateCircuit()
	circuit.Keccak = shc
	for i := range circuit.ExecutionFPIQ {
		circuit.ExecutionFPIQ[i].L2MessageHashes.Values = make([][32]frontend.Variable, c.MaxNbMsgPerExecution)
	}

	return &Compiled{
		Circuit: &circuit,
		Keccak:  sh,
	}, nil
}

func (c *Compiled) getConfig() Config {
	return Config{
		MaxNbDecompression:   len(c.Circuit.DecompressionFPIQ),
		MaxNbExecution:       len(c.Circuit.ExecutionFPIQ),
		MaxNbKeccakF:         c.Keccak.MaxNbKeccakF(),
		MaxNbMsgPerExecution: len(c.Circuit.ExecutionFPIQ[0].L2MessageHashes.Values),
		L2MsgMerkleDepth:     c.Circuit.L2MessageMerkleDepth,
		L2MessageMaxNbMerkle: c.Circuit.L2MessageMaxNbMerkle,
		MaxNbCircuits:        c.Circuit.MaxNbCircuits,
	}
}

func (c Config) allocateCircuit() Circuit {
	return Circuit{
		DecompressionPublicInput: make([]frontend.Variable, c.MaxNbDecompression),
		ExecutionPublicInput:     make([]frontend.Variable, c.MaxNbExecution),
		DecompressionFPIQ:        make([]decompression.FunctionalPublicInputQSnark, c.MaxNbDecompression),
		ExecutionFPIQ:            make([]execution.FunctionalPublicInputQSnark, c.MaxNbExecution),
		L2MessageMerkleDepth:     c.L2MsgMerkleDepth,
		L2MessageMaxNbMerkle:     c.L2MessageMaxNbMerkle,
		MaxNbCircuits:            c.MaxNbCircuits,
	}
}

func (c Config) newKeccakCompiler() *keccak.StrictHasherCompiler {
	nbShnarf := c.MaxNbDecompression
	nbMerkle := c.L2MessageMaxNbMerkle * ((1 << c.L2MsgMerkleDepth) - 1)
	res := keccak.NewStrictHasherCompiler(nbShnarf, nbMerkle, 2)
	for i := 0; i < nbShnarf; i++ {
		res.WithHashLengths(160) // 5 components in every shnarf
	}

	for i := 0; i < nbMerkle; i++ {
		res.WithHashLengths(64) // 2 tree nodes
	}

	// aggregation PI opening
	res.WithHashLengths(32 * c.L2MessageMaxNbMerkle)
	res.WithHashLengths(384)

	return &res
}
