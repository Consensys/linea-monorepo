package assets

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"syscall"
	"time"
	"unsafe"

	"github.com/sirupsen/logrus"
	"golang.org/x/sys/unix"
)

// Per-job sync.Once registry
var (
	onceMap sync.Map // map[string]*sync.Once
	// keep mappings so GC doesn't drop them and to keep pages "active"
	residentMu       sync.Mutex
	residentMappings = make([][]byte, 0, 64)
)

// PrefetchOptions controls behaviour; tune for your environment
type PrefetchOptions struct {
	Parallelism        int           // not used by mmap path here, but kept for compatibility
	ChunkSize          int64         // chunk size when doing fallback mapping-touch
	LargeFileThreshold int64         // >= this size we use fadvise instead of mmap-populate
	PerFileTimeout     time.Duration // timeout per file (used by fallbacks)
	Populate           bool          // whether to use MAP_POPULATE for mmap path
}

func DefaultPrefetchOptions() *PrefetchOptions {
	return &PrefetchOptions{
		Parallelism:        4,
		ChunkSize:          64 << 20, // 64 MiB
		LargeFileThreshold: 10 << 30, // 10 GiB
		PerFileTimeout:     120 * time.Second,
		Populate:           true, // use MAP_POPULATE for mmap
	}
}

// PrefetchForJobOnce will prefetch (mmap-populate or fadvise) only once per jobName
// when called concurrently multiple times only the first caller will execute the work.
// cfgRoot is the base assets dir (e.g. cfg.PathForSetup("execution-limitless"))
// resolverFn should return the list of asset paths for that jobName.
// logger is optional (nil => uses default logrus.StandardLogger).
func PrefetchForJobOnce(
	ctx context.Context, jobName string,
	resolverFn func() ([]string, []string, error),
	opts *PrefetchOptions, logger *logrus.Entry,
) error {
	if jobName == "" {
		return fmt.Errorf("empty jobName")
	}
	if opts == nil {
		opts = DefaultPrefetchOptions()
	}
	if logger == nil {
		logger = logrus.NewEntry(logrus.StandardLogger())
	}

	val, _ := onceMap.LoadOrStore(jobName, &sync.Once{})
	once := val.(*sync.Once)

	var preErr error
	once.Do(func() {
		paths, critical, err := resolverFn()
		if err != nil {
			preErr = fmt.Errorf("resolver error for job %s: %w", jobName, err)
			return
		}
		if len(paths) == 0 {
			logger.Infof("no assets to prefetch for job %s", jobName)
			return
		}

		logger.Infof("Prefetch (once) starting for job=%s: total assets=%d critical=%d",
			jobName, len(paths), len(critical))

		for _, p := range paths {
			select {
			case <-ctx.Done():
				preErr = fmt.Errorf("prefetch canceled for job %s: %w", jobName, ctx.Err())
				return
			default:
			}

			p = filepath.Clean(p)

			fi, err := os.Stat(p)
			if err != nil {
				logger.Warnf("prefetch: stat failed %s: %v", p, err)
				continue
			}
			size := fi.Size()
			if size == 0 {
				logger.Debugf("prefetch: skipping empty file %s", p)
				continue
			}

			// Large file → fadvise
			if opts.LargeFileThreshold > 0 && size >= opts.LargeFileThreshold {
				if err := prefetchLargeWithFadvise(p); err != nil {
					logger.Warnf("prefetch: fadvise failed for %s: %v", p, err)
				} else {
					logger.Infof("prefetch: fadvise(WILLNEED) applied on %s (size=%d)", p, size)
					// Check residency after fadvise
					if frac, err := residentFractionForPath(p); err == nil {
						logger.Infof("prefetch: residency after fadvise %s: %.1f%%", p, frac*100)
					}
				}
				continue
			}

			// Otherwise → mmap+MAP_POPULATE
			if opts.Populate {
				if err := prefetchWithMmapPopulate(p); err != nil {
					logger.Warnf("prefetch: mmap-populate failed for %s: %v", p, err)
					if err2 := prefetchLargeWithFadvise(p); err2 != nil {
						logger.Warnf("prefetch: fallback fadvise also failed for %s: %v", p, err2)
					} else {
						logger.Infof("prefetch: fallback fadvise succeeded for %s", p)
						if frac, err := residentFractionForPath(p); err == nil {
							logger.Infof("prefetch: residency after fallback %s: %.1f%%", p, frac*100)
						}
					}
				} else {
					logger.Infof("prefetch: mmap-populate succeeded for %s (size=%d)", p, size)
					// Check residency after mmap populate
					if frac, err := residentFractionForPath(p); err == nil {
						logger.Infof("prefetch: residency after mmap %s: %.1f%%", p, frac*100)
					}
				}
				continue
			}

			// If Populate disabled, use fadvise
			if err := prefetchLargeWithFadvise(p); err != nil {
				logger.Warnf("prefetch: fadvise failed for %s: %v", p, err)
			} else if frac, err := residentFractionForPath(p); err == nil {
				logger.Infof("prefetch: residency after fadvise %s: %.1f%%", p, frac*100)
			}
		}

		logger.Infof("Prefetch (once) finished for job=%s", jobName)
	})

	return preErr
}

// prefetchLargeWithFadvise tries posix_fadvise(FADV_WILLNEED).
// Best-effort: returns nil if it succeeds, returns error if syscall returns error.
func prefetchLargeWithFadvise(path string) error {
	f, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("open: %w", err)
	}
	defer f.Close()
	fd := int(f.Fd())
	// unix.Fadvise returns error only on failure; treat SUCCEED -> nil
	if err := unix.Fadvise(fd, 0, 0, unix.FADV_WILLNEED); err != nil {
		return fmt.Errorf("fadvise WILLNEED: %w", err)
	}
	return nil
}

// prefetchWithMmapPopulate mmap()s the whole file with MAP_POPULATE,
// stores the mapping in a global slice (to keep it alive) and returns nil on success.
// The mapping is intentionally NOT unmapped so the mapping stays present until process exit.
// This keeps pages resident and prevents GC from dropping the slice.
func prefetchWithMmapPopulate(path string) error {
	f, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("open: %w", err)
	}
	// do NOT defer f.Close() — we can close file descriptor after mmap
	fi, err := f.Stat()
	if err != nil {
		_ = f.Close()
		return fmt.Errorf("stat: %w", err)
	}
	size := int(fi.Size())
	if size == 0 {
		_ = f.Close()
		return nil
	}

	// mmap with MAP_POPULATE to fault pages in immediately.
	flags := unix.MAP_SHARED
	flags |= unix.MAP_POPULATE
	data, err := unix.Mmap(int(f.Fd()), 0, size, unix.PROT_READ, flags)
	// once mmap'd, we can close the fd safely; the mapping remains valid.
	_ = f.Close()
	if err != nil {
		return fmt.Errorf("mmap: %w", err)
	}

	// store mapping so it is not GC'd and to keep pages pinned by active mapping
	residentMu.Lock()
	residentMappings = append(residentMappings, data)
	residentMu.Unlock()
	return nil
}

func residentFractionForPath(p string) (float64, error) {
	f, err := os.Open(p)
	if err != nil {
		return 0, err
	}
	defer f.Close()

	fi, err := f.Stat()
	if err != nil {
		return 0, err
	}
	if fi.Size() == 0 {
		return 0, nil
	}

	data, err := unix.Mmap(int(f.Fd()), 0, int(fi.Size()), unix.PROT_READ, unix.MAP_SHARED)
	if err != nil {
		return 0, err
	}
	defer unix.Munmap(data)

	return residentFraction(data)
}

// residentFraction reports how many pages of 'mapped' are resident in RAM (0–1).
// Optional helper useful for diagnostics; returns fraction [0..1] or an error.
func residentFraction(mapped []byte) (float64, error) {
	if len(mapped) == 0 {
		return 0, fmt.Errorf("empty mapping")
	}
	pageSize := os.Getpagesize()
	npages := (len(mapped) + pageSize - 1) / pageSize
	vec := make([]byte, npages)

	// mincore(void *addr, size_t length, unsigned char *vec)
	// It uses the mincore(2) syscall directly, since unix.Mincore may not be defined
	// in all Go versions. Safe no-op if mincore unavailable.
	_, _, errno := syscall.Syscall(syscall.SYS_MINCORE,
		uintptr((uintptr)(unsafe.Pointer(&mapped[0]))),
		uintptr(len(mapped)),
		uintptr(unsafe.Pointer(&vec[0])),
	)
	if errno != 0 {
		return 0, fmt.Errorf("mincore syscall failed: %v", errno)
	}

	var resident int
	for _, b := range vec {
		if b&1 == 1 {
			resident++
		}
	}
	return float64(resident) / float64(npages), nil
}
