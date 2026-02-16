package controller

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path"
	"path/filepath"
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

func setupFsTest(t *testing.T) (confM, confL *config.Config) {

	// Testdir is going to contain the whole test directory
	testDir := t.TempDir()

	const (
		dirlogs     = "logs"
		proverM     = "prover-full-M"
		proverL     = "prover-full-L"
		execution   = "execution"
		compression = "compression"
		aggregation = "aggregation"
	)

	// Create a configuration using temporary directories
	cmd := `
/bin/sh {{.InFile}}
CODE=$?
if [ $CODE -eq 0 ]; then
	touch {{.OutFile}}
fi
exit $CODE
`
	cmdLarge := `
	/bin/sh {{.InFile}}
	CODE=$?
	CODE=$(($CODE - 12))
	if [ $CODE -eq 0 ]; then
		touch {{.OutFile}}
	fi
	exit $CODE
	`

	cmdLargeInternal := `
/bin/sh {{.InFile}}
CODE=$?
CODE=$(($CODE - 10))
if [ $CODE -eq 0 ]; then
	touch {{.OutFile}}
fi
exit $CODE
`

	// For a prover M
	confM = &config.Config{
		Version: "0.2.4",

		Controller: config.Controller{
			EnableExecution:            true,
			EnableDataAvailability:     true,
			EnableAggregation:          true,
			LocalID:                    proverM,
			Prometheus:                 config.Prometheus{Enabled: false},
			RetryDelays:                []int{0, 1},
			WorkerCmd:                  cmd,
			WorkerCmdLarge:             cmdLargeInternal,
			DeferToOtherLargeCodes:     []int{12, 137},
			RetryLocallyWithLargeCodes: []int{10, 77},
		},

		Execution: config.Execution{
			WithRequestDir: config.WithRequestDir{
				RequestsRootDir: path.Join(testDir, proverM, execution),
			},
		},
		DataAvailability: config.DataAvailability{
			WithRequestDir: config.WithRequestDir{
				RequestsRootDir: path.Join(testDir, proverM, compression),
			},
		},
		Aggregation: config.Aggregation{
			WithRequestDir: config.WithRequestDir{
				RequestsRootDir: path.Join(testDir, proverM, aggregation),
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
		os.MkdirAll(confM.DataAvailability.DirFrom(), permCode),
		os.MkdirAll(confM.DataAvailability.DirTo(), permCode),
		os.MkdirAll(confM.DataAvailability.DirDone(), permCode),
		os.MkdirAll(confM.Aggregation.DirFrom(), permCode),
		os.MkdirAll(confM.Aggregation.DirTo(), permCode),
		os.MkdirAll(confM.Aggregation.DirDone(), permCode),
	)

	if err != nil {
		t.Fatalf("could not create the temporary directories")
	}

	return confM, confL
}

func TestFileWatcherM(t *testing.T) {

	confM, _ := setupFsTest(t)

	// Create a list of files
	eFrom := confM.Execution.DirFrom
	cFrom := confM.DataAvailability.DirFrom
	aFrom := confM.Aggregation.DirFrom

	exitCode := 0 // we are not interesting in the exit code here

	// The jobs, declared in the order in which they are expected to be found

	// Name of the expected inprogress files
	expectedFNames := []struct {
		FName string
		Skip  bool
	}{
		{
			FName: createTestInputFile(eFrom(), 0, 1, execJob, exitCode),
		},
		{
			Skip:  true, // wrong directory
			FName: createTestInputFile(eFrom(), 0, 1, aggregationJob, exitCode),
		},
		{
			FName: createTestInputFile(cFrom(), 0, 1, compressionJob, exitCode),
		},
		{
			FName: createTestInputFile(eFrom(), 1, 2, execJob, exitCode),
		},
		{
			FName: createTestInputFile(cFrom(), 1, 2, compressionJob, exitCode),
		},
		{
			FName: createTestInputFile(aFrom(), 0, 2, aggregationJob, exitCode),
		},
		{
			Skip:  true, // for large only
			FName: createTestInputFile(eFrom(), 2, 4, execJob, exitCode, forLarge),
		},
		{
			FName: createTestInputFile(eFrom(), 4, 5, execJob, exitCode),
		},
		{
			FName: createTestInputFile(cFrom(), 2, 5, compressionJob, exitCode),
		},
		{
			FName: createTestInputFile(aFrom(), 2, 5, aggregationJob, exitCode),
		},
	}

	fw := NewFsWatcher(confM)

	for _, f := range expectedFNames {
		if f.Skip {
			continue
		}
		found := fw.GetBest()
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

	exitCode := 0 // we are not interesting in the exit code here

	// The jobs, declared in the order in which they are expected to be found

	// Name of the expected inprogress files
	expectedFNames := []struct {
		FName string
		Skip  bool
	}{
		{
			Skip:  true, // not large
			FName: createTestInputFile(eFrom, 0, 1, execJob, exitCode),
		},
		{
			Skip:  true, // wrong directory
			FName: createTestInputFile(eFrom, 0, 1, aggregationJob, exitCode),
		},
		{
			FName: createTestInputFile(eFrom, 1, 2, execJob, exitCode, forLarge),
		},
		{
			FName: createTestInputFile(eFrom, 2, 4, execJob, exitCode, forLarge),
		},
		{
			Skip:  true, // not large
			FName: createTestInputFile(eFrom, 4, 5, execJob, exitCode),
		},
		{
			Skip:  true, // wrong dir
			FName: createTestInputFile(eFrom, 2, 5, compressionJob, exitCode),
		},
	}

	fw := NewFsWatcher(confL)

	for _, f := range expectedFNames {
		if f.Skip {
			continue
		}
		found := fw.GetBest()
		if assert.NotNil(t, found, "did not find the job") {
			assert.Equal(t, f.FName, found.OriginalFile)
		}
	}
	assert.Nil(t, fw.GetBest(), "the queue should be empty now")
}

func setupFsTestLimitless(t *testing.T) *config.Config {
	tmpRoot := t.TempDir()

	conf := &config.Config{
		Execution: config.Execution{
			WithRequestDir: config.WithRequestDir{
				RequestsRootDir: filepath.Join(tmpRoot, "prover-execution"),
			},
		},
		ExecutionLimitless: config.ExecutionLimitless{
			MetadataDir:  filepath.Join(tmpRoot, "metadata"),
			WitnessDir:   filepath.Join(tmpRoot, "witness"),
			PollInterval: 1,  // 1s
			Timeout:      30, // 30s
		},
		Controller: config.Controller{
			LocalID:        "local-test",
			RetryDelays:    []int{0, 1},
			WorkerCmd:      "/bin/sh {{.InFile}}; CODE=$?; if [ $CODE -eq 0 ]; then touch {{.OutFile}}; fi; exit $CODE",
			WorkerCmdLarge: "/bin/sh {{.InFile}}; CODE=$?; if [ $CODE -eq 0 ]; then touch {{.OutFile}}; fi; exit $CODE",
			LimitlessJobs: config.LimitlessJobs{
				EnableBootstrapper:  true,
				EnableConglomerator: true,
				EnableGL:            true,
				GLMods:              config.ALL_MODULES[:],
				EnableLPP:           true,
				LPPMods:             config.ALL_MODULES[:],
			},
		},
	}

	// Parse templates for executor fallback
	conf.Controller.WorkerCmdTmpl = template.Must(template.New("worker").Parse(conf.Controller.WorkerCmd))
	conf.Controller.WorkerCmdLargeTmpl = template.Must(template.New("worker-large").Parse(conf.Controller.WorkerCmdLarge))

	// Phase templates for limitless (bootstrap & conglomeration need to create tmp response files)
	conf.Controller.ProverPhaseCmd.BootstrapCmdTmpl = template.Must(template.New("bootstrap").Parse("/bin/sh {{.InFile}}; CODE=$?; if [ $CODE -eq 0 ]; then touch {{.OutFile}}; fi; exit $CODE"))
	conf.Controller.ProverPhaseCmd.ConglomerationCmdTmpl = template.Must(template.New("conglomeration").Parse("/bin/sh {{.InFile}}; CODE=$?; if [ $CODE -eq 0 ]; then touch {{.OutFile}}; fi; exit $CODE"))
	// GL/LPP write to /dev/null -> no tmp response needed
	conf.Controller.ProverPhaseCmd.GLCmdTmpl = template.Must(template.New("gl").Parse("/bin/sh {{.InFile}}"))
	conf.Controller.ProverPhaseCmd.LPPCmdTmpl = template.Must(template.New("lpp").Parse("/bin/sh {{.InFile}}"))

	// helper to mkdir and fail the test on error
	mkdir := func(p string) {
		if err := os.MkdirAll(p, 0o755); err != nil {
			t.Fatalf("mkdir %s: %v", p, err)
		}
	}

	// Bootstrap requests root (requests/from, requests/done, responses)
	mkdir(filepath.Join(conf.Execution.RequestsRootDir, config.RequestsFromSubDir))
	mkdir(filepath.Join(conf.Execution.RequestsRootDir, config.RequestsDoneSubDir))
	mkdir(filepath.Join(conf.Execution.RequestsRootDir, config.RequestsToSubDir))

	// Conglomeration (metadata) - create from/to/done
	mkdir(filepath.Join(conf.ExecutionLimitless.MetadataDir, config.RequestsFromSubDir))
	mkdir(filepath.Join(conf.ExecutionLimitless.MetadataDir, config.RequestsDoneSubDir))
	mkdir(filepath.Join(conf.ExecutionLimitless.MetadataDir, config.RequestsToSubDir))

	// GL + LPP directories per module: create from/to/done for each module
	for _, mod := range conf.Controller.LimitlessJobs.GLMods {
		mkdir(filepath.Join(conf.ExecutionLimitless.WitnessDir, "GL", mod, config.RequestsFromSubDir))
		mkdir(filepath.Join(conf.ExecutionLimitless.WitnessDir, "GL", mod, config.RequestsToSubDir))
		mkdir(filepath.Join(conf.ExecutionLimitless.WitnessDir, "GL", mod, config.RequestsDoneSubDir)) // <- important
	}
	for _, mod := range conf.Controller.LimitlessJobs.LPPMods {
		mkdir(filepath.Join(conf.ExecutionLimitless.WitnessDir, "LPP", mod, config.RequestsFromSubDir))
		mkdir(filepath.Join(conf.ExecutionLimitless.WitnessDir, "LPP", mod, config.RequestsToSubDir))
		mkdir(filepath.Join(conf.ExecutionLimitless.WitnessDir, "LPP", mod, config.RequestsDoneSubDir)) // <- important
	}

	return conf
}

func TestFileWatcherDist(t *testing.T) {

	conf := setupFsTestLimitless(t)

	// --- Define test cases dynamically ---
	type testcase struct {
		name         string
		path         string
		jobName      string
		writeContent []byte
	}

	//nolint:prealloc
	var cases []testcase

	// bootstrap
	cases = append(cases, testcase{
		name:         "bootstrap",
		path:         filepath.Join(conf.Execution.RequestsRootDir, config.RequestsFromSubDir, "101-102-etv0.2.3-stv1.2.3-getZkProof.json"),
		jobName:      jobNameBootstrap,
		writeContent: []byte("{}"),
	})

	// conglomeration
	cases = append(cases, testcase{
		name:         "conglomeration",
		path:         filepath.Join(conf.ExecutionLimitless.MetadataDir, config.RequestsFromSubDir, "101-102-metadata-getZkProof.json"),
		jobName:      jobNameConglomeration,
		writeContent: []byte("{}"),
	})

	// all GL mods
	for i, mod := range conf.Controller.LimitlessJobs.GLMods {
		cases = append(cases, testcase{
			name: "gl-" + mod,
			path: filepath.Join(
				conf.ExecutionLimitless.WitnessDir,
				"GL", mod, "requests",
				fmt.Sprintf("101-102-seg-%d-mod-%d-%s-wit.bin",
					i, i, jobNameGL,
				),
			),
			jobName:      fmt.Sprintf("%s-%s", jobNameGL, mod),
			writeContent: []byte("gl-witness"),
		})
	}

	// all LPP mods
	for i, mod := range conf.Controller.LimitlessJobs.LPPMods {
		cases = append(cases, testcase{
			name: "lpp-" + mod,
			path: filepath.Join(
				conf.ExecutionLimitless.WitnessDir,
				"LPP", mod, "requests",
				fmt.Sprintf("101-102-seg-%d-mod-%d-%s-wit.bin",
					i, i, jobNameLPP,
				),
			),
			jobName:      fmt.Sprintf("%s-%s", jobNameLPP, mod),
			writeContent: []byte("lpp-witness"),
		})
	}

	// --- Write all input files ---
	for _, c := range cases {
		if err := os.WriteFile(c.path, c.writeContent, 0o600); err != nil {
			t.Fatalf("write %s file: %v", c.name, err)
		}
	}

	// Init fsWatcher
	fsWatcher := NewFsWatcher(conf)

	// expected set keyed by base filename
	expected := map[string]testcase{}
	for _, c := range cases {
		expected[filepath.Base(c.path)] = c
	}

	// --- Drain watcher ---
	for len(expected) > 0 {
		job := fsWatcher.GetBest()
		if job == nil {
			t.Fatalf("FsWatcher returned nil but expected %d more files", len(expected))
		}
		found := job.OriginalFile
		tc, ok := expected[found]
		if !ok {
			t.Fatalf("Unexpected file returned: %s", found)
		}

		// check job name
		if job.Def.Name != tc.jobName {
			t.Fatalf("job name mismatch for %s: got %s want %s", found, job.Def.Name, tc.jobName)
		}

		// check file lock
		dirFrom := job.Def.dirFrom()
		originalPath := filepath.Join(dirFrom, job.OriginalFile)
		lockedPath := filepath.Join(dirFrom, job.LockedFile)

		if _, err := os.Stat(originalPath); err == nil {
			t.Fatalf("original file still exists after lock: %s", originalPath)
		}
		if fi, err := os.Stat(lockedPath); err != nil || fi == nil {
			t.Fatalf("locked file does not exist for %s: %s (err=%v)", found, lockedPath, err)
		}

		delete(expected, found)
	}

	assert.Empty(t, expected, "expected files not found")
}

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
