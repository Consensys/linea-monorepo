package controller

import (
	"fmt"
	"text/template"

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

const (
	priorityBootstrap      = 0
	priorityGLExecution    = 1
	priorityRandomBeacon   = 2
	priorityLPPExecution   = 3
	priorityConglomeration = 4
)

const (
	bootstrapInputPattern       = `^[0-9]+-[0-9]+(-etv[0-9\.]+)?(-stv[0-9\.]+)?-getZkProof\.json%v(\.failure\.%v_[0-9]+)*$`
	glInputPattern              = `^[0-9]+-[0-9]+(-etv[0-9\.]+)?(-stv[0-9\.]+)?-getZKProof_Bootstrap_Submodule\.json%v(\.failure\.%v_[0-9]+)*$`
	randomBeaconInputPattern1   = `^[0-9]+-[0-9]+(-etv[0-9\.]+)?(-stv[0-9\.]+)?-getZKProof_Bootstrap_DistMetadata\.json%v(\.failure\.%v_[0-9]+)*$`
	randomBeaconInputPattern2   = `^[0-9]+-[0-9]+(-etv[0-9\.]+)?(-stv[0-9\.]+)?-getZKProof_GL_Beacon\.json%v(\.failure\.%v_[0-9]+)*$`
	lppInputPattern             = `^[0-9]+-[0-9]+(-etv[0-9\.]+)?(-stv[0-9\.]+)?-getZKProof_RndBeacon\.json%v(\.failure\.%v_[0-9]+)*$`
	conglomerationInputPattern1 = `^[0-9]+-[0-9]+(-etv[0-9\.]+)?(-stv[0-9\.]+)?-getZKProof_GL\.json%v(\.failure\.%v_[0-9]+)*$`
	conglomerationInputPattern2 = `^[0-9]+-[0-9]+(-etv[0-9\.]+)?(-stv[0-9\.]+)?-getZKProof_LPP\.json%v(\.failure\.%v_[0-9]+)*$`
)

const (
	bootstrapSubmoduleFile       = "{{.Start}}-{{.End}}-.getZKProof_Bootstrap_Submodule.json"
	bootstrapDistMetadataFile    = "{{.Start}}-{{.End}}-.getZKProof_Bootstrap_DistMetadata.json"
	glBeaconFile                 = "{{.Start}}-{{.End}}-.getZKProof_GL_Beacon.json"
	glOutputFile                 = "{{.Start}}-{{.End}}-.getZKProof_GL.json"
	randomBeaconDistMetadataFile = "{{.Start}}-{{.End}}-.getZKProof_Bootstrap_DistMetadata.json"
	randomBeaconGLFile           = "{{.Start}}-{{.End}}-.getZKProof_GL_Beacon.json"
	randomBeaconOutputFile       = "{{.Start}}-{{.End}}-.getZKProof_RndBeacon.json"
	lppOutputFile                = "{{.Start}}-{{.End}}-.getZKProof_LPP.json"
	conglomerationGLFile         = "{{.Start}}-{{.End}}-.getZKProof_GL.json"
	conglomerationLPPFile        = "{{.Start}}-{{.End}}-.getZKProof_LPP.json"
	conglomerationOutputFile     = "{{.Start}}-{{.End}}-.getZKProof.json"
)

type JobDefinition_Limitless struct {
	// Name of the job
	Name string

	// Priority at which this type of job should be processed. The lower the more of a priority.
	// Typically 0 for bootstrap, 1 for Gl execution, 2 for Random Beacon, 3 for LPP Execution, 4 for Conglomeration.
	Priority int

	// Parameters for the job definition provided by the user
	// There can be multiple i/p request files for a job eg: conglomeration
	ReqRootDir []string

	// The regexp to use to match input files. For instance,
	//
	// 	`^\d+-\d+-etv0.1.2-stv\d.\d.\d-getZkProof.json$`
	//
	// Will tell the controller to accept any version of the state-manager
	// but to only accept execution trace. The regexp should always start "^"
	// and end with "$" otherwise you are going to match in-progress files.
	//
	InputFilesRegexp []*regexp2.Regexp

	// Template to use to generate the output file. The template should have the
	// form of a go template. For instance,
	//
	// 	`{{.From}}-{{.To}}-pv{{.Version}}-stv{{.Stv}}-etv{{.Etv}}-zkProof.json`
	// There can be multiple output files for a job. Eg: GL-Execution
	OutputFileTmpl []*template.Template

	// The associated compiled regexp, this saves on recompiling the regexps
	// everytime we want to use them. If a field is not needed, it can be left
	// at zero.
	ParamsRegexp struct {
		Start       *regexp2.Regexp
		End         *regexp2.Regexp
		Stv         *regexp2.Regexp
		Etv         *regexp2.Regexp
		Cv          *regexp2.Regexp
		ContentHash *regexp2.Regexp
	}

	// Regexp of the failure code so that we can trim it if we want to retry.
	FailureSuffix *regexp2.Regexp
}

// Function to create a JobDefinition_Limitless
func createJobDefinition(name string, priority int,
	reqRootDir, inputFilePattern []string,
	outputTmpl, outputFileName []string) (*JobDefinition_Limitless, error) {

	numReqs, numIPs, numOPTmpl, numOPFileName := len(reqRootDir), len(inputFilePattern), len(outputTmpl), len(outputFileName)
	if numReqs != numIPs || numOPTmpl != numOPFileName {
		return nil, fmt.Errorf(`length mismatch: reqRootDir:%d, inputFilePattern:%d, 
	outputTmpl:%d, and outputFileName:%d must have the same length`, numReqs, numIPs, numOPTmpl, numOPFileName)
	}

	var inputFileRegexps []*regexp2.Regexp
	for _, pattern := range inputFilePattern {
		re, err := regexp2.Compile(pattern, regexp2.None)
		if err != nil {
			return nil, fmt.Errorf("invalid input file pattern: %v", err)
		}
		inputFileRegexps = append(inputFileRegexps, re)
	}

	var outputFileTemplates []*template.Template
	for i, tmpl := range outputTmpl {
		outputFileTemplates = append(outputFileTemplates, tmplMustCompile(tmpl, outputFileName[i]))
	}

	return &JobDefinition_Limitless{
		Name:             name,
		Priority:         priority,
		ReqRootDir:       reqRootDir,
		InputFilesRegexp: inputFileRegexps,
		OutputFileTmpl:   outputFileTemplates,
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
		FailureSuffix: matchFailureSuffix("fail"),
	}, nil
}

func BootstrapDefinition(conf *config.Config) (*JobDefinition_Limitless, error) {
	inpFileExt := ""
	if conf.Bootstrap.CanRunFullLarge {
		inpFileExt = fmt.Sprintf(`\.%v`, config.LargeSuffix)
	}
	inputFilePattern := []string{
		fmt.Sprintf(
			bootstrapInputPattern,
			inpFileExt,
			config.FailSuffix,
		),
	}
	reqRootDir := []string{conf.Bootstrap.RequestsRootDir}
	outputTmpl := []string{"exec-bootstrap-submodule-req-file", "exec-bootstrap-submodule-distmetadata-file"}
	outputFileName := []string{bootstrapSubmoduleFile, bootstrapDistMetadataFile}
	return createJobDefinition(jobNameBootstrap, priorityBootstrap, reqRootDir, inputFilePattern, outputTmpl, outputFileName)
}

func GLExecutionDefinition(conf *config.Config) (*JobDefinition_Limitless, error) {
	inpFileExt := ""
	if conf.GLExecution.CanRunFullLarge {
		inpFileExt = fmt.Sprintf(`\.%v`, config.LargeSuffix)
	}
	inputFilePattern := []string{
		fmt.Sprintf(
			glInputPattern,
			inpFileExt,
			config.FailSuffix,
		),
	}
	reqRootDir := []string{conf.GLExecution.RequestsRootDir}
	outputTmpl := []string{"exec-GL-Beacon-file", "exec-GL-output-file"}
	outputFileName := []string{glBeaconFile, glOutputFile}

	return createJobDefinition(jobNameGLExecution, priorityGLExecution, reqRootDir, inputFilePattern, outputTmpl, outputFileName)
}

func RandomBeaconDefinition(conf *config.Config) (*JobDefinition_Limitless, error) {
	inpFile1Ext, inpFile2Ext := "", ""
	if conf.RandomBeacon.Bootstrap.CanRunFullLarge && conf.RandomBeacon.GL.CanRunFullLarge {
		inpFile1Ext, inpFile2Ext = fmt.Sprintf(`\.%v`, config.LargeSuffix), fmt.Sprintf(`\.%v`, config.LargeSuffix)
	}
	inputFilePattern := []string{
		fmt.Sprintf(
			randomBeaconInputPattern1,
			inpFile1Ext,
			config.FailSuffix,
		),
		fmt.Sprintf(
			randomBeaconInputPattern2,
			inpFile2Ext,
			config.FailSuffix,
		),
	}
	reqRootDir := []string{conf.RandomBeacon.Bootstrap.RequestsRootDir, conf.RandomBeacon.GL.RequestsRootDir}
	outputTmpl := []string{"exec-rndbeacon-output-file"}
	outputFileName := []string{randomBeaconOutputFile}
	return createJobDefinition(jobNameRandomBeacon, priorityRandomBeacon, reqRootDir, inputFilePattern, outputTmpl, outputFileName)
}

func LPPExecutionDefinition(conf *config.Config) (*JobDefinition_Limitless, error) {
	inpFileExt := ""
	if conf.LPPExecution.CanRunFullLarge {
		inpFileExt = fmt.Sprintf(`\.%v`, config.LargeSuffix)
	}
	inputFilePattern := []string{
		fmt.Sprintf(
			lppInputPattern,
			inpFileExt,
			config.FailSuffix,
		),
	}
	reqRootDir := []string{conf.LPPExecution.RequestsRootDir}
	outputTmpl := []string{"exec-LPP-output-file"}
	outputFileName := []string{lppOutputFile}

	return createJobDefinition(jobNameLPPExecution, priorityLPPExecution, reqRootDir, inputFilePattern, outputTmpl, outputFileName)
}

func ConglomerationDefinition(conf *config.Config) (*JobDefinition_Limitless, error) {
	inpFile1Ext, inpFile2Ext := "", ""

	// TODO: Clairfy @linea-prover Can be have multiple limitless prover component running in different modes?
	// For eg. Bootstraper - "full-large" and GL-subprover in "full". If so? how would the file be formated
	if conf.Conglomeration.GL.CanRunFullLarge && conf.Conglomeration.LPP.CanRunFullLarge {
		inpFile1Ext, inpFile2Ext = fmt.Sprintf(`\.%v`, config.LargeSuffix), fmt.Sprintf(`\.%v`, config.LargeSuffix)
	}
	inputFilePattern := []string{
		fmt.Sprintf(
			conglomerationInputPattern1,
			inpFile1Ext,
			config.FailSuffix,
		),
		fmt.Sprintf(
			conglomerationInputPattern2,
			inpFile2Ext,
			config.FailSuffix,
		),
	}
	reqRootDir := []string{conf.Conglomeration.GL.RequestsRootDir, conf.Conglomeration.LPP.RequestsRootDir}
	outputTmpl := []string{"exec-output-file"}
	outputFileName := []string{conglomerationOutputFile}
	return createJobDefinition(jobNameConglomeration, priorityConglomeration, reqRootDir, inputFilePattern, outputTmpl, outputFileName)
}
