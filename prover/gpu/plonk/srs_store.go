//go:build cuda

// srs_store.go indexes and loads KZG SRS assets from disk.
//
// Goal:
//   - Keep lookup logic deterministic and cheap.
//   - Hide file naming/routing details from prover code.
//   - Reuse one simple selection rule for CPU and GPU paths.
//
// File classes:
//  1. CPU gnark memdumps:
//     kzg_srs_{canonical|lagrange}_{n}_{curve}_{flavor}.memdump
//  2. GPU TE-XY memdumps:
//     kzg_srs_{canonical|lagrange}_TE_XY_{n}_{curve}_{flavor}.memdump
//
// Selection rule (same for both classes):
//
//	canonical request (size n): pick smallest file with size >= n
//	lagrange request  (size n): pick file with size == n
//
// Why:
//   - Canonical SRS is reused by commitments with a small +k margin, so >= n is
//     valid and avoids unnecessary regeneration.
//   - Lagrange SRS must match the domain size exactly.
//
// Lookup flow:
//
//	request(curve, n, canonical?)
//	          |
//	          v
//	 +-------------------+
//	 | entries[curve]    |  (or teEntries[curve] for GPU path)
//	 +-------------------+
//	          |
//	          v
//	 linear scan on sorted sizes
//	          |
//	          v
//	  first matching fsEntry
//
// Pseudocode:
//
//	for e in entries_sorted_by_size:
//	    if e.kind != wanted_kind: continue
//	    if wanted_kind == canonical and e.size >= n: return e
//	    if wanted_kind == lagrange  and e.size == n: return e
//	return not_found
package plonk

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os"
	"regexp"
	"sort"
	"strconv"
	"unsafe"

	"github.com/consensys/gnark-crypto/ecc"
	bls12377 "github.com/consensys/gnark-crypto/ecc/bls12-377"
	"github.com/consensys/gnark-crypto/ecc/bls12-377/fp"
	bls12377kzg "github.com/consensys/gnark-crypto/ecc/bls12-377/kzg"
	"github.com/consensys/gnark-crypto/kzg"
	gnarkplonk "github.com/consensys/gnark/backend/plonk"
	"github.com/consensys/gnark/constraint"
)

type SRSStore struct {
	rootDir   string
	entries   map[ecc.ID][]fsEntry
	teEntries map[ecc.ID][]fsEntry
}

type fsEntry struct {
	isCanonical bool
	size        int
	path        string
}

var (
	srsRegexp   = regexp.MustCompile(`^kzg_srs_(canonical|lagrange)_(\d+)_(bls12377|bn254|bw6761)_(aleo|aztec|celo)\.memdump$`)
	srsTERegexp = regexp.MustCompile(`^kzg_srs_(canonical|lagrange)_TE_XY_(\d+)_(bls12377|bn254|bw6761)_(aleo|aztec|celo)\.memdump$`)
)

// NewSRSStore creates a new SRSStore that indexes both original kzg memdump files
// and pre-converted TE_XY memdump files.
func NewSRSStore(rootDir string) (*SRSStore, error) {
	dir, err := os.ReadDir(rootDir)
	if err != nil {
		return nil, err
	}

	store := &SRSStore{
		rootDir:   rootDir,
		entries:   make(map[ecc.ID][]fsEntry),
		teEntries: make(map[ecc.ID][]fsEntry),
	}
	store.entries[ecc.BLS12_377] = []fsEntry{}
	store.teEntries[ecc.BLS12_377] = []fsEntry{}

	for _, entry := range dir {
		if entry.IsDir() {
			continue
		}
		fileName := entry.Name()
		path := rootDir + "/" + fileName

		// Try TE_XY regex first (more specific)
		if matches := srsTERegexp.FindStringSubmatch(fileName); matches != nil {
			curveID := parseCurveID(matches[3])
			if curveID == 0 {
				continue
			}
			size, _ := strconv.Atoi(matches[2])
			store.teEntries[curveID] = append(store.teEntries[curveID], fsEntry{
				isCanonical: matches[1] == "canonical",
				size:        size,
				path:        path,
			})
			continue
		}

		// Try original SRS regex
		if matches := srsRegexp.FindStringSubmatch(fileName); matches != nil {
			curveID := parseCurveID(matches[3])
			if curveID == 0 {
				continue
			}
			size, _ := strconv.Atoi(matches[2])
			store.entries[curveID] = append(store.entries[curveID], fsEntry{
				isCanonical: matches[1] == "canonical",
				size:        size,
				path:        path,
			})
		}
	}

	for _, entries := range store.entries {
		sort.Slice(entries, func(i, j int) bool { return entries[i].size < entries[j].size })
	}
	for _, entries := range store.teEntries {
		sort.Slice(entries, func(i, j int) bool { return entries[i].size < entries[j].size })
	}

	return store, nil
}

func parseCurveID(s string) ecc.ID {
	switch s {
	case "bls12377":
		return ecc.BLS12_377
	default:
		return 0
	}
}

func findInEntries(entries []fsEntry, size int, isCanonical bool) *fsEntry {
	for i := range entries {
		e := &entries[i]
		if e.isCanonical != isCanonical {
			continue
		}
		if isCanonical {
			if e.size >= size {
				return e
			}
		} else {
			if e.size == size {
				return e
			}
		}
	}
	return nil
}

func (store *SRSStore) findEntry(curveID ecc.ID, size int, isCanonical bool) *fsEntry {
	return findInEntries(store.entries[curveID], size, isCanonical)
}

func (store *SRSStore) findTEEntry(curveID ecc.ID, size int, isCanonical bool) *fsEntry {
	return findInEntries(store.teEntries[curveID], size, isCanonical)
}

func (store *SRSStore) loadSRS(ctx context.Context, curveID ecc.ID, entry *fsEntry, size int) (kzg.SRS, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	f, err := os.Open(entry.path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	res := kzg.NewSRS(curveID)
	r := bufio.NewReaderSize(f, 1<<20)
	if err := res.ReadDump(r, size); err != nil {
		return nil, err
	}
	return res, nil
}

// GetSRSCPU returns canonical + lagrange SRS for CPU-oriented tests/workflows.
func (store *SRSStore) GetSRSCPU(ctx context.Context, ccs constraint.ConstraintSystem) (kzg.SRS, kzg.SRS, error) {
	sizeCanonical, sizeLagrange := gnarkplonk.SRSSize(ccs)
	curveID := ecc.BLS12_377

	canonicalEntry := store.findEntry(curveID, sizeCanonical, true)
	if canonicalEntry == nil {
		return nil, nil, fmt.Errorf("canonical SRS not found for curve %s and size >= %d", curveID, sizeCanonical)
	}
	lagrangeEntry := store.findEntry(curveID, sizeLagrange, false)
	if lagrangeEntry == nil {
		return nil, nil, fmt.Errorf("lagrange SRS not found for curve %s and exact size %d", curveID, sizeLagrange)
	}

	canonicalSRS, err := store.loadSRS(ctx, curveID, canonicalEntry, sizeCanonical)
	if err != nil {
		return nil, nil, err
	}
	lagrangeSRS, err := store.loadSRS(ctx, curveID, lagrangeEntry, sizeLagrange)
	if err != nil {
		return nil, nil, err
	}

	return canonicalSRS, lagrangeSRS, nil
}

// GetSRSGPU returns canonical SRS as pre-converted TE points for GPU proving.
func (store *SRSStore) GetSRSGPU(ctx context.Context, ccs constraint.ConstraintSystem) (canonical []G1TEPoint, err error) {
	sizeCanonical, _ := gnarkplonk.SRSSize(ccs)

	canonical, err = store.LoadTEPoints(sizeCanonical, true)
	if err != nil {
		return nil, fmt.Errorf("load canonical TE SRS: %w", err)
	}
	return canonical, nil
}

// GetSRSGPUPinned loads canonical SRS directly into pinned (page-locked)
// host memory, ready for GPU DMA. No intermediate Go heap allocation.
// Caller must Free() the returned G1MSMPoints when done.
func (store *SRSStore) GetSRSGPUPinned(ctx context.Context, ccs constraint.ConstraintSystem) (canonical *G1MSMPoints, err error) {
	sizeCanonical, _ := gnarkplonk.SRSSize(ccs)

	canonical, err = store.LoadTEPointsPinned(sizeCanonical, true)
	if err != nil {
		return nil, fmt.Errorf("load canonical TE SRS pinned: %w", err)
	}
	return canonical, nil
}

// GetSRS keeps backward compatibility with existing call sites that expect both.
func (store *SRSStore) GetSRS(ctx context.Context, ccs constraint.ConstraintSystem) (kzg.SRS, kzg.SRS, error) {
	return store.GetSRSCPU(ctx, ccs)
}

// ---------------------------------------------------------------------------
// TE_XY memdump: raw G1TEPoint data (96 bytes/point, no header)
// ---------------------------------------------------------------------------

// LoadTEPoints loads pre-converted TE points from a raw memdump into a Go slice.
// For canonical SRS any file with size >= n works (returns first n points).
// For lagrange SRS exact size match is required.
func (store *SRSStore) LoadTEPoints(n int, isCanonical bool) ([]G1TEPoint, error) {
	entry := store.findTEEntry(ecc.BLS12_377, n, isCanonical)
	if entry == nil {
		kind := "lagrange"
		if isCanonical {
			kind = "canonical"
		}
		return nil, fmt.Errorf("TE_XY %s SRS not found for size %d", kind, n)
	}

	f, err := os.Open(entry.path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	return readG1TEPointsUnsafe(f, n)
}

// LoadTEPointsPinned loads pre-converted TE points directly into pinned (page-locked)
// host memory, bypassing any Go heap allocation. This is the fastest production path:
// file → pinned memory → GPU DMA, with no intermediate copies.
func (store *SRSStore) LoadTEPointsPinned(n int, isCanonical bool) (*G1MSMPoints, error) {
	entry := store.findTEEntry(ecc.BLS12_377, n, isCanonical)
	if entry == nil {
		kind := "lagrange"
		if isCanonical {
			kind = "canonical"
		}
		return nil, fmt.Errorf("TE_XY %s SRS not found for size %d", kind, n)
	}

	f, err := os.Open(entry.path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	return ReadG1TEPointsPinned(f, n)
}

// LoadPointsAffine loads affine points from disk
func (store *SRSStore) LoadPointsAffine(n int, isCanonical bool) ([]bls12377.G1Affine, error) {
	entry := store.findEntry(ecc.BLS12_377, n, isCanonical)
	if entry == nil {
		return nil, fmt.Errorf("original SRS not found for size %d to convert to affine", n)
	}

	f, err := os.Open(entry.path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	res := kzg.NewSRS(ecc.BLS12_377)
	// r := bufio.NewReaderSize(f, 1<<20)
	if err := res.ReadDump(f, entry.size); err != nil {
		return nil, err
	}
	g1Points := res.(*bls12377kzg.SRS).Pk.G1
	return g1Points[:n], nil
}

// baseName returns the last element of a path.
func baseName(path string) string {
	for i := len(path) - 1; i >= 0; i-- {
		if path[i] == '/' {
			return path[i+1:]
		}
	}
	return path
}

// readG1TEPointsUnsafe reads n G1TEPoints by casting raw bytes directly into the slice.
func readG1TEPointsUnsafe(r io.Reader, n int) ([]G1TEPoint, error) {
	points := make([]G1TEPoint, n)
	buf := unsafe.Slice((*byte)(unsafe.Pointer(&points[0])), n*g1TEPointSize)
	if _, err := io.ReadFull(r, buf); err != nil {
		return nil, err
	}
	return points, nil
}

// writeG1TEPointsUnsafe writes G1TEPoints as raw bytes.
func writeG1TEPointsUnsafe(w io.Writer, points []G1TEPoint) error {
	if len(points) == 0 {
		return nil
	}
	buf := unsafe.Slice((*byte)(unsafe.Pointer(&points[0])), len(points)*g1TEPointSize)
	_, err := w.Write(buf)
	return err
}

// tePointsToAffine batch-converts compact TE XY points back to SW affine.
func tePointsToAffine(tePoints []G1TEPoint) []bls12377.G1Affine {
	n := len(tePoints)
	affine := make([]bls12377.G1Affine, n)
	for i, tp := range tePoints {
		xTE := fp.Element{tp[0], tp[1], tp[2], tp[3], tp[4], tp[5]}
		yTE := fp.Element{tp[6], tp[7], tp[8], tp[9], tp[10], tp[11]}
		// Reconstruct extended coords: T = X*Y, Z = 1
		var tTE, zTE fp.Element
		tTE.Mul(&xTE, &yTE)
		zTE.SetOne()
		var w [24]uint64
		w[0], w[1], w[2], w[3], w[4], w[5] = xTE[0], xTE[1], xTE[2], xTE[3], xTE[4], xTE[5]
		w[6], w[7], w[8], w[9], w[10], w[11] = yTE[0], yTE[1], yTE[2], yTE[3], yTE[4], yTE[5]
		w[12], w[13], w[14], w[15], w[16], w[17] = tTE[0], tTE[1], tTE[2], tTE[3], tTE[4], tTE[5]
		w[18], w[19], w[20], w[21], w[22], w[23] = zTE[0], zTE[1], zTE[2], zTE[3], zTE[4], zTE[5]
		jac := te2jac(w)
		affine[i].FromJacobian(&jac)
	}
	return affine
}
