//go:build cuda

package plonk2

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
	"sort"
	"strconv"
	"sync"
	"testing"

	"github.com/consensys/gnark-crypto/ecc"
	blskzg "github.com/consensys/gnark-crypto/ecc/bls12-377/kzg"
	bnkzg "github.com/consensys/gnark-crypto/ecc/bn254/kzg"
	bwkzg "github.com/consensys/gnark-crypto/ecc/bw6-761/kzg"
	cryptokzg "github.com/consensys/gnark-crypto/kzg"
	gnarkplonk "github.com/consensys/gnark/backend/plonk"
	"github.com/consensys/gnark/constraint"
	"github.com/stretchr/testify/require"
)

const testSRSRootDir = "/home/ubuntu/dev/go/src/github.com/consensys/linea-monorepo/prover/prover-assets/kzgsrs"

type testSRSEntry struct {
	canonical bool
	size      int
	path      string
}

type testSRSAssetStore struct {
	entries map[ecc.ID][]testSRSEntry
}

var (
	testSRSAssetsOnce  sync.Once
	testSRSAssetsStore *testSRSAssetStore
	testSRSAssetsErr   error
	testSRSAssetsRE    = regexp.MustCompile(
		`^kzg_srs_(canonical|lagrange)_(\d+)_(bls12377|bn254|bw6761)_(aleo|aztec|celo)\.memdump$`,
	)
)

func testSRSAssets(tb testing.TB) *testSRSAssetStore {
	tb.Helper()
	testSRSAssetsOnce.Do(func() {
		testSRSAssetsStore, testSRSAssetsErr = newTestSRSAssetStore(testSRSRootDir)
	})
	require.NoError(tb, testSRSAssetsErr, "loading SRS asset index should succeed")
	return testSRSAssetsStore
}

func newTestSRSAssetStore(root string) (*testSRSAssetStore, error) {
	files, err := os.ReadDir(root)
	if err != nil {
		return nil, err
	}
	store := &testSRSAssetStore{entries: make(map[ecc.ID][]testSRSEntry)}
	for _, file := range files {
		if file.IsDir() {
			continue
		}
		matches := testSRSAssetsRE.FindStringSubmatch(file.Name())
		if matches == nil {
			continue
		}
		curveID, ok := testSRSAssetCurve(matches[3])
		if !ok {
			continue
		}
		size, err := strconv.Atoi(matches[2])
		if err != nil {
			return nil, err
		}
		store.entries[curveID] = append(store.entries[curveID], testSRSEntry{
			canonical: matches[1] == "canonical",
			size:      size,
			path:      root + "/" + file.Name(),
		})
	}
	for curveID := range store.entries {
		sort.Slice(store.entries[curveID], func(i, j int) bool {
			return store.entries[curveID][i].size < store.entries[curveID][j].size
		})
	}
	return store, nil
}

func testSRSAssetCurve(name string) (ecc.ID, bool) {
	switch name {
	case "bn254":
		return ecc.BN254, true
	case "bls12377":
		return ecc.BLS12_377, true
	case "bw6761":
		return ecc.BW6_761, true
	default:
		return 0, false
	}
}

func (s *testSRSAssetStore) loadForCCS(tb testing.TB, ccs constraint.ConstraintSystem) (cryptokzg.SRS, cryptokzg.SRS) {
	tb.Helper()
	canonicalSize, lagrangeSize := gnarkplonk.SRSSize(ccs)
	curveID := testSRSAssetECCID(tb, ccs)
	canonical := s.load(tb, curveID, canonicalSize, true)
	lagrange := s.load(tb, curveID, lagrangeSize, false)
	return canonical, lagrange
}

func testSRSAssetECCID(tb testing.TB, ccs constraint.ConstraintSystem) ecc.ID {
	tb.Helper()
	curve, err := curveFromConstraintSystem(ccs)
	require.NoError(tb, err, "constraint system should use a target curve")
	switch curve {
	case CurveBN254:
		return ecc.BN254
	case CurveBLS12377:
		return ecc.BLS12_377
	case CurveBW6761:
		return ecc.BW6_761
	default:
		tb.Fatalf("unsupported curve %s", curve)
		return 0
	}
}

func (s *testSRSAssetStore) loadBN254(tb testing.TB, n int, canonical bool) *bnkzg.SRS {
	tb.Helper()
	srs, ok := s.load(tb, ecc.BN254, n, canonical).(*bnkzg.SRS)
	require.True(tb, ok, "asset SRS should be BN254")
	return srs
}

func (s *testSRSAssetStore) loadBLS12377(tb testing.TB, n int, canonical bool) *blskzg.SRS {
	tb.Helper()
	srs, ok := s.load(tb, ecc.BLS12_377, n, canonical).(*blskzg.SRS)
	require.True(tb, ok, "asset SRS should be BLS12-377")
	return srs
}

func (s *testSRSAssetStore) loadBW6761(tb testing.TB, n int, canonical bool) *bwkzg.SRS {
	tb.Helper()
	srs, ok := s.load(tb, ecc.BW6_761, n, canonical).(*bwkzg.SRS)
	require.True(tb, ok, "asset SRS should be BW6-761")
	return srs
}

func (s *testSRSAssetStore) load(tb testing.TB, curveID ecc.ID, n int, canonical bool) cryptokzg.SRS {
	tb.Helper()
	entry := s.find(curveID, n, canonical)
	kind := "lagrange"
	if canonical {
		kind = "canonical"
	}
	require.NotNilf(tb, entry, "%s SRS asset for %s and size %d should exist", kind, curveID, n)

	file, err := os.Open(entry.path)
	require.NoError(tb, err, "opening %s SRS asset should succeed", kind)
	defer file.Close()

	srs := cryptokzg.NewSRS(curveID)
	reader := bufio.NewReaderSize(file, 1<<20)
	require.NoError(tb, srs.ReadDump(reader, n), "reading %s SRS asset should succeed", kind)
	return srs
}

func (s *testSRSAssetStore) find(curveID ecc.ID, n int, canonical bool) *testSRSEntry {
	for i := range s.entries[curveID] {
		entry := &s.entries[curveID][i]
		if entry.canonical != canonical {
			continue
		}
		if canonical && entry.size >= n {
			return entry
		}
		if !canonical && entry.size == n {
			return entry
		}
	}
	return nil
}

func TestSRSAssetsContainTargetCurves(t *testing.T) {
	store := testSRSAssets(t)
	for _, curveID := range []ecc.ID{ecc.BN254, ecc.BLS12_377, ecc.BW6_761} {
		t.Run(curveID.String(), func(t *testing.T) {
			require.NotNil(t, store.find(curveID, 256, true), "canonical SRS should exist")
			require.NotNil(t, store.find(curveID, 256, false), "lagrange SRS should exist")
		})
	}
}

func (e testSRSEntry) String() string {
	return fmt.Sprintf("canonical=%t size=%d path=%s", e.canonical, e.size, e.path)
}
