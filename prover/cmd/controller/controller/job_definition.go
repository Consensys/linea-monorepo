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
	jobNameExecution         = "execution"
	jobNameBlobDecompression = "compression"
	jobNameAggregation       = "aggregation"
)

// ParamsRegexp represents the associated compiled regexps for a job definition.
type ParamsRegexp struct {
	Start       *regexp2.Regexp
	End         *regexp2.Regexp
	Stv         *regexp2.Regexp
	Etv         *regexp2.Regexp
	Cv          *regexp2.Regexp
	ContentHash *regexp2.Regexp
}

// JobDefinition represents a collection of static parameters allowing to define a job.
type JobDefinition struct {
	// Name of the job
	Name string

	// Priority at which this type of job should be processed. The lower the more of a priority.
	// Typically 0 for execution, 1 for compression and 2 for aggregation.
	Priority int

	// Parameters for the job definition provided by the user
	RequestsRootDir []string

	// The regexp to use to match input files. For instance,
	//
	// 	`^\d+-\d+-etv0.1.2-stv\d.\d.\d-getZkProof.json$`
	//
	// Will tell the controller to accept any version of the state-manager
	// but to only accept execution trace. The regexp should always start "^"
	// and end with "$" otherwise you are going to match in-progress files.
	//
	InputFileRegexp []*regexp2.Regexp

	// Template to use to generate the output file. The template should have the
	// form of a go template. For instance,
	//
	// 	`{{.From}}-{{.To}}-pv{{.Version}}-stv{{.Stv}}-etv{{.Etv}}-zkProof.json`
	//
	OutputFileTmpl []*template.Template

	// The associated compiled regexp, this saves on recompiling the regexps
	// everytime we want to use them. If a field is not needed, it can be left
	// at zero.
	ParamsRegexp []ParamsRegexp

	// Regexp of the failure code so that we can trim it if we want to retry.
	FailureSuffix *regexp2.Regexp
}

// commonJobDefinition creates a new JobDefinition with the provided parameters.
// It sets up the job definition's name, priority, request directories, input file patterns, output templates,
// and parameter regexps. The function returns a JobDefinition and an error if any occurs during the setup.
func commonJobDefinition(name string, priority int,
	reqRootDirs []string, inputFilePatterns []string,
	outputFileTmpls []string, outputFileNames []string,
	paramsRegexp []ParamsRegexp, failSuffix string) (*JobDefinition, error) {

	m, n := len(reqRootDirs), len(inputFilePatterns)
	if m != n {
		return nil, fmt.Errorf(`length mis-match between the number of request files:%d 
		and input file patterns:%d specified in the job definition`, m, n)
	}

	p, q := len(outputFileTmpls), len(outputFileNames)
	if p != q {
		return nil, fmt.Errorf(`length mis-match between the number of output file templates:%d 
		and output file names:%d specified in the job definition`, p, q)
	}

	inputFileRegexps := make([]*regexp2.Regexp, m)
	paramsRegexps := make([]ParamsRegexp, m)
	outputFileTemplates := make([]*template.Template, p)

	for i := range inputFilePatterns {
		inputFileRegexps[i] = regexp2.MustCompile(inputFilePatterns[i], regexp2.None)
		paramsRegexps[i] = paramsRegexp[i]
	}

	for j := range outputFileNames {
		outputFileTemplates[j] = tmplMustCompile(outputFileTmpls[j], outputFileNames[j])
	}

	return &JobDefinition{
		Name:            name,
		Priority:        priority,
		RequestsRootDir: reqRootDirs,
		InputFileRegexp: inputFileRegexps,
		OutputFileTmpl:  outputFileTemplates,
		ParamsRegexp:    paramsRegexps,
		FailureSuffix:   matchFailureSuffix(failSuffix),
	}, nil
}

// ExecutionDefinition creates a job definition for the execution prover job.
// It sets the input file pattern based on the configuration and creates the job definition with the appropriate parameters.
// The function panics on any error since it is called at startup.
func ExecutionDefinition(conf *config.Config) JobDefinition {
	inpFileExt := ""
	if conf.Execution.CanRunFullLarge {
		inpFileExt = fmt.Sprintf(`\.%v`, config.LargeSuffix)
	}

	inputFilePattern := fmt.Sprintf(
		`^[0-9]+-[0-9]+(-etv[0-9\.]+)?(-stv[0-9\.]+)?-getZkProof\.json%v(\.failure\.%v_[0-9]+)*$`,
		inpFileExt,
		config.FailSuffix,
	)

	paramsRegexp := ParamsRegexp{
		Start: regexp2.MustCompile(`^[0-9]+`, regexp2.None),
		End:   regexp2.MustCompile(`(?<=^[0-9]+-)[0-9]+`, regexp2.None),
		Etv:   matchVersionWithPrefix("etv"),
		Stv:   matchVersionWithPrefix("stv"),
	}

	jobDef, err := commonJobDefinition(
		jobNameExecution,
		0,
		[]string{conf.Execution.RequestsRootDir},
		[]string{inputFilePattern},
		[]string{"exec-output-file"},
		[]string{"{{.Start}}-{{.End}}-getZkProof.json"},
		[]ParamsRegexp{paramsRegexp},
		config.FailSuffix,
	)
	if err != nil {
		utils.Panic("could not create job definition: %v", err)
	}
	return *jobDef
}

// CompressionDefinition creates a job definition for the blob decompression prover job.
// It sets the input file pattern based on the configuration and creates the job definition with the appropriate parameters.
// The function panics on any error since it is called at startup.
func CompressionDefinition(conf *config.Config) JobDefinition {
	inputFilePattern := fmt.Sprintf(
		`^[0-9]+-[0-9]+(-bcv[0-9\.]+)?(-ccv[0-9\.]+)?-((0x)?[0-9a-zA-Z]*-)?getZkBlobCompressionProof\.json(\.failure\.%v_[0-9]+)*$`,
		config.FailSuffix,
	)

	paramsRegexp := ParamsRegexp{
		Start:       regexp2.MustCompile(`^[0-9]+`, regexp2.None),
		End:         regexp2.MustCompile(`(?<=^[0-9]+-)[0-9]+`, regexp2.None),
		Cv:          matchVersionWithPrefix("cv"),
		ContentHash: regexp2.MustCompile(`(?<=ccv[0-9\.]+-)(0x)?[0-9a-zA-Z]+-(?=getZk)`, regexp2.None),
	}

	jobDef, err := commonJobDefinition(
		jobNameBlobDecompression,
		1,
		[]string{conf.BlobDecompression.RequestsRootDir},
		[]string{inputFilePattern},
		[]string{"compress-output-file"},
		[]string{"{{.Start}}-{{.End}}-{{.ContentHash}}getZkBlobCompressionProof.json"},
		[]ParamsRegexp{paramsRegexp},
		config.FailSuffix,
	)
	if err != nil {
		utils.Panic("could not create job definition: %v", err)
	}
	return *jobDef
}

// AggregatedDefinition creates a job definition for the aggregated prover job.
// It sets the input file pattern based on the configuration and creates the job definition with the appropriate parameters.
// The function panics on any error since it is called at startup.
func AggregatedDefinition(conf *config.Config) JobDefinition {
	inputFilePattern := fmt.Sprintf(
		`^[0-9]+-[0-9]+(-[a-fA-F0-9]+)?-getZkAggregatedProof\.json(\.failure\.%v_[0-9]+)*$`,
		config.FailSuffix,
	)

	paramsRegexp := ParamsRegexp{
		Start:       regexp2.MustCompile(`^[0-9]+`, regexp2.None),
		End:         regexp2.MustCompile(`(?<=^[0-9]+-)[0-9]+`, regexp2.None),
		ContentHash: regexp2.MustCompile(`(?<=^[0-9]+-[0-9]+-)[a-fA-F0-9]+(?=-getZk)`, regexp2.None),
	}

	jobDef, err := commonJobDefinition(
		jobNameAggregation,
		2,
		[]string{conf.Aggregation.RequestsRootDir},
		[]string{inputFilePattern},
		[]string{"agreg-output-file"},
		[]string{"{{.Start}}-{{.End}}-{{.ContentHash}}-getZkAggregatedProof.json"},
		[]ParamsRegexp{paramsRegexp},
		config.FailSuffix,
	)
	if err != nil {
		utils.Panic("could not create job definition: %v", err)
	}
	return *jobDef
}

// Version prefix template
func matchVersionWithPrefix(pre string) *regexp2.Regexp {
	return regexp2.MustCompile(
		fmt.Sprintf(`(?<=%v)[\d\.]+`, pre),
		regexp2.None,
	)
}

// Match the failure code suffix. This string will essentially match all the
// substrings of the form `.failure.code_<X>` so that they can be replaced with
// the empty string.
func matchFailureSuffix(pre string) *regexp2.Regexp {
	return regexp2.MustCompile(
		fmt.Sprintf(`\.failure\.%v_[0-9]+`, pre),
		regexp2.None,
	)
}

// Generates a template or panics
func tmplMustCompile(name, tmpl string) *template.Template {
	res, err := template.New(name).Parse(tmpl)
	if err != nil {
		utils.Panic("could not generate template: %v", err)
	}
	return res
}

func (jd *JobDefinition) isValidReqRootDirIdx(idx int) error {
	if idx < 0 || idx >= len(jd.RequestsRootDir) {
		return fmt.Errorf("out-of-bound request root dir. index specified for job definition: %s", jd.Name)
	}
	return nil
}

func (jd *JobDefinition) isValidOutputFileIdx(idx int) error {
	if idx < 0 || idx >= len(jd.OutputFileTmpl) {
		return fmt.Errorf("out-of-bound output file template index specified for job definition: %s", jd.Name)
	}
	return nil
}

func (jd *JobDefinition) dirFrom(idx int) string {
	if err := jd.isValidReqRootDirIdx(idx); err != nil {
		utils.Panic(err.Error())
	}
	return filepath.Join(jd.RequestsRootDir[idx], config.RequestsFromSubDir)
}

func (jd *JobDefinition) dirDone(idx int) string {
	if err := jd.isValidReqRootDirIdx(idx); err != nil {
		utils.Panic(err.Error())
	}
	return filepath.Join(jd.RequestsRootDir[idx], config.RequestsDoneSubDir)
}

func (jd *JobDefinition) dirTo(idx int) string {
	if err := jd.isValidReqRootDirIdx(idx); err != nil {
		utils.Panic(err.Error())
	}
	return filepath.Join(jd.RequestsRootDir[idx], config.RequestsToSubDir)
}
