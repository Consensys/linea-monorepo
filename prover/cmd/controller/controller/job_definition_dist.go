package controller

import (
	"fmt"
	"path/filepath"
	"regexp"
	"strconv"

	"github.com/consensys/linea-monorepo/prover/config"
	"github.com/dlclark/regexp2"
)

const (
	jobNameBootstrap      = "bootstrap"
	jobNameConglomeration = "conglomeration"
	jobNameGL             = "gl"
	jobNameLPP            = "lpp"
)

// -------------------- Job Definitions --------------------

// BootstrapDefinition defines a bootstrap job.
func BootstrapDefinition(conf *config.Config) JobDefinition {

	var (
		inputPattern = fmt.Sprintf(
			`^[0-9]+-[0-9]+(-etv[0-9\.]+)?(-stv[0-9\.]+)?-getZkProof\.json(\.failure\.%v_[0-9]+)*$`,
			config.FailSuffix,
		)

		// /tmp/metadata/requests
		responseRootDir = filepath.Join(conf.ExecutionLimitless.MetadataDir, config.RequestsFromSubDir)

		outputTmpl = "{{.Start}}-{{.End}}-metadata-getZkProof.json"
	)

	return newJobDefinition(jobNameBootstrap,
		conf.Execution.RequestsRootDir, inputPattern,
		responseRootDir, outputTmpl,
		0, paramsExecution(), // reuse Execution params (start, end, etv, stv)
	)
}

// ConglomerationDefinition defines a conglomeration job.
func ConglomerationDefinition(conf *config.Config) JobDefinition {

	var (
		inputPattern = fmt.Sprintf(
			`^[0-9]+-[0-9]+-metadata-getZkProof\.json(\.failure\.%v_[0-9]+)*$`,
			config.FailSuffix,
		)

		responseRootDir = filepath.Join(conf.Execution.RequestsRootDir, config.RequestsToSubDir)
		outputTmpl      = "{{.Start}}-{{.End}}-getZkProof.json"
	)

	return newJobDefinition(jobNameConglomeration,
		conf.ExecutionLimitless.MetadataDir, inputPattern,
		responseRootDir, outputTmpl,
		2, ParamRegexps{
			Start: reStart,
			End:   reEnd,
		},
	)
}

// GLDefinitionForModule returns a JobDefinition for a single GL module directory.
// The RequestsRootDir will be <conf.Limitless.WitnessDir>/GL/<module>.
func GLDefinitionForModule(conf *config.Config, module string) JobDefinition {

	var (
		jobName = fmt.Sprintf("%s-%s", jobNameGL, module)

		reqRootDir   = filepath.Join(conf.ExecutionLimitless.WitnessDir, "GL", module)
		inputPattern = fmt.Sprintf(
			`^[0-9]+-[0-9]+-seg-[0-9]+-mod-[0-9]+-gl-wit\.bin(\.failure\.%v_[0-9]+)*$`,
			config.FailSuffix,
		)

		// GL jobs don't produce an output artifact, only /dev/null.
		// So outputTmpl is set empty.
		respRootDir = "/dev/null"
		outputTmpl  = ""
	)

	return newJobDefinition(jobName,
		reqRootDir, inputPattern,
		respRootDir, outputTmpl,
		1, // priority set to default value now and will be overwritten later when the file arrives
		ParamRegexps{
			Start: reStart,
			End:   reEnd,
			SegID: regexp2.MustCompile(`(?<=-seg-)[0-9]+`, regexp2.None),
			ModID: regexp2.MustCompile(`(?<=-mod-)[0-9]+`, regexp2.None),
		},
	)
}

// LPPDefinitionForModule returns a JobDefinition for a single LPP module directory.
// The RequestsRootDir will be <conf.Limitless.WitnessDir>/LPP/<module>.
func LPPDefinitionForModule(conf *config.Config, module string) JobDefinition {

	var (
		jobName = fmt.Sprintf("%s-%s", jobNameLPP, module)

		reqRootDir   = filepath.Join(conf.ExecutionLimitless.WitnessDir, "LPP", module)
		inputPattern = fmt.Sprintf(
			`^[0-9]+-[0-9]+-seg-[0-9]+-mod-[0-9]+-lpp-wit\.bin(\.failure\.%v_[0-9]+)*$`,
			config.FailSuffix,
		)

		// LPP jobs don't produce an output artifact, only /dev/null.
		// So outputTmpl is set empty.
		respRootDir = "/dev/null"
		outputTmpl  = ""
	)

	return newJobDefinition(jobName,
		reqRootDir, inputPattern,
		respRootDir, outputTmpl,
		3, //  priority set to default value now and will be overwritten later when the file arrives
		ParamRegexps{
			Start: reStart,
			End:   reEnd,
			SegID: regexp2.MustCompile(`(?<=-seg-)[0-9]+`, regexp2.None),
			ModID: regexp2.MustCompile(`(?<=-mod-)[0-9]+`, regexp2.None),
		},
	)
}

// -------------------- Utilities --------------------

var segRegexp = regexp.MustCompile(`-seg-(\d+)-`)

func segPriority(filename string) int {
	matches := segRegexp.FindStringSubmatch(filepath.Base(filename))
	if len(matches) > 1 {
		if val, err := strconv.Atoi(matches[1]); err == nil {
			return val
		}
	}
	panic("failed to parse segment ID") // fallback default
}
