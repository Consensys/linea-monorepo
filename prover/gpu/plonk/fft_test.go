//go:build cuda

package plonk_test

import (
	"fmt"
	"math/big"
	"testing"

	"github.com/consensys/gnark-crypto/ecc/bls12-377/fr"
	"github.com/consensys/gnark-crypto/ecc/bls12-377/fr/fft"
	gnarkutils "github.com/consensys/gnark-crypto/utils"

	"github.com/consensys/linea-monorepo/prover/gpu"
	"github.com/consensys/linea-monorepo/prover/gpu/plonk"
	"github.com/stretchr/testify/require"
)

// testFFTForward validates GPU forward NTT against gnark-crypto's CPU FFT.
// Both use DIF: natural-order input → bit-reversed output.
func testFFTForward(t *testing.T, n int) {
	assert := require.New(t)

	t.Helper()

	dev, err := gpu.New()
	assert.NoError(err, "New failed")
	defer dev.Close()

	domain := fft.NewDomain(uint64(n))

	// Random input data
	data := loadTestScalars(n)

	// CPU reference: DIF forward FFT
	cpuData := make(fr.Vector, n)
	copy(cpuData, data)
	domain.FFT(cpuData, fft.DIF)

	// GPU forward FFT
	gpuDomain, err := plonk.NewFFTDomain(dev, n)
	assert.NoError(err, "NewFFTDomain failed")
	defer gpuDomain.Close()

	gpuVec, err := plonk.NewFrVector(dev, n)
	assert.NoError(err, "NewFrVector failed")
	defer gpuVec.Free()

	gpuVec.CopyFromHost(data)
	gpuDomain.FFT(gpuVec)
	dev.Sync()

	gpuResult := make(fr.Vector, n)
	gpuVec.CopyToHost(gpuResult)

	for i := 0; i < n; i++ {
		if !gpuResult[i].Equal(&cpuData[i]) {
			t.Errorf("FFT mismatch at %d:\n  gpu:  %v\n  cpu:  %v", i, gpuResult[i], cpuData[i])
			if i > 10 {
				t.Fatalf("too many mismatches, stopping")
			}
		}
	}
}

func TestFFTSmall(t *testing.T) {
	testFFTForward(t, 1024)
}

func TestFFTMedium(t *testing.T) {
	testFFTForward(t, 1<<15)
}

func TestFFTLarge(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping large FFT test in short mode")
	}
	testFFTForward(t, 1<<20)
}

func TestFFTRoundtrip(t *testing.T) {
	assert := require.New(t)

	const n = 1 << 15

	dev, err := gpu.New()
	assert.NoError(err, "New failed")
	defer dev.Close()

	gpuDomain, err := plonk.NewFFTDomain(dev, n)
	assert.NoError(err, "NewFFTDomain failed")
	defer gpuDomain.Close()

	gpuVec, err := plonk.NewFrVector(dev, n)
	assert.NoError(err, "NewFrVector failed")
	defer gpuVec.Free()

	// Random input
	original := loadTestScalars(n)
	gpuVec.CopyFromHost(original)

	// Forward FFT (DIF): natural input → bit-reversed output
	gpuDomain.FFT(gpuVec)
	// Inverse FFT (DIT): bit-reversed input → natural output, scaled by 1/n
	gpuDomain.FFTInverse(gpuVec)
	dev.Sync()

	result := make(fr.Vector, n)
	gpuVec.CopyToHost(result)

	for i := 0; i < n; i++ {
		if !result[i].Equal(&original[i]) {
			t.Errorf("roundtrip mismatch at %d:\n  got:  %v\n  want: %v", i, result[i], original[i])
			if i > 10 {
				t.Fatalf("too many mismatches, stopping")
			}
		}
	}
}

func TestFFTBitReverse(t *testing.T) {
	assert := require.New(t)

	const n = 1024

	dev, err := gpu.New()
	assert.NoError(err, "New failed")
	defer dev.Close()

	gpuDomain, err := plonk.NewFFTDomain(dev, n)
	assert.NoError(err, "NewFFTDomain failed")
	defer gpuDomain.Close()

	gpuVec, err := plonk.NewFrVector(dev, n)
	assert.NoError(err, "NewFrVector failed")
	defer gpuVec.Free()

	// Input: [0, 1, 2, ..., n-1] as Fr elements
	input := make(fr.Vector, n)
	for i := range input {
		input[i].SetUint64(uint64(i))
	}
	gpuVec.CopyFromHost(input)

	gpuDomain.BitReverse(gpuVec)
	dev.Sync()

	result := make(fr.Vector, n)
	gpuVec.CopyToHost(result)

	// Verify bit-reversal: result[bitrev(i)] == input[i]
	logN := 0
	for tmp := n; tmp > 1; tmp >>= 1 {
		logN++
	}
	for i := 0; i < n; i++ {
		j := bitRev(i, logN)
		if !result[j].Equal(&input[i]) {
			t.Errorf("bit-reverse mismatch: result[%d] != input[%d]", j, i)
			if i > 10 {
				t.Fatalf("too many mismatches, stopping")
			}
		}
	}
}

// bitRev returns the bit-reversal of x within logN bits.
func bitRev(x, logN int) int {
	r := 0
	for i := 0; i < logN; i++ {
		r = (r << 1) | (x & 1)
		x >>= 1
	}
	return r
}

func TestFFTEdgeCases(t *testing.T) {
	assert := require.New(t)

	dev, err := gpu.New()
	assert.NoError(err, "New failed")
	defer dev.Close()

	t.Run("ZeroInput", func(t *testing.T) {
		assert := require.New(t)

		const n = 1024
		domain := fft.NewDomain(uint64(n))

		data := make(fr.Vector, n) // all zero

		cpuData := make(fr.Vector, n)
		domain.FFT(cpuData, fft.DIF)

		gpuDomain, err := plonk.NewFFTDomain(dev, n)
		assert.NoError(err, "NewFFTDomain failed")
		defer gpuDomain.Close()

		gpuVec, err := plonk.NewFrVector(dev, n)
		assert.NoError(err, "NewFrVector failed")
		defer gpuVec.Free()

		gpuVec.CopyFromHost(data)
		gpuDomain.FFT(gpuVec)
		dev.Sync()

		result := make(fr.Vector, n)
		gpuVec.CopyToHost(result)

		for i := 0; i < n; i++ {
			if !result[i].Equal(&cpuData[i]) {
				t.Errorf("zero input: mismatch at %d", i)
			}
		}
	})

	t.Run("AllSame", func(t *testing.T) {
		assert := require.New(t)

		const n = 1024
		domain := fft.NewDomain(uint64(n))

		data := make(fr.Vector, n)
		var val fr.Element
		val.SetUint64(42)
		for i := range data {
			data[i] = val
		}

		cpuData := make(fr.Vector, n)
		copy(cpuData, data)
		domain.FFT(cpuData, fft.DIF)

		gpuDomain, err := plonk.NewFFTDomain(dev, n)
		assert.NoError(err, "NewFFTDomain failed")
		defer gpuDomain.Close()

		gpuVec, err := plonk.NewFrVector(dev, n)
		assert.NoError(err, "NewFrVector failed")
		defer gpuVec.Free()

		gpuVec.CopyFromHost(data)
		gpuDomain.FFT(gpuVec)
		dev.Sync()

		result := make(fr.Vector, n)
		gpuVec.CopyToHost(result)

		for i := 0; i < n; i++ {
			if !result[i].Equal(&cpuData[i]) {
				t.Errorf("all-same input: mismatch at %d", i)
			}
		}
	})
}

func BenchmarkFFT(b *testing.B) {
	assert := require.New(b)

	dev, err := gpu.New()
	assert.NoError(err, "New failed")
	defer dev.Close()

	// Large-size focus includes 1<<27 (4G elements in Fr AoS on host, 8G on device vector).
	// For very large sizes, benchmark kernel iterations directly to avoid host reset dominating.
	sizes := []int{1 << 14, 1 << 18, 1 << 22} //, 1 << 20, 1 << 22, 1 << 24, 1 << 25, 1 << 26, 1 << 27}

	for _, n := range sizes {
		if testing.Short() && n > 1<<20 {
			continue
		}
		b.Run(formatSize(n), func(b *testing.B) {
			assert := require.New(b)

			gpuDomain, err := plonk.NewFFTDomain(dev, n)
			assert.NoError(err, "NewFFTDomain failed")
			defer gpuDomain.Close()

			gpuVec, err := plonk.NewFrVector(dev, n)
			assert.NoError(err, "NewFrVector failed")
			defer gpuVec.Free()

			var data fr.Vector
			data = loadTestScalars(n)
			gpuVec.CopyFromHost(data)
			dev.Sync()

			// Warm-up
			gpuDomain.FFT(gpuVec)
			dev.Sync()

			b.ResetTimer()
			if n >= 1<<25 {
				for i := 0; i < b.N; i++ {
					gpuDomain.FFT(gpuVec)
					dev.Sync()
				}
			} else {
				for i := 0; i < b.N; i++ {
					gpuVec.CopyFromHost(data) // reset data each iteration
					gpuDomain.FFT(gpuVec)
					dev.Sync()
				}
			}
		})
	}
}

// TestCosetFFT validates GPU CosetFFT against CPU reference.
func TestCosetFFT(t *testing.T) {
	assert := require.New(t)

	dev, err := gpu.New()
	assert.NoError(err, "New failed")
	defer dev.Close()

	sizes := []int{1024, 1 << 15}
	for _, n := range sizes {
		t.Run(fmt.Sprintf("n=%d", n), func(t *testing.T) {
			assert := require.New(t)

			domain := fft.NewDomain(uint64(n))
			data := loadTestScalars(n)

			// CPU reference: coset FFT produces natural-order evaluations on coset
			cpuData := make(fr.Vector, n)
			copy(cpuData, data)
			domain.FFT(cpuData, fft.DIF, fft.OnCoset())
			// DIF produces bit-reversed output; bit-reverse to get natural order
			gnarkutils.BitReverse(cpuData)

			// GPU CosetFFT (produces natural-order output)
			gpuDomain, err := plonk.NewFFTDomain(dev, n)
			assert.NoError(err, "NewFFTDomain failed")
			defer gpuDomain.Close()

			gpuVec, err := plonk.NewFrVector(dev, n)
			assert.NoError(err, "NewFrVector failed")
			defer gpuVec.Free()

			gpuVec.CopyFromHost(data)
			gpuDomain.CosetFFT(gpuVec, domain.FrMultiplicativeGen)
			dev.Sync()

			result := make(fr.Vector, n)
			gpuVec.CopyToHost(result)

			for i := 0; i < n; i++ {
				if !result[i].Equal(&cpuData[i]) {
					t.Errorf("CosetFFT mismatch at %d:\n  gpu: %v\n  cpu: %v", i, result[i], cpuData[i])
					if i > 5 {
						t.Fatalf("too many mismatches")
					}
				}
			}
		})
	}
}

// TestCosetFFTRoundtrip validates CosetFFT → CosetFFTInverse recovers original data.
func TestCosetFFTRoundtrip(t *testing.T) {
	assert := require.New(t)

	dev, err := gpu.New()
	assert.NoError(err, "New failed")
	defer dev.Close()

	const n = 1 << 15
	domain := fft.NewDomain(uint64(n))

	gpuDomain, err := plonk.NewFFTDomain(dev, n)
	assert.NoError(err, "NewFFTDomain failed")
	defer gpuDomain.Close()

	gpuVec, err := plonk.NewFrVector(dev, n)
	assert.NoError(err, "NewFrVector failed")
	defer gpuVec.Free()

	original := loadTestScalars(n)
	gpuVec.CopyFromHost(original)

	// Forward coset FFT, then inverse coset FFT — should recover original
	var gInv fr.Element
	gInv.Inverse(&domain.FrMultiplicativeGen)

	gpuDomain.CosetFFT(gpuVec, domain.FrMultiplicativeGen)
	gpuDomain.CosetFFTInverse(gpuVec, gInv)
	dev.Sync()

	result := make(fr.Vector, n)
	gpuVec.CopyToHost(result)

	for i := 0; i < n; i++ {
		if !result[i].Equal(&original[i]) {
			t.Errorf("roundtrip mismatch at %d:\n  got:  %v\n  want: %v", i, result[i], original[i])
			if i > 5 {
				t.Fatalf("too many mismatches")
			}
		}
	}
}

// TestCosetFFTInverse validates GPU CosetFFTInverse against CPU reference.
func TestCosetFFTInverse(t *testing.T) {
	assert := require.New(t)

	dev, err := gpu.New()
	assert.NoError(err, "New failed")
	defer dev.Close()

	const n = 1 << 15
	domain := fft.NewDomain(uint64(n))

	// Start with a polynomial, evaluate on coset (CPU), then inverse on GPU
	poly := loadTestScalars(n)

	// CPU: evaluate on coset
	cpuEvals := make(fr.Vector, n)
	copy(cpuEvals, poly)
	domain.FFT(cpuEvals, fft.DIF, fft.OnCoset())
	gnarkutils.BitReverse(cpuEvals)
	// cpuEvals is now natural-order evaluations on coset

	// GPU: inverse coset FFT should recover the polynomial
	gpuDomain, err := plonk.NewFFTDomain(dev, n)
	assert.NoError(err, "NewFFTDomain failed")
	defer gpuDomain.Close()

	gpuVec, _ := plonk.NewFrVector(dev, n)
	defer gpuVec.Free()

	gpuVec.CopyFromHost(cpuEvals)
	var gInv fr.Element
	gInv.Inverse(&domain.FrMultiplicativeGen)
	gpuDomain.CosetFFTInverse(gpuVec, gInv)
	dev.Sync()

	result := make(fr.Vector, n)
	gpuVec.CopyToHost(result)

	for i := 0; i < n; i++ {
		if !result[i].Equal(&poly[i]) {
			t.Errorf("CosetFFTInverse mismatch at %d:\n  got:  %v\n  want: %v", i, result[i], poly[i])
			if i > 5 {
				t.Fatalf("too many mismatches")
			}
		}
	}
}

// =============================================================================
// Butterfly4 / Decomposed iFFT(4n) Tests
// =============================================================================

// TestButterfly4DecomposedIFFT validates decomposed iFFT(4n) against CPU iFFT(4n).
// The decomposition: CosetFFTInverse on each of 4 blocks + butterfly4 + shard scaling.
func TestButterfly4DecomposedIFFT(t *testing.T) {
	assert := require.New(t)

	dev, err := gpu.New()
	assert.NoError(err, "New failed")
	defer dev.Close()

	sizes := []int{32, 256, 1024}
	for _, n := range sizes {
		t.Run(fmt.Sprintf("n=%d", n), func(t *testing.T) {
			assert := require.New(t)

			bigN := 4 * n

			// Create a random polynomial of degree < 4n
			poly := loadTestScalars(bigN)

			// === CPU reference: evaluate on coset domain1, then iFFT(4n) ===
			domain1 := fft.NewDomain(uint64(bigN))
			cpuEvals := make(fr.Vector, bigN)
			copy(cpuEvals, poly)
			// Forward coset FFT on domain1
			domain1.FFT(cpuEvals, fft.DIF, fft.OnCoset())
			gnarkutils.BitReverse(cpuEvals)
			// cpuEvals now has natural-order evaluations on coset domain1

			// Now inverse coset FFT to recover polynomial
			// DIF forward produces bit-reversed output; DIT inverse expects bit-reversed input.
			cpuRecovered := make(fr.Vector, bigN)
			copy(cpuRecovered, poly)
			domain1.FFT(cpuRecovered, fft.DIF, fft.OnCoset())        // fwd coset: natural → bit-rev
			domain1.FFTInverse(cpuRecovered, fft.DIT, fft.OnCoset()) // inv coset: bit-rev → natural
			// cpuRecovered should equal poly

			// Verify CPU roundtrip
			for i := 0; i < bigN; i++ {
				if !cpuRecovered[i].Equal(&poly[i]) {
					t.Fatalf("CPU roundtrip failed at %d", i)
				}
			}

			// === GPU decomposed approach ===
			// Split evaluations into 4 coset blocks:
			// Coset k has evaluations at u·g₁^k · ω₀^i for i=0..n-1
			// In natural order of domain1: val[4i+k] goes to coset k, index i
			domain0 := fft.NewDomain(uint64(n))
			g1 := domain1.Generator // primitive 4n-th root

			blocks := [4]fr.Vector{}
			for k := 0; k < 4; k++ {
				blocks[k] = make(fr.Vector, n)
				for i := 0; i < n; i++ {
					blocks[k][i] = cpuEvals[4*i+k]
				}
			}

			// Create GPU resources
			gpuDomain, err := plonk.NewFFTDomain(dev, n)
			assert.NoError(err, "NewFFTDomain failed")
			defer gpuDomain.Close()

			gpuBlocks := [4]*plonk.FrVector{}
			for k := 0; k < 4; k++ {
				gpuBlocks[k], err = plonk.NewFrVector(dev, n)
				assert.NoError(err, "NewFrVector failed")
				defer gpuBlocks[k].Free()
			}

			// Step 1: Upload and CosetFFTInverse each block
			u := domain0.FrMultiplicativeGen // coset shift (same for domain0 and domain1)
			for k := 0; k < 4; k++ {
				gpuBlocks[k].CopyFromHost(blocks[k])

				// Coset generator for sub-coset k: u * g₁^k
				var cosetGen, cosetGenInv fr.Element
				var gPow fr.Element
				gPow.Exp(g1, big.NewInt(int64(k)))
				cosetGen.Mul(&u, &gPow)
				cosetGenInv.Inverse(&cosetGen)

				gpuDomain.CosetFFTInverse(gpuBlocks[k], cosetGenInv)
			}

			// Step 2: Butterfly4
			// omega4 = g₁^n (primitive 4th root of unity)
			var omega4, omega4Inv, quarter fr.Element
			omega4.Exp(g1, big.NewInt(int64(n)))
			omega4Inv.Inverse(&omega4)
			quarter.SetUint64(4)
			quarter.Inverse(&quarter)

			plonk.Butterfly4Inverse(gpuBlocks[0], gpuBlocks[1], gpuBlocks[2], gpuBlocks[3],
				omega4Inv, quarter)

			// Step 3: Apply shard scaling: block_l *= u^{-ln}
			// l=0: no scaling needed
			// l=1: *= u^{-n}, l=2: *= u^{-2n}, l=3: *= u^{-3n}
			var uInvN fr.Element
			var uInv fr.Element
			uInv.Inverse(&u)
			uInvN.Exp(uInv, big.NewInt(int64(n)))
			for l := 1; l < 4; l++ {
				var scale fr.Element
				scale.Exp(uInvN, big.NewInt(int64(l)))
				gpuBlocks[l].ScalarMul(scale)
			}

			dev.Sync()

			// Download and compare
			gpuResult := make(fr.Vector, bigN)
			for l := 0; l < 4; l++ {
				tmp := make(fr.Vector, n)
				gpuBlocks[l].CopyToHost(tmp)
				for j := 0; j < n; j++ {
					gpuResult[l*n+j] = tmp[j]
				}
			}

			// Compare against original polynomial
			mismatches := 0
			for i := 0; i < bigN; i++ {
				if !gpuResult[i].Equal(&poly[i]) {
					if mismatches < 10 {
						t.Errorf("decomposed iFFT mismatch at %d:\n  gpu: %v\n  cpu: %v",
							i, gpuResult[i], poly[i])
					}
					mismatches++
				}
			}
			if mismatches > 0 {
				t.Fatalf("total mismatches: %d / %d", mismatches, bigN)
			}
		})
	}
}

// TestButterfly4Simple validates the butterfly4 kernel alone with a known-answer test.
func TestButterfly4Simple(t *testing.T) {
	assert := require.New(t)

	dev, err := gpu.New()
	assert.NoError(err, "New failed")
	defer dev.Close()

	// Use n=1 to test a single butterfly operation.
	// For n=1, each block has 1 element. The butterfly should compute:
	// b0 = (a0+a1+a2+a3)/4, b1 = ((a0-a2)+w*(a1-a3))/4, etc.
	const n = 1

	// Create 4 blocks with known values
	blocks := [4]fr.Vector{}
	for k := 0; k < 4; k++ {
		blocks[k] = make(fr.Vector, n)
		blocks[k][0].SetUint64(uint64(k + 1)) // a0=1, a1=2, a2=3, a3=4
	}

	gpuBlocks := [4]*plonk.FrVector{}
	for k := 0; k < 4; k++ {
		gpuBlocks[k], _ = plonk.NewFrVector(dev, n)
		defer gpuBlocks[k].Free()
		gpuBlocks[k].CopyFromHost(blocks[k])
	}

	// Use the BLS12-377 primitive 4th root of unity
	// For any field, omega4 satisfies omega4^2 = -1
	domain4 := fft.NewDomain(4)
	omega4 := domain4.Generator
	var omega4Inv, quarter fr.Element
	omega4Inv.Inverse(&omega4)
	quarter.SetUint64(4)
	quarter.Inverse(&quarter)

	plonk.Butterfly4Inverse(gpuBlocks[0], gpuBlocks[1], gpuBlocks[2], gpuBlocks[3],
		omega4Inv, quarter)
	dev.Sync()

	// Compute expected on CPU
	var a0, a1, a2, a3 fr.Element
	a0.SetUint64(1)
	a1.SetUint64(2)
	a2.SetUint64(3)
	a3.SetUint64(4)

	// b0 = (a0+a1+a2+a3)/4 = (1+2+3+4)/4 = 10/4
	var t0, t1, t2, diff13 fr.Element
	t0.Add(&a0, &a2)     // 1+3 = 4
	t1.Sub(&a0, &a2)     // 1-3 = -2
	t2.Add(&a1, &a3)     // 2+4 = 6
	diff13.Sub(&a1, &a3) // 2-4 = -2

	var wDiff fr.Element
	wDiff.Mul(&omega4Inv, &diff13) // omega4_inv * (-2)

	var expected [4]fr.Element
	var tmp fr.Element
	tmp.Add(&t0, &t2)
	expected[0].Mul(&tmp, &quarter) // (4+6)/4 = 10/4

	tmp.Add(&t1, &wDiff)
	expected[1].Mul(&tmp, &quarter) // (-2 + omega4_inv*(-2))/4

	tmp.Sub(&t0, &t2)
	expected[2].Mul(&tmp, &quarter) // (4-6)/4 = -2/4

	tmp.Sub(&t1, &wDiff)
	expected[3].Mul(&tmp, &quarter) // (-2 - omega4_inv*(-2))/4

	for k := 0; k < 4; k++ {
		result := make(fr.Vector, n)
		gpuBlocks[k].CopyToHost(result)
		if !result[0].Equal(&expected[k]) {
			t.Errorf("block %d: got %v, want %v", k, result[0], expected[k])
		}
	}
}

func BenchmarkFFTCPU(b *testing.B) {
	sizes := []int{1 << 16, 1 << 18, 1 << 20, 1 << 22, 1 << 24}

	for _, n := range sizes {
		if testing.Short() && n > 1<<20 {
			continue
		}
		b.Run(formatSize(n), func(b *testing.B) {
			domain := fft.NewDomain(uint64(n))
			data := loadTestScalars(n)
			buf := make(fr.Vector, n)

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				copy(buf, data)
				domain.FFT(buf, fft.DIF)
			}
		})
	}
}

// TestFFT22 validates GPU forward FFT at 2^22 size against CPU reference.
// (Extracted from fft22_test.go, converted to external test package.)
func TestFFT22(t *testing.T) {
	assert := require.New(t)

	n := 1 << 22

	dev, err := gpu.New()
	assert.NoError(err, "Failed")
	defer dev.Close()

	domain := fft.NewDomain(uint64(n))
	data := make(fr.Vector, n)
	for i := range data {
		data[i].SetRandom()
	}

	cpuData := make(fr.Vector, n)
	copy(cpuData, data)
	domain.FFT(cpuData, fft.DIF)

	gpuDomain, err := plonk.NewFFTDomain(dev, n)
	assert.NoError(err, "NewFFTDomain failed")
	defer gpuDomain.Close()

	gpuVec, err := plonk.NewFrVector(dev, n)
	assert.NoError(err, "NewFrVector failed")
	defer gpuVec.Free()

	gpuVec.CopyFromHost(data)
	gpuDomain.FFT(gpuVec)
	dev.Sync()

	gpuResult := make(fr.Vector, n)
	gpuVec.CopyToHost(gpuResult)

	mismatches := 0
	for i := 0; i < n; i++ {
		if !gpuResult[i].Equal(&cpuData[i]) {
			mismatches++
		}
	}
	if mismatches > 0 {
		t.Errorf("FAIL: %d/%d mismatches", mismatches, n)
	}
}
