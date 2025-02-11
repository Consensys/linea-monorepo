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
	"github.com/consensys/linea-monorepo/prover/utils"
)

const (
	Bootstrap int = iota
	GL
	RndBeacon
	LPP
	Conglomeration
)

// TODO: Write this test
func TestLimitlessProverFileWatcherL(t *testing.T) {

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
		execBootstrap         = "execution"
		execBootstrapGL       = "bootstrapGl"
		execBootstrapMetadata = "bootstrapMetadata"
		execGLRndBeacon       = "gl-rndbeacon"
		execGLConglomeration  = "gl"
		execRndbeaconLPP      = "rndbeacon"
		execLPPConglomeration = "lpp"
		execConglomeration    = "execution"
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
		ExecRndBeacon: config.RndBeacon{
			GL: config.WithRequestDir{
				RequestsRootDir: []string{path.Join(testDir, proverM, execGLRndBeacon)}, // In practice, there will be `n` files
			},
			BootstrapMetadata: config.WithRequestDir{
				RequestsRootDir: []string{path.Join(testDir, proverM, execBootstrapMetadata)},
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
		ExecConglomeration: config.Conglomeration{
			GL: config.WithRequestDir{
				RequestsRootDir: []string{path.Join(testDir, proverM, execGLConglomeration)},
			},
			LPP: config.WithRequestDir{
				RequestsRootDir: []string{path.Join(testDir, proverM, execLPPConglomeration)},
			},
			BootstrapMetadata: config.WithRequestDir{
				RequestsRootDir: []string{path.Join(testDir, proverM, execBootstrapMetadata)},
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
		os.MkdirAll(confM.ExecRndBeacon.GL.DirFrom(0), permCode),
		os.MkdirAll(confM.ExecRndBeacon.GL.DirDone(0), permCode),
		os.MkdirAll(confM.ExecRndBeacon.BootstrapMetadata.DirFrom(0), permCode),
		os.MkdirAll(confM.ExecRndBeacon.BootstrapMetadata.DirDone(0), permCode),
		os.MkdirAll(confM.ExecRndBeacon.DirTo(0), permCode),

		// LPP: 1 input -> 1 output
		// In practice there will be `n` files
		os.MkdirAll(confM.ExecLPP.DirFrom(0), permCode),
		os.MkdirAll(confM.ExecLPP.DirDone(0), permCode),
		os.MkdirAll(confM.ExecLPP.DirTo(0), permCode),

		// Conglomeration: 3 input -> 1 ouput
		// In practice there will be `2n+1` inputs => 1 output file
		os.MkdirAll(confM.ExecConglomeration.GL.DirFrom(0), permCode),
		os.MkdirAll(confM.ExecConglomeration.GL.DirDone(0), permCode),

		os.MkdirAll(confM.ExecConglomeration.LPP.DirFrom(0), permCode),
		os.MkdirAll(confM.ExecConglomeration.LPP.DirDone(0), permCode),

		os.MkdirAll(confM.ExecConglomeration.BootstrapMetadata.DirFrom(0), permCode),
		os.MkdirAll(confM.ExecConglomeration.BootstrapMetadata.DirDone(0), permCode),

		os.MkdirAll(confM.ExecConglomeration.DirTo(0), permCode),
	)

	if err != nil {
		t.Fatalf("could not create the temporary directories")
	}

	return confM, confL
}

// Creates test input files with specific filenames and exit codes to simulate job files for the file system watcher.
func createLimitlessTestInputFile(
	dirFrom []string,
	start, end, jobType, exitWith int,
	large ...bool,
) (fnames []string) {
	// The filenames are expected to match the regexp pattern that we have in
	// the job definition.
	var fmtStrArr []string
	switch jobType {
	case Bootstrap:
		fmtStrArr = []string{"%v-%v-etv0.1.2-stv1.2.3-getZkProof.json"}
	case GL:
		fmtStrArr = []string{"%v-%v-etv0.1.2-stv1.2.3-getZkProof_Bootstrap_GLSubmodule.json"}
	case RndBeacon:
		fmtStrArr = []string{"%v-%v-etv0.1.2-stv1.2.3-getZkProof_Bootstrap_DistMetadata.json",
			"%v-%v-etv0.1.2-stv1.2.3-getZkProof_Bootstrap_GLSubmodule.json"}
	case LPP:
		fmtStrArr = []string{"%v-%v-etv0.1.2-stv1.2.3-getZkProof_RndBeacon.json"}
	case Conglomeration:
		fmtStrArr = []string{"%v-%v-etv0.1.2-stv1.2.3-getZkProof_Bootstrap_DistMetadata.json",
			"%v-%v-etv0.1.2-stv1.2.3-getZkProof_GLjson",
			"%v-%v-etv0.1.2-stv1.2.3-getZkProof_LPP.json"}
	default:
		panic("incorrect job type")
	}

	m, n := len(dirFrom), len(fmtStrArr)
	if m != n {
		utils.Panic("number of entries in dirFrom:%d should match with the number of formated input files:%d", m, n)
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
