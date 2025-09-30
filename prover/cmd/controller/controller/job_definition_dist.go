package controller

import (
	"fmt"

	"github.com/consensys/linea-monorepo/prover/config"
)

const (
	jobNameBootstrap      = "bootstrap"
	jobNameGL             = "gl"
	jobNameLPP            = "lpp"
	jobNameConglomeration = "conglomeration"
)

// -------------------- Job Definitions --------------------

// BootstrapDefinition defines a bootstrap job.
func BootstrapDefinition(conf *config.Config) JobDefinition {
	inputPattern := fmt.Sprintf(
		`^[0-9]+-[0-9]+(-etv[0-9\.]+)?(-stv[0-9\.]+)?-getZkProof\.json(\.failure\.%v_[0-9]+)*$`,
		config.FailSuffix,
	)

	outputTmpl := "{{.Start}}-{{.End}}-metadata-getZkProof.json"

	return newJobDefinition(
		conf.Execution.RequestsRootDir,
		jobNameBootstrap,
		inputPattern,
		outputTmpl,
		0,
		paramsExecution(), // reuse Execution params (start, end, etv, stv)
	)
}
