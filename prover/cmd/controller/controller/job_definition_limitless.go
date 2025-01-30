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
	job_Exec_Bootstrap_Submodule = "exec-bootstrap-submodule"
	job_Exec_Bootstrap_Metadata  = "exec-bootstrap-metadata"

	// Global-Local subprovers
	job_Exec_GL_RndBeacon = "exec-GL-rndbeacon"
	job_Exec_GL           = "exec-GL"

	// Random Beacon
	job_Exec_RndBeacon_LPP      = "exec-rndbeacon-LPP"
	job_Exec_RndBeacon_Metadata = "exec-rndbeacon-metadata"

	// LPP-subprovers
	job_Exec_LPP = "exec-LPP"

	// Conglomerator
	job_Exec_Congolomerate_LPP      = "exec-congolo-LPP"
	job_Exec_Congolomerate_GL       = "exec-congolo-GL"
	job_Exec_Congolomerate_Metadata = "exec-congolo-metadata"
)

// Priorities
const (
	priority_Exec_Bootstrap_Submodule = 0
	priority_Exec_Bootstrap_Metadata  = 0

	priority_Exec_GL_RndBeacon = 1
	priority_Exec_GL           = 1

	priority_Exec_RndBeacon_LPP      = 2
	priority_Exec_RndBeacon_Metadata = 2

	priority_Exec_LPP = 3

	priority_Exec_Congolomerate_LPP      = 4
	priority_Exec_Congolomerate_GL       = 4
	priority_Exec_Congolomerate_Metadata = 4
)

// Input file patterns
const (
	exec_Bootstrap_Submodule_InputPattern = `^[0-9]+-[0-9]+(-etv[0-9\.]+)?(-stv[0-9\.]+)?-getZkProof\.json%v(\.failure\.%v_[0-9]+)*$`
	exec_Bootstrap_MetaData_InputPattern  = `^[0-9]+-[0-9]+(-etv[0-9\.]+)?(-stv[0-9\.]+)?-getZkProof\.json%v(\.failure\.%v_[0-9]+)*$`

	exec_GL_RndBeacon_InputPattern = `^[0-9]+-[0-9]+(-etv[0-9\.]+)?(-stv[0-9\.]+)?-getZKProof_Bootstrap_Submodule\.json%v(\.failure\.%v_[0-9]+)*$`

	exec_RndBeacon_Metadata_InputPattern = `^[0-9]+-[0-9]+(-etv[0-9\.]+)?(-stv[0-9\.]+)?-getZKProof_Bootstrap_DistMetadata\.json%v(\.failure\.%v_[0-9]+)*$`
	exec_GL_InputPattern                 = `^[0-9]+-[0-9]+(-etv[0-9\.]+)?(-stv[0-9\.]+)?-getZKProof_GL_Beacon\.json%v(\.failure\.%v_[0-9]+)*$`

	exec_LPP_InputPattern = `^[0-9]+-[0-9]+(-etv[0-9\.]+)?(-stv[0-9\.]+)?-getZKProof_RndBeacon\.json%v(\.failure\.%v_[0-9]+)*$`

	exec_Congolomerate_GL_InputPattern       = `^[0-9]+-[0-9]+(-etv[0-9\.]+)?(-stv[0-9\.]+)?-getZKProof_GL\.json%v(\.failure\.%v_[0-9]+)*$`
	exec_Congolomerate_LPP_InputPattern      = `^[0-9]+-[0-9]+(-etv[0-9\.]+)?(-stv[0-9\.]+)?-getZKProof_LPP\.json%v(\.failure\.%v_[0-9]+)*$`
	exec_Congolomerate_Metadata_InputPattern = `^[0-9]+-[0-9]+(-etv[0-9\.]+)?(-stv[0-9\.]+)?-getZKProof_Bootstrap_DistMetadata\.json%v(\.failure\.%v_[0-9]+)*$`
)

// Ouput File patterns and templates
const (
	exec_Bootstrap_Submodule_File = "{{.Start}}-{{.End}}-getZkProof_Bootstrap_Submodule.json"
	exec_Bootstrap_Submodule_Tmpl = "exec-bootstrap-submodule-req-file"

	exec_Bootstrap_DistMetadata_File = "{{.Start}}-{{.End}}-getZKProof_Bootstrap_DistMetadata.json"
	exec_Bootstrap_DistMetadata_Tmpl = "exec-bootstrap-submodule-distmetadata-file"

	// Global-Local subprovers
	exec_GL_Beacon_File = "{{.Start}}-{{.End}}-getZKProof_GL_Beacon.json"
	exec_GL_Beacon_Tmpl = "exec-GL-Beacon-file"

	exec_GL_File = "{{.Start}}-{{.End}}-getZKProof_GL.json"
	exec_GL_Tmpl = "exec-GL-output-file"

	// Random Beacon
	exec_RndBeacon_DistMetadata_File = "{{.Start}}-{{.End}}-getZKProof_Bootstrap_DistMetadata.json"

	exec_RndBeacon_GL_File = "{{.Start}}-{{.End}}-getZKProof_GL_Beacon.json"

	exec_RndBeacon_File = "{{.Start}}-{{.End}}-getZKProof_RndBeacon.json"
	exec_RndBeacon_Tmpl = "exec-rndbeacon-output-file"

	// LPP-subprovers
	exec_LPP_File = "{{.Start}}-{{.End}}-getZKProof_LPP.json"
	exec_LPP_Tmpl = "exec-LPP-output-file"

	// Conglomerator
	// exec_Congolomerate_GL_File = "{{.Start}}-{{.End}}-.getZKProof_GL.json"

	// exec_Congolomerate_LPP_File = "{{.Start}}-{{.End}}-.getZKProof_LPP.json"

	// exec_Congolomerate_Metadata_File = "{{.Start}}-{{.End}}-.getZKProof_Bootstrap_DistMetadata.json"

	exec_Congolomerate_File = "{{.Start}}-{{.End}}-.getZKProof.json"
	exec_Congolomerate_Tmpl = "exec-output-file"
)

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

func BootstrapSubModDefinition(conf *config.Config) (*JobDefinition, error) {
	inpFileExt := ""
	if conf.Bootstrap_Submodule.CanRunFullLarge {
		inpFileExt = fmt.Sprintf(`\.%v`, config.LargeSuffix)
	}
	inputFilePattern := fmt.Sprintf(exec_Bootstrap_Submodule_InputPattern, inpFileExt, config.FailSuffix)
	return createJobDefinition(job_Exec_Bootstrap_Submodule, priority_Exec_Bootstrap_Submodule,
		conf.Bootstrap_Submodule.RequestsRootDir, inputFilePattern, exec_Bootstrap_Submodule_Tmpl, exec_Bootstrap_Submodule_File)
}

func BootstrapMetadataDefinition(conf *config.Config) (*JobDefinition, error) {
	inpFileExt := ""
	if conf.Bootstrap_Metadata.CanRunFullLarge {
		inpFileExt = fmt.Sprintf(`\.%v`, config.LargeSuffix)
	}
	inputFilePattern := fmt.Sprintf(exec_Bootstrap_MetaData_InputPattern, inpFileExt, config.FailSuffix)
	return createJobDefinition(job_Exec_Bootstrap_Metadata, priority_Exec_Bootstrap_Metadata,
		conf.Bootstrap_Metadata.RequestsRootDir, inputFilePattern, exec_Bootstrap_DistMetadata_Tmpl, exec_Bootstrap_DistMetadata_File)
}
