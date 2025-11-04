package iokeccakf

/* io-keccakf package prepare the input and output for keccakf.
The inputs to keccakf are blocks in BaseA or BaseB (little-endian).
The output from keccakf is the hash result in baseB (little-endian).
 Thus, the implementation applies a base conversion over the blocks;
going from uint-BE to BaseA/BaseB-LE. Also, a base conversion over the hash result,
going from BaseB-LE to uint-BE. */

import (
	"slices"

	"github.com/consensys/linea-monorepo/prover/maths/common/smartvectors"
	"github.com/consensys/linea-monorepo/prover/maths/common/vector"
	"github.com/consensys/linea-monorepo/prover/maths/field"
	"github.com/consensys/linea-monorepo/prover/protocol/column"
	"github.com/consensys/linea-monorepo/prover/protocol/distributed/pragmas"
	"github.com/consensys/linea-monorepo/prover/protocol/ifaces"
	"github.com/consensys/linea-monorepo/prover/protocol/wizard"
	sym "github.com/consensys/linea-monorepo/prover/symbolic"
	"github.com/consensys/linea-monorepo/prover/utils"
	"github.com/consensys/linea-monorepo/prover/zkevm/prover/common"
	commonconstraints "github.com/consensys/linea-monorepo/prover/zkevm/prover/common/common_constraints"
	"github.com/consensys/linea-monorepo/prover/zkevm/prover/hash/generic"
)

const (
	BASE_CONVERSION = "BASE_CONVERSION"
	BLOCK           = "BLOCK"
	HASH_OUTPUT     = "HASH_OUTPUT"
	MAXNBYTE        = 2 // maximum number of bytes due to the field size.
)

type lookup struct {
	colMAXNBYTE ifaces.Column
	colBaseA    []ifaces.Column
	colBaseB    []ifaces.Column
}

// KeccakfBlockPreparationInputs stores the inputs for [newBlockBaseConversion]
type KeccakfBlockPreparationInputs struct {
	Lane ifaces.Column
	// IsBeginningOfNewHash is 1 if the associated row is the beginning of a new hash.
	IsBeginningOfNewHash ifaces.Column
	IsLaneActive         ifaces.Column
	// bases for conversion, note that baseA^8 and BaseB^8 should be less than the field modulus.
	BaseA, BaseB int
	// number of bits per baseX (baseA or baseB)
	// the constraint is that BaseX^nbBitPerBaseX < field.Modulus and nbBitPerBaseX is a power of 2.
	NbBitsPerBaseX int
}

// The submodule KeccakfBlockPreparation implements the base conversion over
// the inputs to the keccakf (i.e., blocks/lanes), in order to export them to the keccakf.
// The lanes from the first block of hash should be in baseA and others are in baseB.
type KeccakfBlockPreparation struct {
	Inputs *KeccakfBlockPreparationInputs
	// It is 1 when the lane is from the first block of the hash
	IsFromFirstBlock ifaces.Column
	// IsFromBlockBaseB := 1-isFromFirstBlock
	IsFromBlockBaseB ifaces.Column
	// lanes from first block in baseA, others in baseB
	LaneX []ifaces.Column
	// Size of the module
	Size int
	// lookup table for base conversion
	// the first column in in base uint16, the second in baseA, the third in baseB
	Lookup lookup
}

// NewKeccakfBlockPreparation declare the intermediate columns,
// and the constraints for changing the blocks in base uint64 into baseA/baseB.
// It also change the order of Bytes from Big-Endian to Little-Endian.
func NewKeccakfBlockPreparation(
	comp *wizard.CompiledIOP,
	inp KeccakfBlockPreparationInputs,
) *KeccakfBlockPreparation {

	var (
		nbSlicesBaseX = MAXNBYTE * 8 / inp.NbBitsPerBaseX

		b = &KeccakfBlockPreparation{
			Inputs: &inp,
			Size:   inp.Lane.Size(),
			Lookup: lookup{
				colBaseA: make([]ifaces.Column, nbSlicesBaseX),
				colBaseB: make([]ifaces.Column, nbSlicesBaseX),
			},
			LaneX: make([]ifaces.Column, nbSlicesBaseX),
		}
		// declare the columns
		createCol = common.CreateColFn(comp, BASE_CONVERSION, b.Size, pragmas.RightPadded)

		param            = generic.KeccakUsecase
		nbOfRowsPerLane  = param.LaneSizeBytes() / MAXNBYTE
		nbOfRowsPerBlock = param.NbOfLanesPerBlock() * nbOfRowsPerLane
	)
	for j := 0; j < nbSlicesBaseX; j++ {
		b.LaneX[j] = createCol("LaneX_%v", j)
	}
	b.IsFromFirstBlock = createCol("IsFromFirstBlock")
	b.IsFromBlockBaseB = createCol("IsFromBlockBaseB")

	// table for base conversion (used for converting blocks to what keccakf expect)
	colMAXNBYTE, colBaseA, colBaseB := createLookupTablesBaseX(b.Inputs.BaseB, b.Inputs.BaseA, nbSlicesBaseX)

	b.Lookup.colMAXNBYTE = comp.InsertPrecomputed(ifaces.ColIDf("LOOKUP_MAXNBYTE"), colMAXNBYTE)
	for i := 0; i < nbSlicesBaseX; i++ {
		b.Lookup.colBaseA[i] = comp.InsertPrecomputed(ifaces.ColIDf("LOOKUP_BaseA_%v", i), colBaseA[i])
		b.Lookup.colBaseB[i] = comp.InsertPrecomputed(ifaces.ColIDf("LOOKUP_BaseB_%v", i), colBaseB[i])
	}

	// declare the constraints
	// isFromFirstBlock is well formed
	// isFromFirstBlock = sum_j Shift(l.isFirstLaneFromNewHash,-j) for j:=0,...,
	s := sym.NewConstant(0)
	for j := 0; j < nbOfRowsPerBlock; j++ {
		s = sym.Add(
			s, column.Shift(b.Inputs.IsBeginningOfNewHash, -j),
		)
	}
	comp.InsertGlobal(0, ifaces.QueryIDf("IsFromFirstBlock"),
		sym.Sub(s, b.IsFromFirstBlock))

	commonconstraints.MustBeMutuallyExclusiveBinaryFlags(
		comp,
		b.Inputs.IsLaneActive,
		[]ifaces.Column{
			b.IsFromFirstBlock,
			b.IsFromBlockBaseB},
	)

	// if isFromFirstBlock = 1  ---> convert to keccak.BaseA
	// otherwise convert to keccak.BaseB
	/*comp.InsertInclusionConditionalOnIncluded(0,
		ifaces.QueryIDf("BaseConversion_Into_BaseA_"),
		append(b.Lookup.colBaseA, b.Lookup.colMAXNBYTE),
		append(b.LaneX, b.Inputs.Lane),
		b.IsFromFirstBlock)

	comp.InsertInclusionConditionalOnIncluded(0,
	ifaces.QueryIDf("BaseConversion_Into_BaseB_"),
	append(b.Lookup.colBaseB, b.Lookup.colMAXNBYTE),
	append(b.LaneX, b.Inputs.Lane),
	b.IsFromBlockBaseB)
	*/

	return b
}

// assign column isFromFirstBlock and laneX
func (b *KeccakfBlockPreparation) Run(run *wizard.ProverRuntime) {

	var (
		size                 = b.Size
		isBeginningOfNewHash = b.Inputs.IsBeginningOfNewHash.GetColAssignment(run).IntoRegVecSaveAlloc()
		param                = generic.KeccakUsecase
		nbOfRowsPerLane      = param.LaneSizeBytes() / MAXNBYTE
		numRowsPerBlock      = param.NbOfLanesPerBlock() * nbOfRowsPerLane
		isActive             = b.Inputs.IsLaneActive.GetColAssignment(run).IntoRegVecSaveAlloc()
		colIsFromFirstBlock  = common.NewVectorBuilder(b.IsFromFirstBlock)
		colIsFromOtherBlocks = common.NewVectorBuilder(b.IsFromBlockBaseB)
		ones                 = vector.Repeat(field.One(), numRowsPerBlock)
	)

	// assign isFromFirstBlock
	for j := 0; j < size; j++ {
		if isBeginningOfNewHash[j].IsOne() {
			colIsFromFirstBlock.PushSliceF(ones)
			j = j + (numRowsPerBlock - 1)
		} else {
			colIsFromFirstBlock.PushInt(0)
		}
	}

	isNotFirstBlock := make([]field.Element, size)
	vector.Sub(isNotFirstBlock, isActive, colIsFromFirstBlock.Slice())
	colIsFromOtherBlocks.PushSliceF(isNotFirstBlock)

	colIsFromFirstBlock.PadAndAssign(run, field.Zero())
	colIsFromOtherBlocks.PadAndAssign(run, field.Zero())

	// assign the laneX
	var (
		isFirstBlock = b.IsFromFirstBlock.GetColAssignment(run).IntoRegVecSaveAlloc()
		lane         = b.Inputs.Lane.GetColAssignment(run).IntoRegVecSaveAlloc()
		laneX        = make([]*common.VectorBuilder, len(b.LaneX))
		bytes        = make([]byte, MAXNBYTE)
	)

	// initialize laneX builders
	for j := range laneX {
		laneX[j] = common.NewVectorBuilder(b.LaneX[j])
	}

	// populate laneX
	for j := range lane {
		laneBytes := lane[j].Bytes() // big-endian bytes
		bytes = laneBytes[len(laneBytes)-MAXNBYTE:]
		slices.Reverse(bytes) // to have little-endian order

		// convert lane[j] into uint-LE.
		if isFirstBlock[j].IsOne() {
			a := extractLittleEndianBaseX(bytes, b.Inputs.NbBitsPerBaseX, len(b.LaneX), b.Inputs.BaseA)
			for k := range laneX {
				laneX[k].PushInt(int(a[k]))
			}
		} else {
			a := extractLittleEndianBaseX(bytes, b.Inputs.NbBitsPerBaseX, len(b.LaneX), b.Inputs.BaseB)
			for k := range laneX {
				laneX[k].PushInt(int(a[k]))
			}
		}
	}

	// assign the laneX
	for j := range laneX {
		laneX[j].PadAndAssign(run, field.Zero())
	}
}

// x has nbBits bits, convert it to baseX representation
func uintToBaseX(x uint, base field.Element, nbBits int) field.Element {

	res := field.Zero()
	one := field.One()
	resIsZero := true

	for k := nbBits - 1; k >= 0; k-- {
		// The test allows skipping useless field muls or testing
		// the entire field element.
		if !resIsZero {
			res.Mul(&res, &base)
		}

		// Skips the field addition if the bit is zero
		bit := (x >> k) & 1
		if bit > 0 {
			res.Add(&res, &one)
			resIsZero = false
		}
	}

	return res
}

// / createLookupTablesBaseX constructs lookup tables for representing every possible
// MAXNBYTE-byte integer (from 0 to 2^(8*MAXNBYTE) - 1) in `nbSlices` slices,
// where each slice corresponds to a fixed-width segment of `bitsPerSlice` bits.
//
// For each integer j in [0, 2^(8*MAXNBYTE) - 1], the function decomposes j into
// `nbSlices` chunks of size `bitsPerSlice`, least-significant slice first, then
// encodes each chunk both in base `baseA` and in base `baseB` using `uintToBaseX`.
//
// Parameters:
//   - baseA, baseB: field elements representing the numeric bases for two different encodings.
//   - nbSlices:     number of slices into which each integer is decomposed.
//
// Returns three columns of lookup tables:
//   - uint16Col: SmartVector of all possible field elements j ∈ [0, 2^(8*MAXNBYTE) - 1].
//   - baseACol:  baseACol[i][j] holds the i-th slice of j encoded in baseA.
//   - baseBCol:  baseBCol[i][j] holds the i-th slice of j encoded in baseB.
//
// Example (conceptual):
//
//	Suppose MAXNBYTE = 2 and nbSlices = 2.
//
//	bitsPerSlice = (8 * 2) / 2 = 8
//	mask = 0xFF
//
//	For integer j = 0x1234 (4660 decimal):
//	    slices = [0x34, 0x12]
//	    baseACol[0][4660] = uintToBaseX(0x34, baseA, 8)
//	    baseACol[1][4660] = uintToBaseX(0x12, baseA, 8)
//
//	Thus each row j in the tables represents one full decomposition.
//
// Note:
//   - The decomposition is little-endian with respect to bit order: slice 0 contains the least significant bits.
//   - The total number of table entries grows as 2^(8*MAXNBYTE), which can be extremely large.
//     This function is only feasible for small MAXNBYTE (e.g., 1 or 2).
func createLookupTablesBaseX(baseA, baseB int, nbSlices int) (uint16Col smartvectors.SmartVector, baseACol, baseBCol []smartvectors.SmartVector) {
	var (
		bitsPerSlice = MAXNBYTE * 8 / nbSlices
		u            []field.Element
		v, w         = make([][]field.Element, nbSlices), make([][]field.Element, nbSlices)
		mask         = (1 << bitsPerSlice) - 1
		uintMAXNBYTE = 1<<(8*MAXNBYTE) - 1
		result       = make([]uint, nbSlices) // it holds the decomposition uint in chuncks of size bitsPerSlice
		baseAFr      = field.NewElement(uint64(baseA))
		baseBFr      = field.NewElement(uint64(baseB))
	)

	if MAXNBYTE > 2 {
		// too big to create a lookup table, in this case lane column needs to be decomposed into smaller chunks.
		utils.Panic("MAXNBYTE = %v create an unsupported size of lookuptable", MAXNBYTE)
	}

	// preallocate slices
	for i := 0; i < nbSlices; i++ {
		v[i] = make([]field.Element, 0, uintMAXNBYTE+1)
		w[i] = make([]field.Element, 0, uintMAXNBYTE+1)
	}

	for j := 0; j <= uintMAXNBYTE; j++ {

		// decompse j into chunks of size bitsPerSlice
		n := j
		for i := 0; i < nbSlices; i++ {
			result[i] = uint(n & mask) // extract lowest chunkBits bits
			n >>= bitsPerSlice         // shift n for next chunk
		}

		u = append(u, field.NewElement(uint64(j)))

		for i := 0; i < nbSlices; i++ {
			v[i] = append(v[i], uintToBaseX(result[i], baseAFr, bitsPerSlice))
			w[i] = append(w[i], uintToBaseX(result[i], baseBFr, bitsPerSlice))
		}
	}

	for j := 0; j < nbSlices; j++ {
		baseACol = append(baseACol, smartvectors.NewRegular(v[j]))
		baseBCol = append(baseBCol, smartvectors.NewRegular(w[j]))
	}

	return smartvectors.NewRegular(u), baseACol, baseBCol
}

// extractLittleEndianBaseX extracts `nbSlices` integers from the bitstream `data`,
// interpreting each chunk of `bitsPerChunk` bits as a number in the given `base`.
//
// Each chunk is read in *little-endian bit order.
//
// The function treats each bit as a binary digit (0 or 1), and accumulates its
// contribution in the specified base. In other words, for each chunk:
//
//	val = Σ_{j=0}^{bitsPerChunk-1} (bit_j * base^j)
//
// where bit_j is the j-th bit (starting from the least significant bit in the stream).
//
// Parameters:
//   - data:         the input byte slice containing the bitstream.
//   - bitsPerChunk: the number of bits to read for each integer.
//   - nbSlices:     how many integers to extract from the stream.
//   - base:         the base used for recombination (e.g., 2 for binary, 3 for base-3 interpretation).
//
// Returns:
//
//	A slice of uint64 values, each representing one extracted integer.
//
// Example:
//
//	data := []byte{0b11001010, 0b00110101} // bitstream (LSB-first)
//	bitsPerChunk := 4
//	nbSlices := 4
//	base := uint64(3)
//
//	vals := extractLittleEndianBaseX(data, bitsPerChunk, nbSlices, base)
//
// Combined bitstream (LSB-first):
//
//	[0 1 0 1 0 0 1 1 1 0 1 0 1 1 0 0]
//
// Grouped into 4-bit chunks (little-endian within each group):
//
//	chunk 0 → [0 1 0 1]
//	chunk 1 → [0 0 1 1]
//	chunk 2 → [1 0 1 0]
//	chunk 3 → [1 1 0 0]
//
// Interpretation in base 3 (each bit contributes bit_j * 3^j):
//
//	val[0] = 0*3^0 + 1*3^1 + 0*3^2 + 1*3^3 = 30
//	val[1] = 0*3^0 + 0*3^1 + 1*3^2 + 1*3^3 = 36
//	val[2] = 1*3^0 + 0*3^1 + 1*3^2 + 0*3^3 = 10
//	val[3] = 1*3^0 + 1*3^1 + 0*3^2 + 0*3^3 = 4
//
// So the function returns:
//
//	[]uint64{30, 36, 10, 4}
//
// For base = 2, this behaves exactly like a normal binary little-endian extractor.
func extractLittleEndianBaseX(data []byte, bitsPerChunk int, nbSlices int, base int) []int {
	result := make([]int, 0, nbSlices)
	var bitPos int // current bit offset in the stream
	totalBits := len(data) * 8

	for i := 0; i < nbSlices && bitPos+bitsPerChunk <= totalBits; i++ {
		var val int
		pow := 1

		for j := 0; j < bitsPerChunk; j++ {
			byteIndex := (bitPos + j) / 8
			bitIndex := (bitPos + j) % 8
			bit := (data[byteIndex] >> bitIndex) & 1 // little-endian bit order within each byte

			// accumulate in base `base`, little-endian order
			val += int(bit) * pow
			pow *= base
		}
		result = append(result, val)
		bitPos += bitsPerChunk
	}
	return result
}
