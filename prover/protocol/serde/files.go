package serde

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/consensys/linea-monorepo/prover/utils/profiling"
	"github.com/klauspost/compress/zstd"
	"github.com/sirupsen/logrus"
	"golang.org/x/sync/semaphore"
)

var (

	// numConcurrentDiskWrite governs the number of concurrent disk writes
	numConcurrentDiskWrite = 1
	// diskWriteSemaphore limits concurrent writes to prevent IO congestion.
	// Even at 1, it ensures that massive assets (like Module 11) don't compete
	// for disk bandwidth, which is critical for mechanical or shared cloud drives.
	diskWriteSemaphore = semaphore.NewWeighted(int64(numConcurrentDiskWrite))

	// concurrentDiskRead governs the number of concurrent disk reads
	numConcurrentDiskRead = 1
	// diskReadSemaphore ensures we don't saturate the IO bus during heavy asset loading.
	diskReadSemaphore = semaphore.NewWeighted(int64(numConcurrentDiskRead))
)

// StoreToDisk serializes the asset and writes it atomically to disk.
func StoreToDisk(filePath string, asset any, withCompression bool) error {
	var (
		data     []byte
		serErr   error
		writeErr error
		tSer     time.Duration
		tComp    time.Duration
	)

	// --- 1. Serialization (Linearization) ---
	tSer = profiling.TimeIt(func() {
		data, serErr = Serialize(asset)
	})
	if serErr != nil {
		return fmt.Errorf("serialization failed: %w", serErr)
	}
	rawSize := len(data)

	// --- 2. Optional Compression (Zstd) ---
	if withCompression {
		tComp = profiling.TimeIt(func() {
			var buf bytes.Buffer
			encoder, _ := zstd.NewWriter(&buf)
			_, _ = encoder.Write(data)
			_ = encoder.Close()
			data = buf.Bytes()
		})
		logrus.Debugf("Compression: %s -> %s", formatSize(rawSize), formatSize(len(data)))
	}

	// --- 3. Atomic Write Strategy ---
	// We write to a .tmp file and rename it. This ensures that if the prover crashes
	// during a write, we don't leave a corrupted asset on disk.
	dir := filepath.Dir(filePath)
	_ = os.MkdirAll(dir, 0755)

	tmpFile, err := os.CreateTemp(dir, filepath.Base(filePath)+".tmp.*")
	if err != nil {
		return fmt.Errorf("failed to create temp file: %w", err)
	}
	tmpName := tmpFile.Name()

	// Cleanup closure: delete the temp file if we don't finish the Rename.
	defer func() {
		_ = tmpFile.Close()
		if tmpName != "" {
			_ = os.Remove(tmpName)
		}
	}()

	_ = diskWriteSemaphore.Acquire(context.Background(), 1)
	tWrite := profiling.TimeIt(func() {
		_, writeErr = tmpFile.Write(data)
		if writeErr == nil {
			writeErr = tmpFile.Sync() // Ensure bits are actually on the platter.
		}
	})
	diskWriteSemaphore.Release(1)

	if writeErr != nil {
		return fmt.Errorf("write/sync failed: %w", writeErr)
	}

	_ = tmpFile.Close()

	// Atomic Rename: On POSIX, this replaces the file instantly.
	if err := os.Rename(tmpName, filePath); err != nil {
		return fmt.Errorf("rename failed: %w", err)
	}
	tmpName = "" // Success: disable the defer removal.

	// Optional: Sync the directory itself to ensure the directory entry is durable.
	if dfd, err := os.Open(dir); err == nil {
		_ = dfd.Sync()
		_ = dfd.Close()
	}

	logrus.Infof("Saved %s | Total: %s (Ser: %s, Comp: %s, DiskIO: %s)",
		filePath, formatSize(len(data)), tSer, tComp, tWrite)

	return nil
}

// LoadFromDisk loads a serialized asset from disk.
// Returns an io.Closer that MUST be closed by the caller to release the Mmap region.
// NOTE: The requirement to return an io.Closer is driven entirely by the Uncompressed (Mmap) Path.
// In that path, the memory is borrowed from the OS kernel, and we must have a way to tell the OS to unmap it (Munmap)
// when we are done. For compressed path, we use a default io.NopCloser.
func LoadFromDisk(filePath string, assetPtr any, withCompression bool) (io.Closer, error) {
	var (
		mfile    *MappedFile
		data     []byte
		readErr  error
		deserErr error
		tDecomp  time.Duration
	)

	// We use a NopCloser as the default for the "Compressed" path where data is in Heap.
	// When using Mmap (the uncompressed path), the memory isn't actually "allocated"—it's borrowed from the OS
	// The closer that we set for the uncompressed path is essentially a handle to that "loan."
	// Calling closer.Close() is the signal to the OS that you are done with the memory.
	// See Ln.177 we set closer to mfile for uncompressed path
	var closer io.Closer = io.NopCloser(nil)

	// --- 1. Data Acquisition Phase ---
	if withCompression {
		// Path A: Compressed (Must load into Heap memory)
		logrus.Infof("Loading compressed asset %s...", filePath)

		_ = diskReadSemaphore.Acquire(context.Background(), 1)
		compressedData, err := os.ReadFile(filePath)
		diskReadSemaphore.Release(1)

		if err != nil {
			return nil, fmt.Errorf("failed to read file: %w", err)
		}

		tDecomp = profiling.TimeIt(func() {
			// Zstd is significantly faster than LZ4 for high-compression ratios in Prover assets.
			decoder, _ := zstd.NewReader(bytes.NewReader(compressedData))
			defer decoder.Close()
			data, readErr = io.ReadAll(decoder)
		})

		if readErr != nil {
			return nil, fmt.Errorf("decompression failed: %w", readErr)
		}

	} else {
		// Path B: Uncompressed (Zero-Copy Mmap)
		logrus.Infof("Memory-mapping asset %s...", filePath)

		_ = diskReadSemaphore.Acquire(context.Background(), 1)
		tMmap := profiling.TimeIt(func() {
			mfile, readErr = openMappedFile(filePath)
		})
		diskReadSemaphore.Release(1)

		if readErr != nil {
			return nil, fmt.Errorf("mmap failed: %w", readErr)
		}

		data = mfile.Data()
		closer = mfile // The caller must call Close() to trigger Munmap.
		logrus.Debugf("Mmap established in %s", tMmap)
	}

	// --- 2. Deserialization (Overlay/Swizzle) Phase ---
	// This is where the magic happens: we map Go headers directly onto 'data'.
	tDeser := profiling.TimeIt(func() {
		deserErr = Deserialize(data, assetPtr)
	})

	if deserErr != nil {
		_ = closer.Close() // Safety: Don't leak the mmap if swizzling fails.
		return nil, fmt.Errorf("deserialization failed: %w", deserErr)
	}

	logrus.Infof("Loaded %s [Size: %s] | Deser: %s | Decomp: %s",
		filePath, formatSize(len(data)), tDeser, tDecomp)

	return closer, nil
}

func formatSize(bytes int) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}
