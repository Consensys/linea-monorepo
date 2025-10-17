package ringsis

import (
	"github.com/consensys/gnark-crypto/field/koalabear/fft"
	"github.com/consensys/gnark-crypto/field/koalabear/sis"
	"github.com/consensys/linea-monorepo/prover/maths/common/smartvectors"
	"github.com/consensys/linea-monorepo/prover/maths/field"
	"github.com/consensys/linea-monorepo/prover/maths/field/fext"

	"github.com/consensys/linea-monorepo/prover/utils"
	"github.com/consensys/linea-monorepo/prover/utils/arena"
	"github.com/consensys/linea-monorepo/prover/utils/parallel"
)

const (
	// The seed of the ring sis key
	RING_SIS_SEED int64 = 42069
)

// Key encapsulates the public parameters of an instance of the ring-SIS hash
// instance.
type Key struct {
	// GnarkInternal stores the SIS key itself and some precomputed domain
	// twiddles.
	GnarkInternal *sis.RSis
	// Params provides the parameters of the ring-SIS instance (logTwoBound,
	// degree etc)
	*KeyGen

	// twiddleCosets stores the list of twiddles that we use to implement the
	// SIS parameters. The twiddleAreInternally are only used when dealing with
	// the parameters modulusDegree=64 and logTwoBound=8 and is passed as input
	// to the specially unrolled [sis.FFT64] function. They are thus optionally
	// constructed when [GenerateKey] is called.
	twiddleCosets []field.Element `serde:"omit"`
}

type KeyGen struct {
	*Params
	MaxNumFieldToHash int
}

// GenerateKey generates a ring-SIS key from a set of a [Params] and a max
// number of elements to hash
func GenerateKey(params Params, maxNumFieldToHash int) *Key {

	// Sanity-check : if the logTwoBound is larger or equal than 64 it can
	// create overflows as we cast the small-norm limbs into
	if params.LogTwoBound >= 64 {
		utils.Panic("Log two bound cannot be larger than 64")
	}

	rsis, err := sis.NewRSis(RING_SIS_SEED, params.LogTwoDegree, params.LogTwoBound, maxNumFieldToHash)
	if err != nil {
		panic(err)
	}

	res := &Key{
		GnarkInternal: rsis,
		KeyGen:        &KeyGen{&params, maxNumFieldToHash},
		twiddleCosets: nil,
	}
	return res
}

// Hash interprets the input vector as a sequence of coefficients of size
// r.LogTwoBound bits long and returns the hash of the polynomial corresponding
// to the sum sum_i A[i]*m Mod X^{d}+1
//
// It is equivalent to calling r.Write(element.Marshal()); outBytes = r.Sum(nil);
func (s *Key) Hash(v []field.Element) []field.Element {

	result := make([]field.Element, s.GnarkInternal.Degree)
	err := s.GnarkInternal.Hash(v, result)

	if err != nil {
		panic(err)
	}

	return result
}

// LimbSplit breaks down the entries of `v` into short limbs representing
// `LogTwoBound` bits each. The function then flatten and flatten them in a
// vector, casted as field elements in Montgommery form.
func (s *Key) LimbSplit(vReg []field.Element) []field.Element {

	vr := sis.NewLimbIterator(sis.NewVectorIterator(vReg), s.LogTwoBound/8)
	m := make([]field.Element, len(vReg)*s.NumLimbs())
	var ok bool
	for i := 0; i < len(m); i++ {
		m[i][0], ok = vr.NextLimb()
		if !ok {
			utils.Panic("LimbSplit panic")
		}
	}
	// The limbs are in regular form, we reconvert them back into montgommery
	// form
	for i := 0; i < len(m); i++ {
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
func (s Key) HashModXnMinus1(limbs []fext.Element) []fext.Element {

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

	domain := s.GnarkInternal.Domain
	k := make([]fext.Element, s.modulusDegree())
	a := make([]field.Element, s.modulusDegree())
	r := make([]fext.Element, s.OutputSize())

	for i := 0; i < nbPolyUsed; i++ {
		copy(a, s.GnarkInternal.A[i])
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

		domain.FFTExt(k, fft.DIF)
		domain.FFT(a, fft.DIF)

		var tmp fext.Element
		for i := range r {
			tmp.MulByElement(&k[i], &a[i])
			r[i].Add(&r[i], &tmp)
		}
	}

	if len(inputReader) > 0 {
		utils.Panic("%v elements remains in the reader after hashing", len(inputReader))
	}

	// by linearity, we defer the fft inverse at the end
	domain.FFTInverseExt(r, fft.DIT)

	// also account for the Montgommery issue : in gnark's implementation
	// the key is implictly multiplied by RInv
	for i := range r {
		r[i] = fext.MulRInv(r[i])
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
	for i := range s.GnarkInternal.A {
		for j := range s.GnarkInternal.A[i] {
			t := s.GnarkInternal.A[i][j]
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
//
// The function is optimize to deal with the ring-SIS instances parametrized by
//
//   - modulus degree: 	64  log2(bound): 	8
//   - modulus degree: 	64  log2(bound): 	16
//   - modulus degree: 	32  log2(bound): 	8
func (s *Key) TransversalHash(v []smartvectors.SmartVector, res []field.Element) []field.Element {
	// nbRows is the number of rows in the matrix
	nbRows := len(v)
	if nbRows == 0 || nbRows > s.MaxNumFieldToHash {
		utils.Panic("nbRows=%v out of bounds [1,%v]", nbRows, s.MaxNumFieldToHash)
	}

	// nbCols is the number of columns in the matrix
	nbCols := v[0].Len()
	if nbCols == 0 {
		utils.Panic("Provided a 0-colums matrix")
	}

	for i := range v {
		if v[i].Len() != nbCols {
			utils.Panic("Unexpected : all inputs smart-vectors should have the same length the first one has length %v, but #%v has length %v",
				nbCols, i, v[i].Len())
		}
	}

	sisKeySize := s.OutputSize()

	parallel.Execute(nbCols, func(start, end int) {
		// we transpose the columns using a windowed approach
		// this is done to improve memory accesses when transposing the matrix

		// perf note; we could allocate only blocks of 256 elements here and do the SIS hash
		// block by block, but surprisingly it is slower than the current implementation
		// it would however save some memory allocation.
		windowSize := 4
		n := end - start
		for n%windowSize != 0 {
			windowSize /= 2
		}
		// using arena here just favors contiguous memory allocation
		transposedArena := arena.NewVectorArena[field.Element](nbRows * windowSize)
		transposed := make([][]field.Element, windowSize)
		for i := range transposed {
			transposed[i] = arena.Get[field.Element](transposedArena, nbRows)
		}
		for col := start; col < end; col += windowSize {
			for i := 0; i < nbRows; i++ {
				switch vi := v[i].(type) {
				case *smartvectors.Constant:
					cst := vi.Value
					for j := range transposed {
						transposed[j][i] = cst
					}
				case *smartvectors.Regular:
					for j := range transposed {
						transposed[j][i] = (*vi)[col+j]
					}
				default:
					for j := range transposed {
						transposed[j][i] = v[i].Get(col + j)
					}
				}
			}
			for j := range transposed {
				s.GnarkInternal.Hash(transposed[j], res[(col+j)*sisKeySize:(col+j)*sisKeySize+sisKeySize])
			}
		}
	})

	return res
}
