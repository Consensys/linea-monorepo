package controller

import (
	"fmt"

	"github.com/consensys/linea-monorepo/prover/config"
	"github.com/dlclark/regexp2"
)

// Job definitions are defined such that each job has a single request and response file
// These jobs will execute asynchronously based on their set priorities
const (
	// Bootstrap
	job_Exec_Bootstrap_GLSubmodule  = "exec-bootstrap-GLsubmodule"
	job_Exec_Bootstrap_DistMetadata = "exec-bootstrap-metadata"

	// Global-Local subprovers
	job_Exec_GL_RndBeacon = "exec-GL-rndbeacon"
	job_Exec_GL           = "exec-GL"

	// Random Beacon
	job_Exec_RndBeacon_LPP       = "exec-rndbeacon"
	job_Exec_Bootstrap_RndBeacon = "exec-bootstrap-rndbeacon"

	// LPP-subprovers
	job_Exec_LPP = "exec-LPP"

	// Conglomerator
	job_Exec_Congolomerate_LPP      = "exec-congolo-LPP"
	job_Exec_Congolomerate_GL       = "exec-congolo-GL"
	job_Exec_Congolomerate_Metadata = "exec-congolo-metadata"
)

// Priorities
const (
	priority_Exec_Bootstrap_GLSubmodule  = 0
	priority_Exec_Bootstrap_DistMetadata = 0

	priority_Exec_GL_RndBeacon = 1
	priority_Exec_GL           = 1

	priority_Exec_RndBeacon_LPP       = 2
	priority_Exec_Bootstrap_RndBeacon = 2

	priority_Exec_LPP = 3

	priority_Exec_Congolomerate_LPP      = 4
	priority_Exec_Congolomerate_GL       = 4
	priority_Exec_Congolomerate_Metadata = 4
)

// Input file patterns
const (
	// Bootstrap I/p file is the usual execution req. file
	exec_Bootstrap_InputPattern = `^[0-9]+-[0-9]+(-etv[0-9\.]+)?(-stv[0-9\.]+)?-getZkProof\.json%v(\.failure\.%v_[0-9]+)*$`

	// GL input
	exec_Bootstrap_GL_InputPattern = `^[0-9]+-[0-9]+(-etv[0-9\.]+)?(-stv[0-9\.]+)?-getZkProof_Bootstrap_GLSubmodule\.json%v(\.failure\.%v_[0-9]+)*$`

	// Rnd Beacon I/p
	exec_Bootstrap_RndBeacon_InputPattern = `^[0-9]+-[0-9]+(-etv[0-9\.]+)?(-stv[0-9\.]+)?-getZkProof_Bootstrap_DistMetadata\.json%v(\.failure\.%v_[0-9]+)*$`
	exec_GL_RndBeacon_InputPattern        = `^[0-9]+-[0-9]+(-etv[0-9\.]+)?(-stv[0-9\.]+)?-getZkProof_GL_RndBeacon\.json%v(\.failure\.%v_[0-9]+)*$`

	// LPP Input
	exec_LPP_InputPattern = `^[0-9]+-[0-9]+(-etv[0-9\.]+)?(-stv[0-9\.]+)?-getZkProof_RndBeacon\.json%v(\.failure\.%v_[0-9]+)*$`

	// Conglomerator Input
	exec_Conglomerate_GL_InputPattern           = `^[0-9]+-[0-9]+(-etv[0-9\.]+)?(-stv[0-9\.]+)?-getZkProof_GL\.json%v(\.failure\.%v_[0-9]+)*$`
	exec_Conglomerate_LPP_InputPattern          = `^[0-9]+-[0-9]+(-etv[0-9\.]+)?(-stv[0-9\.]+)?-getZkProof_LPP\.json%v(\.failure\.%v_[0-9]+)*$`
	exec_Conglomerate_DistMetadata_InputPattern = `^[0-9]+-[0-9]+(-etv[0-9\.]+)?(-stv[0-9\.]+)?-getZkProof_Bootstrap_DistMetadata\.json%v(\.failure\.%v_[0-9]+)*$`
)

// Ouput File patterns and templates
const (
	exec_Bootstrap_GLSubmodule_File = "{{.Start}}-{{.End}}-etv{{.Etv}}-stv{{.Stv}}-getZkProof_Bootstrap_GLSubmodule.json"
	exec_Bootstrap_Submodule_Tmpl   = "exec-bootstrap-GLsubmodule-req-file"

	exec_Bootstrap_DistMetadata_File = "{{.Start}}-{{.End}}-etv{{.Etv}}-stv{{.Stv}}-getZkProof_Bootstrap_DistMetadata.json"
	exec_Bootstrap_DistMetadata_Tmpl = "exec-bootstrap-submodule-distmetadata-file"

	// Global-Local subprovers
	exec_GL_RndBeacon_File = "{{.Start}}-{{.End}}-etv{{.Etv}}-stv{{.Stv}}-getZkProof_GL_RndBeacon.json"
	exec_GL_RndBeacon_Tmpl = "exec-GL-Beacon-file"

	exec_GL_File = "{{.Start}}-{{.End}}-etv{{.Etv}}-stv{{.Stv}}-getZkProof_GL.json"
	exec_GL_Tmpl = "exec-GL-output-file"

	// Random Beacon
	exec_RndBeacon_File = "{{.Start}}-{{.End}}-etv{{.Etv}}-stv{{.Stv}}-getZkProof_RndBeacon.json"
	exec_RndBeacon_Tmpl = "exec-rndbeacon-output-file"

	// LPP-subprovers
	exec_LPP_File = "{{.Start}}-{{.End}}-etv{{.Etv}}-stv{{.Stv}}-getZkProof_LPP.json"
	exec_LPP_Tmpl = "exec-LPP-output-file"

	// Conglomerator
	exec_Congolomerate_File = "{{.Start}}-{{.End}}-.getZkProof.json"
	exec_Congolomerate_Tmpl = "exec-output-file"
)

// createJobDefinition creates a new JobDefinition with the provided parameters.
// It sets up the job's name, priority, request directory, input file pattern, and output template.
// The function returns a pointer to the JobDefinition and an error if any occurs during the setup.
func createJobDefinition(name string, priority int,
	reqRootDir, inputFilePattern string,
	outputTmpl, outputFileName string) (*JobDefinition, error) {

	return &JobDefinition{
		Name:     name,
		Priority: priority,

		// Primary and Secondary Request (Input) Files
		RequestsRootDir: reqRootDir,
		InputFileRegexp: regexp2.MustCompile(inputFilePattern, regexp2.None),

		// Output Templates
		OutputFileTmpl: tmplMustCompile(outputTmpl, outputFileName),

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
	}, nil
}

// BootstrapGLSubModDefinition creates a job definition for the Bootstrap GL Submodule job.
// It sets the input file pattern based on the configuration and creates the job definition
// with the appropriate parameters.
func BootstrapGLSubModDefinition(conf *config.Config) (*JobDefinition, error) {
	inpFileExt := ""
	if conf.Bootstrap.CanRunFullLarge {
		inpFileExt = fmt.Sprintf(`\.%v`, config.LargeSuffix)
	}
	inputFilePattern := fmt.Sprintf(exec_Bootstrap_InputPattern, inpFileExt, config.FailSuffix)
	return createJobDefinition(job_Exec_Bootstrap_GLSubmodule, priority_Exec_Bootstrap_GLSubmodule,
		conf.Bootstrap.RequestsRootDir, inputFilePattern, exec_Bootstrap_Submodule_Tmpl, exec_Bootstrap_GLSubmodule_File)
}

// BootstrapDistMetadataDefinition creates a job definition for the Bootstrap Metadata job.
// It sets the input file pattern based on the configuration and creates the job definition
// with the appropriate parameters.
func BootstrapDistMetadataDefinition(conf *config.Config) (*JobDefinition, error) {
	inpFileExt := ""
	if conf.Bootstrap.CanRunFullLarge {
		inpFileExt = fmt.Sprintf(`\.%v`, config.LargeSuffix)
	}
	inputFilePattern := fmt.Sprintf(exec_Bootstrap_InputPattern, inpFileExt, config.FailSuffix)
	return createJobDefinition(job_Exec_Bootstrap_DistMetadata, priority_Exec_Bootstrap_DistMetadata,
		conf.Bootstrap.RequestsRootDir, inputFilePattern, exec_Bootstrap_DistMetadata_Tmpl, exec_Bootstrap_DistMetadata_File)
}

func GLRndBeaconDefinition(conf *config.Config) (*JobDefinition, error) {
	inpFileExt := ""
	if conf.GLExecution.CanRunFullLarge {
		inpFileExt = fmt.Sprintf(`\.%v`, config.LargeSuffix)
	}
	inputFilePattern := fmt.Sprintf(exec_Bootstrap_GL_InputPattern, inpFileExt, config.FailSuffix)
	return createJobDefinition(job_Exec_GL_RndBeacon, priority_Exec_GL_RndBeacon,
		conf.GLExecution.RequestsRootDir, inputFilePattern, exec_GL_RndBeacon_Tmpl, exec_GL_RndBeacon_File)
}

func GLDefinition(conf *config.Config) (*JobDefinition, error) {
	inpFileExt := ""
	if conf.GLExecution.CanRunFullLarge {
		inpFileExt = fmt.Sprintf(`\.%v`, config.LargeSuffix)
	}
	inputFilePattern := fmt.Sprintf(exec_Bootstrap_GL_InputPattern, inpFileExt, config.FailSuffix)
	return createJobDefinition(job_Exec_GL, priority_Exec_GL,
		conf.GLExecution.RequestsRootDir, inputFilePattern, exec_GL_Tmpl, exec_GL_File)
}

func BootstrapRndBeaconDefinition(conf *config.Config) (*JobDefinition, error) {
	inpFileExt := ""
	if conf.RndBeacon.CanRunFullLarge {
		inpFileExt = fmt.Sprintf(`\.%v`, config.LargeSuffix)
	}
	inputFilePattern := fmt.Sprintf(exec_Bootstrap_RndBeacon_InputPattern, inpFileExt, config.FailSuffix)
	return createJobDefinition(job_Exec_Bootstrap_RndBeacon, priority_Exec_Bootstrap_RndBeacon,
		conf.RndBeacon.MetaData.RequestsRootDir, inputFilePattern, exec_RndBeacon_Tmpl, exec_RndBeacon_File)
}

func RndBeaconLPPDefinition(conf *config.Config) (*JobDefinition, error) {
	inpFileExt := ""
	if conf.RndBeacon.CanRunFullLarge {
		inpFileExt = fmt.Sprintf(`\.%v`, config.LargeSuffix)
	}
	inputFilePattern := fmt.Sprintf(exec_GL_RndBeacon_InputPattern, inpFileExt, config.FailSuffix)
	return createJobDefinition(job_Exec_RndBeacon_LPP, priority_Exec_RndBeacon_LPP,
		conf.RndBeacon.GL.RequestsRootDir, inputFilePattern, exec_RndBeacon_Tmpl, exec_RndBeacon_File)
}

func LPPDefinition(conf *config.Config) (*JobDefinition, error) {
	inpFileExt := ""
	if conf.LPPExecution.CanRunFullLarge {
		inpFileExt = fmt.Sprintf(`\.%v`, config.LargeSuffix)
	}
	inputFilePattern := fmt.Sprintf(exec_LPP_InputPattern, inpFileExt, config.FailSuffix)
	return createJobDefinition(job_Exec_LPP, priority_Exec_LPP,
		conf.LPPExecution.RequestsRootDir, inputFilePattern, exec_LPP_Tmpl, exec_LPP_File)
}
