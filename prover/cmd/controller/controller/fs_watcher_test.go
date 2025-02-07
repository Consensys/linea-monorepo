package controller

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path"
	"testing"
	"text/template"

	"github.com/consensys/linea-monorepo/prover/config"
	"github.com/stretchr/testify/assert"
)

const (
	execJob int = iota
	compressionJob
	aggregationJob
	forLarge bool = true
)

func TestLsName(t *testing.T) {
	dir := t.TempDir()

	// When the dir doesn't exist we should return an error
	_, err := lsname("/dir-that-does-not-exist-45464343434")
	assert.Errorf(t, err, "no error on non-existing directory")

	// When the dir exists and is non-empty (cwd will be non-empty)
	ls, err := lsname(".")
	assert.NoErrorf(t, err, "error on current directory")
	assert.NotEmptyf(t, ls, "empty on cwd")

	// When the directory is empty
	ls, err = lsname(dir)
	assert.NoErrorf(t, err, "error on tmp directory")
	assert.Emptyf(t, ls, "non empty dir")
}

func TestFileWatcherM(t *testing.T) {
	confM, _ := setupFsTest(t)

	// Create a list of files
	eFrom := confM.Execution.DirFrom
	cFrom := confM.BlobDecompression.DirFrom
	aFrom := confM.Aggregation.DirFrom

	exitCode := 0 // we are not interested in the exit code here

	// The jobs, declared in the order in which they are expected to be found

	// Name of the expected in-progress files
	expectedFNames := []struct {
		FName []string
		Skip  bool
	}{
		{
			FName: []string{createTestInputFile(eFrom(), 0, 1, execJob, exitCode)},
		},
		{
			Skip:  true, // wrong directory
			FName: []string{createTestInputFile(eFrom(), 0, 1, aggregationJob, exitCode)},
		},
		{
			FName: []string{createTestInputFile(cFrom(), 0, 1, compressionJob, exitCode)},
		},
		{
			FName: []string{createTestInputFile(eFrom(), 1, 2, execJob, exitCode)},
		},
		{
			FName: []string{createTestInputFile(cFrom(), 1, 2, compressionJob, exitCode)},
		},
		{
			FName: []string{createTestInputFile(aFrom(), 0, 2, aggregationJob, exitCode)},
		},
		{
			Skip:  true, // for large only
			FName: []string{createTestInputFile(eFrom(), 2, 4, execJob, exitCode, forLarge)},
		},
		{
			FName: []string{createTestInputFile(eFrom(), 4, 5, execJob, exitCode)},
		},
		{
			FName: []string{createTestInputFile(cFrom(), 2, 5, compressionJob, exitCode)},
		},
		{
			FName: []string{createTestInputFile(aFrom(), 2, 5, aggregationJob, exitCode)},
		},
	}

	fw := NewFsWatcher(confM)
	for _, f := range expectedFNames {
		if f.Skip {
			continue
		}
		t.Logf("Looking for job with file: %s", f.FName)
		found := fw.GetBest()
		if found == nil {
			t.Logf("Did not find the job for file: %s", f.FName)
		}
		if assert.NotNil(t, found, "did not find the job") {
			assert.Equal(t, f.FName, found.OriginalFile)
		}
	}
	assert.Nil(t, fw.GetBest(), "the queue should be empty now")

}

func TestFileWatcherL(t *testing.T) {
	_, confL := setupFsTest(t)

	// Create a list of files
	eFrom := confL.Execution.DirFrom()

	exitCode := 0 // we are not interested in the exit code here

	// The jobs, declared in the order in which they are expected to be found

	// Name of the expected in-progress files
	expectedFNames := []struct {
		FName []string
		Skip  bool
	}{
		{
			Skip:  true, // not large
			FName: []string{createTestInputFile(eFrom, 0, 1, execJob, exitCode)},
		},
		{
			Skip:  true, // wrong directory
			FName: []string{createTestInputFile(eFrom, 0, 1, aggregationJob, exitCode)},
		},
		{
			FName: []string{createTestInputFile(eFrom, 1, 2, execJob, exitCode, forLarge)},
		},
		{
			FName: []string{createTestInputFile(eFrom, 2, 4, execJob, exitCode, forLarge)},
		},
		{
			Skip:  true, // not large
			FName: []string{createTestInputFile(eFrom, 4, 5, execJob, exitCode)},
		},
		{
			Skip:  true, // wrong dir
			FName: []string{createTestInputFile(eFrom, 2, 5, compressionJob, exitCode)},
		},
	}

	fw := NewFsWatcher(confL)

	for _, f := range expectedFNames {
		if f.Skip {
			continue
		}
		t.Logf("Looking for job with file: %s", f.FName)
		found := fw.GetBest()
		if found == nil {
			t.Logf("Did not find the job for file: %s", f.FName)
		}
		if assert.NotNil(t, found, "did not find the job") {
			assert.Equal(t, f.FName, found.OriginalFile)
		}
	}
	assert.Nil(t, fw.GetBest(), "the queue should be empty now")
}

// Sets up the test environment by creating temporary directories and configurations for the prover.
func setupFsTest(t *testing.T) (confM, confL *config.Config) {
	// Testdir is going to contain the whole test directory
	testDir := t.TempDir()

	const (
		dirfrom     = "prover-requests"
		dirto       = "prover-responses"
		dirdone     = "requests-done"
		dirlogs     = "logs"
		proverM     = "prover-full-M"
		proverL     = "prover-full-L"
		execution   = "execution"
		compression = "compression"
		aggregation = "aggregation"

		// Add conf. for Limitless prover: Naming convention: exec<i/p><o/p>
		exec                  = "bootstrap"
		execBootstrapGL       = "bootstrapGl"
		execBootstrapMetadata = "bootstrapMetadata"
		execGLRndBeacon       = "gl-rndbeacon"
		execGLConglomeration  = "gl"
		execRndbeaconLPP      = "rndbeacon"
		execLPPConglomeration = "lpp"
	)

	// Create a configuration using temporary directories
	// Defines three command templates for different types of jobs.
	// These templates will be used to create shell commands for the worker processes.
	cmd := `
/bin/sh {{index .InFile 0}}
CODE=$?
if [ $CODE -eq 0 ]; then
	touch {{index .OutFile 0}}
fi
exit $CODE
`
	cmdLarge := `
/bin/sh {{index .InFile 0}}
CODE=$?
CODE=$(($CODE - 12))
if [ $CODE -eq 0 ]; then
	touch {{index .OutFile 0}}
fi
exit $CODE
`

	cmdLargeInternal := `
/bin/sh {{index .InFile 0}}
CODE=$?
CODE=$(($CODE - 10))
if [ $CODE -eq 0 ]; then
	touch {{index .OutFile 0}}
fi
exit $CODE
`
	// For a prover M
	confM = &config.Config{
		Version: "0.2.4",

		Controller: config.Controller{
			EnableExecution:            true,
			EnableBlobDecompression:    true,
			EnableAggregation:          true,
			LocalID:                    proverM,
			Prometheus:                 config.Prometheus{Enabled: false},
			RetryDelays:                []int{0, 1},
			WorkerCmd:                  cmd,
			WorkerCmdLarge:             cmdLargeInternal,
			DeferToOtherLargeCodes:     []int{12, 137},
			RetryLocallyWithLargeCodes: []int{10, 77},

			// Limitless prover components
			EnableExecBootstrap:      true,
			EnableExecGL:             true,
			EnableExecRndBeacon:      true,
			EnableExecLPP:            true,
			EnableExecConglomeration: true,
		},

		Execution: config.Execution{
			WithRequestDir: config.WithRequestDir{
				RequestsRootDir: path.Join(testDir, proverM, execution),
			},
		},
		BlobDecompression: config.BlobDecompression{
			WithRequestDir: config.WithRequestDir{
				RequestsRootDir: path.Join(testDir, proverM, compression),
			},
		},
		Aggregation: config.Aggregation{
			WithRequestDir: config.WithRequestDir{
				RequestsRootDir: path.Join(testDir, proverM, aggregation),
			},
		},

		// Limitless prover components
		ExecBootstrap: config.Execution{
			WithRequestDir: config.WithRequestDir{
				RequestsRootDir: path.Join(testDir, proverM, exec),
			},
		},
		ExecGL: config.Execution{
			WithRequestDir: config.WithRequestDir{
				RequestsRootDir: path.Join(testDir, proverM, execBootstrapGL),
			},
		},
		ExecRndBeacon: config.RndBeacon{
			GL: config.WithRequestDir{
				RequestsRootDir: path.Join(testDir, proverM, execGLRndBeacon),
			},
			BootstrapMetadata: config.WithRequestDir{
				RequestsRootDir: path.Join(testDir, proverM, execBootstrapMetadata),
			},
		},
		ExecLPP: config.Execution{
			WithRequestDir: config.WithRequestDir{
				RequestsRootDir: path.Join(testDir, proverM, execRndbeaconLPP),
			},
		},
		ExecConglomeration: config.Conglomeration{
			GL: config.WithRequestDir{
				RequestsRootDir: path.Join(testDir, proverM, execGLConglomeration),
			},
			LPP: config.WithRequestDir{
				RequestsRootDir: path.Join(testDir, proverM, execLPPConglomeration),
			},
			BootstrapMetadata: config.WithRequestDir{
				RequestsRootDir: path.Join(testDir, proverM, execBootstrapMetadata),
			},
		},
	}

	_confL := *confM
	confL = &_confL
	confL.Controller.LocalID = proverL
	confL.Controller.WorkerCmdLarge = cmdLarge
	confL.Execution.CanRunFullLarge = true

	// ensure the template are parsed
	confM.Controller.WorkerCmdTmpl = template.Must(template.New("worker").Parse(confM.Controller.WorkerCmd))
	confM.Controller.WorkerCmdLargeTmpl = template.Must(template.New("worker-large").Parse(confM.Controller.WorkerCmdLarge))
	confL.Controller.WorkerCmdTmpl = template.Must(template.New("worker").Parse(confL.Controller.WorkerCmd))
	confL.Controller.WorkerCmdLargeTmpl = template.Must(template.New("worker-large").Parse(confL.Controller.WorkerCmdLarge))

	// Initialize the dirs (and give them all permissions). They will be
	// wiped out after the test anyway.
	permCode := fs.FileMode(0777)
	err := errors.Join(
		os.MkdirAll(confM.Execution.DirFrom(), permCode),
		os.MkdirAll(confM.Execution.DirTo(), permCode),
		os.MkdirAll(confM.Execution.DirDone(), permCode),
		os.MkdirAll(confM.BlobDecompression.DirFrom(), permCode),
		os.MkdirAll(confM.BlobDecompression.DirTo(), permCode),
		os.MkdirAll(confM.BlobDecompression.DirDone(), permCode),
		os.MkdirAll(confM.Aggregation.DirFrom(), permCode),
		os.MkdirAll(confM.Aggregation.DirTo(), permCode),
		os.MkdirAll(confM.Aggregation.DirDone(), permCode),

		// Add stuff for Limitless prover
		os.MkdirAll(confM.ExecBootstrap.DirFrom(), permCode),
		os.MkdirAll(confM.ExecBootstrap.DirTo(), permCode),
		os.MkdirAll(confM.ExecBootstrap.DirDone(), permCode),

		os.MkdirAll(confM.ExecGL.DirFrom(), permCode),
		os.MkdirAll(confM.ExecGL.DirTo(), permCode),
		os.MkdirAll(confM.ExecGL.DirDone(), permCode),

		os.MkdirAll(confM.ExecRndBeacon.GL.DirFrom(), permCode),
		os.MkdirAll(confM.ExecRndBeacon.GL.DirTo(), permCode),
		os.MkdirAll(confM.ExecRndBeacon.GL.DirDone(), permCode),
		os.MkdirAll(confM.ExecRndBeacon.BootstrapMetadata.DirFrom(), permCode),
		os.MkdirAll(confM.ExecRndBeacon.BootstrapMetadata.DirTo(), permCode),
		os.MkdirAll(confM.ExecRndBeacon.BootstrapMetadata.DirDone(), permCode),

		os.MkdirAll(confM.ExecLPP.DirFrom(), permCode),
		os.MkdirAll(confM.ExecLPP.DirTo(), permCode),
		os.MkdirAll(confM.ExecLPP.DirDone(), permCode),

		os.MkdirAll(confM.ExecConglomeration.GL.DirFrom(), permCode),
		os.MkdirAll(confM.ExecConglomeration.GL.DirTo(), permCode),
		os.MkdirAll(confM.ExecConglomeration.GL.DirDone(), permCode),

		os.MkdirAll(confM.ExecConglomeration.LPP.DirFrom(), permCode),
		os.MkdirAll(confM.ExecConglomeration.LPP.DirTo(), permCode),
		os.MkdirAll(confM.ExecConglomeration.LPP.DirDone(), permCode),

		os.MkdirAll(confM.ExecConglomeration.BootstrapMetadata.DirFrom(), permCode),
		os.MkdirAll(confM.ExecConglomeration.BootstrapMetadata.DirTo(), permCode),
		os.MkdirAll(confM.ExecConglomeration.BootstrapMetadata.DirDone(), permCode),
	)

	if err != nil {
		t.Fatalf("could not create the temporary directories")
	}

	return confM, confL
}

// Creates test input files with specific filenames and exit codes to simulate job files for the file system watcher.
func createTestInputFile(
	dirfrom string,
	start, end, jobType, exitWith int,
	large ...bool,
) (fname string) {
	// The filenames are expected to match the regexp pattern that we have in
	// the job definition.
	fmtString := ""
	switch jobType {
	case execJob:
		fmtString = "%v-%v-etv0.1.2-stv1.2.3-getZkProof.json"
	case compressionJob:
		fmtString = "%v-%v-bcv0.1.2-ccv0.1.2-getZkBlobCompressionProof.json"
	case aggregationJob:
		fmtString = "%v-%v-deadbeef57-getZkAggregatedProof.json"
	default:
		panic("incorrect job type")
	}

	filename := fmt.Sprintf(fmtString, start, end)
	if len(large) > 0 && large[0] {
		filename += ".large"
	}
	f, err := os.Create(path.Join(dirfrom, filename))
	if err != nil {
		panic(err)
	}

	// If called (with the test configuration (i.e. with sh), the file will
	// immediately exit with the provided error code)
	f.WriteString(fmt.Sprintf("#!/bin/sh\nexit %v", exitWith))
	f.Close()

	return filename
}
