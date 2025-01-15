package serialization

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/big"
	"sync"
	"unsafe"

	"github.com/consensys/linea-monorepo/prover/maths/common/smartvectors"
	"github.com/consensys/linea-monorepo/prover/maths/field"
	"github.com/consensys/linea-monorepo/prover/protocol/ifaces"
	"github.com/consensys/linea-monorepo/prover/utils/collection"
	"github.com/consensys/linea-monorepo/prover/utils/parallel"
	"github.com/pierrec/lz4/v4"
	"github.com/sirupsen/logrus"
)

// WAssignment is an alias for the mapping type used to represent the assignment
// of a column in a [wizard.ProverRuntime]
type WAssignment = collection.Mapping[ifaces.ColID, smartvectors.SmartVector]


// hashWAssignment computes a hash of the serialized WAssignment structure.
func hashWAssignment(a WAssignment) string {
	serialized, err := json.Marshal(a)
	if err != nil {
		panic("Failed to serialize WAssignment for hashing: " + err.Error())
	}
	h := sha256.New()
	h.Write(serialized)
	return hex.EncodeToString(h.Sum(nil))
}

// SerializeAssignment serializes map representing the column assignment of a
// wizard protocol.
func SerializeAssignment(a WAssignment, numChunks int) []json.RawMessage {
	logrus.Infof("Hash of WAssignment before serialization: %s", hashWAssignment(a))

	var (
		as    = a.InnerMap()
		ser   = map[string]*CompressedSmartVector{}
		names = a.ListAllKeys()
		lock  = &sync.Mutex{}
	)

	parallel.ExecuteChunky(len(names), func(start, stop int) {
		for i := start; i < stop; i++ {
			v := CompressSmartVector(as[names[i]])
			lock.Lock()
			// Convert names[i] to string using fmt.Sprintf or another method
			ser[fmt.Sprintf("%v", names[i])] = v
			lock.Unlock()
		}
	})

	// Calculate the size of `ser` in bytes
	var serSizeBytes uintptr
	for _, v := range ser {
		serSizeBytes += unsafe.Sizeof(*v)
	}

	// Convert size to GB
	serSizeGB := float64(serSizeBytes) / (1024 * 1024 * 1024)
	logrus.Infof("Size of ser : %.6f GB", serSizeGB)

	// Parallelize CBOR serialization by chunking `ser`
	chunkSize := (len(ser) + numChunks - 1) / numChunks // Calculate the size of each chunk
	var serializedChunks = make([]json.RawMessage, numChunks)
	var wg sync.WaitGroup
	var m sync.Mutex
	for i := 0; i < numChunks; i++ {
		wg.Add(1)
		go func(chunkIndex int) {
			defer wg.Done()

			// Select chunk of data for this goroutine
			start := chunkIndex * chunkSize
			stop := start + chunkSize
			if stop > len(names) {
				stop = len(names)
			}

			// Prepare chunk map for serialization
			chunk := make(map[string]*CompressedSmartVector)
			for j := start; j < stop; j++ {
				// Convert names[j] to string
				key := fmt.Sprintf("%v", names[j])
				chunk[key] = ser[key]
			}

			// Serialize the chunk with CBOR
			serializedChunk := serializeAnyWithCborPkg(chunk)
			logrus.Infof("Serialized chunk %d, size: %d bytes", i, len(serializedChunk))

			// Store the result in the slice
			m.Lock()
			serializedChunks[chunkIndex] = serializedChunk
			m.Unlock()
		}(i)
	}
	wg.Wait()

	return serializedChunks
}

// CompressChunks compresses each serialized chunk
func CompressChunks(chunks []json.RawMessage) []json.RawMessage {
	compressedChunks := make([]json.RawMessage, len(chunks))
	var wg sync.WaitGroup

	for i, chunk := range chunks {
		wg.Add(1)
		go func(i int, chunk json.RawMessage) {
			defer wg.Done()
			var compressedData bytes.Buffer
			lz4Writer := lz4.NewWriter(&compressedData)
			_, err := lz4Writer.Write(chunk)
			if err != nil {
				logrus.Errorf("Error compressing chunk %d: %v", i, err)
				panic(err) // handle error as needed
			}
			lz4Writer.Close() // finalize the LZ4 stream
			compressedChunks[i] = compressedData.Bytes()
			logrus.Infof("Compressed chunk %d, size: %d bytes", i, len(compressedChunks[i]))
		}(i, chunk)
	}
	wg.Wait()

	return compressedChunks
}
// DeserializeAssignment deserializes a blob of bytes into a set of column
// assignments representing assigned columns of a Wizard protocol.
func DeserializeAssignment(filepath string, numChunks int) (WAssignment, error) {
	var (
		res  = collection.NewMapping[ifaces.ColID, smartvectors.SmartVector]()
		lock = &sync.Mutex{}
	)

	logrus.Infof("Reading the assignment files")

	// Read and decompress each chunk individually
	var wg sync.WaitGroup
	for i := 0; i < numChunks; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			chunkPath := fmt.Sprintf("%s_chunk_%d", filepath, i)
			chunkData, err := ioutil.ReadFile(chunkPath)
			if err != nil {
				logrus.Errorf("Failed to read chunk %d: %v", i, err)
				return
			}

			logrus.Infof("Reading chunk %d from %s, size: %d bytes", i, chunkPath, len(chunkData))
			lz4Reader := lz4.NewReader(bytes.NewReader(chunkData))
			var decompressedData bytes.Buffer
			n, err := decompressedData.ReadFrom(lz4Reader)
			if err != nil {
				logrus.Errorf("Error decompressing chunk %d: %v", i, err)
				return
			}
			logrus.Infof("Decompressed chunk %d, size: %d bytes", i, n)

			// Deserialize the decompressed chunk
			var chunkMap map[string]*CompressedSmartVector
			if err := deserializeAnyWithCborPkg(decompressedData.Bytes(), &chunkMap); err != nil {
				logrus.Errorf("Error deserializing chunk %d: %v", i, err)
				return
			}

			// Reconstruct the WAssignment
			for k, v := range chunkMap {
				decompressed := v.Decompress()
				lock.Lock()
				res.InsertNew(ifaces.ColID(k), decompressed)
				lock.Unlock()
			}
		}(i)
	}
	wg.Wait()

	logrus.Infof("Hash of WAssignment after deserialization: %s", hashWAssignment(res))

	return res, nil
}

// CompressedSmartVector represents a [smartvectors.SmartVector] in a more
// space-efficient manner.
type CompressedSmartVector struct {
	F []CompressedSVFragment
}

// CompressedSVFragment represent a portion of a SerializableSmartVector
type CompressedSVFragment struct {
	// L is the byte length used by the fragment
	L uint8
	// X is the value used to represent a single field element
	X *big.Int
	// V is a byteslice storing the bytes of a vector if the fragment represent
	// plain values.
	V []byte
	// N is the number of repetion used
	N int
}

func CompressSmartVector(sv smartvectors.SmartVector) *CompressedSmartVector {

	switch v := sv.(type) {
	case *smartvectors.Constant:
		return &CompressedSmartVector{
			F: []CompressedSVFragment{
				newConstantSVFragment(v.Val(), v.Len()),
			},
		}
	case *smartvectors.Regular:
		return &CompressedSmartVector{
			F: []CompressedSVFragment{
				newSliceSVFragment(*v),
			},
		}
	case *smartvectors.PaddedCircularWindow:
		var (
			w          = v.Window()
			offset     = v.Offset()
			paddingVal = v.PaddingVal()
			fullLen    = v.Len()
		)

		// It's a left-padded value
		if offset == 0 {
			return &CompressedSmartVector{
				F: []CompressedSVFragment{
					newSliceSVFragment(w),
					newConstantSVFragment(paddingVal, fullLen-len(w)),
				},
			}
		}

		// It's a right-padded value
		if offset+len(w) == fullLen {
			return &CompressedSmartVector{
				F: []CompressedSVFragment{
					newConstantSVFragment(paddingVal, fullLen-len(w)),
					newSliceSVFragment(w),
				},
			}
		}

	}

	// The other cases are not expected, we still support them via a
	// suboptimal method.
	return &CompressedSmartVector{
		F: []CompressedSVFragment{
			newSliceSVFragment(sv.IntoRegVecSaveAlloc()),
		},
	}
}

func (sv *CompressedSmartVector) Decompress() smartvectors.SmartVector {

	if len(sv.F) == 1 && sv.F[0].isConstant() {
		val := new(field.Element).SetBigInt(sv.F[0].X)
		return smartvectors.NewConstant(*val, sv.F[0].N)
	}

	if len(sv.F) == 1 && sv.F[0].isPlain() {
		return smartvectors.NewRegular(sv.F[0].readSlice())
	}

	if len(sv.F) == 2 && sv.F[0].isConstant() && sv.F[1].isPlain() {

		var (
			paddingVal = new(field.Element).SetBigInt(sv.F[0].X)
			window     = sv.F[1].readSlice()
			size       = sv.F[0].N + len(window)
		)

		return smartvectors.LeftPadded(window, *paddingVal, size)
	}

	if len(sv.F) == 2 && sv.F[1].isConstant() && sv.F[0].isPlain() {

		var (
			paddingVal = new(field.Element).SetBigInt(sv.F[1].X)
			window     = sv.F[0].readSlice()
			size       = sv.F[1].N + len(window)
		)

		return smartvectors.RightPadded(window, *paddingVal, size)
	}

	panic("unexpected pattern")
}

func (f *CompressedSVFragment) isConstant() bool {
	return f.X != nil
}

func (f *CompressedSVFragment) isPlain() bool {
	return f.V != nil
}

func (f *CompressedSVFragment) readSlice() []field.Element {

	var (
		l   = int(f.L)
		buf = bytes.NewBuffer(f.V)
		tmp = [32]byte{}
		n   = f.N
	)

	if l > 0 {
		n = len(f.V) / l
	}

	var (
		res = make([]field.Element, n)
	)

	for i := range res {
		buf.Read(tmp[32-l:])
		res[i].SetBytes(tmp[:])
	}

	return res
}

func newConstantSVFragment(x field.Element, n int) CompressedSVFragment {

	var (
		f big.Int
		_ = x.BigInt(&f)
	)

	return CompressedSVFragment{
		X: &f,
		N: n,
	}
}

func newSliceSVFragment(fv []field.Element) CompressedSVFragment {

	var (
		l int
	)

	for i := range fv {
		l = max(l, (fv[i].BitLen()+7)/8)
	}

	var (
		res    = make([]byte, 0, len(fv)*l)
		resBuf = bytes.NewBuffer(res)
	)

	for i := range fv {
		fbytes := fv[i].Bytes()
		resBuf.Write(fbytes[32-l:])
	}

	compressed := CompressedSVFragment{
		L: uint8(l),
		V: resBuf.Bytes(),
	}

	// Can happen if the caller provides a vector of the form [0, 0, 0, 0]. In
	// that case the value of "n" cannot be infered from the slice because the
	// slice will be empty. The solution is to provide a length to the vector
	if l == 0 {
		compressed.N = len(fv)
	}

	return compressed
}
