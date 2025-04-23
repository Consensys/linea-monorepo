package serialization

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math/big"
	"os"
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
// of a column in a wizard.ProverRuntime.
type WAssignment = collection.Mapping[ifaces.ColID, smartvectors.SmartVector]

// serializeWAssignment serializes a WAssignment to CBOR.
func serializeWAssignment(w WAssignment) (json.RawMessage, error) {
	ser := make(map[string]json.RawMessage, w.Len())
	for name, sv := range w.InnerMap() {
		compressed := CompressSmartVector(sv)
		data, err := compressed.Serialize()
		if err != nil {
			return nil, fmt.Errorf("failed to serialize SmartVector for %v: %w", name, err)
		}
		ser[string(name)] = data
	}
	return serializeAnyWithCborPkg(ser)
}

// deserializeWAssignment deserializes a WAssignment from CBOR.
func deserializeWAssignment(data json.RawMessage) (WAssignment, error) {
	var ser map[string]json.RawMessage
	if err := deserializeAnyWithCborPkg(data, &ser); err != nil {
		return collection.NewMapping[ifaces.ColID, smartvectors.SmartVector](), fmt.Errorf("failed to deserialize WAssignment: %w", err)
	}
	res := collection.NewMapping[ifaces.ColID, smartvectors.SmartVector]()
	for name, raw := range ser {
		var compressed CompressedSmartVector
		if err := compressed.Deserialize(raw); err != nil {
			return collection.NewMapping[ifaces.ColID, smartvectors.SmartVector](), fmt.Errorf("failed to deserialize SmartVector for %v: %w", name, err)
		}
		res.InsertNew(ifaces.ColID(name), compressed.Decompress())
	}
	return res, nil
}

// hashWAssignment computes a hash of the serialized WAssignment structure.
func hashWAssignment(a WAssignment) string {
	serialized, err := serializeWAssignment(a)
	if err != nil {
		logrus.Errorf("Failed to serialize WAssignment for hashing: %v", err)
		return ""
	}
	h := sha256.New()
	h.Write(serialized)
	return hex.EncodeToString(h.Sum(nil))
}

// SerializeAssignment serializes map representing the column assignment of a wizard protocol.
func SerializeAssignment(a WAssignment, numChunks int) ([]json.RawMessage, error) {
	if numChunks <= 0 {
		return nil, fmt.Errorf("invalid numChunks: %d", numChunks)
	}

	logrus.Infof("Hash of WAssignment before serialization: %s", hashWAssignment(a))

	var (
		ser   = make(map[string]*CompressedSmartVector, a.Len())
		names = a.ListAllKeys()
		lock  sync.Mutex
	)

	parallel.ExecuteChunky(len(names), func(start, stop int) {
		for i := start; i < stop; i++ {
			v := CompressSmartVector(a.InnerMap()[names[i]])
			lock.Lock()
			ser[string(names[i])] = v
			lock.Unlock()
		}
	})

	// Calculate the size of `ser` in bytes (approximate)
	serSizeBytes := uint64(len(ser)) * uint64(unsafe.Sizeof(*new(CompressedSmartVector)))
	serSizeGB := float64(serSizeBytes) / (1024 * 1024 * 1024)
	logrus.Infof("Size of ser: %.6f GB", serSizeGB)

	// Parallelize CBOR serialization by chunking `ser`
	chunkSize := (len(names) + numChunks - 1) / numChunks
	serializedChunks := make([]json.RawMessage, numChunks)
	var wg sync.WaitGroup
	var m sync.Mutex
	for i := 0; i < numChunks; i++ {
		wg.Add(1)
		go func(chunkIndex int) {
			defer wg.Done()

			start := chunkIndex * chunkSize
			stop := start + chunkSize
			if stop > len(names) {
				stop = len(names)
			}

			chunk := make(map[string]*CompressedSmartVector)
			for j := start; j < stop; j++ {
				key := string(names[j])
				chunk[key] = ser[key]
			}

			serializedChunk, err := serializeAnyWithCborPkg(chunk)
			if err != nil {
				logrus.Errorf("Failed to serialize chunk %d: %v", chunkIndex, err)
				return
			}
			logrus.Infof("Serialized chunk %d, size: %d bytes", chunkIndex, len(serializedChunk))

			m.Lock()
			serializedChunks[chunkIndex] = serializedChunk
			m.Unlock()
		}(i)
	}
	wg.Wait()

	// Check for errors (since goroutines may have failed silently)
	for i, chunk := range serializedChunks {
		if chunk == nil {
			return nil, fmt.Errorf("failed to serialize chunk %d", i)
		}
	}

	return serializedChunks, nil
}

// CompressChunks compresses each serialized chunk.
func CompressChunks(chunks []json.RawMessage) ([]json.RawMessage, error) {
	compressedChunks := make([]json.RawMessage, len(chunks))
	var wg sync.WaitGroup
	var err error
	var m sync.Mutex

	for i, chunk := range chunks {
		wg.Add(1)
		go func(i int, chunk json.RawMessage) {
			defer wg.Done()
			var compressedData bytes.Buffer
			lz4Writer := lz4.NewWriter(&compressedData)
			_, writeErr := lz4Writer.Write(chunk)
			if writeErr != nil {
				logrus.Errorf("Error compressing chunk %d: %v", i, writeErr)
				m.Lock()
				err = fmt.Errorf("failed to compress chunk %d: %w", i, writeErr)
				m.Unlock()
				return
			}
			if closeErr := lz4Writer.Close(); closeErr != nil {
				logrus.Errorf("Error closing LZ4 writer for chunk %d: %v", i, closeErr)
				m.Lock()
				err = fmt.Errorf("failed to close LZ4 writer for chunk %d: %w", i, closeErr)
				m.Unlock()
				return
			}
			compressedChunks[i] = compressedData.Bytes()
			logrus.Infof("Compressed chunk %d, size: %d bytes", i, len(compressedChunks[i]))
		}(i, chunk)
	}
	wg.Wait()

	if err != nil {
		return nil, err
	}
	return compressedChunks, nil
}

// DeserializeAssignment deserializes a blob of bytes into a set of column assignments.
func DeserializeAssignment(filepath string, numChunks int) (WAssignment, error) {
	var (
		res  = collection.NewMapping[ifaces.ColID, smartvectors.SmartVector]()
		lock sync.Mutex
		err  error
	)

	logrus.Infof("Reading the assignment files")

	var wg sync.WaitGroup
	for i := 0; i < numChunks; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			chunkPath := fmt.Sprintf("%s_chunk_%d", filepath, i)
			chunkData, readErr := os.ReadFile(chunkPath)
			if readErr != nil {
				logrus.Errorf("Failed to read chunk %d: %v", i, readErr)
				lock.Lock()
				err = fmt.Errorf("failed to read chunk %d: %w", i, readErr)
				lock.Unlock()
				return
			}

			logrus.Infof("Reading chunk %d from %s, size: %d bytes", i, chunkPath, len(chunkData))
			lz4Reader := lz4.NewReader(bytes.NewReader(chunkData))
			var decompressedData bytes.Buffer
			n, readErr := decompressedData.ReadFrom(lz4Reader)
			if readErr != nil {
				logrus.Errorf("Error decompressing chunk %d: %v", i, readErr)
				lock.Lock()
				err = fmt.Errorf("failed to decompress chunk %d: %w", i, readErr)
				lock.Unlock()
				return
			}
			logrus.Infof("Decompressed chunk %d, size: %d bytes", i, n)

			var chunkMap map[string]*CompressedSmartVector
			if deserErr := deserializeAnyWithCborPkg(decompressedData.Bytes(), &chunkMap); deserErr != nil {
				logrus.Errorf("Error deserializing chunk %d: %v", i, deserErr)
				lock.Lock()
				err = fmt.Errorf("failed to deserialize chunk %d: %w", i, deserErr)
				lock.Unlock()
				return
			}

			for k, v := range chunkMap {
				decompressed := v.Decompress()
				lock.Lock()
				res.InsertNew(ifaces.ColID(k), decompressed)
				lock.Unlock()
			}
		}(i)
	}
	wg.Wait()

	if err != nil {
		return collection.NewMapping[ifaces.ColID, smartvectors.SmartVector](), err
	}

	logrus.Infof("Hash of WAssignment after deserialization: %s", hashWAssignment(res))
	return res, nil
}

// CompressedSmartVector represents a smartvectors.SmartVector in a space-efficient manner.
type CompressedSmartVector struct {
	F []CompressedSVFragment
}

// Serialize implements Serializable for CompressedSmartVector.
func (c *CompressedSmartVector) Serialize() (json.RawMessage, error) {
	return serializeAnyWithCborPkg(*c)
}

// Deserialize implements Serializable for CompressedSmartVector.
func (c *CompressedSmartVector) Deserialize(data json.RawMessage) error {
	return deserializeAnyWithCborPkg(data, c)
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

		// It's a right-padded value
		if offset == 0 {
			return &CompressedSmartVector{
				F: []CompressedSVFragment{
					newSliceSVFragment(w),
					newConstantSVFragment(paddingVal, fullLen-len(w)),
				},
			}
		}

		// It's a left-padded value
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

	// Padding comes first => Left padding
	if len(sv.F) == 2 && sv.F[0].isConstant() && sv.F[1].isPlain() {
		var (
			paddingVal = new(field.Element).SetBigInt(sv.F[0].X)
			window     = sv.F[1].readSlice()
			size       = sv.F[0].N + len(window)
		)
		return smartvectors.LeftPadded(window, *paddingVal, size)
	}

	// Padding comes later => Right padding
	if len(sv.F) == 2 && sv.F[0].isPlain() && sv.F[1].isConstant() {
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
	// that case the value of "n" cannot be inferred from the slice because the
	// slice will be empty. The solution is to provide a length to the vector
	if l == 0 {
		compressed.N = len(fv)
	}

	return compressed
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
