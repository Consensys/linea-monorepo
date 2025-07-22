package serialization

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"time"

	"github.com/consensys/linea-monorepo/prover/backend/files"
	"github.com/consensys/linea-monorepo/prover/utils/profiling"
	"github.com/pierrec/lz4/v4"
	"github.com/sirupsen/logrus"
	"golang.org/x/sync/semaphore"
)

var (
	// numConcurrentDiskWrite governs the number of concurrent disk writes
	numConcurrentDiskWrite = 1
	// concurrentSubProverSemaphore is a semaphore that controls the number of
	// concurrent disk writes.
	diskWriteSemaphore = semaphore.NewWeighted(int64(numConcurrentDiskWrite))
	// concurrentDiskRead governs the number of concurrent disk reads
	numConcurrentDiskRead = 1
	// diskReadSemaphore is a semaphore that controls the number of concurrent
	// disk reads.
	diskReadSemaphore = semaphore.NewWeighted(int64(numConcurrentDiskRead))
)

// LoadFromDisk loads a serialized asset from disk
func LoadFromDisk(filePath string, assetPtr any, withCompression bool) error {

	f := files.MustRead(filePath)
	defer f.Close()

	var (
		buf       []byte
		desErr    error
		readErr   error
		decompErr error
		tDecomp   time.Duration
	)

	diskReadSemaphore.Acquire(context.Background(), 1)

	logrus.Infof("Reading file %v\n", filePath)

	tRead := profiling.TimeIt(func() {
		buf, readErr = io.ReadAll(f)
	})

	diskReadSemaphore.Release(int64(1))

	if readErr != nil {
		return fmt.Errorf("could not read file %s: %w", filePath, readErr)
	}

	sizeDisk := len(buf)
	sizeUncompressed := sizeDisk
	readingSpeedBytesperSec := float64(sizeDisk) / tRead.Seconds()

	if withCompression {

		tDecomp = profiling.TimeIt(func() {
			r := lz4.NewReader(bytes.NewReader(buf))
			buf, decompErr = io.ReadAll(r)
			if decompErr != nil {
				return
			}
		})

		if decompErr != nil {
			return fmt.Errorf("could not decompress file %s: %w", filePath, decompErr)
		}

		sizeUncompressed = len(buf)
	}

	tDes := profiling.TimeIt(func() {
		desErr = Deserialize(buf, assetPtr)
	})

	if desErr != nil {
		return fmt.Errorf("could not deserialize %s: %w", filePath, desErr)
	}

	logrus.Infof("Read %s in %s, deserialized in %s, decompressed in %s, size-disk: %dB, size-uncompressed: %v, reading-speed: %vB/sec", filePath, tRead, tDes, tDecomp, sizeDisk, sizeUncompressed, readingSpeedBytesperSec)

	return nil
}

// StoreToDisk writes the provided assets to disk using the [Serialize] function.
func StoreToDisk(filePath string, asset any, withCompression bool) error {

	f := files.MustOverwrite(filePath)
	defer f.Close()

	var (
		buf     []byte
		serr    error
		werr    error
		compErr error
		tComp   time.Duration
	)

	tSer := profiling.TimeIt(func() {
		buf, serr = Serialize(asset)
	})

	if serr != nil {
		return fmt.Errorf("could not serialize %s: %w", filePath, serr)
	}

	if withCompression {

		tComp = profiling.TimeIt(func() {

			var (
				b = bytes.NewBuffer(nil)
				w = lz4.NewWriter(b)
			)

			if _, compErr = w.Write(buf); compErr != nil {
				return
			}

			if compErr = w.Flush(); compErr != nil {
				return
			}

			buf = b.Bytes()
		})

		if compErr != nil {
			return fmt.Errorf("could not compress file %s: %w", filePath, compErr)
		}
	}

	diskWriteSemaphore.Acquire(context.Background(), 1)

	tW := profiling.TimeIt(func() {
		_, werr = f.Write(buf)
	})

	diskWriteSemaphore.Release(1)

	if werr != nil {
		return fmt.Errorf("could not write to file %s: %w", filePath, werr)
	}

	logrus.Infof("Wrote %s in %s, serialized in %s, compressed in %s, size: %dB", filePath, tW, tSer, tComp, len(buf))

	return nil
}
