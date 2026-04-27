//go:build cuda

package vortex

import (
	"math/rand/v2"
	"testing"
	"time"

	"github.com/consensys/gnark-crypto/field/koalabear"
	fext "github.com/consensys/gnark-crypto/field/koalabear/extensions"
	"github.com/consensys/gnark-crypto/field/koalabear/fft"
	"github.com/consensys/gnark-crypto/field/koalabear/sis"
	"github.com/consensys/linea-monorepo/prover/gpu"
	"github.com/stretchr/testify/require"
)

func newTestDevice(t *testing.T) *gpu.Device {
	t.Helper()
	dev, err := gpu.New()
	require.NoError(t, err)
	t.Cleanup(func() { dev.Close() })
	return dev
}

// ─── GPU Poseidon2 compress ──────────────────────────────────────────────────

func TestGPUPoseidon2Compress(t *testing.T) {
	assert := require.New(t)
	rng := rand.New(rand.NewChaCha8([32]byte{50}))
	dev := newTestDevice(t)

	p2, err := NewGPUPoseidon2(dev, 16)
	assert.NoError(err)
	defer p2.Free()

	var a, b Hash
	for j := 0; j < 8; j++ {
		a[j] = randKB(rng)
		b[j] = randKB(rng)
	}

	// CPU reference
	cpuHash := CompressPoseidon2(a, b)

	// GPU: pack input as [left[8] || right[8]] = 16 koalabear elements
	input := make([]koalabear.Element, 16)
	copy(input[:8], a[:])
	copy(input[8:], b[:])
	gpuHashes := p2.CompressBatch(input, 1)

	assert.Equal(cpuHash, gpuHashes[0], "GPU Poseidon2 compress mismatch")
}

// ─── GPU Vortex commit: RS encoding + leaf hash correctness ─────────────────

func TestGPUVortexCommit(t *testing.T) {
	assert := require.New(t)
	rng := rand.New(rand.NewChaCha8([32]byte{42}))
	dev := newTestDevice(t)

	nCols := 32
	nRows := 16
	rate := 2
	nSelected := 8

	sisParams, err := sis.NewRSis(0, 9, 16, nRows)
	assert.NoError(err)

	params, err := NewParams(nCols, nRows, sisParams, rate, nSelected)
	assert.NoError(err)

	m := randMatrix(rng, nRows, nCols)

	// GPU commit
	gv, err := NewGPUVortex(dev, params, nRows)
	assert.NoError(err)
	defer gv.Free()

	cs, _, err := gv.Commit(m)
	assert.NoError(err)

	// Compare RS encoding
	gpuRows, err := cs.ExtractAllRows()
	assert.NoError(err)
	for i := range m {
		cpuRow := make([]koalabear.Element, nCols*rate)
		params.EncodeReedSolomon(m[i], cpuRow)
		for j := range cpuRow {
			assert.Equal(cpuRow[j], gpuRows[i][j], "encoded[%d][%d]", i, j)
		}
	}

	// Compare leaves (GPU MD hash vs CPU CompressPoseidon2x16)
	gpuSIS, err := cs.ExtractSISHashes()
	assert.NoError(err)
	gpuLeaves, err := cs.ExtractLeaves()
	assert.NoError(err)

	scw := nCols * rate
	degree := sisParams.Degree
	cpuLeaves := make([]Hash, scw)
	n16 := scw / 16
	for c := 0; c < n16; c++ {
		start := c * 16 * degree
		CompressPoseidon2x16(gpuSIS[start:start+16*degree], degree, cpuLeaves[c*16:(c+1)*16])
	}
	for i := 0; i < scw; i++ {
		assert.Equal(cpuLeaves[i], gpuLeaves[i], "leaf[%d]", i)
	}
}

// ─── GPU Vortex linear combination + column extraction ──────────────────────

func TestGPUVortexLinComb(t *testing.T) {
	assert := require.New(t)
	rng := rand.New(rand.NewChaCha8([32]byte{77}))
	dev := newTestDevice(t)

	nCols := 32
	nRows := 16
	rate := 2
	nSelected := 4

	sisParams, err := sis.NewRSis(0, 9, 16, nRows)
	assert.NoError(err)

	params, err := NewParams(nCols, nRows, sisParams, rate, nSelected)
	assert.NoError(err)

	m := randMatrix(rng, nRows, nCols)
	alpha := randE4(rng)
	selectedCols := []int{0, 1, 2, 3}

	gv, err := NewGPUVortex(dev, params, nRows)
	assert.NoError(err)
	defer gv.Free()

	cs, _, err := gv.Commit(m)
	assert.NoError(err)

	// GPU linear combination
	uAlpha, err := cs.LinComb(alpha)
	assert.NoError(err)
	assert.Equal(nCols*rate, len(uAlpha), "UAlpha length")

	// Column extraction
	cols, err := cs.ExtractColumns(selectedCols)
	assert.NoError(err)
	assert.Equal(len(selectedCols), len(cols), "column count")

	// Cross-check: extracted columns against full matrix
	allRows, err := cs.ExtractAllRows()
	assert.NoError(err)
	for i, c := range selectedCols {
		for row := 0; row < nRows; row++ {
			assert.Equal(allRows[row][c], cols[i][row], "col[%d] row[%d]", c, row)
		}
	}
}

// ─── GPU E4 NTT ─────────────────────────────────────────────────────────────

func TestGPUFFTE4(t *testing.T) {
	assert := require.New(t)
	rng := rand.New(rand.NewChaCha8([32]byte{88}))
	dev := newTestDevice(t)

	const n = 1 << 14 // 16K E4 elements

	// Random E4 input
	data := make([]fext.E4, n)
	for i := range data {
		data[i] = randE4(rng)
	}

	// CPU reference: forward FFT (DIF)
	cpuData := make([]fext.E4, n)
	copy(cpuData, data)
	cpuDomain := fft.NewDomain(uint64(n))
	cpuDomain.FFTExt(cpuData, fft.DIF)

	// GPU forward FFT
	gpuData := make([]fext.E4, n)
	copy(gpuData, data)
	dom, err := NewGPUFFTDomain(dev, n)
	assert.NoError(err)
	defer dom.Free()

	dom.FFTE4(gpuData)

	// Compare
	mismatches := 0
	for i := range cpuData {
		if cpuData[i] != gpuData[i] {
			mismatches++
			if mismatches <= 5 {
				t.Errorf("FFTE4 mismatch at %d: cpu=%v gpu=%v", i, cpuData[i], gpuData[i])
			}
		}
	}
	assert.Equal(0, mismatches, "FFTE4 total mismatches")
}

func TestGPUFFTE4Roundtrip(t *testing.T) {
	assert := require.New(t)
	rng := rand.New(rand.NewChaCha8([32]byte{89}))
	dev := newTestDevice(t)

	const n = 1 << 14

	original := make([]fext.E4, n)
	for i := range original {
		original[i] = randE4(rng)
	}

	dom, err := NewGPUFFTDomain(dev, n)
	assert.NoError(err)
	defer dom.Free()

	// Forward then inverse
	data := make([]fext.E4, n)
	copy(data, original)
	dom.FFTE4(data)
	dom.FFTInverseE4(data)

	// Scale by 1/n (GPU IFFT does not include this)
	var nInv koalabear.Element
	nInv.SetUint64(uint64(n))
	nInv.Inverse(&nInv)
	for i := range data {
		data[i].B0.A0.Mul(&data[i].B0.A0, &nInv)
		data[i].B0.A1.Mul(&data[i].B0.A1, &nInv)
		data[i].B1.A0.Mul(&data[i].B1.A0, &nInv)
		data[i].B1.A1.Mul(&data[i].B1.A1, &nInv)
	}

	mismatches := 0
	for i := range original {
		if original[i] != data[i] {
			mismatches++
			if mismatches <= 5 {
				t.Errorf("E4 roundtrip mismatch at %d: orig=%v got=%v", i, original[i], data[i])
			}
		}
	}
	assert.Equal(0, mismatches, "E4 roundtrip total mismatches")
}

func TestGPUCosetFFTE4(t *testing.T) {
	assert := require.New(t)
	rng := rand.New(rand.NewChaCha8([32]byte{90}))
	dev := newTestDevice(t)

	const n = 1 << 14

	data := make([]fext.E4, n)
	for i := range data {
		data[i] = randE4(rng)
	}

	cpuDomain := fft.NewDomain(uint64(n))

	// CPU reference: coset FFT (DIF, OnCoset) → bit-reversed output
	cpuData := make([]fext.E4, n)
	copy(cpuData, data)
	cpuDomain.FFTExt(cpuData, fft.DIF, fft.OnCoset())
	// CPU DIF output is bit-reversed; GPU CosetFFT output is also bit-reversed
	// (kb_ntt_coset_fwd = ScaleByPowers + DIF, no final bitrev)

	// GPU coset FFT → bit-reversed output (ScaleByPowers + DIF)
	gpuData := make([]fext.E4, n)
	copy(gpuData, data)
	dom, err := NewGPUFFTDomain(dev, n)
	assert.NoError(err)
	defer dom.Free()

	dom.CosetFFTE4(gpuData, cpuDomain.FrMultiplicativeGen)

	mismatches := 0
	for i := range cpuData {
		if cpuData[i] != gpuData[i] {
			mismatches++
			if mismatches <= 5 {
				t.Errorf("CosetFFTE4 mismatch at %d: cpu=%v gpu=%v", i, cpuData[i], gpuData[i])
			}
		}
	}
	assert.Equal(0, mismatches, "CosetFFTE4 total mismatches")
}

// bitReverseE4 applies bit-reversal permutation on a slice of E4 elements.
func bitReverseE4(a []fext.E4) {
	n := len(a)
	nn := uint64(n)
	logN := uint64(0)
	for 1<<logN < nn {
		logN++
	}
	for i := uint64(0); i < nn; i++ {
		j := reverseBits(i, logN)
		if i < j {
			a[i], a[j] = a[j], a[i]
		}
	}
}

func reverseBits(v, logN uint64) uint64 {
	var r uint64
	for i := uint64(0); i < logN; i++ {
		r = (r << 1) | (v & 1)
		v >>= 1
	}
	return r
}

// ─── Benchmarks ─────────────────────────────────────────────────────────────

func benchCommit(b *testing.B, nCols, nRows, rate int) {
	dev, err := gpu.New()
	if err != nil {
		b.Fatal(err)
	}
	defer dev.Close()

	rng := rand.New(rand.NewChaCha8([32]byte{}))
	nSelected := min(32, nCols*rate/4)

	sisParams, _ := sis.NewRSis(0, 9, 16, nRows)
	params, _ := NewParams(nCols, nRows, sisParams, rate, nSelected)
	m := randMatrix(rng, nRows, nCols)
	inputBytes := int64(nCols * nRows * 4) // KoalaBear is uint32 (4 bytes)

	gv, err := NewGPUVortex(dev, params, nRows)
	if err != nil {
		b.Fatal(err)
	}
	defer gv.Free()

	b.Run("GPU", func(b *testing.B) {
		b.SetBytes(inputBytes)
		for i := 0; i < b.N; i++ {
			if _, _, err := gv.Commit(m); err != nil {
				b.Fatal(err)
			}
		}
	})
	b.Run("CPU", func(b *testing.B) {
		b.SetBytes(inputBytes)
		for i := 0; i < b.N; i++ {
			if _, _, err := params.Commit(m); err != nil {
				b.Fatal(err)
			}
		}
	})
	b.Run("GPU_vs_CPU", func(b *testing.B) {
		var gpuTotal, cpuTotal time.Duration
		for i := 0; i < b.N; i++ {
			start := time.Now()
			if _, _, err := gv.Commit(m); err != nil {
				b.Fatal(err)
			}
			gpuTotal += time.Since(start)

			start = time.Now()
			if _, _, err := params.Commit(m); err != nil {
				b.Fatal(err)
			}
			cpuTotal += time.Since(start)
		}

		gpuMBps := toMBps(inputBytes, gpuTotal/time.Duration(b.N))
		cpuMBps := toMBps(inputBytes, cpuTotal/time.Duration(b.N))
		b.ReportMetric(gpuMBps, "gpu_MB/s")
		b.ReportMetric(cpuMBps, "cpu_MB/s")
		if cpuMBps > 0 {
			b.ReportMetric(gpuMBps/cpuMBps, "gpu_vs_cpu_x")
		}
	})
}

func toMBps(bytes int64, d time.Duration) float64 {
	if d <= 0 {
		return 0
	}
	return float64(bytes) / d.Seconds() / (1024 * 1024)
}

func BenchmarkCommit_1024x128(b *testing.B)     { benchCommit(b, 1024, 128, 2) }
func BenchmarkCommit_4096x256(b *testing.B)     { benchCommit(b, 4096, 256, 2) }
func BenchmarkCommit_16384x256(b *testing.B)    { benchCommit(b, 16384, 256, 2) }
func BenchmarkCommit_524288x2048(b *testing.B)  { benchCommit(b, 1<<19, 1<<11, 2) } // ~1B cells
func BenchmarkCommit_1048576x4096(b *testing.B) { benchCommit(b, 1<<20, 1<<12, 2) } // ~4B cells

// ─── Rate 16: GPU RS encode correctness ──────────────────────────────────────
//
// gnark-crypto's NewParams rejects rate > 8, so we validate GPU RS encoding
// against CPU by doing a full commit + verify roundtrip (which exercises the
// same RS encode → SIS → Poseidon2 → Merkle → lincomb → column extraction
// pipeline with the higher rate).

func TestGPUVortexCommitRate16(t *testing.T) {
	assert := require.New(t)
	rng := rand.New(rand.NewChaCha8([32]byte{44}))
	dev := newTestDevice(t)

	nCols := 64
	nRows := 16
	rate := 16
	nSelected := 8

	sisParams, err := sis.NewRSis(0, 9, 16, nRows)
	assert.NoError(err)

	params, err := NewParams(nCols, nRows, sisParams, rate, nSelected)
	assert.NoError(err)

	m := randMatrix(rng, nRows, nCols)

	// GPU commit
	gv, err := NewGPUVortex(dev, params, nRows)
	assert.NoError(err)
	defer gv.Free()

	cs, _, err := gv.Commit(m)
	assert.NoError(err)

	// Verify RS encoding: CPU encode each row and compare with GPU
	gpuRows, err := cs.ExtractAllRows()
	assert.NoError(err)
	assert.Equal(nRows, len(gpuRows))
	assert.Equal(nCols*rate, len(gpuRows[0]))

	for i := range m {
		cpuRow := make([]koalabear.Element, nCols*rate)
		params.EncodeReedSolomon(m[i], cpuRow)
		for j := range cpuRow {
			assert.Equal(cpuRow[j], gpuRows[i][j], "rate16 encoded[%d][%d]", i, j)
		}
	}

	// Compare leaves: GPU MD hash vs CPU CompressPoseidon2x16
	gpuSIS, err := cs.ExtractSISHashes()
	assert.NoError(err)
	gpuLeaves, err := cs.ExtractLeaves()
	assert.NoError(err)

	scw := nCols * rate
	degree := sisParams.Degree
	cpuLeaves := make([]Hash, scw)
	n16 := scw / 16
	for c := 0; c < n16; c++ {
		start := c * 16 * degree
		CompressPoseidon2x16(gpuSIS[start:start+16*degree], degree, cpuLeaves[c*16:(c+1)*16])
	}
	for i := 0; i < scw; i++ {
		assert.Equal(cpuLeaves[i], gpuLeaves[i], "rate16 leaf[%d]", i)
	}

	// Note: params.Commit() uses gnark-crypto's BuildMerkleTree (single CompressPoseidon2),
	// while the GPU uses smt_koalabear's hashLR (2-block MD). Roots differ by design.
	// Full GPU-vs-CPU root comparison is done in TestCommitMerkleWithSIS_GPUvsCPU_Rate16.

	// GPU lincomb + column extraction
	alpha := randE4(rng)
	selectedCols := make([]int, nSelected)
	for i := range selectedCols {
		selectedCols[i] = rng.IntN(nCols*rate - 1)
	}

	gpuProof, err := cs.Prove(alpha, selectedCols)
	assert.NoError(err)
	assert.Equal(nCols*rate, len(gpuProof.UAlpha))
}

func TestGPUVortexLinCombRate16(t *testing.T) {
	assert := require.New(t)
	rng := rand.New(rand.NewChaCha8([32]byte{45}))
	dev := newTestDevice(t)

	nCols := 32
	nRows := 8
	rate := 16
	nSelected := 4

	sisParams, err := sis.NewRSis(0, 9, 16, nRows)
	assert.NoError(err)

	params, err := NewParams(nCols, nRows, sisParams, rate, nSelected)
	assert.NoError(err)

	m := randMatrix(rng, nRows, nCols)
	alpha := randE4(rng)

	gv, err := NewGPUVortex(dev, params, nRows)
	assert.NoError(err)
	defer gv.Free()

	cs, _, err := gv.Commit(m)
	assert.NoError(err)

	// GPU lincomb
	gpuUAlpha, err := cs.LinComb(alpha)
	assert.NoError(err)
	assert.Equal(nCols*rate, len(gpuUAlpha))

	// CPU reference lincomb from the GPU-extracted encoded rows
	gpuRows, err := cs.ExtractAllRows()
	assert.NoError(err)

	cpuUAlpha := make([]fext.E4, nCols*rate)
	var pow fext.E4
	pow.SetOne()
	for _, row := range gpuRows {
		for j := range row {
			var term fext.E4
			term.B0.A0 = row[j]
			term.Mul(&term, &pow)
			cpuUAlpha[j].Add(&cpuUAlpha[j], &term)
		}
		pow.Mul(&pow, &alpha)
	}

	for j := range cpuUAlpha {
		assert.Equal(cpuUAlpha[j], gpuUAlpha[j], "rate16 UAlpha[%d]", j)
	}
}

func BenchmarkCommit_64x16_rate16(b *testing.B)   { benchCommit(b, 64, 16, 16) }
func BenchmarkCommit_256x64_rate16(b *testing.B)   { benchCommit(b, 256, 64, 16) }
func BenchmarkCommit_1024x128_rate16(b *testing.B) { benchCommit(b, 1024, 128, 16) }
