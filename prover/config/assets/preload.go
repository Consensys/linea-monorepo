package assets

import (
	"context"
	"errors"
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
	// Global map[string]*sync.Once
	onceMap sync.Map
	// keep mappings so GC doesn't drop them and to keep pages "active"
	residentMu       sync.Mutex
	residentMappings = make([][]byte, 0, 64)
)

// PreloadOptions controls behaviour; tune for your environment
type PreloadOptions struct {
	ChunkSize          int64         // chunk size when doing fallback mapping-touch
	LargeFileThreshold int64         // >= this size we use fadvise instead of mmap-populate
	PerFileTimeout     time.Duration // timeout per file (used by fallbacks)
	Populate           bool          // whether to use MAP_POPULATE for mmap path
}

func DefaultPreloadOptions() *PreloadOptions {
	return &PreloadOptions{
		ChunkSize:          64 << 20, // 64 MiB
		LargeFileThreshold: 10 << 30, // 10 GiB
		PerFileTimeout:     120 * time.Second,
		Populate:           true, // use MAP_POPULATE for mmap
	}
}

// PreLoadOnceForJob will preload (mmap-populate or fadvise) only once per jobName
// when called concurrently multiple times only the first caller will execute the work.
// resolverFn should return the list of asset paths for that jobName.
func PreloadOnceForJob(
	ctx context.Context, jobName string,
	resolverFn func() ([]string, []string, error),
	opts *PreloadOptions, logger *logrus.Entry,
) error {
	if jobName == "" {
		return fmt.Errorf("empty jobName")
	}
	if opts == nil {
		opts = DefaultPreloadOptions()
	}
	if logger == nil {
		logger = logrus.NewEntry(logrus.StandardLogger())
	}

	val, _ := onceMap.LoadOrStore(jobName, &sync.Once{})
	once, ok := val.(*sync.Once)
	if !ok {
		return errors.New("cannot cast to *sync.Once")
	}

	var preErr error
	once.Do(func() {
		paths, critical, err := resolverFn()
		if err != nil {
			preErr = fmt.Errorf("resolver error for job %s: %w", jobName, err)
			return
		}
		if len(paths) == 0 {
			logger.Infof("no assets to preload for job %s", jobName)
			return
		}

		logger.Infof("preload (once) starting for job=%s: total assets=%d critical=%d",
			jobName, len(paths), len(critical))

		for _, p := range paths {
			// Context cancellation
			select {
			case <-ctx.Done():
				preErr = fmt.Errorf("preload canceled for job %s: %w", jobName, ctx.Err())
				return
			default:
			}

			if err := preloadPath(p, opts, logger); err != nil {
				logger.Warnf("preload failed for %s: %v", p, err)
			}
		}

		logger.Infof("preload (once) finished for job=%s", jobName)
	})

	return preErr
}

// preloadPath applies the correct preload strategy for a single file.
// It keeps the same logic as before but isolates it for readability.
func preloadPath(p string, opts *PreloadOptions, logger *logrus.Entry) error {
	p = filepath.Clean(p)

	fi, err := os.Stat(p)
	if err != nil {
		return fmt.Errorf("stat failed: %w", err)
	}
	size := fi.Size()
	if size == 0 {
		logger.Debugf("preload: skipping empty file %s", p)
		return nil
	}

	// 1. Large file → fadvise
	if opts.LargeFileThreshold > 0 && size >= opts.LargeFileThreshold {
		if err := preloadLarge(p); err != nil {
			return fmt.Errorf("fadvise failed: %w", err)
		}
		logger.Infof("preload: fadvise(WILLNEED) applied on %s (size=%d)", p, size)
		logResidency(p, "fadvise", logger)
		return nil
	}

	// 2. Otherwise → mmap+MAP_POPULATE (if enabled)
	if opts.Populate {
		if err := preLoadNormal(p); err != nil {
			logger.Warnf("preload: mmap-populate failed for %s: %v", p, err)

			// fallback → fadvise
			if err2 := preloadLarge(p); err2 != nil {
				return fmt.Errorf("fallback fadvise failed: %v (original mmap error: %v)", err2, err)
			}
			logger.Infof("preload: fallback fadvise succeeded for %s", p)
			logResidency(p, "fallback", logger)
			return nil
		}

		logger.Infof("preload: mmap-populate succeeded for %s (size=%d)", p, size)
		logResidency(p, "mmap", logger)
		return nil
	}

	// 3. Populate disabled → fadvise only
	if err := preloadLarge(p); err != nil {
		return fmt.Errorf("fadvise failed: %w", err)
	}
	logResidency(p, "fadvise", logger)
	return nil
}

// preloadLarge tries posix_fadvise(FADV_WILLNEED).
// Best-effort: returns nil if it succeeds, returns error if syscall returns error.
func preloadLarge(path string) error {
	f, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("open: %w", err)
	}
	defer f.Close()
	fd := int(f.Fd())

	// unix.Fadvise returns error only on failure; treat SUCCEED -> nil
	// This tells the kernel: the user plans to read this file soon and requests reading it into page cache in background
	// It is asynchronous and non-blocking — it just hints to the kernel.
	// Effect: the kernel may start preloading pages in background via readahead.
	if err := unix.Fadvise(fd, 0, 0, unix.FADV_WILLNEED); err != nil {
		return fmt.Errorf("fadvise WILLNEED: %w", err)
	}
	return nil
}

// preLoadNormal mmap()s the whole file with MAP_POPULATE,
// stores the mapping in a global slice (to keep it alive) and returns nil on success.
// The mapping is intentionally NOT unmapped so the mapping stays present until process exit.
// This keeps pages resident and prevents GC from dropping the slice.
func preLoadNormal(path string) error {
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

	// Map a segment of the file into the process’s address space (read-only).
	// OS won’t actually read data yet — it will lazily page-fault when you touch those pages
	data, err := unix.Mmap(int(f.Fd()), 0, size, unix.PROT_READ, flags)
	// once mmap'd, we can close the fd safely; the mapping remains valid.
	_ = f.Close()
	if err != nil {
		return fmt.Errorf("mmap: %w", err)
	}

	// Store mapping so it is not GC'd and to keep pages pinned by active mapping
	residentMu.Lock()
	residentMappings = append(residentMappings, data)
	residentMu.Unlock()
	return nil
}

// logResidency logs how much of the file is currently resident in RAM.
func logResidency(p, phase string, logger *logrus.Entry) {
	if frac, err := residentFractionForPath(p); err == nil {
		logger.Infof("preload: residency after %s %s: %.1f%%", phase, p, frac*100)
	}
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
