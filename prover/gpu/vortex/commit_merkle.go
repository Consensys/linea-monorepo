// GPU-accelerated Vortex commit for the protocol compiler.
//
// Two modes:
//
//   CommitMerkleWithSIS     — legacy drop-in, D2H of full encoded matrix.
//   CommitSIS               — device-resident: encoded matrix stays on GPU.
//
// CommitSIS returns a *CommitState handle. The protocol compiler stores it
// in prover state and later calls:
//
//   cs.LinComb(α)            → UAlpha[j] = Σᵢ αⁱ · row[i][j]   (GPU kernel)
//   cs.ExtractColumns(cols)  → selected columns only             (small D2H)
//
// This eliminates the ~8 GiB D2H transfer of the full encoded matrix.

//go:build cuda

package vortex

import (
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/consensys/gnark-crypto/field/koalabear"
	refvortex "github.com/consensys/gnark-crypto/field/koalabear/vortex"
	"github.com/consensys/linea-monorepo/prover/crypto/state-management/smt_koalabear"
	"github.com/consensys/linea-monorepo/prover/crypto/vortex/vortex_koalabear"
	"github.com/consensys/linea-monorepo/prover/gpu"
	"github.com/consensys/linea-monorepo/prover/maths/common/smartvectors"
	"github.com/consensys/linea-monorepo/prover/maths/field"
	"github.com/sirupsen/logrus"
)

// gpuVortexCache caches GPUVortex instances keyed by (nCols, maxNRows, rate).
// Avoids re-allocating ~12 GB of GPU memory per commit call.
var (
	gpuVortexMu    sync.Mutex
	gpuVortexCache = map[gpuVortexKey]*GPUVortex{}

	gpuBenchTimings = os.Getenv("LINEA_GPU_BENCH_TIMINGS") != ""
)

type gpuVortexKey struct {
	nCols int
	nRows int
	rate  int
}

// EvictPipelineCache frees all cached GPUVortex pipelines, reclaiming GPU memory.
// Call after a recursion level's OpenSelectedColumns to free pipelines that won't
// be reused (each level uses different parameters).
func EvictPipelineCache() {
	gpuVortexMu.Lock()
	defer gpuVortexMu.Unlock()
	for key, gv := range gpuVortexCache {
		gv.Free()
		delete(gpuVortexCache, key)
	}
}

func getOrCreateGPUVortex(dev *gpu.Device, params *Params, maxNRows int) (*GPUVortex, error) {
	key := gpuVortexKey{
		nCols: params.inner.NbColumns,
		nRows: maxNRows,
		rate:  params.inner.ReedSolomonInvRate,
	}
	gpuVortexMu.Lock()
	defer gpuVortexMu.Unlock()

	if gv, ok := gpuVortexCache[key]; ok {
		return gv, nil
	}
	gv, err := NewGPUVortex(dev, params, maxNRows)
	if err != nil {
		return nil, err
	}
	gpuVortexCache[key] = gv
	return gv, nil
}

func materializeRows(polysMatrix []smartvectors.SmartVector) [][]koalabear.Element {
	rows := make([][]koalabear.Element, len(polysMatrix))
	for i := range polysMatrix {
		if reg, ok := polysMatrix[i].(*smartvectors.Regular); ok {
			rows[i] = []koalabear.Element(*reg)
			continue
		}
		rows[i] = smartvectors.IntoRegVec(polysMatrix[i])
	}
	return rows
}

func initGPUForRows(params *vortex_koalabear.Params, maxRows int) (*gpu.Device, *Params, *GPUVortex) {
	dev := gpu.GetDevice()
	if dev == nil {
		return nil, nil, nil
	}
	gpuParams, err := NewParams(params.NbColumns, maxRows, params.Key.SisGnarkCrypto, params.RsParams.Rate, 256)
	if err != nil {
		logrus.WithError(err).Warn("GPU params init failed")
		return nil, nil, nil
	}
	gv, err := getOrCreateGPUVortex(dev, gpuParams, maxRows)
	if err != nil {
		logrus.WithError(err).Warn("GPU vortex init failed")
		return nil, nil, nil
	}
	return dev, gpuParams, gv
}

func initGPU(params *vortex_koalabear.Params) (*gpu.Device, *Params, *GPUVortex) {
	return initGPUForRows(params, params.MaxNbRows)
}

func writeSmartVectorRow(s smartvectors.SmartVector, dst []koalabear.Element) {
	if reg, ok := s.(*smartvectors.Regular); ok {
		copy(dst, []koalabear.Element(*reg))
		return
	}
	s.WriteInSlice(dst)
}

func logGPUTiming(format string, args ...any) {
	if gpuBenchTimings {
		fmt.Printf("[gpu/vortex] "+format+"\n", args...)
	}
}

// PreWarmGPU creates the GPU pipeline during compilation so that the
// first Prove() call doesn't pay the ~5s initialization cost.
// Safe to call multiple times or when GPU is unavailable (no-op).
func PreWarmGPU(params *vortex_koalabear.Params) {
	if params == nil {
		return
	}
	initGPU(params)
}

// PreWarmGPUForRows eagerly initializes a GPU pipeline sized for `maxRows`.
// Useful to avoid first-use latency in prover steps that commit matrices with
// round-specific row counts.
func PreWarmGPUForRows(params *vortex_koalabear.Params, maxRows int) {
	if params == nil || maxRows <= 0 {
		return
	}
	initGPUForRows(params, maxRows)
}

// cloneSMTTreeFromRef converts a gnark-crypto Merkle tree into an smt_koalabear
// tree without rehashing, by deep-copying tree levels into SMT layout.
func cloneSMTTreeFromRef(src *refvortex.MerkleTree) *smt_koalabear.Tree {
	if src == nil || len(src.Levels) == 0 {
		return nil
	}
	depth := src.Depth()

	leaves := make([]field.Octuplet, len(src.Levels[depth]))
	for i := range leaves {
		leaves[i] = field.Octuplet(src.Levels[depth][i])
	}

	occupiedNodes := make([][]field.Octuplet, depth-1)
	for level := 1; level < depth; level++ {
		// SMT level 0 is just above leaves; ref tree stores levels from root.
		srcLevel := src.Levels[depth-level]
		dstLevel := make([]field.Octuplet, len(srcLevel))
		for i := range dstLevel {
			dstLevel[i] = field.Octuplet(srcLevel[i])
		}
		occupiedNodes[level-1] = dstLevel
	}

	return &smt_koalabear.Tree{
		Depth:          depth,
		Root:           field.Octuplet(src.Levels[0][0]),
		OccupiedLeaves: leaves,
		OccupiedNodes:  occupiedNodes,
		// Full trees never hit empty nodes while proving openings.
		EmptyNodes: make([]field.Octuplet, depth-1),
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// CommitSIS — GPU-accelerated commit with host-resident result
// ─────────────────────────────────────────────────────────────────────────────

// CommitSIS commits on GPU and keeps the encoded matrix on device.
//
// Returns:
//   - *CommitState: device-resident handle for LinComb + ExtractColumns
//   - *smt_koalabear.Tree: Merkle tree (host-side)
//   - []field.Element: SIS column hashes (nil if needSISHashes is false)
func CommitSIS(
	params *vortex_koalabear.Params,
	polysMatrix []smartvectors.SmartVector,
	needSISHashes bool,
) (*CommitState, *smt_koalabear.Tree, []field.Element) {
	tAll := time.Now()
	var (
		tInit, tCommit, tTree, tSIS time.Duration
		usedCPUFallback             bool
	)

	defer func() {
		logGPUTiming(
			"CommitSIS rows=%d cols=%d needSIS=%t cpuFallback=%t init=%v commit=%v tree=%v sis=%v total=%v",
			len(polysMatrix), params.NbColumns, needSISHashes, usedCPUFallback,
			tInit, tCommit, tTree, tSIS, time.Since(tAll),
		)
	}()

	if params.Key.LogTwoBound() != gpuSISLogTwoBound {
		usedCPUFallback = true
		return commitSISCPU(params, polysMatrix)
	}

	t0 := time.Now()
	dev, _, gv := initGPU(params)
	tInit = time.Since(t0)
	if gv == nil {
		usedCPUFallback = true
		return commitSISCPU(params, polysMatrix)
	}

	t0 = time.Now()
	cs, _, err := gv.CommitDirect(len(polysMatrix), func(i int, dst []koalabear.Element) {
		writeSmartVectorRow(polysMatrix[i], dst)
	})
	tCommit = time.Since(t0)
	if err != nil {
		logrus.WithError(err).Warn("GPU CommitSIS failed, falling back to CPU")
		usedCPUFallback = true
		return commitSISCPU(params, polysMatrix)
	}

	t0 = time.Now()
	tree := cloneSMTTreeFromRef(cs.MerkleTree())
	tTree = time.Since(t0)
	if tree == nil {
		logrus.Warn("GPU CommitSIS tree conversion failed, falling back to CPU")
		usedCPUFallback = true
		return commitSISCPU(params, polysMatrix)
	}

	// Extract SIS hashes before downloading encoded matrix (both read from
	// the shared pipeline's device buffers).
	var sisHashes []field.Element
	if needSISHashes {
		t0 = time.Now()
		sisHashes, err = cs.ExtractSISHashes()
		tSIS = time.Since(t0)
		if err != nil {
			logrus.WithError(err).Warn("GPU SIS hash extraction failed, falling back to CPU")
			usedCPUFallback = true
			return commitSISCPU(params, polysMatrix)
		}
	}

	// D2D snapshot: copy encoded matrix to per-round GPU buffer (~35x faster
	// than D2H). The GPU pipeline is cached and shared across SIS rounds —
	// each CommitDirect call overwrites d_encoded_col. The snapshot decouples
	// this round's data from the shared pipeline.
	if err := cs.SnapshotEncoded(dev); err != nil {
		logrus.WithError(err).Warn("GPU snapshot failed, falling back to CPU")
		usedCPUFallback = true
		return commitSISCPU(params, polysMatrix)
	}
	return cs, tree, sisHashes
}

// commitSISCPU is the CPU fallback. Commits on CPU and wraps the result.
func commitSISCPU(
	params *vortex_koalabear.Params,
	polysMatrix []smartvectors.SmartVector,
) (*CommitState, *smt_koalabear.Tree, []field.Element) {
	encoded, _, tree, colHashes := params.CommitMerkleWithSIS(polysMatrix)
	cs := &CommitState{encodedMatrix: encoded, nRows: len(polysMatrix)}
	return cs, tree, colHashes
}

// ─────────────────────────────────────────────────────────────────────────────
// CommitMerkleWithSIS — legacy drop-in (full D2H of encoded matrix)
// ─────────────────────────────────────────────────────────────────────────────

// gpuSISLogTwoBound is the SIS LogTwoBound value the GPU kernels are validated
// for. The SIS decomposition produces ceil(31/LogTwoBound) limbs per element;
// changing LogTwoBound alters buffer sizes and kernel indexing. Fall back to
// CPU for other values until the CUDA kernels are generalized.
const gpuSISLogTwoBound = 16

// CommitMerkleWithSIS is a GPU-accelerated drop-in replacement for
// vortex_koalabear.Params.CommitMerkleWithSIS. Returns the full encoded
// matrix on host (D2H transfer). Use CommitSIS for device-resident mode.
func CommitMerkleWithSIS(
	params *vortex_koalabear.Params,
	polysMatrix []smartvectors.SmartVector,
) (vortex_koalabear.EncodedMatrix, vortex_koalabear.Commitment, *smt_koalabear.Tree, []field.Element) {

	if params.Key.LogTwoBound() != gpuSISLogTwoBound {
		return params.CommitMerkleWithSIS(polysMatrix)
	}

	_, _, gv := initGPU(params)
	if gv == nil {
		logrus.Warn("GPU not available, falling back to CPU CommitMerkleWithSIS")
		return params.CommitMerkleWithSIS(polysMatrix)
	}

	rows := materializeRows(polysMatrix)
	encodedRows, colHashes, _, _, gpuTree, err := gv.CommitAndExtract(rows)
	if err != nil {
		logrus.WithError(err).Warn("GPU commit failed, falling back to CPU")
		return params.CommitMerkleWithSIS(polysMatrix)
	}

	encodedMatrix := make(vortex_koalabear.EncodedMatrix, len(rows))
	for i := range encodedRows {
		encodedMatrix[i] = smartvectors.NewRegular(encodedRows[i])
	}
	tree := cloneSMTTreeFromRef(gpuTree)
	if tree == nil {
		logrus.Warn("GPU Merkle tree conversion failed, falling back to CPU")
		return params.CommitMerkleWithSIS(polysMatrix)
	}
	return encodedMatrix, tree.Root, tree, colHashes
}
