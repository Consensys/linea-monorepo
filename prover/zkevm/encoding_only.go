package zkevm

import (
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/consensys/linea-monorepo/prover/backend/files"
	"github.com/consensys/linea-monorepo/prover/config"
	"github.com/consensys/linea-monorepo/prover/protocol/compiler/lookup"
	"github.com/consensys/linea-monorepo/prover/protocol/compiler/mimc"
	"github.com/consensys/linea-monorepo/prover/protocol/compiler/specialqueries"
	"github.com/consensys/linea-monorepo/prover/protocol/serialization"
	"github.com/consensys/linea-monorepo/prover/protocol/wizard"
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

func (z *ZkEvm) AssignAndEncodeInFile(filepath string, input *Witness) {
	// Start encoding and measure time
	encodingStart := time.Now()
	run := wizard.ProverOnlyFirstRound(z.WizardIOP, z.prove(input))
	b := serialization.SerializeAssignment(run.Columns)
	encodingDuration := time.Since(encodingStart).Seconds()
	fmt.Printf("[%v] encoding complete, total size: %d bytes, took %.2f seconds\n", time.Now(), len(b), encodingDuration)

	// Start writing and measure time
	writingStart := time.Now()
	f := files.MustOverwrite(filepath)
	f.Write(b)
	f.Close()
	writingDuration := time.Since(writingStart).Seconds()
	fmt.Printf("[%v] writing complete, took %.2f seconds\n", time.Now(), writingDuration)

	// Summary of total time
	totalDuration := encodingDuration + writingDuration
	fmt.Printf("[%v] blob total size %v bytes, took %.2f sec total (encode + write)\n", time.Now(), len(b), totalDuration)
}

func (z *ZkEvm) AssignAndEncodeInChunks(filepath string, input *Witness, numChunks int) {
	// Start encoding and measure time
	encodingStart := time.Now()
	run := wizard.ProverOnlyFirstRound(z.WizardIOP, z.prove(input))
	b := serialization.SerializeAssignment(run.Columns)
	// b := serialization.SerializeAssignmentWithoutCompression(run.Columns)
	encodingDuration := time.Since(encodingStart).Seconds()
	fmt.Printf("[%v] encoding complete, total size: %d bytes, took %.2f seconds\n", time.Now(), len(b), encodingDuration)

	// Determine the size of each chunk
	chunkSize := (len(b) + numChunks - 1) / numChunks // Round up to ensure all data is included
	fmt.Printf("[%v] calculated chunk size: %d bytes for %d chunks\n", time.Now(), chunkSize, numChunks)

	// Start writing process timing
	writingStart := time.Now()
	var wg sync.WaitGroup
	for i := 0; i < numChunks; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			// Calculate chunk boundaries
			start := i * chunkSize
			end := start + chunkSize
			if end > len(b) {
				end = len(b) // Adjust for the last chunk
			}
			chunk := b[start:end]
			chunkPath := fmt.Sprintf("%s_chunk_%d", filepath, i)

			// Measure writing time for each chunk
			writeStart := time.Now()
			f, err := os.Create(chunkPath)
			if err != nil {
				fmt.Printf("[%v] error creating file %s: %v\n", time.Now(), chunkPath, err)
				return
			}
			defer f.Close()

			_, err = f.Write(chunk)
			if err != nil {
				fmt.Printf("[%v] error writing to file %s: %v\n", time.Now(), chunkPath, err)
				return
			}
			writeDuration := time.Since(writeStart).Seconds()
			fmt.Printf("[%v] completed writing chunk %d to %s, took %.2f seconds\n", time.Now(), i, chunkPath, writeDuration)
		}(i)
	}
	wg.Wait()
	writingDuration := time.Since(writingStart).Seconds() // Total writing time

	// Total encoding and writing summary
	totalDuration := encodingDuration + writingDuration
	fmt.Printf("[%v] blob total size %v bytes, took %.2f sec total (encode + write)\n", time.Now(), len(b), totalDuration)
	fmt.Printf("[%v] total encoding time: %.2f seconds, total writing time: %.2f seconds\n", time.Now(), encodingDuration, writingDuration)
}