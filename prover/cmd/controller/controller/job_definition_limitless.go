package controller

import (
	"fmt"

	"github.com/consensys/linea-monorepo/prover/config"
	"github.com/consensys/linea-monorepo/prover/utils"
)

const (
	// Job definitions
	jobExecBootstrap        = "execBootstrap"
	jobExecGL               = "execGL"
	jobExecRndBeacon        = "execRndbeacon"
	jobExecLPP              = "execLPP"
	jobExecCongolomerateLPP = "execConglomeration"

	// Priorities
	priorityExecBootstrap       = 0
	priorityExecGL              = 1
	priorityExecRndBeacon       = 2
	priorityExecLPP             = 3
	priorityExecCongolomeration = 4

	// Input file patterns
	execBootstrapInputPattern                    = `^[0-9]+-[0-9]+(-etv[0-9\.]+)?(-stv[0-9\.]+)?-getZkProof\.json%v(\.failure\.%v_[0-9]+)*$`
	execBootstrapGLInputPattern                  = `^[0-9]+-[0-9]+(-etv[0-9\.]+)?(-stv[0-9\.]+)?-getZkProof_Bootstrap_GLSubmodule\.json%v(\.failure\.%v_[0-9]+)*$`
	execBootstrapRndBeaconInputPattern           = `^[0-9]+-[0-9]+(-etv[0-9\.]+)?(-stv[0-9\.]+)?-getZkProof_Bootstrap_DistMetadata\.json%v(\.failure\.%v_[0-9]+)*$`
	execGLRndBeaconInputPattern                  = `^[0-9]+-[0-9]+(-etv[0-9\.]+)?(-stv[0-9\.]+)?-getZkProof_GL_RndBeacon\.json%v(\.failure\.%v_[0-9]+)*$`
	execLPPInputPattern                          = `^[0-9]+-[0-9]+(-etv[0-9\.]+)?(-stv[0-9\.]+)?-getZkProof_RndBeacon\.json%v(\.failure\.%v_[0-9]+)*$`
	execConglomerateGLInputPattern               = `^[0-9]+-[0-9]+(-etv[0-9\.]+)?(-stv[0-9\.]+)?-getZkProof_GL\.json%v(\.failure\.%v_[0-9]+)*$`
	execConglomerateLPPInputPattern              = `^[0-9]+-[0-9]+(-etv[0-9\.]+)?(-stv[0-9\.]+)?-getZkProof_LPP\.json%v(\.failure\.%v_[0-9]+)*$`
	execConglomerateBootstrapDistMetadataPattern = `^[0-9]+-[0-9]+(-etv[0-9\.]+)?(-stv[0-9\.]+)?-getZkProof_Bootstrap_DistMetadata\.json%v(\.failure\.%v_[0-9]+)*$`

	// Output file templates and patterns
	execBootstrapGLSubmoduleTemplate  = "execBootstrapGLSubmoduleReqFile"
	execBootstrapGLSubmoduleFile      = "{{ index .Job.Start .Idx }}-{{ index .Job.End .Idx }}-etv{{ index .Job.Etv .Idx }}-stv{{ index .Job.Stv .Idx }}-getZkProof_Bootstrap_GLSubmodule.json"
	execBootstrapDistMetadataTemplate = "execBootstrapSubmoduleDistMetadataFile"
	execBootstrapDistMetadataFile     = "{{ index .Job.Start .Idx }}-{{ index .Job.End .Idx }}-etv{{ index .Job.Etv .Idx }}-stv{{ index .Job.Stv .Idx }}-getZkProof_Bootstrap_DistMetadata.json"
	execGLRndBeaconTemplate           = "execGLBeaconFile"
	execGLRndBeaconFile               = "{{ index .Job.Start .Idx }}-{{ index .Job.End .Idx }}-etv{{ index .Job.Etv .Idx }}-stv{{ index .Job.Stv .Idx }}-getZkProof_GL_RndBeacon.json"
	execGLTemplate                    = "execGLOutputFile"
	execGLFile                        = "{{ index .Job.Start .Idx }}-{{ index .Job.End .Idx }}-etv{{ index .Job.Etv .Idx }}-stv{{ index .Job.Stv .Idx }}-getZkProof_GL.json"
	execRndBeaconTemplate             = "execRndBeaconOutputFile"
	execRndBeaconFile                 = "{{ index .Job.Start .Idx }}-{{ index .Job.End .Idx }}-etv{{ index .Job.Etv .Idx }}-stv{{ index .Job.Stv .Idx }}-getZkProof_RndBeacon.json"
	execLPPTemplate                   = "execLPPOutputFile"
	execLPPFile                       = "{{ index .Job.Start .Idx }}-{{ index .Job.End .Idx }}-etv{{ index .Job.Etv .Idx }}-stv{{ index .Job.Stv .Idx }}-getZkProof_LPP.json"
	execConglomerateTemplate          = "execOutputFile"
	execConglomerateFile              = "{{ index .Job.Start .Idx }}-{{ index .Job.End .Idx }}-getZkProof.json"
)

func ExecBootstrapDefinition(conf *config.Config) (*JobDefinition, error) {
	inpFileExt := ""
	if conf.ExecBootstrap.CanRunFullLarge {
		inpFileExt = fmt.Sprintf(`\.%v`, config.LargeSuffix)
	}

	// Input files
	reqDirs := conf.ExecBootstrap.RequestsRootDir
	inputFilePatterns := []string{fmt.Sprintf(execBootstrapInputPattern, inpFileExt, config.FailSuffix)}

	// Output files
	outputTmpls := []string{execBootstrapGLSubmoduleTemplate, execBootstrapDistMetadataTemplate}
	outputFiles := []string{execBootstrapGLSubmoduleFile, execBootstrapDistMetadataFile}

	return commonJobDefinition(jobExecBootstrap, priorityExecBootstrap,
		reqDirs, inputFilePatterns, outputTmpls, outputFiles, cmnExecParamsRegexp(1), config.FailSuffix)
}

func ExecGLDefinition(conf *config.Config) (*JobDefinition, error) {
	inpFileExt := ""
	if conf.ExecGL.CanRunFullLarge {
		inpFileExt = fmt.Sprintf(`\.%v`, config.LargeSuffix)
	}

	// Input files
	reqDirs := conf.ExecGL.RequestsRootDir
	inputFilePatterns := []string{fmt.Sprintf(execBootstrapGLInputPattern, inpFileExt, config.FailSuffix)}

	// Output files
	outputTmpls := []string{execGLRndBeaconTemplate, execGLTemplate}
	outputFiles := []string{execGLRndBeaconFile, execGLFile}

	return commonJobDefinition(jobExecGL, priorityExecGL,
		reqDirs, inputFilePatterns, outputTmpls, outputFiles, cmnExecParamsRegexp(1), config.FailSuffix)
}

func ExecRndBeaconDefinition(conf *config.Config) (*JobDefinition, error) {
	inpFileExt := ""
	if conf.ExecRndBeacon.CanRunFullLarge {
		inpFileExt = fmt.Sprintf(`\.%v`, config.LargeSuffix)
	}

	// Input files
	reqDirs := utils.CombineRequests(conf.ExecRndBeacon.BootstrapMetadata.RequestsRootDir, conf.ExecRndBeacon.GL.RequestsRootDir)
	inputFilePatterns := []string{
		fmt.Sprintf(execBootstrapRndBeaconInputPattern, inpFileExt, config.FailSuffix),
		fmt.Sprintf(execGLRndBeaconInputPattern, inpFileExt, config.FailSuffix),
	}

	// Output files
	outputTmpls := []string{execRndBeaconTemplate}
	outputFiles := []string{execRndBeaconFile}

	return commonJobDefinition(jobExecRndBeacon, priorityExecRndBeacon,
		reqDirs, inputFilePatterns, outputTmpls, outputFiles, cmnExecParamsRegexp(2), config.FailSuffix)
}

func ExecLPPDefinition(conf *config.Config) (*JobDefinition, error) {
	inpFileExt := ""
	if conf.ExecLPP.CanRunFullLarge {
		inpFileExt = fmt.Sprintf(`\.%v`, config.LargeSuffix)
	}

	// Input files
	reqDirs := conf.ExecLPP.RequestsRootDir
	inputFilePatterns := []string{fmt.Sprintf(execLPPInputPattern, inpFileExt, config.FailSuffix)}

	// Output files
	outputTmpls := []string{execLPPTemplate}
	outputFiles := []string{execLPPFile}

	return commonJobDefinition(jobExecLPP, priorityExecLPP,
		reqDirs, inputFilePatterns, outputTmpls, outputFiles, cmnExecParamsRegexp(1), config.FailSuffix)
}

func ExecConglomerationDefinition(conf *config.Config) (*JobDefinition, error) {
	inpFileExt := ""
	if conf.ExecConglomeration.CanRunFullLarge {
		inpFileExt = fmt.Sprintf(`\.%v`, config.LargeSuffix)
	}

	// Input files
	reqDirs := utils.CombineRequests(conf.ExecConglomeration.BootstrapMetadata.RequestsRootDir, conf.ExecConglomeration.GL.RequestsRootDir, conf.ExecConglomeration.LPP.RequestsRootDir)
	inputFilePatterns := []string{
		fmt.Sprintf(execConglomerateBootstrapDistMetadataPattern, inpFileExt, config.FailSuffix),
		fmt.Sprintf(execConglomerateGLInputPattern, inpFileExt, config.FailSuffix),
		fmt.Sprintf(execConglomerateLPPInputPattern, inpFileExt, config.FailSuffix),
	}

	// Output files
	outputTmpls := []string{execConglomerateTemplate}
	outputFiles := []string{execConglomerateFile}

	return commonJobDefinition(jobExecCongolomerateLPP, priorityExecCongolomeration,
		reqDirs, inputFilePatterns, outputTmpls, outputFiles, cmnExecParamsRegexp(3), config.FailSuffix)
}
