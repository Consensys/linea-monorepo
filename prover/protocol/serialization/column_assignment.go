package serialization

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"

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

// CompressedSmartVector represents a [smartvectors.SmartVector] in a more
// space-efficient manner.
type CompressedSmartVector struct {
	F []CompressedSVFragment
}

// CompressedSVFragment represent a portion of a SerializableSmartVector
type CompressedSVFragment struct {
	L uint8
	X *big.Int
	V []byte
	N int
}

// Hashing Functions

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

// Serialization Functions

// SerializeAssignment serializes map representing the column assignment of a wizard protocol.
func SerializeAssignment(a WAssignment, numChunks int) ([]json.RawMessage, error) {
	logrus.Infof("Hash of WAssignment before serialization: %s", hashWAssignment(a))

	ser := prepareSerializedMap(a)
	serSizeGB := calculateSizeInGB(ser)
	logrus.Infof("Size of ser : %.6f GB", serSizeGB)

	return parallelSerializeChunks(ser, numChunks)
}

// prepareSerializedMap prepares the serialized map from WAssignment.
func prepareSerializedMap(a WAssignment) map[string]*CompressedSmartVector {
	as := a.InnerMap()
	ser := make(map[string]*CompressedSmartVector)
	names := a.ListAllKeys()
	lock := &sync.Mutex{}

	parallel.ExecuteChunky(len(names), func(start, stop int) {
		for i := start; i < stop; i++ {
			v := CompressSmartVector(as[names[i]])
			lock.Lock()
			ser[fmt.Sprintf("%v", names[i])] = v
			lock.Unlock()
		}
	})

	return ser
}

// calculateSizeInGB calculates the size of the serialized map in GB.
func calculateSizeInGB(ser map[string]*CompressedSmartVector) float64 {
	var serSizeBytes uintptr
	for _, v := range ser {
		serSizeBytes += unsafe.Sizeof(*v)
	}
	return float64(serSizeBytes) / (1024 * 1024 * 1024)
}

// parallelSerializeChunks serializes the chunks in parallel.
func parallelSerializeChunks(ser map[string]*CompressedSmartVector, numChunks int) ([]json.RawMessage, error) {
	chunkSize := (len(ser) + numChunks - 1) / numChunks
	serializedChunks := make([]json.RawMessage, numChunks)
	var wg sync.WaitGroup
	var m sync.Mutex

	for i := 0; i < numChunks; i++ {
		wg.Add(1)
		go func(chunkIndex int) {
			defer wg.Done()
			chunk := prepareChunk(ser, chunkIndex, chunkSize)
			serializedChunk, err := serializeAnyWithCborPkg(chunk)
			if err != nil {
				return
			}
			logrus.Infof("Serialized chunk %d, size: %d bytes", chunkIndex, len(serializedChunk))
			m.Lock()
			serializedChunks[chunkIndex] = serializedChunk
			m.Unlock()
		}(i)
	}
	wg.Wait()

	return serializedChunks, nil
}

// prepareChunk prepares a chunk of data for serialization.
func prepareChunk(ser map[string]*CompressedSmartVector, chunkIndex, chunkSize int) map[string]*CompressedSmartVector {
	start := chunkIndex * chunkSize
	stop := start + chunkSize
	if stop > len(ser) {
		stop = len(ser)
	}

	chunk := make(map[string]*CompressedSmartVector)
	for k, v := range ser {
		if start <= 0 && stop > 0 {
			chunk[k] = v
		}
		start--
		stop--
	}

	return chunk
}

// Compression Functions

// CompressChunks compresses each serialized chunk.
func CompressChunks(chunks []json.RawMessage) []json.RawMessage {
	compressedChunks := make([]json.RawMessage, len(chunks))
	var wg sync.WaitGroup

	for i, chunk := range chunks {
		wg.Add(1)
		go func(i int, chunk json.RawMessage) {
			defer wg.Done()
			compressedChunks[i] = compressChunk(chunk)
			logrus.Infof("Compressed chunk %d, size: %d bytes", i, len(compressedChunks[i]))
		}(i, chunk)
	}
	wg.Wait()

	return compressedChunks
}

// compressChunk compresses a single chunk.
func compressChunk(chunk json.RawMessage) json.RawMessage {
	var compressedData bytes.Buffer
	lz4Writer := lz4.NewWriter(&compressedData)
	_, err := lz4Writer.Write(chunk)
	if err != nil {
		logrus.Errorf("Error compressing chunk: %v", err)
		panic(err)
	}
	lz4Writer.Close()
	return compressedData.Bytes()
}

// Deserialization Functions

// DeserializeAssignment deserializes a blob of bytes into a set of column assignments representing assigned columns of a Wizard protocol.
func DeserializeAssignment(filepath string, numChunks int) (WAssignment, error) {
	res := collection.NewMapping[ifaces.ColID, smartvectors.SmartVector]()
	lock := &sync.Mutex{}

	logrus.Infof("Reading the assignment files")

	var wg sync.WaitGroup
	for i := 0; i < numChunks; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			chunkData := readChunk(filepath, i)
			decompressedData := decompressChunk(chunkData)
			chunkMap := deserializeChunk(decompressedData)
			reconstructAssignment(chunkMap, res, lock)
		}(i)
	}
	wg.Wait()

	logrus.Infof("Hash of WAssignment after deserialization: %s", hashWAssignment(res))

	return res, nil
}

// readChunk reads a chunk from a file.
func readChunk(filepath string, chunkIndex int) []byte {
	chunkPath := fmt.Sprintf("%s_chunk_%d", filepath, chunkIndex)
	chunkData, err := os.ReadFile(chunkPath)
	if err != nil {
		logrus.Errorf("Failed to read chunk %d: %v", chunkIndex, err)
		return nil
	}
	logrus.Infof("Reading chunk %d from %s, size: %d bytes", chunkIndex, chunkPath, len(chunkData))
	return chunkData
}

// decompressChunk decompresses a chunk of data.
func decompressChunk(chunkData []byte) []byte {
	lz4Reader := lz4.NewReader(bytes.NewReader(chunkData))
	var decompressedData bytes.Buffer
	n, err := decompressedData.ReadFrom(lz4Reader)
	if err != nil {
		logrus.Errorf("Error decompressing chunk: %v", err)
		return nil
	}
	logrus.Infof("Decompressed chunk, size: %d bytes", n)
	return decompressedData.Bytes()
}

// deserializeChunk deserializes a chunk of data.
func deserializeChunk(decompressedData []byte) map[string]*CompressedSmartVector {
	var chunkMap map[string]*CompressedSmartVector
	if err := deserializeAnyWithCborPkg(decompressedData, &chunkMap); err != nil {
		logrus.Errorf("Error deserializing chunk: %v", err)
		return nil
	}
	return chunkMap
}

// reconstructAssignment reconstructs the WAssignment from the deserialized chunk.
func reconstructAssignment(chunkMap map[string]*CompressedSmartVector, res WAssignment, lock *sync.Mutex) {
	for k, v := range chunkMap {
		decompressed := v.Decompress()
		lock.Lock()
		res.InsertNew(ifaces.ColID(k), decompressed)
		lock.Unlock()
	}
}

// SmartVector Compression and Decompression

// CompressSmartVector compresses a SmartVector.
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
		return compressPaddedCircularWindow(v)
	default:
		return &CompressedSmartVector{
			F: []CompressedSVFragment{
				newSliceSVFragment(sv.IntoRegVecSaveAlloc()),
			},
		}
	}
}

// compressPaddedCircularWindow compresses a PaddedCircularWindow SmartVector.
func compressPaddedCircularWindow(v *smartvectors.PaddedCircularWindow) *CompressedSmartVector {
	w := v.Window()
	offset := v.Offset()
	paddingVal := v.PaddingVal()
	fullLen := v.Len()

	if offset == 0 {
		return &CompressedSmartVector{
			F: []CompressedSVFragment{
				newSliceSVFragment(w),
				newConstantSVFragment(paddingVal, fullLen-len(w)),
			},
		}
	}

	if offset+len(w) == fullLen {
		return &CompressedSmartVector{
			F: []CompressedSVFragment{
				newConstantSVFragment(paddingVal, fullLen-len(w)),
				newSliceSVFragment(w),
			},
		}
	}

	return nil
}

// Decompress decompresses a CompressedSmartVector.
func (sv *CompressedSmartVector) Decompress() smartvectors.SmartVector {
	if len(sv.F) == 1 && sv.F[0].isConstant() {
		val := new(field.Element).SetBigInt(sv.F[0].X)
		return smartvectors.NewConstant(*val, sv.F[0].N)
	}

	if len(sv.F) == 1 && sv.F[0].isPlain() {
		return smartvectors.NewRegular(sv.F[0].readSlice())
	}

	if len(sv.F) == 2 && sv.F[0].isConstant() && sv.F[1].isPlain() {
		paddingVal := new(field.Element).SetBigInt(sv.F[0].X)
		window := sv.F[1].readSlice()
		size := sv.F[0].N + len(window)
		return smartvectors.LeftPadded(window, *paddingVal, size)
	}

	if len(sv.F) == 2 && sv.F[1].isConstant() && sv.F[0].isPlain() {
		paddingVal := new(field.Element).SetBigInt(sv.F[1].X)
		window := sv.F[0].readSlice()
		size := sv.F[1].N + len(window)
		return smartvectors.RightPadded(window, *paddingVal, size)
	}

	panic("unexpected pattern")
}

// CompressedSVFragment Methods

func (f *CompressedSVFragment) isConstant() bool {
	return f.X != nil
}

func (f *CompressedSVFragment) isPlain() bool {
	return f.V != nil
}

func (f *CompressedSVFragment) readSlice() []field.Element {
	l := int(f.L)
	buf := bytes.NewBuffer(f.V)
	tmp := [32]byte{}
	n := f.N

	if l > 0 {
		n = len(f.V) / l
	}

	res := make([]field.Element, n)
	for i := range res {
		buf.Read(tmp[32-l:])
		res[i].SetBytes(tmp[:])
	}

	return res
}

// Helper Functions

func newConstantSVFragment(x field.Element, n int) CompressedSVFragment {
	var f big.Int
	_ = x.BigInt(&f)
	return CompressedSVFragment{
		X: &f,
		N: n,
	}
}

func newSliceSVFragment(fv []field.Element) CompressedSVFragment {
	l := 0
	for i := range fv {
		l = max(l, (fv[i].BitLen()+7)/8)
	}

	res := make([]byte, 0, len(fv)*l)
	resBuf := bytes.NewBuffer(res)
	for i := range fv {
		fbytes := fv[i].Bytes()
		resBuf.Write(fbytes[32-l:])
	}

	compressed := CompressedSVFragment{
		L: uint8(l),
		V: resBuf.Bytes(),
	}

	if l == 0 {
		compressed.N = len(fv)
	}

	return compressed
}
