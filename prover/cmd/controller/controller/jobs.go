package controller

import (
	"fmt"
	"path"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/consensys/linea-monorepo/prover/config"
	"github.com/consensys/linea-monorepo/prover/utils"
	"github.com/dlclark/regexp2"
	"github.com/sirupsen/logrus"
)

// Defines a job to be processed. Jobs are parsed from filenames using regexp.
type Job struct {
	// Configuration parameters relative to the job
	Def *JobDefinition

	// Original name of the file when it was found
	OriginalFile []string
	// Name of the locked file. If this value is set, it means that the job
	// was successfully locked.
	LockedFile []string

	// Height of the file in the priority queue
	Start []int
	End   []int

	// Execution Trace version
	Etv []string

	// State Manager Trace version
	Stv []string

	// Compressor version
	Cv []string

	// The hex string of the content hash
	ContentHash []string
}

// OutputFileRessouce collects all the data needed to fill the output template
// file.
type OutputFileResource struct {
	Job Job

	// TODO: Remove this attribute. Not required and change the regex patterns defined in the job definition
	Idx int
}

// Parse a filename into a Job. Returns an error if the file does not
// corresponds to the specified template of the job type.
func NewJob(jdef *JobDefinition, filenames []string) (j *Job, err error) {
	m, n := len(jdef.InputFileRegexp), len(filenames)
	if m != n {
		return nil, fmt.Errorf(`length mis-match between the number of input files specified in the 
				job definition: %d and the function params: %d`, m, n)
	}

	// Define job
	j = &Job{
		Def: jdef, OriginalFile: make([]string, m), LockedFile: make([]string, m),
		Start: make([]int, m), End: make([]int, m), Etv: make([]string, m), Stv: make([]string, m),
		Cv: make([]string, m), ContentHash: make([]string, m),
	}

	for i := 0; i < m; i++ {
		// Validate the filename against the inbound regexp
		if ok, err := jdef.InputFileRegexp[i].MatchString(filenames[i]); !ok || err != nil {
			return nil, fmt.Errorf(
				"filename %v does not match the inbound regexp for `%v` job: `%v. "+
					"Err if any = %v",
				filenames[i], jdef.Name, jdef.InputFileRegexp[i].String(), err,
			)
		}
		j.OriginalFile[i] = filenames[i]

		regs := jdef.ParamsRegexp

		// If the regexps in the job definition are provided, use them to extract
		// the parameters of the job. If one regexp is provided but is invalid, this
		// will panic. That is because since, we already matched the
		// `InputFileRegexp`, we assume that the parameters regexp can only be
		// matched.
		j.Start[i] = intIfRegexpNotNil(regs[i].Start, filenames[i])
		j.End[i] = intIfRegexpNotNil(regs[i].End, filenames[i])
		j.Cv[i] = stringIfRegexpNotNil(regs[i].Cv, filenames[i])
		j.Etv[i] = stringIfRegexpNotNil(regs[i].Etv, filenames[i])
		j.Stv[i] = stringIfRegexpNotNil(regs[i].Stv, filenames[i])
		j.ContentHash[i] = stringIfRegexpNotNil(regs[i].ContentHash, filenames[i])
	}

	return j, nil
}

// Returns the full path to the inprogress file
func (j *Job) InProgressPath(ipIdx int) string {
	if err := j.Def.isValidReqRootDirIdx(ipIdx); err != nil {
		utils.Panic("InProgressPath panic:%v", err.Error())
	}
	return filepath.Join(j.Def.dirFrom(ipIdx), j.LockedFile[ipIdx])
}

// func (j *Job) InProgressPath() []string {
// 	dirs := j.Def.dirFrom()
// 	inProgressPaths := make([]string, len(dirs))
// 	for ipIdx := 0; ipIdx < len(inProgressPaths); ipIdx++ {
// 		inProgressPaths[ipIdx] = filepath.Join(dirs[ipIdx], j.LockedFile[ipIdx])
// 	}
// 	return inProgressPaths
// }

// Returns the name of the output file for the job at the specified index
func (j *Job) ResponseFile(opIdx int) (s string, err error) {

	// Sanity check
	if err := j.Def.isValidOutputFileIdx(opIdx); err != nil {
		return "", err
	}

	// Run the template
	// REMARK: Check how it behaves on runtime
	w := &strings.Builder{}
	err = j.Def.OutputFileTmpl[opIdx].Execute(w, OutputFileResource{
		Job: *j,
	})
	if err != nil {
		return "", err
	}
	s = w.String()

	// Clear the `--` from the response file name. These can happen if an
	// optional field was not provided in the request file and is given in the
	// response.
	s = strings.ReplaceAll(s, "--", "-")

	// Append the dir_to filepath
	s = path.Join(j.Def.dirTo(opIdx), s)

	return s, nil
}

// Returns the name of the output file for the job
func (j *Job) TmpResponseFile(c *config.Config, opIdx int) (s string) {
	if err := j.Def.isValidOutputFileIdx(opIdx); err != nil {
		utils.Panic("TmpResponseFile panic:%v", err.Error())
	}
	return path.Join(j.Def.dirTo(opIdx), "tmp-response-file."+c.Controller.LocalID+".json")
}

// This function returns the name of the input file, modified to indicate that it should be retried in "large mode".
// It will fail if the job's configuration does not include a suffix for retrying in large mode.
// However, this situation is unexpected because the configuration validation ensures that if an exit code requires
// deferring the job to a larger machine, the suffix must be set.
// Additionally, if the prover's status code is zero (indicating success), the function will return an error.
func (j *Job) DeferToLargeFile(status Status, ipIdx int) (s string, err error) {

	// Sanity check
	if err := j.Def.isValidReqRootDirIdx(ipIdx); err != nil {
		return "", err
	}

	// It's an invariant of the executor to not forget to set the status
	if status.ExitCode == 0 {
		return "", fmt.Errorf(
			"cant defer to large %v, status code was zero",
			j.OriginalFile[ipIdx],
		)
	}

	const suffixLarge = config.LargeSuffix

	// Issue a warning if the files name already contains the
	// suffix. We may be in a situation where the large prover is trying to
	// defer over the same file. It is very likely an error. We will still
	// rename it to "<...>.large.large". That way, the file will not be picked
	// up a second time by the same large prover, creating an infinite retry
	// loop.
	if strings.HasSuffix(j.OriginalFile[ipIdx], suffixLarge) {
		logrus.Warnf(
			"Deferring the large machine but the input file `%v` already has"+
				" the suffix %v. Still renaming it to %v, but it will likely"+
				// Returns the name of the input file modified so that it is retried in		" not be picked up again",
				j.OriginalFile[ipIdx], suffixLarge, s,
		)
	}

	// Remove the suffix .failure.code_[0-9]+ from all the strings of the input
	// file. That way we do not propagate the previous errors.
	origFile, err := j.Def.FailureSuffix.Replace(j.OriginalFile[ipIdx], "", -1, -1)
	if err != nil {
		// he assumption here is that the above function may return an error
		// but this error can only depend on the regexp, the replacement,
		// the startAt and the count/ Thus, if it fails, the error is
		// unrelated to the input stream, which is the only user-provided
		// parameter.
		panic(err)
	}

	return fmt.Sprintf(
		"%v/%v.%v.failure.%v_%v",
		j.Def.dirFrom(ipIdx), origFile,
		suffixLarge,
		config.FailSuffix, status.ExitCode,
	), nil
}

// Returns the done file following the jobs status
func (j *Job) DoneFile(status Status, ipIdx int) string {

	// Sanity check
	if err := j.Def.isValidReqRootDirIdx(ipIdx); err != nil {
		utils.Panic("DoneFile panic:%v", err.Error())
	}

	// Remove the suffix .failure.code_[0-9]+ from all the strings
	origFile, err := j.Def.FailureSuffix.Replace(j.OriginalFile[ipIdx], "", -1, -1)
	if err != nil {
		// he assumption here is that the above function may return an error
		// but this error can only depend on the regexp, the replacement,
		// the startAt and the count/ Thus, if it fails, the error is
		// unrelated to the input stream, which is the only user-provided
		// parameter.
		panic(err)
	}

	if status.ExitCode == CodeSuccess {
		return fmt.Sprintf("%v/%v.%v", j.Def.dirDone(ipIdx), origFile, config.SuccessSuffix)
	} else {
		return fmt.Sprintf("%v/%v.failure.%v_%v", j.Def.dirDone(ipIdx), origFile, config.FailSuffix, status.ExitCode)
	}
}

// Returns the score of a JOB. The score is obtained as 100*job.Stop + P, where
// P is 0 if the job is an execution job, 1 if the job is a compression job and
// 2 if the job is an aggregation job. The lower the score the higher will be
// the priority of the job. The 100 value is chosen to make the score easy to
// mentally compute. ASSUMED 0 index here
func (j *Job) Score() int {
	return 100*j.End[0] + j.Def.Priority
}

// If the regexp is provided and non-nil, return the first match and returns the
// empty string if no match is found. If the regexp is not provided, returns the
// empty strings
func stringIfRegexpNotNil(r *regexp2.Regexp, s string) (res string) {
	if r != nil {
		res, err := r.FindStringMatch(s)
		if err != nil || res == nil {
			return ""
		}
		return res.String()
	}
	return ""
}

// Parse an integer from a string using a regexp if provided. If the regexp is
// non-nil but does not match the string s, the function panics. If the regexp
// is not provided, the function returns 0.
func intIfRegexpNotNil(r *regexp2.Regexp, s string) int {
	// Map the result as an integer
	match := stringIfRegexpNotNil(r, s)
	if len(s) == 0 {
		return 0
	}

	res, err := strconv.Atoi(match)
	if err != nil {
		// If this happens, it means that the provided regexp does not match
		// decimal digits strings.
		panic(err)
	}
	return res
}
