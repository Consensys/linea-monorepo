package controller

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/exp/slices"
)

func TestRunCommand(t *testing.T) {

	// Bootstrap the test directory
	confM, confL := setupFsTest(t)

	var (
		eFrom   string = confM.Execution.DirFrom(0)
		cFrom   string = confM.BlobDecompression.DirFrom(0)
		aFrom   string = confM.Aggregation.DirFrom(0)
		exit0   int    = 0
		exit2   int    = 2
		exit10  int    = 10
		exit12  int    = 12
		exit77  int    = 77
		exit137 int    = 137
	)

	// Populate the filesystem with job files

	// Execution
	createTestInputFile(eFrom, 0, 1, execJob, exit0)
	createTestInputFile(eFrom, 1, 2, execJob, exit12, forLarge)
	createTestInputFile(eFrom, 2, 3, execJob, exit77)
	createTestInputFile(eFrom, 3, 4, execJob, exit77, forLarge)
	createTestInputFile(eFrom, 4, 5, execJob, exit137)
	createTestInputFile(eFrom, 5, 6, execJob, exit137, forLarge)
	createTestInputFile(eFrom, 6, 7, execJob, exit2)
	createTestInputFile(eFrom, 7, 8, execJob, exit2)
	createTestInputFile(eFrom, 8, 9, execJob, exit10)
	createTestInputFile(eFrom, 9, 10, execJob, exit12)

	// Compression
	createTestInputFile(cFrom, 0, 2, compressionJob, exit0)
	createTestInputFile(cFrom, 2, 4, compressionJob, exit2)
	createTestInputFile(cFrom, 4, 6, compressionJob, exit77)
	createTestInputFile(cFrom, 6, 8, compressionJob, exit137)

	// Aggregation
	createTestInputFile(aFrom, 0, 2, aggregationJob, exit0)
	createTestInputFile(aFrom, 2, 4, aggregationJob, exit2)
	createTestInputFile(aFrom, 4, 6, aggregationJob, exit77)
	createTestInputFile(aFrom, 6, 8, aggregationJob, exit137)

	ctxM, stopM := context.WithCancel(context.Background())
	ctxL, stopL := context.WithCancel(context.Background())

	go runController(ctxM, confM)
	go runController(ctxL, confL)

	// Give one sec, for the test to complete
	<-time.After(4 * time.Second)

	// Shutdown the two concurrent controllers
	stopM()
	stopL()

	// After this test the test directory should have the following structure
	expectedStructure := []struct {
		Path    string
		Entries []string
	}{
		{
			Path:    confM.Execution.DirFrom(0),
			Entries: []string{}, // all files should be processed
		},
		{
			Path: confM.Execution.DirDone(0),
			Entries: []string{
				"0-1-etv0.1.2-stv1.2.3-getZkProof.json.success",
				"1-2-etv0.1.2-stv1.2.3-getZkProof.json.large.success",
				"2-3-etv0.1.2-stv1.2.3-getZkProof.json.failure.code_67",
				"3-4-etv0.1.2-stv1.2.3-getZkProof.json.large.failure.code_65",
				"4-5-etv0.1.2-stv1.2.3-getZkProof.json.large.failure.code_125",
				"5-6-etv0.1.2-stv1.2.3-getZkProof.json.large.failure.code_125",
				"6-7-etv0.1.2-stv1.2.3-getZkProof.json.failure.code_2",
				"7-8-etv0.1.2-stv1.2.3-getZkProof.json.failure.code_2",
				"8-9-etv0.1.2-stv1.2.3-getZkProof.json.success",
				"9-10-etv0.1.2-stv1.2.3-getZkProof.json.large.success",
			},
		},
		{
			Path: confM.Execution.DirTo(0),
			Entries: []string{
				"0-1-getZkProof.json",
				"1-2-getZkProof.json",
				"8-9-getZkProof.json",
				"9-10-getZkProof.json",
			},
		},
		{
			Path:    confM.BlobDecompression.DirFrom(0),
			Entries: []string{},
		},
		{
			Path: confM.BlobDecompression.DirDone(0),
			Entries: []string{
				"0-2-bcv0.1.2-ccv0.1.2-getZkBlobCompressionProof.json.success",
				"2-4-bcv0.1.2-ccv0.1.2-getZkBlobCompressionProof.json.failure.code_2",
				"4-6-bcv0.1.2-ccv0.1.2-getZkBlobCompressionProof.json.failure.code_77",
				"6-8-bcv0.1.2-ccv0.1.2-getZkBlobCompressionProof.json.failure.code_137",
			},
		},
		{
			Path: confM.BlobDecompression.DirTo(0),
			Entries: []string{
				"0-2-getZkBlobCompressionProof.json",
			},
		},
		{
			Path:    confM.Aggregation.DirFrom(0),
			Entries: []string{},
		},
		{
			Path: confM.Aggregation.DirDone(0),
			Entries: []string{
				"0-2-deadbeef57-getZkAggregatedProof.json.success",
				"2-4-deadbeef57-getZkAggregatedProof.json.failure.code_2",
				"4-6-deadbeef57-getZkAggregatedProof.json.failure.code_77",
				"6-8-deadbeef57-getZkAggregatedProof.json.failure.code_137",
			},
		},
		{
			Path:    confM.Aggregation.DirTo(0),
			Entries: []string{"0-2-deadbeef57-getZkAggregatedProof.json"},
		},
	}

	for _, dirVal := range expectedStructure {
		dir, err := os.Open(dirVal.Path)
		require.NoErrorf(t, err, "dir %v", dirVal.Path)
		filesFound, err := dir.Readdirnames(-1)
		require.NoErrorf(t, err, "dir %v", dirVal.Path)
		slices.Sort(filesFound)
		assert.Equalf(t, dirVal.Entries, filesFound, "dir %v", dirVal.Path)
	}

}
