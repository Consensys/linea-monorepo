package ringsis

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"runtime"
	"sync"

	"github.com/consensys/accelerated-crypto-monorepo/maths/common/smartvectors"
	"github.com/consensys/accelerated-crypto-monorepo/maths/field"
	"github.com/consensys/accelerated-crypto-monorepo/utils"
	"github.com/consensys/accelerated-crypto-monorepo/utils/parallel"
	"github.com/consensys/gnark-crypto/ecc/bn254/fr"
	"github.com/consensys/gnark-crypto/ecc/bn254/fr/fft"
	"github.com/consensys/gnark-crypto/ecc/bn254/fr/sis"
)

const (
	// The seed of the ring sis key
	RING_SIS_SEED int64 = 42069
)

// Encapsulate the setup of a ring sis instance
type Key struct {
	// prevents from hashing concurrently with the same key object
	lock             *sync.Mutex
	*sis.RSis            // associated gnark object
	*Params              // parameters of the SIS instance
	MaxNbFieldToHash int // reexport gnark-crypto's field
}

// Generate a key from a set of parameters and a max number of elements to hash
func (p *Params) GenerateKey(maxNbToHash int) Key {

	// Sanity-check : if the logTwoBound is larger or equal than 64
	// it can create overflows when computing 1 << max as a u64.
	if p.LogTwoBound_ >= 64 {
		utils.Panic("log two bound cannot be larger than 64")
	}

	key, err := sis.NewRSis(RING_SIS_SEED, p.LogTwoDegree, p.LogTwoBound_, maxNbToHash)
	if err != nil {
		panic(err)
	}
	return Key{
		lock:             &sync.Mutex{},
		RSis:             key,
		Params:           p,
		MaxNbFieldToHash: maxNbToHash,
	}
}

// Hash interprets the input vector as a sequence of coefficients of size r.LogTwoBound bits long,
// and return the hash of the polynomial corresponding to the sum sum_i A[i]*m Mod X^{d}+1
//
// It is equivalent to calling r.Write(element.Marshal()); outBytes = r.Sum(nil);
func (s *Key) Hash(v []field.Element) []field.Element {
	// since hashing writes into internal buffers
	// we need to guard against races conditions.
	s.lock.Lock()
	defer s.lock.Unlock()

	// write the input as byte
	s.Reset()
	for i := range v {
		_, err := s.Write(v[i].Marshal())
		if err != nil {
			panic(err)
		}
	}
	sum := s.Sum(make([]byte, 0, field.Bytes*s.Degree))

	// unmarshal the result
	var rlen [4]byte
	binary.BigEndian.PutUint32(rlen[:], uint32(len(sum)/fr.Bytes))
	reader := io.MultiReader(bytes.NewReader(rlen[:]), bytes.NewReader(sum))
	var result fr.Vector
	_, err := result.ReadFrom(reader)
	if err != nil {
		panic(err)
	}

	return result
}

// LimbSplit the inputs
func (s *Key) LimbSplit(v []field.Element) []field.Element {

	writer := bytes.Buffer{}
	for i := range v {
		b := v[i].Bytes() // big endian serialization
		writer.Write(b[:])
	}

	buf := writer.Bytes()
	m := make([]field.Element, len(v)*s.NumLimbs())
	sis.LimbDecomposeBytes(buf, m, s.LogTwoBound)

	// The limbs are in regular form, we reconvert them back
	// into montgommery form
	for i := range m {
		m[i] = MulR(m[i])
	}

	return m
}

// Applies the SIS hash modulo X^n - 1, (instead of X^n + 1)
// This is used for self-recursion purpose
func (key Key) HashModXnMinus1(limbs []field.Element) []field.Element {

	// we use a subslice as reader
	inputReader := limbs
	if len(limbs) > len(key.A)*key.Degree {
		utils.Panic("wrong size : %v > %v", len(limbs), len(key.A)*key.Degree)
	}

	nbPolyUsed := utils.DivCeil(len(limbs), key.Degree)

	/*
		To perform the modulo X^n - 1 operation, we use the FFT method

		Let a, k, r be vectors of size d

		k <- FFT(k)
		a <- FFT(a)
		r += k * a (element-wise)
		r <- FFTInv(r)
	*/

	domain := key.Domain
	k := make([]field.Element, key.Degree)
	a := make([]field.Element, key.Degree)
	r := make([]field.Element, key.Degree)

	for i := 0; i < nbPolyUsed; i++ {
		copy(a, key.A[i])
		copy(k, inputReader)

		// consume the "reader"
		if i < nbPolyUsed-1 {
			inputReader = inputReader[key.Degree:]
		} else {
			// we may need to zero-pad the last one if incomplete
			for i := len(inputReader); i < key.Degree; i++ {
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

	if len(inputReader) > key.Degree {
		utils.Panic("%v elements remains in the reader after hashing", len(inputReader))
	}

	// by linearity, we defer the fft inverse at the end
	domain.FFTInverse(r, fft.DIT)

	// also account for the Montgommery issue : in gnark's implementation
	// the key is implictly multiplied by RInv
	for i := range r {
		r[i] = MulRInv(r[i])
	}

	return r
}

// Return the SIS key multiplied by RInv and laid out on a single-vector
func (s *Key) LaidOutKey() []field.Element {
	res := make([]field.Element, 0, len(s.A)*s.Degree)
	for i := range s.A {
		for j := range s.A[i] {
			t := s.A[i][j]
			t = MulRInv(t)
			res = append(res, t)
		}
	}
	return res
}

// Returns a repr for the key
func (s *Key) Repr() string {
	return fmt.Sprintf("SISKEY_%v_%v_%v", s.LogTwoDegree, s.LogTwoBound, s.MaxNbFieldToHash)
}

// Evaluate sis hashes transversally over a list of smart-vectors. Each smart-vector
// is seen as the row of a matrix. All rows must have the same size or panic.
// The function returns the hash of the columns. The column hashes are concatenated
// into a single array.
func (s *Key) TransversalHash(v []smartvectors.SmartVector) []field.Element {

	// Assert that all smart-vectors have the same numCols
	numCols := v[0].Len()
	for i := range v {
		if v[i].Len() != numCols {
			utils.Panic("Unexpected : all inputs smart-vectors should have the same length the first one has length %v, but #%v has length %v",
				numCols, i, v[i].Len())
		}
	}

	res := make([]field.Element, numCols*s.OutputSize())

	// Assert that the length does not overcome the maximal hashing capacity
	numRows := len(v)
	if numRows > s.MaxNbFieldToHash {
		utils.Panic("attempted to hash %v rows, but the limit is %v", numCols, s.MaxNbFieldToHash)
	}

	// Will contain keys per threads
	keys := make([]*Key, runtime.GOMAXPROCS(0))
	buffers := make([][]field.Element, runtime.GOMAXPROCS(0))

	parallel.ExecuteThreadAware(
		numCols,
		func(threadID int) {
			keys[threadID] = s.CopyWithFreshBuffer()
			buffers[threadID] = make([]field.Element, numRows)
		},
		func(col, threadID int) {
			buffer := buffers[threadID]
			key := keys[threadID]
			for row := 0; row < numRows; row++ {
				buffer[row] = v[row].Get(col)
			}
			copy(res[col*key.OutputSize():(col+1)*key.OutputSize()], key.Hash(buffer))
		})

	return res
}

// Creates a copy of the key with fresh buffers. Shallow copies the
// the key itself.
func (s *Key) CopyWithFreshBuffer() *Key {

	// since hashing writes into internal buffers
	// we need to guard against races conditions.
	s.lock.Lock()
	defer s.lock.Unlock()

	clonedRsis := s.RSis.CopyWithFreshBuffer()
	return &Key{
		lock:             &sync.Mutex{},
		RSis:             &clonedRsis,
		Params:           s.Params,
		MaxNbFieldToHash: s.MaxNbFieldToHash,
	}
}

// Hash a vector of limbs. Unoptimized, only there for testing purpose.
func (key *Key) HashFromLimbs(limbs []field.Element) []field.Element {

	nbPolyUsed := utils.DivCeil(len(limbs), key.Degree)

	if nbPolyUsed > len(key.Ag) {
		utils.Panic("too many inputs max is %v but has %v", len(key.Ag)*key.Degree, len(limbs))
	}

	// we can hash now.
	res := make([]field.Element, key.Degree)

	// method 1: fft
	k := make([]field.Element, key.Degree)
	for i := 0; i < nbPolyUsed; i++ {

		// The last poly may be incomplete
		copy(k, limbs[i*key.Degree:])

		// so we may need to zero-pad the last one if incomplete. This
		// loop will be skipped if limns[i*key.Degree] is larger than
		// the degree
		for i := len(limbs[i*key.Degree:]); i < key.Degree; i++ {
			k[i].SetZero()
		}

		key.Domain.FFT(k, fft.DIF, fft.OnCoset(), fft.WithNbTasks(1))
		var tmp field.Element
		for j := range res {
			tmp.Mul(&k[j], &key.Ag[i][j])
			res[j].Add(&res[j], &tmp)
		}
	}

	// also account for the Montgommery issue : in gnark's implementation
	// the key is implictly multiplied by RInv
	for j := range res {
		res[j] = MulRInv(res[j])
	}

	key.Domain.FFTInverse(res, fft.DIT, fft.OnCoset(), fft.WithNbTasks(1)) // -> reduces mod Xᵈ+1
	return res
}

// For testing/debugging purposes
func HashLimbsWithSlice(keySlice []field.Element, limbs []field.Element, domain *fft.Domain, degree int) []field.Element {

	nbPolyUsed := utils.DivCeil(len(limbs), degree)

	if len(limbs) > len(keySlice) {
		utils.Panic("too many inputs max is %v but has %v", len(limbs), len(keySlice))
	}

	// we can hash now.
	res := make([]field.Element, degree)

	// method 1: fft
	limbsChunk := make([]field.Element, degree)
	keyChunk := make([]field.Element, degree)
	for i := 0; i < nbPolyUsed; i++ {
		// extract the key element, without copying
		copy(keyChunk, keySlice[i*degree:(i+1)*degree])
		domain.FFT(keyChunk, fft.DIF, fft.OnCoset(), fft.WithNbTasks(1))

		// extract a copy of the limbs
		// The last poly may be incomplete
		copy(limbsChunk, limbs[i*degree:])

		// so we may need to zero-pad the last one if incomplete. This
		// loop will be skipped if limns[i*degree] is larger than
		// the degree
		for i := len(limbs[i*degree:]); i < degree; i++ {
			limbsChunk[i].SetZero()
		}

		domain.FFT(limbsChunk, fft.DIF, fft.OnCoset(), fft.WithNbTasks(1))

		var tmp field.Element
		for j := range res {
			tmp.Mul(&keyChunk[j], &limbsChunk[j])
			res[j].Add(&res[j], &tmp)
		}
	}

	domain.FFTInverse(res, fft.DIT, fft.OnCoset(), fft.WithNbTasks(1)) // -> reduces mod Xᵈ+1
	return res
}
