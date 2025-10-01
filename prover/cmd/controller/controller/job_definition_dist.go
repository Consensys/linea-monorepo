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

// GLDefinitionForModule returns a JobDefinition for a single GL module directory.
// The RequestsRootDir will be <conf.Limitless.WitnessDir>/GL/<module>.
func GLDefinitionForModule(conf *config.Config, module string) JobDefinition {

	var (
		inputPattern = fmt.Sprintf(
			`^[0-9]+-[0-9]+-seg-[0-9]+-mod-[0-9]+-gl-wit\.bin(\.failure\.%v_[0-9]+)*$`,
			config.FailSuffix,
		)
		outputTmpl = "{{.Start}}-{{.End}}-seg-{{.Seg}}-mod-{{.Mod}}-gl-wit.bin"

		rootDir = filepath.Join(conf.Limitless.WitnessDir, "GL", module)
	)

	return newJobDefinition(
		rootDir,
		jobNameGL,
		inputPattern,
		outputTmpl,
		0, // priority set to default value now and will be overwritten later when the file arrives
		ParamRegexps{
			Start: regexp2.MustCompile(`^[0-9]+`, regexp2.None),
			End:   regexp2.MustCompile(`(?<=^[0-9]+-)[0-9]+`, regexp2.None),
			Seg:   regexp2.MustCompile(`(?<=-seg-)[0-9]+`, regexp2.None),
			Mod:   regexp2.MustCompile(`(?<=-mod-)[0-9]+`, regexp2.None),
		},
	)
}

// LPPDefinitionForModule returns a JobDefinition for a single LPP module directory.
// The RequestsRootDir will be <conf.Limitless.WitnessDir>/LPP/<module>.
func LPPDefinitionForModule(conf *config.Config, module string) JobDefinition {

	var (
		inputPattern = fmt.Sprintf(
			`^[0-9]+-[0-9]+-seg-[0-9]+-mod-[0-9]+-lpp-wit\.bin(\.failure\.%v_[0-9]+)*$`,
			config.FailSuffix,
		)

		outputTmpl = "{{.Start}}-{{.End}}-seg-{{.Seg}}-mod-{{.Mod}}-lpp-wit.bin"

		rootDir = filepath.Join(conf.Limitless.WitnessDir, "LPP", module)
	)

	return newJobDefinition(
		rootDir,
		jobNameLPP,
		inputPattern,
		outputTmpl,
		0, //  priority set to default value now and will be overwritten later when the file arrives
		ParamRegexps{
			Start: regexp2.MustCompile(`^[0-9]+`, regexp2.None),
			End:   regexp2.MustCompile(`(?<=^[0-9]+-)[0-9]+`, regexp2.None),
			Seg:   regexp2.MustCompile(`(?<=-seg-)[0-9]+`, regexp2.None),
			Mod:   regexp2.MustCompile(`(?<=-mod-)[0-9]+`, regexp2.None),
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
	return 1 // fallback default
}

// setPriority sets the priority of the job based on the file name once it arrives
func (jd *JobDefinition) setPriority(fileName string) {
	jd.Priority = segPriority(fileName)
}
