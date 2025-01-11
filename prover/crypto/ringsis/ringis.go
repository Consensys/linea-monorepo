package ringsis

import (
	"runtime"

	"github.com/bits-and-blooms/bitset"
	"github.com/consensys/gnark-crypto/ecc/bls12-377/fr/fft"
	"github.com/consensys/gnark-crypto/ecc/bls12-377/fr/sis"
	"github.com/consensys/linea-monorepo/prover/maths/common/smartvectors"
	"github.com/consensys/linea-monorepo/prover/maths/field"
	"github.com/consensys/linea-monorepo/prover/utils"
	"github.com/consensys/linea-monorepo/prover/utils/parallel"
)

const (
	// The seed of the ring sis key
	RING_SIS_SEED int64 = 42069
)

// Key encapsulates the public parameters of an instance of the ring-SIS hash
// instance.
type Key struct {
	// gnarkInternal stores the SIS key itself and some precomputed domain
	// twiddles.
	gnarkInternal *sis.RSis
	// Params provides the parameters of the ring-SIS instance (logTwoBound,
	// degree etc)
	Params
}

// GenerateKey generates a ring-SIS key from a set of a [Params] and a max
// number of elements to hash
func GenerateKey(params Params, maxNumFieldToHash int) Key {

	// Sanity-check : if the logTwoBound is larger or equal than 64 it can
	// create overflows as we cast the small-norm limbs into
	if params.LogTwoBound >= 64 {
		utils.Panic("Log two bound cannot be larger than 64")
	}

	rsis, err := sis.NewRSis(RING_SIS_SEED, params.LogTwoDegree, params.LogTwoBound, maxNumFieldToHash)
	if err != nil {
		panic(err)
	}

	res := Key{
		gnarkInternal: rsis,
		Params:        params,
	}

	return res
}

// Ag returns the SIS key
func (s *Key) Ag() [][]field.Element {
	return s.gnarkInternal.Ag
}

// Hash interprets the input vector as a sequence of coefficients of size
// r.LogTwoBound bits long and returns the hash of the polynomial corresponding
// to the sum sum_i A[i]*m Mod X^{d}+1
//
// It is equivalent to calling r.Write(element.Marshal()); outBytes = r.Sum(nil);
func (s *Key) Hash(v []field.Element) []field.Element {

	// write the input as byte
	sum := make([]field.Element, s.OutputSize())
	if err := s.gnarkInternal.Hash(v, sum); err != nil {
		panic(err)
	}
	return sum
}

// LimbSplit breaks down the entries of `v` into short limbs representing
// `LogTwoBound` bits each. The function then flatten and flatten them in a
// vector, casted as field elements in Montgommery form.
func (s *Key) LimbSplit(vReg []field.Element) []field.Element {
	m := make([]field.Element, len(vReg)*s.NumLimbs())

	it := sis.NewLimbIterator(sis.NewVectorIterator(vReg), s.LogTwoBound/8)

	// The limbs are in regular form, we reconvert them back into montgommery
	// form
	var ok bool
	for i := range m {
		m[i][0], ok = it.NextLimb()
		if !ok {
			// the rest is 0 we can stop (note that if we change the padding
			// policy we may need to change this)
			break
		}
		m[i] = field.MulR(m[i])
	}

	return m
}

// HashModXnMinus1 applies the SIS hash modulo X^n - 1, (instead of X^n + 1).
// This is used as part of the self-recursion procedure. Note that this **does
// not implement** a collision-resistant hash function and it is meant to. Its
// purpose is to help constructing a witness for the correctness of computation
// of a batch of SIS hashes.
//
// The vector of limbs has to be provided in Montgommery form.
func (s Key) HashModXnMinus1(limbs []field.Element) []field.Element {

	// inputReader is a subslice of `limbs` and it is meant to be used as a
	// reader. We periodically pop the first element by reassigning the input
	// reader to `inputReader[1:]`.
	inputReader := limbs

	if len(limbs) > s.maxNumLimbsHashable() {
		utils.Panic("Wrong size : %v > %v", len(limbs), s.maxNumLimbsHashable())
	}

	nbPolyUsed := utils.DivCeil(len(limbs), s.modulusDegree())

	/*
		To perform the modulo X^n - 1 operation, we use the FFT method

		Let a, k, r be vectors of size d

		k <- FFT(k)
		a <- FFT(a)
		r += k * a (element-wise)
		r <- FFTInv(r)
	*/

	domain := s.gnarkInternal.Domain
	k := make([]field.Element, s.modulusDegree())
	a := make([]field.Element, s.modulusDegree())
	r := make([]field.Element, s.OutputSize())

	for i := 0; i < nbPolyUsed; i++ {
		copy(a, s.gnarkInternal.A[i])
		copy(k, inputReader)

		// consume the "reader"
		if i < nbPolyUsed-1 {
			inputReader = inputReader[s.modulusDegree():]
		} else {
			// we may need to zero-pad the last one if incomplete
			for i := len(inputReader); i < s.modulusDegree(); i++ {
				k[i].SetZero()
			}
			// we can give the inputReader back
			inputReader = nil
		}

		domain.FFT(k, fft.DIF)
		domain.FFT(a, fft.DIF)

		var tmp field.Element
		for i := range r {
			tmp.Mul(&k[i], &a[i])
			r[i].Add(&r[i], &tmp)
		}
	}

	if len(inputReader) > 0 {
		utils.Panic("%v elements remains in the reader after hashing", len(inputReader))
	}

	// by linearity, we defer the fft inverse at the end
	domain.FFTInverse(r, fft.DIT)

	// also account for the Montgommery issue : in gnark's implementation
	// the key is implictly multiplied by RInv
	for i := range r {
		r[i] = field.MulRInv(r[i])
	}

	return r
}

// FlattenedKey returns the SIS key multiplied by RInv and laid out in a
// single-vector of field elements. This function is meant to be used as part
// of the self-recursion of Vortex and serves constructing an assignment for
// the column storing the chunks of the SIS key.
//
// The function stores the key polynomial per polynomials. In coefficient form
// and multiplied by RInv so that it can account for the "Montgommery skip".
// See [Key.Hash] for more details.
func (s *Key) FlattenedKey() []field.Element {
	res := make([]field.Element, 0, s.maxNumLimbsHashable())
	for i := range s.gnarkInternal.A {
		for j := range s.gnarkInternal.A[i] {
			t := s.gnarkInternal.A[i][j]
			t = field.MulRInv(t)
			res = append(res, t)
		}
	}
	return res
}

// TransversalHash evaluates SIS hashes transversally over a list of smart-vectors.
// Each smart-vector is seen as the row of a matrix. All rows must have the same
// size or panic. The function returns the hash of the columns. The column hashes
// are concatenated into a single array.
func (s *Key) TransversalHash(v []smartvectors.SmartVector) []field.Element {

	// nbRows stores the number of rows in the matrix to hash it must be
	// strictly positive and be within the bounds of MaxNumFieldHashable.
	nbRows := len(v)

	if nbRows == 0 || nbRows > s.MaxNumFieldHashable() {
		utils.Panic("Attempted to hash %v rows, must be in [1:%v]", nbRows, s.MaxNumFieldHashable())
	}

	// nbCols stores the number of columns in the matrix to hash et must be
	// positive and all the rows must have the same size.
	nbCols := v[0].Len()

	if nbCols == 0 {
		utils.Panic("Provided a 0-column matrix")
	}

	for i := range v {
		if v[i].Len() != nbCols {
			utils.Panic("Unexpected : all inputs smart-vectors should have the same length the first one has length %v, but #%v has length %v",
				nbCols, i, v[i].Len())
		}
	}

	/*
		v contains a list of rows. We want to hash the columns, in a cache-friendly
		manner.
		we will work with "tiles" of chunks of columns.

		for example, if we consider the matrix
		v[0] -> [ 1  2  3  4  ]
		v[1] -> [ 5  6  7  8  ]
		v[2] -> [ 9  10 11 12 ]
		v[3] -> [ 13 14 15 16 ]

		we want to compute
		res = [ H(1,5,9,13) H(2,6,10,14) H(3,7,11,15) H(4,8,12,16) ]

		note that the output size of the hash is s.OutputSize() (i.e it's a slice)
		and that we will decompose the columns in "Limbs" of size s.LogTwoBound;
		this limbs are then interpreted as a slice of coefficients of
		a polynomial of size s.OutputSize()

		that is, we can decompose H(1,5,9,13) as;
		k0 := limbs(1,5) 	= [a b c d e f g h]
		k1 := limbs(9,13) 	= [i j k l m n o p]

		In practice, s.OutputSize() is a reasonable size (< 1024) so we can slide our tiles
		over the partial columns and compute the hash of the columns in parallel.

	*/

	nbBytePerLimb := s.LogTwoBound / 8
	nbLimbsPerField := field.Bytes / nbBytePerLimb
	nbFieldPerPoly := s.modulusDegree() / nbLimbsPerField

	// let's estimate a good size for the tile
	const (
		cacheLineSize        = 64
		elementsPerCacheLine = cacheLineSize / field.Bytes
		cacheSize            = 30 * 1024 // 32KB (but we remove 2kb for some margin)
	)
	// we need space for the part of the column we process, Ag[i], the limbs and the output
	// note that tile height must be a multiple of nbFieldPerPoly
	// for now we just set it to 1
	// TODO @gbotrel experiment with larger tile height
	availableL1Cache := cacheSize - (s.OutputSize() * field.Bytes * 5) // Ag[i], limbs, output, fft twiddles and cosets

	// it makes sense to have at least elementsPerCacheLine for the tile width
	tileWidth := elementsPerCacheLine

	oneHashSize := nbFieldPerPoly * field.Bytes

	delta := tileWidth * oneHashSize
	for delta < availableL1Cache {
		tileWidth += elementsPerCacheLine
		delta = tileWidth * oneHashSize
	}

	if tileWidth > nbCols {
		tileWidth = nbCols
	}
	// ensure that the tile width divides the number of columns
	// nbIterations := nbCols / tileWidth

	N := s.OutputSize()

	constPoly := make(field.Vector, N)
	res := make(field.Vector, nbCols*N)

	nbCpus := runtime.GOMAXPROCS(0)

	if nbCpus >= nbCols {
		// 1 cpu per col is fine.
		tileWidth = 1
	} else {
		// we want to have at least 1 tile width per cpu;
		for tileWidth*nbCpus > nbCols {
			tileWidth--
		}
	}

	// block is the number of tiles processed concurrently by all cpus
	blockWidth := tileWidth * nbCpus
	// nbBlocks := nbCols / blockWidth
	remainingIterations := nbCols % blockWidth
	rn := nbCols
	if remainingIterations > 0 {
		rn -= remainingIterations
	}

	type worker struct {
		k   []field.Element
		it  columnIterator
		lit *sis.LimbIterator
	}

	nbPolys := utils.DivCeil(len(v), nbFieldPerPoly)
	constBlock := bitset.New(uint(nbPolys))
	for start := 0; start < len(v); start += nbFieldPerPoly {
		end := start + nbFieldPerPoly
		if end > len(v) {
			end = len(v)
		}
		polID := start / nbFieldPerPoly

		// first, we iterate over the lines; if all are smartvectors.Constant, we contribute the contribution
		// once only, and add it at the end before reducing
		constantBlock := true
		for row := start; row < end; row++ {
			if _, ok := v[row].(*smartvectors.Constant); !ok {
				constantBlock = false
				break
			}
		}

		if constantBlock {
			constBlock.Set(uint(polID))
		}
	}

	// TODO @gbotrel use go routines directly so that we don't process the full row at once
	// but chunk it for larger matrices to fit in L2 - L3 cache.
	parallel.Execute(nbCols, func(colStart, colEnd int) {

		w := worker{
			k: make([]field.Element, N),
			it: columnIterator{
				v: v,
			},
		}
		w.lit = sis.NewLimbIterator(&w.it, s.LogTwoBound/8)

		for rowStart := 0; rowStart < len(v); rowStart += nbFieldPerPoly {
			rowEnd := rowStart + nbFieldPerPoly
			if rowEnd > len(v) {
				rowEnd = len(v)
			}
			polID := rowStart / nbFieldPerPoly

			if colStart == 0 {
				// first worker process the constants;
				w.it.rowStart = rowStart
				w.it.rowEnd = rowEnd
				w.it.isConstIT = true
				w.it.colIndex = 0
				w.lit = sis.NewLimbIterator(&w.it, s.LogTwoBound/8)

				s.gnarkInternal.InnerHash(w.lit, constPoly, w.k, polID)
			}

			// if it's a constant block, we are done.
			if constBlock.Test(uint(polID)) {
				continue
			}

			w.it.isConstIT = false
			w.it.rowStart = rowStart
			w.it.rowEnd = rowEnd

			for colID := colStart; colID < colEnd; colID++ {
				w.it.colIndex = colID
				w.it.rowStart = rowStart
				w.lit.Reset(&w.it)
				s.gnarkInternal.InnerHash(w.lit, res[colID*N:colID*N+N], w.k, polID)
			}
		}

	})

	// now for each subslice in results, we do the FFT inverse to reduce mod Xᵈ+1
	parallel.Execute(nbCols, func(start, stop int) {
		for j := start; j < stop; j++ {
			vRes := field.Vector(res[j*N : (j+1)*N])
			vRes.Add(vRes, constPoly)
			s.gnarkInternal.Domain.FFTInverse(res[j*N:(j+1)*N], fft.DIT, fft.OnCoset(), fft.WithNbTasks(1))
		}
	})

	return res
}

// columnIterator is a helper struct to iterate over the columns of a matrix
// it implements the SIS.ElementIterator interface
type columnIterator struct {
	v                []smartvectors.SmartVector
	rowStart, rowEnd int
	colIndex         int
	isConstIT        bool
}

// func newColumnIterator(v []smartvectors.SmartVector, colIndex int, constIT bool) *columnIterator {
// 	return &columnIterator{v: v, colIndex: colIndex, isConstIT: constIT}
// }

func (it *columnIterator) Next() (field.Element, bool) {
	if it.rowEnd == it.rowStart {
		return field.Element{}, false
	}
	row := it.v[it.rowStart]
	_, constRow := row.(*smartvectors.Constant)
	it.rowStart++
	if (it.isConstIT && constRow) || (!it.isConstIT && !constRow) {
		return row.Get(it.colIndex), true
	}
	return field.Element{}, true
}
