package circuits

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"

	"github.com/consensys/gnark-crypto/ecc"
	"github.com/consensys/gnark-crypto/kzg"
	"github.com/consensys/gnark/backend/plonk"
	"github.com/consensys/gnark/constraint"
	"github.com/sirupsen/logrus"

	kzg377 "github.com/consensys/gnark-crypto/ecc/bls12-377/kzg"
	kzg254 "github.com/consensys/gnark-crypto/ecc/bn254/kzg"
	kzgbw6 "github.com/consensys/gnark-crypto/ecc/bw6-761/kzg"
)

type SRSStore struct {
	entries map[ecc.ID][]fsEntry // for each curve, list SRS of different sizes and types (canonical/lagrange)
}

type fsEntry struct {
	isCanonical bool   // canonical or lagrange
	size        int    // domain size
	path        string // path to the file it can be loaded from
	cached      kzg.SRS
}

// NewSRSStore creates a new SRSStore
func NewSRSStore(rootDir string) (*SRSStore, error) {
	// list all the files in rootDir
	// for each file, make a fsEntry but do not load the SRS (lazy loaded on demand)
	// store the fsEntry in map[string]fsEntry, with the key being the file name

	dir, err := os.ReadDir(rootDir)
	if err != nil {
		return nil, err
	}

	srsStore := &SRSStore{
		entries: make(map[ecc.ID][]fsEntry),
	}
	srsStore.entries[ecc.BLS12_377] = []fsEntry{}
	srsStore.entries[ecc.BN254] = []fsEntry{}
	srsStore.entries[ecc.BW6_761] = []fsEntry{}

	srsRegexp := regexp.MustCompile(`^(kzg_srs)_(canonical|lagrange)_(\d+)_(bls12377|bn254|bw6761)_(aleo|aztec|celo)\.memdump$`)

	for _, entry := range dir {
		if entry.IsDir() {
			continue
		}
		// parse the file name
		// create a fsEntry
		// store it in the map

		fileName := entry.Name()
		matches := srsRegexp.FindStringSubmatch(fileName)
		if matches == nil {
			continue
		}

		isCanonical := matches[2] == "canonical"
		size, _ := strconv.Atoi(matches[3])
		var curveID ecc.ID
		switch matches[4] {
		case "bls12377":
			curveID = ecc.BLS12_377
		case "bn254":
			curveID = ecc.BN254
		case "bw6761":
			curveID = ecc.BW6_761
		default:
			return nil, errors.New("curve not supported")
		}

		srsStore.entries[curveID] = append(srsStore.entries[curveID], fsEntry{
			isCanonical: isCanonical,
			size:        size,
			path:        filepath.Join(rootDir, fileName),
		})

	}

	// sort the entries by size
	for _, entries := range srsStore.entries {
		sort.Slice(entries, func(i, j int) bool {
			return entries[i].size < entries[j].size
		})
	}

	return srsStore, nil
}

func (store *SRSStore) GetSRS(_ context.Context, ccs constraint.ConstraintSystem) (kzg.SRS, kzg.SRS, error) {
	sizeCanonical, sizeLagrange := plonk.SRSSize(ccs)
	curveID := fieldToCurve(ccs.Field())

	var lagrangeSRS kzg.SRS
	loadLagrangeErr := make(chan error, 2)
	go func() { // attempt to find lagrange SRS
		var (
			err error
			f   *os.File
		)
		for i := range store.entries[curveID] {
			entry := &store.entries[curveID][i] // reference in case we need to update the cache
			if !entry.isCanonical && entry.size == sizeLagrange {
				if entry.cached != nil {
					lagrangeSRS = entry.cached
					break
				}
				lagrangeSRS = kzg.NewSRS(curveID)
				if f, err = os.Open(entry.path); err == nil {
					err = errors.Join(lagrangeSRS.ReadDump(f), f.Close())
				}

				entry.cached = lagrangeSRS

				break
			}
		}

		if err != nil {
			logrus.WithError(err).Warn("failed to load lagrange SRS")
		} else if lagrangeSRS == nil {
			logrus.Warn("lagrange SRS not found")
		}
		loadLagrangeErr <- err
		close(loadLagrangeErr)
	}()

	// find the canonical srs
	var canonicalSRS kzg.SRS

	for i := range store.entries[curveID] {
		entry := &store.entries[curveID][i] // reference in case we need to update the cache
		if entry.isCanonical && entry.size >= sizeCanonical {
			if entry.cached != nil {
				canonicalSRS = entry.cached
				break
			}

			canonicalSRS = kzg.NewSRS(curveID)
			f, err := os.Open(entry.path)
			if err != nil {
				return nil, nil, err
			}
			err = errors.Join(
				canonicalSRS.ReadDump(f, sizeCanonical),
				f.Close(),
			)
			if err != nil {
				return nil, nil, err
			}

			entry.cached = canonicalSRS

			break
		}
	}

	if canonicalSRS == nil {
		return nil, nil, fmt.Errorf("could not find canonical SRS for curve %s and size %d", curveID, sizeCanonical)
	}

	// wait for lagrange SRS to be loaded
	if loadLagrangeErr := <-loadLagrangeErr; lagrangeSRS == nil || loadLagrangeErr != nil {
		// we can compute it from the canonical one.
		if sizeCanonical < sizeLagrange {
			panic("canonical SRS is smaller than lagrange SRS")
		}
		logrus.Debugf("computing lagrange SRS from canonical SRS %d -> %d\n", sizeCanonical, sizeLagrange)
		var err error
		if lagrangeSRS, err = toLagrange(canonicalSRS, sizeLagrange); err != nil {
			return nil, nil, err
		}
	}

	return canonicalSRS, lagrangeSRS, nil
}

func toLagrange(srs kzg.SRS, sizeLagrange int) (kzg.SRS, error) {
	var err error
	switch srs := srs.(type) {
	case *kzg254.SRS:
		lagrange := &kzg254.SRS{}
		lagrange.Pk.G1, err = kzg254.ToLagrangeG1(srs.Pk.G1[:sizeLagrange])
		return lagrange, err
	case *kzg377.SRS:
		lagrange := &kzg377.SRS{}
		lagrange.Pk.G1, err = kzg377.ToLagrangeG1(srs.Pk.G1[:sizeLagrange])
		return lagrange, err
	case *kzgbw6.SRS:
		lagrange := &kzgbw6.SRS{}
		lagrange.Pk.G1, err = kzgbw6.ToLagrangeG1(srs.Pk.G1[:sizeLagrange])
		return lagrange, err
	default:
		panic("unknown SRS type")
	}
}
