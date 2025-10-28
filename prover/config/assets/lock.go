package assets

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/sirupsen/logrus"
	"golang.org/x/sys/unix"
)

var (
	// lockedMappings holds mmap byte slices keyed by absolute path.
	// We intentionally store the byte slice to keep the mapping alive
	// and to call munlock/munmap on shutdown if needed.
	lockedMappings sync.Map // map[string][]byte
)

// UnlockAllLockedFiles - unlocks all the locked files to the RAM
func UnlockAllLockedFiles() error {
	lockedPaths := getLockedMappingKeys()

	if len(lockedPaths) == 0 {
		return nil
	}

	logger := logrus.NewEntry(logrus.StandardLogger())
	for _, lockedPath := range lockedPaths {
		if err := unlockFileFromRAM(lockedPath, logger); err != nil {
			return err
		}
	}
	return nil
}

// lockFileIntoRAM mmaps the whole file read-only and mlock()s it so pages are
// pinned in RAM. On success the mapping is kept in memory (and in lockedMappings).
//
// If mlock is not possible (insufficient RLIMIT_MEMLOCK, missing capability),
// the function returns an error. Caller may choose to fallback to fadvise.
//
// Important: the process must have enough RLIMIT_MEMLOCK (or CAP_IPC_LOCK).
// For very large files (e.g. 15GiB) you must increase the limit (see notes).
func lockFileIntoRAM(path string, logger *logrus.Entry) error {
	if logger == nil {
		logger = logrus.NewEntry(logrus.StandardLogger())
	}
	if path == "" {
		return fmt.Errorf("empty path")
	}
	abs, err := filepath.Abs(path)
	if err == nil {
		path = abs
	}

	// If already locked, return early.
	if _, ok := lockedMappings.Load(path); ok {
		logger.Debugf("LockFileIntoRAM: already locked: %s", path)
		return nil
	}

	f, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("open %s: %w", path, err)
	}
	defer f.Close()

	fi, err := f.Stat()
	if err != nil {
		return fmt.Errorf("stat %s: %w", path, err)
	}
	size := fi.Size()
	if size == 0 {
		return fmt.Errorf("cannot lock empty file %s", path)
	}

	// Check current RLIMIT_MEMLOCK
	var rlim unix.Rlimit
	if err := unix.Getrlimit(unix.RLIMIT_MEMLOCK, &rlim); err == nil {
		// rlim.Cur is the soft limit in bytes (0 = unlimited? usually RLIM_INFINITY)
		// If limit smaller than file size, warn and return error.
		if rlim.Cur != unix.RLIM_INFINITY && uint64(size) > rlim.Cur {
			return fmt.Errorf("rlimit MEMLOCK too small (%d) for file %s size %d; increase RLIMIT_MEMLOCK or run with CAP_IPC_LOCK", rlim.Cur, path, size)
		}
	} else {
		// If we can't read rlimit, continue and try mlock anyway.
		logger.Warnf("could not read RLIMIT_MEMLOCK: %v (will still try mlock)", err)
	}

	flags := unix.MAP_SHARED

	// MAP_POPULATE tells the kernel reads in all the necessary pages from the file to memory,
	// reducing page faults later when the memory is actually accessed.
	flags |= unix.MAP_POPULATE

	// Map the whole file (read-only, shared, populate)
	data, err := unix.Mmap(int(f.Fd()), 0, int(size), unix.PROT_READ, flags)
	if err != nil {
		return fmt.Errorf("mmap %s: %w", path, err)
	}

	// Try mlock the mapped region to lock a region of a process’s virtual memory into RAM, which prevents that memory from
	// being swapped out to disk by the operating system’s kernel memory manage
	if err := unix.Mlock(data); err != nil {
		// cleanup the mapping if mlock fails
		_ = unix.Munmap(data)
		return fmt.Errorf("mlock failed for %s: %w", path, err)
	}

	// Keep the mapping in the registry so it is not GC'd or unmapped.
	lockedMappings.Store(path, data)
	logger.Infof("Locked file into RAM: %s (size=%d bytes)", path, size)
	return nil
}

// unlockFileFromRAM will munlock and munmap the previously-locked mapping.
// It is safe to call even if the file was not locked.
func unlockFileFromRAM(path string, logger *logrus.Entry) error {
	if logger == nil {
		logger = logrus.NewEntry(logrus.StandardLogger())
	}
	if path == "" {
		return fmt.Errorf("empty path")
	}
	abs, err := filepath.Abs(path)
	if err == nil {
		path = abs
	}

	v, ok := lockedMappings.Load(path)
	if !ok {
		logger.Infof("UnlockFileFromRAM: not locked: %s", path)
		return nil
	}
	data := v.([]byte)

	if err := unix.Munlock(data); err != nil {
		logger.Warnf("munlock failed for %s: %v", path, err)
		// continue to munmap anyway
	}
	if err := unix.Munmap(data); err != nil {
		// In rare cases munmap can fail; surface it
		return fmt.Errorf("munmap failed for %s: %w", path, err)
	}
	lockedMappings.Delete(path)
	logger.Infof("Unlocked and unmapped file: %s", path)
	return nil
}

func getLockedMappingKeys() []string {
	var keys []string
	lockedMappings.Range(func(key, value any) bool {
		if s, ok := key.(string); ok {
			keys = append(keys, s)
		}
		return true // continue iteration
	})
	return keys
}
