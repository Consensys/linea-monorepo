//go:build cuda

package vortex

import (
	"testing"

	"github.com/consensys/gnark-crypto/ecc/bls12-377/fr/mimc"
	"github.com/consensys/linea-monorepo/prover/circuits/pi-interconnection/keccak/prover/crypto/ringsis"
	"github.com/consensys/linea-monorepo/prover/circuits/pi-interconnection/keccak/prover/crypto/state-management/smt"
	"github.com/consensys/linea-monorepo/prover/circuits/pi-interconnection/keccak/prover/maths/common/smartvectors"
	"github.com/consensys/linea-monorepo/prover/circuits/pi-interconnection/keccak/prover/maths/field"
	"github.com/consensys/linea-monorepo/prover/circuits/pi-interconnection/keccak/prover/utils/types"
	"github.com/consensys/linea-monorepo/prover/gpu"
	"github.com/stretchr/testify/require"
)

func TestBuildSISMiMCTreeGPUvsCPU(t *testing.T) {
	const (
		numLeaves = 1024
		chunkSize = 64
	)

	colHashes := make([]field.Element, numLeaves*chunkSize)
	for i := range colHashes {
		colHashes[i].SetUint64(uint64(i*17 + 3))
	}

	gpuTree, err := buildSISMiMCTreeGPU(colHashes, chunkSize)
	require.NoError(t, err)

	leaves := hashSisHashForTest(colHashes, chunkSize)
	cpuTree := smt.BuildCompleteMiMC(leaves)

	require.Equal(t, cpuTree.Root, gpuTree.Root, "GPU MiMC tree root should match CPU")
	require.Equal(t, cpuTree.OccupiedLeaves, gpuTree.OccupiedLeaves, "GPU MiMC leaves should match CPU")
	require.Equal(t, cpuTree.OccupiedNodes, gpuTree.OccupiedNodes, "GPU MiMC internal nodes should match CPU")
}

func TestBuildNoSISMiMCTreeGPUvsCPU(t *testing.T) {
	const (
		numRows = 9
		numCols = 1024
	)

	rows := make([]smartvectors.SmartVector, numRows)
	for i := range rows {
		if i%4 == 0 {
			rows[i] = smartvectors.NewConstant(field.NewElement(uint64(i+11)), numCols)
			continue
		}
		rows[i] = smartvectors.Rand(numCols)
	}

	params := NewParams(1, numCols, numRows, ringsis.StdParams)
	gpuTree, gpuColHashes, err := buildNoSISMiMCTreeGPU(rows)
	require.NoError(t, err)

	cpuColHashes := params.noSisTransversalHash(rows)
	leaves := make([]types.Bytes32, len(cpuColHashes))
	for i := range leaves {
		leaves[i] = cpuColHashes[i].Bytes()
	}
	cpuTree := smt.BuildCompleteMiMC(leaves)

	require.Equal(t, cpuColHashes, gpuColHashes, "GPU no-SIS column hashes should match CPU")
	require.Equal(t, cpuTree.Root, gpuTree.Root, "GPU no-SIS MiMC tree root should match CPU")
	require.Equal(t, cpuTree.OccupiedLeaves, gpuTree.OccupiedLeaves, "GPU no-SIS MiMC leaves should match CPU")
	require.Equal(t, cpuTree.OccupiedNodes, gpuTree.OccupiedNodes, "GPU no-SIS MiMC internal nodes should match CPU")
}

func TestBuildSISMiMCTreeGPUFromRowsVsCPU(t *testing.T) {
	const (
		numRows = 19
		numCols = 1024
	)

	rows := make([]smartvectors.SmartVector, numRows)
	for i := range rows {
		if i%5 == 0 {
			rows[i] = smartvectors.NewConstant(field.NewElement(uint64(i+7)), numCols)
			continue
		}
		rows[i] = smartvectors.Rand(numCols)
	}

	params := NewParams(1, numCols, numRows, ringsis.StdParams)
	gpuTree, gpuColHashes, err := buildSISMiMCTreeGPUFromRows(EncodedMatrix(rows), params.Key)
	require.NoError(t, err)

	cpuColHashes := params.Key.TransversalHash(rows)
	cpuTree := smt.BuildCompleteMiMC(params.hashSisHash(cpuColHashes))

	require.Equal(t, cpuColHashes, gpuColHashes, "GPU SIS column hashes should match CPU")
	require.Equal(t, cpuTree.Root, gpuTree.Root, "GPU SIS MiMC tree root should match CPU")
	require.Equal(t, cpuTree.OccupiedLeaves, gpuTree.OccupiedLeaves, "GPU SIS MiMC leaves should match CPU")
	require.Equal(t, cpuTree.OccupiedNodes, gpuTree.OccupiedNodes, "GPU SIS MiMC internal nodes should match CPU")
}

func TestMergeSplitSISMiMCTreesMatchesFullTree(t *testing.T) {
	const (
		numLeaves = 1024
		chunkSize = 64
	)

	colHashes := make([]field.Element, numLeaves*chunkSize)
	for i := range colHashes {
		colHashes[i].SetUint64(uint64(i*19 + 5))
	}

	fullTree, err := buildSISMiMCTreeGPU(colHashes, chunkSize)
	require.NoError(t, err)

	split := len(colHashes) / 2
	leftTree, err := buildSISMiMCTreeGPU(colHashes[:split], chunkSize)
	require.NoError(t, err)
	rightTree, err := buildSISMiMCTreeGPU(colHashes[split:], chunkSize)
	require.NoError(t, err)

	merged, err := mergeSplitSISMiMCTrees(leftTree, rightTree)
	require.NoError(t, err)
	require.Equal(t, fullTree.Root, merged.Root, "merged split tree root should match the full tree")
	require.Equal(t, fullTree.OccupiedLeaves, merged.OccupiedLeaves, "merged leaves should match the full tree")
	require.Equal(t, fullTree.OccupiedNodes, merged.OccupiedNodes, "merged internal nodes should match the full tree")
}

func TestBuildSISMiMCTreeGPUSplitFromRowsVsCPU(t *testing.T) {
	if gpu.PhysicalDeviceCount() < 2 {
		t.Skip("requires two visible CUDA devices")
	}

	const (
		numRows = 19
		numCols = 2048
	)

	rows := make([]smartvectors.SmartVector, numRows)
	for i := range rows {
		if i%5 == 0 {
			rows[i] = smartvectors.NewConstant(field.NewElement(uint64(i+7)), numCols)
			continue
		}
		rows[i] = smartvectors.Rand(numCols)
	}

	params := NewParams(1, numCols, numRows, ringsis.StdParams)
	gpuTree, gpuColHashes, err := buildSISMiMCTreeGPUSplitFromRows(
		EncodedMatrix(rows),
		params.Key,
	)
	require.NoError(t, err)

	cpuColHashes := params.Key.TransversalHash(rows)
	cpuTree := smt.BuildCompleteMiMC(params.hashSisHash(cpuColHashes))

	require.Equal(t, cpuColHashes, gpuColHashes, "split GPU SIS column hashes should match CPU")
	require.Equal(t, cpuTree.Root, gpuTree.Root, "split GPU SIS MiMC tree root should match CPU")
	require.Equal(t, cpuTree.OccupiedLeaves, gpuTree.OccupiedLeaves, "split GPU SIS MiMC leaves should match CPU")
	require.Equal(t, cpuTree.OccupiedNodes, gpuTree.OccupiedNodes, "split GPU SIS MiMC internal nodes should match CPU")
}

func hashSisHashForTest(colHashes []field.Element, chunkSize int) []types.Bytes32 {
	numChunks := len(colHashes) / chunkSize
	leaves := make([]types.Bytes32, numChunks)
	for chunkID := 0; chunkID < numChunks; chunkID++ {
		startChunk := chunkID * chunkSize
		hasher := mimc.NewFieldHasher()
		digest := hasher.SumElements(colHashes[startChunk : startChunk+chunkSize])
		leaves[chunkID] = digest.Bytes()
	}
	return leaves
}

func BenchmarkSISMiMCTreeProductionCPU(b *testing.B) {
	benchmarkSISMiMCTree(b, false)
}

func BenchmarkSISMiMCTreeProductionGPU(b *testing.B) {
	benchmarkSISMiMCTree(b, true)
}

func BenchmarkSISMiMCTreeFromRowsProductionGPU_812Rows(b *testing.B) {
	benchmarkSISMiMCTreeFromRows(b, 812, false)
}

func BenchmarkSISMiMCTreeFromRowsProductionGPU_812RegularRows(b *testing.B) {
	benchmarkSISMiMCTreeFromRows(b, 812, true)
}

func BenchmarkSISMiMCTreeFromRowsProductionGPU_288RegularRows(b *testing.B) {
	benchmarkSISMiMCTreeFromRows(b, 288, true)
}

func BenchmarkSISMiMCTreeFromRowsProductionGPU_108RegularRows(b *testing.B) {
	benchmarkSISMiMCTreeFromRows(b, 108, true)
}

func BenchmarkSISMiMCTreeFromRowsProductionGPU_1880RegularRows(b *testing.B) {
	benchmarkSISMiMCTreeFromRows(b, 1880, true)
}

func BenchmarkSISMiMCTreeFromRowsProductionGPUSplit_1880RegularRows(b *testing.B) {
	benchmarkSISMiMCTreeFromRowsSplit(b, 1880, true)
}

func BenchmarkSISMiMCTreeFromRowsProductionGPUSplit_812RegularRows(b *testing.B) {
	benchmarkSISMiMCTreeFromRowsSplit(b, 812, true)
}

func BenchmarkSISMiMCTreeFromRowsProductionGPUSplit_288RegularRows(b *testing.B) {
	benchmarkSISMiMCTreeFromRowsSplit(b, 288, true)
}

func BenchmarkSISMiMCTreeFromRowsProductionGPUSplit_108RegularRows(b *testing.B) {
	benchmarkSISMiMCTreeFromRowsSplit(b, 108, true)
}

func benchmarkSISMiMCTree(b *testing.B, useGPU bool) {
	const (
		blowUpFactor = 2
		numColumns   = 1 << 18
		maxRows      = 1880
	)

	params := NewParams(blowUpFactor, numColumns, maxRows, ringsis.StdParams)
	numLeaves := params.NumEncodedCols()
	chunkSize := params.Key.OutputSize()
	colHashes := make([]field.Element, numLeaves*chunkSize)
	for i := range colHashes {
		colHashes[i].SetUint64(uint64(i*17 + 3))
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if useGPU {
			tree, err := buildSISMiMCTreeGPU(colHashes, chunkSize)
			require.NoError(b, err)
			require.NotEqual(b, types.Bytes32{}, tree.Root)
			continue
		}
		leaves := params.hashSisHash(colHashes)
		tree := smt.BuildCompleteMiMC(leaves)
		require.NotEqual(b, types.Bytes32{}, tree.Root)
	}
}

func benchmarkSISMiMCTreeFromRows(b *testing.B, numRows int, regularRows bool) {
	const numCols = 1 << 19

	rows := make([]smartvectors.SmartVector, numRows)
	for i := range rows {
		if regularRows {
			row := make([]field.Element, numCols)
			for j := range row {
				row[j].SetUint64(uint64(i + j + 1))
			}
			rows[i] = smartvectors.NewRegular(row)
			continue
		}
		rows[i] = smartvectors.NewConstant(field.NewElement(uint64(i+1)), numCols)
	}

	params := NewParams(1, numCols, numRows, ringsis.StdParams)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		tree, colHashes, err := buildSISMiMCTreeGPUFromRows(
			EncodedMatrix(rows),
			params.Key,
		)
		require.NoError(b, err)
		require.NotEqual(b, types.Bytes32{}, tree.Root)
		require.Len(b, colHashes, numCols*params.Key.OutputSize())
	}
}

func benchmarkSISMiMCTreeFromRowsSplit(b *testing.B, numRows int, regularRows bool) {
	if gpu.PhysicalDeviceCount() < 2 {
		b.Skip("requires two visible CUDA devices")
	}

	const numCols = 1 << 19

	rows := make([]smartvectors.SmartVector, numRows)
	for i := range rows {
		if regularRows {
			row := make([]field.Element, numCols)
			for j := range row {
				row[j].SetUint64(uint64(i + j + 1))
			}
			rows[i] = smartvectors.NewRegular(row)
			continue
		}
		rows[i] = smartvectors.NewConstant(field.NewElement(uint64(i+1)), numCols)
	}

	params := NewParams(1, numCols, numRows, ringsis.StdParams)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		tree, colHashes, err := buildSISMiMCTreeGPUSplitFromRows(
			EncodedMatrix(rows),
			params.Key,
		)
		require.NoError(b, err)
		require.NotEqual(b, types.Bytes32{}, tree.Root)
		require.Len(b, colHashes, numCols*params.Key.OutputSize())
	}
}
