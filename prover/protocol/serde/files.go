package serde

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/klauspost/compress/zstd"
	"github.com/sirupsen/logrus"
)

// LoadFromDisk loads a serialized asset from disk.
// It returns an io.Closer that MUST be closed by the caller once the
// asset is no longer needed (specifically for Mmap mode).
func LoadFromDisk(filePath string, assetPtr any, withCompression bool) (io.Closer, error) {
	start := time.Now()

	var (
		mfile *MappedFile
		data  []byte
		err   error
	)

	// We use a closer to ensure the memory stays valid as long as the caller needs it.
	var closer io.Closer = io.NopCloser(nil)

	if withCompression {
		// --- Path A: Compressed (Heap Memory) ---
		logrus.Infof("Loading compressed file %s...", filePath)

		compressedData, err := os.ReadFile(filePath)
		if err != nil {
			return nil, fmt.Errorf("failed to read file %s: %w", filePath, err)
		}

		decoder, err := zstd.NewReader(bytes.NewReader(compressedData))
		if err != nil {
			return nil, fmt.Errorf("failed to create zstd reader: %w", err)
		}
		defer decoder.Close()

		data, err = io.ReadAll(decoder)
		if err != nil {
			return nil, fmt.Errorf("failed to decompress %s: %w", filePath, err)
		}

		logrus.Infof("Decompressed %s in %s (Disk: %s -> Mem: %s)",
			filePath, time.Since(start), formatSize(len(compressedData)), formatSize(len(data)))

	} else {
		// --- Path B: Uncompressed (Memory Mapped) ---
		logrus.Infof("Mmapping file %s...", filePath)

		mfile, err = openMappedFile(filePath)
		if err != nil {
			return nil, fmt.Errorf("failed to mmap %s: %w", filePath, err)
		}
		data = mfile.data

		// The MappedFile struct implements io.Closer.
		// Returning it ensures the finalizer (munmap) won't run until the caller is done.
		closer = mfile
	}

	// 2. Deserialize (Overlay/Swizzle)
	tDeserialize := time.Now()
	if err := Deserialize(data, assetPtr); err != nil {
		closer.Close() // Clean up on error
		return nil, fmt.Errorf("failed to deserialize %s: %w", filePath, err)
	}

	logrus.Infof("Loaded & Deserialized %s in %s (Overlay took: %s) | Size: %s",
		filePath, time.Since(start), time.Since(tDeserialize), formatSize(len(data)))

	// While runtime.KeepAlive(mfile) prevents collection during this function,
	// returning the closer allows the caller to prevent collection during DeepCmp.
	return closer, nil
}

// StoreToDisk serializes the asset and writes it to the specified file.
// It uses an atomic write strategy (write temp -> rename) to ensure readers never see partial files.
func StoreToDisk(filePath string, asset any, withCompression bool) error {
	start := time.Now()

	// 1. Serialize
	data, err := Serialize(asset)
	if err != nil {
		return fmt.Errorf("failed to serialize asset: %w", err)
	}
	rawSize := len(data)
	tSerialize := time.Since(start)

	// 2. Compress (Optional)
	if withCompression {
		tCompStart := time.Now()
		var buf bytes.Buffer

		// Use default compression level
		encoder, err := zstd.NewWriter(&buf)
		if err != nil {
			return fmt.Errorf("failed to create zstd writer: %w", err)
		}

		if _, err := encoder.Write(data); err != nil {
			encoder.Close()
			return fmt.Errorf("compression write failed: %w", err)
		}

		if err := encoder.Close(); err != nil {
			return fmt.Errorf("compression close failed: %w", err)
		}

		data = buf.Bytes()
		logrus.Debugf("Compressed data: %s -> %s (Time: %s)",
			formatSize(rawSize), formatSize(len(data)), time.Since(tCompStart))
	}

	// 3. Atomic Write
	// Create temp file in the same directory to ensure we can Rename (atomic on POSIX)
	dir := filepath.Dir(filePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create dir %s: %w", dir, err)
	}

	tmpFile, err := os.CreateTemp(dir, filepath.Base(filePath)+".tmp.*")
	if err != nil {
		return fmt.Errorf("failed to create temp file: %w", err)
	}
	tmpName := tmpFile.Name()

	// Cleanup if something goes wrong before the rename
	defer func() {
		_ = tmpFile.Close()
		if _, err := os.Stat(tmpName); err == nil {
			_ = os.Remove(tmpName)
		}
	}()

	// Write data
	if _, err := tmpFile.Write(data); err != nil {
		return fmt.Errorf("failed to write to temp file: %w", err)
	}

	// Fsync to ensure durability
	if err := tmpFile.Sync(); err != nil {
		return fmt.Errorf("failed to fsync temp file: %w", err)
	}

	// Close explicitly before renaming
	if err := tmpFile.Close(); err != nil {
		return fmt.Errorf("failed to close temp file: %w", err)
	}

	// Atomic Rename
	if err := os.Rename(tmpName, filePath); err != nil {
		return fmt.Errorf("failed to rename temp file to %s: %w", filePath, err)
	}

	logrus.Infof("Saved %s in %s (Ser: %s, Total Size: %s)",
		filePath, time.Since(start), tSerialize, formatSize(len(data)))

	return nil
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
