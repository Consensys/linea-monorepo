// Tests and benchmarks for GPU CommitMerkleWithSIS drop-in replacement.
//
// Compares GPU vortex.CommitMerkleWithSIS against the CPU
// vortex_koalabear.Params.CommitMerkleWithSIS using the same SmartVector
// inputs (mix of Constant and Regular SmartVectors, matching production).
//
// Parameters from protocol/compiler/standard_benchmark_test.go:
//   - RS rate = 2, SIS(logTwoDegree=9, logTwoBound=16)
//   - Matrix ~1B cells (1<<19 cols × 1<<11 rows)
//   - ~15% constant rows (mimics SmartVector constant columns in production)

//go:build cuda

package vortex

import (
	"fmt"
	"math/rand/v2"
	"testing"
	"time"

	"github.com/consensys/gnark-crypto/field/koalabear"
	"github.com/consensys/gnark-crypto/field/koalabear/sis"
	"github.com/consensys/linea-monorepo/prover/crypto/vortex/vortex_koalabear"
	"github.com/consensys/linea-monorepo/prover/gpu"
	"github.com/consensys/linea-monorepo/prover/maths/common/smartvectors"
	"github.com/consensys/linea-monorepo/prover/maths/field"
	"github.com/stretchr/testify/require"
)

const constFraction = 0.15 // ~15% constant rows

// randSmartVectorMatrix creates nRows SmartVectors of length nCols.
// ~constFraction are Constant SmartVectors, the rest are Regular.
func randSmartVectorMatrix(rng *rand.Rand, nRows, nCols int) []smartvectors.SmartVector {
	m := make([]smartvectors.SmartVector, nRows)
	nConst := int(float64(nRows) * constFraction)
	for i := range m {
		if i < nConst {
			v := randKB(rng)
			m[i] = smartvectors.NewConstant(field.Element(v), nCols)
		} else {
			row := make([]field.Element, nCols)
			for j := range row {
				row[j] = field.Element(randKB(rng))
			}
			m[i] = smartvectors.NewRegular(row)
		}
	}
	return m
}

// randMatrixMixed creates nRows × nCols with ~constFraction constant rows.
func randMatrixMixed(rng *rand.Rand, nRows, nCols int) [][]koalabear.Element {
	m := make([][]koalabear.Element, nRows)
	nConst := int(float64(nRows) * constFraction)
	for i := range m {
		m[i] = make([]koalabear.Element, nCols)
		if i < nConst {
			v := randKB(rng)
			for j := range m[i] {
				m[i][j] = v
			}
		} else {
			for j := range m[i] {
				m[i][j] = randKB(rng)
			}
		}
	}
	return m
}

// ─── Drop-in CommitMerkleWithSIS: GPU vs CPU with SmartVectors ──────────────

func TestCommitMerkleWithSIS_GPUvsCPU(t *testing.T) {
	assert := require.New(t)
	rng := rand.New(rand.NewChaCha8([32]byte{0xCA}))

	nCols := 1024
	nRows := 128
	rate := 2

	params := vortex_koalabear.NewParams(rate, nCols, nRows, 9, 16)
	m := randSmartVectorMatrix(rng, nRows, nCols)

	// CPU
	cpuEncoded, cpuCommit, cpuTree, cpuHashes := params.CommitMerkleWithSIS(m)

	// GPU
	gpuEncoded, gpuCommit, gpuTree, gpuHashes := CommitMerkleWithSIS(&params, m)

	// Compare commitments (Merkle roots)
	assert.Equal(cpuCommit, gpuCommit, "GPU commitment ≠ CPU commitment")

	// Compare tree roots
	assert.Equal(cpuTree.Root, gpuTree.Root, "GPU tree root ≠ CPU tree root")

	// Compare SIS column hashes
	assert.Equal(len(cpuHashes), len(gpuHashes), "SIS hash length mismatch")
	for i := range cpuHashes {
		assert.Equal(cpuHashes[i], gpuHashes[i], "SIS hash mismatch at index %d", i)
	}

	// Compare encoded matrices
	assert.Equal(len(cpuEncoded), len(gpuEncoded), "encoded matrix row count mismatch")
	for i := range cpuEncoded {
		cpuRow := smartvectors.IntoRegVec(cpuEncoded[i])
		gpuRow := smartvectors.IntoRegVec(gpuEncoded[i])
		assert.Equal(len(cpuRow), len(gpuRow), "encoded row %d length mismatch", i)
		for j := range cpuRow {
			assert.Equal(cpuRow[j], gpuRow[j], "encoded matrix mismatch at [%d][%d]", i, j)
		}
	}
}

func TestCommitMerkleWithSIS_Diagnostic(t *testing.T) {
	rng := rand.New(rand.NewChaCha8([32]byte{0xCA}))

	nCols := 256
	nRows := 32
	rate := 2

	params := vortex_koalabear.NewParams(rate, nCols, nRows, 9, 16)
	m := randSmartVectorMatrix(rng, nRows, nCols)

	// CPU path
	cpuEncoded, _, _, cpuHashes := params.CommitMerkleWithSIS(m)

	// GPU path (using CommitMerkleWithSIS drop-in)
	gpuEncoded, _, _, gpuHashes := CommitMerkleWithSIS(&params, m)

	// Step 1: Compare encoded matrices row by row
	encMismatch := 0
	for i := 0; i < len(cpuEncoded); i++ {
		cpuRow := smartvectors.IntoRegVec(cpuEncoded[i])
		gpuRow := smartvectors.IntoRegVec(gpuEncoded[i])
		if len(cpuRow) != len(gpuRow) {
			t.Errorf("row %d length: CPU=%d GPU=%d", i, len(cpuRow), len(gpuRow))
			continue
		}
		for j := range cpuRow {
			if cpuRow[j] != gpuRow[j] {
				if encMismatch < 3 {
					t.Logf("encoded[%d][%d]: CPU=0x%x GPU=0x%x", i, j, cpuRow[j][0], gpuRow[j][0])
				}
				encMismatch++
			}
		}
	}
	t.Logf("Encoded matrix mismatches: %d / %d", encMismatch, len(cpuEncoded)*nCols*rate)

	// Step 2: Compare SIS column hashes
	sisMismatch := 0
	minLen := len(cpuHashes)
	if len(gpuHashes) < minLen {
		minLen = len(gpuHashes)
	}
	t.Logf("SIS hash count: CPU=%d GPU=%d", len(cpuHashes), len(gpuHashes))
	for i := 0; i < minLen; i++ {
		if cpuHashes[i] != gpuHashes[i] {
			if sisMismatch < 3 {
				t.Logf("SIS[%d]: CPU=0x%x GPU=0x%x", i, cpuHashes[i][0], gpuHashes[i][0])
			}
			sisMismatch++
		}
	}
	t.Logf("SIS hash mismatches: %d / %d", sisMismatch, minLen)
}

func TestCommitMerkleWithSIS_GPUvsCPU_Rate4(t *testing.T) {
	testCommitMerkleWithSISRate(t, 4)
}

func TestCommitMerkleWithSIS_GPUvsCPU_Rate8(t *testing.T) {
	testCommitMerkleWithSISRate(t, 8)
}

func TestCommitMerkleWithSIS_GPUvsCPU_Rate16(t *testing.T) {
	testCommitMerkleWithSISRate(t, 16)
}

func testCommitMerkleWithSISRate(t *testing.T, rate int) {
	assert := require.New(t)
	rng := rand.New(rand.NewChaCha8([32]byte{byte(rate)}))

	nCols := 256
	nRows := 32

	params := vortex_koalabear.NewParams(rate, nCols, nRows, 9, 16)
	m := randSmartVectorMatrix(rng, nRows, nCols)

	// CPU
	_, cpuCommit, cpuTree, cpuHashes := params.CommitMerkleWithSIS(m)

	// GPU
	_, gpuCommit, gpuTree, gpuHashes := CommitMerkleWithSIS(&params, m)

	assert.Equal(cpuCommit, gpuCommit, "GPU commitment ≠ CPU commitment (rate=%d)", rate)
	assert.Equal(cpuTree.Root, gpuTree.Root, "GPU tree root ≠ CPU tree root (rate=%d)", rate)
	assert.Equal(len(cpuHashes), len(gpuHashes), "SIS hash length mismatch (rate=%d)", rate)
	for i := range cpuHashes {
		assert.Equal(cpuHashes[i], gpuHashes[i], "SIS hash mismatch at %d (rate=%d)", i, rate)
	}
}

// ─── Low-level GPU correctness tests (RS encoding + SIS + columns) ──────────

func TestGPUEncodeAndSIS(t *testing.T) {
	assert := require.New(t)
	rng := rand.New(rand.NewChaCha8([32]byte{0xCA}))
	dev := newTestDevice(t)

	nCols := 1024
	nRows := 128
	rate := 2
	nSelected := 32

	sisParams, err := sis.NewRSis(0, 9, 16, nRows)
	assert.NoError(err)

	params, err := NewParams(nCols, nRows, sisParams, rate, nSelected)
	assert.NoError(err)

	m := randMatrixMixed(rng, nRows, nCols)

	// CPU RS encode
	cpuEncoded := make([][]koalabear.Element, nRows)
	for i := range m {
		cpuEncoded[i] = make([]koalabear.Element, nCols*rate)
		params.EncodeReedSolomon(m[i], cpuEncoded[i])
	}

	// GPU commit + extract
	gv, err := NewGPUVortex(dev, params, nRows)
	assert.NoError(err)
	defer gv.Free()

	cs, _, err := gv.Commit(m)
	assert.NoError(err)

	gpuRows, err := cs.ExtractAllRows()
	assert.NoError(err)

	// Compare RS-encoded rows
	assert.Equal(len(cpuEncoded), len(gpuRows), "row count mismatch")
	for i := range cpuEncoded {
		for j := range cpuEncoded[i] {
			assert.Equal(cpuEncoded[i][j], gpuRows[i][j], "encoded[%d][%d] mismatch", i, j)
		}
	}

	// Compare SIS hashes
	gpuSIS, err := cs.ExtractSISHashes()
	assert.NoError(err)

	degree := sisParams.Degree
	scw := nCols * rate
	cpuSIS := make([]koalabear.Element, scw*degree)
	for col := 0; col < scw; col++ {
		column := make([]koalabear.Element, nRows)
		for row := 0; row < nRows; row++ {
			column[row] = cpuEncoded[row][col]
		}
		sisParams.Hash(column, cpuSIS[col*degree:(col+1)*degree])
	}
	assert.Equal(len(cpuSIS), len(gpuSIS), "SIS hash length mismatch")
	for i := range cpuSIS {
		assert.Equal(cpuSIS[i], gpuSIS[i], "SIS[%d] mismatch", i)
	}

	// Compare leaves (GPU MD hash vs CPU CompressPoseidon2x16)
	gpuLeaves, err := cs.ExtractLeaves()
	assert.NoError(err)

	cpuLeaves := make([]Hash, scw)
	n16 := scw / 16
	for c := 0; c < n16; c++ {
		start := c * 16 * degree
		CompressPoseidon2x16(gpuSIS[start:start+16*degree], degree, cpuLeaves[c*16:(c+1)*16])
	}
	for i := 0; i < scw; i++ {
		assert.Equal(cpuLeaves[i], gpuLeaves[i], "leaf[%d] mismatch", i)
	}
}

func TestGPUColumnExtraction(t *testing.T) {
	assert := require.New(t)
	rng := rand.New(rand.NewChaCha8([32]byte{0xBE}))
	dev := newTestDevice(t)

	nCols := 256
	nRows := 32
	rate := 2
	nSelected := 8

	sisParams, err := sis.NewRSis(0, 9, 16, nRows)
	assert.NoError(err)

	params, err := NewParams(nCols, nRows, sisParams, rate, nSelected)
	assert.NoError(err)

	m := randMatrixMixed(rng, nRows, nCols)

	gv, err := NewGPUVortex(dev, params, nRows)
	assert.NoError(err)
	defer gv.Free()

	cs, _, err := gv.Commit(m)
	assert.NoError(err)

	allRows, err := cs.ExtractAllRows()
	assert.NoError(err)

	// Extract specific columns and verify against full matrix
	selectedCols := make([]int, nSelected)
	for i := range selectedCols {
		selectedCols[i] = rng.IntN(nCols*rate - 1)
	}
	cols, err := cs.ExtractColumns(selectedCols)
	assert.NoError(err)

	for i, c := range selectedCols {
		for row := 0; row < nRows; row++ {
			assert.Equal(allRows[row][c], cols[i][row],
				"col extraction mismatch: col=%d row=%d", c, row)
		}
	}
}

// ─── Benchmarks (drop-in CommitMerkleWithSIS) ───────────────────────────────

// BenchmarkCommitMerkleWithSIS_Small: 4096 × 256 ≈ 1M cells, rate=2.
func BenchmarkCommitMerkleWithSIS_Small(b *testing.B) {
	benchCommitMerkleWithSIS(b, 4096, 256, 2)
}

// BenchmarkCommitMerkleWithSIS_Typical: 16384 × 256, rate=2.
func BenchmarkCommitMerkleWithSIS_Typical(b *testing.B) {
	benchCommitMerkleWithSIS(b, 16384, 256, 2)
}

// BenchmarkCommitMerkleWithSIS_MedLarge: 1<<16 × 1<<10, rate=2.
func BenchmarkCommitMerkleWithSIS_MedLarge(b *testing.B) {
	benchCommitMerkleWithSIS(b, 1<<16, 1<<10, 2)
}

// BenchmarkCommitMerkleWithSIS_Large: 1<<19 × 1<<11 ≈ 1B cells, rate=2.
func BenchmarkCommitMerkleWithSIS_Large(b *testing.B) {
	benchCommitMerkleWithSIS(b, 1<<19, 1<<11, 2)
}

func benchCommitMerkleWithSIS(b *testing.B, nCols, nRows, rate int) {
	rng := rand.New(rand.NewChaCha8([32]byte{}))
	inputBytes := int64(nCols * nRows * 4)

	params := vortex_koalabear.NewParams(rate, nCols, nRows, 9, 16)
	m := randSmartVectorMatrix(rng, nRows, nCols)

	b.Logf("matrix: %dx%d (%s cells, %s encoded, %s input, %.0f%% const rows)",
		nCols, nRows,
		fmtCount(int64(nCols)*int64(nRows)),
		fmtCount(int64(nCols)*int64(nRows)*int64(rate)),
		fmtBytes(inputBytes),
		constFraction*100)

	// Warmup GPU
	CommitMerkleWithSIS(&params, m)

	b.Run("GPU_CommitMerkleWithSIS", func(b *testing.B) {
		b.SetBytes(inputBytes)
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			CommitMerkleWithSIS(&params, m)
		}
	})

	b.Run("CPU_CommitMerkleWithSIS", func(b *testing.B) {
		b.SetBytes(inputBytes)
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			params.CommitMerkleWithSIS(m)
		}
	})

	b.Run("Speedup", func(b *testing.B) {
		var gpuTotal, cpuTotal time.Duration
		for i := 0; i < b.N; i++ {
			start := time.Now()
			CommitMerkleWithSIS(&params, m)
			gpuTotal += time.Since(start)

			start = time.Now()
			params.CommitMerkleWithSIS(m)
			cpuTotal += time.Since(start)
		}
		n := time.Duration(b.N)
		gpuAvg := gpuTotal / n
		cpuAvg := cpuTotal / n
		speedup := float64(cpuAvg) / float64(gpuAvg)
		b.ReportMetric(float64(gpuAvg.Milliseconds()), "gpu_ms")
		b.ReportMetric(float64(cpuAvg.Milliseconds()), "cpu_ms")
		b.ReportMetric(speedup, "speedup_x")
	})
}

// ─── Low-level benchmarks (GPU commit only, no D2H) ────────────────────────

func BenchmarkCommitMerkle_Large(b *testing.B) {
	benchCommitMerkle(b, 1<<19, 1<<11, 2)
}

func BenchmarkCommitMerkle_Medium(b *testing.B) {
	benchCommitMerkle(b, 1<<18, 1<<12, 2)
}

func BenchmarkCommitMerkle_Rate8(b *testing.B) {
	benchCommitMerkle(b, 1<<16, 1<<11, 8)
}

func benchCommitMerkle(b *testing.B, nCols, nRows, rate int) {
	rng := rand.New(rand.NewChaCha8([32]byte{}))
	nSelected := min(256, nCols*rate/4)

	sisParams, _ := sis.NewRSis(0, 9, 16, nRows)
	params, _ := NewParams(nCols, nRows, sisParams, rate, nSelected)

	m := randMatrixMixed(rng, nRows, nCols)
	inputBytes := int64(nCols * nRows * 4)

	b.Logf("matrix: %dx%d (%s cells, %s encoded, %s input, %.0f%% const rows)",
		nCols, nRows,
		fmtCount(int64(nCols)*int64(nRows)),
		fmtCount(int64(nCols)*int64(nRows)*int64(rate)),
		fmtBytes(inputBytes),
		constFraction*100)

	dev, err := gpu.New()
	if err != nil {
		b.Fatal(err)
	}
	defer dev.Close()

	gv, err := NewGPUVortex(dev, params, nRows)
	if err != nil {
		b.Fatal(err)
	}
	defer gv.Free()

	if _, _, err := gv.Commit(m); err != nil {
		b.Fatal(err)
	}

	b.Run("GPU", func(b *testing.B) {
		b.SetBytes(inputBytes)
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			if _, _, err := gv.Commit(m); err != nil {
				b.Fatal(err)
			}
		}
	})

	b.Run("CPU", func(b *testing.B) {
		b.SetBytes(inputBytes)
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			if _, _, err := params.Commit(m); err != nil {
				b.Fatal(err)
			}
		}
	})

	b.Run("Speedup", func(b *testing.B) {
		var gpuTotal, cpuTotal time.Duration
		for i := 0; i < b.N; i++ {
			start := time.Now()
			_, _, _ = gv.Commit(m)
			gpuTotal += time.Since(start)

			start = time.Now()
			_, _, _ = params.Commit(m)
			cpuTotal += time.Since(start)
		}
		n := time.Duration(b.N)
		gpuAvg := gpuTotal / n
		cpuAvg := cpuTotal / n
		speedup := float64(cpuAvg) / float64(gpuAvg)
		b.ReportMetric(float64(gpuAvg.Milliseconds()), "gpu_ms")
		b.ReportMetric(float64(cpuAvg.Milliseconds()), "cpu_ms")
		b.ReportMetric(speedup, "speedup_x")
	})
}

// ─── Helpers ─────────────────────────────────────────────────────────────────

func fmtCount(n int64) string {
	switch {
	case n >= 1<<30:
		return fmt.Sprintf("%.1fG", float64(n)/float64(1<<30))
	case n >= 1<<20:
		return fmt.Sprintf("%.1fM", float64(n)/float64(1<<20))
	default:
		return fmt.Sprintf("%d", n)
	}
}

func fmtBytes(n int64) string {
	switch {
	case n >= 1<<30:
		return fmt.Sprintf("%.1f GiB", float64(n)/float64(1<<30))
	case n >= 1<<20:
		return fmt.Sprintf("%.1f MiB", float64(n)/float64(1<<20))
	default:
		return fmt.Sprintf("%d B", n)
	}
}
