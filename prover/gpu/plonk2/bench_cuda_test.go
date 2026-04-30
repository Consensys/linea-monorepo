//go:build cuda

package plonk2

import (
	"fmt"
	"os"
	"testing"

	blsfr "github.com/consensys/gnark-crypto/ecc/bls12-377/fr"
	gnarkunsafe "github.com/consensys/gnark-crypto/utils/unsafe"
	"github.com/stretchr/testify/require"

	"github.com/consensys/linea-monorepo/prover/gpu"
	oldplonk "github.com/consensys/linea-monorepo/prover/gpu/plonk"
)

const benchSRSRootDir = "/home/ubuntu/dev/go/src/github.com/consensys/linea-monorepo/prover/prover-assets/kzgsrs"
const benchBLS12377Scalars = benchSRSRootDir + "/random_scalars_134217728_bls12377.memdump"

func BenchmarkG1MSMCommitRawBLS12377(b *testing.B) {
	dev, err := gpu.New()
	require.NoError(b, err, "creating CUDA device should succeed")
	defer dev.Close()

	store, err := oldplonk.NewSRSStore(benchSRSRootDir)
	require.NoError(b, err, "creating SRS store should succeed")

	for _, n := range []int{1 << 14, 1 << 16} {
		b.Run(fmt.Sprintf("n=%s", benchFormatSize(n)), func(b *testing.B) {
			points, err := store.LoadPointsAffine(n, true)
			require.NoError(b, err, "loading affine SRS should succeed")
			msm, err := NewG1MSM(dev, CurveBLS12377, rawBLS12377G1Slice(points))
			require.NoError(b, err, "creating resident MSM should succeed")
			defer msm.Close()

			scalars := loadBenchBLS12377Scalars(b, n)
			rawScalars := cloneRaw(rawBLS12377(scalars))
			require.NoError(b, dev.Sync(), "device sync before benchmark should succeed")

			b.SetBytes(int64(n) * int64(blsfr.Bytes))
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_, err := msm.CommitRaw(rawScalars)
				require.NoError(b, err, "resident MSM commitment should succeed")
			}
		})
	}
}

func BenchmarkCompareBLS12377MSMPlonkVsPlonk2_CUDA(b *testing.B) {
	dev, err := gpu.New()
	require.NoError(b, err, "creating CUDA device should succeed")
	defer dev.Close()

	store, err := oldplonk.NewSRSStore(benchSRSRootDir)
	require.NoError(b, err, "creating SRS store should succeed")

	for _, n := range []int{1 << 10, 1 << 12, 1 << 14, 1 << 16, 1 << 18, 1 << 20} {
		scalars := loadBenchBLS12377Scalars(b, n)
		rawScalars := cloneRaw(rawBLS12377(scalars))

		b.Run(fmt.Sprintf("gpu-plonk/n=%s", benchFormatSize(n)), func(b *testing.B) {
			pts, err := store.LoadTEPointsPinned(n, true)
			require.NoError(b, err, "loading pinned TE SRS should succeed")
			msm, err := oldplonk.NewG1MSM(dev, pts)
			require.NoError(b, err, "creating gpu/plonk MSM should succeed")
			defer msm.Close()
			require.NoError(b, msm.PinWorkBuffers(), "pinning gpu/plonk MSM work buffers should succeed")
			require.NoError(b, dev.Sync(), "device sync before benchmark should succeed")

			b.SetBytes(int64(n) * int64(blsfr.Bytes))
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_ = msm.MultiExp(scalars)
			}
		})

		b.Run(fmt.Sprintf("gpu-plonk2/n=%s", benchFormatSize(n)), func(b *testing.B) {
			points, err := store.LoadPointsAffine(n, true)
			require.NoError(b, err, "loading affine SRS should succeed")
			msm, err := NewG1MSM(dev, CurveBLS12377, rawBLS12377G1Slice(points))
			require.NoError(b, err, "creating gpu/plonk2 MSM should succeed")
			defer msm.Close()
			require.NoError(b, msm.PinWorkBuffers(), "pinning gpu/plonk2 MSM work buffers should succeed")
			require.NoError(b, dev.Sync(), "device sync before benchmark should succeed")

			b.SetBytes(int64(n) * int64(blsfr.Bytes))
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_, err := msm.CommitRaw(rawScalars)
				require.NoError(b, err, "gpu/plonk2 MSM should succeed")
			}
		})
	}
}

func loadBenchBLS12377Scalars(tb testing.TB, n int) []blsfr.Element {
	tb.Helper()
	f, err := os.Open(benchBLS12377Scalars)
	require.NoError(tb, err, "opening benchmark scalar dump should succeed")
	defer f.Close()

	require.NoError(tb, gnarkunsafe.ReadMarker(f), "reading scalar dump marker should succeed")
	scalars, _, err := gnarkunsafe.ReadSlice[[]blsfr.Element](f, n)
	require.NoError(tb, err, "reading benchmark scalars should succeed")
	return scalars
}

func benchFormatSize(n int) string {
	if n >= 1<<20 {
		return fmt.Sprintf("%dM", n>>20)
	}
	return fmt.Sprintf("%dK", n>>10)
}
