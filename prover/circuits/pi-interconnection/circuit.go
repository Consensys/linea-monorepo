package pi_interconnection

import (
	"errors"
	"fmt"
	"math/big"
	"slices"

	"github.com/consensys/gnark-crypto/ecc"
	"github.com/consensys/gnark/constraint"
	"github.com/consensys/gnark/frontend/cs/scs"
	"github.com/consensys/linea-monorepo/prover/circuits"
	"github.com/consensys/linea-monorepo/prover/config"
	public_input "github.com/consensys/linea-monorepo/prover/public-input"

	"github.com/consensys/gnark/frontend"
	"github.com/consensys/gnark/std/compress"
	"github.com/consensys/gnark/std/hash"
	"github.com/consensys/gnark/std/hash/mimc"
	"github.com/consensys/gnark/std/lookup/logderivlookup"
	"github.com/consensys/gnark/std/math/cmp"
	decompression "github.com/consensys/linea-monorepo/prover/circuits/blobdecompression/v1"
	"github.com/consensys/linea-monorepo/prover/circuits/execution"
	"github.com/consensys/linea-monorepo/prover/circuits/internal"
	"github.com/consensys/linea-monorepo/prover/circuits/pi-interconnection/keccak"
	"github.com/consensys/linea-monorepo/prover/crypto/mimc/gkrmimc"
	"github.com/consensys/linea-monorepo/prover/crypto/ringsis"
	"github.com/consensys/linea-monorepo/prover/protocol/compiler"
	"github.com/consensys/linea-monorepo/prover/protocol/compiler/cleanup"
	mimcComp "github.com/consensys/linea-monorepo/prover/protocol/compiler/mimc"
	"github.com/consensys/linea-monorepo/prover/protocol/compiler/selfrecursion"
	"github.com/consensys/linea-monorepo/prover/protocol/compiler/vortex"
	"github.com/consensys/linea-monorepo/prover/protocol/wizard"
	"github.com/consensys/linea-monorepo/prover/utils"
)

type Circuit struct {
	AggregationPublicInput   [2]frontend.Variable `gnark:",public"` // the public input of the aggregation circuit; divided big-endian into two 16-byte chunks
	ExecutionPublicInput     []frontend.Variable  `gnark:",public"`
	DecompressionPublicInput []frontend.Variable  `gnark:",public"`

	DecompressionFPIQ []decompression.FunctionalPublicInputQSnark
	ExecutionFPIQ     []execution.FunctionalPublicInputQSnark

	public_input.AggregationFPIQSnark

	Keccak keccak.StrictHasherCircuit

	// config
	L2MessageMerkleDepth int
	L2MessageMaxNbMerkle int

	MaxNbCircuits    int // possibly useless TODO consider removing
	UseGkrMimc       bool
	MockKeccakWizard bool // for testing purposes, bypass expensive keccak verification
}

// type alias to denote a wizard-compilation suite. This is used when calling
// compile and provides internal parameters for the wizard package.
type compilationSuite = []func(*wizard.CompiledIOP)

func (c *Circuit) Define(api frontend.API) error {
	maxNbDecompression, maxNbExecution := len(c.DecompressionPublicInput), len(c.ExecutionPublicInput)
	if len(c.DecompressionFPIQ) != maxNbDecompression || len(c.ExecutionFPIQ) != maxNbExecution {
		return errors.New("public / functional public input length mismatch")
	}

	c.AggregationFPIQSnark.RangeCheck(api)

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
	finalStateRootHashes.Insert(c.InitialStateRootHash)
	for _, pi := range c.ExecutionFPIQ {
		finalStateRootHashes.Insert(pi.FinalStateRootHash)
	}

	blobBatchHashes := internal.ChecksumSubSlices(api, hshM, batchHashes, internal.VarSlice{Values: nbBatchesSums, Length: c.NbDecompression})

	shnarfParams := make([]ShnarfIteration, len(c.DecompressionPublicInput))
	for i, piq := range c.DecompressionFPIQ {
		piq.RangeCheck(api)

		shnarfParams[i] = ShnarfIteration{ // prepare shnarf verification data
			BlobDataSnarkHash:    utils.ToBytes(api, piq.SnarkHash),
			NewStateRootHash:     utils.ToBytes(api, finalStateRootHashes.Lookup(nbBatchesSums[i])[0]),
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

	pi := public_input.AggregationFPISnark{
		AggregationFPIQSnark:   c.AggregationFPIQSnark,
		NbL2Messages:           merkleLeavesConcat.Length,
		L2MsgMerkleTreeRoots:   make([][32]frontend.Variable, c.L2MessageMaxNbMerkle),
		FinalBlockNumber:       rExecution.LastF(func(i int) frontend.Variable { return c.ExecutionFPIQ[i].FinalBlockNumber }),
		FinalBlockTimestamp:    rExecution.LastF(func(i int) frontend.Variable { return c.ExecutionFPIQ[i].FinalBlockTimestamp }),
		FinalRollingHash:       rExecution.LastArray32F(func(i int) [32]frontend.Variable { return c.ExecutionFPIQ[i].FinalRollingHash }),
		FinalRollingHashNumber: rExecution.LastF(func(i int) frontend.Variable { return c.ExecutionFPIQ[i].FinalRollingHashNumber }),
		FinalShnarf:            rDecompression.LastArray32(shnarfs),
		L2MsgMerkleTreeDepth:   c.L2MessageMerkleDepth,
	}

	for i := range pi.L2MsgMerkleTreeRoots {
		pi.L2MsgMerkleTreeRoots[i] = MerkleRootSnark(&hshK, merkleLeavesConcat.Values[i*merkleNbLeaves:(i+1)*merkleNbLeaves])
	}

	// "open" aggregation public input
	aggregationPIBytes := pi.Sum(api, &hshK)
	twoPow8 := big.NewInt(256)
	api.AssertIsEqual(c.AggregationPublicInput[0], compress.ReadNum(api, aggregationPIBytes[:16], twoPow8))
	api.AssertIsEqual(c.AggregationPublicInput[1], compress.ReadNum(api, aggregationPIBytes[16:], twoPow8))

	if c.MockKeccakWizard {
		return nil
	} else {
		return hshK.Finalize()
	}
}

func MerkleRootSnark(hshK keccak.BlockHasher, leaves [][32]frontend.Variable) [32]frontend.Variable {

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

type Compiled struct {
	Circuit *Circuit
	Keccak  keccak.CompiledStrictHasher
}

func Compile(c config.PublicInput, wizardCompilationOpts ...func(iop *wizard.CompiledIOP)) (*Compiled, error) {

	if c.L2MsgMaxNbMerkle <= 0 {
		merkleNbLeaves := 1 << c.L2MsgMerkleDepth
		c.L2MsgMaxNbMerkle = (c.MaxNbExecution*c.ExecutionMaxNbMsg + merkleNbLeaves - 1) / merkleNbLeaves
	}

	sh := newKeccakCompiler(c).Compile(wizardCompilationOpts...)
	shc, err := sh.GetCircuit()
	if err != nil {
		return nil, err
	}

	circuit := allocateCircuit(c)
	circuit.Keccak = shc
	for i := range circuit.ExecutionFPIQ {
		circuit.ExecutionFPIQ[i].L2MessageHashes.Values = make([][32]frontend.Variable, c.ExecutionMaxNbMsg)
	}

	return &Compiled{
		Circuit: &circuit,
		Keccak:  sh,
	}, nil
}

func (c *Compiled) getConfig() (config.PublicInput, error) {
	executionNbMsg := 0
	execs := c.Circuit.ExecutionFPIQ
	if len(c.Circuit.ExecutionFPIQ) != 0 {
		executionNbMsg = len(execs[0].L2MessageHashes.Values)
		for i := range execs {
			if len(execs[i].L2MessageHashes.Values) != executionNbMsg {
				return config.PublicInput{}, errors.New("inconsistent max number of L2 message hashes")
			}
		}
	}
	return config.PublicInput{
		MaxNbDecompression: len(c.Circuit.DecompressionFPIQ),
		MaxNbExecution:     len(c.Circuit.ExecutionFPIQ),
		ExecutionMaxNbMsg:  executionNbMsg,
		L2MsgMerkleDepth:   c.Circuit.L2MessageMerkleDepth,
		L2MsgMaxNbMerkle:   c.Circuit.L2MessageMaxNbMerkle,
		MaxNbCircuits:      c.Circuit.MaxNbCircuits,
	}, nil
}

func allocateCircuit(c config.PublicInput) Circuit {
	return Circuit{
		DecompressionPublicInput: make([]frontend.Variable, c.MaxNbDecompression),
		ExecutionPublicInput:     make([]frontend.Variable, c.MaxNbExecution),
		DecompressionFPIQ:        make([]decompression.FunctionalPublicInputQSnark, c.MaxNbDecompression),
		ExecutionFPIQ:            make([]execution.FunctionalPublicInputQSnark, c.MaxNbExecution),
		L2MessageMerkleDepth:     c.L2MsgMerkleDepth,
		L2MessageMaxNbMerkle:     c.L2MsgMaxNbMerkle,
		MaxNbCircuits:            c.MaxNbCircuits,
		MockKeccakWizard:         c.MockKeccakWizard,
		UseGkrMimc:               true,
	}
}

func newKeccakCompiler(c config.PublicInput) *keccak.StrictHasherCompiler {
	nbShnarf := c.MaxNbDecompression
	nbMerkle := c.L2MsgMaxNbMerkle * ((1 << c.L2MsgMerkleDepth) - 1)
	res := keccak.NewStrictHasherCompiler(nbShnarf, nbMerkle, 2)
	for i := 0; i < nbShnarf; i++ {
		res.WithHashLengths(160) // 5 components in every shnarf
	}

	for i := 0; i < nbMerkle; i++ {
		res.WithHashLengths(64) // 2 tree nodes
	}

	// aggregation PI opening
	res.WithHashLengths(32 * c.L2MsgMaxNbMerkle)
	res.WithHashLengths(384)

	return &res
}

type builder struct {
	*config.PublicInput
}

func NewBuilder(c config.PublicInput) circuits.Builder {
	return builder{&c}
}

func (b builder) Compile() (constraint.ConstraintSystem, error) {
	c, err := Compile(*b.PublicInput, WizardCompilationParameters()...)
	if err != nil {
		return nil, err
	}
	const estimatedNbConstraints = 35_000_000
	cs, err := frontend.Compile(ecc.BLS12_377.ScalarField(), scs.NewBuilder, c.Circuit, frontend.WithCapacity(estimatedNbConstraints))
	if err != nil {
		return nil, err
	}
	if nbC := cs.GetNbConstraints(); nbC > estimatedNbConstraints || estimatedNbConstraints-nbC > 5_000_000 {
		return nil, fmt.Errorf("constraint estimate is off; got %d", nbC)
	}
	return cs, nil
}

func WizardCompilationParameters() []func(iop *wizard.CompiledIOP) {
	var (
		sisInstance = ringsis.Params{LogTwoBound: 16, LogTwoDegree: 6}

		fullCompilationSuite = compilationSuite{

			compiler.Arcane(1<<10, 1<<18, false),
			vortex.Compile(
				2,
				vortex.ForceNumOpenedColumns(256),
				vortex.WithSISParams(&sisInstance),
			),

			selfrecursion.SelfRecurse,
			cleanup.CleanUp,
			mimcComp.CompileMiMC,
			compiler.Arcane(1<<10, 1<<16, false),
			vortex.Compile(
				8,
				vortex.ForceNumOpenedColumns(64),
				vortex.WithSISParams(&sisInstance),
			),

			selfrecursion.SelfRecurse,
			cleanup.CleanUp,
			mimcComp.CompileMiMC,
			compiler.Arcane(1<<10, 1<<13, false),
			vortex.Compile(
				8,
				vortex.ForceNumOpenedColumns(64),
				vortex.ReplaceSisByMimc(),
			),
		}
	)

	return fullCompilationSuite

}

// GetMaxNbCircuitsSum computes MaxNbDecompression + MaxNbExecution from the compiled constraint system
// TODO replace with something cleaner, using the config
func GetMaxNbCircuitsSum(cs constraint.ConstraintSystem) int {
	return cs.GetNbPublicVariables() - 2
}

type InnerCircuitType uint8

const (
	Execution     InnerCircuitType = 0
	Decompression InnerCircuitType = 1
)

func InnerCircuitTypesToIndexes(cfg *config.PublicInput, types []InnerCircuitType) []int {
	indexes := utils.RightPad(utils.Partition(utils.RangeSlice[int](len(types)), types), 2)
	return utils.RightPad(
		append(utils.RightPad(indexes[Execution], cfg.MaxNbExecution), indexes[Decompression]...), cfg.MaxNbExecution+cfg.MaxNbDecompression)

}
