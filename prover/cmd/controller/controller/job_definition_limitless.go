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

func createJobDefinition(conf *config.Config, jobName, inputFilePattern, outputTmpl, outputFileName string, priority int, requestsRootDir string) JobDefinition {
	return JobDefinition{
		RequestsRootDir: requestsRootDir,
		Name:            jobName,
		InputFileRegexp: regexp2.MustCompile(
			inputFilePattern,
			regexp2.None,
		),
		OutputFileTmpl: tmplMustCompile(
			outputTmpl,
			outputFileName,
		),
		Priority: priority,
		ParamsRegexp: struct {
			Start       *regexp2.Regexp
			End         *regexp2.Regexp
			Stv         *regexp2.Regexp
			Etv         *regexp2.Regexp
			Cv          *regexp2.Regexp
			ContentHash *regexp2.Regexp
		}{
			Start: regexp2.MustCompile(`^[0-9]+`, regexp2.None),
			End:   regexp2.MustCompile(`(?<=^[0-9]+-)[0-9]+`, regexp2.None),
			Etv:   matchVersionWithPrefix("etv"),
			Stv:   matchVersionWithPrefix("stv"),
		},
		FailureSuffix: matchFailureSuffix(config.FailSuffix),
	}
}

func BootstrapDefinition(conf *config.Config) JobDefinition {
	inpFileExt := ""
	if conf.Execution.CanRunFullLarge {
		inpFileExt = fmt.Sprintf(`\.%v`, config.LargeSuffix)
	}
	inputFilePattern := fmt.Sprintf(
		`^[0-9]+-[0-9]+(-etv[0-9\.]+)?(-stv[0-9\.]+)?-getZkProof\.json%v(\.failure\.%v_[0-9]+)*$`,
		inpFileExt,
		config.FailSuffix,
	)
	return createJobDefinition(conf, jobNameBootstrap, inputFilePattern, "bootstrap-exec-output-file", "{{.Start}}-{{.End}}-.getZKExecBootstrap.json", 0, conf.Bootstrap.RequestsRootDir)
}

func GLExecutionDefinition(conf *config.Config) JobDefinition {
	inputFilePattern := fmt.Sprintf(
		`^[0-9]+-[0-9]+(-etv[0-9\.]+)?(-stv[0-9\.]+)?-getZKExecBootstrap\.json(\.failure\.%v_[0-9]+)*$`,
		config.FailSuffix,
	)
	return createJobDefinition(conf, jobNameGLExecution, inputFilePattern, "gl-exec-output-file", "{{.Start}}-{{.End}}-.getZKExecGLProof.json", 1, conf.GLExecution.RequestsRootDir)
}

func RandomBeaconDefinition(conf *config.Config) JobDefinition {
	inputFilePattern := fmt.Sprintf(
		`^[0-9]+-[0-9]+(-etv[0-9\.]+)?(-stv[0-9\.]+)?-getZKExecGLProof\.json(\.failure\.%v_[0-9]+)*$`,
		config.FailSuffix,
	)
	return createJobDefinition(conf, jobNameRandomBeacon, inputFilePattern, "rnd-beacon-exec-output-file", "{{.Start}}-{{.End}}-.getZKExecRandBeacon.json", 2, conf.RandomBeacon.RequestsRootDir)
}

func LPPExecutionDefinition(conf *config.Config) JobDefinition {
	inputFilePattern := fmt.Sprintf(
		`^[0-9]+-[0-9]+(-etv[0-9\.]+)?(-stv[0-9\.]+)?-getZKExecRandBeacon\.json(\.failure\.%v_[0-9]+)*$`,
		config.FailSuffix,
	)
	return createJobDefinition(conf, jobNameLPPExecution, inputFilePattern, "lpp-exec-output-file", "{{.Start}}-{{.End}}-.getZKExecLPPProof.json", 3, conf.LPPExecution.RequestsRootDir)
}

// // TODO: Figure out: How will the request dir. work here?
// // Are we combining the responses from GL and LPP in to one file
// // Or We can set optional req dir. in the JobDefinition struct
// func ConglomerationDefinition(conf *config.Config) JobDefinition { // return JobDefinition{} // }
