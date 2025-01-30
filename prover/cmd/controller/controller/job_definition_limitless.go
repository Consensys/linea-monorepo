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

const (
	bootstrapOutputTmpl       = "exec-bootstrap-submodule-req-file"
	bootstrapDistMetadataTmpl = "exec-bootstrap-submodule-distmetadata-file"
	glBeaconOutputTmpl        = "exec-GL-Beacon-file"
	glOutputTmpl              = "exec-GL-output-file"
	randomBeaconOutputTmpl    = "exec-rndbeacon-output-file"
	lppOutputTmpl             = "exec-LPP-output-file"
	conglomerationOutputTmpl  = "exec-output-file"
)

func createJobDefinition(name string, priority int,
	reqRootDir, inputFilePattern []string,
	outputTmpl, outputFileName []string) (*JobDefinition, error) {

	numReqs, numIPs, numOPTmpl, numOPFileName := len(reqRootDir), len(inputFilePattern), len(outputTmpl), len(outputFileName)

	// Currently JobDefinition supports only primary and secondary inputs/output files
	// Length cannot exceed 2
	if numReqs > 2 || numIPs > 2 || numOPTmpl > 2 || numOPFileName > 2 {
		return nil, fmt.Errorf("input and output parameters length cannot be greater than 2")
	}

	if numReqs != numIPs || numOPTmpl != numOPFileName {
		return nil, fmt.Errorf(`length mismatch: reqRootDir:%d, inputFilePattern:%d, 
	outputTmpl:%d, and outputFileName:%d must have the same length`, numReqs, numIPs, numOPTmpl, numOPFileName)
	}

	// Set primary request root directory and compile primary input file regexps
	primaryReqRootDir := reqRootDir[0]
	inpReq1FileRegexp, err := regexp2.Compile(inputFilePattern[0], regexp2.None)
	if err != nil {
		return nil, fmt.Errorf("invalid input file pattern: %v", err)
	}

	// Set secondary request root directory and compile secondary input file regexps
	var inpReq2FileRegexp *regexp2.Regexp
	var secReqRootDir string
	if numReqs == 2 {
		secReqRootDir = reqRootDir[1]
		inpReq2FileRegexp, err = regexp2.Compile(inputFilePattern[1], regexp2.None)
		if err != nil {
			return nil, fmt.Errorf("invalid input file pattern: %v", err)
		}
	}

	// Compile output file templates
	opFile1Template := tmplMustCompile(outputTmpl[0], outputFileName[0])
	var opFile2Template *template.Template
	if numOPTmpl == 2 {
		opFile2Template = tmplMustCompile(outputTmpl[1], outputFileName[1])
	}

	return &JobDefinition{
		Name:     name,
		Priority: priority,

		// Primary and Secondary Request (Input) Files
		RequestsRootDir:    primaryReqRootDir,
		InputFileRegexp:    inpReq1FileRegexp,
		SecRequestsRootDir: secReqRootDir,
		SecInputFileRegexp: inpReq2FileRegexp,

		// Output Templates
		OutputFileTmpl:    opFile1Template,
		SecOutputFileTmpl: opFile2Template,

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

func BootstrapDefinition(conf *config.Config) (*JobDefinition, error) {
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
	outputTmpl := []string{bootstrapOutputTmpl, bootstrapDistMetadataTmpl}
	outputFileName := []string{bootstrapSubmoduleFile, bootstrapDistMetadataFile}
	return createJobDefinition(jobNameBootstrap, priorityBootstrap, reqRootDir, inputFilePattern, outputTmpl, outputFileName)
}

func GLExecutionDefinition(conf *config.Config) (*JobDefinition, error) {
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
	outputTmpl := []string{glBeaconOutputTmpl, glOutputTmpl}
	outputFileName := []string{glBeaconFile, glOutputFile}

	return createJobDefinition(jobNameGLExecution, priorityGLExecution, reqRootDir, inputFilePattern, outputTmpl, outputFileName)
}

func RandomBeaconDefinition(conf *config.Config) (*JobDefinition, error) {
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
	outputTmpl := []string{randomBeaconOutputTmpl}
	outputFileName := []string{randomBeaconOutputFile}
	return createJobDefinition(jobNameRandomBeacon, priorityRandomBeacon, reqRootDir, inputFilePattern, outputTmpl, outputFileName)
}

func LPPExecutionDefinition(conf *config.Config) (*JobDefinition, error) {
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
	outputTmpl := []string{lppOutputTmpl}
	outputFileName := []string{lppOutputFile}

	return createJobDefinition(jobNameLPPExecution, priorityLPPExecution, reqRootDir, inputFilePattern, outputTmpl, outputFileName)
}

func ConglomerationDefinition(conf *config.Config) (*JobDefinition, error) {
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
	outputTmpl := []string{conglomerationOutputTmpl}
	outputFileName := []string{conglomerationOutputFile}
	return createJobDefinition(jobNameConglomeration, priorityConglomeration, reqRootDir, inputFilePattern, outputTmpl, outputFileName)
}
