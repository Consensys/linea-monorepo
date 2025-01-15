package serialization

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math/big"
	"sync"
	"time"
	"unsafe"

	"github.com/consensys/linea-monorepo/prover/maths/common/smartvectors"
	"github.com/consensys/linea-monorepo/prover/maths/field"
	"github.com/consensys/linea-monorepo/prover/protocol/ifaces"
	"github.com/consensys/linea-monorepo/prover/utils/collection"
	"github.com/consensys/linea-monorepo/prover/utils/parallel"
	"github.com/pierrec/lz4/v4"
)

// WAssignment is an alias for the mapping type used to represent the assignment
// of a column in a [wizard.ProverRuntime]
type WAssignment = collection.Mapping[ifaces.ColID, smartvectors.SmartVector]

// SerializeAssignment serializes map representing the column assignment of a
// wizard protocol.

func SerializeAssignment(a WAssignment) []byte {
	var (
		as    = a.InnerMap()
		ser   = map[string]*CompressedSmartVector{}
		names = a.ListAllKeys()
		lock  = &sync.Mutex{}
	)

	// Step 1: Measure time for parallel.ExecuteChunky
	start := time.Now()
	parallel.ExecuteChunky(len(names), func(start, stop int) {
		for i := start; i < stop; i++ {
			v := CompressSmartVector(as[names[i]])
			lock.Lock()
			// Convert names[i] to string using fmt.Sprintf or another method
			ser[fmt.Sprintf("%v", names[i])] = v
			lock.Unlock()
		}
	})
	fmt.Printf("Time taken for parallel.ExecuteChunky: %v\n", time.Since(start))

	// Calculate the approximate size of `ser` in bytes
	var serSizeBytes uintptr
	lock.Lock()
	for k, v := range ser {
		serSizeBytes += unsafe.Sizeof(k) + unsafe.Sizeof(*v)
	}
	lock.Unlock()

	// Convert size to GB
	serSizeGB := float64(serSizeBytes) / (1024 * 1024 * 1024)
	fmt.Printf("Size of ser (approximate): %.6f GB\n", serSizeGB)

	// Step 2: Parallelize CBOR serialization by chunking `ser`
	start = time.Now()
	numChunks := 50
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

			// Store the result in the slice
			m.Lock()
			serializedChunks[chunkIndex] = serializedChunk
			m.Unlock()
		}(i)
	}
	wg.Wait()
	fmt.Printf("Time taken for CBOR serialization in chunks: %v\n", time.Since(start))

	// Calculate the combined size of `serializedChunks` in GB
	totalCBORSize := 0
	for _, chunk := range serializedChunks {
		totalCBORSize += len(chunk)
	}
	cborDataSizeGB := float64(totalCBORSize) / (1024 * 1024 * 1024)
	fmt.Printf("Total size of CBOR serialized data: %.6f GB\n", cborDataSizeGB)

	// Step 3: Concatenate all serialized chunks and compress with LZ4
	start = time.Now()
	var compressedData bytes.Buffer
	lz4Writer := lz4.NewWriter(&compressedData)
	for _, chunk := range serializedChunks {
		_, err := lz4Writer.Write(chunk)
		if err != nil {
			panic(err) // handle error as needed
		}
	}
	lz4Writer.Close() // finalize the LZ4 stream
	fmt.Printf("Time taken for LZ4 compression: %v\n", time.Since(start))

	// Log size of `compressedData.Bytes()` in GB
	compressedDataSizeGB := float64(compressedData.Len()) / (1024 * 1024 * 1024)
	fmt.Printf("Size of compressedData.Bytes(): %.6f GB\n", compressedDataSizeGB)

	return compressedData.Bytes()
}

func SerializeAssignmentWithoutCompression(a WAssignment) []byte {
	var (
		as    = a.InnerMap()
		ser   = map[string]*CompressedSmartVector{}
		names = a.ListAllKeys()
		lock  = &sync.Mutex{}
	)

	// Step 1: Measure time for parallel.ExecuteChunky
	start := time.Now()
	parallel.ExecuteChunky(len(names), func(start, stop int) {
		for i := start; i < stop; i++ {
			v := CompressSmartVector(as[names[i]])
			lock.Lock()
			ser[fmt.Sprintf("%v", names[i])] = v
			lock.Unlock()
		}
	})
	fmt.Printf("Time taken for parallel.ExecuteChunky: %v\n", time.Since(start))

	// Log approximate size of `ser` in GB
	serSizeGB := float64(unsafe.Sizeof(ser)) / (1024 * 1024 * 1024)
	fmt.Printf("Size of ser (approximate): %.6f GB\n", serSizeGB)

	// Step 2: Parallelize CBOR serialization by chunking `ser`
	start = time.Now()
	chunkSize := (len(ser) + 49) / 50
	var serializedChunks = make([]json.RawMessage, 50)
	var wg sync.WaitGroup
	var m sync.Mutex
	for i := 0; i < 50; i++ {
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
				key := fmt.Sprintf("%v", names[j])
				chunk[key] = ser[key]
			}

			serializedChunk := serializeAnyWithCborPkg(chunk)

			m.Lock()
			serializedChunks[chunkIndex] = serializedChunk
			m.Unlock()
		}(i)
	}
	wg.Wait()
	fmt.Printf("Time taken for CBOR serialization in chunks : %v\n", time.Since(start))

	// Calculate the combined size of `serializedChunks` in GB
	totalCBORSize := 0
	for _, chunk := range serializedChunks {
		totalCBORSize += len(chunk)
	}
	cborDataSizeGB := float64(totalCBORSize) / (1024 * 1024 * 1024)
	fmt.Printf("Total size of CBOR serialized data: %.6f GB\n", cborDataSizeGB)

	// Step 3: Parallel concatenation of all serialized chunks
	start = time.Now()
	numConcatenators := 50
	chunksPerGroup := (len(serializedChunks) + numConcatenators - 1) / numConcatenators
	subResults := make([][]byte, numConcatenators)
	var concatWg sync.WaitGroup

	for i := 0; i < numConcatenators; i++ {
		concatWg.Add(1)
		go func(groupIndex int) {
			defer concatWg.Done()

			// Concatenate this group of chunks
			start := groupIndex * chunksPerGroup
			stop := start + chunksPerGroup
			if stop > len(serializedChunks) {
				stop = len(serializedChunks)
			}

			var groupResult []byte
			for _, chunk := range serializedChunks[start:stop] {
				groupResult = append(groupResult, chunk...)
			}
			subResults[groupIndex] = groupResult
		}(i)
	}
	concatWg.Wait()

	// Final concatenation of subResults
	var finalResult []byte
	for _, subResult := range subResults {
		finalResult = append(finalResult, subResult...)
	}
	fmt.Printf("Time taken for parallel concatenation: %v\n", time.Since(start))

	return finalResult
}

// DeserializeAssignment deserialize a blob of bytes into a set of column
// assignment representing assigned columns of a Wizard protocol.
func DeserializeAssignment(data []byte) (WAssignment, error) {

	var (
		ser  = map[string]*CompressedSmartVector{}
		res  = collection.NewMapping[ifaces.ColID, smartvectors.SmartVector]()
		lock = &sync.Mutex{}
	)

	if err := deserializeAnyWithCborPkg(data, &ser); err != nil {
		return WAssignment{}, err
	}

	names := make([]string, 0, len(ser))
	for n := range ser {
		names = append(names, n)
	}

	parallel.ExecuteChunky(len(names), func(start, stop int) {
		for i := start; i < stop; i++ {
			v := ser[names[i]].Decompress()
			lock.Lock()
			res.InsertNew(ifaces.ColID(names[i]), v)
			lock.Unlock()
		}
	})

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
