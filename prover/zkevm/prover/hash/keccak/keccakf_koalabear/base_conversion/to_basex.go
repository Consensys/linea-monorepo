package baseconversion

/* iokeccakf package prepares the input and output for keccakf.
The inputs to keccakf are blocks in BaseA or BaseB (little-endian).
The output from keccakf is the hash result in baseB (little-endian).
 Thus, the implementation applies a base conversion over the blocks;
going from uint to BaseA/BaseB. Also, a base conversion over the hash result,
going from BaseB to uint. */

import (
	"slices"

	"github.com/consensys/linea-monorepo/prover/maths/common/smartvectors"
	"github.com/consensys/linea-monorepo/prover/maths/field"
	"github.com/consensys/linea-monorepo/prover/protocol/distributed/pragmas"
	"github.com/consensys/linea-monorepo/prover/protocol/ifaces"
	"github.com/consensys/linea-monorepo/prover/protocol/wizard"
	"github.com/consensys/linea-monorepo/prover/utils"
	"github.com/consensys/linea-monorepo/prover/zkevm/prover/common"
)

const (
	MAXNBYTE                  = 2 // maximum number of bytes due to the field size.
	KECCAKF_INPUT_PREPARATION = "KECCAKF_INPUT_PREPARATION"
)

// lookup holds the lookup tables for base conversion.
type lookup struct {
	colMAXNBYTE ifaces.Column     // holds the uint values from 0 to 2^(8*MAXNBYTE)-1
	colBasex    [][]ifaces.Column // holds the baseA representation of uint values
}

// @ azam decide if to add the constraint on isBaseX.
// ToBaseXInputs stores the data [NewToBaseX] requireds as inputs.
type ToBaseXInputs struct {
	// for keccakf each lane is 64 bits (8 bytes),
	// but due to the MAXNBYTE constraint, a lane is stored over multiple rows.
	Lane ifaces.Column
	// it is 1 when the lane is active
	IsLaneActive ifaces.Column
	// bases for conversion
	BaseX []int
	// number of bits per baseX (baseA or baseB)
	// the constraint is that BaseX^nbBitPerBaseX < field.Modulus and nbBitPerBaseX is a power of 2.
	NbBitsPerBaseX int
	// the flag indicating where to convert the lane and to with base.
	IsBaseX []ifaces.Column
}

// The submodule ToBaseX implements the base conversion over
// the inputs to the keccakf (i.e., blocks/lanes), in order to export them to the keccakf.
// The lanes from the first block of hash should be in baseA and others are in baseB.
type ToBaseX struct {
	Inputs *ToBaseXInputs
	LaneX  []ifaces.Column // lanes from first block in baseA, others in baseB
	Lookup lookup          // lookup table for base conversion
	Size   int             // number of rows
}

// NewToBaseX declare the intermediate columns,
// and the constraints for changing the blocks in base uint64 into baseA/baseB.
// Note : it does not check the validity of the inputs. Particularly it does not check that
// IsBaseX are MustBeMutuallyExclusiveBinaryFlags w.r.t the IsLaneActive column.
func NewToBaseX(
	comp *wizard.CompiledIOP,
	inp ToBaseXInputs,
) *ToBaseX {

	var (
		nbSlicesBaseX = MAXNBYTE * 8 / inp.NbBitsPerBaseX

		b = &ToBaseX{
			Inputs: &inp,
			Size:   inp.Lane.Size(),
			Lookup: lookup{
				colBasex: make([][]ifaces.Column, len(inp.BaseX)),
			},
			LaneX: make([]ifaces.Column, nbSlicesBaseX),
		}
		// declare the columns
		createCol = common.CreateColFn(comp, KECCAKF_INPUT_PREPARATION, b.Size, pragmas.RightPadded)
	)
	for j := 0; j < nbSlicesBaseX; j++ {
		b.LaneX[j] = createCol("LaneX_%v", j)
	}

	for j := 0; j < len(b.Inputs.IsBaseX); j++ {
		b.Lookup.colBasex[j] = make([]ifaces.Column, nbSlicesBaseX)
	}

	// table for base conversion (used for converting blocks to what keccakf expect)
	colMAXNBYTE, colBasex := createLookupTablesBaseX(b.Inputs.BaseX, nbSlicesBaseX)

	b.Lookup.colMAXNBYTE = comp.InsertPrecomputed(ifaces.ColIDf("LOOKUP_MAXNBYTE"), colMAXNBYTE)
	for k := range b.Inputs.BaseX {
		for i := 0; i < nbSlicesBaseX; i++ {
			b.Lookup.colBasex[k][i] = comp.InsertPrecomputed(ifaces.ColIDf("LOOKUP_BaseA_%v_%v", k, i), colBasex[k][i])
		}
	}

	// if isBaseX = 1  ---> convert to basex
	// otherwise convert to keccak.BaseB
	for k := range b.Inputs.IsBaseX {
		comp.InsertInclusionConditionalOnIncluded(0,
			ifaces.QueryIDf("BaseConversion_Into_BaseX_%v", k),
			append(b.Lookup.colBasex[k], b.Lookup.colMAXNBYTE),
			append(b.LaneX, b.Inputs.Lane),
			b.Inputs.IsBaseX[k],
		)
	}

	return b
}

// assign column isFromFirstBlock and laneX
func (b *ToBaseX) Run(run *wizard.ProverRuntime) {
	var (
		isBaseX = make([][]field.Element, len(b.Inputs.BaseX))
	)

	for i := range b.Inputs.BaseX {
		// get isBaseX column
		isBaseX[i] = b.Inputs.IsBaseX[i].GetColAssignment(run).IntoRegVecSaveAlloc()
	}

	// assign the laneX
	var (
		lane  = b.Inputs.Lane.GetColAssignment(run).IntoRegVecSaveAlloc()
		laneX = make([]*common.VectorBuilder, len(b.LaneX))
		bytes = make([]byte, MAXNBYTE)
		base  int
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

		// get the base
		for i := range b.Inputs.BaseX {
			if isBaseX[i][j].IsOne() {
				base = b.Inputs.BaseX[i]
				continue
			}
		}

		a := extractLittleEndianBaseX(bytes, b.Inputs.NbBitsPerBaseX, len(b.LaneX), base)
		for k := range laneX {
			laneX[k].PushInt(int(a[k]))
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
func createLookupTablesBaseX(basex []int, nbSlices int) (uint16Col smartvectors.SmartVector, basexCol [][]smartvectors.SmartVector) {
	var (
		bitsPerSlice = MAXNBYTE * 8 / nbSlices
		u            []field.Element
		v            = make([][][]field.Element, len(basex))
		mask         = (1 << bitsPerSlice) - 1
		uintMAXNBYTE = 1<<(8*MAXNBYTE) - 1
		result       = make([]uint, nbSlices) // it holds the decomposition uint in chuncks of size bitsPerSlice
		basexFr      = make([]field.Element, len(basex))
	)

	basexCol = make([][]smartvectors.SmartVector, len(basex))

	for i := range basex {
		basexFr[i] = field.NewElement(uint64(basex[i]))
		v[i] = make([][]field.Element, nbSlices)
	}

	if MAXNBYTE > 2 {
		// too big to create a lookup table, in this case lane column needs to be decomposed into smaller chunks.
		utils.Panic("MAXNBYTE = %v create an unsupported size of lookuptable", MAXNBYTE)
	}

	// preallocate slices
	for j := range basex {
		for i := 0; i < nbSlices; i++ {
			v[j][i] = make([]field.Element, 0, uintMAXNBYTE+1)
		}
	}

	for j := 0; j <= uintMAXNBYTE; j++ {
		// decompse j into chunks of size bitsPerSlice
		n := j
		for i := 0; i < nbSlices; i++ {
			result[i] = uint(n & mask) // extract lowest chunkBits bits
			n >>= bitsPerSlice         // shift n for next chunk
		}

		u = append(u, field.NewElement(uint64(j)))
		for k := range basex {
			for i := 0; i < nbSlices; i++ {
				v[k][i] = append(v[k][i], uintToBaseX(result[i], basexFr[k], bitsPerSlice))

			}
		}
	}
	for k := range basex {
		for j := 0; j < nbSlices; j++ {
			basexCol[k] = append(basexCol[k], smartvectors.NewRegular(v[k][j]))
		}
	}

	return smartvectors.NewRegular(u), basexCol
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
//	data := []byte{0b11001010, 0b00110101}  which is the byte slice for 0x35ca
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

	// fast path, used for the padded rows
	if base == 0 {
		return make([]int, nbSlices)
	}

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
