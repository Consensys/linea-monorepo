package controller

import (
	"testing"
	"text/template"

	"github.com/consensys/linea-monorepo/prover/config"
	"github.com/stretchr/testify/assert"
)

func TestExecBootstrapRetryWithLarge(t *testing.T) {
	// A test command useful for testing the command generation
	var testDefinition = JobDefinition{

		// Give a name to the command
		Name: jobExecBootstrap,

		// The template of the output file (returns a constant template with no
		// parameters)
		OutputFileTmpl: []*template.Template{
			template.Must(template.New("output-file").Parse("output-fill-constant")),
			template.Must(template.New("output-file").Parse("output-fill-constant")),
		},

		RequestsRootDir:  []string{"./testdata"},
		ResponsesRootDir: []string{"./responses", "./responses"},
	}

	jobs := []struct {
		Job
		ExpCode int
	}{
		{
			Job: Job{
				Def:        &testDefinition,
				LockedFile: []string{"exit-0.sh"},
				// Not directly needed but helpful to track the process name
				Start: []int{0},
				End:   []int{0},
			},
			ExpCode: 0,
		},
		{
			Job: Job{
				Def:        &testDefinition,
				LockedFile: []string{"exit-1.sh"},
				// Not directly needed but helpful to track the process name
				Start: []int{1},
				End:   []int{1},
			},
			ExpCode: 1,
		},
		{
			Job: Job{
				Def:        &testDefinition,
				LockedFile: []string{"exit-77.sh"},
				// Not directly needed but helpful to track the process name
				Start: []int{2},
				End:   []int{2},
			},
			ExpCode: 77 + 10,
		},
		{
			Job: Job{
				Def:        &testDefinition,
				LockedFile: []string{"sigkill.sh"},
				// Not directly needed but helpful to track the process name
				Start: []int{3},
				End:   []int{3},
			},
			ExpCode: 137,
		},
	}

	e := NewExecutor(&config.Config{
		Controller: config.Controller{
			WorkerCmdTmpl: template.Must(
				template.New("test-cmd").
					Parse("/bin/sh {{index .InFile 0}}"),
			),
			// And the large fields. The commands adds a +10 to the return code
			// to leave an evidence that the return code was obtained through
			// running the large command.
			WorkerCmdLargeTmpl: template.Must(
				template.New("test-cmd-large").
					Parse(`/bin/sh -c "/bin/sh {{index .InFile 0}}"; exit $(($? + 10))`),
			),
			RetryLocallyWithLargeCodes: config.DefaultRetryLocallyWithLargeCodes,
		},
	})

	for i := range jobs {
		status := e.Run(&jobs[i].Job)
		assert.Equalf(t, jobs[i].ExpCode, status.ExitCode, "got status %++v", status)
	}
}
