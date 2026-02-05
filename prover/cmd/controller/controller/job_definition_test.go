package controller

import (
	"fmt"
	"path/filepath"
	"testing"

	"github.com/consensys/linea-monorepo/prover/config"
	"github.com/stretchr/testify/assert"
)

type inpFileNamesCases struct {
	Ext, Fail      string
	Fnames         []string
	ShouldMatch    bool
	Explainer      string
	ExpectedOutput []string
	ExpToLarge     []string
	ExpSuccess     []string
	ExpFailW2      []string
}

// This tests ensures that the naming convention is respected by the file-watcher
// i.e., files with the right naming only are recognized. And the corresponding
// output files are also recognized.
func TestExecutionInFileRegexp(t *testing.T) {

	var (
		correctM           = "102-103-etv0.2.3-stv1.2.3-getZkProof.json"
		correctL           = "102-103-etv0.2.3-stv1.2.3-getZkProof.json.large"
		correctWithFailM   = "102-103-etv0.2.3-stv1.2.3-getZkProof.json.failure.code_77"
		correctWithFailL   = "102-103-etv0.2.3-stv1.2.3-getZkProof.json.large.failure.code_77"
		correctWith2FailsM = "102-103-etv0.2.3-stv1.2.3-getZkProof.json.failure.code_77.failure.code_77"
		correctWith2FailsL = "102-103-etv0.2.3-stv1.2.3-getZkProof.json.large.failure.code_77.failure.code_77"
		missingEtv         = "102-103-stv1.2.3-getZkProof.json"
		missingStv         = "102-103-etv0.2.3-getZkProof.json"
		notAPoint          = "102-103-etv0.2.3-getZkProofAjson"
		badName            = "102-103-etv0.2.3-stv1.2.3-getAggregatedProof.json"
	)

	// The responses in case of success
	var (
		respM = "responses/102-103-getZkProof.json"
		respL = "responses/102-103-getZkProof.json"
		// #nosec G101 -- Not a credential
		respWithFailM = "responses/102-103-getZkProof.json"
		// #nosec G101 -- Not a credential
		respWithFailL = "responses/102-103-getZkProof.json"
		// #nosec G101 -- Not a credential
		respWith2FailsM = "responses/102-103-getZkProof.json"
		// #nosec G101 -- Not a credential
		respWith2FailsL = "responses/102-103-getZkProof.json"
	)

	// The rename in case it is deferred to the large prover
	var (
		toLargeM           = "requests/102-103-etv0.2.3-stv1.2.3-getZkProof.json.large.failure.code_137"
		toLargeWithFailM   = "requests/102-103-etv0.2.3-stv1.2.3-getZkProof.json.large.failure.code_137"
		toLargeWith2FailsM = "requests/102-103-etv0.2.3-stv1.2.3-getZkProof.json.large.failure.code_137"
		toLargeWoEtv       = "requests/102-103-stv1.2.3-getZkProof.json.large.failure.code_137"
		toLargeWoStv       = "requests/102-103-etv0.2.3-getZkProof.json.large.failure.code_137"
	)

	// The rename in case it is a success
	var (
		successM           = "requests-done/102-103-etv0.2.3-stv1.2.3-getZkProof.json.success"
		successMWoStv      = "requests-done/102-103-etv0.2.3-getZkProof.json.success"
		successtWoEtv      = "requests-done/102-103-stv1.2.3-getZkProof.json.success"
		successL           = "requests-done/102-103-etv0.2.3-stv1.2.3-getZkProof.json.large.success"
		successWithFailM   = "requests-done/102-103-etv0.2.3-stv1.2.3-getZkProof.json.success"
		successWithFailL   = "requests-done/102-103-etv0.2.3-stv1.2.3-getZkProof.json.large.success"
		successWith2FailsM = "requests-done/102-103-etv0.2.3-stv1.2.3-getZkProof.json.success"
		successWith2FailsL = "requests-done/102-103-etv0.2.3-stv1.2.3-getZkProof.json.large.success"
	)

	// The rename in case it is a panic (code = 2)
	var (
		failM           = "requests-done/102-103-etv0.2.3-stv1.2.3-getZkProof.json.failure.code_2"
		failMWoStv      = "requests-done/102-103-etv0.2.3-getZkProof.json.failure.code_2"
		failtWoEtv      = "requests-done/102-103-stv1.2.3-getZkProof.json.failure.code_2"
		failL           = "requests-done/102-103-etv0.2.3-stv1.2.3-getZkProof.json.large.failure.code_2"
		failWithFailM   = "requests-done/102-103-etv0.2.3-stv1.2.3-getZkProof.json.failure.code_2"
		failWithFailL   = "requests-done/102-103-etv0.2.3-stv1.2.3-getZkProof.json.large.failure.code_2"
		failWith2FailsM = "requests-done/102-103-etv0.2.3-stv1.2.3-getZkProof.json.failure.code_2"
		failWith2FailsL = "requests-done/102-103-etv0.2.3-stv1.2.3-getZkProof.json.large.failure.code_2"
	)

	testcase := []inpFileNamesCases{
		{
			Ext: "", Fail: "code", ShouldMatch: true,
			Fnames:         []string{correctM, correctWithFailM, correctWith2FailsM, missingEtv, missingStv},
			Explainer:      "happy path, case M",
			ExpectedOutput: []string{respM, respWithFailM, respWith2FailsM, respM, respM},
			ExpToLarge:     []string{toLargeM, toLargeWithFailM, toLargeWith2FailsM, toLargeWoEtv, toLargeWoStv},
			ExpSuccess:     []string{successM, successWithFailM, successWith2FailsM, successtWoEtv, successMWoStv},
			ExpFailW2:      []string{failM, failWithFailM, failWith2FailsM, failtWoEtv, failMWoStv},
		},
		{
			Ext: "large", Fail: "code", ShouldMatch: true,
			Fnames:         []string{correctL, correctWithFailL, correctWith2FailsL},
			Explainer:      "happy path, case L",
			ExpectedOutput: []string{respL, respWithFailL, respWith2FailsL},
			ExpSuccess:     []string{successL, successWithFailL, successWith2FailsL},
			ExpFailW2:      []string{failL, failWithFailL, failWith2FailsL},
		},
		{
			Ext: "", Fail: "code", ShouldMatch: false,
			Fnames:    []string{correctL, correctWithFailL, correctWith2FailsL},
			Explainer: "M does not pick the files reserved for L",
		},
		{
			Ext: "large", Fail: "code", ShouldMatch: false,
			Fnames:    []string{correctM, correctWithFailM, correctWith2FailsM},
			Explainer: "L does not pick the files reserved for M",
		},
		{
			Ext: "", Fail: "code", ShouldMatch: false,
			Fnames:    []string{notAPoint, badName},
			Explainer: "M does not pick obviously invalid files",
		},
		{
			Ext: "large", Fail: "code", ShouldMatch: false,
			Fnames:    []string{missingEtv, missingStv, notAPoint, badName},
			Explainer: "L does not pick obviously invalid files",
		},
	}

	for _, c := range testcase {

		conf := config.Config{}
		conf.Version = "0.1.2"
		conf.Execution.CanRunFullLarge = c.Ext == "large"
		// conf.Execution.FilterInExtension = c.Ext

		def := ExecutionDefinition(&conf)

		t.Run(c.Explainer, func(t *testing.T) {
			runInpFileTestCase(t, &def, c)
		})
	}
}

func TestCompressionInFileRegexp(t *testing.T) {

	var (
		correctM           = "102-103-bcv0.2.3-ccv0.2.3-getZkBlobCompressionProof.json"
		correctWithFailM   = "102-103-bcv0.2.3-ccv0.2.3-getZkBlobCompressionProof.json.failure.code_77"
		correctWith2FailsM = "102-103-bcv0.2.3-ccv0.2.3-getZkBlobCompressionProof.json.failure.code_77.failure.code_77"
		notAPoint          = "102-103-bcv0.2.3-ccv0.2.3-getZkCompressionProofAjson"
		badName            = "102-103-bcv0.2.3-ccv0.2.3-getAggregatedProof.json"
		missingCv          = "102-103-getZkBlobCompressionProof.json"
		etvNoCv            = "102-103-bcv0.2.3-etv0.2.3-getZkBlobCompressionProof.json"
		missingBCv         = "102-103-ccv0.2.3-getZkBlobCompressionProof.json"
		missingCCv         = "102-103-bcv0.2.3-getZkBlobCompressionProof.json"
		withBlobHash       = "102-103-bcv0.2.3-ccv0.2.3-abcdef-getZkBlobCompressionProof.json"
		withBlobHash0x     = "102-103-bcv0.2.3-ccv0.2.3-0xabcdef-getZkBlobCompressionProof.json"
		with0x             = "102-103-bcv0.2.3-ccv0.2.3-0x-getZkBlobCompressionProof.json"
		withDoubleDash     = "102-103-bcv0.2.3-ccv0.2.3--getZkBlobCompressionProof.json"
	)

	var (
		respM = "responses/102-103-getZkBlobCompressionProof.json"
		// #nosec G101 -- Not a credential
		respWithFailM = "responses/102-103-getZkBlobCompressionProof.json"
		// #nosec G101 -- Not a credential
		respWith2FailsM = "responses/102-103-getZkBlobCompressionProof.json"
		// #nosec G101 -- Not a credential
		respWithBlobHash = "responses/102-103-abcdef-getZkBlobCompressionProof.json"
		// #nosec G101 -- Not a credential
		respWithBlobHash0x = "responses/102-103-0xabcdef-getZkBlobCompressionProof.json"
		// #nosec G101 -- Not a credential
		respWith0x = "responses/102-103-0x-getZkBlobCompressionProof.json"
		// #nosec G101 -- Not a credential
		respWithNoDoubleDash = "responses/102-103-getZkBlobCompressionProof.json"
	)

	testcase := []inpFileNamesCases{
		{
			Ext: "", Fail: "code", ShouldMatch: true,
			Fnames:         []string{correctM, correctWithFailM, correctWith2FailsM, withBlobHash, withBlobHash0x, withDoubleDash, with0x, missingCv, missingBCv, missingCCv},
			Explainer:      "happy path, case M",
			ExpectedOutput: []string{respM, respWithFailM, respWith2FailsM, respWithBlobHash, respWithBlobHash0x, respWithNoDoubleDash, respWith0x, respM, respM, respM},
		},
		{
			Ext: "", Fail: "code", ShouldMatch: false,
			Fnames:    []string{etvNoCv, notAPoint, badName},
			Explainer: "M does not pick obviously invalid files",
		},
	}

	for _, c := range testcase {

		conf := config.Config{}
		conf.Version = "0.1.2"
		conf.Execution.CanRunFullLarge = c.Ext == "large"
		// conf.Execution.FilterInExtension = c.Ext

		def := CompressionDefinition(&conf)

		t.Run(c.Explainer, func(t *testing.T) {
			runInpFileTestCase(t, &def, c)
		})
	}
}

func TestAggregatedInFileRegexp(t *testing.T) {

	var (
		correctM           = "102-103-abcdef0123-getZkAggregatedProof.json"
		correctWithFailM   = "102-103-abcdef0123-getZkAggregatedProof.json.failure.code_77"
		correctWith2FailsM = "102-103-abcdef0123-getZkAggregatedProof.json.failure.code_77.failure.code_77"
		missingContentHash = "102-103-getZkAggregatedProof.json"
		withEtv            = "102-103-etv0.2.3-getZkAggregatedProof.json"
		notAPoint          = "102-103-getZkAggregatedProofAjson"
		badName            = "102-103-abcdef0123-getCompressionProof.json"
	)

	var (
		// #nosec G101 -- Not a credential
		respM = "responses/102-103-getZkAggregatedProof.json"
		// #nosec G101 -- Not a credential
		respWithContentHash = "responses/102-103-abcdef0123-getZkAggregatedProof.json"
	)

	testcase := []inpFileNamesCases{
		{
			Ext: "", Fail: "code", ShouldMatch: true,
			Fnames:         []string{correctM, correctWithFailM, correctWith2FailsM, missingContentHash},
			Explainer:      "happy path, case M",
			ExpectedOutput: []string{respWithContentHash, respWithContentHash, respWithContentHash, respM},
		},
		{
			Ext: "", Fail: "code", ShouldMatch: false,
			Fnames:    []string{withEtv, notAPoint, badName},
			Explainer: "M does not pick obviously invalid files",
		},
	}

	for _, c := range testcase {

		conf := config.Config{}
		conf.Version = "0.1.2"
		conf.Execution.CanRunFullLarge = c.Ext == "large"
		// conf.Execution.FilterInExtension = c.Ext

		def := AggregatedDefinition(&conf)

		t.Run(c.Explainer, func(t *testing.T) {
			runInpFileTestCase(t, &def, c)
		})
	}
}

func runInpFileTestCase(t *testing.T, def *JobDefinition, c inpFileNamesCases) {

	for i, fname := range c.Fnames {

		// NB: if the regexp matches but the fields cannot be parsed
		// this will panic and fail the test. This is intentional. All
		// errors must be caught by the input file regexp.
		job, err := NewJob(def, fname)

		if c.ShouldMatch {
			if !assert.NoError(t, err, fname) {
				// stop there for this iteration
				continue
			}

			// Then try to format the response of the job
			resp, err := job.ResponseFile()
			if assert.NoErrorf(t, err, "cannot produce a response for job %s", fname) {
				assert.Equal(t, c.ExpectedOutput[i], resp, "wrong output file")
			}

			// Try the name of the large one. If the case is specifying some
			// expected values
			if len(c.ExpToLarge) > 0 {
				toLarge, err := job.DeferToLargeFile(
					Status{ExitCode: 137},
				)

				if assert.NoError(t, err, "cannot produce name for the too large job") {
					assert.Equal(t, c.ExpToLarge[i], toLarge)
				}
			}

			// Try the success file
			if len(c.ExpSuccess) > 0 {
				toSuccess := job.DoneFile(Status{ExitCode: 0})
				assert.Equal(t, c.ExpSuccess[i], toSuccess)
			}

			// Try the code 2 file
			if len(c.ExpFailW2) > 0 {
				toFail2 := job.DoneFile(Status{ExitCode: 2})
				assert.Equal(t, c.ExpFailW2[i], toFail2)
			}

		} else {
			assert.Errorf(
				t, err, fname,
				"%v should not match %s",
				fname, def.InputFileRegexp.String(),
			)
		}
	}
}

func TestFailSuffixMatching(t *testing.T) {

	testcases := []struct {
		s      string
		ncodes int
	}{
		{s: "abds.failure.code_1.failure.code_2", ncodes: 2},
		{s: "abds.failure.code_1", ncodes: 1},
		{s: "abds.failure.code1", ncodes: 0},
		{s: "abds.failure.code__1", ncodes: 0},
		{s: "abds", ncodes: 0},
	}

	r := matchFailureSuffix("code")

	for _, c := range testcases {
		// Count the number of matches
		found := 0
		m, _ := r.FindStringMatch(c.s)
		for m != nil {
			found++
			m, _ = r.FindNextMatch(m)
		}
		assert.Equalf(t, c.ncodes, found, "failed to parse %v", c.s)
	}
}

// This test ensures that the naming convention is respected by the file-watcher
// for bootstrap jobs: only valid inputs are recognized, and the corresponding
// output files are also generated.
func TestBootstrapInFileRegexp(t *testing.T) {

	conf := config.Config{}
	conf.Version = "0.1.2"

	var (
		correct           = "102-103-etv0.2.3-stv1.2.3-getZkProof.json"
		correctWithFail   = "102-103-etv0.2.3-stv1.2.3-getZkProof.json.failure.code_77"
		correctWith2Fails = "102-103-etv0.2.3-stv1.2.3-getZkProof.json.failure.code_77.failure.code_77"
		missingEtv        = "102-103-stv1.2.3-getZkProof.json"
		missingStv        = "102-103-etv0.2.3-getZkProof.json"
		notAPoint         = "102-103-etv0.2.3-getZkProofAjson"
		badName           = "102-103-etv0.2.3-stv1.2.3-getAggregatedProof.json"
	)

	// The requests in case of success
	var (
		resp         = "requests/102-103-metadata-getZkProof.json"
		respWithFail = "requests/102-103-metadata-getZkProof.json"
		// #nosec G101 -- Not a credential
		respWith2Fails = "requests/102-103-metadata-getZkProof.json"
	)

	// The rename in case it is a success
	var (
		success           = "requests-done/102-103-etv0.2.3-stv1.2.3-getZkProof.json.success"
		successWoStv      = "requests-done/102-103-etv0.2.3-getZkProof.json.success"
		successWoEtv      = "requests-done/102-103-stv1.2.3-getZkProof.json.success"
		successWithFail   = "requests-done/102-103-etv0.2.3-stv1.2.3-getZkProof.json.success"
		successWith2Fails = "requests-done/102-103-etv0.2.3-stv1.2.3-getZkProof.json.success"
	)

	// The rename in case it is a panic (code = 2)
	var (
		fail           = "requests-done/102-103-etv0.2.3-stv1.2.3-getZkProof.json.failure.code_2"
		failWoStv      = "requests-done/102-103-etv0.2.3-getZkProof.json.failure.code_2"
		failWoEtv      = "requests-done/102-103-stv1.2.3-getZkProof.json.failure.code_2"
		failWithFail   = "requests-done/102-103-etv0.2.3-stv1.2.3-getZkProof.json.failure.code_2"
		failWith2Fails = "requests-done/102-103-etv0.2.3-stv1.2.3-getZkProof.json.failure.code_2"
	)

	testcase := []inpFileNamesCases{
		{
			Ext: "", Fail: "code", ShouldMatch: true,
			Fnames:         []string{correct, correctWithFail, correctWith2Fails, missingEtv, missingStv},
			Explainer:      "happy path, Bootstrapper picks valid files",
			ExpectedOutput: []string{resp, respWithFail, respWith2Fails, resp, resp},
			ExpSuccess:     []string{success, successWithFail, successWith2Fails, successWoEtv, successWoStv},
			ExpFailW2:      []string{fail, failWithFail, failWith2Fails, failWoEtv, failWoStv},
		},
		{
			Ext: "", Fail: "code", ShouldMatch: false,
			Fnames:    []string{notAPoint, badName},
			Explainer: "Bootstrapper rejects invalid files",
		},
	}

	for _, c := range testcase {
		def := BootstrapDefinition(&conf)
		t.Run(c.Explainer, func(t *testing.T) {
			runInpFileTestCase(t, &def, c)
		})
	}
}

// This test ensures that the naming convention is respected by the file-watcher
// for conglomeration jobs: only valid inputs are recognized, and the corresponding
// output files are also generated.
func TestConglomerationInFileRegexp(t *testing.T) {

	conf := config.Config{}
	conf.Version = "0.1.2"

	var (
		correct           = "200-201-metadata-getZkProof.json"
		correctWithFail   = "200-201-metadata-getZkProof.json.failure.code_77"
		correctWith2Fails = "200-201-metadata-getZkProof.json.failure.code_77.failure.code_77"
		notAPoint         = "200-201-metadata-getZkProofAjson"
		badName           = "200-201-getAggregatedProof.json"
	)

	// The requests in case of success
	var (
		// #nosec G101 -- this is a test fixture filename, not credentials
		resp = "responses/200-201-getZkProof.json"
		// #nosec G101 -- this is a test fixture filename, not credentials
		respWithFail = "responses/200-201-getZkProof.json"
		// #nosec G101 -- this is a test fixture filename, not credentials
		respWith2Fails = "responses/200-201-getZkProof.json"
	)

	// The rename in case it is a success
	var (
		success           = "requests-done/200-201-metadata-getZkProof.json.success"
		successWithFail   = "requests-done/200-201-metadata-getZkProof.json.success"
		successWith2Fails = "requests-done/200-201-metadata-getZkProof.json.success"
	)

	// The rename in case it is a panic (code = 2)
	var (
		fail           = "requests-done/200-201-metadata-getZkProof.json.failure.code_2"
		failWithFail   = "requests-done/200-201-metadata-getZkProof.json.failure.code_2"
		failWith2Fails = "requests-done/200-201-metadata-getZkProof.json.failure.code_2"
	)

	testcase := []inpFileNamesCases{
		{
			Ext: "", Fail: "code", ShouldMatch: true,
			Fnames:         []string{correct, correctWithFail, correctWith2Fails},
			Explainer:      "happy path, Conglomerator picks valid files",
			ExpectedOutput: []string{resp, respWithFail, respWith2Fails},
			ExpSuccess:     []string{success, successWithFail, successWith2Fails},
			ExpFailW2:      []string{fail, failWithFail, failWith2Fails},
		},
		{
			Ext: "", Fail: "code", ShouldMatch: false,
			Fnames:    []string{notAPoint, badName},
			Explainer: "Conglomerator rejects invalid files",
		},
	}

	for _, c := range testcase {
		def := ConglomerationDefinition(&conf)
		t.Run(c.Explainer, func(t *testing.T) {
			runInpFileTestCase(t, &def, c)
		})
	}
}

// This test ensures that the naming convention is respected by the file-watcher
// for GL module jobs: only valid inputs are recognized. GL jobs do not produce
// any output artifact (only /dev/null), so ExpectedOutput is empty, but success
// and failure renames still happen.
func TestGLModsInFileRegexp(t *testing.T) {
	conf := config.Config{}
	conf.Version = "0.1.2"
	conf.ExecutionLimitless.WitnessDir = "/tmp/exec-limitless/witness"

	var (
		correct           = "22504197-22504198-seg-0-mod-0-gl-wit.bin"
		correctWithFail   = "22504197-22504198-seg-0-mod-0-gl-wit.bin.failure.code_77"
		correctWith2Fails = "22504197-22504198-seg-0-mod-0-gl-wit.bin.failure.code_77.failure.code_77"
		notAPoint         = "22504197-22504198-seg-0-mod-0-gl-witA.bin"
		badName           = "22504197-22504198-gl-wit.json"
	)

	var (
		resp = "/dev/null"
	)

	for _, module := range config.ALL_MODULES {
		testcase := []inpFileNamesCases{
			{
				Ext: "", Fail: "code", ShouldMatch: true,
				Fnames:    []string{correct, correctWithFail, correctWith2Fails},
				Explainer: fmt.Sprintf("happy path, GL module %s picks valid files", module),
				ExpectedOutput: []string{
					resp, resp, resp, // no real output files, GL jobs write to /dev/null
				},
				ExpSuccess: []string{
					filepath.Join(conf.ExecutionLimitless.WitnessDir, "GL", module, "requests-done", correct+".success"),
					filepath.Join(conf.ExecutionLimitless.WitnessDir, "GL", module, "requests-done", correct+".success"),
					filepath.Join(conf.ExecutionLimitless.WitnessDir, "GL", module, "requests-done", correct+".success"),
				},
				ExpFailW2: []string{
					filepath.Join(conf.ExecutionLimitless.WitnessDir, "GL", module, "requests-done", correct+".failure.code_2"),
					filepath.Join(conf.ExecutionLimitless.WitnessDir, "GL", module, "requests-done", correct+".failure.code_2"),
					filepath.Join(conf.ExecutionLimitless.WitnessDir, "GL", module, "requests-done", correct+".failure.code_2"),
				},
			},
			{
				Ext: "", Fail: "code", ShouldMatch: false,
				Fnames:    []string{notAPoint, badName},
				Explainer: fmt.Sprintf("GL module %s rejects invalid files", module),
			},
		}

		for _, c := range testcase {
			def := GLDefinitionForModule(&conf, module)
			t.Run(c.Explainer, func(t *testing.T) {
				runInpFileTestCase(t, &def, c)
			})
		}
	}
}

// This test ensures that the naming convention is respected by the file-watcher
// for LPP module jobs: only valid inputs are recognized. LPP jobs do not produce
// any output artifact (only /dev/null), so ExpectedOutput is empty, but success
// and failure renames still happen.
func TestLPPModsInFileRegexp(t *testing.T) {
	conf := config.Config{}
	conf.Version = "0.1.2"
	conf.ExecutionLimitless.WitnessDir = "/tmp/exec-limitless/witness"

	var (
		correct           = "22504197-22504198-seg-0-mod-0-lpp-wit.bin"
		correctWithFail   = "22504197-22504198-seg-0-mod-0-lpp-wit.bin.failure.code_77"
		correctWith2Fails = "22504197-22504198-seg-0-mod-0-lpp-wit.bin.failure.code_77.failure.code_77"
		notAPoint         = "22504197-22504198-seg-0-mod-0-lpp-witA.bin"
		badName           = "22504197-22504198-lpp-wit.json"
	)

	var (
		resp = "/dev/null"
	)

	for _, module := range config.ALL_MODULES {
		testcase := []inpFileNamesCases{
			{
				Ext: "", Fail: "code", ShouldMatch: true,
				Fnames:    []string{correct, correctWithFail, correctWith2Fails},
				Explainer: fmt.Sprintf("happy path, lpp module %s picks valid files", module),
				ExpectedOutput: []string{
					resp, resp, resp, // no real output files, lpp jobs write to /dev/null
				},
				ExpSuccess: []string{
					filepath.Join(conf.ExecutionLimitless.WitnessDir, "LPP", module, "requests-done", correct+".success"),
					filepath.Join(conf.ExecutionLimitless.WitnessDir, "LPP", module, "requests-done", correct+".success"),
					filepath.Join(conf.ExecutionLimitless.WitnessDir, "LPP", module, "requests-done", correct+".success"),
				},
				ExpFailW2: []string{
					filepath.Join(conf.ExecutionLimitless.WitnessDir, "LPP", module, "requests-done", correct+".failure.code_2"),
					filepath.Join(conf.ExecutionLimitless.WitnessDir, "LPP", module, "requests-done", correct+".failure.code_2"),
					filepath.Join(conf.ExecutionLimitless.WitnessDir, "LPP", module, "requests-done", correct+".failure.code_2"),
				},
			},
			{
				Ext: "", Fail: "code", ShouldMatch: false,
				Fnames:    []string{notAPoint, badName},
				Explainer: fmt.Sprintf("LPP module %s rejects invalid files", module),
			},
		}

		for _, c := range testcase {
			def := LPPDefinitionForModule(&conf, module)
			t.Run(c.Explainer, func(t *testing.T) {
				runInpFileTestCase(t, &def, c)
			})
		}
	}
}
