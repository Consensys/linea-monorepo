package controller

import (
	"fmt"

	"github.com/consensys/linea-monorepo/prover/config"
	"github.com/dlclark/regexp2"
)

const (
	jobNameBootstrap      = "bootstrap-execution"
	jobNameGLExecution    = "gl-execution"
	jobNameRandomBeacon   = "randomness-beacon-execution"
	jobNameLPPExecution   = "lpp-execution"
	jobNameConglomeration = "conglomeration-execution"
)

// BootstrapDefinition: Defines the "one-at-all" bootstrap job
func BootstrapDefinition(conf *config.Config) JobDefinition {
	jobDef := commonExecJobDef(conf, jobNameBootstrap, 0)

	// Format the extension part of the regexp if provided
	inpFileExt := ""
	if conf.Execution.CanRunFullLarge {
		inpFileExt = fmt.Sprintf(`\.%v`, config.LargeSuffix)
	}

	// Set the InputFileRegexp specific to ExecutionDefinition
	jobDef.InputFileRegexp = regexp2.MustCompile(
		fmt.Sprintf(
			`^[0-9]+-[0-9]+(-etv[0-9\.]+)?(-stv[0-9\.]+)?-getZkProof\.json%v(\.failure\.%v_[0-9]+)*$`,
			inpFileExt,
			config.FailSuffix,
		),
		regexp2.None,
	)

	return jobDef
}

// Function to define Conglomeration job
func ConglomerationDefinition(conf *config.Config) JobDefinition {
	jobDef := commonExecJobDef(conf, jobNameConglomeration, 4)

	// Set the OutputFileTmpl specific to ConglomerationDefinition
	jobDef.OutputFileTmpl = tmplMustCompile(
		"exec-output-file",
		"{{.Start}}-{{.End}}-getZkProof.json",
	)

	return jobDef
}

// TODO: Define other jobs
