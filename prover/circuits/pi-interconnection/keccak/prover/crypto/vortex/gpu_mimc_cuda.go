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
	gpuSISSplitMinRows  = 256

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
	if len(v) < piGPUSISMinRows() && !shouldAttemptSplitSISGPU(v) {
		return nil, nil, false
	}
	tree, colHashes, err := buildSISMiMCTreeGPUSplitFromRows(v, key)
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

func piGPUSISSplitMinRows() int {
	raw := os.Getenv("LINEA_PROVER_GPU_PI_SIS_SPLIT_MIN_ROWS")
	if raw == "" {
		return gpuSISSplitMinRows
	}
	v, err := strconv.Atoi(raw)
	if err != nil || v < 0 {
		logrus.WithField("value", raw).Warn("invalid LINEA_PROVER_GPU_PI_SIS_SPLIT_MIN_ROWS; using default")
		return gpuSISSplitMinRows
	}
	return v
}

func shouldAttemptSplitSISGPU(v EncodedMatrix) bool {
	if len(v) < piGPUSISSplitMinRows() {
		return false
	}
	if os.Getenv("LINEA_PROVER_GPU_PI_DISABLE_SECONDARY_DEVICE") != "" {
		return false
	}
	if len(v) == 0 || v[0].Len() < 2 || !utils.IsPowerOfTwo(v[0].Len()) {
		return false
	}
	_, primaryID, err := piPrimaryGPUDevice()
	if err != nil {
		logrus.WithError(err).Warn("PI Vortex GPU primary device unavailable for SIS split threshold")
		return false
	}
	_, ok, err := piSecondaryGPUDeviceID(primaryID)
	if err != nil {
		logrus.WithError(err).Warn("PI Vortex GPU secondary device unavailable for SIS split threshold")
		return false
	}
	return ok
}

func buildSISMiMCTreeGPU(colHashes []field.Element, chunkSize int) (*smt.Tree, error) {
	tree, _, err := buildMiMCTreeGPUFromChunks(colHashes, chunkSize)
	return tree, err
}

func buildSISMiMCTreeGPUFromRows(v EncodedMatrix, key *ringsis.Key) (*smt.Tree, []field.Element, error) {
	dev, _, err := piPrimaryGPUDevice()
	if err != nil {
		return nil, nil, err
	}
	return buildSISMiMCTreeGPUFromRowsOnDevice(dev, v, key)
}

func buildSISMiMCTreeGPUSplitFromRows(v EncodedMatrix, key *ringsis.Key) (*smt.Tree, []field.Element, error) {
	if os.Getenv("LINEA_PROVER_GPU_PI_DISABLE_SECONDARY_DEVICE") != "" {
		return buildSISMiMCTreeGPUFromRows(v, key)
	}
	primary, primaryID, err := piPrimaryGPUDevice()
	if err != nil {
		return nil, nil, err
	}
	secondaryID, ok, err := piSecondaryGPUDeviceID(primaryID)
	if err != nil {
		return nil, nil, err
	}
	if !ok {
		return buildSISMiMCTreeGPUFromRowsOnDevice(primary, v, key)
	}
	if len(v) == 0 || v[0].Len() < 2 {
		return buildSISMiMCTreeGPUFromRowsOnDevice(primary, v, key)
	}

	numCols := v[0].Len()
	if !utils.IsPowerOfTwo(numCols) {
		return buildSISMiMCTreeGPUFromRowsOnDevice(primary, v, key)
	}
	split := numCols / 2
	leftRows, rightRows, err := splitSISRows(v, split)
	if err != nil {
		return nil, nil, err
	}

	secondary := gpu.GetDeviceN(secondaryID)
	if secondary == nil {
		return nil, nil, fmt.Errorf("secondary GPU device %d is unavailable", secondaryID)
	}
	logrus.Infof(
		"PI Vortex GPU SIS split across devices: primary=%d secondary=%d rows=%d cols=%d split=%d",
		primaryID, secondaryID, len(v), numCols, split,
	)

	type result struct {
		tree      *smt.Tree
		colHashes []field.Element
		err       error
	}
	leftCh := make(chan result, 1)
	rightCh := make(chan result, 1)

	go func() {
		runtime.LockOSThread()
		defer runtime.UnlockOSThread()
		tree, colHashes, err := buildSISMiMCTreeGPUFromRowsOnDevice(primary, leftRows, key)
		leftCh <- result{tree: tree, colHashes: colHashes, err: err}
	}()
	go func() {
		runtime.LockOSThread()
		defer runtime.UnlockOSThread()
		tree, colHashes, err := buildSISMiMCTreeGPUFromRowsOnDevice(secondary, rightRows, key)
		rightCh <- result{tree: tree, colHashes: colHashes, err: err}
	}()

	left := <-leftCh
	right := <-rightCh
	if left.err != nil {
		return nil, nil, fmt.Errorf("primary split SIS commitment: %w", left.err)
	}
	if right.err != nil {
		return nil, nil, fmt.Errorf("secondary split SIS commitment: %w", right.err)
	}

	tree, err := mergeSplitSISMiMCTrees(left.tree, right.tree)
	if err != nil {
		return nil, nil, err
	}
	colHashes := make([]field.Element, 0, len(left.colHashes)+len(right.colHashes))
	colHashes = append(colHashes, left.colHashes...)
	colHashes = append(colHashes, right.colHashes...)
	return tree, colHashes, nil
}

func buildSISMiMCTreeGPUFromRowsOnDevice(dev *gpu.Device, v EncodedMatrix, key *ringsis.Key) (*smt.Tree, []field.Element, error) {
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

	if dev == nil {
		return nil, nil, fmt.Errorf("GPU device is unavailable")
	}
	if err := dev.Bind(); err != nil {
		return nil, nil, fmt.Errorf("bind GPU device %d: %w", dev.DeviceID(), err)
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

func piPrimaryGPUDevice() (*gpu.Device, int, error) {
	dev, deviceID, err := gpu.DeviceFromEnvOrCurrent()
	if err != nil {
		return nil, 0, err
	}
	if dev == nil {
		dev = gpu.GetDevice()
		if dev == nil {
			return nil, 0, fmt.Errorf("GPU device is unavailable")
		}
		deviceID = dev.DeviceID()
		if err := dev.Bind(); err != nil {
			return nil, 0, fmt.Errorf("bind GPU device %d: %w", deviceID, err)
		}
	}
	return dev, deviceID, nil
}

func piSecondaryGPUDeviceID(primaryID int) (int, bool, error) {
	raw := os.Getenv("LINEA_PROVER_GPU_PI_SECONDARY_DEVICE_ID")
	if raw != "" {
		id, err := strconv.Atoi(raw)
		if err != nil {
			return 0, false, fmt.Errorf("invalid LINEA_PROVER_GPU_PI_SECONDARY_DEVICE_ID %q: %w", raw, err)
		}
		if id < 0 {
			return 0, false, fmt.Errorf("LINEA_PROVER_GPU_PI_SECONDARY_DEVICE_ID must be non-negative, got %d", id)
		}
		if id == primaryID {
			return 0, false, fmt.Errorf("PI secondary device matches primary device %d", primaryID)
		}
		return id, true, nil
	}

	n := gpu.PhysicalDeviceCount()
	if n < 2 {
		return 0, false, nil
	}
	return (primaryID + 1) % n, true, nil
}

func splitSISRows(v EncodedMatrix, split int) (EncodedMatrix, EncodedMatrix, error) {
	left := make(EncodedMatrix, len(v))
	right := make(EncodedMatrix, len(v))
	for i := range v {
		if v[i].Len() <= split || split <= 0 {
			return nil, nil, fmt.Errorf("invalid split %d for row %d length %d", split, i, v[i].Len())
		}
		switch row := v[i].(type) {
		case *smartvectors.Regular:
			left[i] = smartvectors.NewRegular((*row)[:split])
			right[i] = smartvectors.NewRegular((*row)[split:])
		case *smartvectors.Constant:
			left[i] = smartvectors.NewConstant(row.Value, split)
			right[i] = smartvectors.NewConstant(row.Value, row.Len()-split)
		default:
			return nil, nil, fmt.Errorf("unsupported smart vector row %d of type %T", i, v[i])
		}
	}
	return left, right, nil
}

func mergeSplitSISMiMCTrees(left, right *smt.Tree) (*smt.Tree, error) {
	if left == nil || right == nil {
		return nil, fmt.Errorf("cannot merge nil split SIS tree")
	}
	if left.Config == nil || right.Config == nil {
		return nil, fmt.Errorf("cannot merge split SIS tree without config")
	}
	if left.Config.Depth != right.Config.Depth {
		return nil, fmt.Errorf("split SIS tree depth mismatch: %d != %d", left.Config.Depth, right.Config.Depth)
	}
	if len(left.OccupiedLeaves) != len(right.OccupiedLeaves) {
		return nil, fmt.Errorf("split SIS leaf count mismatch: %d != %d", len(left.OccupiedLeaves), len(right.OccupiedLeaves))
	}
	if left.Config.Depth == 0 {
		return nil, fmt.Errorf("split SIS tree depth is too small to merge")
	}

	depth := left.Config.Depth + 1
	tree := smt.NewEmptyTree(&smt.Config{HashFunc: hashtypes.MiMC, Depth: depth})
	tree.OccupiedLeaves = make([]types.Bytes32, 0, len(left.OccupiedLeaves)+len(right.OccupiedLeaves))
	tree.OccupiedLeaves = append(tree.OccupiedLeaves, left.OccupiedLeaves...)
	tree.OccupiedLeaves = append(tree.OccupiedLeaves, right.OccupiedLeaves...)

	tree.OccupiedNodes = make([][]types.Bytes32, depth-1)
	for level := 0; level < left.Config.Depth-1; level++ {
		tree.OccupiedNodes[level] = make([]types.Bytes32, 0, len(left.OccupiedNodes[level])+len(right.OccupiedNodes[level]))
		tree.OccupiedNodes[level] = append(tree.OccupiedNodes[level], left.OccupiedNodes[level]...)
		tree.OccupiedNodes[level] = append(tree.OccupiedNodes[level], right.OccupiedNodes[level]...)
	}
	tree.OccupiedNodes[depth-2] = []types.Bytes32{left.Root, right.Root}

	hasher := mimc.NewMiMC()
	left.Root.WriteTo(hasher)
	right.Root.WriteTo(hasher)
	tree.Root = types.AsBytes32(hasher.Sum(nil))
	return tree, nil
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
