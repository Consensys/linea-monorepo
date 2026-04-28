//go:build cuda && iciclebench

package plonk2

import (
	"fmt"
	"os"
	goRuntime "runtime"
	"sync"
	"testing"
	"unsafe"

	bls12377 "github.com/consensys/gnark-crypto/ecc/bls12-377"
	blsfr "github.com/consensys/gnark-crypto/ecc/bls12-377/fr"
	icicleCore "github.com/ingonyama-zk/icicle-gnark/v3/wrappers/golang/core"
	icicleBLS12377 "github.com/ingonyama-zk/icicle-gnark/v3/wrappers/golang/curves/bls12377"
	icicleBLS12377MSM "github.com/ingonyama-zk/icicle-gnark/v3/wrappers/golang/curves/bls12377/msm"
	icicleBLS12377NTT "github.com/ingonyama-zk/icicle-gnark/v3/wrappers/golang/curves/bls12377/ntt"
	icicleBN254 "github.com/ingonyama-zk/icicle-gnark/v3/wrappers/golang/curves/bn254"
	icicleBN254NTT "github.com/ingonyama-zk/icicle-gnark/v3/wrappers/golang/curves/bn254/ntt"
	icicleBW6761 "github.com/ingonyama-zk/icicle-gnark/v3/wrappers/golang/curves/bw6761"
	icicleBW6761NTT "github.com/ingonyama-zk/icicle-gnark/v3/wrappers/golang/curves/bw6761/ntt"
	icicleRuntime "github.com/ingonyama-zk/icicle-gnark/v3/wrappers/golang/runtime"
	icicleConfig "github.com/ingonyama-zk/icicle-gnark/v3/wrappers/golang/runtime/config_extension"
	"github.com/stretchr/testify/require"

	"github.com/consensys/linea-monorepo/prover/gpu"
	oldplonk "github.com/consensys/linea-monorepo/prover/gpu/plonk"
)

var icicleBenchOnce sync.Once

func BenchmarkCompareNTTPlonk2VsICICLE(b *testing.B) {
	const n = 1 << 20

	b.Run("bn254", func(b *testing.B) {
		dev, err := gpu.New()
		require.NoError(b, err, "creating gnark GPU device should succeed")
		defer dev.Close()

		input := deterministicBN254(n, 101)
		rawInput := cloneRaw(rawBN254(input))
		benchPlonk2FFT(b, dev, CurveBN254, fftSpecBN254(n), rawInput, n)
	})
	b.Run("bls12-377", func(b *testing.B) {
		dev, err := gpu.New()
		require.NoError(b, err, "creating gnark GPU device should succeed")
		defer dev.Close()

		input := deterministicBLS12377(n, 102)
		rawInput := cloneRaw(rawBLS12377(input))
		benchPlonk2FFT(b, dev, CurveBLS12377, fftSpecBLS12377(n), rawInput, n)
	})
	b.Run("bw6-761", func(b *testing.B) {
		dev, err := gpu.New()
		require.NoError(b, err, "creating gnark GPU device should succeed")
		defer dev.Close()

		input := deterministicBW6761(n, 103)
		rawInput := cloneRaw(rawBW6761(input))
		benchPlonk2FFT(b, dev, CurveBW6761, fftSpecBW6761(n), rawInput, n)
	})

	b.Run("icicle-bn254", func(b *testing.B) {
		goRuntime.LockOSThread()
		defer goRuntime.UnlockOSThread()
		initICICLEBench(b)
		input := icicleBN254.GenerateScalars(n)
		initICICLEBN254Domain(b, n)
		defer requireICICLESuccess(b, icicleBN254NTT.ReleaseDomain(), "release BN254 NTT domain")
		benchICICLENTT(b, input, icicleBN254NTT.GetDefaultNttConfig(), icicleBN254NTT.Ntt)
	})
	b.Run("icicle-bls12-377", func(b *testing.B) {
		goRuntime.LockOSThread()
		defer goRuntime.UnlockOSThread()
		initICICLEBench(b)
		input := icicleBLS12377.GenerateScalars(n)
		initICICLEBLS12377Domain(b, n)
		defer requireICICLESuccess(b, icicleBLS12377NTT.ReleaseDomain(), "release BLS12-377 NTT domain")
		benchICICLENTT(b, input, icicleBLS12377NTT.GetDefaultNttConfig(), icicleBLS12377NTT.Ntt)
	})
	b.Run("icicle-bw6-761", func(b *testing.B) {
		goRuntime.LockOSThread()
		defer goRuntime.UnlockOSThread()
		initICICLEBench(b)
		input := icicleBW6761.GenerateScalars(n)
		initICICLEBW6761Domain(b, n)
		defer requireICICLESuccess(b, icicleBW6761NTT.ReleaseDomain(), "release BW6-761 NTT domain")
		benchICICLENTT(b, input, icicleBW6761NTT.GetDefaultNttConfig(), icicleBW6761NTT.Ntt)
	})
}

func BenchmarkCompareMSMBLS12377SRS(b *testing.B) {
	dev, err := gpu.New()
	require.NoError(b, err, "creating gnark GPU device should succeed")
	defer dev.Close()

	store, err := oldplonk.NewSRSStore(benchSRSRootDir)
	require.NoError(b, err, "creating SRS store should succeed")

	for _, n := range []int{1 << 14, 1 << 16} {
		scalars := loadBenchBLS12377Scalars(b, n)

		b.Run(fmt.Sprintf("old-plonk/n=%s", benchFormatSize(n)), func(b *testing.B) {
			tePoints, err := store.LoadTEPoints(n, true)
			require.NoError(b, err, "loading TE SRS should succeed")
			pts, err := oldplonk.PinG1TEPoints(tePoints)
			require.NoError(b, err, "pinning TE SRS should succeed")
			msm, err := oldplonk.NewG1MSM(dev, pts)
			require.NoError(b, err, "creating old plonk MSM should succeed")
			defer msm.Close()
			require.NoError(b, msm.PinWorkBuffers(), "pinning old plonk MSM work buffers should succeed")
			require.NoError(b, dev.Sync(), "device sync before benchmark should succeed")

			b.SetBytes(int64(n) * int64(blsfr.Bytes))
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_ = msm.MultiExp(scalars)
			}
		})

		b.Run(fmt.Sprintf("plonk2/n=%s", benchFormatSize(n)), func(b *testing.B) {
			points, err := store.LoadPointsAffine(n, true)
			require.NoError(b, err, "loading affine SRS should succeed")
			msm, err := NewG1MSM(dev, CurveBLS12377, rawBLS12377G1Slice(points))
			require.NoError(b, err, "creating plonk2 MSM should succeed")
			defer msm.Close()
			require.NoError(b, msm.PinWorkBuffers(), "pinning plonk2 MSM work buffers should succeed")
			rawScalars := cloneRaw(rawBLS12377(scalars))
			require.NoError(b, dev.Sync(), "device sync before benchmark should succeed")

			b.SetBytes(int64(n) * int64(blsfr.Bytes))
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_, err := msm.CommitRaw(rawScalars)
				require.NoError(b, err, "plonk2 MSM should succeed")
			}
		})

		b.Run(fmt.Sprintf("icicle/n=%s", benchFormatSize(n)), func(b *testing.B) {
			goRuntime.LockOSThread()
			defer goRuntime.UnlockOSThread()
			initICICLEBench(b)
			points, err := store.LoadPointsAffine(n, true)
			require.NoError(b, err, "loading affine SRS should succeed")
			iciclePoints := icicleBLS12377AffineFromGnark(points)
			icicleScalars := icicleBLS12377ScalarsFromGnark(scalars)
			benchICICLEBLS12377MSM(b, icicleScalars, iciclePoints)
		})
	}
}

func benchPlonk2FFT(
	b *testing.B,
	dev *gpu.Device,
	curve Curve,
	spec FFTDomainSpec,
	rawInput []uint64,
	n int,
) {
	b.Helper()
	require.NoError(b, dev.Bind(), "binding gnark GPU device should succeed")
	domain, err := NewFFTDomain(dev, spec)
	require.NoError(b, err, "creating plonk2 FFT domain should succeed")
	defer domain.Free()
	vec, err := NewFrVector(dev, curve, n)
	require.NoError(b, err, "allocating plonk2 FFT vector should succeed")
	defer vec.Free()
	require.NoError(b, vec.CopyFromHostRaw(rawInput), "copying FFT input should succeed")
	require.NoError(b, dev.Sync(), "device sync before benchmark should succeed")

	b.SetBytes(int64(len(rawInput) * 8))
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		require.NoError(b, domain.FFT(vec), "plonk2 FFT should succeed")
		require.NoError(b, dev.Sync(), "plonk2 FFT sync should succeed")
	}
}

func benchICICLENTT[ScalarT any, ConfigT any](
	b *testing.B,
	input icicleCore.HostSlice[ScalarT],
	cfg icicleCore.NTTConfig[ConfigT],
	nttFn func(
		icicleCore.HostOrDeviceSlice,
		icicleCore.NTTDir,
		*icicleCore.NTTConfig[ConfigT],
		icicleCore.HostOrDeviceSlice,
	) icicleRuntime.EIcicleError,
) {
	b.Helper()
	var deviceInput icicleCore.DeviceSlice
	input.CopyToDevice(&deviceInput, true)
	defer func() {
		activateICICLECUDA(b)
		requireICICLESuccess(b, deviceInput.Free(), "freeing ICICLE NTT input")
	}()

	var deviceOutput icicleCore.DeviceSlice
	_, err := deviceOutput.Malloc(input.SizeOfElement(), input.Len())
	requireICICLESuccess(b, err, "allocating ICICLE NTT output")
	defer func() {
		activateICICLECUDA(b)
		requireICICLESuccess(b, deviceOutput.Free(), "freeing ICICLE NTT output")
	}()

	cfg.IsAsync = false
	cfg.StreamHandle = nil
	requireICICLESuccess(b, icicleRuntime.DeviceSynchronize(), "ICICLE NTT setup sync")

	b.SetBytes(int64(input.Len() * input.SizeOfElement()))
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		activateICICLECUDA(b)
		requireICICLESuccess(b, nttFn(deviceInput, icicleCore.KForward, &cfg, deviceOutput), "ICICLE NTT")
		requireICICLESuccess(b, icicleRuntime.DeviceSynchronize(), "ICICLE NTT sync")
	}
}

func benchICICLEBLS12377MSM(
	b *testing.B,
	scalars icicleCore.HostSlice[icicleBLS12377.ScalarField],
	points icicleCore.HostSlice[icicleBLS12377.Affine],
) {
	b.Helper()
	var deviceScalars icicleCore.DeviceSlice
	scalars.CopyToDevice(&deviceScalars, true)
	defer func() {
		activateICICLECUDA(b)
		requireICICLESuccess(b, deviceScalars.Free(), "freeing ICICLE MSM scalars")
	}()

	var devicePoints icicleCore.DeviceSlice
	points.CopyToDevice(&devicePoints, true)
	defer func() {
		activateICICLECUDA(b)
		requireICICLESuccess(b, devicePoints.Free(), "freeing ICICLE MSM points")
	}()

	var projective icicleBLS12377.Projective
	var deviceOutput icicleCore.DeviceSlice
	_, err := deviceOutput.Malloc(projective.Size(), 1)
	requireICICLESuccess(b, err, "allocating ICICLE MSM output")
	defer func() {
		activateICICLECUDA(b)
		requireICICLESuccess(b, deviceOutput.Free(), "freeing ICICLE MSM output")
	}()

	cfg := icicleBLS12377MSM.GetDefaultMSMConfig()
	cfg.StreamHandle = nil
	cfg.IsAsync = false
	cfg.AreScalarsMontgomeryForm = true
	cfg.AreBasesMontgomeryForm = true
	requireICICLESuccess(b, icicleRuntime.DeviceSynchronize(), "ICICLE MSM setup sync")

	b.SetBytes(int64(scalars.Len() * scalars.SizeOfElement()))
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		activateICICLECUDA(b)
		requireICICLESuccess(b, icicleBLS12377MSM.Msm(deviceScalars, devicePoints, &cfg, deviceOutput), "ICICLE MSM")
		requireICICLESuccess(b, icicleRuntime.DeviceSynchronize(), "ICICLE MSM sync")
	}
}

func initICICLEBench(tb testing.TB) {
	tb.Helper()
	if os.Getenv("ICICLE_BACKEND_INSTALL_DIR") == "" {
		tb.Skip("ICICLE_BACKEND_INSTALL_DIR must point to the ICICLE install lib/backend directory")
	}
	icicleBenchOnce.Do(func() {
		requireICICLESuccess(tb, icicleRuntime.LoadBackendFromEnvOrDefault(), "loading ICICLE backend")
	})
	activateICICLECUDA(tb)
}

func activateICICLECUDA(tb testing.TB) {
	tb.Helper()
	device := icicleRuntime.CreateDevice("CUDA", 0)
	requireICICLESuccess(tb, icicleRuntime.SetDevice(&device), "selecting ICICLE CUDA device")
}

func initICICLEBN254Domain(tb testing.TB, n int) {
	tb.Helper()
	var rouIcicle icicleBN254.ScalarField
	requireICICLESuccess(
		tb,
		icicleRuntime.EIcicleError(setICICLEBN254RootRaw(unsafe.Pointer(&rouIcicle), n)),
		"building BN254 ICICLE root of unity",
	)
	cfg, cleanup := icicleNTTDomainConfig()
	defer cleanup()
	requireICICLESuccess(
		tb,
		icicleRuntime.EIcicleError(
			initICICLEBN254DomainRaw(unsafe.Pointer(&rouIcicle), unsafe.Pointer(&cfg)),
		),
		"initializing BN254 ICICLE NTT domain",
	)
}

func initICICLEBLS12377Domain(tb testing.TB, n int) {
	tb.Helper()
	var rouIcicle icicleBLS12377.ScalarField
	requireICICLESuccess(
		tb,
		icicleRuntime.EIcicleError(setICICLEBLS12377RootRaw(unsafe.Pointer(&rouIcicle), n)),
		"building BLS12-377 ICICLE root of unity",
	)
	cfg, cleanup := icicleNTTDomainConfig()
	defer cleanup()
	requireICICLESuccess(
		tb,
		icicleRuntime.EIcicleError(
			initICICLEBLS12377DomainRaw(unsafe.Pointer(&rouIcicle), unsafe.Pointer(&cfg)),
		),
		"initializing BLS12-377 ICICLE NTT domain",
	)
}

func initICICLEBW6761Domain(tb testing.TB, n int) {
	tb.Helper()
	var rouIcicle icicleBW6761.ScalarField
	requireICICLESuccess(
		tb,
		icicleRuntime.EIcicleError(setICICLEBW6761RootRaw(unsafe.Pointer(&rouIcicle), n)),
		"building BW6-761 ICICLE root of unity",
	)
	cfg, cleanup := icicleNTTDomainConfig()
	defer cleanup()
	requireICICLESuccess(
		tb,
		icicleRuntime.EIcicleError(
			initICICLEBW6761DomainRaw(unsafe.Pointer(&rouIcicle), unsafe.Pointer(&cfg)),
		),
		"initializing BW6-761 ICICLE NTT domain",
	)
}

func icicleNTTDomainConfig() (icicleCore.NTTInitDomainConfig, func()) {
	cfg := icicleCore.GetDefaultNTTInitDomainConfig()
	ext := icicleConfig.Create()
	ext.SetBool(icicleCore.CUDA_NTT_FAST_TWIDDLES_MODE, true)
	cfg.Ext = ext.AsUnsafePointer()
	return cfg, func() { icicleConfig.Delete(ext) }
}

func icicleBLS12377ScalarsFromGnark(
	scalars []blsfr.Element,
) icicleCore.HostSlice[icicleBLS12377.ScalarField] {
	out := make(icicleCore.HostSlice[icicleBLS12377.ScalarField], len(scalars))
	raw := rawBLS12377(scalars)
	for i := range scalars {
		out[i].FromLimbs(splitUint64ToUint32(raw[i*4 : (i+1)*4]))
	}
	return out
}

func icicleBLS12377AffineFromGnark(
	points []bls12377.G1Affine,
) icicleCore.HostSlice[icicleBLS12377.Affine] {
	out := make(icicleCore.HostSlice[icicleBLS12377.Affine], len(points))
	for i := range points {
		raw := rawBLS12377G1(&points[i])
		out[i].FromLimbs(
			splitUint64ToUint32(raw[:6]),
			splitUint64ToUint32(raw[6:]),
		)
	}
	return out
}

func splitUint64ToUint32(v []uint64) []uint32 {
	out := make([]uint32, len(v)*2)
	for i, limb := range v {
		out[2*i] = uint32(limb)
		out[2*i+1] = uint32(limb >> 32)
	}
	return out
}

func requireICICLESuccess(tb testing.TB, err icicleRuntime.EIcicleError, msg string) {
	tb.Helper()
	require.Equal(tb, icicleRuntime.Success, err, "%s: %s", msg, err.AsString())
}
