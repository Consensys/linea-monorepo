//go:build cuda

package vortex

/*
#cgo LDFLAGS: -L${SRCDIR}/../../../../../../gpu/cuda/build -lgnark_gpu -L/usr/local/cuda/lib64 -lcudart -lstdc++ -lm
#cgo CFLAGS: -I${SRCDIR}/../../../../../../gpu/cuda/include

#include "gnark_gpu.h"
*/
import "C"

import (
	"fmt"
	"os"
	"runtime"
	"strconv"
	"sync"
	"unsafe"

	"github.com/consensys/gnark-crypto/ecc/bls12-377/fr"
	"github.com/consensys/gnark-crypto/ecc/bls12-377/fr/fft"
	"github.com/consensys/gnark-crypto/ecc/bls12-377/fr/mimc"
	"github.com/consensys/linea-monorepo/prover/circuits/pi-interconnection/keccak/prover/crypto/ringsis"
	"github.com/consensys/linea-monorepo/prover/circuits/pi-interconnection/keccak/prover/crypto/state-management/hashtypes"
	"github.com/consensys/linea-monorepo/prover/circuits/pi-interconnection/keccak/prover/crypto/state-management/smt"
	"github.com/consensys/linea-monorepo/prover/circuits/pi-interconnection/keccak/prover/maths/common/smartvectors"
	"github.com/consensys/linea-monorepo/prover/circuits/pi-interconnection/keccak/prover/maths/field"
	"github.com/consensys/linea-monorepo/prover/circuits/pi-interconnection/keccak/prover/utils"
	"github.com/consensys/linea-monorepo/prover/circuits/pi-interconnection/keccak/prover/utils/parallel"
	"github.com/consensys/linea-monorepo/prover/circuits/pi-interconnection/keccak/prover/utils/types"
	"github.com/consensys/linea-monorepo/prover/gpu"
	"github.com/sirupsen/logrus"
)

const (
	gpuSISDegree        = 64
	gpuSISLogTwoBound   = 16
	gpuSISLimbsPerField = 16
	gpuSISMinRows       = 512

	gpuSISRowRegular  = 0
	gpuSISRowConstant = 1
)

type gpuSISStaticData struct {
	twiddles       []field.Element
	twiddlesInv    []field.Element
	coset          []field.Element
	cosetInv       []field.Element
	cardinalityInv []field.Element
	mimcConstants  []field.Element
}

type gpuSISKeyCacheKey struct {
	key      *ringsis.Key
	numPolys int
}

var (
	gpuSISStaticOnce sync.Once
	gpuSISStatic     gpuSISStaticData
	gpuSISStaticErr  error

	gpuSISKeyCacheMu sync.Mutex
	gpuSISKeyCache   = map[gpuSISKeyCacheKey][]field.Element{}
)

func tryCommitSISGPU(v EncodedMatrix, key *ringsis.Key) (*smt.Tree, []field.Element, bool) {
	if !usePIGPUSIS() {
		return nil, nil, false
	}
	if len(v) < piGPUSISMinRows() {
		return nil, nil, false
	}
	tree, colHashes, err := buildSISMiMCTreeGPUFromRows(v, key)
	if err != nil {
		logrus.WithError(err).Warn("PI Vortex GPU SIS failed; falling back to CPU")
		return nil, nil, false
	}
	return tree, colHashes, true
}

func tryBuildSISMiMCTreeGPU(colHashes []field.Element, chunkSize int) (*smt.Tree, bool) {
	if !usePIGPUMiMC() {
		return nil, false
	}
	tree, err := buildSISMiMCTreeGPU(colHashes, chunkSize)
	if err != nil {
		logrus.WithError(err).Warn("PI Vortex GPU MiMC tree failed; falling back to CPU")
		return nil, false
	}
	return tree, true
}

func tryCommitNoSISMiMCGPU(v []smartvectors.SmartVector) (*smt.Tree, []field.Element, bool) {
	if !usePIGPUMiMC() {
		return nil, nil, false
	}
	tree, colHashes, err := buildNoSISMiMCTreeGPU(v)
	if err != nil {
		logrus.WithError(err).Warn("PI Vortex GPU no-SIS MiMC failed; falling back to CPU")
		return nil, nil, false
	}
	return tree, colHashes, true
}

func usePIGPUMiMC() bool {
	return os.Getenv("LINEA_PROVER_GPU_PI_VORTEX") == "1" ||
		os.Getenv("LINEA_PROVER_GPU_PI_MIMC") == "1"
}

func usePIGPUSIS() bool {
	return os.Getenv("LINEA_PROVER_GPU_PI_VORTEX") == "1" ||
		os.Getenv("LINEA_PROVER_GPU_PI_SIS") == "1"
}

func piGPUSISMinRows() int {
	raw := os.Getenv("LINEA_PROVER_GPU_PI_SIS_MIN_ROWS")
	if raw == "" {
		return gpuSISMinRows
	}
	v, err := strconv.Atoi(raw)
	if err != nil || v < 0 {
		logrus.WithField("value", raw).Warn("invalid LINEA_PROVER_GPU_PI_SIS_MIN_ROWS; using default")
		return gpuSISMinRows
	}
	return v
}

func buildSISMiMCTreeGPU(colHashes []field.Element, chunkSize int) (*smt.Tree, error) {
	tree, _, err := buildMiMCTreeGPUFromChunks(colHashes, chunkSize)
	return tree, err
}

func buildSISMiMCTreeGPUFromRows(v EncodedMatrix, key *ringsis.Key) (*smt.Tree, []field.Element, error) {
	if len(v) == 0 {
		return nil, nil, fmt.Errorf("empty matrix")
	}
	if key.LogTwoDegree != utils.Log2Ceil(gpuSISDegree) || key.LogTwoBound != gpuSISLogTwoBound {
		return nil, nil, fmt.Errorf(
			"unsupported SIS params: degree=%d logTwoBound=%d",
			key.OutputSize(),
			key.LogTwoBound,
		)
	}

	numRows := len(v)
	numCols := v[0].Len()
	if numCols == 0 {
		return nil, nil, fmt.Errorf("matrix has zero columns")
	}
	for i := range v {
		if v[i].Len() != numCols {
			return nil, nil, fmt.Errorf("row %d has length %d, expected %d", i, v[i].Len(), numCols)
		}
	}
	if !utils.IsPowerOfTwo(numCols) {
		return nil, nil, fmt.Errorf("numCols=%d is not a power of two", numCols)
	}
	if numRows > key.MaxNumFieldToHash {
		return nil, nil, fmt.Errorf("numRows=%d exceeds SIS key capacity %d", numRows, key.MaxNumFieldToHash)
	}

	rowPtrs := make([]uintptr, numRows)
	rowKinds := make([]uint8, numRows)
	rowConstants := make([]field.Element, numRows)
	for i := range v {
		switch vi := v[i].(type) {
		case *smartvectors.Regular:
			rowKinds[i] = gpuSISRowRegular
			rowPtrs[i] = uintptr(unsafe.Pointer(&(*vi)[0]))
		case *smartvectors.Constant:
			rowKinds[i] = gpuSISRowConstant
			rowConstants[i] = vi.Value
		default:
			return nil, nil, fmt.Errorf("unsupported smart vector row %d of type %T", i, v[i])
		}
	}

	numPolys := utils.DivCeil(numRows*gpuSISLimbsPerField, gpuSISDegree)
	ag, err := cachedFlattenSISKey(key, numPolys)
	if err != nil {
		return nil, nil, err
	}
	static, err := getGPUSISStaticData()
	if err != nil {
		return nil, nil, err
	}

	dev := gpu.GetDevice()
	if dev == nil {
		return nil, nil, fmt.Errorf("GPU device is unavailable")
	}
	if err := dev.Bind(); err != nil {
		return nil, nil, fmt.Errorf("bind GPU device: %w", err)
	}

	colHashes := make([]field.Element, numCols*gpuSISDegree)
	nodes := make([]field.Element, 2*numCols-1)

	errCode := C.gnark_gpu_bls12377_sis_mimc_tree(
		C.gnark_gpu_context_t(dev.Handle()),
		(*C.uintptr_t)(unsafe.Pointer(&rowPtrs[0])),
		(*C.uint8_t)(unsafe.Pointer(&rowKinds[0])),
		(*C.uint64_t)(unsafe.Pointer(&rowConstants[0])),
		C.size_t(numRows),
		C.size_t(numCols),
		(*C.uint64_t)(unsafe.Pointer(&ag[0])),
		C.size_t(numPolys),
		(*C.uint64_t)(unsafe.Pointer(&static.twiddles[0])),
		(*C.uint64_t)(unsafe.Pointer(&static.twiddlesInv[0])),
		(*C.uint64_t)(unsafe.Pointer(&static.coset[0])),
		(*C.uint64_t)(unsafe.Pointer(&static.cosetInv[0])),
		(*C.uint64_t)(unsafe.Pointer(&static.cardinalityInv[0])),
		(*C.uint64_t)(unsafe.Pointer(&static.mimcConstants[0])),
		(*C.uint64_t)(unsafe.Pointer(&colHashes[0])),
		(*C.uint64_t)(unsafe.Pointer(&nodes[0])),
	)
	runtime.KeepAlive(v)
	runtime.KeepAlive(rowPtrs)
	runtime.KeepAlive(rowKinds)
	runtime.KeepAlive(rowConstants)
	runtime.KeepAlive(ag)
	runtime.KeepAlive(static)
	if errCode != C.GNARK_GPU_SUCCESS {
		return nil, nil, fmt.Errorf("gnark_gpu_bls12377_sis_mimc_tree: %s", gpuErrorString(errCode))
	}

	return bottomUpMiMCTreeFromField(nodes, numCols), colHashes, nil
}

func buildNoSISMiMCTreeGPU(v []smartvectors.SmartVector) (*smt.Tree, []field.Element, error) {
	if len(v) == 0 {
		return nil, nil, fmt.Errorf("empty matrix")
	}

	numRows := len(v)
	numCols := v[0].Len()
	if numCols == 0 {
		return nil, nil, fmt.Errorf("matrix has zero columns")
	}
	for i := range v {
		if v[i].Len() != numCols {
			return nil, nil, fmt.Errorf("row %d has length %d, expected %d", i, v[i].Len(), numCols)
		}
	}
	if !utils.IsPowerOfTwo(numCols) {
		return nil, nil, fmt.Errorf("numCols=%d is not a power of two", numCols)
	}

	columnChunks := make([]field.Element, numCols*numRows)
	parallel.Execute(numCols, func(start, stop int) {
		for col := start; col < stop; col++ {
			dst := columnChunks[col*numRows : (col+1)*numRows]
			for row := range v {
				switch vi := v[row].(type) {
				case *smartvectors.Constant:
					dst[row] = vi.Value
				case *smartvectors.Regular:
					dst[row] = (*vi)[col]
				default:
					dst[row] = v[row].Get(col)
				}
			}
		}
	})

	return buildMiMCTreeGPUFromChunks(columnChunks, numRows)
}

func buildMiMCTreeGPUFromChunks(chunks []field.Element, chunkSize int) (*smt.Tree, []field.Element, error) {
	if chunkSize <= 0 {
		return nil, nil, fmt.Errorf("invalid chunk size %d", chunkSize)
	}
	if len(chunks) == 0 || len(chunks)%chunkSize != 0 {
		return nil, nil, fmt.Errorf("input length %d is not a multiple of chunk size %d", len(chunks), chunkSize)
	}
	numLeaves := len(chunks) / chunkSize
	if !utils.IsPowerOfTwo(numLeaves) {
		return nil, nil, fmt.Errorf("numLeaves=%d is not a power of two", numLeaves)
	}

	dev := gpu.GetDevice()
	if dev == nil {
		return nil, nil, fmt.Errorf("GPU device is unavailable")
	}
	if err := dev.Bind(); err != nil {
		return nil, nil, fmt.Errorf("bind GPU device: %w", err)
	}

	static, err := getGPUSISStaticData()
	if err != nil {
		return nil, nil, err
	}

	totalNodes := 2*numLeaves - 1
	nodes := make([]field.Element, totalNodes)

	errCode := C.gnark_gpu_bls12377_mimc_sis_tree(
		C.gnark_gpu_context_t(dev.Handle()),
		(*C.uint64_t)(unsafe.Pointer(&chunks[0])),
		C.size_t(numLeaves),
		C.size_t(chunkSize),
		(*C.uint64_t)(unsafe.Pointer(&static.mimcConstants[0])),
		(*C.uint64_t)(unsafe.Pointer(&nodes[0])),
	)
	runtime.KeepAlive(chunks)
	runtime.KeepAlive(static)
	if errCode != C.GNARK_GPU_SUCCESS {
		return nil, nil, fmt.Errorf("gnark_gpu_bls12377_mimc_sis_tree: %s", gpuErrorString(errCode))
	}

	leaves := append([]field.Element(nil), nodes[:numLeaves]...)
	return bottomUpMiMCTreeFromField(nodes, numLeaves), leaves, nil
}

func cachedFlattenSISKey(key *ringsis.Key, numPolys int) ([]field.Element, error) {
	cacheKey := gpuSISKeyCacheKey{key: key, numPolys: numPolys}

	gpuSISKeyCacheMu.Lock()
	if ag, ok := gpuSISKeyCache[cacheKey]; ok {
		gpuSISKeyCacheMu.Unlock()
		return ag, nil
	}
	gpuSISKeyCacheMu.Unlock()

	ag, err := flattenSISKey(key, numPolys)
	if err != nil {
		return nil, err
	}

	gpuSISKeyCacheMu.Lock()
	if cached, ok := gpuSISKeyCache[cacheKey]; ok {
		gpuSISKeyCacheMu.Unlock()
		return cached, nil
	}
	gpuSISKeyCache[cacheKey] = ag
	gpuSISKeyCacheMu.Unlock()
	return ag, nil
}

func flattenSISKey(key *ringsis.Key, numPolys int) ([]field.Element, error) {
	agByPoly := key.Ag()
	if numPolys > len(agByPoly) {
		return nil, fmt.Errorf("numPolys=%d exceeds SIS key length %d", numPolys, len(agByPoly))
	}
	ag := make([]field.Element, numPolys*gpuSISDegree)
	for i := 0; i < numPolys; i++ {
		if len(agByPoly[i]) != gpuSISDegree {
			return nil, fmt.Errorf("SIS key polynomial %d has length %d", i, len(agByPoly[i]))
		}
		copy(ag[i*gpuSISDegree:(i+1)*gpuSISDegree], agByPoly[i])
	}
	return ag, nil
}

func getGPUSISStaticData() (gpuSISStaticData, error) {
	gpuSISStaticOnce.Do(func() {
		twiddles, twiddlesInv, coset, cosetInv, cardinalityInv, err := sisFFTTables()
		if err != nil {
			gpuSISStaticErr = err
			return
		}
		constants, err := mimcConstants()
		if err != nil {
			gpuSISStaticErr = err
			return
		}
		gpuSISStatic = gpuSISStaticData{
			twiddles:       twiddles,
			twiddlesInv:    twiddlesInv,
			coset:          coset,
			cosetInv:       cosetInv,
			cardinalityInv: cardinalityInv,
			mimcConstants:  constants,
		}
	})
	return gpuSISStatic, gpuSISStaticErr
}

func sisFFTTables() (
	twiddles []field.Element,
	twiddlesInv []field.Element,
	coset []field.Element,
	cosetInv []field.Element,
	cardinalityInv []field.Element,
	err error,
) {
	shift, err := fr.Generator(2 * gpuSISDegree)
	if err != nil {
		return nil, nil, nil, nil, nil, err
	}
	domain := fft.NewDomain(gpuSISDegree, fft.WithShift(shift))

	twiddlesByStage, err := domain.Twiddles()
	if err != nil {
		return nil, nil, nil, nil, nil, err
	}
	twiddlesInvByStage, err := domain.TwiddlesInv()
	if err != nil {
		return nil, nil, nil, nil, nil, err
	}
	twiddles, err = flattenSISTwiddles(twiddlesByStage)
	if err != nil {
		return nil, nil, nil, nil, nil, err
	}
	twiddlesInv, err = flattenSISTwiddles(twiddlesInvByStage)
	if err != nil {
		return nil, nil, nil, nil, nil, err
	}
	coset, err = domain.CosetTable()
	if err != nil {
		return nil, nil, nil, nil, nil, err
	}
	cosetInv, err = domain.CosetTableInv()
	if err != nil {
		return nil, nil, nil, nil, nil, err
	}
	cardinalityInv = []field.Element{domain.CardinalityInv}
	return twiddles, twiddlesInv, coset, cosetInv, cardinalityInv, nil
}

func flattenSISTwiddles(twiddlesByStage [][]field.Element) ([]field.Element, error) {
	const numStages = 6
	if len(twiddlesByStage) != numStages {
		return nil, fmt.Errorf("unexpected SIS twiddle stage count %d", len(twiddlesByStage))
	}
	res := make([]field.Element, 0, 69)
	for stage := 0; stage < numStages; stage++ {
		expectedLen := 1 + (gpuSISDegree >> (stage + 1))
		if len(twiddlesByStage[stage]) != expectedLen {
			return nil, fmt.Errorf(
				"unexpected SIS twiddle stage %d length %d, expected %d",
				stage,
				len(twiddlesByStage[stage]),
				expectedLen,
			)
		}
		res = append(res, twiddlesByStage[stage]...)
	}
	return res, nil
}

func mimcConstants() ([]field.Element, error) {
	bigConstants := mimc.GetConstants()
	if len(bigConstants) != 62 {
		return nil, fmt.Errorf("unexpected MiMC constant count: %d", len(bigConstants))
	}
	constants := make([]field.Element, len(bigConstants))
	for i := range bigConstants {
		constants[i].SetBigInt(&bigConstants[i])
	}
	return constants, nil
}

func bottomUpMiMCTreeFromField(nodes []field.Element, numLeaves int) *smt.Tree {
	depth := utils.Log2Ceil(numLeaves)
	tree := smt.NewEmptyTree(&smt.Config{HashFunc: hashtypes.MiMC, Depth: depth})

	tree.OccupiedLeaves = make([]types.Bytes32, numLeaves)
	copyFieldElementsAsBytes(tree.OccupiedLeaves, nodes[:numLeaves])

	offset := numLeaves
	if depth == 0 {
		tree.Root = nodes[0].Bytes()
		return tree
	}

	tree.OccupiedNodes = make([][]types.Bytes32, depth-1)
	for level := 1; level < depth; level++ {
		levelSize := numLeaves >> level
		tree.OccupiedNodes[level-1] = make([]types.Bytes32, levelSize)
		copyFieldElementsAsBytes(
			tree.OccupiedNodes[level-1],
			nodes[offset:offset+levelSize],
		)
		offset += levelSize
	}

	tree.Root = nodes[len(nodes)-1].Bytes()
	return tree
}

func copyFieldElementsAsBytes(dst []types.Bytes32, src []field.Element) {
	const parallelThreshold = 4096
	if len(dst) != len(src) {
		utils.Panic("mismatched byte conversion lengths: dst=%d src=%d", len(dst), len(src))
	}
	if len(src) < parallelThreshold {
		for i := range src {
			dst[i] = src[i].Bytes()
		}
		return
	}
	parallel.Execute(len(src), func(start, stop int) {
		for i := start; i < stop; i++ {
			dst[i] = src[i].Bytes()
		}
	})
}

func gpuErrorString(code C.gnark_gpu_error_t) string {
	switch code {
	case C.GNARK_GPU_ERROR_CUDA:
		return "CUDA error"
	case C.GNARK_GPU_ERROR_INVALID_ARG:
		return "invalid argument"
	case C.GNARK_GPU_ERROR_OUT_OF_MEMORY:
		return "out of GPU memory"
	case C.GNARK_GPU_ERROR_SIZE_MISMATCH:
		return "size mismatch"
	default:
		return fmt.Sprintf("unknown error code %d", int(code))
	}
}
