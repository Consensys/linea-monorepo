// File: prover/protocol/serde/files.go
package serde

import (
	"fmt"
	"time"

	"github.com/sirupsen/logrus"
)

// LoadFromDisk replaces the legacy read-all approach with Memory Mapping.
// It is vastly faster for startup.
func LoadFromDisk(filePath string, assetPtr any) error {
	start := time.Now()

	// 1. Mmap the file
	// This takes microseconds, regardless of file size (e.g. 100GB).
	mfile, err := OpenMappedFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to mmap %s: %w", filePath, err)
	}

	// Note: We cannot close mfile immediately because assetPtr will point into its data.
	// The mfile acts as the heap. It will be closed by GC Finalizer or manual management
	// if we introduce a Lifecycle manager later.

	// 2. Deserialize (Overlay/Swizzle)
	// This performs pointer swizzling and slice header construction.
	// It creates a "View" of the data.
	if err := Deserialize(mfile.data, assetPtr); err != nil {
		// If deserialization fails, we must unmap to avoid leaks
		mfile.Close()
		return fmt.Errorf("failed to deserialize %s: %w", filePath, err)
	}

	logrus.Infof("Zero-Copy Loaded %s in %s (Size: %d bytes)",
		filePath, time.Since(start), len(mfile.data))

	return nil
}
