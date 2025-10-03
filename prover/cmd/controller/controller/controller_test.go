package controller

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"syscall"
	"testing"
	"text/template"
	"time"

	"github.com/consensys/linea-monorepo/prover/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/exp/slices"
)

func TestRunCommand(t *testing.T) {

	// Bootstrap the test directory
	confM, confL := setupFsTest(t)

	var (
		eFrom   string = confM.Execution.DirFrom()
		cFrom   string = confM.BlobDecompression.DirFrom()
		aFrom   string = confM.Aggregation.DirFrom()
		exit0   int    = 0
		exit2   int    = 2
		exit10  int    = 10
		exit12  int    = 12
		exit77  int    = 77
		exit137 int    = 137
	)

	// Populate the filesystem with job files

	// execution
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

	// compression
	createTestInputFile(cFrom, 0, 2, compressionJob, exit0)
	createTestInputFile(cFrom, 2, 4, compressionJob, exit2)
	createTestInputFile(cFrom, 4, 6, compressionJob, exit77)
	createTestInputFile(cFrom, 6, 8, compressionJob, exit137)

	// aggregation
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
			Path:    confM.Execution.DirFrom(),
			Entries: []string{}, // all files should be processed
		},
		{
			Path: confM.Execution.DirDone(),
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
			Path: confM.Execution.DirTo(),
			Entries: []string{
				"0-1-getZkProof.json",
				"1-2-getZkProof.json",
				"8-9-getZkProof.json",
				"9-10-getZkProof.json",
			},
		},
		{
			Path:    confM.BlobDecompression.DirFrom(),
			Entries: []string{},
		},
		{
			Path: confM.BlobDecompression.DirDone(),
			Entries: []string{
				"0-2-bcv0.1.2-ccv0.1.2-getZkBlobCompressionProof.json.success",
				"2-4-bcv0.1.2-ccv0.1.2-getZkBlobCompressionProof.json.failure.code_2",
				"4-6-bcv0.1.2-ccv0.1.2-getZkBlobCompressionProof.json.failure.code_77",
				"6-8-bcv0.1.2-ccv0.1.2-getZkBlobCompressionProof.json.failure.code_137",
			},
		},
		{
			Path: confM.BlobDecompression.DirTo(),
			Entries: []string{
				"0-2-getZkBlobCompressionProof.json",
			},
		},
		{
			Path:    confM.Aggregation.DirFrom(),
			Entries: []string{},
		},
		{
			Path: confM.Aggregation.DirDone(),
			Entries: []string{
				"0-2-deadbeef57-getZkAggregatedProof.json.success",
				"2-4-deadbeef57-getZkAggregatedProof.json.failure.code_2",
				"4-6-deadbeef57-getZkAggregatedProof.json.failure.code_77",
				"6-8-deadbeef57-getZkAggregatedProof.json.failure.code_137",
			},
		},
		{
			Path:    confM.Aggregation.DirTo(),
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

func TestSpotInstanceMode(t *testing.T) {

	t.Skipf("this breaks the CI pipeline")

	var (
		cfg    = setupFsTestSpotInstance(t)
		nbTest = 5
	)

	for i := 0; i < nbTest; i++ {
		// Create the input file
		createTestInputFile(cfg.Execution.DirFrom(), i, i, execJob, 0)
	}

	// This bit of code finds the current process with the goal of self-SIGTERMING
	// a few seconds after we start the controller.
	p, err := os.FindProcess(os.Getpid())
	if err != nil {
		panic("could not find the current process")
	}

	go runController(context.Background(), cfg)

	// Wait for the controller to process 2 jobs and be in the middle of the
	// 3rd job.
	time.Sleep(5 * time.Second)

	if err := p.Signal(os.Signal(syscall.SIGTERM)); err != nil {
		panic("panic could not self-send a SIGTERM")
	}

	var (
		contentFrom, eFrom = lsname(cfg.Execution.DirFrom())
		contentDone, eDone = lsname(cfg.Execution.DirDone())
	)

	if eFrom != nil || eDone != nil {
		t.Fatalf("could not list the directories: %v %v", eFrom, eDone)
	}

	assert.Len(t, contentFrom, 3)
	assert.Len(t, contentDone, 2)
}

func setupFsTestSpotInstance(t *testing.T) (cfg *config.Config) {

	// Testdir is going to contain the whole test directory
	testDir := t.TempDir()

	cfg = &config.Config{
		Version: "0.2.4",

		Controller: config.Controller{
			SpotInstanceMode: true,
			EnableExecution:  true,
			LocalID:          "test-prover-id",
			Prometheus:       config.Prometheus{Enabled: false},
			RetryDelays:      []int{0, 1},
			WorkerCmdTmpl:    template.Must(template.New("test-cmd").Parse("sleep 2")),
		},

		Execution: config.Execution{
			WithRequestDir: config.WithRequestDir{
				RequestsRootDir: path.Join(testDir, "execution"),
			},
		},
	}

	// Initialize the dirs (and give them all permissions). They will be
	// wiped out after the test anyway.
	permCode := fs.FileMode(0777)
	err := errors.Join(
		os.MkdirAll(cfg.Execution.DirFrom(), permCode),
		os.MkdirAll(cfg.Execution.DirTo(), permCode),
		os.MkdirAll(cfg.Execution.DirDone(), permCode),
	)

	if err != nil {
		t.Fatal(err)
	}

	return cfg
}

// --- Replace TestRunDistController with this version ---
func TestRunDistController(t *testing.T) {
	conf := setupFsTestLimitless(t)

	// --- Bootstrap jobs (must live under Execution.RequestsRootDir/requests) ---
	bsSucc := createLimitlessInputFile(
		filepath.Join(conf.Execution.RequestsRootDir), // <- IMPORTANT: requests-root
		"bootstrap",
		101, 102, 0, "",
	)
	bsFail := createLimitlessInputFile(
		filepath.Join(conf.Execution.RequestsRootDir),
		"bootstrap",
		103, 104, 77, "",
	)

	// --- Conglomeration jobs (metadata requests) ---
	cgSucc := createLimitlessInputFile(
		filepath.Join(conf.ExecutionLimitless.MetadataDir), "conglomeration",
		201, 202, 0, "",
	)
	cgFail := createLimitlessInputFile(
		filepath.Join(conf.ExecutionLimitless.MetadataDir), "conglomeration",
		203, 204, 2, "",
	)

	// --- GL and LPP jobs for all modules ---
	glSucc, glFail := []string{}, []string{}
	lppSucc, lppFail := []string{}, []string{}

	for _, mod := range config.ALL_MODULES {
		glSucc = append(glSucc, createLimitlessInputFile(
			filepath.Join(conf.ExecutionLimitless.WitnessDir, "GL"),
			"gl", 300, 301, 0, mod,
		))
		glFail = append(glFail, createLimitlessInputFile(
			filepath.Join(conf.ExecutionLimitless.WitnessDir, "GL"),
			"gl", 302, 303, 137, mod,
		))
		lppSucc = append(lppSucc, createLimitlessInputFile(
			filepath.Join(conf.ExecutionLimitless.WitnessDir, "LPP"),
			"lpp", 400, 401, 0, mod,
		))
		lppFail = append(lppFail, createLimitlessInputFile(
			filepath.Join(conf.ExecutionLimitless.WitnessDir, "LPP"),
			"lpp", 402, 403, 137, mod,
		))
	}

	// Run controller
	ctx, stop := context.WithCancel(context.Background())
	go runController(ctx, conf)
	time.Sleep(4 * time.Second)
	stop()

	// --- Assertions ---

	// Bootstrap: done files live under the requests-root requests-done
	require.FileExists(t, filepath.Join(conf.Execution.RequestsRootDir, config.RequestsDoneSubDir, bsSucc+".success"))
	require.FileExists(t, filepath.Join(conf.Execution.RequestsRootDir, config.RequestsDoneSubDir, bsFail+".failure.code_77"))

	// Conglomeration (metadata) - these go to metadata/requests-done
	require.FileExists(t, filepath.Join(conf.ExecutionLimitless.MetadataDir, config.RequestsDoneSubDir, cgSucc+".success"))
	require.FileExists(t, filepath.Join(conf.ExecutionLimitless.MetadataDir, config.RequestsDoneSubDir, cgFail+".failure.code_2"))

	// GL + LPP for all modules (requests-done under each module)
	for i, mod := range config.ALL_MODULES {
		// GL
		require.FileExists(t,
			filepath.Join(conf.ExecutionLimitless.WitnessDir, "GL", mod, config.RequestsDoneSubDir, glSucc[i]+".success"))
		require.FileExists(t,
			filepath.Join(conf.ExecutionLimitless.WitnessDir, "GL", mod, config.RequestsDoneSubDir, glFail[i]+".failure.code_137"))

		// LPP
		require.FileExists(t,
			filepath.Join(conf.ExecutionLimitless.WitnessDir, "LPP", mod, config.RequestsDoneSubDir, lppSucc[i]+".success"))
		require.FileExists(t,
			filepath.Join(conf.ExecutionLimitless.WitnessDir, "LPP", mod, config.RequestsDoneSubDir, lppFail[i]+".failure.code_137"))
	}
}

// --- Slightly updated createLimitlessInputFile: keeps behaviour but creates files under <dir>/<requests-from> or <dir>/<mod>/requests -->
func createLimitlessInputFile(dir, jobType string, start, end, exitWith int, mod string) string {
	var fname string
	switch jobType {
	case "bootstrap":
		fname = fmt.Sprintf("%d-%d-etv0.2.3-stv1.2.3-getZkProof.json", start, end)
	case "conglomeration":
		fname = fmt.Sprintf("%d-%d-metadata-getZkProof.json", start, end)
	case "gl":
		fname = fmt.Sprintf("%d-%d-seg-0-mod-0-gl-wit.bin", start, end)
	case "lpp":
		fname = fmt.Sprintf("%d-%d-seg-0-mod-0-lpp-wit.bin", start, end)
	default:
		panic("bad jobType")
	}

	// Module subdir if needed
	if mod != "" {
		dir = filepath.Join(dir, mod, config.RequestsFromSubDir)
	} else {
		// For bootstrap the caller passes the requests-root, so append requests-from there.
		dir = filepath.Join(dir, config.RequestsFromSubDir)
	}

	if err := os.MkdirAll(dir, 0o755); err != nil {
		panic(err)
	}

	full := filepath.Join(dir, fname)
	content := fmt.Sprintf("#!/bin/sh\nexit %d\n", exitWith)
	if err := os.WriteFile(full, []byte(content), 0o755); err != nil {
		panic(err)
	}
	return fname
}
