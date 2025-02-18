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
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

const (
	execBootstrapPriority int = iota
	execGLPriority
	execRndBeaconPriority
	execLPPPriority
	execConglomerationPriority
)

func TestLimitlessProverFileWatcherM(t *testing.T) {
	confM, _ := setupLimitlessFsTest(t)

	// We are not interested in the exit code here
	exitCode := 0

	// Create a list of files for each job type
	execBootstrapFrom := []string{confM.ExecBootstrap.DirFrom(0)}
	execGLFrom := []string{confM.ExecGL.DirFrom(0)}
	execRndBeaconFrom := []string{
		confM.ExecRndBeacon.DirFrom(0),
		confM.ExecRndBeacon.DirFrom(1),
	}
	execLPPFrom := []string{confM.ExecLPP.DirFrom(0)}
	execConglomerationFrom := []string{
		confM.ExecConglomeration.DirFrom(0),
		confM.ExecConglomeration.DirFrom(1),
		confM.ExecConglomeration.DirFrom(2),
	}

	// The jobs, declared in the order in which they are expected to be found
	// NOTE: It is important to test for the same starting and ending block ranges
	expectedFNames := []struct {
		FName []string
		Skip  bool
	}{
		{
			FName: createLimitlessTestInputFiles(execBootstrapFrom, 0, 1, execBootstrapPriority, exitCode),
		},
		{
			FName: createLimitlessTestInputFiles(execGLFrom, 0, 1, execGLPriority, exitCode),
		},
		{
			FName: createLimitlessTestInputFiles(execRndBeaconFrom, 0, 1, execRndBeaconPriority, exitCode),
		},
		{
			FName: createLimitlessTestInputFiles(execLPPFrom, 0, 1, execLPPPriority, exitCode),
		},
		{
			FName: createLimitlessTestInputFiles(execConglomerationFrom, 0, 1, execConglomerationPriority, exitCode),
		},
		{
			Skip:  true, // not large
			FName: createLimitlessTestInputFiles(execConglomerationFrom, 0, 1, execConglomerationPriority, exitCode),
		},
		{
			Skip:  true, // wrong dir
			FName: createLimitlessTestInputFiles(execBootstrapFrom, 0, 1, execConglomerationPriority, exitCode),
		},
	}

	fw := NewFsWatcher(confM)
	for _, f := range expectedFNames {
		if f.Skip {
			continue
		}
		t.Logf("Looking for job with file: %s", f.FName)
		found := fw.GetBest()
		t.Logf("Found job: %+v", found)
		if found == nil {
			t.Logf("Did not find the job for file: %s", f.FName)
		}
		if assert.NotNil(t, found, "did not find the job") {
			assert.Equal(t, f.FName, found.OriginalFile)
		}
	}
	assert.Nil(t, fw.GetBest(), "the queue should be empty now")

}

func TestLimitlessProverFileWatcherL(t *testing.T) {
	_, confL := setupLimitlessFsTest(t)

	// We are not interested in the exit code here
	exitCode := 0

	// Create a list of files for each job type
	execBootstrapFrom := []string{confL.ExecBootstrap.DirFrom(0)}
	execGLFrom := []string{confL.ExecGL.DirFrom(0)}
	execRndBeaconFrom := []string{
		confL.ExecRndBeacon.DirFrom(0),
		confL.ExecRndBeacon.DirFrom(1),
	}
	execLPPFrom := []string{confL.ExecLPP.DirFrom(0)}
	execConglomerationFrom := []string{
		confL.ExecConglomeration.DirFrom(0),
		confL.ExecConglomeration.DirFrom(1),
		confL.ExecConglomeration.DirFrom(2),
	}

	// The jobs, declared in the order in which they are expected to be found
	// NOTE: It is important to test for the same starting and ending block ranges
	expectedFNames := []struct {
		FName []string
		Skip  bool
	}{
		{
			FName: createLimitlessTestInputFiles(execBootstrapFrom, 0, 1, execBootstrapPriority, exitCode, forLarge),
		},
		{
			FName: createLimitlessTestInputFiles(execGLFrom, 0, 1, execGLPriority, exitCode, forLarge),
		},
		{
			FName: createLimitlessTestInputFiles(execRndBeaconFrom, 0, 1, execRndBeaconPriority, exitCode, forLarge),
		},
		{
			FName: createLimitlessTestInputFiles(execLPPFrom, 0, 1, execLPPPriority, exitCode, forLarge),
		},
		{
			FName: createLimitlessTestInputFiles(execConglomerationFrom, 0, 1, execConglomerationPriority, exitCode, forLarge),
		},
		{
			Skip:  true, // not large
			FName: createLimitlessTestInputFiles(execConglomerationFrom, 4, 5, execConglomerationPriority, exitCode),
		},
		{
			Skip:  true, // wrong dir
			FName: createLimitlessTestInputFiles(execBootstrapFrom, 0, 1, execConglomerationPriority, exitCode),
		},
	}

	fw := NewFsWatcher(confL)
	for _, f := range expectedFNames {
		if f.Skip {
			continue
		}
		t.Logf("Looking for job with file: %s", f.FName)
		found := fw.GetBest()
		t.Logf("Found job: %+v", found)
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
func setupLimitlessFsTest(t *testing.T) (confM, confL *config.Config) {
	// Testdir is going to contain the whole test directory
	testDir := t.TempDir()

	const (
		dirfrom = "prover-requests"
		dirto   = "prover-responses"
		dirdone = "requests-done"
		dirlogs = "logs"
		proverM = "prover-full-M"
		proverL = "prover-full-L"

		// Add conf. for Limitless prover: Naming convention: exec<i/p><o/p>
		execBootstrap                       = "execution"
		execBootstrapGL                     = "bootstrap-gl"
		execBootstrapMetadata               = "bootstrap-metadata"
		execBootstrapMetadataRndBeacon      = "bootstrap-metadata-rndbeacon"
		execGLRndBeacon                     = "gl-rndbeacon"
		execGLConglomeration                = "gl"
		execRndbeaconLPP                    = "rndbeacon"
		execLPPConglomeration               = "lpp"
		execBootstrapMetadataConglomeration = "bootstrap-metadata-conglomeration"
		execConglomeration                  = "execution"
	)

	cmd := `
		{{- range .InFile }}
		/bin/sh {{ . }}
		{{- end }}
		CODE=$?
		if [ $CODE -eq 0 ]; then
			{{- range .OutFile }}
			touch {{ . }}
			{{- end }}
		fi
		exit $CODE
		`
	cmdLarge := `
		{{- range .InFile }}
		/bin/sh {{ . }}
		{{- end }}
		CODE=$?
		CODE=$(($CODE - 12))
		if [ $CODE -eq 0 ]; then
			{{- range .OutFile }}
			touch {{ . }}
			{{- end }}
		fi
		exit $CODE
		`

	cmdLargeInternal := `
		{{- range .InFile }}
		/bin/sh {{ . }}
		{{- end }}
		CODE=$?
		CODE=$(($CODE - 10))
		if [ $CODE -eq 0 ]; then
			{{- range .OutFile }}
			touch {{ . }}
			{{- end }}
		fi
		exit $CODE
		`

	// For a prover M
	confM = &config.Config{
		Version: "0.2.4",

		Controller: config.Controller{
			// Disable legacy
			EnableExecution:         false,
			EnableBlobDecompression: false,
			EnableAggregation:       false,

			// Limitless prover components
			EnableExecBootstrap:        true,
			EnableExecGL:               true,
			EnableExecRndBeacon:        true,
			EnableExecLPP:              true,
			EnableExecConglomeration:   true,
			LocalID:                    proverM,
			Prometheus:                 config.Prometheus{Enabled: false},
			RetryDelays:                []int{0, 1},
			WorkerCmd:                  cmd,
			WorkerCmdLarge:             cmdLargeInternal,
			DeferToOtherLargeCodes:     []int{12, 137},
			RetryLocallyWithLargeCodes: []int{10, 77},
		},

		// Limitless prover components
		ExecBootstrap: config.Execution{
			WithRequestDir: config.WithRequestDir{
				RequestsRootDir: []string{path.Join(testDir, proverM, execBootstrap)},
			},
			WithResponseDir: config.WithResponseDir{
				ResponsesRootDir: []string{
					path.Join(testDir, proverM, execBootstrapGL),
					path.Join(testDir, proverM, execBootstrapMetadata)},
			},
		},
		ExecGL: config.Execution{
			WithRequestDir: config.WithRequestDir{
				RequestsRootDir: []string{path.Join(testDir, proverM, execBootstrapGL)},
			},
			WithResponseDir: config.WithResponseDir{
				ResponsesRootDir: []string{
					path.Join(testDir, proverM, execGLRndBeacon),
					path.Join(testDir, proverM, execGLConglomeration)},
			},
		},
		ExecRndBeacon: config.Execution{
			WithRequestDir: config.WithRequestDir{
				RequestsRootDir: []string{path.Join(testDir, proverM, execBootstrapMetadataRndBeacon),
					path.Join(testDir, proverM, execGLRndBeacon),
				},
			},
			WithResponseDir: config.WithResponseDir{
				ResponsesRootDir: []string{path.Join(testDir, proverM, execRndbeaconLPP)},
			},
		},
		ExecLPP: config.Execution{
			WithRequestDir: config.WithRequestDir{
				RequestsRootDir: []string{path.Join(testDir, proverM, execRndbeaconLPP)},
			},
			WithResponseDir: config.WithResponseDir{
				ResponsesRootDir: []string{path.Join(testDir, proverM, execLPPConglomeration)},
			},
		},
		ExecConglomeration: config.Execution{
			WithRequestDir: config.WithRequestDir{
				RequestsRootDir: []string{path.Join(testDir, proverM, execBootstrapMetadataConglomeration),
					path.Join(testDir, proverM, execGLConglomeration),
					path.Join(testDir, proverM, execLPPConglomeration),
				},
			},
			WithResponseDir: config.WithResponseDir{
				ResponsesRootDir: []string{path.Join(testDir, proverM, execConglomeration)},
			},
		},
	}

	_confL := *confM
	confL = &_confL
	confL.Controller.LocalID = proverL
	confL.Controller.WorkerCmdLarge = cmdLarge

	// Allow Limitless job to run in large mode
	confL.ExecBootstrap.CanRunFullLarge = true
	confL.ExecGL.CanRunFullLarge = true
	confL.ExecRndBeacon.CanRunFullLarge = true
	confL.ExecLPP.CanRunFullLarge = true
	confL.ExecConglomeration.CanRunFullLarge = true

	// ensure the template are parsed
	confM.Controller.WorkerCmdTmpl = template.Must(template.New("worker").Parse(confM.Controller.WorkerCmd))
	confM.Controller.WorkerCmdLargeTmpl = template.Must(template.New("worker-large").Parse(confM.Controller.WorkerCmdLarge))
	confL.Controller.WorkerCmdTmpl = template.Must(template.New("worker").Parse(confL.Controller.WorkerCmd))
	confL.Controller.WorkerCmdLargeTmpl = template.Must(template.New("worker-large").Parse(confL.Controller.WorkerCmdLarge))

	// Initialize the dirs (and give them all permissions). They will be
	// wiped out after the test anyway.
	permCode := fs.FileMode(0777)
	err := errors.Join(

		// Add stuff for Limitless prover

		// Bootstrap: 1 input -> 2 output
		os.MkdirAll(confM.ExecBootstrap.DirFrom(0), permCode),
		os.MkdirAll(confM.ExecBootstrap.DirDone(0), permCode),
		os.MkdirAll(confM.ExecBootstrap.DirTo(0), permCode),
		os.MkdirAll(confM.ExecBootstrap.DirTo(1), permCode),

		// GL: 1 input -> 2 output
		os.MkdirAll(confM.ExecGL.DirFrom(0), permCode),
		os.MkdirAll(confM.ExecGL.DirDone(0), permCode),
		// In practice there will be `n` files here
		os.MkdirAll(confM.ExecGL.DirTo(0), permCode),
		os.MkdirAll(confM.ExecGL.DirTo(1), permCode),

		// RndBeacon: 2 input -> 1 output
		// In practice there will be `n` files here
		os.MkdirAll(confM.ExecRndBeacon.DirFrom(0), permCode),
		os.MkdirAll(confM.ExecRndBeacon.DirDone(0), permCode),
		os.MkdirAll(confM.ExecRndBeacon.DirFrom(1), permCode),
		os.MkdirAll(confM.ExecRndBeacon.DirDone(1), permCode),
		os.MkdirAll(confM.ExecRndBeacon.DirTo(0), permCode),

		// LPP: 1 input -> 1 output
		// In practice there will be `n` files
		os.MkdirAll(confM.ExecLPP.DirFrom(0), permCode),
		os.MkdirAll(confM.ExecLPP.DirDone(0), permCode),
		os.MkdirAll(confM.ExecLPP.DirTo(0), permCode),

		// Conglomeration: 3 input -> 1 ouput
		// In practice there will be `2n+1` inputs => 1 output file
		os.MkdirAll(confM.ExecConglomeration.DirFrom(0), permCode),
		os.MkdirAll(confM.ExecConglomeration.DirDone(0), permCode),
		os.MkdirAll(confM.ExecConglomeration.DirFrom(1), permCode),
		os.MkdirAll(confM.ExecConglomeration.DirDone(1), permCode),
		os.MkdirAll(confM.ExecConglomeration.DirFrom(2), permCode),
		os.MkdirAll(confM.ExecConglomeration.DirDone(2), permCode),
		os.MkdirAll(confM.ExecConglomeration.DirTo(0), permCode),
	)

	if err != nil {
		t.Fatalf("could not create the temporary directories")
	}

	return confM, confL
}

// Creates test input files with specific filenames and exit codes to simulate job files for the file system watcher.
func createLimitlessTestInputFiles(
	dirFrom []string,
	start, end, jobType, exitWith int,
	large ...bool,
) (fnames []string) {
	// The filenames are expected to match the regexp pattern that we have in
	// the job definition.
	var fmtStrArr []string
	switch jobType {
	case execBootstrapPriority:
		fmtStrArr = []string{"%v-%v-etv0.1.2-stv1.2.3-getZkProof.json"}
	case execGLPriority:
		fmtStrArr = []string{"%v-%v-etv0.1.2-stv1.2.3-getZkProof_Bootstrap_GLSubmodule.json"}
	case execRndBeaconPriority:
		fmtStrArr = []string{"%v-%v-etv0.1.2-stv1.2.3-getZkProof_Bootstrap_DistMetadata.json",
			"%v-%v-etv0.1.2-stv1.2.3-getZkProof_GL_RndBeacon.json"}
	case execLPPPriority:
		fmtStrArr = []string{"%v-%v-etv0.1.2-stv1.2.3-getZkProof_RndBeacon.json"}
	case execConglomerationPriority:
		fmtStrArr = []string{"%v-%v-etv0.1.2-stv1.2.3-getZkProof_Bootstrap_DistMetadata.json",
			"%v-%v-etv0.1.2-stv1.2.3-getZkProof_GL.json",
			"%v-%v-etv0.1.2-stv1.2.3-getZkProof_LPP.json"}
	default:
		panic("incorrect job type")
	}

	m, n := len(dirFrom), len(fmtStrArr)
	if m != n {
		logrus.Debugf("number of entries in dirFrom:%d should match with the number of formated input files:%d", m, n)
		return nil
	}

	fnames = make([]string, len(fmtStrArr))
	for i, fmtString := range fmtStrArr {
		fnames[i] = fmt.Sprintf(fmtString, start, end)
		if len(large) > 0 && large[0] {
			fnames[i] += ".large"
		}
		f, err := os.Create(path.Join(dirFrom[i], fnames[i]))
		if err != nil {
			panic(err)
		}

		// If called (with the test configuration (i.e. with sh), the file will
		// immediately exit with the provided error code)
		f.WriteString(fmt.Sprintf("#!/bin/sh\nexit %v", exitWith))
		f.Close()
	}

	return fnames
}
