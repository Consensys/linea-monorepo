package controller

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"text/template"
	"time"

	"github.com/consensys/linea-monorepo/prover/config"
	"github.com/stretchr/testify/assert"
)

func TestRetryWithLarge(t *testing.T) {

	// A test command useful for testing the command generation
	var testDefinition = JobDefinition{
		// Give a name to the command
		Name: jobNameExecution,
		// The template of the output file (returns a constant template with no
		// parameters)
		OutputFileTmpl: template.Must(
			template.New("output-file").
				Parse("output-fill-constant"),
		),
		RequestRootDir: "./testdata",
	}

	jobs := []struct {
		Job
		ExpCode int
	}{
		{
			Job: Job{
				Def:        &testDefinition,
				LockedFile: "exit-0.sh",
				// Not directly needed but helpful to track the process name
				Start: 0,
				End:   0,
			},
			ExpCode: 0,
		},
		{
			Job: Job{
				Def:        &testDefinition,
				LockedFile: "exit-1.sh",
				// Not directly needed but helpful to track the process name
				Start: 1,
				End:   1,
			},
			ExpCode: 1,
		},
		{
			Job: Job{
				Def:        &testDefinition,
				LockedFile: "exit-77.sh",
				// Not directly needed but helpful to track the process name
				Start: 2,
				End:   2,
			},
			ExpCode: 77 + 10,
		},
		{
			Job: Job{
				Def:        &testDefinition,
				LockedFile: "sigkill.sh",
				// Not directly needed but helpful to track the process name
				Start: 3,
				End:   3,
			},
			ExpCode: 137,
		},
	}

	e := NewExecutor(&config.Config{
		Controller: config.Controller{
			WorkerCmdTmpl: template.Must(
				template.New("test-cmd").
					Parse("/bin/sh {{.InFile}}"),
			),
			// And the large fields. The commands adds a +10 to the return code
			// to leave an evidence that the return code was obtained through
			// running the large command.
			WorkerCmdLargeTmpl: template.Must(
				template.New("test-cmd-large").
					Parse(`/bin/sh -c "/bin/sh {{.InFile}}"; exit $(($? + 10))`),
			),
			RetryLocallyWithLargeCodes: config.DefaultRetryLocallyWithLargeCodes,
		},

		// Execution: config.Execution{
		// 	CanRunFullLarge: true,
		// },
	})

	for i := range jobs {
		status := e.Run(context.Background(), &jobs[i].Job)
		assert.Equalf(t, jobs[i].ExpCode, status.ExitCode, "got status %++v", status)
	}
}

func TestEarlyExitOnSpotInstanceMode(t *testing.T) {

	t.Skipf("this breaks the CI pipeline")

	// A test command useful for testing the command generation
	testDefinition := JobDefinition{
		// Give a name to the command
		Name: jobNameExecution,
		// The template of the output file (returns a constant template with no
		// parameters)
		OutputFileTmpl: template.Must(
			template.New("output-file").
				Parse("output-fill-constant"),
		),
		RequestRootDir: "./testdata",
	}

	e := NewExecutor(&config.Config{
		Controller: config.Controller{
			WorkerCmdTmpl: template.Must(
				template.New("test-cmd").
					Parse("/bin/sh {{.InFile}}"),
			),
		},
	})

	job := &Job{
		Def:        &testDefinition,
		LockedFile: "sleep-4.sh",
		// Not directly needed but helpful to track the process name
		Start: 0,
		End:   0,
	}

	var (
		// The context auto-cancels after 2 seconds and will not let the original
		// command finish.
		ctx, cancelMainExpiration = context.WithTimeout(context.Background(), 2*time.Second)
	)

	status := e.Run(ctx, job)
	cancelMainExpiration()

	assert.Equal(t, CodeKilledByExtSig, status.ExitCode)

	// We wait 3 more second to ensure all sub-process have exited
	time.Sleep(3 * time.Second)
}

func TestBuildCmdForLimitless(t *testing.T) {
	// minimal config to create the job definitions used by the limitless prover
	tmpRoot := t.TempDir()
	reqRoot := filepath.Join(tmpRoot, "requests-root")

	conf := config.Config{
		Execution: config.Execution{
			WithRequestDir: config.WithRequestDir{
				RequestsRootDir: reqRoot,
			},
		},
		ExecutionLimitless: config.ExecutionLimitless{
			MetadataDir: filepath.Join(tmpRoot, "metadata"),
			WitnessDir:  filepath.Join(tmpRoot, "witness"),
		},
	}

	// ensure the phase command templates are set (these are the templates
	// chosen by buildCmd for limitless jobs)
	conf.Controller.ProverPhaseCmd.BootstrapCmdTmpl = template.Must(template.New("bootstrap").Parse("bootstrap --in {{.InFile}} --out {{.OutFile}}"))
	conf.Controller.ProverPhaseCmd.ConglomerationCmdTmpl = template.Must(template.New("conglomeration").Parse("conglomeration --in {{.InFile}} --out {{.OutFile}}"))
	conf.Controller.ProverPhaseCmd.GLCmdTmpl = template.Must(template.New("gl").Parse("gl --in {{.InFile}} --out {{.OutFile}}"))
	conf.Controller.ProverPhaseCmd.LPPCmdTmpl = template.Must(template.New("lpp").Parse("lpp --in {{.InFile}} --out {{.OutFile}}"))

	// provide a fallback worker templates so buildCmd never receives a nil tmpl
	conf.Controller.WorkerCmdTmpl = template.Must(template.New("worker").Parse("worker --in {{.InFile}} --out {{.OutFile}}"))
	conf.Controller.WorkerCmdLargeTmpl = template.Must(template.New("worker-large").Parse("worker-large --in {{.InFile}} --out {{.OutFile}}"))

	// Executor under test
	exec := NewExecutor(&conf)

	// ---------- bootstrap ----------
	bootstrapDef := BootstrapDefinition(&conf)
	bootstrapName := "101-102-etv0.2.3-stv1.2.3-getZkProof.json"
	jobBootstrap, err := NewJob(&bootstrapDef, bootstrapName)
	if !assert.NoError(t, err, "NewJob(bootstrap) should succeed") {
		t.FailNow()
	}

	cmd, err := exec.buildCmd(jobBootstrap, false)
	if !assert.NoError(t, err, "buildCmd(bootstrap)") {
		t.FailNow()
	}

	// bootstrap OutFile should be a tmp response file (not /dev/null)
	// use Job.TmpResponseFile to get expected value
	expectedOut := jobBootstrap.TmpResponseFile(&conf)
	assert.True(t, strings.Contains(cmd, expectedOut), "bootstrap OutFile not used in command: %s", cmd)
	assert.True(t, strings.HasPrefix(cmd, "bootstrap "), "bootstrap template should be used")

	// ---------- conglomeration ----------
	conglomDef := ConglomerationDefinition(&conf)
	conglomName := "101-102-metadata-getZkProof.json"
	jobConglom, err := NewJob(&conglomDef, conglomName)
	if !assert.NoError(t, err, "NewJob(conglomeration) should succeed") {
		t.FailNow()
	}

	cmd, err = exec.buildCmd(jobConglom, false)
	if !assert.NoError(t, err, "buildCmd(conglomeration) should succeed") {
		t.FailNow()
	}

	expectedOut = jobConglom.TmpResponseFile(&conf)
	assert.True(t, strings.Contains(cmd, expectedOut), "conglomeration OutFile not used in command: %s", cmd)
	assert.True(t, strings.HasPrefix(cmd, "conglomeration "), "conglomeration template should be used")

	// ---------- GL (module) ----------
	glDef := GLDefinitionForModule(&conf, "ARITH-OPS")
	glName := "101-102-seg-0-mod-0-gl-wit.bin"
	jobGL, err := NewJob(&glDef, glName)
	if !assert.NoError(t, err, "NewJob(gl) should succeed") {
		t.FailNow()
	}

	cmd, err = exec.buildCmd(jobGL, false)
	if !assert.NoError(t, err, "buildCmd(gl) should succeed") {
		t.FailNow()
	}

	// GL jobs use /dev/null as out
	assert.True(t, strings.Contains(cmd, "/dev/null"), "GL command should target /dev/null: %s", cmd)
	assert.True(t, strings.HasPrefix(cmd, "gl "), "GL template should be used")

	// ---------- LPP (module) ----------
	lppDef := LPPDefinitionForModule(&conf, "ARITH-OPS")
	lppName := "101-102-seg-0-mod-0-lpp-wit.bin"
	jobLPP, err := NewJob(&lppDef, lppName)
	if !assert.NoError(t, err, "NewJob(lpp) should succeed") {
		t.FailNow()
	}

	cmd, err = exec.buildCmd(jobLPP, false)
	if !assert.NoError(t, err, "buildCmd(lpp) should succeed") {
		t.FailNow()
	}

	// LPP jobs use /dev/null as out
	assert.True(t, strings.Contains(cmd, "/dev/null"), "LPP command should target /dev/null: %s", cmd)
	assert.True(t, strings.HasPrefix(cmd, "lpp "), "LPP template should be used")
}

// TestExecutorRunLimitless verifies Executor.Run executes limitless phase commands
// and returns the script exit codes for bootstrap, conglomeration, GL and LPP.
func TestExecutorRunLimitless(t *testing.T) {
	tmp := t.TempDir()
	// small helper to create directories and fail the test on error
	mkdir := func(p string) {
		if err := os.MkdirAll(p, 0o755); err != nil {
			t.Fatalf("mkdir %s: %v", p, err)
		}
	}

	// minimal config used by JobDefinitions
	reqRoot := filepath.Join(tmp, "requests-root")
	conf := &config.Config{
		Execution: config.Execution{
			WithRequestDir: config.WithRequestDir{
				RequestsRootDir: reqRoot,
			},
		},
		ExecutionLimitless: config.ExecutionLimitless{
			MetadataDir:  filepath.Join(tmp, "metadata"),
			WitnessDir:   filepath.Join(tmp, "witness"),
			PollInterval: 1,
			Timeout:      30,
		},
	}

	// ensure directories used by job definitions exist
	mkdir(filepath.Join(conf.Execution.RequestsRootDir, config.RequestsFromSubDir))      // bootstrap expects requests here
	mkdir(filepath.Join(conf.Execution.RequestsRootDir, config.RequestsToSubDir))        // some defs reference this for responses
	mkdir(filepath.Join(conf.ExecutionLimitless.MetadataDir, config.RequestsFromSubDir)) // conglomeration input dir

	// GL and LPP module requests directories
	glReqDir := filepath.Join(conf.ExecutionLimitless.WitnessDir, "GL", "ARITH-OPS", "requests")
	mkdir(glReqDir)
	lppReqDir := filepath.Join(conf.ExecutionLimitless.WitnessDir, "LPP", "ARITH-OPS", "requests")
	mkdir(lppReqDir)

	// set phase templates so buildCmd uses them for limitless jobs
	conf.Controller.ProverPhaseCmd.BootstrapCmdTmpl = template.Must(template.New("bootstrap").Parse("/bin/sh {{.InFile}}"))
	conf.Controller.ProverPhaseCmd.ConglomerationCmdTmpl = template.Must(template.New("conglomeration").Parse("/bin/sh {{.InFile}}"))
	conf.Controller.ProverPhaseCmd.GLCmdTmpl = template.Must(template.New("gl").Parse("/bin/sh {{.InFile}}"))
	conf.Controller.ProverPhaseCmd.LPPCmdTmpl = template.Must(template.New("lpp").Parse("/bin/sh {{.InFile}}"))

	// fallback worker templates (not used for limitless jobs, but keep them non-nil)
	conf.Controller.WorkerCmdTmpl = template.Must(template.New("worker").Parse("/bin/sh {{.InFile}}"))
	conf.Controller.WorkerCmdLargeTmpl = template.Must(template.New("worker-large").Parse("/bin/sh {{.InFile}}"))

	// Executor under test
	ex := NewExecutor(conf)

	type caseDef struct {
		def      JobDefinition
		filename string // base filename (the input file name that matches the job input regexp)
		dir      string // directory where to write the file (def.dirFrom())
		wantCode int
	}

	// create cases for various limitless jobs and exit codes
	cases := []caseDef{
		// bootstrap: file lives under Execution.RequestsRootDir/requests
		{
			def:      BootstrapDefinition(conf),
			filename: "101-102-etv0.2.3-stv1.2.3-getZkProof.json",
			dir:      filepath.Join(conf.Execution.RequestsRootDir, config.RequestsFromSubDir),
			wantCode: 0,
		},
		{
			def:      BootstrapDefinition(conf),
			filename: "101-102-etv0.2.3-stv1.2.3-getZkProof.json",
			dir:      filepath.Join(conf.Execution.RequestsRootDir, config.RequestsFromSubDir),
			wantCode: 77,
		},

		// conglomeration: file lives under ExecutionLimitless.MetadataDir/requests
		{
			def:      ConglomerationDefinition(conf),
			filename: "201-202-metadata-getZkProof.json",
			dir:      filepath.Join(conf.ExecutionLimitless.MetadataDir, config.RequestsFromSubDir),
			wantCode: 0,
		},
		{
			def:      ConglomerationDefinition(conf),
			filename: "201-202-metadata-getZkProof.json",
			dir:      filepath.Join(conf.ExecutionLimitless.MetadataDir, config.RequestsFromSubDir),
			wantCode: 137,
		},

		// gl module: file lives under Witness/GL/<module>/requests
		{
			def:      GLDefinitionForModule(conf, "ARITH-OPS"),
			filename: "301-302-seg-0-mod-0-gl-wit.bin",
			dir:      glReqDir,
			wantCode: 2,
		},
		{
			def:      GLDefinitionForModule(conf, "ARITH-OPS"),
			filename: "301-302-seg-0-mod-0-gl-wit.bin",
			dir:      glReqDir,
			wantCode: 0,
		},

		// lpp module: file lives under Witness/LPP/<module>/requests
		{
			def:      LPPDefinitionForModule(conf, "ARITH-OPS"),
			filename: "401-402-seg-0-mod-0-lpp-wit.bin",
			dir:      lppReqDir,
			wantCode: 0,
		},
		{
			def:      LPPDefinitionForModule(conf, "ARITH-OPS"),
			filename: "401-402-seg-0-mod-0-lpp-wit.bin",
			dir:      lppReqDir,
			wantCode: 77,
		},
	}

	// for each case: write the script file (matching job input filename), create Job via NewJob,
	// set LockedFile to the same base filename, and run executor.Run
	for idx, c := range cases {
		// write script in expected directory; the file name must match the job input regexp
		target := filepath.Join(c.dir, c.filename)
		content := fmt.Sprintf("#!/bin/sh\nexit %d\n", c.wantCode)
		if err := os.WriteFile(target, []byte(content), 0o600); err != nil {
			t.Fatalf("case %d: write script %s: %v", idx, target, err)
		}

		// create a Job using NewJob so params (Start/End/...) are parsed correctly
		job, err := NewJob(&c.def, c.filename)
		if !assert.NoError(t, err, "NewJob should succeed for case %d (%s)", idx, c.filename) {
			continue
		}
		// emulate locked file as watcher would: set LockedFile (basename)
		job.LockedFile = c.filename

		// run the executor and assert exit code is what the script returned
		// use a short timeout context to avoid hanging tests
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		status := ex.Run(ctx, job)
		assert.Equalf(t, c.wantCode, status.ExitCode, "case %d: want exit %d got %d (job=%s)", idx, c.wantCode, status.ExitCode, c.def.Name)
	}
}
