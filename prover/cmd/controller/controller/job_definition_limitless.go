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

func createJobDefinition(name string, priority int, reqRootDir string, inputFilePattern string, optReqRootDir string, optInputFilePattern string, outputTmpl string, outputFileName string) JobDefinition {
	jd := JobDefinition{
		Name:            name,
		Priority:        priority,
		RequestsRootDir: reqRootDir,
		InputFileRegexp: regexp2.MustCompile(
			inputFilePattern,
			regexp2.None,
		),
		OutputFileTmpl: tmplMustCompile(
			outputTmpl,
			outputFileName,
		),
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

	// Additional check for optional dirs => Congolomeration
	if optInputFilePattern != "" && optReqRootDir != "" {
		jd.OptReqRootDir = optReqRootDir
		jd.OptInputFileRegexp = regexp2.MustCompile(optInputFilePattern, regexp2.None)
	}

	return jd
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
	return createJobDefinition(jobNameBootstrap, 0, conf.Bootstrap.RequestsRootDir, inputFilePattern, "",
		"", "bootstrap-exec-output-file", "{{.Start}}-{{.End}}-.getZKExecBootstrap.json")
}

func GLExecutionDefinition(conf *config.Config) JobDefinition {
	inputFilePattern := fmt.Sprintf(
		`^[0-9]+-[0-9]+(-etv[0-9\.]+)?(-stv[0-9\.]+)?-getZKExecBootstrap\.json(\.failure\.%v_[0-9]+)*$`,
		config.FailSuffix,
	)
	return createJobDefinition(jobNameGLExecution, 1, conf.GLExecution.RequestsRootDir, inputFilePattern, "",
		"", "gl-exec-output-file", "{{.Start}}-{{.End}}-.getZKExecGLProof.json")
}

func RandomBeaconDefinition(conf *config.Config) JobDefinition {
	inputFilePattern := fmt.Sprintf(
		`^[0-9]+-[0-9]+(-etv[0-9\.]+)?(-stv[0-9\.]+)?-getZKExecGLProof\.json(\.failure\.%v_[0-9]+)*$`,
		config.FailSuffix,
	)
	return createJobDefinition(jobNameRandomBeacon, 2, conf.RandomBeacon.RequestsRootDir, inputFilePattern, "",
		"", "rnd-beacon-exec-output-file", "{{.Start}}-{{.End}}-.getZKExecRandBeacon.json")
}

func LPPExecutionDefinition(conf *config.Config) JobDefinition {
	inputFilePattern := fmt.Sprintf(
		`^[0-9]+-[0-9]+(-etv[0-9\.]+)?(-stv[0-9\.]+)?-getZKExecRandBeacon\.json(\.failure\.%v_[0-9]+)*$`,
		config.FailSuffix,
	)
	return createJobDefinition(jobNameLPPExecution, 3, conf.LPPExecution.RequestsRootDir, inputFilePattern, "",
		"", "lpp-exec-output-file", "{{.Start}}-{{.End}}-.getZKExecLPPProof.json")
}

func ConglomerationDefinition(conf *config.Config) JobDefinition {
	inputFilePattern := fmt.Sprintf(
		`^[0-9]+-[0-9]+(-etv[0-9\.]+)?(-stv[0-9\.]+)?-getZKExecGLProof\.json(\.failure\.%v_[0-9]+)*$`,
		config.FailSuffix,
	)

	optFilePattern := fmt.Sprintf(
		`^[0-9]+-[0-9]+(-etv[0-9\.]+)?(-stv[0-9\.]+)?-getZKExecLPPProof\.json(\.failure\.%v_[0-9]+)*$`,
		config.FailSuffix,
	)

	return createJobDefinition(jobNameConglomeration, 4, conf.Conglomeration.GLResp.RequestsRootDir,
		inputFilePattern, conf.Conglomeration.LPPResp.RequestsRootDir, optFilePattern,
		"exec-output-file", "{{.Start}}-{{.End}}-.getZKProof.json")
}
