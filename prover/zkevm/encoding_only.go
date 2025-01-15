package zkevm

import (
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/consensys/linea-monorepo/prover/config"
	"github.com/consensys/linea-monorepo/prover/protocol/compiler/lookup"
	"github.com/consensys/linea-monorepo/prover/protocol/compiler/mimc"
	"github.com/consensys/linea-monorepo/prover/protocol/compiler/specialqueries"
	"github.com/consensys/linea-monorepo/prover/protocol/serialization"
	"github.com/consensys/linea-monorepo/prover/protocol/wizard"
	"github.com/sirupsen/logrus"
)

var (
	encodeOnlyZkevm            *ZkEvm
	onceEncodeOnlyZkevm        = sync.Once{}
	encodeOnlyCompilationSuite = compilationSuite{
		mimc.CompileMiMC,
		specialqueries.RangeProof,
		lookup.CompileLogDerivative,
	}
)

func EncodeOnlyZkEvm(tl *config.TracesLimits) *ZkEvm {
	onceEncodeOnlyZkevm.Do(func() {
		encodeOnlyZkevm = fullZKEVMWithSuite(tl, encodeOnlyCompilationSuite)
	})

	return encodeOnlyZkevm
}


func (z *ZkEvm) AssignAndEncodeInChunks(filepath string, input *Witness, numChunks int) {
	// Start serialization and measure time
	encodingStart := time.Now()
	run := wizard.ProverOnlyFirstRound(z.WizardIOP, z.prove(input))
	serializedChunks := serialization.SerializeAssignment(run.Columns, numChunks)
	encodingDuration := time.Since(encodingStart).Seconds()

	// Calculate total size of serialized data
	totalSerializedSize := 0
	for _, chunk := range serializedChunks {
		totalSerializedSize += len(chunk)
	}
	logrus.Infof("[%v] encoding complete, total serialized size: %d bytes, took %.2f seconds", time.Now(), totalSerializedSize, encodingDuration)

	// Start compression and measure time
	compressionStart := time.Now()
	compressedSerializedChunks := serialization.CompressChunks(serializedChunks)
	compressionDuration := time.Since(compressionStart).Seconds()

	// Calculate total size of compressed data
	totalCompressedSize := 0
	for _, chunk := range compressedSerializedChunks {
		totalCompressedSize += len(chunk)
	}
	logrus.Infof("[%v] compression complete, total compressed size: %d bytes, took %.2f seconds", time.Now(), totalCompressedSize, compressionDuration)

	// Start writing process timing
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

	// Total process summary
	totalDuration := encodingDuration + compressionDuration + writingDuration
	logrus.Infof("[%v] blob total serialized size %d bytes, total compressed size %d bytes, took %.2f sec total (encoding + compression + write)", time.Now(), totalSerializedSize, totalCompressedSize, totalDuration)
	logrus.Infof("[%v] total encoding time: %.2f seconds, total compression time: %.2f seconds, total writing time: %.2f seconds", time.Now(), encodingDuration, compressionDuration, writingDuration)
}