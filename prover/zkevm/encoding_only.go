package zkevm

import (
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/consensys/linea-monorepo/prover/config"
	"github.com/consensys/linea-monorepo/prover/protocol/compiler/logderivativesum"
	"github.com/consensys/linea-monorepo/prover/protocol/compiler/mimc"
	"github.com/consensys/linea-monorepo/prover/protocol/compiler/specialqueries"
	"github.com/consensys/linea-monorepo/prover/protocol/serialization"
	"github.com/consensys/linea-monorepo/prover/protocol/wizard"
	"github.com/sirupsen/logrus"
)

var (
	encodeOnlyZkevm            *ZkEvm
	onceEncodeOnlyZkevm        = sync.Once{}
	encodeOnlyCompilationSuite = CompilationSuite{
		mimc.CompileMiMC,
		specialqueries.RangeProof,
		logderivativesum.CompileLookups,
	}
)

func EncodeOnlyZkEvm(tl *config.TracesLimits) *ZkEvm {
	onceEncodeOnlyZkevm.Do(func() {
		encodeOnlyZkevm = FullZKEVMWithSuite(tl, encodeOnlyCompilationSuite)
	})

	return encodeOnlyZkevm
}

func (z *ZkEvm) AssignAndEncodeInChunks(filepath string, input *Witness, numChunks int) {

	// Start encoding and measure time
	encodingStart := time.Now()
	run := wizard.RunProverUntilRound(z.WizardIOP, z.prove(input), 1)
	firstRoundOnlyDuration := time.Since(encodingStart).Seconds()
	logrus.Infof("ProverOnlyFirstRound complete, took %.2f seconds", firstRoundOnlyDuration)

	// Start serialization
	serializationStart := time.Now()
	serializedChunks := serialization.SerializeAssignment(run.Columns, numChunks)
	serializationDuration := time.Since(serializationStart).Seconds()
	logrus.Infof("CBOR serialization complete, took %.2f seconds", serializationDuration)

	// Calculate total size of serialized data
	totalSerializedSize := 0
	for _, chunk := range serializedChunks {
		totalSerializedSize += len(chunk)
	}
	encodingDuration := time.Since(encodingStart).Seconds()
	logrus.Infof("Encoding (ProverOnlyFirstRound + Serialization) complete, total serialized size: %d bytes, took %.2f seconds", totalSerializedSize, encodingDuration)

	// Start compression and measure time
	compressionStart := time.Now()
	compressedSerializedChunks := serialization.CompressChunks(serializedChunks)
	compressionDuration := time.Since(compressionStart).Seconds()

	// Calculate total size of compressed data
	totalCompressedSize := 0
	for _, chunk := range compressedSerializedChunks {
		totalCompressedSize += len(chunk)
	}
	logrus.Infof("Compression complete, total compressed size: %d bytes, took %.2f seconds", totalCompressedSize, compressionDuration)

	// Start writing to files
	writingStart := time.Now()
	var wg sync.WaitGroup
	for i := 0; i < numChunks; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			chunk := compressedSerializedChunks[i]
			chunkPath := fmt.Sprintf("%s_chunk_%d", filepath, i)

			// Measure writing time for each chunk
			writeStart := time.Now()
			f, err := os.Create(chunkPath)
			if err != nil {
				logrus.Errorf("[%v] error creating file %s: %v", time.Now(), chunkPath, err)
				return
			}
			defer f.Close()

			_, err = f.Write(chunk)
			if err != nil {
				logrus.Errorf("[%v] error writing to file %s: %v", time.Now(), chunkPath, err)
				return
			}
			writeDuration := time.Since(writeStart).Seconds()
			logrus.Infof("Completed writing chunk %d to %s, took %.2f seconds", i, chunkPath, writeDuration)
		}(i)
	}
	wg.Wait()
	writingDuration := time.Since(writingStart).Seconds() // Total writing time
	logrus.Infof("Writing complete, total compressed size: %d bytes, took %.2f seconds", totalCompressedSize, writingDuration)

	// Total summary
	totalDuration := encodingDuration + compressionDuration + writingDuration
	logrus.Infof("Total serialized size %d bytes, total compressed size %d bytes, took %.2f sec total (encoding + compression + writing)", totalSerializedSize, totalCompressedSize, totalDuration)
	logrus.Infof("Compression Ratio: %.2f", float64(totalSerializedSize)/float64(totalCompressedSize))
}
