package controller

import (
	"fmt"
	"path/filepath"
	"text/template"

	"github.com/consensys/linea-monorepo/prover/config"
	"github.com/consensys/linea-monorepo/prover/utils"
	"github.com/dlclark/regexp2"
)

const (
	jobNameExecution        = "execution"
	jobNameDataAvailability = "compression"
	jobNameAggregation      = "aggregation"
)

// ParamRegexps groups together common regexps used by JobDefinitions
type ParamRegexps struct {
	Start, End, Stv, Etv, Cv, ContentHash, SegID, ModID *regexp2.Regexp
}

// JobDefinition represents a collection of static parameters allowing to define
// a job.
type JobDefinition struct {
	// Name of the job
	Name string

	// Parameters for the job definition provided by the user
	RequestRootDir string

	// The regexp to use to match input files.
	InputFileRegexp *regexp2.Regexp

	// Response root directory. If it is not set, it will be the value returned by .dirTo
	// For example: Its not set for execution, compression, or aggregation jobs
	// It is set explicitly only for the limitless prover jobs
	ResponseRootDir string

	// Template to use to generate the output file. May be empty for jobs that
	// have no output (e.g., GL/LPP that use /dev/null).
	OutputFileTmpl *template.Template

	// Priority at which this type of job should be processed. The lower the
	// more of a priority.
	Priority int

	// Compiled regexps for capturing parameters.
	ParamsRegexp ParamRegexps

	// Regexp of the failure code so that we can trim it if we want to retry.
	FailureSuffix *regexp2.Regexp
}

// -------------------- Job Definitions --------------------

// ExecutionDefinition defines an execution prover job.
func ExecutionDefinition(conf *config.Config) JobDefinition {
	inpFileExt := ""
	if conf.Execution.CanRunFullLarge {
		inpFileExt = fmt.Sprintf(`\.%v`, config.LargeSuffix)
	}

	inputPattern := fmt.Sprintf(
		`^[0-9]+-[0-9]+(-etv[0-9\.]+)?(-stv[0-9\.]+)?-getZkProof\.json%v(\.failure\.%v_[0-9]+)*$`,
		inpFileExt,
		config.FailSuffix,
	)

	return newJobDefinition(jobNameExecution,
		conf.Execution.RequestsRootDir, inputPattern,
		"", "{{.Start}}-{{.End}}-getZkProof.json",
		0, paramsExecution(),
	)
}

// CompressionDefinition defines a compression prover job.
func CompressionDefinition(conf *config.Config) JobDefinition {
	inputPattern := fmt.Sprintf(
		`^[0-9]+-[0-9]+(-bcv[0-9\.]+)?(-ccv[0-9\.]+)?-((0x)?[0-9a-zA-Z]*-)?getZkBlobCompressionProof\.json(\.failure\.%v_[0-9]+)*$`,
		config.FailSuffix,
	)

	return newJobDefinition(jobNameDataAvailability,
		conf.DataAvailability.RequestsRootDir, inputPattern,
		"", "{{.Start}}-{{.End}}-{{.ContentHash}}getZkBlobCompressionProof.json",
		1, paramsCompression(),
	)
}

// AggregatedDefinition defines an aggregation prover job.
func AggregatedDefinition(conf *config.Config) JobDefinition {
	inputPattern := fmt.Sprintf(
		`^[0-9]+-[0-9]+(-[a-fA-F0-9]+)?-getZkAggregatedProof\.json(\.failure\.%v_[0-9]+)*$`,
		config.FailSuffix,
	)

	return newJobDefinition(jobNameAggregation,
		conf.Aggregation.RequestsRootDir, inputPattern,
		"", "{{.Start}}-{{.End}}-{{.ContentHash}}-getZkAggregatedProof.json",
		2,
		paramsAggregation(),
	)
}

// -------------------- Param Regex Helpers --------------------

var (
	reStart = regexp2.MustCompile(`^[0-9]+`, regexp2.None)
	reEnd   = regexp2.MustCompile(`(?<=^[0-9]+-)[0-9]+`, regexp2.None)
)

func paramsExecution() ParamRegexps {
	return ParamRegexps{
		Start: reStart,
		End:   reEnd,
		Etv:   matchVersionWithPrefix("etv"),
		Stv:   matchVersionWithPrefix("stv"),
	}
}

func paramsCompression() ParamRegexps {
	return ParamRegexps{
		Start: reStart,
		End:   reEnd,
		Cv:    matchVersionWithPrefix("cv"),
		ContentHash: regexp2.MustCompile(
			`(?<=ccv[0-9\.]+-)(0x)?[0-9a-zA-Z]+-(?=getZk)`,
			regexp2.None,
		),
	}
}

func paramsAggregation() ParamRegexps {
	return ParamRegexps{
		Start: reStart,
		End:   reEnd,
		ContentHash: regexp2.MustCompile(
			`(?<=^[0-9]+-[0-9]+-)[a-fA-F0-9]+(?=-getZk)`,
			regexp2.None,
		),
	}
}

// -------------------- Utilities --------------------

// Create a new JobDefinition with consistent defaults
func newJobDefinition(
	name, reqRootDir, inputPattern,
	respRootDir, outputTmpl string,
	priority int, params ParamRegexps,
) JobDefinition {
	return JobDefinition{
		Name:            name,
		RequestRootDir:  reqRootDir,
		InputFileRegexp: regexp2.MustCompile(inputPattern, regexp2.None),
		ResponseRootDir: respRootDir,
		OutputFileTmpl:  tmplMustCompile(name+"-output-file", outputTmpl),
		Priority:        priority,
		ParamsRegexp:    params,
		FailureSuffix:   matchFailureSuffix(config.FailSuffix),
	}
}

func matchVersionWithPrefix(pre string) *regexp2.Regexp {
	return regexp2.MustCompile(fmt.Sprintf(`(?<=%v)[\d\.]+`, pre), regexp2.None)
}

func matchFailureSuffix(pre string) *regexp2.Regexp {
	return regexp2.MustCompile(fmt.Sprintf(`\.failure\.%v_[0-9]+`, pre), regexp2.None)
}

func tmplMustCompile(name, tmpl string) *template.Template {
	res, err := template.New(name).Parse(tmpl)
	if err != nil {
		utils.Panic("could not generate template: %v", err)
	}
	return res
}

// -------------------- Directory Helpers --------------------

func (jd *JobDefinition) dirFrom() string {
	return filepath.Join(jd.RequestRootDir, config.RequestsFromSubDir)
}

func (jd *JobDefinition) dirDone() string {
	return filepath.Join(jd.RequestRootDir, config.RequestsDoneSubDir)
}

func (jd *JobDefinition) dirTo() string {

	if jd.ResponseRootDir != "" {
		return jd.ResponseRootDir
	}

	return filepath.Join(jd.RequestRootDir, config.RequestsToSubDir)
}

// WritesToDevNull reports whether this job definition discards responses.
func (jd *JobDefinition) WritesToDevNull() bool {
	return jd.ResponseRootDir == "/dev/null"
}
